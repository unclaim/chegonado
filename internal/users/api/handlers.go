package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/unclaim/chegonado.git/internal/shared/common_errors"
	"github.com/unclaim/chegonado.git/internal/shared/utils"

	"github.com/unclaim/chegonado.git/internal/users/domain"
	"github.com/unclaim/chegonado.git/pkg/security/session"
	"github.com/unclaim/chegonado.git/pkg/security/token"
)

type UserHandler struct {
	Tokens  token.TokenManager
	Service domain.UsersService
}

func NewUserHandler(tokens token.TokenManager, service domain.UsersService) *UserHandler {
	return &UserHandler{
		Tokens:  tokens,
		Service: service,
	}
}

// AdminRemoveUserTypeHandler - обработчик для удаления пользователя.
func (uh *UserHandler) AdminRemoveUserTypeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("не указан идентификатор пользователя"), http.StatusBadRequest)
		return
	}

	err := uh.Service.AdminRemoveUserTypeService(r.Context(), userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			common_errors.NewAppError(w, r, fmt.Errorf("пользователь с указанным идентификатором не найден"), http.StatusNotFound)
		} else {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка при удалении пользователя: %w", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminFetchUserTypesHandler - обработчик для получения списка пользователей типа 'USER'.
func (uh *UserHandler) AdminFetchUserTypesHandler(w http.ResponseWriter, r *http.Request) {
	users, err := uh.Service.AdminFetchUserTypesService(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить пользователей: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось закодировать ответ: %w", err), http.StatusInternalServerError)
		return
	}
}

// AdminListUserTypesHandler - обработчик для получения количества пользователей по типам.
func (uh *UserHandler) AdminListUserTypesHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.AdminListUserTypesService(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить данные: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось сформировать ответ: %w", err), http.StatusInternalServerError)
		return
	}
}

// handleGet - обработчик для получения данных текущего пользователя.
func (uh *UserHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	user, err := uh.Service.HandleGetService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки данных пользователя: %w", err), http.StatusInternalServerError)
		return
	}
	utils.NewResponse(w, http.StatusOK, user)
}

// GetBotsHandler - обработчик для получения списка ботов.
func (uh *UserHandler) GetBotsHandler(w http.ResponseWriter, r *http.Request) {
	users, err := uh.Service.HandleGetBotsService(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить список ботов: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, users)
}

// ReviewsBotsHandler - обработчик для получения отзывов ботов.
func (uh *UserHandler) ReviewsBotsHandler(w http.ResponseWriter, r *http.Request) {
	reviews, err := uh.Service.ReviewsBotsService(r.Context())
	if err != nil {
		if len(reviews) == 0 {
			common_errors.NewAppError(w, r, fmt.Errorf("отзывы не найдены"), http.StatusNotFound)
		} else {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении отзывов: %w", err), http.StatusInternalServerError)
		}
		return
	}

	utils.NewResponse(w, http.StatusOK, reviews)
}

// CheckUserHandler - обработчик для проверки существования пользователя.
func (uh *UserHandler) CheckUserHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.CheckUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при разборе данных запроса: %v", err), http.StatusBadRequest)
		return
	}

	exists, err := uh.Service.CheckUserService(r.Context(), req)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при поиске пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		common_errors.NewAppError(w, r, fmt.Errorf("пользователь не найден"), http.StatusNotFound)
		return
	}

	response := map[string]bool{"exists": true}
	utils.NewResponse(w, http.StatusOK, response)
}

// GettingOrderExecutorsHandler - обработчик для получения списка исполнителей заказа.
func (uh *UserHandler) GettingOrderExecutorsHandler(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	proStr := r.URL.Query().Get("pro")
	onlineStr := r.URL.Query().Get("online")
	categories := r.URL.Query().Get("categories")
	location := r.URL.Query().Get("location")

	users, count, err := uh.Service.HandleGettingOrderExecutorsService(r.Context(), limitStr, offsetStr, proStr, onlineStr, categories, location)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при извлечении пользователей: %v", err), http.StatusInternalServerError)
		return
	}

	response := domain.OrderExecutorsResponse{
		StatusCode: http.StatusOK,
		Body:       users,
		TotalCount: count,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetUserCategoriesHandler - обработчик для получения категорий пользователя.
func (uh *UserHandler) GetUserCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("id")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil || userIdInt <= 0 {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный идентификатор пользователя: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.GetUserCategoriesService(r.Context(), userIdInt)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка базы данных: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// OrdersHandler - обработчик для работы с заказами.
func (uh *UserHandler) OrdersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orders, err := uh.Service.OrdersHandlerGetService(r.Context())
		if err != nil {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения заказов: %v", err), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка сериализации данных: %v", err), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		var newOrder domain.Order
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка декодирования данных: %v", err), http.StatusBadRequest)
			return
		}
		createdOrder, err := uh.Service.OrdersHandlerPostService(r.Context(), newOrder)
		if err != nil {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка создания заказа: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(createdOrder); err != nil {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка сериализации данных: %v", err), http.StatusInternalServerError)
			return
		}

	default:
		common_errors.NewAppError(w, r, fmt.Errorf("неподдерживаемый метод запроса"), http.StatusMethodNotAllowed)
		return
	}
}

