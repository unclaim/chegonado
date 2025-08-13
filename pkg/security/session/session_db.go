// Package session предоставляет функциональность для управления сессиями пользователей.
// Он включает в себя создание, проверку и уничтожение сессий в базе данных.
package session

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"log/slog" // Импортируем пакет slog для логирования

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/unclaim/chegonado/internal/shared/utils/randutils"
)

// SessionsDB представляет собой структуру, которая содержит пул соединений с базой данных.
type SessionsDB struct {
	dbpool *pgxpool.Pool
}

// NewSessionsDB создает новый экземпляр SessionsDB с заданным пулом соединений.
// Параметры:
//   - dbpool: Пул соединений к базе данных.
//
// Возвращает указатель на новый экземпляр SessionsDB.
func NewSessionsDB(dbpool *pgxpool.Pool) *SessionsDB {
	return &SessionsDB{
		dbpool: dbpool,
	}
}

// Check проверяет наличие активной сессии для данного HTTP-запроса.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - r: HTTP-запрос, содержащий куки с идентификатором сессии.
//
// Возвращает указатель на объект Session и ошибку, если она произошла.
func (sm *SessionsDB) Check(ctx context.Context, r *http.Request) (*Session, error) {

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
		}),
	)
	sessionCookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		logger.Info("Cookie not found",
			slog.Group("req",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String())),
			slog.Int("status", http.StatusNotFound),
			slog.Duration("duration", time.Second))
		return nil, ErrNoAuth
	}
	sess := &Session{}
	row := sm.dbpool.QueryRow(ctx, `SELECT user_id FROM sessions WHERE id = $1`, sessionCookie.Value)
	err = row.Scan(&sess.UserID)
	if err == sql.ErrNoRows {
		slog.Warn("Проверка сессии: не найдено записей", "error", err.Error())
		return nil, ErrNoAuth
	} else if err != nil {
		slog.Error("Ошибка при проверке сессии:", "error", err.Error())
		return nil, err
	}

	sess.ID = sessionCookie.Value
	return sess, nil
}

// Create создает новую сессию для указанного пользователя и устанавливает соответствующий куки.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - w: HTTP-ответ для установки куки.
//   - user: Интерфейс пользователя, для которого создается сессия.
//   - r: HTTP-запрос для получения информации о клиенте (IP-адрес и User-Agent).
//
// Возвращает ошибку, если она произошла.
func (sm *SessionsDB) Create(ctx context.Context, w http.ResponseWriter, user UserInterface, r *http.Request) error {
	sessID := randutils.RandStringRunes(32)

	ip := getClientIP(r)

	city, location, geoErr := getGeoInfo(ip)
	if geoErr != nil {
		slog.Warn("Не удалось получить геоинформацию:", "error", geoErr)
		city, location = "Неизвестно", "Неизвестно"
	}

	userAgent := r.Header.Get("User-Agent")
	browser, operatingSystem := parseUserAgent(userAgent)

	_, dbErr := sm.dbpool.Exec(ctx,
		"INSERT INTO sessions(id, user_id, ip, browser, operating_system, city, location) VALUES($1, $2, $3, $4, $5, $6, $7)",
		sessID, user.GetID(), ip, browser, operatingSystem, city, location,
	)
	if dbErr != nil {
		slog.Error("Ошибка при создании сессии:", "error", dbErr)
		return dbErr
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   sessID,
		Expires: time.Now().Add(90 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	return nil
}

// DestroyCurrent удаляет текущую активную сессию пользователя из базы данных и аннулирует cookie идентификатор сессии.
//
// Параметры:
//
//	ctx - Контекст запроса, используемый для передачи контекста выполнения операции.
//	w - Объект ResponseWriter для отправки HTTP-ответа клиенту.
//	r - Запрос от клиента, содержащий данные о текущей сессии.
//
// Возвращаемые значения:
//
//	Ошибка, если возникла проблема во время удаления сессии или обработки запроса.
//
// Функция проверяет наличие активной сессии пользователя и, если такая существует, удаляет её запись из таблицы `sessions` в базе данных.
// После успешного удаления записи устанавливает просроченный срок действия cookies 'session_id', тем самым автоматически очищая клиентские куки браузера.
// Если возникают проблемы с выполнением SQL-запроса или обработкой запроса, возвращается соответствующая ошибка.
//
// Пример использования:
//
//	err := sm.DestroyCurrent(context.Background(), responseWriter, request)
//	if err != nil {
//	    log.Println(err)
//	}
func (sm *SessionsDB) DestroyCurrent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err == nil {
		// Удаляем текущую сессию из базы данных
		if _, execErr := sm.dbpool.Exec(ctx,
			"DELETE FROM sessions WHERE id = $1", sess.ID); execErr != nil {
			slog.Error("Не удалось удалить текущую сессию:", "error", execErr)
			return fmt.Errorf("не удалось удалить сессию: %w", execErr)
		}
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Expires: time.Now().AddDate(0, 0, -1), // Устанавливаем истекший срок действия куки
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}

// DestroyAll уничтожает все сессии для указанного пользователя.
// Параметры:
//   - ctx: Контекст выполнения запроса.
//   - w: HTTP-ответ (не используется в данной функции).
//   - user: Интерфейс пользователя для удаления всех его сессий.
//
// Возвращает ошибку при неудаче.
func (sm *SessionsDB) DestroyAll(ctx context.Context, w http.ResponseWriter, user UserInterface) error {
	result, err := sm.dbpool.Exec(ctx,
		"DELETE FROM sessions WHERE user_id = $1", user.GetID())
	if err != nil {
		slog.Error("Ошибка при удалении всех сессий:", "error", err)
		return err
	}

	slog.Info("Уничтожены все сессии", "count", result.RowsAffected(), "для пользователя", user.GetID())

	return nil
}
