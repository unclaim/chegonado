package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// AppConfig содержит все настройки приложения.
type AppConfig struct {
	App              App              `yaml:"app"`
	Server           Server           `yaml:"server"`
	Database         Database         `yaml:"database"`
	Cache            Cache            `yaml:"cache"`
	Logging          Logging          `yaml:"logging"`
	Monitoring       Monitoring       `yaml:"monitoring"`
	API              API              `yaml:"api"`
	ExternalServices ExternalServices `yaml:"external_services"`
	Security         Security         `yaml:"security"`
	Deployment       Deployment       `yaml:"deployment"`
	FileStorage      FileStorage      `yaml:"file_storage"`
	SMTPConfig       *SMTPConfig      `yaml:"smtp_config"`
}

// SMTPConfig содержит настройки для SMTP-сервера.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
}

// App содержит основные параметры приложения.
type App struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
	ID          string `yaml:"id"`
	Debug       bool   `yaml:"debug"`
}

// Server содержит параметры HTTP-сервера.
type Server struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	ReadTimeout      string `yaml:"read_timeout"`
	WriteTimeout     string `yaml:"write_timeout"`
	IdleTimeout      string `yaml:"idle_timeout"`
	MaxHeaderSize    int    `yaml:"max_header_size"`
	KeepaliveTimeout string `yaml:"keepalive_timeout"`
}

// Database содержит параметры подключения к базе данных.
type Database struct {
	Type              string `yaml:"type"`
	URL               string `yaml:"url"`
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	User              string `yaml:"user"`
	Password          string `yaml:"password"`
	Name              string `yaml:"name"`
	MaxConnections    int    `yaml:"max_connections"`
	ConnectionTimeout string `yaml:"connection_timeout"`
	SSLMode           string `yaml:"sslmode"`
}

// Cache содержит параметры для кэширования.
type Cache struct {
	Enabled  bool   `yaml:"enabled"`
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TTL      string `yaml:"ttl"`
}

// Logging содержит параметры для логирования.
type Logging struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
	File   struct {
		Path      string `yaml:"path"`
		Retention string `yaml:"retention"`
		Rotation  struct {
			MaxSize    int `yaml:"max_size"`
			MaxBackups int `yaml:"max_backups"`
		} `yaml:"rotation"`
	} `yaml:"file"`
	Format string `yaml:"format"`
}

// Monitoring содержит параметры для мониторинга.
type Monitoring struct {
	Enabled     bool        `yaml:"enabled"`
	Prometheus  Prometheus  `yaml:"prometheus"`
	Alerts      Alerts      `yaml:"alerts"`
	HealthCheck HealthCheck `yaml:"health_check"`
}

// Prometheus содержит параметры для Prometheus.
type Prometheus struct {
	Endpoint       string `yaml:"endpoint"`
	ScrapeInterval string `yaml:"scrape_interval"`
	Timeout        string `yaml:"timeout"`
}

// Alerts содержит параметры для алертов.
type Alerts struct {
	Email Email `yaml:"email"`
}

// Email содержит параметры для отправки почты.
type Email struct {
	Enabled    bool     `yaml:"enabled"`
	Recipients []string `yaml:"recipients"`
	SMTPServer string   `yaml:"smtp_server"`
	SMTPPort   int      `yaml:"smtp_port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"smtp_password"`
	From       string   `yaml:"from"`
	FromName   string   `yaml:"from_name"`
}

// HealthCheck содержит параметры для проверки здоровья приложения.
type HealthCheck struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

// API содержит параметры для API.
type API struct {
	Version      string       `yaml:"version"`
	RateLimiting RateLimiting `yaml:"rate_limiting"`
	CORS         CORS         `yaml:"cors"`
}

// RateLimiting содержит параметры для ограничения скорости запросов.
type RateLimiting struct {
	Enabled  bool   `yaml:"enabled"`
	Limit    int    `yaml:"limit"`
	Interval string `yaml:"interval"`
}

