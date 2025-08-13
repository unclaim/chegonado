package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/unclaim/chegonado/internal/shared/common_errors"
	"github.com/unclaim/chegonado/internal/users/domain"

	"golang.org/x/crypto/bcrypt"
)

// Helper функция для генерации криптографически стойкого 6-значного кода
// Генерация криптографически стойкого 6-значного кода
func GenerateSecureCode() (int, error) {
	var p [4]byte
	if _, err := rand.Read(p[:]); err != nil {
		return 0, fmt.Errorf("ошибка при генерации случайных байтов: %w", err)
	}
	// Преобразуем байты в число
	randomNumber := int(p[0])<<24 | int(p[1])<<16 | int(p[2])<<8 | int(p[3])
	// Берем остаток по диапазону 100000-999999
	code := 100000 + (randomNumber % 900000)
	return code, nil
}

// NewResponse отправляет ответ клиенту в виде JSON.
// Используется для отправки как успешных ответов, так и обернутых ошибок.
func NewResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		StatusCode: statusCode,
		Body:       data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Не удалось закодировать JSON-ответ",
			slog.Any("data", data),
			slog.String("error", err.Error()),
		)
		// На случай, если исходная запись JSON не удалась, возвращаем HTTP ошибку
		http.Error(w, "Внутренняя ошибка сервера при формировании ответа", http.StatusInternalServerError)
	}
}

// Response представляет собой общую структуру HTTP-ответа,
// включающую статус-код и тело, которое может быть как данными, так и ошибкой.
type Response struct {
	StatusCode int         `json:"statusCode,omitempty"` // Код статуса HTTP-ответа
	Body       interface{} `json:"body,omitempty"`       // Тело ответа (может содержать любые данные или ErrorResponse)
}

func ComparePasswords(currentHash, newPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(newPassword))
	if err == nil {
		return errors.New("новый пароль не может совпадать со старым паролем")
	}
	return nil
}
func ParseRowToUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{}

	err := row.Scan(
		&user.ID,
		&user.Version,
		&user.Blacklisted,
		&user.Sex,
		&user.FollowersCount,
		&user.Verified,
		&user.NoAds,
		&user.CanUploadShot,
		&user.Pro,
		&user.Type,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Username,
		&user.PasswordHash,
		&user.Bdate,
		&user.Phone,
		&user.Email,
		&user.HTMLURL,
		&user.AvatarURL,
		&user.Bio,
		&user.Location,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return user, nil
}

func ParseRowToUserReset(row pgx.Row) (*domain.User, error) {
	user := &domain.User{}

	err := row.Scan(
		&user.ID,
		&user.Version,
		&user.Username,
		&user.Email,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при сканировании данных пользователя: %w", err)
	}
	return user, nil
}

func IsValidEmailFormat(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
func IsDuplicatedKeyError(err error) bool {
	var perr *pgconn.PgError
	if errors.As(err, &perr) {
		return perr.Code == "DUPLICATED_KEY"
	}
	return false
}

func HandleDuplicateError(err error) error {
	if err == nil {
		return nil
	}

	if IsDuplicatedKeyError(err) {
		return fmt.Errorf("попробуйте изменить уникальное значение, данное поле уже занято: %w", err)
	}

	return err
}
func SaveToJSON(data domain.User, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла %s: %w", filename, err)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			slog.Debug("ошибка закрытия файла %s: %v", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	return nil
}
func ValidateEmail(email string) error {
	if !IsValidEmailFormat(email) {
		return errors.New("недопустимый формат адреса электронной почты")
	}
	return nil
}
func CheckMethod(w http.ResponseWriter, r *http.Request, allowedMethods []string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}
	common_errors.NewAppError(w, r, fmt.Errorf("неверный метод (%s)", r.Method), http.StatusMethodNotAllowed)
	return false
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if ok := errors.As(err, &pgErr); ok {
		return pgErr.Code == "23505"
	}
	return false
}

func ComparePasswordAndHash(pass, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return false, err
	}
	return true, nil
}
func DeleteUserFiles(userID int64) error {
	directory := filepath.Join("./uploads", fmt.Sprint(userID))
	err := os.RemoveAll(directory)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return fmt.Errorf("директория пользователя с ID %d не найдена: %w", userID, err)
		case os.IsPermission(err):
			return fmt.Errorf("недостаточно прав для удаления директории пользователя с ID %d: %w", userID, err)
		default:
			return fmt.Errorf("не удалось удалить файлы пользователя с ID %d: %w", userID, err)
		}
	}
	return nil
}
