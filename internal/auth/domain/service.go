package domain

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/unclaim/chegonado.git/internal/auth"
	"github.com/unclaim/chegonado.git/internal/shared/common_errors"
	"github.com/unclaim/chegonado.git/internal/shared/config"
	"github.com/unclaim/chegonado.git/internal/shared/ports"
	"github.com/unclaim/chegonado.git/internal/shared/utils"
	"github.com/unclaim/chegonado.git/internal/users/domain"
	"github.com/unclaim/chegonado.git/pkg/security/session"
)

// AuthService реализует интерфейс domain.AuthService.
type AuthService struct {
	AuthRepo    AuthRepository
	Sessions    session.SessionManager
	EmailSender ports.EmailSender
	Config      config.AppConfig
	bus         EventBus
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(repo AuthRepository, sessions session.SessionManager, emailSender ports.EmailSender, config config.AppConfig, bus EventBus) *AuthService {
	return &AuthService{
		AuthRepo:    repo,
		Sessions:    sessions,
		EmailSender: emailSender,
		Config:      config,
		bus:         bus,
	}
}

// SendVerificationCodeService - отправляет код верификации на email.
func (s *AuthService) SendVerificationCodeService(ctx context.Context, emailUser string) error {
	code, err := utils.GenerateSecureCode()
	if err != nil {
		return common_errors.WrapServiceError("не удалось сгенерировать код верификации", err)
	}
	expiresAt := time.Now().Add(6 * time.Minute)

	if err := s.AuthRepo.SaveVerificationCode(ctx, emailUser, int64(code), expiresAt); err != nil {
		return common_errors.WrapServiceError("ошибка сохранения кода верификации", err)
	}

	templatePath := "../../web/templates/emails/verification_email.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return common_errors.WrapServiceError("не удалось загрузить шаблон письма", err)
	}

	var body bytes.Buffer
	data := struct {
		Code int
	}{
		Code: code,
	}
	if err = tmpl.Execute(&body, data); err != nil {
		return common_errors.WrapServiceError("ошибка при подготовке тела письма", err)
	}

	go func() {
		if err := s.EmailSender.SendEmail(emailUser, "Код верификации", &body); err != nil {
			slog.Error("Не удалось отправить код верификации", "ошибка", err)
		}
	}()

	return nil
}

// Signup регистрирует нового пользователя.
func (s *AuthService) Signup(ctx context.Context, req SignUpRequest, w http.ResponseWriter, r *http.Request) (*domain.User, error) {
	if req.FirstName == "" || req.LastName == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, common_errors.NewServiceError("Все поля обязательны для заполнения")
	}

	u, err := s.AuthRepo.CreateUser(ctx,
		req.FirstName,
		req.LastName,
		req.Username,
		req.Email,
		req.Password,
	)
	if err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании пользователя", err)
	}

	if err := s.Sessions.Create(ctx, w, u, r); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании сессии", err)
	}

	return u, nil
}

// Login аутентифицирует пользователя и создает сессию.
func (s *AuthService) Login(ctx context.Context, login string, password string, w http.ResponseWriter, r *http.Request) (*domain.User, error) {
	if login == "" || password == "" {
		return nil, common_errors.NewServiceError("Поля логина и пароля обязательны для заполнения")
	}

	u, err := s.AuthRepo.CheckPasswordByLoginOrEmail(ctx, login, login, password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common_errors.NewServiceError("Неверные учетные данные")
		}
		return nil, common_errors.WrapServiceError("ошибка при проверке учётных данных", err)
	}

	if err := s.Sessions.Create(ctx, w, u, r); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании сессии", err)
	}

	return u, nil
}

// Signout завершает текущую сессию пользователя.
func (s *AuthService) Signout(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	_, err := session.SessionFromContext(ctx)
	if err != nil {
		return common_errors.NewServiceError("Вы не авторизованы")
	}

	err = s.Sessions.DestroyCurrent(ctx, w, r)
	if err != nil {
		return common_errors.WrapServiceError("ошибка завершения сессии", err)
	}

	return nil
}

