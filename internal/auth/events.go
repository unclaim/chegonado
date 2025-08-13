package auth

// UserRegisteredEvent — событие, которое публикуется после успешной регистрации пользователя.
type UserRegisteredEvent struct {
	UserID int64
	Email  string
}
