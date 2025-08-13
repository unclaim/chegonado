package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/unclaim/chegonado/internal/chat/domain"
)

// ChatRepository реализует интерфейс ChatRepositoryPort.
type ChatRepository struct {
	db *pgxpool.Pool
}

// NewChatRepository создает новый экземпляр ChatRepository.
func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateMessage сохраняет новое сообщение в базе данных.
func (r *ChatRepository) CreateMessage(ctx context.Context, message domain.Message) error {
	// Проверка существования получателя
	var recipientExists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", message.RecipientID).Scan(&recipientExists)
	if err != nil {
		return fmt.Errorf("ошибка проверки получателя: %w", err)
	}
	if !recipientExists {
		return fmt.Errorf("получатель с id %d не найден", message.RecipientID)
	}

	// Проверка существования отправителя (по желанию)
	var senderExists bool
	err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", message.SenderID).Scan(&senderExists)
	if err != nil {
		return fmt.Errorf("ошибка проверки отправителя: %w", err)
	}
	if !senderExists {
		return fmt.Errorf("отправитель с id %d не найден", message.SenderID)
	}

	insertQuery := `INSERT INTO messages (sender_id, recipient_id, content) VALUES ($1, $2, $3)`
	_, err = r.db.Exec(ctx, insertQuery, message.SenderID, message.RecipientID, message.Content)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении сообщения: %w", err)
	}
	return nil
}

// FindUnreadMessagesByUserID находит непрочитанные сообщения для пользователя.
func (r *ChatRepository) FindUnreadMessagesByUserID(ctx context.Context, userID int64) ([]domain.Message, error) {
	query := `SELECT m.id, m.sender_id, m.content, m.created_at, m.is_read, u.id, u.username, u.avatar_url 
	FROM messages m 
	JOIN users u ON m.sender_id = u.id 
	WHERE m.recipient_id = $1 AND m.is_read = FALSE 
	ORDER BY m.created_at ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса к базе данных: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var user domain.User
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.IsRead, &user.ID, &user.Username, &user.AvatarURL); err != nil {
			// Логируем ошибку и продолжаем, чтобы не падать на одном "битом" сообщении
			fmt.Printf("ошибка при разборе данных сообщения: %v", err)
			continue
		}
		msg.Sender = user
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при закрытии соединений: %w", err)
	}
	return messages, nil
}

// FindReadMessagesByUserID находит прочитанные (архивные) сообщения для пользователя.
func (r *ChatRepository) FindReadMessagesByUserID(ctx context.Context, userID int64) ([]domain.Message, error) {
	query := `SELECT m.id, m.sender_id, m.content, m.created_at, m.is_read, u.id, u.username, u.avatar_url 
	FROM messages m 
	JOIN users u ON m.sender_id = u.id 
	WHERE m.recipient_id = $1 AND m.is_read = TRUE 
	ORDER BY m.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var user domain.User
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.IsRead, &user.ID, &user.Username, &user.AvatarURL); err != nil {
			fmt.Printf("ошибка при разборе данных сообщения: %v", err)
			continue
		}
		msg.Sender = user
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при закрытии соединений: %w", err)
	}
	return messages, nil
}

// UpdateMessageAsRead обновляет статус сообщения.
func (r *ChatRepository) UpdateMessageAsRead(ctx context.Context, messageID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET is_read = TRUE WHERE id = $1`, messageID)
	if err != nil {
		return fmt.Errorf("ошибка базы данных при пометке сообщения как прочитанного: %w", err)
	}
	return nil
}
