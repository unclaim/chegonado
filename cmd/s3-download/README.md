Отличная идея! Наличие утилиты для переноса данных из S3 обратно в локальное хранилище — это очень полезно. Она может пригодиться для локальной разработки, создания резервных копий или миграции на другую систему.
Я подготовил полный код для такой утилиты. Вы можете сохранить её в отдельный файл, например, cmd/s3-download/main.go, и запускать по необходимости.
Описание решения
Подключение к S3: Утилита подключится к вашему S3-бакету, используя те же параметры, что и основное приложение.
Получение списка файлов: Программа запросит у S3 список всех объектов (файлов) в вашем бакете.
Создание локальной структуры: Для каждого файла из S3 будет создана соответствующая папка на локальном диске, если она ещё не существует.
Скачивание файлов: Каждый файл будет последовательно скачиваться из S3 и сохраняться в нужную локальную папку.
Логирование: Программа будет выводить в консоль информацию о том, какой файл скачивается, чтобы вы могли следить за процессом.
Полный код утилиты для скачивания из S3
Создайте новый файл cmd/s3-download/main.go и вставьте в него этот код.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/unclaim/chegonado.git/internal/shared/config"
)

const (
	localBasePath = "uploads" // Корневая папка для сохранения файлов
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

// downloadFileFromS3 скачивает один файл из S3.
func downloadFileFromS3(ctx context.Context, downloader *s3manager.Downloader, bucket, key string) error {
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
	defer file.Close()

	_, err = downloader.DownloadWithContext(ctx, file,
		&s3.GetObjectInput{
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
	cfg, err := config.LoadConfig("../../configs/config.yaml")
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
	s3Client, err := NewS3Client(s3cfg)
	if err != nil {
		log.Fatalf("Не удалось создать S3-клиент: %v", err)
	}
	
	// Используем Downloader для эффективного скачивания файлов
	downloader := s3manager.NewDownloaderWithClient(s3Client)

	log.Println("Начинаю скачивание файлов из S3...")

	// Получаем список всех объектов в бакете
	err = s3Client.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3cfg.Bucket),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, object := range page.Contents {
			key := aws.StringValue(object.Key)
			log.Printf("Скачиваю файл: %s", key)
			
			if err := downloadFileFromS3(ctx, downloader, s3cfg.Bucket, key); err != nil {
				log.Printf("Ошибка при скачивании файла %s: %v", key, err)
			}
		}
		return true // Продолжаем итерацию
	})

	if err != nil {
		log.Fatalf("Ошибка при получении списка файлов из S3: %v", err)
	}

	log.Println("Скачивание файлов завершено.")
}


Как использовать
Сохраните код: Создайте папку cmd/s3-download и файл main.go внутри неё. Вставьте туда код выше.
Настройте .env: Убедитесь, что ваш файл .env содержит правильные параметры для S3 (S3_ENDPOINT, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY и S3_BUCKET).
Запустите утилиту: Откройте терминал в корневой папке вашего проекта и выполните команду:
go run cmd/s3-download/main.go


Проверка: После завершения работы утилиты в корневой папке проекта появится директория uploads, в которой будет воссоздана вся структура файлов из S3.
Эта утилита — простой и эффективный способ получить локальную копию всех ваших файлов из S3. Если у вас возникнут вопросы или потребуется добавить в неё какие-то дополнительные функции, просто скажите.
