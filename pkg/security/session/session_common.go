/*
Package session предоставляет функциональность для управления пользовательскими сессиями в веб-приложении.
Он включает создание, проверку и уничтожение сессий, а также middleware для защиты маршрутов,
требующих аутентификации.

Пакет определяет структуру Session для хранения данных, связанных с сессией, и интерфейс
для управления сессиями. Также предоставляется middleware для проверки сессий в входящих HTTP-запросах.
*/

package session

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/unclaim/chegonado.git/internal/shared/common_errors"
)

// Session представляет собой пользовательскую сессию с соответствующей информацией.
type Session struct {
	UserID          int64     `json:"user_id"`          // Уникальный идентификатор пользователя
	ID              string    `json:"id"`               // Уникальный идентификатор сессии
	IP              string    `json:"ip"`               // IP-адрес пользователя
	Browser         string    `json:"browser"`          // Браузер, используемый пользователем
	OperatingSystem string    `json:"operating_system"` // Операционная система устройства пользователя
	City            string    `json:"city"`             // Город, из которого пользователь получает доступ
	Location        string    `json:"location"`         // Общее местоположение пользователя
	CreatedAt       time.Time `json:"created_at"`       // Время создания сессии
	FirstLogin      time.Time `json:"first_login"`      // Время первого входа пользователя
	LastLogin       time.Time `json:"last_login"`       // Время последнего входа пользователя
}

// UserInterface определяет методы, которые должен реализовать пользователь для работы с сессиями.
type UserInterface interface {
	GetID() int64         // Возвращает уникальный идентификатор пользователя
	GetUsrVersion() int64 // Возвращает версию данных или профиля пользователя
}

// SessionManager определяет методы для управления сессиями.
type SessionManager interface {
	Check(context.Context, *http.Request) (*Session, error)                          // Проверяет и извлекает сессию из запроса
	Create(context.Context, http.ResponseWriter, UserInterface, *http.Request) error // Создает новую сессию для пользователя
	DestroyCurrent(context.Context, http.ResponseWriter, *http.Request) error        // Уничтожает текущую сессию для запроса
	DestroyAll(context.Context, http.ResponseWriter, UserInterface) error            // Уничтожает все сессии, связанные с пользователем
}

// key используется как ключ контекста для хранения и извлечения сессий.
type key int

const sessionKey key = iota

// ErrNoAuth возвращается, когда в контексте не найдена аутентифицированная сессия.
var ErrNoAuth = errors.New("ошибка получения сессии")

// SessionFromContext извлекает Session из контекста, если она существует.
func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(sessionKey).(*Session)
	if !ok {
		return nil, ErrNoAuth
	}
	return sess, nil
}

// noAuthUrls содержит конечные точки (эндпоинты), которые не требуют аутентификации.
var noAuthUrls = map[string]struct{}{
	"/api/tasks/categories":         {},
	"/api/unclaimeds":               {},
	"/api/user/login":               {},
	"/api/user/signup":              {},
	"/api/username/":                {},
	"/api/user/check-session":       {},
	"/public/":                      {},
	"/password_reset":               {},
	"/send_password_reset":          {},
	"/reset_password":               {},
	"/update_password":              {},
	"/password_reset_confirmation":  {},
	"/api/check-session":            {},
	"/api/user/send_password_reset": {},
	"/api/user/update_password":     {},
	"/api/user/reset_password":      {},
	"/api/tasks":                    {},
	"/api/bots":                     {},
	"/api/categories":               {},
	"/api/subcategories":            {},
	"/api/reviews_bots":             {},
	"/auth/login/verify-code":       {},
	"/auth/login/send-code":         {},
	"/api/auth/signup/verify-code":  {},
	"/auth/resend-code":             {},
	"/api/auth/signup/send-code":    {},
	"/api/user/check-user":          {},
}

// AuthMiddleware является HTTP middleware, который проверяет наличие действительной сессии.
// Если действительная сессия не найдена и путь запроса требует аутентификации,
// он отвечает статусом "Forbidden". В противном случае он позволяет доступ к защищенным маршрутам,
// добавляя сессию в контекст запроса.
func AuthMiddleware(sm SessionManager, ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentPath := r.URL.Path

		if _, ok := noAuthUrls[currentPath]; ok {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(currentPath, "/api/username/") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(currentPath, "/swagger/") {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(ctx, r)
		if err != nil {
			common_errors.NewAppError(w, r, errors.New("доступ запрещен"), http.StatusForbidden)
			return
		}

		updatedCtx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(updatedCtx))
	})
}
