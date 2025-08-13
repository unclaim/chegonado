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

	// Убедись, что путь к твоему пакету с конфигом правильный.
	cfg "github.com/unclaim/chegonado.git/internal/shared/config"
)

// localBasePath — это корневая папка, куда будут скачиваться файлы.
const localBasePath = "uploads"

// S3Config содержит параметры для подключения к S3.
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// NewS3Client создает и возвращает новый S3-клиент.
func NewS3Client(ctx context.Context, cfg *S3Config) (*s3.Client, error) {
	// Создаем провайдер статических учетных данных.
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	// Загружаем конфигурацию, используя провайдер учетных данных.
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(creds),
		config.WithRegion("us-east-1"), // Регион можно изменить, если нужно
		// Использование актуального WithEndpointResolver.
		// Теперь функция-резолвер принимает service и region.
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.Endpoint,
				SigningRegion: "us-east-1",
				Source:        aws.EndpointSourceCustom,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфигурацию AWS: %w", err)
	}

	// Создаем новый S3-клиент с настройкой стиля пути.
	return s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

// downloadFileFromS3 скачивает один файл из S3 по его ключу.
func downloadFileFromS3(ctx context.Context, downloader *manager.Downloader, bucket, key string) error {
	filePath := filepath.Join(localBasePath, key)
	fileDir := filepath.Dir(filePath)

	// Создаём все необходимые директории. os.MkdirAll безопасен, если директории уже существуют.
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории %s: %w", fileDir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии файла %s: %v", filePath, closeErr)
		}
	}()

	// Скачиваем файл с помощью Downloader.
	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("не удалось скачать файл %s: %w", key, err)
	}

	log.Printf("Файл скачан: %s", filePath)
	return nil
}

func main() {
	ctx := context.Background()

	// Загрузка конфигурации из YAML-файла.
	configPath := "../../configs/config.yaml"
	appCfg, err := cfg.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации из %s: %v", configPath, err)
	}

	// Подготовка S3-конфигурации из загруженных данных.
	s3cfg := &S3Config{
		Endpoint:        appCfg.FileStorage.S3.Endpoint,
		AccessKeyID:     appCfg.FileStorage.S3.AccessKeyID,
		SecretAccessKey: appCfg.FileStorage.S3.SecretAccessKey,
		Bucket:          appCfg.FileStorage.S3.Bucket,
	}

	// Создание S3-клиента.
	s3Client, err := NewS3Client(ctx, s3cfg)
	if err != nil {
		log.Fatalf("Не удалось создать S3-клиент: %v", err)
	}

	downloader := manager.NewDownloader(s3Client)
	log.Println("Начинаю скачивание файлов из S3...")

	// Используем пагинатор для итерации по всем объектам в бакете.
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3cfg.Bucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("Ошибка при получении страницы списка файлов: %v", err)
		}

		if page == nil {
			continue // Пропускаем пустые страницы, если таковые есть
		}

		for _, object := range page.Contents {
			// Проверяем, что ключ объекта существует.
			if object.Key == nil {
				continue
			}

			key := *object.Key

			// Пропускаем "директории" (объекты, заканчивающиеся на '/')
			if object.Size != nil && *object.Size == 0 && key[len(key)-1] == '/' {
				log.Printf("Пропускаю 'директорию': %s", key)
				continue
			}

			log.Printf("Скачиваю файл: %s", key)

			if err := downloadFileFromS3(ctx, downloader, s3cfg.Bucket, key); err != nil {
				// Логируем ошибку, но продолжаем скачивать остальные файлы.
				log.Printf("Ошибка при скачивании файла %s: %v", key, err)
			}
		}
	}

	log.Println("Скачивание файлов завершено.")
}
