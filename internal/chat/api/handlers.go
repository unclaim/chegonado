package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/unclaim/chegonado/internal/chat/domain"
	"github.com/unclaim/chegonado/internal/shared/common_errors"
	"github.com/unclaim/chegonado/internal/shared/utils"
	"github.com/unclaim/chegonado/pkg/security/session"
)

// ChatHandler отвечает за обработку HTTP-запросов, связанных с чатом.
type ChatHandler struct {
	chatService *domain.ChatService
}

// NewChatHandler создает новый экземпляр ChatHandler.
func NewChatHandler(service *domain.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: service,
	}
}

// MarkMessagesAsRead помечает сообщение как прочитанное.
func (h *ChatHandler) MarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	var req domain.MarkAsReadRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный формат запроса: %w", err), http.StatusBadRequest)
		return
	}

	err := h.chatService.MarkMessageAsRead(r.Context(), req.MessageID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidMessageID) {
			common_errors.NewAppError(w, r, err, http.StatusBadRequest)
		} else {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка при пометке сообщения как прочитанного: %w", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetArchivedMessages возвращает архивированные сообщения пользователя.
func (h *ChatHandler) GetArchivedMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}

	messages, err := h.chatService.GetArchiveMessages(ctx, sess.UserID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки архивированных сообщений: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, messages)
}

// GetInboxMessages возвращает входящие сообщения пользователя.
func (h *ChatHandler) GetInboxMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}

	messages, err := h.chatService.GetInboxMessages(ctx, sess.UserID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки сообщений: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, messages)
}

// SendMessageHandler отправляет сообщение другому пользователю.
func (h *ChatHandler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("необходимо авторизоваться: %w", err), http.StatusUnauthorized)
		return
	}

	var req domain.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании данных: %w", err), http.StatusBadRequest)
		return
	}

	if sess.UserID == 0 {
		common_errors.NewAppError(w, r, fmt.Errorf("неопределенный пользователь"), http.StatusUnauthorized)
		return
	}
	log.Println(req.RecipientID, sess.UserID)
	err = h.chatService.SendMessage(ctx, sess.UserID, req)
	if err != nil {
		if errors.Is(err, domain.ErrMessageTooLong) {
			common_errors.NewAppError(w, r, err, http.StatusBadRequest)
		} else {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка при отправке сообщения: %w", err), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{"message": "сообщение успешно отправлено"}
	utils.NewResponse(w, http.StatusOK, response)
}
