package infra

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	cfg "github.com/unclaim/chegonado.git/internal/shared/config"
	"github.com/unclaim/chegonado.git/internal/shared/utils"
)

// S3Repository implements the domain.FileStorageRepository interface for S3 storage.
type S3Repository struct {
	s3Client *s3.Client
	bucket   string
	endpoint string
}

// NewS3Repository creates a new S3Repository instance.
func NewS3Repository(cfg *cfg.AppConfig) (*S3Repository, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.FileStorage.S3.AccessKeyID, cfg.FileStorage.S3.SecretAccessKey, "")

	// Используем WithBaseEndpoint для задания статического URL эндпоинта.
	// Это самая новая и правильная функция для этой задачи.
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(cfg.FileStorage.S3.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить AWS-конфигурацию: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Repository{
		s3Client: s3Client,
		bucket:   cfg.FileStorage.S3.Bucket,
		endpoint: cfg.FileStorage.S3.Endpoint,
	}, nil
}

// SaveFile saves a file to S3 and returns its full URL.
func (r *S3Repository) SaveFile(ctx context.Context, filePath string, file io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	_, err = r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &filePath,
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить файл в S3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucket, filePath), nil
}

// DeleteFile deletes a file from S3.
func (r *S3Repository) DeleteFile(ctx context.Context, filePath string) error {
	_, err := r.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucket,
		Key:    &filePath,
	})
	if err != nil {
		return fmt.Errorf("не удалось удалить файл из S3: %w", err)
	}
	return nil
}

// CheckFileExists checks if a file exists in S3.
func (r *S3Repository) CheckFileExists(ctx context.Context, filePath string) (bool, error) {
	_, err := r.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &r.bucket,
		Key:    &filePath,
	})
	if err != nil {
		var notFoundError *types.NotFound
		if errors.As(err, &notFoundError) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка при проверке существования файла в S3: %w", err)
	}
	return true, nil
}

// CreateDefaultAvatar creates a default avatar in S3.
func (r *S3Repository) CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error) {
	tempFile, err := os.CreateTemp("", "default_avatar.png")
	if err != nil {
		return "", fmt.Errorf("не удалось создать временный файл: %w", err)
	}
	tempFileName := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return "", err
	}
	defer func() {
		if err := os.Remove(tempFileName); err != nil {
			log.Printf("Не удалось удалить временный файл %s: %v", tempFileName, err)
		}
	}()

	if err := utils.GenerateAvatar(email, tempFileName); err != nil {
		return "", fmt.Errorf("не удалось сгенерировать аватар: %w", err)
	}

	fileReader, err := os.Open(tempFileName)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть временный файл: %w", err)
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			log.Printf("Не удалось закрыть fileReader: %v", err)
		}
	}()

	avatarPath := fmt.Sprintf("users/%d/avatars/default_avatar.png", userID)

	_, err = r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &r.bucket,
		Key:         &avatarPath,
		Body:        fileReader,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить аватар по умолчанию в S3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucket, avatarPath), nil
}
