package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/unclaim/chegonado.git/internal/users/domain"
	"github.com/unclaim/chegonado.git/pkg/infrastructure/eventbus"
	"github.com/unclaim/chegonado.git/pkg/security/session"
)

// AuthRepository определяет методы для взаимодействия с хранилищем данных.
type AuthRepository interface {
	CreateUser(ctx context.Context, firstName, lastName, username, email, password string) (*domain.User, error)
	CheckPasswordByLoginOrEmail(ctx context.Context, username, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetSessByID(ctx context.Context, sessionID string) (int64, error)
	GetSessionsByUserID(ctx context.Context, userID int64) ([]session.Session, error)
	RevokeSession(ctx context.Context, sessionID string, userID int64) (int64, error)
	UpdatePassword(ctx context.Context, email, newPassword string) error
	UpdateThePasswordInTheSettings(ctx context.Context, userID int64, newPassword string) error
	CheckPasswordByUserID(ctx context.Context, uid int64, pass string) (*domain.User, error)
	CheckVerifiedByUserID(ctx context.Context, uid int64) (bool, error)
	ReadVerificationCode(ctx context.Context, email string) (*domain.VerifyCode, error)
	Verified(ctx context.Context, email string) error
	SaveVerificationCode(ctx context.Context, email string, code int64, expiresAt time.Time) error
	CleanUpVerificationCodes(ctx context.Context) error
	DeleteVerificationCode(ctx context.Context, email string) error
}

// AuthServicePort определяет методы, которые должны быть реализованы в сервисе аутентификации.
// Этот интерфейс будет использоваться в AuthHandler для инверсии зависимостей.
type AuthServicePort interface {
	SendPasswordResetService(ctx context.Context, emailUser string) error
	Signup(ctx context.Context, req SignUpRequest, w http.ResponseWriter, r *http.Request) (*domain.User, error)
	Login(ctx context.Context, login string, password string, w http.ResponseWriter, r *http.Request) (*domain.User, error)
	Signout(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	CheckSessionService(ctx context.Context, sessionID string) (*domain.User, error)
	GetActiveUserSessionsService(ctx context.Context, userID int64) ([]session.Session, error)
	RevokeSessionService(ctx context.Context, sessionID string, userID int64) error
	ResetPasswordService(ctx context.Context, token, email, newPassword string) error
	ResetPasswordPageService(token string) (*domain.ResetPasswordResponse, error)
	PasswordService(ctx context.Context, w http.ResponseWriter, r *http.Request, request domain.PasswordRequest) error
	LoginWithEmailCodeService(ctx context.Context, email string) error
	VerifyLoginCodeService(ctx context.Context, email string, code string, w http.ResponseWriter, r *http.Request) (*domain.User, error)
	SignupWithEmailCodeService(ctx context.Context, email string) error
	VerifySignupCodeService(ctx context.Context, email string, code string, w http.ResponseWriter, r *http.Request) (*domain.User, error)
	ResendVerificationCodeService(ctx context.Context, email string) error
}

// EventBus — интерфейс для публикации событий.
type EventBus interface {
	Publish(event eventbus.Event)
}
