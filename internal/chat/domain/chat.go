package domain

import (
	"context"
	"time"
)

// Message представляет собой структуру сообщения.
type Message struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"senderId"`
	RecipientID int64     `json:"recipientId"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	IsRead      bool      `json:"isRead"`
	Sender      User      `json:"sender"`
}

// User представляет упрощенную структуру пользователя для отображения в чате.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

// MessageRequest — структура для входящего запроса на отправку сообщения.
type MessageRequest struct {
	RecipientID int64  `json:"recipient_id"`
	Content     string `json:"message"`
}

// MarkAsReadRequest — структура для входящего запроса на пометку сообщения как прочитанного.
type MarkAsReadRequest struct {
	MessageID int64 `json:"messageId"`
}

// ChatServicePort — интерфейс для бизнес-логики.
type ChatServicePort interface {
	SendMessage(ctx context.Context, senderID int64, req MessageRequest) error
	GetInboxMessages(ctx context.Context, userID int64) ([]Message, error)
	GetArchiveMessages(ctx context.Context, userID int64) ([]Message, error)
	MarkMessageAsRead(ctx context.Context, messageID int64) error
}

// ChatRepositoryPort — интерфейс для работы с хранилищем данных.
type ChatRepositoryPort interface {
	CreateMessage(ctx context.Context, message Message) error
	FindUnreadMessagesByUserID(ctx context.Context, userID int64) ([]Message, error)
	FindReadMessagesByUserID(ctx context.Context, userID int64) ([]Message, error)
	UpdateMessageAsRead(ctx context.Context, messageID int64) error
}
