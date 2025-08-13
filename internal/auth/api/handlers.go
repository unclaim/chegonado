package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/unclaim/chegonado.git/internal/auth/domain"
	"github.com/unclaim/chegonado.git/internal/shared/common_errors"
	"github.com/unclaim/chegonado.git/internal/shared/utils"
	u "github.com/unclaim/chegonado.git/internal/users/domain"
	"github.com/unclaim/chegonado.git/pkg/security/session"
)

var ErrInvalidCredentials = errors.New("неверные учетные данные")
var ErrUserAlreadyExists = errors.New("пользователь с указанным именем или e-mail уже существует")
var ErrAllFieldsRequired = errors.New("все поля обязательны для заполнения")
var ErrAlreadyLoggedIn = errors.New("вы уже вошли в систему")

// AuthHandler теперь зависит от интерфейса domain.AuthServicePort.
type AuthHandler struct {
	AuthService domain.AuthServicePort
}

// NewAuthHandler создает новый экземпляр AuthHandler.
func NewAuthHandler(authService domain.AuthServicePort) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

// @Summary Запрос на сброс пароля
// @Description Отправляет на указанный email ссылку для сброса пароля.
// @Tags Пароли
// @Accept json
// @Produce json
// @Param email body domain.LoginEmailRequest true "Email пользователя для сброса пароля"
// @Success 200 {object} utils.Response "Ссылка успешно отправлена"
// @Router /auth/password/reset [post]
func (ah *AuthHandler) SendPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("поле 'email' обязательно для заполнения"), http.StatusBadRequest)
		return
	}
	err := ah.AuthService.SendPasswordResetService(r.Context(), req.Email)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при отправке ссылки для сброса пароля: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Ссылка для сброса пароля успешно отправлена на указанный email. Проверьте почту и следуйте инструкциям в письме."}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя с логином и паролем.
// @Tags Регистрация
// @Accept json
// @Produce json
// @Param signupRequest body domain.SignUpRequest true "Данные для регистрации"
// @Success 200 {object} utils.Response "Успешная регистрация"
// @Router /auth/signup [post]
func (ah *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {

	var signupRequest domain.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&signupRequest); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный формат JSON: %v", err), http.StatusBadRequest)
		return
	}

	newUser, err := ah.AuthService.Signup(r.Context(), signupRequest, w, r)
	if err != nil {
		if errors.Is(err, ErrAllFieldsRequired) {
			common_errors.NewAppError(w, r, err, http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrUserAlreadyExists) {
			common_errors.NewAppError(w, r, err, http.StatusConflict)
			return
		}
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка регистрации: %w", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Пользователь успешно зарегистрирован",
		"user":    newUser,
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Выход из системы
// @Description Деавторизует текущего пользователя, удаляя сессию.
// @Tags Аутентификация
// @Produce json
// @Success 200 {object} utils.Response "Успешный выход"
// @Router /auth/signout [post]
func (ah *AuthHandler) Signout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := ah.AuthService.Signout(ctx, w, r)
	if err != nil {
		if errors.Is(err, ErrAlreadyLoggedIn) {
			common_errors.NewAppError(w, r, err, http.StatusUnauthorized)
			return
		}
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка выхода из системы: %w", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Вы успешно вышли из системы"}
	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Авторизация пользователя
// @Description Проводит авторизацию пользователя по логину и паролю.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Данные для входа"
// @Success 200 {object} utils.Response "Успешный вход"
// @Router /user/login [post]
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при чтении данных запроса: %w", err), http.StatusBadRequest)
		return
	}

	loggedInUser, err := ah.AuthService.Login(r.Context(), req.Username, req.Password, w, r)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			common_errors.NewAppError(w, r, ErrInvalidCredentials, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ErrAllFieldsRequired) {
			common_errors.NewAppError(w, r, err, http.StatusBadRequest)
			return
		}
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка входа: %w", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"exists": true,
		"user":   loggedInUser,
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Проверка сессии
// @Description Проверяет наличие и валидность сессии пользователя.
// @Tags Сессии
// @Produce json
// @Success 200 {object} utils.Response "Сессия активна"
// @Router /auth/session [get]
func (ah *AuthHandler) CheckSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("cookie сессии не найдена или повреждена: %v", err), http.StatusUnauthorized)
		return
	}

	user, err := ah.AuthService.CheckSessionService(ctx, sessionCookie.Value)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при проверке сессии: %v", err), http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"isAuthenticated": true,
		"user":            user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сериализации ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Получить активные сессии
