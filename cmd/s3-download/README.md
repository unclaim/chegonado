# Утилита для скачивания файлов из S3-хранилища
Этот проект содержит Go-скрипт, предназначенный для скачивания всех файлов из указанного S3-совместимого бакета в локальную директорию. Скрипт сохраняет исходную структуру папок и использует современные методы для эффективной работы с большими объёмами данных.
Скрипт полностью обновлён для использования AWS SDK for Go (v2).
# Требования
Для запуска утилиты вам потребуется:
Go (версия 1.16 или выше)
Доступ к S3-совместимому хранилищу (например, Amazon S3, MinIO, Yandex Cloud Object Storage и др.)
# Зависимости Go:
go get github.com/unclaim/chegonado.git
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager


# Конфигурация
Скрипт использует файл конфигурации config.yaml, расположенный в ../../configs/config.yaml. Прежде чем запускать скрипт, убедитесь, что все параметры S3 настроены правильно.
Пример config.yaml:
fileStorage:
  s3:
    endpoint: "http://127.0.0.1:9000" # URL вашего S3-совместимого хранилища
    accessKeyID: "your_access_key_id"
    secretAccessKey: "your_secret_access_key"
    bucket: "your-bucket-name"


# Как это работает
Подключение: Скрипт инициализирует подключение к S3-клиенту, используя параметры из config.yaml.
Итерация по объектам: Скрипт использует s3.NewListObjectsV2Paginator для эффективного перебора всех объектов в указанном бакете. Этот метод оптимально обрабатывает большие списки файлов, делая запросы постранично.
Создание локальной структуры: Для каждого файла, найденного в S3, скрипт воссоздаёт соответствующую структуру папок на локальном диске, начиная с базовой директории uploads.
Скачивание: Затем он использует s3.manager.Downloader для скачивания файла. Этот менеджер автоматически управляет загрузкой, разбивая файл на части, что ускоряет процесс и повышает надёжность.
Логирование: Процесс скачивания логируется, предоставляя информацию о каждом обработанном файле.
Запуск:
go run main.go


# Обновление до AWS SDK v2: Ключевые изменения
Этот скрипт был обновлён с AWS SDK v1 до v2. Ниже приведён обзор основных изменений, которые были внесены в код.
Функция / Метод (v1)
Обновление (v2)
Описание
session.NewSession()
config.LoadDefaultConfig()
Инициализация конфигурации стала более модульной. LoadDefaultConfig автоматически загружает настройки из окружения, общих файлов конфигурации и других источников.
s3manager.NewDownloaderWithClient()
manager.NewDownloader()
Менеджер скачивания перенесён в отдельный пакет feature/s3/manager.
s3Client.ListObjectsV2PagesWithContext()
s3.NewListObjectsV2Paginator()
Итерация по объектам теперь выполняется с помощью итератора-пагинатора, что упрощает и оптимизирует работу с большими наборами данных.
aws.String()
aws.String() или &"string_literal"
Хелперы aws.String остаются, но для получения значения из *string рекомендуется использовать aws.ToString().
downloader.DownloadWithContext()
downloader.Download()
Все вызовы теперь принимают context.Context в качестве первого аргумента.

# Примеры обновлённых функций
Вот как выглядят ключевые функции скрипта после обновления:
NewS3Client
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


downloadFileFromS3
func downloadFileFromS3(ctx context.Context, downloader *manager.Downloader, bucket, key string) error {
    // ... (создание файла)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %w", filePath, err)
	}
	defer file.Close()
    
	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("не удалось скачать файл %s: %w", key, err)
	}

	return nil
}