// CORS содержит параметры для CORS.
type CORS struct {
	AllowOrigin  []string `yaml:"allow_origin"`
	AllowMethods []string `yaml:"allow_methods"`
}

// ExternalServices содержит настройки для сторонних сервисов.
type ExternalServices struct {
	PaymentGateway      ExternalService `yaml:"payment_gateway"`
	NotificationService ExternalService `yaml:"notification_service"`
}

// ExternalService содержит общие параметры для сторонних сервисов.
type ExternalService struct {
	URL     string `yaml:"url"`
	APIKey  string `yaml:"api_key"`
	Secret  string `yaml:"secret"`
	Timeout string `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// Security содержит параметры безопасности.
type Security struct {
	EnableHTTPS           bool         `yaml:"enable_https"`
	AllowedIPs            []string     `yaml:"allowed_ips"`
	APISecurity           APISecurity  `yaml:"api_security"`
	CSRFProtection        bool         `yaml:"csrf_protection"`
	ContentSecurityPolicy CSP          `yaml:"content_security_policy"`
	InputValidation       bool         `yaml:"input_validation"`
	LoggingAudit          LoggingAudit `yaml:"logging_audit"`
	JWTSecret             string       `yaml:"jwt_secret"`
	PasswordSalt          string       `yaml:"password_salt"`
	SessionSecret         string       `yaml:"session_secret"`
}

// APISecurity содержит параметры для безопасности API.
type APISecurity struct {
	EnableAPIKey bool   `yaml:"enable_api_key"`
	APIKeyParam  string `yaml:"api_key_param"`
	APIKey       string `yaml:"api_key"`
}

// CSP содержит параметры для Content Security Policy.
type CSP struct {
	Enabled bool   `yaml:"enabled"`
	Policy  string `yaml:"policy"`
}

// LoggingAudit содержит параметры для аудита.
type LoggingAudit struct {
	Enabled             bool `yaml:"enabled"`
	LogSensitiveActions bool `yaml:"log_sensitive_actions"`
}

// Deployment содержит параметры для развертывания.
type Deployment struct {
	Strategy     string       `yaml:"strategy"`
	ReplicaCount int          `yaml:"replica_count"`
	HealthChecks HealthChecks `yaml:"health_checks"`
}

// HealthChecks содержит параметры для проверок готовности и жизнеспособности.
type HealthChecks struct {
	Readiness Readiness `yaml:"readiness"`
	Liveness  Liveness  `yaml:"liveness"`
}

// Readiness содержит параметры для проверки готовности.
type Readiness struct {
	Path     string `yaml:"path"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

// Liveness содержит параметры для проверки жизнеспособности.
type Liveness struct {
	Path     string `yaml:"path"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

// FileStorage содержит параметры для файлового хранилища.
type FileStorage struct {
	Type string `yaml:"type"`
	S3   S3     `yaml:"s3"`
}

// S3 содержит параметры для S3-совместимого хранилища.
type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
}

// LoadConfig загружает конфигурацию из файла и переменных окружения.
// Переменные окружения имеют приоритет.
func LoadConfig(filename string) (*AppConfig, error) {
	var config AppConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML-файла: %w", err)
	}

	_ = godotenv.Load("../../.env")
	if AppName := os.Getenv("NAME"); AppName != "" {
		config.App.Name = AppName
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if cacheURL := os.Getenv("REDIS_URL"); cacheURL != "" {
		config.Cache.URL = cacheURL
	}
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}
	if readTimeoutStr := os.Getenv("SERVER_READ_TIMEOUT"); readTimeoutStr != "" {
		config.Server.ReadTimeout = readTimeoutStr
	}
	if writeTimeoutStr := os.Getenv("SERVER_WRITE_TIMEOUT"); writeTimeoutStr != "" {
		config.Server.WriteTimeout = writeTimeoutStr
	}
	if idleTimeoutStr := os.Getenv("SERVER_IDLE_TIMEOUT"); idleTimeoutStr != "" {
		config.Server.IdleTimeout = idleTimeoutStr
	}
	if keepaliveTimeoutStr := os.Getenv("SERVER_KEEPALIVE_TIMEOUT"); keepaliveTimeoutStr != "" {
		config.Server.KeepaliveTimeout = keepaliveTimeoutStr
	}
	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		config.Database.Port = dbPort
	}
	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		config.Database.Name = dbName
	}
	if cacheHost := os.Getenv("REDIS_HOST"); cacheHost != "" {
		config.Cache.Host = cacheHost
	}
	if cachePortStr := os.Getenv("REDIS_PORT"); cachePortStr != "" {
		if cachePort, err := strconv.Atoi(cachePortStr); err == nil {
			config.Cache.Port = cachePort
		}
	}
	if cachePassword := os.Getenv("REDIS_PASSWORD"); cachePassword != "" {
		config.Cache.Password = cachePassword
	}
	if cacheDBStr := os.Getenv("REDIS_DB"); cacheDBStr != "" {
		if cacheDB, err := strconv.Atoi(cacheDBStr); err == nil {
			config.Cache.DB = cacheDB
		}
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}
	if sessionSecret := os.Getenv("SESSION_SECRET"); sessionSecret != "" {
		config.Security.SessionSecret = sessionSecret
	}
	if passwordSalt := os.Getenv("PASSWORD_SALT"); passwordSalt != "" {
		config.Security.PasswordSalt = passwordSalt
	}
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		config.Security.APISecurity.APIKey = apiKey
	}
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.Monitoring.Alerts.Email.SMTPServer = smtpHost
	}
	if smtpPortStr := os.Getenv("SMTP_PORT"); smtpPortStr != "" {
		if smtpPort, err := strconv.Atoi(smtpPortStr); err == nil {
			config.Monitoring.Alerts.Email.SMTPPort = smtpPort
		}
	}
	if smtpUser := os.Getenv("SMTP_USERNAME"); smtpUser != "" {
		config.Monitoring.Alerts.Email.Username = smtpUser
	}
	if smtpPass := os.Getenv("SMTP_PASSWORD"); smtpPass != "" {
		config.Monitoring.Alerts.Email.Password = smtpPass
	}
	if smtpFrom := os.Getenv("SMTP_FROM"); smtpFrom != "" {
		config.Monitoring.Alerts.Email.From = smtpFrom
	}
	if smtpFromName := os.Getenv("SMTP_FROM_NAME"); smtpFromName != "" {
		config.Monitoring.Alerts.Email.FromName = smtpFromName
	}
	if paymentKey := os.Getenv("PAYMENT_API_KEY"); paymentKey != "" {
		config.ExternalServices.PaymentGateway.APIKey = paymentKey
	}
	if notificationSecret := os.Getenv("NOTIFICATION_SECRET"); notificationSecret != "" {
		config.ExternalServices.NotificationService.Secret = notificationSecret
	}

	// Настройки файлового хранилища
	if fileStorageType := os.Getenv("FILE_STORAGE_TYPE"); fileStorageType != "" {
		config.FileStorage.Type = fileStorageType
	}
	if s3Endpoint := os.Getenv("S3_ENDPOINT"); s3Endpoint != "" {
		config.FileStorage.S3.Endpoint = s3Endpoint
	}
	if s3AccessKeyID := os.Getenv("S3_ACCESS_KEY_ID"); s3AccessKeyID != "" {
		config.FileStorage.S3.AccessKeyID = s3AccessKeyID
	}
	if s3SecretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY"); s3SecretAccessKey != "" {
		config.FileStorage.S3.SecretAccessKey = s3SecretAccessKey
	}
	if s3Bucket := os.Getenv("S3_BUCKET"); s3Bucket != "" {
		config.FileStorage.S3.Bucket = s3Bucket
	}

	return &config, nil
}