// SaveCompanyInfoHandler - обработчик для сохранения информации о компании.
func (uh *UserHandler) SaveCompanyInfoHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	logoURL := r.FormValue("logo_url")
	websiteURL := r.FormValue("website_url")

	if err := uh.Service.SaveCompanyInfoService(r.Context(), name, logoURL, websiteURL); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при сохранении информации о компании: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/company", http.StatusFound)
}

// AccountUpdateEmailHandler - обработчик для обновления email.
func (uh *UserHandler) AccountUpdateEmailHandler(w http.ResponseWriter, r *http.Request) {

	var updateRequest struct {
		UserEmail string `json:"user_email"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil || updateRequest.UserEmail == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный или пустой email: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if err := uh.Service.HandleAccountUpdateEmailService(r.Context(), r, updateRequest.UserEmail); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось обновить адрес электронной почты: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/account/unverified", http.StatusFound)
}

// BlockUserHandler - обработчик для блокировки пользователя.
func (uh *UserHandler) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неправильный формат идентификатора пользователя: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.HandleBlockUserService(r.Context(), r, int64(userIdInt))
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось заблокировать пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// SocialProfileHandler - обработчик для социальных ссылок профиля.
func (uh *UserHandler) SocialProfileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// uh.handlePost(w, r)
	case http.MethodGet:
		uh.handleGet(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		w.Header().Set("Allow", "GET, POST")
		return
	}
}

// SubscribeHandler - обработчик для подписки на пользователя.
func (uh *UserHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	followedIdStr := r.PathValue("id")
	followedId, err := strconv.ParseInt(followedIdStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования значения ID пользователя: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.SubscribeHandlerService(r.Context(), r, followedId)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка подписки: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// UnblockHandler - обработчик для разблокировки пользователя.
func (uh *UserHandler) UnblockHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	if userId == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("не указан идентификатор пользователя"), http.StatusBadRequest)
		return
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования идентификатора пользователя: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.UnblockHandlerService(r.Context(), r, int64(userIdInt))
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка снятия блокировки: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// UnsubscribeHandler - обработчик для отписки от пользователя.
func (uh *UserHandler) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	followedIdStr := r.PathValue("id")
	followedId, err := strconv.ParseInt(followedIdStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования id пользователя: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.UnsubscribeHandlerService(r.Context(), r, followedId)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка отмены подписки: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// GetUserPersonalDataHandler - обработчик для получения персональных данных пользователя.
func (uh *UserHandler) GetUserPersonalDataHandler(w http.ResponseWriter, r *http.Request) {

	response, err := uh.Service.GetUserPersonalDataService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки сообщений: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateUserPersonalDataHandler - обработчик для обновления персональных данных.
func (uh *UserHandler) UpdateUserPersonalDataHandler(w http.ResponseWriter, r *http.Request) {
	var request domain.User
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный запрос: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	response, err := uh.Service.UpdateUserPersonalDataService(r.Context(), r, request)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обновления данных: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetUserPhoneNumberHandler - обработчик для получения номера телефона.
func (uh *UserHandler) GetUserPhoneNumberHandler(w http.ResponseWriter, r *http.Request) {

	response, err := uh.Service.GetUserPhoneNumberService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки номера телефона: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateUserPhoneNumberHandler - обработчик для обновления номера телефона.
func (uh *UserHandler) UpdateUserPhoneNumberHandler(w http.ResponseWriter, r *http.Request) {
	var request domain.PhoneNumberUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный запрос: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	response, err := uh.Service.UpdateUserPhoneNumberService(r.Context(), r, request)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сохранения номера телефона: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetUserBiographyHandler - обработчик для получения биографии.
func (uh *UserHandler) GetUserBiographyHandler(w http.ResponseWriter, r *http.Request) {

	response, err := uh.Service.GetUserBiographyService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка загрузки биографии: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateUserBiographyHandler - обработчик для обновления биографии.
func (uh *UserHandler) UpdateUserBiographyHandler(w http.ResponseWriter, r *http.Request) {
	var request domain.Bio
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный запрос: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	response, err := uh.Service.UpdateUserBiographyService(r.Context(), r, request)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сохранения биографии: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка формирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetAccountInfoHandler - обработчик для получения основной информации об аккаунте.
// Теперь он отвечает за создание и установку CSRF-токена.
func (uh *UserHandler) HandleGetAccountInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: Вызываем сервис для получения данных пользователя.
	response, err := uh.Service.GetAccountInfoService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить данные аккаунта: %v", err), http.StatusUnauthorized)
		return
	}

	// Шаг 2: Получаем сессию для создания CSRF-токена.
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии для CSRF-токена: %v", err), http.StatusUnauthorized)
		return
	}

	// Шаг 3: Создаём и устанавливаем CSRF-токен в заголовок и куки.
	expirationTime := time.Now().Add(24 * time.Hour)
	CSRFToken, err := uh.Tokens.Create(sess, expirationTime.Unix())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось создать CSRF-токен: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", CSRFToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    CSRFToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	// Шаг 4: Отправляем ответ.
	utils.NewResponse(w, http.StatusOK, response)
}

// DeleteAccountHandler - обработчик для удаления аккаунта.
func (uh *UserHandler) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.DeleteAccountService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось удалить пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// DeleteConfirmationHandler - обработчик для подтверждения удаления.
func (uh UserHandler) DeleteConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	response := uh.Service.DeleteConfirmationService(status)

	utils.NewResponse(w, http.StatusOK, response)
}

// HandleDownloadExportDate - обработчик для скачивания экспортированных данных.
func (uh *UserHandler) HandleDownloadExportDate(w http.ResponseWriter, r *http.Request) {
	// Подготовка данных для скачивания.
	fileContent, filename, err := uh.Service.DownloadExportDateService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при подготовке файла для скачивания: %v", err), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки для скачивания.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Отправляем буфер клиенту.
	if _, err := io.Copy(w, fileContent); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось передать файл клиенту: %v", err), http.StatusInternalServerError)
		return
	}
}

// ExportDateHandler - обработчик для получения даты последнего экспорта.
func (uh *UserHandler) ExportDateHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.ExportDateService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить данные для экспорта: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// GetUserByQueryHandler - обработчик для получения пользователя по ID.
func (uh *UserHandler) GetUserByQueryHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("отсутствует обязательный параметр id"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный формат id: %v", err), http.StatusBadRequest)
		return
	}

	user, err := uh.Service.GetUserByQueryService(r.Context(), userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при поиске пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, user)
}

// ProfileGetHandler - обработчик для получения данных профиля.
func (uh *UserHandler) ProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.ProfileGetService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения данных пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сериализации данных: %v", err), http.StatusInternalServerError)
		return
	}
}

// ProfilePostHandler - обработчик для обновления данных профиля.
func (uh *UserHandler) ProfilePostHandler(w http.ResponseWriter, r *http.Request) {
	var update domain.User
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка десериализации данных: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if err := uh.Service.ProfilePostService(r.Context(), r, update); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обновления профиля: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateAccountHandler - обработчик для обновления данных аккаунта.
func (uh *UserHandler) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest domain.AccountRequest
	decoder := json.NewDecoder(r.Body)
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if err := decoder.Decode(&accountRequest); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка декодирования данных: %v", err), http.StatusBadRequest)
		return
	}

	if err := uh.Service.UpdateAccountService(r.Context(), r, accountRequest); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обновления данных пользователя: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RequestExportHandler - обработчик для запроса экспорта данных.
func (uh *UserHandler) RequestExportHandler(w http.ResponseWriter, r *http.Request) {
	if err := uh.Service.RequestExportService(r.Context(), r); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обработки запроса на экспорт: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Запрос на экспорт данных обработан. Архив отправлен на ваш email.",
	}
	utils.NewResponse(w, http.StatusOK, response)
}

// UserSkillsGetHandler - обработчик для получения навыков пользователя.
func (uh *UserHandler) UserSkillsGetHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.UserSkillsGetService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения навыков: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// UserSkillsPostHandler - обработчик для обновления навыков пользователя.
func (uh *UserHandler) UserSkillsPostHandler(w http.ResponseWriter, r *http.Request) {
	var skills domain.UserSkills
	if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный формат передаваемых данных: %v", err), http.StatusBadRequest)
		return
	}
	// Check the error after closing
	defer func() {
		if err := r.Body.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if err := uh.Service.UserSkillsPostService(r.Context(), r, skills.Skills); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обновления навыков: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ClearUserSubcategoriesHandler - обработчик для удаления подкатегорий пользователя.
func (uh *UserHandler) ClearUserSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var request domain.DeleteSubcategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный формат входных данных: %v", err), http.StatusBadRequest)
		return
	}

	response, err := uh.Service.ClearUserSubcategoriesService(r.Context(), r, request)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка удаления подкатегории: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// GetUserSubcategoriesHandler - обработчик для получения подкатегорий пользователя.
func (uh *UserHandler) GetUserSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	response, err := uh.Service.GetUserSubcategoriesService(r.Context(), r)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения специальностей: %v", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, response)
}

// SetUserSubcategoriesHandler - обработчик для установки подкатегорий пользователя.
func (uh *UserHandler) SetUserSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var userServices []domain.UserUslugy
	if err := json.NewDecoder(r.Body).Decode(&userServices); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("невозможно декодировать тело запроса: %v", err), http.StatusBadRequest)
		return
	}

	newIDs, err := uh.Service.SetUserSubcategoriesService(r.Context(), r, userServices)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка вставки услуг: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(map[string][]int64{"ids": newIDs}); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

// updateProfileDataHandler - обновляет данные профиля пользователя.
// func (uh *UserHandler) updateProfileDataHandler(ctx context.Context, userID int64, update domain.User) error {
// 	err := uh.UsersRepo.UpdateProfile(ctx, userID, *update.FirstName, *update.LastName, *update.MiddleName, *update.Location, *update.Bio, update.NoAds)
// 	return err
// }

// GetUserProfileHandler - обработчик для получения данных профиля по ID.
func (uh *UserHandler) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: Получаем ID пользователя из URL.
	userIdStr := r.PathValue("id")
	if userIdStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("не указан ID пользователя"), http.StatusBadRequest)
		return
	}

	userIdInt, err := strconv.Atoi(userIdStr)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неправильный формат ID пользователя: %v", err), http.StatusBadRequest)
		return
	}

	// Шаг 2: Вызываем сервис для получения всех данных профиля.
	// Вся логика, включая работу с сессией, теперь находится внутри сервиса.
	response, err := uh.Service.GetUserProfileService(r.Context(), r, int64(userIdInt))
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения данных профиля: %v", err), http.StatusInternalServerError)
		return
	}

	// Шаг 3: Отправляем успешный ответ.
	utils.NewResponse(w, http.StatusOK, response)
}

// GetUserProfileDataHandler - обработчик для получения данных профиля по имени пользователя.
func (uh *UserHandler) GetUserProfileDataHandler(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: Получаем имя пользователя из URL
	username := r.PathValue("username")
	if username == "" {
		// Обрабатываем ошибку, если имя пользователя не указано
		common_errors.NewAppError(w, r, fmt.Errorf("не указано имя пользователя"), http.StatusBadRequest)
		return
	}

	// Шаг 2: Вызываем сервис для получения всех данных профиля
	// Передаём только имя пользователя и сам HTTP-запрос, чтобы сервис мог получить данные сессии
	response, err := uh.Service.GetUserProfileDataService(r.Context(), r, username)
	if err != nil {
		// Обрабатываем ошибки, возвращаемые сервисом
		if errors.Is(err, pgx.ErrNoRows) {
			common_errors.NewAppError(w, r, fmt.Errorf("пользователь не найден"), http.StatusNotFound)
		} else {
			common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения данных профиля: %w", err), http.StatusInternalServerError)
		}
		return
	}
	// Шаг 3: Отправляем успешный ответ
	utils.NewResponse(w, http.StatusOK, response)
}

// func (uh *UserHandler) handlePost(w http.ResponseWriter, r *http.Request) {
// 	sess, err := session.SessionFromContext(r.Context())
// 	if err != nil {
// 		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %v", err), http.StatusUnauthorized)
// 		return
// 	}

// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

// 	userID := sess.UserID

// 	vk := r.FormValue("vk")
// 	telegram := r.FormValue("telegram")
// 	whatsapp := r.FormValue("whatsapp")
// 	web := r.FormValue("web")
// 	twitter := r.FormValue("twitter")

// 	if err := uh.UsersRepo.UpdateUserLinks(r.Context(), userID, vk, telegram, whatsapp, web, twitter); err != nil {
// 		common_errors.NewAppError(w, r, fmt.Errorf("ошибка сохранения данных: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }
