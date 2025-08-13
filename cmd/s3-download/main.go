package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/unclaim/chegonado.git/internal/shared/config"
)

const (
	localBasePath = "uploads" // Корневая папка для сохранения файлов
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

// downloadFileFromS3 downloads one file from S3 using AWS SDK v2.
func downloadFileFromS3(ctx context.Context, downloader *manager.Downloader, bucket, key string) error {
	filePath := filepath.Join(localBasePath, key)
	fileDir := filepath.Dir(filePath)

	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(fileDir, 0755); mkdirErr != nil {
			return fmt.Errorf("ошибка создания директории %s: %w", fileDir, mkdirErr)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %w", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Ошибка закрытия файла: %v", err)
		}
	}()

	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("не удалось скачать файл %s: %w", key, err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Загружаем конфигурацию
	cfg, err := cfg.LoadConfig("../../configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключение к S3
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

	// Используем Downloader из нового пакета manager
	downloader := manager.NewDownloader(s3Client)

	log.Println("Начинаю скачивание файлов из S3...")

	// Используем Paginator для итерации по объектам
	p := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3cfg.Bucket),
	})

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			log.Fatalf("Ошибка при получении страницы с файлами из S3: %v", err)
		}

		for _, object := range page.Contents {
			key := aws.ToString(object.Key)
			log.Printf("Скачиваю файл: %s", key)

			if err := downloadFileFromS3(ctx, downloader, s3cfg.Bucket, key); err != nil {
				log.Printf("Ошибка при скачивании файла %s: %v", key, err)
			}
		}
	}

	log.Println("Скачивание файлов завершено.")
}
