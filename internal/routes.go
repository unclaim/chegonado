// internal/routes.go
package internal

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/unclaim/chegonado.git/docs" // Импортируем документацию Swagger
	"github.com/unclaim/chegonado.git/internal/auth/api"
	chatAPI "github.com/unclaim/chegonado.git/internal/chat/api"
	filestorageAPI "github.com/unclaim/chegonado.git/internal/filestorage/api"
	tasksAPI "github.com/unclaim/chegonado.git/internal/tasks/api"
	usersAPI "github.com/unclaim/chegonado.git/internal/users/api"
	"github.com/unclaim/chegonado.git/pkg/index"
	"github.com/unclaim/chegonado.git/pkg/private"
	"github.com/unclaim/chegonado.git/pkg/public"
	"github.com/unclaim/chegonado.git/pkg/security/session"
)

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "Total number of HTTP requests"}, []string{"method"})

func init() {
	prometheus.MustRegister(requestCount)
}

// SetupRoutes настраивает все HTTP-маршруты приложения
func SetupRoutes(ah *api.AuthHandler, uh *usersAPI.UserHandler, th *tasksAPI.TasksHandler, fs *filestorageAPI.FileStorageHandler, ch *chatAPI.ChatHandler, sessionsManager *session.SessionsDB, ctx context.Context) http.Handler {
	mux := http.NewServeMux()

	// Обновление email адреса пользователя
	mux.HandleFunc("POST /account/update_email", uh.AccountUpdateEmailHandler)

	// Работа с профилем социальных сетей пользователя
	mux.HandleFunc("/account/social_profiles", uh.SocialProfileHandler)

	apiMux := http.NewServeMux()

	// Email для регистрации
	apiMux.HandleFunc("POST /auth/signup/send-code", ah.SendEmailCodeForSignupHandler)
	// Email для регистрации
	apiMux.HandleFunc("POST /auth/signup/verify-code", ah.VerifyEmailCodeForSignupHandler)

	apiMux.HandleFunc("POST /auth/login/send-code", ah.SendEmailCodeForLoginHandler)
	apiMux.HandleFunc("POST /auth/login/verify-code", ah.VerifyEmailCodeForLoginHandler)
	// Получает активные сессии пользователя
	apiMux.HandleFunc("GET /account/sessions", ah.GetActiveUserSessionsHandler)
	apiMux.HandleFunc("POST /auth/resend-code", ah.ResendVerificationCodeHandler)
	// Аннулирует активную сессию пользователя
	apiMux.HandleFunc("POST /account/sessions/revoke", ah.RevokeSessionHandler)

	// Авторизация пользователя
	apiMux.HandleFunc("POST /user/login", ah.Login)

	// Регистрация нового пользователя
	apiMux.HandleFunc("POST /user/signup", ah.Signup)

	// Выход пользователя из системы
	apiMux.HandleFunc("POST /user/logout", ah.Signout)

	// Проверка состояния сессии пользователя
	apiMux.HandleFunc("GET /check-session", ah.CheckSessionHandler)

	// Отправляет форму сброса пароля
	apiMux.HandleFunc("/user/reset_password", ah.ResetPasswordPageHandler)

	// Отправляет письмо для сброса пароля
	apiMux.HandleFunc("POST /user/send_password_reset", ah.SendPasswordResetHandler)

	// Обрабатывает изменение пароля
	apiMux.HandleFunc("POST /user/update_password", ah.ResetPasswordHandler)

	// Проверка существования пользователя
	apiMux.HandleFunc("POST /user/check-user", uh.CheckUserHandler)

	// Получение списка ботов
	apiMux.HandleFunc("GET /bots", uh.GetBotsHandler)

	// Получение отзывов о ботах
	apiMux.HandleFunc("GET /reviews_bots", uh.ReviewsBotsHandler)

	// Получает список компетенций пользователя
	apiMux.HandleFunc("GET /user/skills", uh.UserSkillsGetHandler)

	// Добавляет новый навык пользователю
	apiMux.HandleFunc("POST /user/skills", uh.UserSkillsPostHandler)

	// Предоставляет свободные задания специалистам
	apiMux.HandleFunc("GET /unclaimeds", uh.GettingOrderExecutorsHandler)

	// Получает установленные категории пользователя
	apiMux.HandleFunc("GET /unclaimeds/{id}/category", uh.GetUserCategoriesHandler)

	// Показывает профиль пользователя
	apiMux.HandleFunc("GET /profile/{id}", uh.GetUserProfileHandler)

	// Получает профиль пользователя по id
	apiMux.HandleFunc("GET /sexy/{id}", uh.GetUserByQueryHandler)

	// Получает профиль пользователя по username
	apiMux.HandleFunc("GET /username/{username}", uh.GetUserProfileDataHandler)

	// Позволяет подписаться на пользователя
	apiMux.HandleFunc("POST /profile/{id}/subscribe", uh.SubscribeHandler)

	// Позволяет отменить подписку на пользователя
	apiMux.HandleFunc("POST /profile/{id}/unsubscribe", uh.UnsubscribeHandler)

	// Отправляет сообщение пользователю
	apiMux.HandleFunc("POST /profile/{id}/send_message", ch.SendMessageHandler)

	// Блокировка пользователя другим пользователем
	apiMux.HandleFunc("POST /profile/{id}/block", uh.BlockUserHandler)

	// Разблокировка заблокированного пользователя
	apiMux.HandleFunc("POST /profile/{id}/unblock", uh.UnblockHandler)

	// Получает информацию о персональном аккаунте
	apiMux.HandleFunc("GET /account", uh.HandleGetAccountInfoHandler)

	// Обновляет информацию личного кабинета
	apiMux.HandleFunc("POST /account/update", uh.UpdateAccountHandler)

	// Меняет пароль пользователя
	apiMux.HandleFunc("POST /account/password", ah.PasswordHandler)

	// Удаляет аккаунт пользователя
	apiMux.HandleFunc("POST /account/delete", uh.DeleteAccountHandler)

	// Получает профиль пользователя
	apiMux.HandleFunc("GET /account/profile", uh.ProfileGetHandler)

	// Обновляет профиль пользователя
	apiMux.HandleFunc("POST /account/profile/update", uh.ProfilePostHandler)

	// Загружает изображение профиля пользователя
	apiMux.HandleFunc("POST /account/avatar", fs.UploadAvatarHandler)

	// Возврат изображения профиля пользователя
	apiMux.HandleFunc("GET /account/avatar/view", fs.GetAvatarHandler)

	// Удаляет изображение профиля пользователя
	apiMux.HandleFunc("POST /account/avatar/delete", fs.DeleteAvatarHandler)

	// Экспорт данных пользователя
	apiMux.HandleFunc("GET /account/export", uh.ExportDateHandler)

	// Запрашивает экспорт данных пользователя
	apiMux.HandleFunc("POST /account/export/request", uh.RequestExportHandler)

	// Скачивает экспортированные данные пользователя
	apiMux.HandleFunc("POST /account/export/download", uh.HandleDownloadExportDate)

	// Отмечает сообщения как прочтённые
	apiMux.HandleFunc("POST /mark_message_as_read", ch.MarkMessagesAsRead)

	// Получает входящую почту пользователя
	apiMux.HandleFunc("GET /messages", ch.GetInboxMessages)

	// Получает архивированные сообщения пользователя
	apiMux.HandleFunc("GET /messages/archive", ch.GetArchivedMessages)

	// Устанавливает подкатегории пользователя
	apiMux.HandleFunc("POST /user/subcategories", uh.SetUserSubcategoriesHandler)

	// Получает выбранные подкатегории пользователя
	apiMux.HandleFunc("GET /user/subcategories", uh.GetUserSubcategoriesHandler)

	// Очищает выбор подкатегорий пользователя
	apiMux.HandleFunc("DELETE /user/subcategories", uh.ClearUserSubcategoriesHandler)

	// Получает точные типы пользователей администратором
	apiMux.HandleFunc("GET /admin/users", uh.AdminListUserTypesHandler)

	// Получает доступные типы пользователей администратором
	apiMux.HandleFunc("GET /admin/user_types", uh.AdminFetchUserTypesHandler)

	// Удаляет определённый тип пользователя администратором
	apiMux.HandleFunc("DELETE /admin/user_types/{user_id}", uh.AdminRemoveUserTypeHandler)

	// Получает персональные данные пользователя
	apiMux.HandleFunc("GET /account/personal-data", uh.GetUserPersonalDataHandler)

	// Обновляет личные данные пользователя
	apiMux.HandleFunc("PUT /account/personal-data", uh.UpdateUserPersonalDataHandler)

	// Получает номер телефона пользователя
	apiMux.HandleFunc("GET /account/phone-number", uh.GetUserPhoneNumberHandler)

	// Обновляет номер телефона пользователя
	apiMux.HandleFunc("PUT /account/phone-number", uh.UpdateUserPhoneNumberHandler)

	// Получает биографию пользователя
	apiMux.HandleFunc("GET /account/bio", uh.GetUserBiographyHandler)

	// Обновляет биографию пользователя
	apiMux.HandleFunc("PUT /account/bio", uh.UpdateUserBiographyHandler)

	// Получение списка заданий
	apiMux.HandleFunc("GET /tasks", th.GetTasks)

	// Поиск заданий по фильтру
	apiMux.HandleFunc("GET /tasks/search", th.SearchTasks)

	// Детали конкретного задания
	apiMux.HandleFunc("GET /tasks/{id}", th.GetTaskHandler)

	// Удаление задания
	apiMux.HandleFunc("DELETE /tasks/{id}", th.DeleteTaskByID)

	// Получение списка категорий заданий
	apiMux.HandleFunc("GET /tasks/categories", th.Categories)

	// Реакция на задание
	apiMux.HandleFunc("GET /task/response", th.ResponseHandler)

	// Получение откликов на задания
	apiMux.HandleFunc("GET /tasks/executor", th.GetTasksResponses)

	// Получение количества откликов специалиста
	apiMux.HandleFunc("GET /tasks/count/executor", th.GetTasksCountResponses)

	// Получение общего числа созданных заданий пользователем
	apiMux.HandleFunc("GET /tasks/count/user", th.GetTasksCount)

	// Получение заданий конкретного пользователя
	apiMux.HandleFunc("GET /tasks/user", th.GetTasksByUserID)

	// Проверка статуса контракта между заказчиком и исполнителем
	apiMux.HandleFunc("GET /contracts/check/{task_id}/{customer_id}/{executor_id}", th.CheckContract)

	// Получение откликов на конкретное задание
	apiMux.HandleFunc("GET /tasks/{task_id}/responses", th.GetResponsesHandler)

	// Создание отклика на задание
	apiMux.HandleFunc("POST /tasks/{id}/response", th.CreateResponse)

	// Создание нового задания
	apiMux.HandleFunc("POST /tasks/new", th.CreateTask)

	// Отмена задания
	apiMux.HandleFunc("POST /tasks/cancel", th.CancelTaskHandler)

	// Фиксирует просмотр задания
	apiMux.HandleFunc("POST /tasks/{task_id}/view", th.RecordTaskViewHandler)

	// Возвращает количество просмотров задания
	apiMux.HandleFunc("GET /tasks/{task_id}/views/count", th.GetTaskViewsCountHandler)

	// Создание контракта
	apiMux.HandleFunc("POST /contract", th.CreateContract)

	// Обновляет отчет по заданию
	apiMux.HandleFunc("PUT /update_report", th.UpdateReport)

	// Создание отчета по заданию
	apiMux.HandleFunc("POST /reports/create", th.CreateReport)

	// Проверяет наличие отчета по контракту
	apiMux.HandleFunc("GET /reports/check/{contract_id}", th.GetContractReportExists)

	// Получает новые отклики на задания
	apiMux.HandleFunc("GET /new_responses", th.ResponseViews)

	// Получает список категорий
	apiMux.HandleFunc("GET /categories", th.GetCategories)

	// Получает категорию по её ID
	apiMux.HandleFunc("GET /categories/{id}", th.CategoryId)

	// Получает список подкатегорий
	apiMux.HandleFunc("GET /subcategories", th.GetSubcategories)

	// Возвращает общее число откликов
	apiMux.HandleFunc("GET /total_responsess", th.TotalResponses)

	// Удаляет отклик на задание
	apiMux.HandleFunc("DELETE /response", th.DeleteResponseHandler)

	// Создание отзыва
	apiMux.HandleFunc("POST /reviews", th.CreateReview)

	// Получает статистику отзывов пользователя
	apiMux.HandleFunc("GET /users/{user_id}/reviews", th.GetUserReviewsStats)

	// Получает полный список отзывов пользователя
	apiMux.HandleFunc("GET /users/{user_id}/reviews/list", th.GetReviewsByUser)

	// Передача запросов в API-контроллеры
	mux.Handle("/api/", http.StripPrefix("/api", apiMux)) // Используем apiMux

	// Публичная страница сайта
	mux.HandleFunc("GET /public/", public.Public)

	// Частная страница доступна только авторизованным пользователям
	mux.HandleFunc("GET /private/", private.Private)

	// Главная страница приложения
	mux.HandleFunc("/", index.ServeIndexPage)

	// Аутентификационный middleware для всех маршрутов
	// Обратите внимание: http.Handle("/", ...) устанавливает обработчик по умолчанию для корневого пути.
	// Это может перекрыть другие маршруты, если не будет использован более специфичный порядок.
	// Лучше использовать session.AuthMiddleware как обертку для всего маршрутизатора.
	// Например: http.Handle("/", session.AuthMiddleware(sessionsManager, ctx, mux))
	// Но для вашего случая, где AuthMiddleware применяется ко всему серверу, это нормально.
	finalHandler := session.AuthMiddleware(sessionsManager, ctx, mux)

	// Обслуживание статичных изображений
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("../../uploads"))))

	// Обслуживание статических ресурсов проекта
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../assets"))))

	// Документация Swagger UI
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Сбор метрик Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	return finalHandler
}
