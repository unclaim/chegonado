package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/unclaim/chegonado.git/docs"
	"github.com/unclaim/chegonado.git/internal"
)

// Информация о сервере
// @title Unclaimeds API
// @version 1.0
// @description Сервис помогает искать специалистов различных профессий...

// @host localhost:8585
// @BasePath /api
// Основная точка входа
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var wait = time.Second * 15
	var showInfo bool

	flag.BoolVar(&showInfo, "show-info", false, "Показать информацию о приложении")
	flag.Parse()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	ctx := context.Background()

	// Инициализируем приложение, передавая путь к файлу конфигурации.
	// Теперь InitApplication загрузит всё необходимое.
	deps, err := internal.InitApplication(ctx, showInfo)
	if err != nil {
		log.Fatalf("Ошибка инициализации приложения: %v\n", err)
	}
	// Убедимся, что пул подключений к БД закроется при завершении работы приложения
	defer deps.DBPool.Close()

	if !deps.Config.Monitoring.HealthCheck.Enabled {
		log.Println("Health check отключён.")
	}

	// Настраиваем маршруты, передавая зависимости
	rootHandler := internal.SetupRoutes(
		deps.AuthHandler,
		deps.UserHandler,
		deps.TaskHandler,
		deps.FileStorageHandler,
		deps.ChatHandler,
		deps.SessionsManager,
		deps.Context,
	)
	log.Println("Запуск сервера")
	// Запускаем сервер
	internal.StartServer(deps.Config, quit, wait, deps.Context, rootHandler)
}
