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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/unclaim/chegonado/internal/shared/config"
)

const (
	localBasePath = "../../uploads" // Базовый путь для локальных файлов
)

// S3Config содержит параметры для подключения к S3.
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// NewS3Client создает и возвращает новый S3-клиент.
func NewS3Client(cfg *S3Config) (*s3.S3, error) {
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать AWS сессию: %w", err)
	}

	return s3.New(sess), nil
}

// updateDBLink обновляет URL аватара в базе данных.
func updateDBLink(ctx context.Context, db *pgxpool.Pool, userID int64, newURL string) error {
	_, err := db.Exec(ctx, "UPDATE users SET avatar_url = $1 WHERE id = $2", newURL, userID)
	if err != nil {
		return fmt.Errorf("не удалось обновить URL аватара для пользователя %d: %w", userID, err)
	}
	return nil
}

// uploadFileToS3 загружает файл в S3 и возвращает его полный URL.
func uploadFileToS3(ctx context.Context, client *s3.S3, bucket, filePath string, file io.Reader) (string, error) {
	// Читаем содержимое файла в байтовый слайс
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	// Создаем io.ReadSeeker из байтового слайса
	reader := bytes.NewReader(buf.Bytes())
	_, err = client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
		Body:   reader,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить файл в S3: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", client.Endpoint, bucket, filePath), nil
}

// migrateUserAvatars переносит аватары одного пользователя.
func migrateUserAvatars(ctx context.Context, db *pgxpool.Pool, s3Client *s3.S3, s3Bucket, userFolder string) {
	userIDStr := filepath.Base(userFolder)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Предупреждение: не удалось разобрать ID пользователя из папки '%s', пропускаем.", userFolder)
		return
	}

	avatarsPath := filepath.Join(userFolder, "avatars")
	if _, err := os.Stat(avatarsPath); os.IsNotExist(err) {
		return // Папка с аватарами не найдена, ничего не делаем.
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

		// Копируем файл в буфер, чтобы можно было использовать его дважды
		var buf bytes.Buffer
		tee := io.TeeReader(file, &buf)

		// Используем filepath.Rel, чтобы получить путь относительно папки uploads.
		s3FilePath, err := filepath.Rel(localBasePath, filePath)
		if err != nil {
			return fmt.Errorf("не удалось получить относительный путь: %w", err)
		}
		// Заменяем разделители пути, чтобы они соответствовали формату URL.
		s3FilePath = strings.ReplaceAll(s3FilePath, "\\", "/")

		s3URL, err := uploadFileToS3(ctx, s3Client, s3Bucket, s3FilePath, tee)
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

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключение к базе данных
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

	// Подключение к S3
	s3cfg := &S3Config{
		Endpoint:        cfg.FileStorage.S3.Endpoint,
		AccessKeyID:     cfg.FileStorage.S3.AccessKeyID,
		SecretAccessKey: cfg.FileStorage.S3.SecretAccessKey,
		Bucket:          cfg.FileStorage.S3.Bucket,
	}
	s3Client, err := NewS3Client(s3cfg)
	if err != nil {
		log.Fatalf("Не удалось создать S3-клиент: %v", err)
	}

	// Начинаем миграцию
	log.Println("Начинаю миграцию локальных аватаров в S3...")

	err = filepath.Walk(filepath.Join(localBasePath, "users"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Base(path) != "users" {
			// Это папка пользователя, запускаем миграцию для неё
			migrateUserAvatars(ctx, dbpool, s3Client, s3cfg.Bucket, path)
			return filepath.SkipDir // Пропускаем поддиректории, так как мы их обработали вручную
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Ошибка при обходе директорий: %v", err)
	}

	log.Println("Миграция завершена.")
}
