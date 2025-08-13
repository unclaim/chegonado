# Утилита для миграции аватаров пользователей в S3
Этот проект содержит Go-скрипт для миграции локально сохранённых аватаров пользователей в S3-совместимое хранилище. Скрипт сканирует локальную файловую систему, загружает файлы в S3 и обновляет соответствующие URL-адреса в базе данных PostgreSQL.
Скрипт полностью обновлён для использования AWS SDK for Go (v2), что обеспечивает высокую производительность и надёжность.
# Требования
Для запуска утилиты вам потребуются:
- Go (версия 1.16 или выше)
- PostgreSQL
- Доступ к S3-совместимому хранилищу (например, Amazon S3, MinIO, Yandex Cloud Object Storage и др.)
# Зависимости Go:
```go
go get github.com/jackc/pgx/v4/pgxpool
go get github.com/unclaim/chegonado.git
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
```

# Конфигурация
Скрипт использует файл конфигурации config.yaml, расположенный в ../../configs/config.yaml. Прежде чем запускать скрипт, убедитесь, что все параметры настроены правильно.
# Пример config.yaml:
```yaml
database:
  host: "localhost"
  port: "5432"
  user: "your_db_user"
  password: "your_db_password"
  name: "your_db_name"
  sslmode: "disable" # или "require", в зависимости от настроек
  url: "" # Если указан, переопределяет другие параметры
fileStorage:
  s3:
    endpoint: "http://127.0.0.1:9000" # URL вашего S3-совместимого хранилища
    accessKeyID: "your_access_key_id"
    secretAccessKey: "your_secret_access_key"
    bucket: "your-bucket-name"
```

# Как это работает
Подключение: Скрипт инициализирует подключение к базе данных PostgreSQL и S3-клиенту, используя данные из config.yaml.
Обход директорий: Скрипт начинает обход локальной файловой системы, начиная с директории ../../uploads/users. Он ищет папки, названия которых соответствуют ID пользователей (например, users/99/).
Обработка файлов: Внутри каждой папки пользователя скрипт ищет подпапку avatars и обходит все файлы-аватары внутри неё.
Загрузка в S3: Каждый найденный файл загружается в S3 с помощью метода s3.PutObject.
Обновление базы данных: После успешной загрузки, скрипт формирует полный URL-адрес файла в S3 и выполняет SQL-запрос UPDATE users SET avatar_url = $1 WHERE id = $2 для обновления соответствующей записи в таблице users.
Логирование: Процесс миграции логируется, предоставляя информацию о каждом обработанном файле и его новом URL.
# Запуск:
```go
go run main.go

# Обновление до AWS SDK v2: Ключевые изменения
Этот скрипт был обновлён с AWS SDK v1 до v2. Ниже приведён обзор основных изменений, которые были внесены в код.
Функция / Метод (v1)
Обновление (v2)
Описание
session.NewSession()
config.LoadDefaultConfig()
Инициализация конфигурации стала более модульной. LoadDefaultConfig автоматически загружает настройки из окружения, общих файлов конфигурации и других источников.
s3.PutObjectWithContext()
s3.PutObject()
Все вызовы теперь принимают context.Context в качестве первого аргумента. Это позволяет контролировать время выполнения и отмену операций.
client.Endpoint (непрямой доступ)
client.Options().BaseEndpoint
URL эндпоинта теперь доступен через client.Options(). Внимание: Он возвращает *string, поэтому его нужно разыменовать для использования в форматировании строк.
aws.String() для форматирования
aws.ToString()
Для безопасного получения строкового значения из указателя *string (например, client.Options().BaseEndpoint) рекомендуется использовать aws.ToString(). Это предотвращает ошибки форматирования.

# Примеры обновлённых функций
Вот как выглядят ключевые функции скрипта после обновления:
NewS3Client
```go
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


uploadFileToS3
func uploadFileToS3(ctx context.Context, client *s3.Client, bucket, filePath string, file io.Reader) (string, error) {
	// ... (чтение файла)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &filePath,
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить файл в S3: %w", err)
	}

	// ИСПРАВЛЕНИЕ: Используем aws.ToString() для безопасного получения строкового значения.
	endpoint := aws.ToString(client.Options().BaseEndpoint)
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, filePath), nil
}

```