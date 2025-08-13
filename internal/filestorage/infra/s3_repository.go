package infra

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/unclaim/chegonado.git/internal/shared/config"
	"github.com/unclaim/chegonado.git/internal/shared/utils"
)

// S3Repository реализует интерфейс domain.FileStorageRepository для S3-хранилища.
type S3Repository struct {
	s3Client *s3.S3
	bucket   string
	endpoint string
}

// NewS3Repository создает новый экземпляр S3Repository.
func NewS3Repository(cfg *config.AppConfig) (*S3Repository, error) {
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.FileStorage.S3.AccessKeyID, cfg.FileStorage.S3.SecretAccessKey, ""),
		Endpoint:         aws.String(cfg.FileStorage.S3.Endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать AWS сессию: %w", err)
	}

	return &S3Repository{
		s3Client: s3.New(sess),
		bucket:   cfg.FileStorage.S3.Bucket,
		endpoint: cfg.FileStorage.S3.Endpoint,
	}, nil
}

// SaveFile сохраняет файл в S3 и возвращает его полный URL.
func (r *S3Repository) SaveFile(ctx context.Context, filePath string, file io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	reader := bytes.NewReader(buf.Bytes())

	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(filePath),
		Body:   reader,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить файл в S3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucket, filePath), nil
}

// DeleteFile удаляет файл из S3.
func (r *S3Repository) DeleteFile(ctx context.Context, filePath string) error {
	_, err := r.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return fmt.Errorf("не удалось удалить файл из S3: %w", err)
	}
	return nil
}

// CheckFileExists проверяет существование файла в S3.
func (r *S3Repository) CheckFileExists(ctx context.Context, filePath string) (bool, error) {
	_, err := r.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return false, nil
		}
		return false, fmt.Errorf("ошибка при проверке существования файла в S3: %w", err)
	}
	return true, nil
}

// CreateDefaultAvatar создает аватар по умолчанию в S3.
func (r *S3Repository) CreateDefaultAvatar(ctx context.Context, email string, userID int64) (string, error) {
	// Создаём временный файл на диске с помощью os.CreateTemp
	tempFile, err := os.CreateTemp("", "default_avatar.png")
	if err != nil {
		return "", fmt.Errorf("не удалось создать временный файл: %w", err)
	}
	tempFileName := tempFile.Name()
	// Закрываем временный файл и проверяем ошибку
	if err := tempFile.Close(); err != nil {
		return "", err
	}
	// Планируем удаление временного файла с обработкой ошибок
	defer func() {
		if err := os.Remove(tempFileName); err != nil {
			log.Printf("Failed to remove temp file %s: %v", tempFileName, err)
		}
	}()

	// Генерируем аватар в этот временный файл
	if err := utils.GenerateAvatar(email, tempFileName); err != nil {
		return "", fmt.Errorf("не удалось сгенерировать аватар: %w", err)
	}

	// Открываем временный файл для чтения
	fileReader, err := os.Open(tempFileName)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть временный файл: %w", err)
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			log.Printf("Failed to close fileReader: %v", err)
		}
	}()

	avatarPath := fmt.Sprintf("users/%d/avatars/default_avatar.png", userID)

	// Загружаем файл в S3
	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(avatarPath),
		Body:        fileReader,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить аватар по умолчанию в S3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucket, avatarPath), nil
}
Вышла новая вторая версия aws переделать код который показан выше адаптировать под новую версию убрать устаревшие методы заменить на новые 
вот документация 
Идти
Перейти к основному содержанию
Поиск пакетов или символов

Зачем ехать
Значок выпадающего меню
Учиться
Документы
Значок выпадающего меню
Пакеты
Сообщество
Значок выпадающего меню
Откройте для себя пакеты
 
github.com/aws/aws-sdk-go-v2
 
AWS

Идти
AWS
пакет


Главный
Подробности
проверенный Действующий файл go.mod 
проверенный Распространяемая лицензия 
проверенный Помеченная версия 
проверенный Стабильная версия 
Узнайте больше о лучших практиках
Хранилище
github.com/aws/aws-sdk-go-v2
Дюны
Логотип Open Source Insights Аналитика из открытых источников

тип HandleFailRefreshCredentialsCacheStrategy
 Документация ¶
Обзор ¶
Утилиты преобразования значений и указателей
HTTP-клиент SDK по умолчанию
Package aws предоставляет основные утилиты SDK и общие типы. Используйте функцию этого пакета утилиты для упрощения настройки и чтения параметров операций API.

Утилиты преобразования значений и указателей ¶
Этот пакет включает вспомогательную утилиту преобразования для каждого скалярного типа пакета SDK Использование API. Эти утилиты делают получение указателя скаляра и разыменование на указатель проще.

Каждая утилита преобразования бывает двух видов. Value to pointer и pointer to value. Указатель на значение безопасно разыменует указатель и вернет его значение. Если указатель был равен нулю, будет возвращено нулевое значение скаляра.

Значение функций указателя будет названо в соответствии со скалярным типом. Так что приобретите *строка из строкового значения используйте функцию "Строка". Это позволяет легко получить указатель на литеральное строковое значение, потому что получение адреса Литерал требует, чтобы сначала присвоить значение переменной.

var strPtr *string

// Without the SDK's conversion functions
str := "my string"
strPtr = &str

// With the SDK's conversion functions
strPtr = aws.String("my string")

// Convert *string to string value
str = aws.ToString(strPtr)
Помимо скаляров, в пакет AWS также входят утилиты преобразования для map и slice для часто используемых типов в параметрах API. Карта и срез Функции преобразования используют шаблон именования, аналогичный скалярному преобразованию Функции.

var strPtrs []*string
var strs []string = []string{"Go", "Gophers", "Go"}

// Convert []string to []*string
strPtrs = aws.StringSlice(strs)

// Convert []*string to []string
strs = aws.ToStringSlice(strPtrs)
HTTP-клиент SDK по умолчанию ¶
SDK будет использовать протокол http. DefaultClient, если HTTP-клиент не предоставлен Сеанс SDK или конструктор клиента сервиса. Это означает, что если метод http. DefaultClient изменяется другими компонентами вашего приложения модификации также будут подхвачены SDK.

В некоторых случаях это может быть намеренно, но лучше создать пользовательский HTTP-клиент для явного обмена через ваше приложение. Вы можете настроить SDK для использования пользовательского клиента HTTP, установив параметр HTTPClient значения типа Config SDK при создании клиента сессии или сервиса.

Package AWS предоставляет основные функции для выполнения запросов к сервисам AWS.

