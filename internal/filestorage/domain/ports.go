package domain

import (
	"context"
	"io"
)

// FileStorageService определяет контракт для бизнес-логики работы с файлами.
type FileStorageService interface {
	UploadAvatar(ctx context.Context, userID int64, file io.Reader, filename string, size int64) (string, error)
	GetAvatarURL(ctx context.Context, userID int64) (string, error)
	DeleteAvatar(ctx context.Context, userID int64) error
	CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error)
}

// FileStorageRepository определяет контракт для взаимодействия с файловым хранилищем.
type FileStorageRepository interface {
	SaveFile(ctx context.Context, filePath string, file io.Reader) (string, error)
	DeleteFile(ctx context.Context, filePath string) error
	CheckFileExists(ctx context.Context, filePath string) (bool, error)
	CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error)
}
