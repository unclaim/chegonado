package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/unclaim/chegonado/internal/filestorage/domain"
	"github.com/unclaim/chegonado/internal/shared/common_errors"
	"github.com/unclaim/chegonado/internal/shared/utils"
	"github.com/unclaim/chegonado/pkg/security/session"
)

// FileStorageHandler содержит сервис для работы с файлами.
type FileStorageHandler struct {
	service domain.FileStorageService
}

// NewFileStorageHandler создает новый экземпляр FileStorageHandler.
func NewFileStorageHandler(service domain.FileStorageService) *FileStorageHandler {
	return &FileStorageHandler{service: service}
}

// UploadAvatarHandler обрабатывает загрузку аватара.
func (h *FileStorageHandler) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("отсутствует авторизация: %v", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	file, header, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			common_errors.NewAppError(w, r, fmt.Errorf("файл аватара не предоставлен"), http.StatusBadRequest)
			return
		}
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении файла аватара: %w", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			slog.Error("Ошибка закрытия файла: %v", slog.Any("error", cerr))
		}
	}()

	avatarURL, err := h.service.UploadAvatar(ctx, userID, file, header.Filename, header.Size)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при загрузке аватара: %w", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message":   "Аватар успешно загружен",
		"avatarURL": avatarURL,
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// GetAvatarHandler получает URL аватара пользователя.
func (h *FileStorageHandler) GetAvatarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить сессию: %v", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	avatarURL, err := h.service.GetAvatarURL(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить URL аватара: %v", err), http.StatusInternalServerError)
		return
	}
	if avatarURL == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("аватар не найден"), http.StatusNotFound)
		return
	}

	response := struct {
		StatusCode int    `json:"statusCode"`
		Body       string `json:"body"`
	}{
		StatusCode: http.StatusOK,
		Body:       avatarURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка: %v", err), http.StatusInternalServerError)
		return
	}
}

// DeleteAvatarHandler удаляет аватар пользователя.
func (h *FileStorageHandler) DeleteAvatarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить сессию: %v", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	if err := h.service.DeleteAvatar(ctx, userID); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось удалить аватар: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Аватар успешно удален"}`))
}