Индекс ¶
Константы
func Bool(v bool) *bool
func BoolMap(vs map[string]bool) map[string]*bool
func BoolSlice(vs []bool) []*bool
func Byte(v байт) *byte
func ByteMap(vs map[string]byte) map[string]*byte
func ByteSlice(vs []byte) []*байт
func Duration(v time. Продолжительность) *время. Длительность
func DurationMap(vs map[string]time. Длительность) map[строка]*время. Длительность
func DurationSlice(vs []time. Продолжительность) []*время. Длительность
func Float32(v float32) *float32
func Float32Map(vs map[string]float32) map[string]*float32
func Float32Slice(vs []float32) []*float32
func Float64(v float64) *float64
func Float64Map(vs map[string]float64) map[string]*float64
func Float64Slice(vs []float64) []*float64
func GetDisableHTTPS(options ... interface{}) (значение bool, найдено bool)
func GetResolvedRegion(options ... interface{}) (строка значения, найденный bool)
func Int(v int) *int
func Int16(v int16) *int16
func Int16Map(vs map[string]int16) map[string]*int16
func Int16Slice(vs []int16) []*int16
func Int32(v int32) *int32
func Int32Map(vs map[string]int32) map[string]*int32
func Int32Slice(vs []int32) []*int32
func Int64(v int64) *int64
func Int64Map(vs map[string]int64) map[string]*int64
func Int64Slice(vs []int64) []*int64
func Int8(v int8) *int8
func Int8Map(vs map[string]int8) map[string]*int8
func Int8Slice(vs []int8) []*int8
func IntMap(vs map[string]int) map[string]*int
func IntSlice(vs []int) []*int
func IsCredentialsProvider(provider, target CredentialsProvider) bool
func String(v string) *string
func StringMap(vs map[string]string) map[string]*string
func StringSlice(vs []string) []*string
func Time(v time.Time) *time.Time
func TimeMap(vs map[string]time.Time) map[string]*time.Time
func TimeSlice(vs []time.Time) []*time.Time
func ToBool(p *bool) (v bool)
func ToBoolMap(vs map[string]*bool) map[string]bool
func ToBoolSlice(vs []*bool) []bool
func ToByte(p *byte) (v byte)
func ToByteMap(vs map[string]*byte) map[string]byte
func ToByteSlice(vs []*byte) []byte
func ToDuration(p *time.Duration) (v time.Duration)
func ToDurationMap(vs map[string]*time.Duration) map[string]time.Duration
func ToDurationSlice(vs []*time.Duration) []time.Duration
func ToFloat32(p *float32) (v float32)
func ToFloat32Map(vs map[string]*float32) map[string]float32
func ToFloat32Slice(vs []*float32) []float32
func ToFloat64(p *float64) (v float64)
func ToFloat64Map(vs map[string]*float64) map[string]float64
func ToFloat64Slice(vs []*float64) []float64
func ToInt(p *int) (v int)
func ToInt16(p *int16) (v int16)
func ToInt16Map(vs map[string]*int16) map[string]int16
func ToInt16Slice(vs []*int16) []int16
func ToInt32(p *int32) (v int32)
func ToInt32Map(vs map[string]*int32) map[string]int32
func ToInt32Slice(vs []*int32) []int32
func ToInt64(p *int64) (v int64)
func ToInt64Map(vs map[string]*int64) map[string]int64
func ToInt64Slice(vs []*int64) []int64
func ToInt8(p *int8) (v int8)
func ToInt8Map(vs map[string]*int8) map[string]int8
func ToInt8Slice(vs []*int8) []int8
func ToIntMap(vs map[string]*int) map[string]int
func ToIntSlice(vs []*int) []int
func ToString(p *string) (v string)
func ToStringMap(vs map[string]*string) map[string]string
func ToStringSlice(vs []*string) []string
func ToTime(p *time.Time) (v time.Time)
func ToTimeMap(vs map[string]*time.Time) map[string]time.Time
func ToTimeSlice(vs []*time.Time) []time.Time
func ToUint(p *uint) (v uint)
func ToUint16(p *uint16) (v uint16)
func ToUint16Map(vs map[string]*uint16) map[string]uint16
func ToUint16Slice(vs []*uint16) []uint16
func ToUint32(p *uint32) (v uint32)
func ToUint32Map(vs map[string]*uint32) map[string]uint32
func ToUint32Slice(vs []*uint32) []uint32
func ToUint64(p *uint64) (v uint64)
func ToUint64Map(vs map[string]*uint64) map[string]uint64
func ToUint64Slice(vs []*uint64) []uint64
func ToUint8(p *uint8) (v uint8)
func ToUint8Map(vs map[string]*uint8) map[string]uint8
func ToUint8Slice(vs []*uint8) []uint8
func ToUintMap(vs map[string]*uint) map[string]uint
func ToUintSlice(vs []*uint) []uint
func Uint(v uint) *uint
func Uint16(v uint16) *uint16
func Uint16Map(vs map[string]uint16) map[string]*uint16
func Uint16Slice(vs []uint16) []*uint16
func Uint32(v uint32) *uint32
func Uint32Map(vs map[string]uint32) map[string]*uint32
func Uint32Slice(vs []uint32) []*uint32
func Uint64(v uint64) *uint64
func Uint64Map(vs map[string]uint64) map[string]*uint64
func Uint64Slice(vs []uint64) []*uint64
func Uint8(v uint8) *uint8
func Uint8Map(vs map[string]uint8) map[string]*uint8
func Uint8Slice(vs []uint8) []*uint8
func UintMap(vs map[string]uint) map[string]*uint
func UintSlice(vs []uint) []*uint
type AccountIDEndpointMode
type AdjustExpiresByCredentialsCacheStrategy
type AnonymousCredentials
func (AnonymousCredentials) Retrieve(context.Context) (Credentials, error)
type ClientLogMode
func (m *ClientLogMode) ClearDeprecatedUsage()
func (m *ClientLogMode) ClearRequest()
func (m *ClientLogMode) ClearRequestEventMessage()
func (m *ClientLogMode) ClearRequestWithBody()
func (m *ClientLogMode) ClearResponse()
func (m *ClientLogMode) ClearResponseEventMessage()
func (m *ClientLogMode) ClearResponseWithBody()
func (m *ClientLogMode) ClearRetries()
func (m *ClientLogMode) ClearSigning()
func (m ClientLogMode) IsDeprecatedUsage() bool
func (m ClientLogMode) IsRequest() bool
func (m ClientLogMode) IsRequestEventMessage() bool
func (m ClientLogMode) IsRequestWithBody() bool
func (m ClientLogMode) IsResponse() bool
func (m ClientLogMode) IsResponseEventMessage() bool
func (m ClientLogMode) IsResponseWithBody() bool
func (m ClientLogMode) IsRetries() bool
func (m ClientLogMode) IsSigning() bool
type Config
func NewConfig() *Config
func (c Config) Copy() Config
type CredentialProviderSource
type CredentialSource
type Credentials
func (v Credentials) Expired() bool
func (v Credentials) HasKeys() bool
type CredentialsCache
func NewCredentialsCache(provider CredentialsProvider, optFns ...func(options *CredentialsCacheOptions)) *CredentialsCache
func (p *CredentialsCache) Invalidate()
func (p *CredentialsCache) IsCredentialsProvider(target CredentialsProvider) bool
func (p *CredentialsCache) ProviderSources() []CredentialSource
func (p *CredentialsCache) Retrieve(ctx context.Context) (Credentials, error)
type CredentialsCacheOptions
type CredentialsProvider
type CredentialsProviderFunc
func (fn CredentialsProviderFunc) Retrieve(ctx context.Context) (Credentials, error)
type DefaultsMode
func (d *DefaultsMode) SetFromString(v string) (ok bool)
type DualStackEndpointState
func GetUseDualStackEndpoint(options ...interface{}) (value DualStackEndpointState, found bool)
type Endpointdeprecated
type EndpointDiscoveryEnableState
type EndpointNotFoundError
func (e *EndpointNotFoundError) Error() string
func (e *EndpointNotFoundError) Unwrap() error
type EndpointResolverdeprecated
type EndpointResolverFuncdeprecated
func (e EndpointResolverFunc) ResolveEndpoint(service, region string) (Endpoint, error)
type EndpointResolverWithOptionsdeprecated
type EndpointResolverWithOptionsFuncdeprecated
func (e EndpointResolverWithOptionsFunc) ResolveEndpoint(service, region string, options ...interface{}) (Endpoint, error)
type EndpointSourcedeprecated
type ExecutionEnvironmentID
type FIPSEndpointState
func GetUseFIPSEndpoint(options ...interface{}) (value FIPSEndpointState, found bool)
type HTTPClient
type HandleFailRefreshCredentialsCacheStrategy
type MissingRegionError
func (*MissingRegionError) Error() string
type NopRetryer
func (NopRetryer) GetAttemptToken(context.Context) (func(error) error, error)
func (NopRetryer) GetInitialToken() func(error) error
func (NopRetryer) GetRetryToken(context.Context, error) (func(error) error, error)
func (NopRetryer) IsErrorRetryable(error) bool
func (NopRetryer) MaxAttempts() int
func (NopRetryer) RetryDelay(int, error) (time.Duration, error)
type RequestCanceledError
func (*RequestCanceledError) CanceledError() bool
func (e *RequestCanceledError) Error() string
func (e *RequestCanceledError) Unwrap() error
type RequestChecksumCalculation
type ResponseChecksumValidation
type RetryMode
func ParseRetryMode(v string) (mode RetryMode, err error)
func (m RetryMode) String() string
type Retryer
type RetryerV2
type RuntimeEnvironment
type Ternary
func BoolTernary(v bool) Ternary
func (t Ternary) Bool() bool
func (t Ternary) String() string
Constants ¶
View Source
const (
	// AccountIDEndpointModeUnset indicates the AWS account ID will not be used for endpoint routing
	AccountIDEndpointModeUnset AccountIDEndpointMode = ""

	// AccountIDEndpointModePreferred indicates the AWS account ID will be used for endpoint routing if present
	AccountIDEndpointModePreferred = "preferred"

	// AccountIDEndpointModeRequired indicates an error will be returned if the AWS account ID is not resolved from identity
	AccountIDEndpointModeRequired = "required"

	// AccountIDEndpointModeDisabled indicates the AWS account ID will be ignored during endpoint routing
	AccountIDEndpointModeDisabled = "disabled"
)
View Source
const SDKName = "aws-sdk-go-v2"
SDKName is the name of this AWS SDK

View Source
const SDKVersion = goModuleVersion
SDKVersion is the version of this SDK

Variables ¶
This section is empty.

Functions ¶
func Bool ¶
func Bool(v bool) *bool
Bool returns a pointer value for the bool value passed in.

func BoolMap ¶
func BoolMap(vs map[string]bool) map[string]*bool
BoolMap returns a map of bool pointers from the values passed in.

func BoolSlice ¶
func BoolSlice(vs []bool) []*bool
BoolSlice returns a slice of bool pointers from the values passed in.

func Byte ¶
added in v0.25.0
func Byte(v byte) *byte
Byte returns a pointer value for the byte value passed in.

func ByteMap ¶
added in v0.25.0
func ByteMap(vs map[string]byte) map[string]*byte
ByteMap returns a map of byte pointers from the values passed in.

func ByteSlice ¶
added in v0.25.0
func ByteSlice(vs []byte) []*byte
ByteSlice returns a slice of byte pointers from the values passed in.

func Duration ¶
added in v1.13.0
func Duration(v time.Duration) *time.Duration
Duration returns a pointer value for the time.Duration value passed in.

func DurationMap ¶
added in v1.13.0
func DurationMap(vs map[string]time.Duration) map[string]*time.Duration
DurationMap returns a map of time.Duration pointers from the values passed in.

func DurationSlice ¶
added in v1.13.0
func DurationSlice(vs []time.Duration) []*time.Duration
DurationSlice returns a slice of time.Duration pointers from the values passed in.

func Float32 ¶
added in v0.13.0
func Float32(v float32) *float32
Float32 returns a pointer value for the float32 value passed in.

func Float32Map ¶
added in v0.13.0
func Float32Map(vs map[string]float32) map[string]*float32
Float32Map возвращает карту указателей float32 из значений прошел.

функционер Float32Slice ¶
Добавлено в версии 0.13.0
func Float32Slice(vs []float32) []*float32
Float32Slice возвращает срез указателей float32 из значений прошел.

функционер Float64 ¶
func Float64(v float64) *float64
Float64 возвращает значение указателя для переданного значения float64.

func Float64Карта ¶
func Float64Map(vs map[string]float64) map[string]*float64
Float64Map возвращает карту указателей float64 из значений прошел.

функционер Float64Slice ¶
func Float64Slice(vs []float64) []*float64
Float64Slice возвращает срез указателей float64 из значений прошел.

func GetDisableHTTPS ¶
Добавлено в v1.11.0
func GetDisableHTTPS(options ...interface{}) (value bool, found bool)
GetDisableHTTPS принимает EndpointResolverOptions службы и возвращает значение DisableHTTPS. Возвращает логическое значение false, если предоставленные параметры не имеют метода для получения DisableHTTPS.

func GetResolvedRegion ¶
Добавлено в v1.11.0
func GetResolvedRegion(options ...interface{}) (value string, found bool)
GetResolvedRegion принимает EndpointResolverOptions службы и возвращает значение ResolvedRegion. Возвращает логическое значение false, если предоставленные параметры не имеют метода для извлечения ResolvedRegion.

func Int ¶
func Int(v int) *int
Int возвращает значение указателя для переданного значения int.

функционер Int16 ¶
Добавлено в версии 0.13.0
func Int16(v int16) *int16
Int16 возвращает значение указателя для переданного значения int16.

func Int16Map ¶
Добавлено в версии 0.13.0
func Int16Map(vs map[string]int16) map[string]*int16
Int16Map возвращает карту указателей int16 из значений прошел.

func Int16Slice ¶
Добавлено в версии 0.13.0
func Int16Slice(vs []int16) []*int16
Int16Slice возвращает срез указателей int16 из значений прошел.

функционер Int32 ¶
Добавлено в версии 0.13.0
func Int32(v int32) *int32
Int32 возвращает значение указателя для переданного значения int32.

func Int32Map ¶
Добавлено в версии 0.13.0
func Int32Map(vs map[string]int32) map[string]*int32
Int32Map возвращает карту указателей int32 из значений прошел.

