package domain

// LoginEmailRequest представляет структуру запроса для входа по email.
type LoginEmailRequest struct {
	Email string `json:"email"`
}

// SignupEmailRequest представляет структуру запроса для регистрации по email.
type SignupEmailRequest struct {
	Email string `json:"email"`
}

// ResendCodeRequest представляет структуру запроса для повторной отправки кода.
type ResendCodeRequest struct {
	Email string `json:"email"`
}

// VerifyEmailCodeRequest представляет структуру запроса для проверки кода.
type VerifyEmailCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// ResetPasswordRequest представляет структуру запроса для сброса пароля.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

// SessionRequest представляет структуру запроса для отмены сессии.
type SessionRequest struct {
	SessionID string `json:"session_id"`
}

type LoginRequest struct {
	Username string `json:"username,omitzero"`
	Password string `json:"password_hash,omitzero"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name,omitzero"`
	LastName  string `json:"last_name,omitzero"`
	Username  string `json:"username,omitzero"`
	Email     string `json:"email,omitzero"`
	Password  string `json:"password_hash,omitzero"`
}

type AccountRequest struct {
	CSRFToken string  `json:"csrf_token,omitzero"`
	Username  *string `json:"username,omitzero"`
	Email     string  `json:"email,omitzero"`
	NoAds     bool    `json:"no_ads,omitzero"`
}

// SignUpRequest представляет структуру запроса для регистрации нового пользователя.
type SignUpRequest struct {
	FirstName string `json:"first_name,omitzero"`    // Имя пользователя
	LastName  string `json:"last_name,omitzero"`     // Фамилия пользователя
	Username  string `json:"username,omitzero"`      // Имя пользователя для входа
	Email     string `json:"email,omitzero"`         // Электронная почта пользователя
	Password  string `json:"password_hash,omitzero"` // Хэш нового пароля
}
