package domain

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/unclaim/chegonado.git/internal/filestorage/infra"
)

const (
	MaxAvatarSize = 800 * 1024 // 800 Кб
)

var allowedExtensions = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
	"bmp":  true,
	"tiff": true,
	"webp": true,
	"svg":  true,
}

// Service реализует интерфейс FileStorageService.
type Service struct {
	repo FileStorageRepository
	db   *pgxpool.Pool
}

// NewService создает новый экземпляр FileStorageService.
func NewService(repo FileStorageRepository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

// UploadAvatar загружает аватар, валидирует его и обновляет URL в базе данных.
func (s *Service) UploadAvatar(ctx context.Context, userID int64, file io.Reader, filename string, size int64) (string, error) {
	if size > MaxAvatarSize {
		return "", fmt.Errorf("размер аватара превышает допустимый предел в %dКб", MaxAvatarSize/1024)
	}

	ext := s.getFileExtension(filename)
	if ext == "" {
		return "", fmt.Errorf("неподдерживаемое расширение файла")
	}

	avatarPath := fmt.Sprintf("users/%d/avatars/avatar.%s", userID, ext)

	avatarURL, err := s.repo.SaveFile(ctx, avatarPath, file)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке файла аватара: %w", err)
	}

	if err := s.UpdateAvatarURL(ctx, userID, avatarURL); err != nil {
		if delErr := s.repo.DeleteFile(ctx, avatarPath); delErr != nil {
			fmt.Printf("предупреждение: не удалось удалить аватар из хранилища после ошибки в БД: %v\n", delErr)
		}
		return "", fmt.Errorf("ошибка при обновлении URL аватара: %w", err)
	}

	return avatarURL, nil
}

// GetAvatarURL возвращает URL аватара пользователя из базы данных.
func (s *Service) GetAvatarURL(ctx context.Context, userID int64) (string, error) {
	var avatarURL sql.NullString
	err := s.db.QueryRow(ctx, "SELECT avatar_url FROM users WHERE id = $1", userID).Scan(&avatarURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("не удалось получить URL аватара: %w", err)
	}

	if avatarURL.Valid {
		return avatarURL.String, nil
	}
	return "", nil
}

// DeleteAvatar удаляет аватар пользователя из хранилища и БД.
func (s *Service) DeleteAvatar(ctx context.Context, userID int64) error {
	avatarURL, err := s.GetAvatarURL(ctx, userID)
	if err != nil {
		return err
	}

	if avatarURL == "" {
		return nil
	}

	var pathToDelete string
	if s.repo == nil {
		return fmt.Errorf("не инициализирован репозиторий")
	}

	switch s.repo.(type) {
	case *infra.S3Repository:
		parts := strings.Split(avatarURL, "/")
		pathToDelete = strings.Join(parts[len(parts)-4:], "/")
	case *infra.LocalRepository:
		pathToDelete = strings.TrimPrefix(avatarURL, "/")
	default:
		return fmt.Errorf("неизвестный тип репозитория")
	}

	if err := s.repo.DeleteFile(ctx, pathToDelete); err != nil {
		return fmt.Errorf("не удалось удалить файл аватара из хранилища: %w", err)
	}

	if err := s.UpdateAvatarURL(ctx, userID, ""); err != nil {
		return fmt.Errorf("не удалось очистить URL аватара в БД: %w", err)
	}

	return nil
}

// UpdateAvatarURL обновляет URL аватара пользователя в БД.
func (s *Service) UpdateAvatarURL(ctx context.Context, userID int64, url string) error {
	result, err := s.db.Exec(ctx, "UPDATE users SET avatar_url = $1 WHERE id = $2", url, userID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении аватара пользователя с ID %d: %w", userID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("пользователь с ID %d не найден", userID)
	}
	return nil
}

// getFileExtension проверяет расширение файла.
func (s *Service) getFileExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename)[1:])
	if allowedExtensions[ext] {
		return ext
	}
	return ""
}

// CreateDefaultAvatar создает аватар по умолчанию для пользователя.
func (s *Service) CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error) {
	avatarURL, err := s.repo.CreateDefaultAvatar(ctx, email, userID)
	if err != nil {
		return "", fmt.Errorf("не удалось создать аватар по умолчанию: %w", err)
	}

	if err := s.UpdateAvatarURL(ctx, userID, avatarURL); err != nil {
		return "", fmt.Errorf("не удалось обновить URL аватара: %w", err)
	}

	return avatarURL, nil
}
