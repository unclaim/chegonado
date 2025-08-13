package public

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIResponse представляет общий ответ API, содержащий код состояния и тело ответа.
// Описание: Общий ответ API с кодом состояния и телом ответа.
type APIResponse struct {
	// StatusCode: Код состояния HTTP (например, 200, 404 и т.д.).
	// Обязательное поле.
	StatusCode int `json:"statusCode"`

	// Body: Тело ответа, может содержать разные данные в зависимости от контекста.
	// Необязательное поле.
	Body interface{} `json:"body,omitempty"`
}

// Обработчик общедоступного ресурса, доступного без аутентификации.
func Public(w http.ResponseWriter, r *http.Request) {
	logger := slog.Default() // Получаем стандартный логгер

	// Формируем успешный ответ
	response := &APIResponse{
		StatusCode: http.StatusOK,
		Body:       map[string]string{"сообщение": "Добро пожаловать! Это общедоступная страница."},
	}

	// Преобразуем ответ в JSON
	respBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("ошибка преобразования ответа в JSON", "ошибка", err)
		http.Error(w, "ошибка подготовки ответа", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Отправляем сформированный ответ клиенту
	if _, err = w.Write(respBytes); err != nil {
		logger.Error("ошибка отправки ответа", "ошибка", err)
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		return
	}
}
