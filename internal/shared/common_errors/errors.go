package common_errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// DebugMode определяет, нужно ли включать трассировку стека в ответах об ошибках.
// В продакшене рекомендуется устанавливать в false.
const DebugMode = true

// NewAppError является основной функцией для централизованной обработки ошибок API.
// Она логирует ошибку, генерирует структурированный JSON-ответ с информацией об ошибке
// и отправляет его клиенту с соответствующим HTTP-статусом.
func NewAppError(w http.ResponseWriter, req *http.Request, err error, statusCode int) {
	// Дополнительные атрибуты для логирования: метод, путь, IP-адрес клиента.
	ipAddress := getClientIP(req)

	slog.Error(
		"Произошла ошибка API", // Сообщение для лога
		slog.String("error", err.Error()),
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
		slog.String("remote_ip", ipAddress),
		slog.Int("status_code", statusCode),
	)

	var trace []byte
	if DebugMode {
		// Получаем трассировку стека только в режиме отладки.
		trace = debug.Stack()
	}

	errorResp := &ErrorResponse{
		ErrorMessage: fmt.Sprintf("%s", err),
		ErrorType:    determineErrorType(statusCode),
		StackTrace:   formatStackTrace(trace), // Красиво оформляем стектрейс
	}

	response := &Response{
		StatusCode: statusCode,
		Body:       errorResp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		// Если не удалось отправить JSON-ответ, логируем это и отправляем обычную ошибку HTTP.
		slog.Error("Не удалось отправить JSON-ответ клиенту",
			slog.String("original_error", err.Error()),
			slog.String("encoding_error", encodeErr.Error()),
			slog.Int("status_code", statusCode),
		)
		http.Error(w, "Внутренняя ошибка сервера при формировании ответа", http.StatusInternalServerError)
	}
}

// GetClientIP извлекает IP-адрес клиента из заголовков запроса или RemoteAddr.
func getClientIP(r *http.Request) string {
	// Сначала пробуем получить из заголовка X-Real-IP (часто используется прокси-серверами).
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Затем проверяем X-Forwarded-For (также используется прокси-серверами, может содержать несколько IP).
	forwardIP := r.Header.Get("X-Forwarded-For")
	if forwardIP != "" {
		// Берем первый IP, если их несколько.
		parts := strings.Split(forwardIP, ",")
		return strings.TrimSpace(parts[0])
	}

	// Если заголовков нет, берем IP из RemoteAddr.
	addrParts := strings.Split(r.RemoteAddr, ":")
	if len(addrParts) > 0 {
		return addrParts[0]
	}

	return "unknown"
}

// formatStackTrace форматирует необработанный байтовый срез трассировки стека в читабельный срез строк.
func formatStackTrace(trace []byte) []string {
	if len(trace) == 0 {
		return nil
	}

	lines := bytes.Split(trace, []byte("\n"))

	// Максимальное количество строк стектрейса для вывода, чтобы не перегружать ответ.
	maxLines := 10
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	filteredLines := filterRelevantLines(lines)

	return filteredLines
}

// filterRelevantLines отфильтровывает нерелевантные строки из трассировки стека.
func filterRelevantLines(lines [][]byte) []string {
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		str := string(line)
		if containsImportant(str) {
			formatted := formatStackFrame(str)
			result = append(result, formatted)
		}
	}
	return result
}

// containsImportant проверяет, является ли строка трассировки стека важной (т.е. содержит ли она информацию о коде приложения).
func containsImportant(line string) bool {
	// Игнорируем стандартные горутинные записи и смещения, которые обычно не несут полезной информации.
	if strings.HasPrefix(line, "goroutine ") || strings.HasPrefix(line, "created by ") || strings.HasSuffix(line, "+0x") || strings.HasPrefix(line, "[signal ") {
		return false
	}
	// Важные строки трассировки обычно содержат отступы (табуляции).
	return strings.Contains(line, "\t")
}

// formatStackFrame форматирует отдельный кадр трассировки стека.
func formatStackFrame(frame string) string {
	parts := strings.Split(frame, "\t")
	if len(parts) < 2 {
		return strings.TrimSpace(frame) // Если формат неожиданный, возвращаем как есть.
	}

	functionInfo := strings.TrimSpace(parts[0])
	locationInfo := strings.TrimSpace(parts[1])

	cleanLocation := cleanFilePath(locationInfo)

	return functionInfo + " (" + cleanLocation + ")"
}

// cleanFilePath удаляет базовый путь проекта из пути к файлу в трассировке стека
// для более компактного и читабельного вывода.
func cleanFilePath(path string) string {
	// Важно: замените "your-app-name/" на фактическое имя вашего Go-модуля (из go.mod)
	// Например, если ваш go.mod содержит `module chegonado.com/app`,
	// то здесь должно быть `chegonado.com/app/`
	const projectBase = "your-app-name/" // <--- ОБЯЗАТЕЛЬНО ЗАМЕНИТЕ ЭТО НА ИМЯ ВАШЕГО МОДУЛЯ

	if index := strings.Index(path, projectBase); index >= 0 {
		// Возвращаем путь относительно корня вашего модуля.
		return path[index+len(projectBase):]
	}

	// Также можно добавить более общее сокращение для Go-модулей, если нужно.
	if index := strings.Index(path, "/go/src/"); index >= 0 {
		return path[index+len("/go/src/"):]
	}

	return path
}

// determineErrorType сопоставляет HTTP-статус-код с предопределенным строковым типом ошибки.
func determineErrorType(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "NotFound"
	case http.StatusBadRequest:
		return "BadRequest"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusGone:
		return "Gone"
	case http.StatusPreconditionFailed:
		return "PreconditionFailed"
	case http.StatusUnprocessableEntity:
		return "UnprocessableEntity"
	case http.StatusLocked:
		return "Locked"
	case http.StatusTooManyRequests:
		return "TooManyRequests"
	case http.StatusServiceUnavailable:
		return "ServiceUnavailable"
	case http.StatusGatewayTimeout:
		return "GatewayTimeout"
	case http.StatusMethodNotAllowed:
		return "MethodNotAllowed"
	case http.StatusInternalServerError:
		fallthrough
	case http.StatusBadGateway:
		fallthrough
	case http.StatusHTTPVersionNotSupported:
		return "InternalServerError"
	default:
		return "UnknownError"
	}
}

// ErrorResponse представляет структуру для ошибок API, отправляемых клиенту.
type ErrorResponse struct {
	ErrorMessage string   `json:"errorMessage,omitempty"` // Сообщение об ошибке, предназначенное для клиента
	ErrorType    string   `json:"errorType,omitempty"`    // Категория ошибки (например, "BadRequest", "NotFound")
	StackTrace   []string `json:"stackTrace,omitempty"`   // Стек вызовов (только в режиме отладки, для разработчиков)
}

// Response представляет собой общую структуру HTTP-ответа,
// включающую статус-код и тело, которое может быть как данными, так и ошибкой.
type Response struct {
	StatusCode int         `json:"statusCode,omitempty"` // Код статуса HTTP-ответа
	Body       interface{} `json:"body,omitempty"`       // Тело ответа (может содержать любые данные или ErrorResponse)
}
