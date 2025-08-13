package infra

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/unclaim/chegonado.git/internal/shared/utils"
)

const (
	AvatarBasePath = "uploads"
)

// LocalRepository реализует интерфейс domain.FileStorageRepository для локальной ФС.
type LocalRepository struct {
}

// NewLocalRepository создает новый экземпляр LocalRepository.
func NewLocalRepository() *LocalRepository {
	return &LocalRepository{}
}

// SaveFile сохраняет файл на диск и возвращает его относительный URL.
func (r *LocalRepository) SaveFile(ctx context.Context, filePath string, file io.Reader) (string, error) {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			return "", fmt.Errorf("ошибка создания директории %s: %w", dir, mkdirErr)
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла %s: %w", filePath, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			slog.Error("Ошибка закрытия файла: %v", slog.Any("error", cerr))
		}
	}()

	written, copyErr := io.Copy(out, file)
	if copyErr != nil {
		return "", fmt.Errorf("ошибка копирования данных в файл %s (%d байт скопировано): %w", filePath, written, copyErr)
	}

	return "/" + filePath, nil
}

// DeleteFile удаляет файл с диска.
func (r *LocalRepository) DeleteFile(ctx context.Context, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}

// CheckFileExists проверяет, существует ли файл на диске.
func (r *LocalRepository) CheckFileExists(ctx context.Context, filePath string) (bool, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка при проверке существования файла: %w", err)
	}
	return true, nil
}

// CreateDefaultAvatar создает аватар по умолчанию на локальном диске.
func (r *LocalRepository) CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error) {
	const AvatarBasePath = "uploads"
	userDir := filepath.Join(AvatarBasePath, fmt.Sprintf("users/%d/avatars", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка при создании папки для пользователя %d: %w", userID, err)
	}

	avatarPath := filepath.Join(userDir, "default_avatar.png")
	if err := utils.GenerateAvatar(email, avatarPath); err != nil {
		return "", fmt.Errorf("ошибка при генерации аватара для пользователя %d: %w", userID, err)
	}

	return "/" + avatarPath, nil
}