// SendPasswordResetService отправляет email для сброса пароля.
func (s *AuthService) SendPasswordResetService(ctx context.Context, emailUser string) error {
	user, err := s.AuthRepo.FindUserByEmail(ctx, emailUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common_errors.NewServiceError(fmt.Sprintf("Пользователь с email %s не найден", emailUser))
		}
		return common_errors.WrapServiceError("ошибка при получении пользователя", err)
	}

	expirationTime := time.Now().Add(time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   user.Email,
		ExpiresAt: expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.Config.Security.JWTSecret))
	if err != nil {
		return common_errors.WrapServiceError("ошибка при создании токена сброса пароля", err)
	}

	resetLink := fmt.Sprintf("http://localhost:3000/reset?token=%s&auto=true", tokenString)

	templatePath := "../../web/templates/emails/reset_password_template.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return common_errors.WrapServiceError("не удалось загрузить шаблон письма", err)
	}

	var body bytes.Buffer
	data := struct {
		ResetLink string
		Username  string
	}{
		ResetLink: resetLink,
		Username:  user.Email,
	}
	if err = tmpl.Execute(&body, data); err != nil {
		return common_errors.WrapServiceError("ошибка при подготовке тела письма", err)
	}

	go func() {
		if err := s.EmailSender.SendEmail(user.Email, "Сброс пароля", &body); err != nil {
			slog.Error("Не удалось отправить email для сброса пароля", "ошибка", err)
		}
	}()

	return nil
}

// CheckSessionService проверяет сессию по ID.
func (s *AuthService) CheckSessionService(ctx context.Context, sessionID string) (*domain.User, error) {
	userID, err := s.AuthRepo.GetSessByID(ctx, sessionID)
	if err != nil {
		return nil, common_errors.WrapServiceError("не удалось получить ID пользователя по сессии", err)
	}

	user, err := s.AuthRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, common_errors.WrapServiceError("не удалось получить пользователя по ID", err)
	}

	return user, nil
}

// GetActiveUserSessionsService возвращает список активных сессий пользователя.
func (s *AuthService) GetActiveUserSessionsService(ctx context.Context, userID int64) ([]session.Session, error) {
	sessions, err := s.AuthRepo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, common_errors.WrapServiceError("не удалось получить список активных сессий", err)
	}
	return sessions, nil
}

// RevokeSessionService отменяет сессию по ID.
func (s *AuthService) RevokeSessionService(ctx context.Context, sessionID string, userID int64) error {
	rowsAffected, err := s.AuthRepo.RevokeSession(ctx, sessionID, userID)
	if err != nil {
		return common_errors.WrapServiceError("ошибка при отмене сеанса", err)
	}
	if rowsAffected == 0 {
		return common_errors.NewServiceError("Сеанс не найден или уже аннулирован")
	}
	return nil
}

// ResetPasswordService сбрасывает пароль пользователя.
func (s *AuthService) ResetPasswordService(ctx context.Context, token, email, newPassword string) error {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Config.Security.JWTSecret), nil
	})

	if err != nil || claims.ExpiresAt < time.Now().Unix() || claims.Subject != email {
		return common_errors.WrapServiceError("Истёкший или недействительный токен", err)
	}

	if err := s.AuthRepo.UpdatePassword(ctx, email, newPassword); err != nil {
		return common_errors.WrapServiceError("ошибка обновления пароля", err)
	}

	return nil
}

// ResetPasswordPageService содержит логику для проверки токена сброса пароля.
func (s *AuthService) ResetPasswordPageService(token string) (*domain.ResetPasswordResponse, error) {
	claims := &jwt.StandardClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Config.Security.JWTSecret), nil
	})

	if err != nil || parsedToken == nil {
		return nil, common_errors.WrapServiceError("Недействительный токен", err)
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, common_errors.NewServiceError("Истёк срок действия токена")
	}

	response := &domain.ResetPasswordResponse{
		Token: token,
		Email: claims.Subject,
	}

	return response, nil
}

// PasswordService содержит логику для смены пароля и управления сессиями.
func (s *AuthService) PasswordService(ctx context.Context, w http.ResponseWriter, r *http.Request, request domain.PasswordRequest) error {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return common_errors.WrapServiceError("ошибка при получении сессии", err)
	}

	if request.NewPassword1 == "" || request.NewPassword1 != request.NewPassword2 {
		return common_errors.NewServiceError("Новый пароль не задан или не совпал с подтверждением")
	}

	user, err := s.AuthRepo.CheckPasswordByUserID(ctx, sess.UserID, request.OldPassword)
	if err != nil {
		return common_errors.WrapServiceError("ошибка проверки старого пароля", err)
	}

	if err := s.AuthRepo.UpdateThePasswordInTheSettings(ctx, user.ID, request.NewPassword1); err != nil {
		return common_errors.WrapServiceError("ошибка обновления пароля", err)
	}

	user.Version++

	if err := s.Sessions.DestroyAll(ctx, w, user); err != nil {
		return common_errors.WrapServiceError("ошибка удаления старых сессий", err)
	}

	if err := s.Sessions.Create(ctx, w, user, r); err != nil {
		return common_errors.WrapServiceError("ошибка создания новой сессии", err)
	}

	return nil
}