func Int32Slice ¶
Добавлено в версии 0.13.0
func Int32Slice(vs []int32) []*int32
Int32Slice возвращает срез указателей int32 из значений прошел.

функционер Int64 ¶
func Int64(v int64) *int64
Int64 возвращает значение указателя для переданного значения int64.

func Int64Map ¶
func Int64Map(vs map[string]int64) map[string]*int64
Int64Map возвращает карту указателей int64 из значений прошел.

func Int64Slice ¶
func Int64Slice(vs []int64) []*int64
Int64Slice возвращает срез указателей int64 из значений прошел.

функционер Int8 ¶
Добавлено в версии 0.13.0
func Int8(v int8) *int8
Int8 возвращает значение указателя для переданного значения int8.

func Int8Map ¶
Добавлено в версии 0.13.0
func Int8Map(vs map[string]int8) map[string]*int8
Int8Map возвращает карту указателей int8 из значений прошел.

func Int8Slice ¶
added in v0.13.0
func Int8Slice(vs []int8) []*int8
Int8Slice returns a slice of int8 pointers from the values passed in.

func IntMap ¶
func IntMap(vs map[string]int) map[string]*int
IntMap returns a map of int pointers from the values passed in.

func IntSlice ¶
func IntSlice(vs []int) []*int
IntSlice returns a slice of int pointers from the values passed in.

func IsCredentialsProvider ¶
added in v1.17.0
func IsCredentialsProvider(provider, target CredentialsProvider) bool
IsCredentialsProvider returns whether the target CredentialProvider is the same type as provider when comparing the implementation type.

If provider has a method IsCredentialsProvider(CredentialsProvider) bool it will be responsible for validating whether target matches the credential provider type.

When comparing the CredentialProvider implementations provider and target for equality, the following rules are used:

If provider is of type T and target is of type V, true if type *T is the same as type *V, otherwise false
If provider is of type *T and target is of type V, true if type *T is the same as type *V, otherwise false
If provider is of type T and target is of type *V, true if type *T is the same as type *V, otherwise false
If provider is of type *T and target is of type *V,true if type *T is the same as type *V, otherwise false
func String ¶
func String(v string) *string
String returns a pointer value for the string value passed in.

func StringMap ¶
func StringMap(vs map[string]string) map[string]*string
StringMap returns a map of string pointers from the values passed in.

func StringSlice ¶
func StringSlice(vs []string) []*string
StringSlice returns a slice of string pointers from the values passed in.

func Time ¶
func Time(v time.Time) *time.Time
Time returns a pointer value for the time.Time value passed in.

func TimeMap ¶
func TimeMap(vs map[string]time.Time) map[string]*time.Time
TimeMap returns a map of time.Time pointers from the values passed in.

func TimeSlice ¶
func TimeSlice(vs []time.Time) []*time.Time
TimeSlice returns a slice of time.Time pointers from the values passed in.

func ToBool ¶
added in v0.25.0
func ToBool(p *bool) (v bool)
ToBool returns bool value dereferenced if the passed in pointer was not nil. Returns a bool zero value if the pointer was nil.

func ToBoolMap ¶
added in v0.25.0
func ToBoolMap(vs map[string]*bool) map[string]bool
ToBoolMap returns a map of bool values, that are dereferenced if the passed in pointer was not nil. The bool zero value is used if the pointer was nil.

func ToBoolSlice ¶
added in v0.25.0
func ToBoolSlice(vs []*bool) []bool
ToBoolSlice returns a slice of bool values, that are dereferenced if the passed in pointer was not nil. Returns a bool zero value if the pointer was nil.

func ToByte ¶
added in v0.25.0
func ToByte(p *byte) (v byte)
ToByte returns byte value dereferenced if the passed in pointer was not nil. Returns a byte zero value if the pointer was nil.

func ToByteMap ¶
added in v0.25.0
func ToByteMap(vs map[string]*byte) map[string]byte
ToByteMap returns a map of byte values, that are dereferenced if the passed in pointer was not nil. The byte zero value is used if the pointer was nil.

func ToByteSlice ¶
added in v0.25.0
func ToByteSlice(vs []*byte) []byte
ToByteSlice returns a slice of byte values, that are dereferenced if the passed in pointer was not nil. Returns a byte zero value if the pointer was nil.

func ToDuration ¶
added in v1.13.0
func ToDuration(p *time.Duration) (v time.Duration)
ToDuration returns time.Duration value dereferenced if the passed in pointer was not nil. Returns a time.Duration zero value if the pointer was nil.

func ToDurationMap ¶
added in v1.13.0
func ToDurationMap(vs map[string]*time.Duration) map[string]time.Duration
ToDurationMap returns a map of time.Duration values, that are dereferenced if the passed in pointer was not nil. The time.Duration zero value is used if the pointer was nil.

func ToDurationSlice ¶
added in v1.13.0
func ToDurationSlice(vs []*time.Duration) []time.Duration
ToDurationSlice returns a slice of time.Duration values, that are dereferenced if the passed in pointer was not nil. Returns a time.Duration zero value if the pointer was nil.

func ToFloat32 ¶
added in v0.25.0
func ToFloat32(p *float32) (v float32)
ToFloat32 returns float32 value dereferenced if the passed in pointer was not nil. Returns a float32 zero value if the pointer was nil.

func ToFloat32Map ¶
added in v0.25.0
func ToFloat32Map(vs map[string]*float32) map[string]float32
ToFloat32Map returns a map of float32 values, that are dereferenced if the passed in pointer was not nil. The float32 zero value is used if the pointer was nil.

func ToFloat32Slice ¶
added in v0.25.0
func ToFloat32Slice(vs []*float32) []float32
ToFloat32Slice returns a slice of float32 values, that are dereferenced if the passed in pointer was not nil. Returns a float32 zero value if the pointer was nil.

func ToFloat64 ¶
added in v0.25.0
func ToFloat64(p *float64) (v float64)
ToFloat64 returns float64 value dereferenced if the passed in pointer was not nil. Returns a float64 zero value if the pointer was nil.

func ToFloat64Map ¶
added in v0.25.0
func ToFloat64Map(vs map[string]*float64) map[string]float64
ToFloat64Map returns a map of float64 values, that are dereferenced if the passed in pointer was not nil. The float64 zero value is used if the pointer was nil.

func ToFloat64Slice ¶
added in v0.25.0
func ToFloat64Slice(vs []*float64) []float64
ToFloat64Slice returns a slice of float64 values, that are dereferenced if the passed in pointer was not nil. Returns a float64 zero value if the pointer was nil.

func ToInt ¶
added in v0.25.0
func ToInt(p *int) (v int)
ToInt returns int value dereferenced if the passed in pointer was not nil. Returns a int zero value if the pointer was nil.

func ToInt16 ¶
added in v0.25.0
func ToInt16(p *int16) (v int16)
ToInt16 returns int16 value dereferenced if the passed in pointer was not nil. Returns a int16 zero value if the pointer was nil.

func ToInt16Map ¶
added in v0.25.0
func ToInt16Map(vs map[string]*int16) map[string]int16
ToInt16Map returns a map of int16 values, that are dereferenced if the passed in pointer was not nil. The int16 zero value is used if the pointer was nil.

func ToInt16Slice ¶
added in v0.25.0
func ToInt16Slice(vs []*int16) []int16
ToInt16Slice returns a slice of int16 values, that are dereferenced if the passed in pointer was not nil. Returns a int16 zero value if the pointer was nil.

func ToInt32 ¶
added in v0.25.0
func ToInt32(p *int32) (v int32)
ToInt32 returns int32 value dereferenced if the passed in pointer was not nil. Returns a int32 zero value if the pointer was nil.

func ToInt32Map ¶
added in v0.25.0
func ToInt32Map(vs map[string]*int32) map[string]int32
ToInt32Map returns a map of int32 values, that are dereferenced if the passed in pointer was not nil. The int32 zero value is used if the pointer was nil.

func ToInt32Slice ¶
added in v0.25.0
func ToInt32Slice(vs []*int32) []int32
ToInt32Slice returns a slice of int32 values, that are dereferenced if the passed in pointer was not nil. Returns a int32 zero value if the pointer was nil.

func ToInt64 ¶
added in v0.25.0
func ToInt64(p *int64) (v int64)
ToInt64 returns int64 value dereferenced if the passed in pointer was not nil. Returns a int64 zero value if the pointer was nil.

func ToInt64Map ¶
added in v0.25.0
func ToInt64Map(vs map[string]*int64) map[string]int64
ToInt64Map returns a map of int64 values, that are dereferenced if the passed in pointer was not nil. The int64 zero value is used if the pointer was nil.

func ToInt64Slice ¶
added in v0.25.0
func ToInt64Slice(vs []*int64) []int64
ToInt64Slice returns a slice of int64 values, that are dereferenced if the passed in pointer was not nil. Returns a int64 zero value if the pointer was nil.

func ToInt8 ¶
added in v0.25.0
func ToInt8(p *int8) (v int8)
ToInt8 returns int8 value dereferenced if the passed in pointer was not nil. Returns a int8 zero value if the pointer was nil.

func ToInt8Map ¶
added in v0.25.0
func ToInt8Map(vs map[string]*int8) map[string]int8
ToInt8Map returns a map of int8 values, that are dereferenced if the passed in pointer was not nil. The int8 zero value is used if the pointer was nil.

func ToInt8Slice ¶
added in v0.25.0
func ToInt8Slice(vs []*int8) []int8
ToInt8Slice returns a slice of int8 values, that are dereferenced if the passed in pointer was not nil. Returns a int8 zero value if the pointer was nil.

func ToIntMap ¶
added in v0.25.0
func ToIntMap(vs map[string]*int) map[string]int
ToIntMap returns a map of int values, that are dereferenced if the passed in pointer was not nil. The int zero value is used if the pointer was nil.

func ToIntSlice ¶
added in v0.25.0
func ToIntSlice(vs []*int) []int
ToIntSlice returns a slice of int values, that are dereferenced if the passed in pointer was not nil. Returns a int zero value if the pointer was nil.

func ToString ¶
added in v0.25.0
func ToString(p *string) (v string)
ToString returns string value dereferenced if the passed in pointer was not nil. Returns a string zero value if the pointer was nil.

func ToStringMap ¶
added in v0.25.0
func ToStringMap(vs map[string]*string) map[string]string
ToStringMap returns a map of string values, that are dereferenced if the passed in pointer was not nil. The string zero value is used if the pointer was nil.

