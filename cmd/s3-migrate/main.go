package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v4/pgxpool"
	cfg "github.com/unclaim/chegonado.git/internal/shared/config"
)

const (
	localBasePath = "../../uploads" // Базовый путь для локальных файлов
)

// S3Config contains parameters for connecting to S3.
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// NewS3Client creates and returns a new S3-client using AWS SDK v2.
func NewS3Client(ctx context.Context, cfg *S3Config) (*s3.Client, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(creds),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(cfg.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить AWS-конфигурацию: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return s3Client, nil
}

// updateDBLink updates the avatar URL in the database.
func updateDBLink(ctx context.Context, db *pgxpool.Pool, userID int64, newURL string) error {
	_, err := db.Exec(ctx, "UPDATE users SET avatar_url = $1 WHERE id = $2", newURL, userID)
	if err != nil {
		return fmt.Errorf("не удалось обновить URL аватара для пользователя %d: %w", userID, err)
	}
	return nil
}

// uploadFileToS3 uploads a file to S3 and returns its full URL using AWS SDK v2.
func uploadFileToS3(ctx context.Context, client *s3.Client, bucket, filePath string, file io.Reader) (string, error) {
	// Read the file content into a byte slice
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &filePath,
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить файл в S3: %w", err)
	}

	// ИСПРАВЛЕНИЕ: Используем aws.ToString() для безопасного получения строкового значения
	// из указателя *string, который возвращает client.Options().BaseEndpoint.
	endpoint := aws.ToString(client.Options().BaseEndpoint)
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, filePath), nil
}

// migrateUserAvatars migrates one user's avatars.
func migrateUserAvatars(ctx context.Context, db *pgxpool.Pool, s3Client *s3.Client, s3Bucket, s3Endpoint, userFolder string) {
	userIDStr := filepath.Base(userFolder)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Предупреждение: не удалось разобрать ID пользователя из папки '%s', пропускаем.", userFolder)
		return
	}

	avatarsPath := filepath.Join(userFolder, "avatars")
	if _, err := os.Stat(avatarsPath); os.IsNotExist(err) {
		return // Avatar folder not found, do nothing.
	}

	err = filepath.Walk(avatarsPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		log.Printf("Обработка файла: %s", filePath)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("не удалось открыть файл: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				// Handle the error appropriately, e.g., log it
				log.Printf("Error closing file: %v", err)
			}
		}()

		// Use filepath.Rel to get the path relative to the localBasePath folder.
		s3FilePath, err := filepath.Rel(localBasePath, filePath)
		if err != nil {
			return fmt.Errorf("не удалось получить относительный путь: %w", err)
		}
		// Replace path separators to match the URL format.
		s3FilePath = strings.ReplaceAll(s3FilePath, "\\", "/")

		s3URL, err := uploadFileToS3(ctx, s3Client, s3Bucket, s3FilePath, file)
		if err != nil {
			return fmt.Errorf("ошибка загрузки файла в S3: %w", err)
		}

		err = updateDBLink(ctx, db, userID, s3URL)
		if err != nil {
			return fmt.Errorf("ошибка обновления БД: %w", err)
		}

		log.Printf("Успешно перенесен аватар пользователя %d: %s -> %s", userID, filePath, s3URL)
		return nil
	})

	if err != nil {
		log.Printf("Ошибка при обработке папки пользователя %d: %v", userID, err)
	}
}

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := cfg.LoadConfig("../../configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Connect to the database
	dbURL := cfg.Database.URL
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)
	}
	dbpool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbpool.Close()

	// Connect to S3
	s3cfg := &S3Config{
		Endpoint:        cfg.FileStorage.S3.Endpoint,
		AccessKeyID:     cfg.FileStorage.S3.AccessKeyID,
		SecretAccessKey: cfg.FileStorage.S3.SecretAccessKey,
		Bucket:          cfg.FileStorage.S3.Bucket,
	}
	s3Client, err := NewS3Client(ctx, s3cfg)
	if err != nil {
		log.Fatalf("Не удалось создать S3-клиент: %v", err)
	}

	// Start migration
	log.Println("Начинаю миграцию локальных аватаров в S3...")

	err = filepath.Walk(filepath.Join(localBasePath, "users"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Base(path) != "users" {
			// This is a user folder, start migration for it
			migrateUserAvatars(ctx, dbpool, s3Client, s3cfg.Bucket, s3cfg.Endpoint, path)
			return filepath.SkipDir // Skip subdirectories as we handled them manually
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Ошибка при обходе директорий: %v", err)
	}

	log.Println("Миграция завершена.")
}
