package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/unclaim/chegonado/internal/auth"
	"github.com/unclaim/chegonado/internal/auth/api"
	"github.com/unclaim/chegonado/internal/auth/domain"
	"github.com/unclaim/chegonado/internal/auth/infra"
	chatAPI "github.com/unclaim/chegonado/internal/chat/api"
	chatDomain "github.com/unclaim/chegonado/internal/chat/domain"
	chatInfra "github.com/unclaim/chegonado/internal/chat/infra"
	filestorageAPI "github.com/unclaim/chegonado/internal/filestorage/api"
	filestorageDomain "github.com/unclaim/chegonado/internal/filestorage/domain"
	filestorageInfra "github.com/unclaim/chegonado/internal/filestorage/infra"
	gamificationDomain "github.com/unclaim/chegonado/internal/gamification/domain"
	gamificationInfra "github.com/unclaim/chegonado/internal/gamification/infra"
	"github.com/unclaim/chegonado/internal/shared/config"
	tasksAPI "github.com/unclaim/chegonado/internal/tasks/api"
	tasksDomain "github.com/unclaim/chegonado/internal/tasks/domain"
	tasksInfra "github.com/unclaim/chegonado/internal/tasks/infra"
	usersAPI "github.com/unclaim/chegonado/internal/users/api"
	usersDomain "github.com/unclaim/chegonado/internal/users/domain"
	usersInfra "github.com/unclaim/chegonado/internal/users/infra"
	"github.com/unclaim/chegonado/pkg/infrastructure/email"
	"github.com/unclaim/chegonado/pkg/infrastructure/eventbus"
	"github.com/unclaim/chegonado/pkg/security/session"
	"github.com/unclaim/chegonado/pkg/security/token"
)

type AppDependencies struct {
	Config             *config.AppConfig
	DBPool             *pgxpool.Pool
	Tokens             *token.JwtToken
	SessionsManager    *session.SessionsDB
	AuthHandler        *api.AuthHandler
	UserHandler        *usersAPI.UserHandler
	TaskHandler        *tasksAPI.TasksHandler
	ChatHandler        *chatAPI.ChatHandler
	FileStorageHandler *filestorageAPI.FileStorageHandler
	Context            context.Context
}

func InitApplication(ctx context.Context, showInfo bool) (*AppDependencies, error) {
	// Мы передаём путь к файлу конфигурации как параметр.
	// Это делает InitApplication более гибким.
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	if showInfo {
		fmt.Printf("Приложение: %s, Окружение: %s, Версия: %s, ID: %s\n",
			cfg.App.Name, cfg.App.Environment, cfg.App.Version, cfg.App.ID)
	}

	// Создание клиента для отправки почты.
	emailConfig := &config.SMTPConfig{
		Host:     cfg.Monitoring.Alerts.Email.SMTPServer,
		Port:     fmt.Sprintf("%d", cfg.Monitoring.Alerts.Email.SMTPPort),
		Username: cfg.Monitoring.Alerts.Email.Username,
		Password: cfg.Monitoring.Alerts.Email.Password,
		From:     cfg.Monitoring.Alerts.Email.From,
		FromName: cfg.Monitoring.Alerts.Email.FromName,
	}
	emailSender := email.NewSMTPClient(emailConfig)

	// Подключение к базе данных. Используем DATABASE_URL, если он задан.
	dbURL := cfg.Database.URL
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)
	}
	dbpool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	// Инициализация токенов с использованием секрета из конфига.
	tokens, err := token.NewJwtToken(cfg.Security.JWTSecret)
	if err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("невозможно инициализировать токены: %w", err)
	}

	sm := session.NewSessionsDB(dbpool)

	// 1. Инициализируем Event Bus
	bus := eventbus.NewEventBus()

	// 2. Инициализируем домен "gamification" и его зависимости
	gamificationRepo := gamificationInfra.NewGamificationRepository()
	gamificationService := gamificationDomain.NewGamificationService(gamificationRepo)

	usersRepo := usersInfra.NewUsersRepository(dbpool)
	usersService := usersDomain.NewUsersService(usersRepo, emailSender, tokens, *cfg)
	userHandler := usersAPI.NewUserHandler(tokens, usersService)
	chatRepo := chatInfra.NewChatRepository(dbpool)
	chatService := chatDomain.NewChatService(chatRepo)
	chatHandler := chatAPI.NewChatHandler(chatService)

	// === Блок инициализации файлового хранилища ===
	var fileStorageRepo filestorageDomain.FileStorageRepository
	switch cfg.FileStorage.Type {
	case "s3":
		fileStorageRepo, err = filestorageInfra.NewS3Repository(cfg)
		if err != nil {
			dbpool.Close()
			return nil, fmt.Errorf("не удалось инициализировать S3-репозиторий: %w", err)
		}
		slog.Info("Используется S3-хранилище для файлов")
	case "local":
		fileStorageRepo = filestorageInfra.NewLocalRepository()
		slog.Info("Используется локальное хранилище для файлов")
	default:
		dbpool.Close()
		return nil, fmt.Errorf("неизвестный тип файлового хранилища: %s", cfg.FileStorage.Type)
	}

	fileStorageService := filestorageDomain.NewService(fileStorageRepo, dbpool)
	fileStorageHandlers := filestorageAPI.NewFileStorageHandler(fileStorageService)
	// ===========================================

	tasksRepo := tasksInfra.NewTasksRepository(dbpool)
	tasksService := tasksDomain.NewTasksService(tasksRepo)
	tasksHandler := tasksAPI.NewTasksHandler(tasksService, tokens)

	authRepo := infra.NewAuthRepository(dbpool, fileStorageService)
	authService := domain.NewAuthService(authRepo, sm, emailSender, config.AppConfig{}, bus)
	authHandler := api.NewAuthHandler(authService)
	// ===========================================
	// САМЫЙ ВАЖНЫЙ ШАГ: РЕГИСТРАЦИЯ ОБРАБОТЧИКОВ!
	// ===========================================
	// Мы регистрируем два разных обработчика для одного и того же события.

	bus.Subscribe(auth.UserRegisteredEvent{}, func(event eventbus.Event) {
		if e, ok := event.(auth.UserRegisteredEvent); ok {
			gamificationService.HandleUserRegistered(e)
		}
	})
	return &AppDependencies{
		Config:             cfg,
		DBPool:             dbpool,
		Tokens:             tokens,
		SessionsManager:    sm,
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		TaskHandler:        tasksHandler,
		ChatHandler:        chatHandler,
		FileStorageHandler: fileStorageHandlers,
		Context:            ctx,
	}, nil
}

func StartServer(cfg *config.AppConfig, quit chan os.Signal, wait time.Duration, ctx context.Context, handler http.Handler) {
	server := &http.Server{
		Handler:        handler,
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		ReadTimeout:    parseDuration(cfg.Server.ReadTimeout),
		WriteTimeout:   parseDuration(cfg.Server.WriteTimeout),
		IdleTimeout:    parseDuration(cfg.Server.IdleTimeout),
		MaxHeaderBytes: cfg.Server.MaxHeaderSize,
	}

	go func() {
		fmt.Printf("Запуск сервера на http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка запуска сервера: %s\n", err)
		}
	}()

	<-quit

	ctxShutdown, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %s\n", err)
	} else {
		log.Println("Сервер успешно завершил работу.")
	}
}

// Вспомогательные функции для парсинга
func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatalf("Некорректный формат длительности: %v", err)
	}
	return d
}