func ToStringSlice ¶
added in v0.25.0
func ToStringSlice(vs []*string) []string
ToStringSlice returns a slice of string values, that are dereferenced if the passed in pointer was not nil. Returns a string zero value if the pointer was nil.

func ToTime ¶
added in v0.25.0
func ToTime(p *time.Time) (v time.Time)
ToTime returns time.Time value dereferenced if the passed in pointer was not nil. Returns a time.Time zero value if the pointer was nil.

func ToTimeMap ¶
added in v0.25.0
func ToTimeMap(vs map[string]*time.Time) map[string]time.Time
ToTimeMap returns a map of time.Time values, that are dereferenced if the passed in pointer was not nil. The time.Time zero value is used if the pointer was nil.

func ToTimeSlice ¶
added in v0.25.0
func ToTimeSlice(vs []*time.Time) []time.Time
ToTimeSlice returns a slice of time.Time values, that are dereferenced if the passed in pointer was not nil. Returns a time.Time zero value if the pointer was nil.

func ToUint ¶
added in v0.25.0
func ToUint(p *uint) (v uint)
ToUint returns uint value dereferenced if the passed in pointer was not nil. Returns a uint zero value if the pointer was nil.

func ToUint16 ¶
added in v0.25.0
func ToUint16(p *uint16) (v uint16)
ToUint16 returns uint16 value dereferenced if the passed in pointer was not nil. Returns a uint16 zero value if the pointer was nil.

func ToUint16Map ¶
added in v0.25.0
func ToUint16Map(vs map[string]*uint16) map[string]uint16
ToUint16Map returns a map of uint16 values, that are dereferenced if the passed in pointer was not nil. The uint16 zero value is used if the pointer was nil.

func ToUint16Slice ¶
added in v0.25.0
func ToUint16Slice(vs []*uint16) []uint16
ToUint16Slice returns a slice of uint16 values, that are dereferenced if the passed in pointer was not nil. Returns a uint16 zero value if the pointer was nil.

func ToUint32 ¶
added in v0.25.0
func ToUint32(p *uint32) (v uint32)
ToUint32 returns uint32 value dereferenced if the passed in pointer was not nil. Returns a uint32 zero value if the pointer was nil.

func ToUint32Map ¶
added in v0.25.0
func ToUint32Map(vs map[string]*uint32) map[string]uint32
ToUint32Map returns a map of uint32 values, that are dereferenced if the passed in pointer was not nil. The uint32 zero value is used if the pointer was nil.

func ToUint32Slice ¶
added in v0.25.0
func ToUint32Slice(vs []*uint32) []uint32
ToUint32Slice returns a slice of uint32 values, that are dereferenced if the passed in pointer was not nil. Returns a uint32 zero value if the pointer was nil.

func ToUint64 ¶
added in v0.25.0
func ToUint64(p *uint64) (v uint64)
ToUint64 returns uint64 value dereferenced if the passed in pointer was not nil. Returns a uint64 zero value if the pointer was nil.

func ToUint64Map ¶
added in v0.25.0
func ToUint64Map(vs map[string]*uint64) map[string]uint64
ToUint64Map returns a map of uint64 values, that are dereferenced if the passed in pointer was not nil. The uint64 zero value is used if the pointer was nil.

func ToUint64Slice ¶
added in v0.25.0
func ToUint64Slice(vs []*uint64) []uint64
ToUint64Slice returns a slice of uint64 values, that are dereferenced if the passed in pointer was not nil. Returns a uint64 zero value if the pointer was nil.

func ToUint8 ¶
added in v0.25.0
func ToUint8(p *uint8) (v uint8)
ToUint8 returns uint8 value dereferenced if the passed in pointer was not nil. Returns a uint8 zero value if the pointer was nil.

func ToUint8Map ¶
added in v0.25.0
func ToUint8Map(vs map[string]*uint8) map[string]uint8
ToUint8Map returns a map of uint8 values, that are dereferenced if the passed in pointer was not nil. The uint8 zero value is used if the pointer was nil.

func ToUint8Slice ¶
added in v0.25.0
func ToUint8Slice(vs []*uint8) []uint8
ToUint8Slice returns a slice of uint8 values, that are dereferenced if the passed in pointer was not nil. Returns a uint8 zero value if the pointer was nil.

func ToUintMap ¶
added in v0.25.0
func ToUintMap(vs map[string]*uint) map[string]uint
ToUintMap returns a map of uint values, that are dereferenced if the passed in pointer was not nil. The uint zero value is used if the pointer was nil.

func ToUintSlice ¶
added in v0.25.0
func ToUintSlice(vs []*uint) []uint
ToUintSlice returns a slice of uint values, that are dereferenced if the passed in pointer was not nil. Returns a uint zero value if the pointer was nil.

func Uint ¶
added in v0.13.0
func Uint(v uint) *uint
Uint returns a pointer value for the uint value passed in.

func Uint16 ¶
added in v0.13.0
func Uint16(v uint16) *uint16
Uint16 returns a pointer value for the uint16 value passed in.

func Uint16Map ¶
added in v0.13.0
func Uint16Map(vs map[string]uint16) map[string]*uint16
Uint16Map returns a map of uint16 pointers from the values passed in.

func Uint16Slice ¶
added in v0.13.0
func Uint16Slice(vs []uint16) []*uint16
Uint16Slice returns a slice of uint16 pointers from the values passed in.

func Uint32 ¶
added in v0.13.0
func Uint32(v uint32) *uint32
Uint32 returns a pointer value for the uint32 value passed in.

func Uint32Map ¶
added in v0.13.0
func Uint32Map(vs map[string]uint32) map[string]*uint32
Uint32Map returns a map of uint32 pointers from the values passed in.

func Uint32Slice ¶
added in v0.13.0
func Uint32Slice(vs []uint32) []*uint32
Uint32Slice returns a slice of uint32 pointers from the values passed in.

func Uint64 ¶
added in v0.13.0
func Uint64(v uint64) *uint64
Uint64 returns a pointer value for the uint64 value passed in.

func Uint64Map ¶
added in v0.13.0
func Uint64Map(vs map[string]uint64) map[string]*uint64
Uint64Map returns a map of uint64 pointers from the values passed in.

func Uint64Slice ¶
added in v0.13.0
func Uint64Slice(vs []uint64) []*uint64
Uint64Slice returns a slice of uint64 pointers from the values passed in.

func Uint8 ¶
added in v0.13.0
func Uint8(v uint8) *uint8
Uint8 returns a pointer value for the uint8 value passed in.

func Uint8Map ¶
added in v0.13.0
func Uint8Map(vs map[string]uint8) map[string]*uint8
Uint8Map returns a map of uint8 pointers from the values passed in.

func Uint8Slice ¶
added in v0.13.0
func Uint8Slice(vs []uint8) []*uint8
Uint8Slice returns a slice of uint8 pointers from the values passed in.

func UintMap ¶
added in v0.13.0
func UintMap(vs map[string]uint) map[string]*uint
UintMap returns a map of uint pointers from the values passed in.

func UintSlice ¶
added in v0.13.0
func UintSlice(vs []uint) []*uint
UintSlice returns a slice of uint pointers from the values passed in.

Types ¶
type AccountIDEndpointMode ¶
added in v1.28.0
type AccountIDEndpointMode string
AccountIDEndpointMode controls how a resolved AWS account ID is handled for endpoint routing.

type AdjustExpiresByCredentialsCacheStrategy ¶
added in v1.16.0
type AdjustExpiresByCredentialsCacheStrategy interface {
	// Given a Credentials as input, applying any mutations and
	// returning the potentially updated Credentials, or error.
	AdjustExpiresBy(Credentials, time.Duration) (Credentials, error)
}
AdjustExpiresByCredentialsCacheStrategy is an interface for CredentialCache to allow CredentialsProvider to intercept adjustments to Credentials expiry based on expectations and use cases of CredentialsProvider.

Credential caches may use default implementation if nil.

type AnonymousCredentials ¶
type AnonymousCredentials struct{}
AnonymousCredentials provides a sentinel CredentialsProvider that should be used to instruct the SDK's signing middleware to not sign the request.

Using `nil` credentials when configuring an API client will achieve the same result. The AnonymousCredentials type allows you to configure the SDK's external config loading to not attempt to source credentials from the shared config or environment.

For example you can use this CredentialsProvider with an API client's Options to instruct the client not to sign a request for accessing public S3 bucket objects.

The following example demonstrates using the AnonymousCredentials to prevent SDK's external config loading attempt to resolve credentials.

cfg, err := config.LoadDefaultConfig(context.TODO(),
     config.WithCredentialsProvider(aws.AnonymousCredentials{}),
)
if err != nil {
     log.Fatalf("failed to load config, %v", err)
}

client := s3.NewFromConfig(cfg)
Alternatively you can leave the API client Option's `Credential` member to nil. If using the `NewFromConfig` constructor you'll need to explicitly set the `Credentials` member to nil, if the external config resolved a credential provider.

client := s3.New(s3.Options{
     // Credentials defaults to a nil value.
})
This can also be configured for specific operations calls too.

cfg, err := config.LoadDefaultConfig(context.TODO())
if err != nil {
     log.Fatalf("failed to load config, %v", err)
}

client := s3.NewFromConfig(config)

result, err := client.GetObject(context.TODO(), s3.GetObject{
     Bucket: aws.String("example-bucket"),
     Key: aws.String("example-key"),
}, func(o *s3.Options) {
     o.Credentials = nil
     // Or
     o.Credentials = aws.AnonymousCredentials{}
})
func (AnonymousCredentials) Retrieve ¶
added in v0.25.0
func (AnonymousCredentials) Retrieve(context.Context) (Credentials, error)
Retrieve implements the CredentialsProvider interface, but will always return error, and cannot be used to sign a request. The AnonymousCredentials type is used as a sentinel type instructing the AWS request signing middleware to not sign a request.

type ClientLogMode ¶
added in v0.30.0
type ClientLogMode uint64
ClientLogMode represents the logging mode of SDK clients. The client logging mode is a bit-field where each bit is a flag that describes the logging behavior for one or more client components. The entire 64-bit group is reserved for later expansion by the SDK.

Example: Setting ClientLogMode to enable logging of retries and requests

clientLogMode := aws.LogRetries | aws.LogRequest
Example: Adding an additional log mode to an existing ClientLogMode value

