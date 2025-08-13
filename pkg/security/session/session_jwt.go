// Package session предоставляет функциональность для управления сессиями пользователей с использованием JWT.
// Он включает в себя создание, проверку и уничтожение сессий.
package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log/slog" // Импортируем пакет slog для логирования

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/unclaim/chegonado/internal/shared/utils/randutils"
)

// SessionsJWT представляет собой структуру для управления сессиями с использованием JWT.
type SessionsJWT struct {
	Secret []byte // Секретный ключ для подписи JWT
}

// SessionJWTClaims представляет собой структуру для хранения данных о сессии в JWT.
type SessionJWTClaims struct {
	UserID             int64 `json:"uid"` // Идентификатор пользователя
	jwt.StandardClaims       // Стандартные поля JWT
}

// NewSessionsJWT создает новый экземпляр SessionsJWT с заданным секретом.
// Параметры:
//   - secret: Секретный ключ для подписи JWT.
//
// Возвращает указатель на новый экземпляр SessionsJWT.
func NewSessionsJWT(secret string) *SessionsJWT {
	return &SessionsJWT{
		Secret: []byte(secret),
	}
}

// parseSecretGetter возвращает секретный ключ для подписи JWT.
// Параметры:
//   - token: Токен JWT, который нужно проверить.
//
// Возвращает секретный ключ и ошибку, если она произошла.
func (sm *SessionsJWT) parseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("неверный метод подписи")
	}
	return sm.Secret, nil
}

// Check проверяет наличие активной сессии для данного HTTP-запроса.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - r: HTTP-запрос, содержащий куки с идентификатором сессии.
//
// Возвращает указатель на объект Session и ошибку, если она произошла.
func (sm *SessionsJWT) Check(ctx context.Context, r *http.Request) (*Session, error) {
	sessionCookie, err := r.Cookie("session")
	if err == http.ErrNoCookie {
		slog.Info("Проверка сессии: отсутствует куки")
		return nil, ErrNoAuth
	}

	payload := &SessionJWTClaims{}
	_, err = jwt.ParseWithClaims(sessionCookie.Value, payload, sm.parseSecretGetter)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать jwt токен: %v", err)
	}

	if payload.Valid() != nil {
		return nil, fmt.Errorf("недействительный jwt токен: %v", err)
	}

	return &Session{
		ID:     payload.Id,
		UserID: payload.UserID,
	}, nil
}

// Create создает новую сессию для указанного пользователя и устанавливает соответствующий куки.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - w: HTTP-ответ для установки куки.
//   - user: Интерфейс пользователя, для которого создается сессия.
//
// Возвращает ошибку при неудаче.
func (sm *SessionsJWT) Create(ctx context.Context, w http.ResponseWriter, user UserInterface) error {
	data := SessionJWTClaims{
		UserID: int64(user.GetID()),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(90 * 24 * time.Hour).Unix(), // 90 дней
			IssuedAt:  time.Now().Unix(),
			Id:        randutils.RandStringRunes(32),
		},
	}
	sessVal, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, data).SignedString(sm.Secret)

	cookie := &http.Cookie{
		Name:    "session",
		Value:   sessVal,
		Expires: time.Now().Add(90 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	slog.Info("Создана новая сессия", "user_id", user.GetID())
	return nil
}

// DestroyCurrent уничтожает текущую сессию пользователя и удаляет соответствующий куки.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - w: HTTP-ответ для удаления куки.
//   - r: HTTP-запрос для получения информации о текущей сессии.
//
// Возвращает ошибку при неудаче.
func (sm *SessionsJWT) DestroyCurrent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cookie := http.Cookie{
		Name:    "session",
		Expires: time.Now().AddDate(0, 0, -1), // Устанавливаем истекший срок действия куки
		Path:    "/",
	}
	http.SetCookie(w, &cookie)

	slog.Info("Уничтожена текущая сессия")

	return nil
}

// DestroyAll уничтожает все сессии для указанного пользователя.
// В данной реализации функция не может удалить другие активные сессии из-за использования JWT,
// так как они не хранятся на сервере.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - w: HTTP-ответ (не используется в данной функции).
//   - user: Интерфейс пользователя для удаления всех его сессий.
//
// Возвращает ошибку при неудаче.
func (sm *SessionsJWT) DestroyAll(ctx context.Context, w http.ResponseWriter, user UserInterface) error {
	slog.Warn("Попытка уничтожить все сессии невозможна из-за использования JWT.")

	return nil
}