// LoginWithEmailCodeService отправляет 6-значный код для входа.
func (s *AuthService) LoginWithEmailCodeService(ctx context.Context, email string) error {
	user, err := s.AuthRepo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common_errors.NewServiceError(fmt.Sprintf("Пользователь с email %s не найден", email))
		}
		return common_errors.WrapServiceError("ошибка при поиске пользователя", err)
	}
	if err := s.SendVerificationCodeService(ctx, user.Email); err != nil {
		return common_errors.WrapServiceError("ошибка при отправке кода верификации", err)
	}
	return nil
}

// VerifyLoginCodeService проверяет код, присланный пользователем, и удаляет его после входа.
func (s *AuthService) VerifyLoginCodeService(ctx context.Context, email string, code string, w http.ResponseWriter, r *http.Request) (*domain.User, error) {
	result, err := s.AuthRepo.ReadVerificationCode(ctx, email)
	if err != nil {
		return nil, common_errors.WrapServiceError("не удалось прочитать код верификации", err)
	}

	if time.Now().After(result.ExpiresAt) {
		return nil, common_errors.NewServiceError("Код верификации истёк")
	}

	codeInt, err := strconv.Atoi(code)
	if err != nil {
		return nil, common_errors.WrapServiceError("Некорректный формат кода верификации", err)
	}
	if int64(codeInt) != result.Code {
		return nil, common_errors.NewServiceError("Код верификации неверный")
	}

	user, err := s.AuthRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, common_errors.WrapServiceError("пользователь не найден", err)
	}
	if err := s.Sessions.Create(ctx, w, user, r); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании сессии", err)
	}
	if err := s.AuthRepo.DeleteVerificationCode(ctx, email); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при удалении кода верификации", err)
	}
	return user, nil
}

// SignupWithEmailCodeService отправляет 6-значный код для упрощенной регистрации.
func (s *AuthService) SignupWithEmailCodeService(ctx context.Context, email string) error {
	// Попробуй найти пользователя по email
	_, err := s.AuthRepo.UserByEmail(ctx, email)
	// --- Добавляем отладочную строчку здесь! ---
	fmt.Printf("В SignupWithEmailCodeService для email %s: err = %v\n", email, err)
	// Если err == nil, значит пользователь уже найден в базе данных
	if err == nil {
		return fmt.Errorf("пользователь с email %s уже существует", email)
	}

	// Если err — это НЕ sql.ErrNoRows, значит произошла какая-то другая ошибка
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}

	// Если дошли до сюда, значит err == sql.ErrNoRows (пользователь не найден)
	// Продолжаем отправлять код
	if err := s.SendVerificationCodeService(ctx, email); err != nil {
		return fmt.Errorf("ошибка при отправке кода верификации: %w", err)
	}
	return nil
}

// VerifySignupCodeService проверяет код, автоматически регистрирует пользователя и удаляет код.
func (s *AuthService) VerifySignupCodeService(ctx context.Context, email string, code string, w http.ResponseWriter, r *http.Request) (*domain.User, error) {
	result, err := s.AuthRepo.ReadVerificationCode(ctx, email)
	if err != nil {
		return nil, common_errors.WrapServiceError("не удалось прочитать код верификации", err)
	}
	codeInt, err := strconv.Atoi(code)
	if err != nil {
		return nil, common_errors.WrapServiceError("Некорректный формат кода верификации", err)
	}
	if int64(codeInt) != result.Code {
		return nil, common_errors.NewServiceError("Код верификации неверный")
	}

	u, err := s.AuthRepo.CreateUser(ctx, email, email, email, email, "password")
	if err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании пользователя", err)
	}
	if err := s.Sessions.Create(ctx, w, u, r); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при создании сессии", err)
	}
	if err := s.AuthRepo.DeleteVerificationCode(ctx, email); err != nil {
		return nil, common_errors.WrapServiceError("ошибка при удалении кода верификации", err)
	}
	// Публикация события после успешной регистрации.
	s.bus.Publish(auth.UserRegisteredEvent{
		UserID: u.ID,
		Email:  email,
	})
	return u, nil
}

// ResendVerificationCodeService повторно отправляет код верификации.
func (s *AuthService) ResendVerificationCodeService(ctx context.Context, email string) error {
	_, err := s.AuthRepo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common_errors.NewServiceError(fmt.Sprintf("Пользователь с email %s не найден", email))
		}
		return common_errors.WrapServiceError("ошибка при поиске пользователя", err)
	}
	if err := s.SendVerificationCodeService(ctx, email); err != nil {
		return common_errors.WrapServiceError("ошибка при повторной отправке кода верификации", err)
	}
	return nil
}