clientLogMode |= aws.LogResponse
const (
	LogSigning ClientLogMode = 1 << (64 - 1 - iota)
	LogRetries
	LogRequest
	LogRequestWithBody
	LogResponse
	LogResponseWithBody
	LogDeprecatedUsage
	LogRequestEventMessage
	LogResponseEventMessage
)
Supported ClientLogMode bits that can be configured to toggle logging of specific SDK events.

func (*ClientLogMode) ClearDeprecatedUsage ¶
added in v1.11.0
func (m *ClientLogMode) ClearDeprecatedUsage()
ClearDeprecatedUsage clears the DeprecatedUsage logging mode bit

func (*ClientLogMode) ClearRequest ¶
added in v0.30.0
func (m *ClientLogMode) ClearRequest()
ClearRequest clears the Request logging mode bit

func (*ClientLogMode) ClearRequestEventMessage ¶
added in v1.11.0
func (m *ClientLogMode) ClearRequestEventMessage()
ClearRequestEventMessage clears the RequestEventMessage logging mode bit

func (*ClientLogMode) ClearRequestWithBody ¶
added in v0.30.0
func (m *ClientLogMode) ClearRequestWithBody()
ClearRequestWithBody clears the RequestWithBody logging mode bit

func (*ClientLogMode) ClearResponse ¶
added in v0.30.0
func (m *ClientLogMode) ClearResponse()
ClearResponse clears the Response logging mode bit

func (*ClientLogMode) ClearResponseEventMessage ¶
added in v1.11.0
func (m *ClientLogMode) ClearResponseEventMessage()
ClearResponseEventMessage clears the ResponseEventMessage logging mode bit

func (*ClientLogMode) ClearResponseWithBody ¶
added in v0.30.0
func (m *ClientLogMode) ClearResponseWithBody()
ClearResponseWithBody clears the ResponseWithBody logging mode bit

func (*ClientLogMode) ClearRetries ¶
added in v0.30.0
func (m *ClientLogMode) ClearRetries()
ClearRetries clears the Retries logging mode bit

func (*ClientLogMode) ClearSigning ¶
added in v0.30.0
func (m *ClientLogMode) ClearSigning()
ClearSigning clears the Signing logging mode bit

func (ClientLogMode) IsDeprecatedUsage ¶
added in v1.11.0
func (m ClientLogMode) IsDeprecatedUsage() bool
IsDeprecatedUsage returns whether the DeprecatedUsage logging mode bit is set

func (ClientLogMode) IsRequest ¶
added in v0.30.0
func (m ClientLogMode) IsRequest() bool
IsRequest returns whether the Request logging mode bit is set

func (ClientLogMode) IsRequestEventMessage ¶
added in v1.11.0
func (m ClientLogMode) IsRequestEventMessage() bool
IsRequestEventMessage returns whether the RequestEventMessage logging mode bit is set

func (ClientLogMode) IsRequestWithBody ¶
added in v0.30.0
func (m ClientLogMode) IsRequestWithBody() bool
IsRequestWithBody returns whether the RequestWithBody logging mode bit is set

func (ClientLogMode) IsResponse ¶
added in v0.30.0
func (m ClientLogMode) IsResponse() bool
IsResponse returns whether the Response logging mode bit is set

func (ClientLogMode) IsResponseEventMessage ¶
added in v1.11.0
func (m ClientLogMode) IsResponseEventMessage() bool
IsResponseEventMessage returns whether the ResponseEventMessage logging mode bit is set

func (ClientLogMode) IsResponseWithBody ¶
added in v0.30.0
func (m ClientLogMode) IsResponseWithBody() bool
IsResponseWithBody returns whether the ResponseWithBody logging mode bit is set

func (ClientLogMode) IsRetries ¶
added in v0.30.0
func (m ClientLogMode) IsRetries() bool
IsRetries returns whether the Retries logging mode bit is set

func (ClientLogMode) IsSigning ¶
added in v0.30.0
func (m ClientLogMode) IsSigning() bool
IsSigning returns whether the Signing logging mode bit is set

type Config ¶
type Config struct {
	// The region to send requests to. This parameter is required and must
	// be configured globally or on a per-client basis unless otherwise
	// noted. A full list of regions is found in the "Regions and Endpoints"
	// document.
	//
	// See http://docs.aws.amazon.com/general/latest/gr/rande.html for
	// information on AWS regions.
	Region string

	// The credentials object to use when signing requests.
	// Use the LoadDefaultConfig to load configuration from all the SDK's supported
	// sources, and resolve credentials using the SDK's default credential chain.
	Credentials CredentialsProvider

	// The Bearer Authentication token provider to use for authenticating API
	// operation calls with a Bearer Authentication token. The API clients and
	// operation must support Bearer Authentication scheme in order for the
	// token provider to be used. API clients created with NewFromConfig will
	// automatically be configured with this option, if the API client support
	// Bearer Authentication.
	//
	// The SDK's config.LoadDefaultConfig can automatically populate this
	// option for external configuration options such as SSO session.
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html
	BearerAuthTokenProvider smithybearer.TokenProvider

	// The HTTP Client the SDK's API clients will use to invoke HTTP requests.
	// The SDK defaults to a BuildableClient allowing API clients to create
	// copies of the HTTP Client for service specific customizations.
	//
	// Use a (*http.Client) for custom behavior. Using a custom http.Client
	// will prevent the SDK from modifying the HTTP client.
	HTTPClient HTTPClient

	// An endpoint resolver that can be used to provide or override an endpoint
	// for the given service and region.
	//
	// See the `aws.EndpointResolver` documentation for additional usage
	// information.
	//
	// Deprecated: See Config.EndpointResolverWithOptions
	EndpointResolver EndpointResolver

	// An endpoint resolver that can be used to provide or override an endpoint
	// for the given service and region.
	//
	// When EndpointResolverWithOptions is specified, it will be used by a
	// service client rather than using EndpointResolver if also specified.
	//
	// See the `aws.EndpointResolverWithOptions` documentation for additional
	// usage information.
	//
	// Deprecated: with the release of endpoint resolution v2 in API clients,
	// EndpointResolver and EndpointResolverWithOptions are deprecated.
	// Providing a value for this field will likely prevent you from using
	// newer endpoint-related service features. See API client options
	// EndpointResolverV2 and BaseEndpoint.
	EndpointResolverWithOptions EndpointResolverWithOptions

	// RetryMaxAttempts specifies the maximum number attempts an API client
	// will call an operation that fails with a retryable error.
	//
	// API Clients will only use this value to construct a retryer if the
	// Config.Retryer member is not nil. This value will be ignored if
	// Retryer is not nil.
	RetryMaxAttempts int

	// RetryMode specifies the retry model the API client will be created with.
	//
	// API Clients will only use this value to construct a retryer if the
	// Config.Retryer member is not nil. This value will be ignored if
	// Retryer is not nil.
	RetryMode RetryMode

	// Retryer is a function that provides a Retryer implementation. A Retryer
	// guides how HTTP requests should be retried in case of recoverable
	// failures. When nil the API client will use a default retryer.
	//
	// In general, the provider function should return a new instance of a
	// Retryer if you are attempting to provide a consistent Retryer
	// configuration across all clients. This will ensure that each client will
	// be provided a new instance of the Retryer implementation, and will avoid
	// issues such as sharing the same retry token bucket across services.
	//
	// If not nil, RetryMaxAttempts, and RetryMode will be ignored by API
	// clients.
	Retryer func() Retryer

	// ConfigSources are the sources that were used to construct the Config.
	// Allows for additional configuration to be loaded by clients.
	ConfigSources []interface{}

	// APIOptions provides the set of middleware mutations modify how the API
	// client requests will be handled. This is useful for adding additional
	// tracing data to a request, or changing behavior of the SDK's client.
	APIOptions []func(*middleware.Stack) error

	// The logger writer interface to write logging messages to. Defaults to
	// standard error.
	Logger logging.Logger

	// Configures the events that will be sent to the configured logger. This
	// can be used to configure the logging of signing, retries, request, and
	// responses of the SDK clients.
	//
	// See the ClientLogMode type documentation for the complete set of logging
	// modes and available configuration.
	ClientLogMode ClientLogMode

	// The configured DefaultsMode. If not specified, service clients will
	// default to legacy.
	//
	// Supported modes are: auto, cross-region, in-region, legacy, mobile,
	// standard
	DefaultsMode DefaultsMode

	// The RuntimeEnvironment configuration, only populated if the DefaultsMode
	// is set to DefaultsModeAuto and is initialized by
	// `config.LoadDefaultConfig`. You should not populate this structure
	// programmatically, or rely on the values here within your applications.
	RuntimeEnvironment RuntimeEnvironment

	// AppId is an optional application specific identifier that can be set.
	// When set it will be appended to the User-Agent header of every request
	// in the form of App/{AppId}. This variable is sourced from environment
	// variable AWS_SDK_UA_APP_ID or the shared config profile attribute sdk_ua_app_id.
	// See https://docs.aws.amazon.com/sdkref/latest/guide/settings-reference.html for
	// more information on environment variables and shared config settings.
	AppID string

	// BaseEndpoint is an intermediary transfer location to a service specific
	// BaseEndpoint on a service's Options.
	BaseEndpoint *string

	// DisableRequestCompression toggles if an operation request could be
	// compressed or not. Will be set to false by default. This variable is sourced from
	// environment variable AWS_DISABLE_REQUEST_COMPRESSION or the shared config profile attribute
	// disable_request_compression
	DisableRequestCompression bool

	// RequestMinCompressSizeBytes sets the inclusive min bytes of a request body that could be
	// compressed. Will be set to 10240 by default and must be within 0 and 10485760 bytes inclusively.
	// This variable is sourced from environment variable AWS_REQUEST_MIN_COMPRESSION_SIZE_BYTES or
	// the shared config profile attribute request_min_compression_size_bytes
	RequestMinCompressSizeBytes int64

	// Controls how a resolved AWS account ID is handled for endpoint routing.
	AccountIDEndpointMode AccountIDEndpointMode

	// RequestChecksumCalculation determines when request checksum calculation is performed.
	//
	// There are two possible values for this setting:
	//
	// 1. RequestChecksumCalculationWhenSupported (default): The checksum is always calculated
	//    if the operation supports it, regardless of whether the user sets an algorithm in the request.
	//
	// 2. RequestChecksumCalculationWhenRequired: The checksum is only calculated if the user
	//    explicitly sets a checksum algorithm in the request.
	//
	// This setting is sourced from the environment variable AWS_REQUEST_CHECKSUM_CALCULATION
	// or the shared config profile attribute "request_checksum_calculation".
	RequestChecksumCalculation RequestChecksumCalculation

	// ResponseChecksumValidation determines when response checksum validation is performed
	//
	// There are two possible values for this setting:
	//
	// 1. ResponseChecksumValidationWhenSupported (default): The checksum is always validated
	//    if the operation supports it, regardless of whether the user sets the validation mode to ENABLED in request.
	//
	// 2. ResponseChecksumValidationWhenRequired: The checksum is only validated if the user
	//    explicitly sets the validation mode to ENABLED in the request
	// This variable is sourced from environment variable AWS_RESPONSE_CHECKSUM_VALIDATION or
	// the shared config profile attribute "response_checksum_validation".
	ResponseChecksumValidation ResponseChecksumValidation

	// Registry of HTTP interceptors.
	Interceptors smithyhttp.InterceptorRegistry

	// Priority list of preferred auth scheme IDs.
	AuthSchemePreference []string

	// ServiceOptions provides service specific configuration options that will be applied
	// when constructing clients for specific services. Each callback function receives the service ID
	// and the service's Options struct, allowing for dynamic configuration based on the service.
	ServiceOptions []func(string, any)
}
Конфигурация предоставляет конфигурацию службы для клиентов службы.