// @Description Возвращает список всех активных сессий текущего пользователя.
// @Tags Сессии
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response "Список сессий"
// @Router /account/sessions [get]
func (ah *AuthHandler) GetActiveUserSessionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("необходимо авторизоваться: %v", err), http.StatusUnauthorized)
		return
	}

	userID := sess.UserID

	sessions, err := ah.AuthService.GetActiveUserSessionsService(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессий: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println(sessions)
	response := map[string]interface{}{
		"sessions": sessions,
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Отменить сессию
// @Description Аннулирует указанную сессию пользователя по ID.
// @Tags Сессии
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response "Сессия успешно отменена"
// @Router /auth/sessions/revoke [post]
func (ah *AuthHandler) RevokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("необходимо авторизоваться: %v", err), http.StatusUnauthorized)
		return
	}

	var req domain.SessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при разборе запроса: %v", err), http.StatusBadRequest)
		return
	}

	err = ah.AuthService.RevokeSessionService(ctx, req.SessionID, sess.UserID)
	if err != nil {
		if err.Error() == "сеанс не найден или уже аннулирован" {
			common_errors.NewAppError(w, r, err, http.StatusNotFound)
			return
		}
		common_errors.NewAppError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Сеанс успешно отменён",
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Сброс пароля
// @Description Устанавливает новый пароль, используя токен сброса.
// @Tags Пароли
// @Accept json
// @Produce json
// @Success 200 {string} string "Пароль успешно обновлён"
// @Router /auth/password/reset/confirm [post]
func (ah *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный формат запроса: %v", err), http.StatusBadRequest)
		return
	}

	if err := ah.AuthService.ResetPasswordService(r.Context(), req.Token, req.Email, req.NewPassword); err != nil {
		common_errors.NewAppError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Пароль успешно обновлён")); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка записи ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Получить страницу сброса пароля
// @Description Возвращает информацию для страницы сброса пароля, используя токен.
// @Tags Пароли
// @Produce json
// @Param token query string true "Токен сброса пароля"
// @Success 200 {object} map[string]string "Информация о сбросе"
// @Router /auth/password/reset/page [get]
func (ah *AuthHandler) ResetPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("токен не найден"), http.StatusBadRequest)
		return
	}

	response, err := ah.AuthService.ResetPasswordPageService(token)
	if err != nil {
		common_errors.NewAppError(w, r, err, http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сериализации ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Изменить пароль
// @Description Позволяет авторизованному пользователю изменить свой пароль.
// @Tags Пароли
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response "Пароль успешно изменен"
// @Router /auth/password [post]
func (ah *AuthHandler) PasswordHandler(w http.ResponseWriter, r *http.Request) {

	var request u.PasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка декодирования данных: %v", err), http.StatusBadRequest)
		return
	}

	if err := ah.AuthService.PasswordService(r.Context(), w, r, request); err != nil {
		common_errors.NewAppError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Пароль успешно изменен",
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Отправить код для входа
// @Description Отправляет 6-значный код на email для входа в систему без пароля.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param email body domain.LoginEmailRequest true "Email пользователя"
// @Success 200 {object} utils.Response "Код отправлен"
// @Router /auth/login/send-code [post]
func (ah *AuthHandler) SendEmailCodeForLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	if err := ah.AuthService.LoginWithEmailCodeService(r.Context(), req.Email); err != nil {
		common_errors.NewAppError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Код для входа отправлен на ваш email."}
	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Проверить код и войти
// @Description Проверяет 6-значный код, присланный на email, и авторизует пользователя.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.VerifyEmailCodeRequest true "Email и код"
// @Success 200 {object} utils.Response "Вход выполнен успешно"
// @Router /auth/login/verify-code [post]
func (ah *AuthHandler) VerifyEmailCodeForLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.VerifyEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	user, err := ah.AuthService.VerifyLoginCodeService(r.Context(), req.Email, req.Code, w, r)
	if err != nil {
		common_errors.NewAppError(w, r, err, http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "Вход выполнен успешно",
		"user":    user,
	}
	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Отправить код для регистрации
// @Description Отправляет 6-значный код на email для упрощенной регистрации.
// @Tags Регистрация
// @Accept json
// @Produce json
// @Param email body domain.SignupEmailRequest true "Email нового пользователя"
// @Success 200 {object} utils.Response "Код отправлен"
// @Router /auth/signup/send-code [post]
func (ah *AuthHandler) SendEmailCodeForSignupHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.SignupEmailRequest
	// Декодируем JSON-запрос из тела
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	// Вызываем сервис для отправки кода
	if err := ah.AuthService.SignupWithEmailCodeService(r.Context(), req.Email); err != nil {
		common_errors.NewAppError(w, r, err, http.StatusConflict) // Общая ошибка для конфликтов
		return
	}

	// Отправляем успешный ответ
	response := map[string]string{"message": "Код для регистрации отправлен на ваш email."}
	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Проверить код и зарегистрировать пользователя
// @Description Проверяет 6-значный код и автоматически создает аккаунт.
// @Tags Регистрация
// @Accept json
// @Produce json
// @Param request body domain.VerifyEmailCodeRequest true "Email и код"
// @Success 200 {object} utils.Response "Регистрация прошла успешно"
// @Router /auth/signup/verify-code [post]
func (ah *AuthHandler) VerifyEmailCodeForSignupHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.VerifyEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	user, err := ah.AuthService.VerifySignupCodeService(r.Context(), req.Email, req.Code, w, r)
	if err != nil {
		common_errors.NewAppError(w, r, err, http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "Регистрация прошла успешно",
		"user":    user,
	}
	utils.NewResponse(w, http.StatusOK, response)
}

// @Summary Повторная отправка кода
// @Description Отправляет новый 6-значный код на email для верификации.
// @Tags Верификация
// @Accept json
// @Produce json
// @Param email body domain.ResendCodeRequest true "Email пользователя"
// @Success 200 {object} utils.Response "Новый код отправлен"
// @Router /auth/resend-code [post]
func (ah *AuthHandler) ResendVerificationCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.ResendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %v", err), http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("поле 'email' обязательно для заполнения"), http.StatusBadRequest)
		return
	}

	if err := ah.AuthService.ResendVerificationCodeService(r.Context(), req.Email); err != nil {
		common_errors.NewAppError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Новый код верификации успешно отправлен на указанный email."}
	utils.NewResponse(w, http.StatusOK, response)
}
