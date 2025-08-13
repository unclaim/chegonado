package domain

import (
	"context"
	"errors"
	"time"
)

var ErrMessageTooLong = errors.New("сообщение слишком длинное")
var ErrInvalidMessageID = errors.New("недопустимый ID сообщения")

// ChatService реализует бизнес-логику для работы с чатом.
type ChatService struct {
	repo ChatRepositoryPort
}

// NewChatService создает новый экземпляр ChatService.
func NewChatService(repo ChatRepositoryPort) *ChatService {
	return &ChatService{repo: repo}
}

// SendMessage отправляет сообщение.
func (s *ChatService) SendMessage(ctx context.Context, senderID int64, req MessageRequest) error {
	if len(req.Content) > 500 {
		return ErrMessageTooLong
	}

	message := Message{
		SenderID:    senderID,
		RecipientID: req.RecipientID,
		Content:     req.Content,
		CreatedAt:   time.Now(),
		IsRead:      false,
	}

	return s.repo.CreateMessage(ctx, message)
}

// GetInboxMessages получает непрочитанные сообщения пользователя.
func (s *ChatService) GetInboxMessages(ctx context.Context, userID int64) ([]Message, error) {
	return s.repo.FindUnreadMessagesByUserID(ctx, userID)
}

// GetArchiveMessages получает прочитанные (архивные) сообщения пользователя.
func (s *ChatService) GetArchiveMessages(ctx context.Context, userID int64) ([]Message, error) {
	return s.repo.FindReadMessagesByUserID(ctx, userID)
}

// MarkMessageAsRead помечает сообщение как прочитанное.
func (s *ChatService) MarkMessageAsRead(ctx context.Context, messageID int64) error {
	if messageID <= 0 {
		return ErrInvalidMessageID
	}
	return s.repo.UpdateMessageAsRead(ctx, messageID)
}