функционирующий NewConfig ¶
func NewConfig() *Config
NewConfig возвращает новый указатель Config, который может быть связан с построителем для установки нескольких значений конфигурации в строке без использования указателей.

func (Config) Копирование ¶
func (c Config) Copy() Config
Copy вернет поверхностную копию объекта Config.

тип CredentialProviderSource ¶
Добавлено в v1.36.3
type CredentialProviderSource interface {
	ProviderSources() []CredentialSource
}
CredentialProviderSource позволяет любому поставщику учетных данных отслеживать Все поставщики, у которых был найден поставщик учетных данных. Например, если учетные данные поступили от Вызов роли, указанной в профиле, этот метод выдаст весь хлебный след

type CredentialSource ¶
Добавлено в v1.36.3
type CredentialSource int
CredentialSource — это источник поставщика учетных данных. Поставщик может иметь несколько источников учетных данных: например, поставщик, который считывает профиль, вызывает ECS получить учетные данные, а затем взять на себя роль с помощью STS, будет иметь все это как часть своей цепочки поставщиков.

const (
	// CredentialSourceUndefined is the sentinel zero value
	CredentialSourceUndefined CredentialSource = iota
	// CredentialSourceCode credentials resolved from code, cli parameters, session object, or client instance
	CredentialSourceCode
	// CredentialSourceEnvVars credentials resolved from environment variables
	CredentialSourceEnvVars
	// CredentialSourceEnvVarsSTSWebIDToken credentials resolved from environment variables for assuming a role with STS using a web identity token
	CredentialSourceEnvVarsSTSWebIDToken
	// CredentialSourceSTSAssumeRole credentials resolved from STS using AssumeRole
	CredentialSourceSTSAssumeRole
	// CredentialSourceSTSAssumeRoleSaml credentials resolved from STS using assume role with SAML
	CredentialSourceSTSAssumeRoleSaml
	// CredentialSourceSTSAssumeRoleWebID credentials resolved from STS using assume role with web identity
	CredentialSourceSTSAssumeRoleWebID
	// CredentialSourceSTSFederationToken credentials resolved from STS using a federation token
	CredentialSourceSTSFederationToken
	// CredentialSourceSTSSessionToken credentials resolved from STS using a session token 	S
	CredentialSourceSTSSessionToken
	// CredentialSourceProfile  credentials resolved from a config file(s) profile with static credentials
	CredentialSourceProfile
	// CredentialSourceProfileSourceProfile credentials resolved from a source profile in a config file(s) profile
	CredentialSourceProfileSourceProfile
	// CredentialSourceProfileNamedProvider credentials resolved from a named provider in a config file(s) profile (like EcsContainer)
	CredentialSourceProfileNamedProvider
	// CredentialSourceProfileSTSWebIDToken  credentials resolved from configuration for assuming a role with STS using web identity token in a config file(s) profile
	CredentialSourceProfileSTSWebIDToken
	// CredentialSourceProfileSSO credentials resolved from an SSO session in a config file(s) profile
	CredentialSourceProfileSSO
	// CredentialSourceSSO credentials resolved from an SSO session
	CredentialSourceSSO
	// CredentialSourceProfileSSOLegacy credentials resolved from an SSO session in a config file(s) profile using legacy format
	CredentialSourceProfileSSOLegacy
	// CredentialSourceSSOLegacy credentials resolved from an SSO session using legacy format
	CredentialSourceSSOLegacy
	// CredentialSourceProfileProcess credentials resolved from a process in a config file(s) profile
	CredentialSourceProfileProcess
	// CredentialSourceProcess credentials resolved from a process
	CredentialSourceProcess
	// CredentialSourceHTTP credentials resolved from an HTTP endpoint
	CredentialSourceHTTP
	// CredentialSourceIMDS credentials resolved from the instance metadata service (IMDS)
	CredentialSourceIMDS
)
тип Учетные данные ¶
type Credentials struct {
	// AWS Access key ID
	AccessKeyID string

	// AWS Secret Access Key
	SecretAccessKey string

	// AWS Session Token
	SessionToken string

	// Source of the credentials
	Source string

	// States if the credentials can expire or not.
	CanExpire bool

	// The time the credentials will expire at. Should be ignored if CanExpire
	// is false.
	Expires time.Time

	// The ID of the account for the credentials.
	AccountID string
}
Credentials – это значение учетных данных AWS для отдельных полей учетных данных.

func (Учетные данные) Срок действия истек ¶
func (v Credentials) Expired() bool
Expired возвращается, если срок действия учетных данных истек.

func (Credentials) HasKeys ¶
func (v Credentials) HasKeys() bool
HasKeys возвращает, если ключи учетных данных заданы.

тип CredentialsCache ¶
Добавлено в версии 0.25.0
type CredentialsCache struct {
	// contains filtered or unexported fields
}
CredentialsCache обеспечивает кэширование и безопасное извлечение учетных данных с параллелизмом с помощью метода retrieve поставщика.

CredentialsCache будет искать дополнительные интерфейсы на поставщике для настройки Как кэш учетных данных обрабатывает кэширование учетных данных.

HandleFailRefreshCredentialsCacheStrategy — позволяет поставщику обрабатывать Сбои при обновлении учетных данных. Это может вернуть обновленные учетные данные или попробуйте другим способом получения учетных данных.

AdjustExpiresByCredentialsCacheStrategy — позволяет поставщику настраивать способ. credentials Expires изменено. Это может изменить то, как учетные данные Срок действия корректируется в зависимости от параметра CredentialsCache ExpiryWindow. Например, предоставление нижнего предела, чтобы не уменьшать срок действия ниже.

func NewCredentialsCache ¶
Добавлено в версии 0.31.0
func NewCredentialsCache(provider CredentialsProvider, optFns ...func(options *CredentialsCacheOptions)) *CredentialsCache
NewCredentialsCache возвращает CredentialsCache, который является оболочкой поставщика. Поставщик ожидается, что он не будет равен нулю. Вариативный список из одной или нескольких функций может быть следующим для изменения конфигурации CredentialsCache. Это позволяет Настройка окна истечения срока действия учетных данных и джиттера.

func (*CredentialsCache) Аннулировать ¶
Добавлено в версии 0.25.0
func (p *CredentialsCache) Invalidate()
Команда Invalidate аннулирует кэшированные учетные данные. Следующий вызов команды Retrieve приведет к вызову метода Retrieve поставщика.

func (*CredentialsCache) IsCredentialsProvider ¶
Добавлено в версии 1.17.0
func (p *CredentialsCache) IsCredentialsProvider(target CredentialsProvider) bool
IsCredentialsProvider возвращает, является ли поставщик учетных данных, обернутый в CredentialsCache соответствует типу целевого провайдера.

func (*CredentialsCache) ProviderSources ¶
Добавлено в v1.36.3
func (p *CredentialsCache) ProviderSources() []CredentialSource
ProviderSources возвращает список мест, где базовый поставщик учетных данных был найден источник, если таковой имеется. Возвращает пустое значение, если поставщик не реализует Интерфейс

func (*CredentialsCache) Получить ¶
Добавлено в версии 0.25.0
func (p *CredentialsCache) Retrieve(ctx context.Context) (Credentials, error)
Retrieve возвращает учетные данные. Если учетные данные уже были Полученные, но не просроченные, будут возвращены кэшированные учетные данные. Если метод учетные данные еще не были восстановлены или срок действия Retrieve поставщика будет вызван метод.

Возвращает и error, если метод retrieve поставщика возвращает ошибку.

type CredentialsCacheOptions ¶
Добавлено в версии 0.31.0
type CredentialsCacheOptions struct {

	// ExpiryWindow will allow the credentials to trigger refreshing prior to
	// the credentials actually expiring. This is beneficial so race conditions
	// with expiring credentials do not cause request to fail unexpectedly
	// due to ExpiredTokenException exceptions.
	//
	// An ExpiryWindow of 10s would cause calls to IsExpired() to return true
	// 10 seconds before the credentials are actually expired. This can cause an
	// increased number of requests to refresh the credentials to occur.
	//
	// If ExpiryWindow is 0 or less it will be ignored.
	ExpiryWindow time.Duration

	// ExpiryWindowJitterFrac provides a mechanism for randomizing the
	// expiration of credentials within the configured ExpiryWindow by a random
	// percentage. Valid values are between 0.0 and 1.0.
	//
	// As an example if ExpiryWindow is 60 seconds and ExpiryWindowJitterFrac
	// is 0.5 then credentials will be set to expire between 30 to 60 seconds
	// prior to their actual expiration time.
	//
	// If ExpiryWindow is 0 or less then ExpiryWindowJitterFrac is ignored.
	// If ExpiryWindowJitterFrac is 0 then no randomization will be applied to the window.
	// If ExpiryWindowJitterFrac < 0 the value will be treated as 0.
	// If ExpiryWindowJitterFrac > 1 the value will be treated as 1.
	ExpiryWindowJitterFrac float64
}
CredentialsCacheOptions — это опции

