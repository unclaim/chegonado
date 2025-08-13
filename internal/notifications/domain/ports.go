package domain

import "context"

// NotificationsService — интерфейс для бизнес-логики уведомлений.
type NotificationsService interface {
	// ...
}

// NotificationsRepository — интерфейс для хранения уведомлений.
type NotificationsRepository interface {
	// ...
}

// EmailAdapter — интерфейс для отправки электронной почты.
// Мы будем использовать его в сервисе.
type EmailAdapter interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}