тип CredentialsProvider ¶
type CredentialsProvider interface {
	// Retrieve returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve(ctx context.Context) (Credentials, error)
}
CredentialsProvider — это интерфейс для любого компонента, который будет предоставлять credentials Credentials. CredentialsProvider необходим для управления собственным Просроченное состояние, а что значит быть просроченным.

Реализация поставщика учетных данных может быть упакована в CredentialCache , чтобы кэшировать полученное значение учетных данных. Без кэша SDK будет Попытайтесь получить учетные данные для каждого запроса.

введите CredentialsProviderFunc ¶
Добавлено в версии 0.25.0
type CredentialsProviderFunc func(context.Context) (Credentials, error)
CredentialsProviderFunc предоставляет вспомогательную систему, оборачивающую значение функции в удовлетворять требованиям интерфейса CredentialsProvider.

func (CredentialsProviderFunc) Получить ¶
Добавлено в версии 0.25.0
func (fn CredentialsProviderFunc) Retrieve(ctx context.Context) (Credentials, error)
Извлекает делегаты в значение функции, которое обертывает CredentialsProviderFunc.

type DefaultsMode ¶
Добавлено в версии 1.13.0
type DefaultsMode string
DefaultsMode — это настройка режима SDK по умолчанию.

const (
	// DefaultsModeAuto is an experimental mode that builds on the standard mode.
	// The SDK will attempt to discover the execution environment to determine the
	// appropriate settings automatically.
	//
	// Note that the auto detection is heuristics-based and does not guarantee 100%
	// accuracy. STANDARD mode will be used if the execution environment cannot
	// be determined. The auto detection might query EC2 Instance Metadata service
	// (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html),
	// which might introduce latency. Therefore we recommend choosing an explicit
	// defaults_mode instead if startup latency is critical to your application
	DefaultsModeAuto DefaultsMode = "auto"

	// DefaultsModeCrossRegion builds on the standard mode and includes optimization
	// tailored for applications which call AWS services in a different region
	//
	// Note that the default values vended from this mode might change as best practices
	// may evolve. As a result, it is encouraged to perform tests when upgrading
	// the SDK
	DefaultsModeCrossRegion DefaultsMode = "cross-region"

	// DefaultsModeInRegion builds on the standard mode and includes optimization
	// tailored for applications which call AWS services from within the same AWS
	// region
	//
	// Note that the default values vended from this mode might change as best practices
	// may evolve. As a result, it is encouraged to perform tests when upgrading
	// the SDK
	DefaultsModeInRegion DefaultsMode = "in-region"

	// DefaultsModeLegacy provides default settings that vary per SDK and were used
	// prior to establishment of defaults_mode
	DefaultsModeLegacy DefaultsMode = "legacy"

	// DefaultsModeMobile builds on the standard mode and includes optimization
	// tailored for mobile applications
	//
	// Note that the default values vended from this mode might change as best practices
	// may evolve. As a result, it is encouraged to perform tests when upgrading
	// the SDK
	DefaultsModeMobile DefaultsMode = "mobile"

	// DefaultsModeStandard provides the latest recommended default values that
	// should be safe to run in most scenarios
	//
	// Note that the default values vended from this mode might change as best practices
	// may evolve. As a result, it is encouraged to perform tests when upgrading
	// the SDK
	DefaultsModeStandard DefaultsMode = "standard"
)
Константы DefaultsMode.

func (*DefaultsMode) SetFromString ¶
Добавлено в версии 1.13.0
func (d *DefaultsMode) SetFromString(v string) (ok bool)
SetFromString задает значение DefaultsMode в одну из предопределенных констант, которая соответствует предоставленную строку при сравнении с помощью EqualFold. Если значение не совпадает с известным Constant ему будет присвоено значение as-is, и функция вернет false. В качестве частного случая, если метод если value является строкой нулевой длины, режим будет установлен в LegacyDefaultsMode.

тип DualStackEndpointState ¶
Добавлено в v1.11.0
type DualStackEndpointState uint
DualStackEndpointState — это константа для описания поведения разрешения конечной точки с двойным стеком.

const (
	// DualStackEndpointStateUnset is the default value behavior for dual-stack endpoint resolution.
	DualStackEndpointStateUnset DualStackEndpointState = iota

	// DualStackEndpointStateEnabled enables dual-stack endpoint resolution for service endpoints.
	DualStackEndpointStateEnabled

	// DualStackEndpointStateDisabled disables dual-stack endpoint resolution for endpoints.
	DualStackEndpointStateDisabled
)
func GetUseDualStackEndpoint ¶
Добавлено в v1.11.0
func GetUseDualStackEndpoint(options ...interface{}) (value DualStackEndpointState, found bool)
GetUseDualStackEndpoint принимает EndpointResolverOptions службы и возвращает значение UseDualStackEndpoint. Возвращает логическое значение false, если предоставленные параметры не имеют метода для получения DualStackEndpointState.

тип
Конечная точка
Устаревшие
тип EndpointDiscoveryEnableState ¶
Добавлено в версии 1.7.0
type EndpointDiscoveryEnableState uint
EndpointDiscoveryEnableState указывает, является ли обнаружение конечной точки включено, выключено, авто или неустановленное состояние.

Поведение по умолчанию (Авто или Сброс) указывает на операции, для которых требуется конечная точка По умолчанию при обнаружении будет использоваться обнаружение конечных точек. Операции, которые при необходимости использовать Endpoint Discovery не будет использовать Endpoint Discovery если EndpointDiscovery не включен явным образом.

const (
	// EndpointDiscoveryUnset represents EndpointDiscoveryEnableState is unset.
	// Users do not need to use this value explicitly. The behavior for unset
	// is the same as for EndpointDiscoveryAuto.
	EndpointDiscoveryUnset EndpointDiscoveryEnableState = iota

	// EndpointDiscoveryAuto represents an AUTO state that allows endpoint
	// discovery only when required by the api. This is the default
	// configuration resolved by the client if endpoint discovery is neither
	// enabled or disabled.
	EndpointDiscoveryAuto // default state

	// EndpointDiscoveryDisabled indicates client MUST not perform endpoint
	// discovery even when required.
	EndpointDiscoveryDisabled

	// EndpointDiscoveryEnabled indicates client MUST always perform endpoint
	// discovery if supported for the operation.
	EndpointDiscoveryEnabled
)
Значения перечисления для EndpointDiscoveryEnableState

введите EndpointNotFoundError ¶
Добавлено в версии 0.25.0
type EndpointNotFoundError struct {
	Err error
}
EndpointNotFoundError — это сигнальная ошибка, указывающая на то, что Реализация EndpointResolver не смогла разрешить конечную точку для Данная услуга и регион. Резолверы должны использовать это для указания на то, что API Клиент должен откатиться назад и попытаться использовать свой внутренний преобразователь по умолчанию, чтобы Разрешите конечную точку.

func (*EndpointNotFoundError) Ошибка ¶
Добавлено в версии 0.25.0
func (e *EndpointNotFoundError) Error() string
Ошибка — это сообщение об ошибке.

func (*EndpointNotFoundError) Развернуть ¶
Добавлено в версии 0.25.0
func (e *EndpointNotFoundError) Unwrap() error
Unwrap возвращает основную ошибку.

тип
EndpointResolver
Устаревшие
type EndpointResolver interface {
	ResolveEndpoint(service, region string) (Endpoint, error)
}
EndpointResolver — это резолвер конечных точек, который может быть использован для предоставления или Переопределить конечную точку для данной службы и региона. Клиенты API попытаться сначала использовать EndpointResolver для разрешения конечной точки, если доступный. Если EndpointResolver возвращает ошибку EndpointNotFoundError, Клиенты API будут возвращаться к попыткам разрешить конечную точку с помощью его Внутренний сопоставитель конечных точек по умолчанию.

Устарело: Глобальный интерфейс разрешения конечных точек является устаревшим. The API Для конечной точки разрешение теперь уникально для каждой службы и задается с помощью команды EndpointResolverV2 в параметрах клиента службы. Установка значения для EndpointResolver на aws. Конфигурация или опции клиента сервиса помешают вам от использования любых функций, связанных с конечными точками, выпущенных после внедрение EndpointResolverV2. Вы также можете столкнуться с поломкой или Непредвиденное поведение при использовании старого глобального интерфейса со службами, которые используйте множество настроек, связанных с конечными точками, таких как S3.

тип
EndpointResolverFunc
Устаревшие
type EndpointResolverFunc func(service, region string) (Endpoint, error)
EndpointResolverFunc является оболочкой функции, удовлетворяющей интерфейсу EndpointResolver.

Устарело: Глобальный интерфейс разрешения конечных точек является устаревшим. Видеть Документация по прекращению поддержки в EndpointResolver.

func (EndpointResolverFunc) ResolveEndpoint ¶
func (e EndpointResolverFunc) ResolveEndpoint(service, region string) (Endpoint, error)
ResolveEndpoint вызывает функцию wrapped и возвращает результаты.

тип
EndpointResolverWithOptions
Устаревшие
Добавлено в v1.11.0
тип
EndpointResolverWithOptionsFunc
Устаревшие
Добавлено в v1.11.0
тип
EndpointSource
Устаревшие
Добавлено в версии 1.1.0
тип ExecutionEnvironmentID ¶
Добавлено в версии 1.13.0
type ExecutionEnvironmentID string
ExecutionEnvironmentID – это идентификатор среды выполнения AWS.

тип FIPSEndpointState ¶
Добавлено в v1.11.0
type FIPSEndpointState uint
FIPSEndpointState — это константа для описания поведения разрешения конечной точки FIPS.

const (
	// FIPSEndpointStateUnset is the default value behavior for FIPS endpoint resolution.
	FIPSEndpointStateUnset FIPSEndpointState = iota

	// FIPSEndpointStateEnabled enables FIPS endpoint resolution for service endpoints.
	FIPSEndpointStateEnabled

	// FIPSEndpointStateDisabled disables FIPS endpoint resolution for endpoints.
	FIPSEndpointStateDisabled
)
func GetUseFIPSEndpoint ¶
Добавлено в v1.11.0
func GetUseFIPSEndpoint(options ...interface{}) (value FIPSEndpointState, found bool)
GetUseFIPSEndpoint принимает EndpointResolverOptions службы и возвращает значение UseDualStackEndpoint. Возвращает логическое значение false, если предоставленные параметры не имеют метода для получения DualStackEndpointState.

тип HTTPClient ¶
Добавлено в версии 0.10.0
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
HTTPClient предоставляет интерфейс для предоставления пользовательских HTTPClients. Вообще *http. Клиента достаточно для большинства случаев использования. HTTPClient не должен Перейдите по 301 или 302 редиректам.

тип HandleFailRefreshCredentialsCacheStrategy ¶
Добавлено в версии 1.16.0
type HandleFailRefreshCredentialsCacheStrategy interface {
	// Given the previously cached Credentials, if any, and refresh error, may
	// returns new or modified set of Credentials, or error.
	//
	// Credential caches may use default implementation if nil.
	HandleFailToRefresh(context.Context, Credentials, error) (Credentials, error)
}
HandleFailRefreshCredentialsCacheStrategy — это интерфейс для CredentialsCache, чтобы разрешить CredentialsProvider как не удалось обновить Учетные данные обрабатываются.

тип MissingRegionError ¶
Добавлено в версии 0.21.0
type MissingRegionError struct{}
MissingRegionError — это ошибка, которая возвращается, если конфигурация региона Значение не найдено.

func (*MissingRegionError) Ошибка ¶
Добавлено в версии 0.21.0
func (*MissingRegionError) Error() string
тип NopRetryer ¶
Добавлено в версии 0.31.0
type NopRetryer struct{}
NopRetryer предоставляет реализацию RequestRetryDecider, которая будет помечать Все ошибки попыток не подлежат повторению, максимальное количество попыток равно 1.

func (NopRetryer) GetAttemptToken ¶
Добавлено в v1.14.0
func (NopRetryer) GetAttemptToken(context.Context) (func(error) error, error)
GetAttemptToken возвращает функцию-заглушку, которая ничего не делает.

func (NopRetryer) GetInitialToken ¶
Добавлено в версии 0.31.0
func (NopRetryer) GetInitialToken() func(error) error
GetInitialToken возвращает функцию-заглушку, которая ничего не делает.

func (NopRetryer) GetRetryToken ¶
Добавлено в версии 0.31.0
func (NopRetryer) GetRetryToken(context.Context, error) (func(error) error, error)
GetRetryToken возвращает функцию-заглушку, которая ничего не делает.

func (NopRetryer) IsErrorRetryable ¶
Добавлено в версии 0.31.0
func (NopRetryer) IsErrorRetryable(error) bool
IsErrorRetryable возвращает false для всех значений ошибок.

func (NopRetryer) MaxAttempts ¶
Добавлено в версии 0.31.0
func (NopRetryer) MaxAttempts() int
MaxAttempts всегда возвращает 1 для исходной попытки.

func (NopRetryer) RetryDelay ¶
Добавлено в версии 0.31.0
func (NopRetryer) RetryDelay(int, error) (time.Duration, error)
RetryDelay недействителен для NopRetryer. Всегда будет возвращать ошибку.

type RequestCanceledError ¶
Добавлено в версии 0.20.0
type RequestCanceledError struct {
	Err error
}
RequestCanceledError — это ошибка, которая будет возвращена запросом API Это было отменено. Запросы с заданным контекстом могут возвращать эту ошибку, если аннулированный.

func (*RequestCanceledError) CanceledError ¶
Добавлено в версии 0.20.0
func (*RequestCanceledError) CanceledError() bool
CanceledError returns true to satisfy interfaces checking for canceled errors.

func (*RequestCanceledError) Error ¶
added in v0.20.0
func (e *RequestCanceledError) Error() string
func (*RequestCanceledError) Unwrap ¶
added in v0.20.0
func (e *RequestCanceledError) Unwrap() error
Unwrap returns the underlying error, if there was one.

type RequestChecksumCalculation ¶
added in v1.33.0
type RequestChecksumCalculation int
RequestChecksumCalculation controls request checksum calculation workflow

const (
	// RequestChecksumCalculationUnset is the unset value for RequestChecksumCalculation
	RequestChecksumCalculationUnset RequestChecksumCalculation = iota

	// RequestChecksumCalculationWhenSupported indicates request checksum will be calculated
	// if the operation supports input checksums
	RequestChecksumCalculationWhenSupported

	// RequestChecksumCalculationWhenRequired indicates request checksum will be calculated
	// if required by the operation or if user elects to set a checksum algorithm in request
	RequestChecksumCalculationWhenRequired
)
type ResponseChecksumValidation ¶
added in v1.33.0
type ResponseChecksumValidation int
ResponseChecksumValidation controls response checksum validation workflow

const (
	// ResponseChecksumValidationUnset is the unset value for ResponseChecksumValidation
	ResponseChecksumValidationUnset ResponseChecksumValidation = iota

	// ResponseChecksumValidationWhenSupported indicates response checksum will be validated
	// if the operation supports output checksums
	ResponseChecksumValidationWhenSupported

	// ResponseChecksumValidationWhenRequired indicates response checksum will only
	// be validated if the operation requires output checksum validation
	ResponseChecksumValidationWhenRequired
)
type RetryMode ¶
added in v1.14.0
type RetryMode string
RetryMode provides the mode the API client will use to create a retryer based on.

const (
	// RetryModeStandard model provides rate limited retry attempts with
	// exponential backoff delay.
	RetryModeStandard RetryMode = "standard"

	// RetryModeAdaptive model provides attempt send rate limiting on throttle
	// responses in addition to standard mode's retry rate limiting.
	//
	// Adaptive retry mode is experimental and is subject to change in the
	// future.
	RetryModeAdaptive RetryMode = "adaptive"
)
func ParseRetryMode ¶
added in v1.14.0
func ParseRetryMode(v string) (mode RetryMode, err error)
ParseRetryMode attempts to parse a RetryMode from the given string. Returning error if the value is not a known RetryMode.

func (RetryMode) String ¶
added in v1.14.0
func (m RetryMode) String() string
type Retryer ¶
type Retryer interface {
	// IsErrorRetryable returns if the failed attempt is retryable. This check
	// should determine if the error can be retried, or if the error is
	// terminal.
	IsErrorRetryable(error) bool

	// MaxAttempts returns the maximum number of attempts that can be made for
	// an attempt before failing. A value of 0 implies that the attempt should
	// be retried until it succeeds if the errors are retryable.
	MaxAttempts() int

	// RetryDelay returns the delay that should be used before retrying the
	// attempt. Will return error if the delay could not be determined.
	RetryDelay(attempt int, opErr error) (time.Duration, error)

	// GetRetryToken attempts to deduct the retry cost from the retry token pool.
	// Returning the token release function, or error.
	GetRetryToken(ctx context.Context, opErr error) (releaseToken func(error) error, err error)

	// GetInitialToken returns the initial attempt token that can increment the
	// retry token pool if the attempt is successful.
	GetInitialToken() (releaseToken func(error) error)
}
Retryer is an interface to determine if a given error from a attempt should be retried, and if so what backoff delay to apply. The default implementation used by most services is the retry package's Standard type. Which contains basic retry logic using exponential backoff.

type RetryerV2 ¶
added in v1.14.0
type RetryerV2 interface {
	Retryer

	// GetInitialToken returns the initial attempt token that can increment the
	// retry token pool if the attempt is successful.
	//
	// Deprecated: This method does not provide a way to block using Context,
	// nor can it return an error. Use RetryerV2, and GetAttemptToken instead.
	GetInitialToken() (releaseToken func(error) error)

	// GetAttemptToken returns the send token that can be used to rate limit
	// attempt calls. Will be used by the SDK's retry package's Attempt
	// middleware to get a send token prior to calling the temp and releasing
	// the send token after the attempt has been made.
	GetAttemptToken(context.Context) (func(error) error, error)
}
RetryerV2 is an interface to determine if a given error from an attempt should be retried, and if so what backoff delay to apply. The default implementation used by most services is the retry package's Standard type. Which contains basic retry logic using exponential backoff.

RetryerV2 replaces the Retryer interface, deprecating the GetInitialToken method in favor of GetAttemptToken which takes a context, and can return an error.

The SDK's retry package's Attempt middleware, and utilities will always wrap a Retryer as a RetryerV2. Delegating to GetInitialToken, only if GetAttemptToken is not implemented.

type RuntimeEnvironment ¶
added in v1.13.0
type RuntimeEnvironment struct {
	EnvironmentIdentifier     ExecutionEnvironmentID
	Region                    string
	EC2InstanceMetadataRegion string
}
RuntimeEnvironment is a collection of values that are determined at runtime based on the environment that the SDK is executing in. Some of these values may or may not be present based on the executing environment and certain SDK configuration properties that drive whether these values are populated..

type Ternary ¶
added in v0.20.0
type Ternary int
Ternary is an enum allowing an unknown or none state in addition to a bool's true and false.

const (
	UnknownTernary Ternary = iota
	FalseTernary
	TrueTernary
)
Enumerations for the values of the Ternary type.

func BoolTernary ¶
added in v0.20.0
func BoolTernary(v bool) Ternary
BoolTernary returns a true or false Ternary value for the bool provided.

func (Ternary) Bool ¶
added in v0.20.0
func (t Ternary) Bool() bool
Bool returns true if the value is TrueTernary, false otherwise.

func (Ternary) String ¶
added in v0.20.0
func (t Ternary) String() string
 Source Files ¶
View all Source files
accountid_endpoint_mode.go
checksum.go
config.go
context.go
credential_cache.go
credentials.go
defaultsmode.go
doc.go
endpoints.go
errors.go
from_ptr.go
go_module_metadata.go
logging.go
request.go
retryer.go
runtime.go
to_ptr.go
types.go
version.go
 Directories ¶
Show internal
Collapse all
arn
Package arn provides a parser for interacting with Amazon Resource Names.
defaults
Package defaults provides recommended configuration values for AWS SDKs and CLIs.
middleware
protocol
ec2query
query
restjson
xml
eventstream Module
ratelimit
retry
Package retry provides interfaces and implementations for SDK request retry behavior.
signer
v4
Package v4 implements the AWS signature version 4 algorithm (commonly known as SigV4).
transport
http
Why Go
Use Cases
Case Studies
Get Started
Playground
Tour
Stack Overflow
Help
Packages
Standard Library
Sub-repositories
About Go Packages
About
Download
Blog
Issue Tracker
Release Notes
Brand Guidelines
Code of Conduct
Connect
Twitter
GitHub
Slack
r/golang
Meetup
Golang Weekly
Gopher in flight goggles
Copyright
Terms of Service
Privacy Policy
Report an Issue
System theme
Theme Toggle


Shortcuts Modal

Google logo
go.dev использует файлы cookie от Google для предоставления и повышения качества своих услуг, а также для Анализируйте трафик. Подробнее.
Хорошо