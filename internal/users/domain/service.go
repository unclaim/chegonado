package domain

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/unclaim/chegonado.git/internal/shared/config"
	"github.com/unclaim/chegonado.git/internal/shared/ports"
	"github.com/unclaim/chegonado.git/pkg/security/session"
	"github.com/unclaim/chegonado.git/pkg/security/token"
)

// UsersServiceImp implements the UsersService interface.
type UsersServiceImp struct {
	UsersRepo   UserRepositoryPort
	EmailSender ports.EmailSender
	Tokens      token.TokenManager
	Config      *config.AppConfig
}

// NewUsersService creates a new instance of UsersServiceImp.
func NewUsersService(repo UserRepositoryPort, emailSender ports.EmailSender, tokens token.TokenManager, config config.AppConfig) *UsersServiceImp {
	return &UsersServiceImp{
		UsersRepo:   repo,
		EmailSender: emailSender,
		Tokens:      tokens,
		Config:      &config,
	}
}

// UpdateAccountService - сервис для обновления данных аккаунта.
func (s *UsersServiceImp) UpdateAccountService(ctx context.Context, r *http.Request, accountRequest AccountRequest) error {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return fmt.Errorf("ошибка авторизации: %v", err)
	}

	ok, err := s.Tokens.Check(sess, accountRequest.CSRFToken)
	if !ok || err != nil {
		return fmt.Errorf("ошибка проверки токена безопасности: %v", err)
	}

	if accountRequest.Username == nil || *accountRequest.Username == "" || accountRequest.Email == "" {
		return fmt.Errorf("имя пользователя и электронная почта обязательны")
	}

	email := strings.ToLower(accountRequest.Email)

	if err := s.UsersRepo.UpdateUser(ctx, sess.UserID, accountRequest.Username, email, accountRequest.NoAds); err != nil {
		return fmt.Errorf("ошибка обновления данных пользователя: %v", err)
	}

	return nil
}

// AdminRemoveUserTypeService - сервис для удаления пользователя.
func (s *UsersServiceImp) AdminRemoveUserTypeService(ctx context.Context, userID string) error {
	return s.UsersRepo.RemoveUserByID(ctx, userID)
}

// AdminFetchUserTypesService - сервис для получения списка пользователей типа 'USER'.
func (s *UsersServiceImp) AdminFetchUserTypesService(ctx context.Context) ([]User, error) {
	return s.UsersRepo.FetchUsersByType(ctx, "USER")
}

// HandleGetBotsService - сервис для получения списка ботов.
func (s *UsersServiceImp) HandleGetBotsService(ctx context.Context) ([]User, error) {
	return s.UsersRepo.GetBots(ctx)
}

// ReviewsBotsService - сервис для получения отзывов ботов.
func (s *UsersServiceImp) ReviewsBotsService(ctx context.Context) ([]ReviewsBots, error) {
	return s.UsersRepo.GetReviewsBots(ctx)
}

// GetUserCategoriesService - сервис для получения категорий пользователя.
func (s *UsersServiceImp) GetUserCategoriesService(ctx context.Context, userIdInt int) ([]CategoryResponse, error) {
	return s.UsersRepo.GetUserCategories(ctx, userIdInt)
}

// OrdersHandlerGetService - сервис для получения заказов.
func (s *UsersServiceImp) OrdersHandlerGetService(ctx context.Context) ([]Order, error) {
	// Временные данные, пока не будет репозитория для заказов
	orders := []Order{
		{ID: 1, Item: "Товар 1", Amount: 2},
		{ID: 2, Item: "Товар 2", Amount: 5},
	}
	return orders, nil
}

// OrdersHandlerPostService - сервис для создания заказа.
func (s *UsersServiceImp) OrdersHandlerPostService(ctx context.Context, newOrder Order) (Order, error) {
	// Здесь должна быть логика сохранения заказа в репозиторий
	// Для примера просто возвращаем переданный объект
	return newOrder, nil
}

// HandleCompanyInfoService - сервис для получения информации о компании.
func (s *UsersServiceImp) HandleCompanyInfoService(ctx context.Context) (Company, error) {
	return s.UsersRepo.GetCompanyInfo(ctx)
}

// SaveCompanyInfoService - сервис для сохранения информации о компании.
func (s *UsersServiceImp) SaveCompanyInfoService(ctx context.Context, name, logoURL, websiteURL string) error {
	return s.UsersRepo.SaveCompanyInfo(ctx, name, logoURL, websiteURL)
}

// AdminListUserTypesService - сервис для получения количества пользователей по типам.
func (s *UsersServiceImp) AdminListUserTypesService(ctx context.Context) (UserResponse, error) {
	userCounts, totalCount, err := s.UsersRepo.ListUserCounts(ctx)
	if err != nil {
		return UserResponse{}, err
	}

	response := UserResponse{
		UserCounts: userCounts,
		TotalUserCount: TotalUserCount{
			Total: totalCount,
		},
	}
	return response, nil
}

// GetUserByQueryService - сервис для получения пользователя по ID.
func (s *UsersServiceImp) GetUserByQueryService(ctx context.Context, userID int64) (User, error) {
	user, err := s.UsersRepo.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("ошибка при поиске пользователя: %v", err)
	}

	if user == nil {
		return User{}, fmt.Errorf("пользователь с указанным id не найден: %v", err)
	}

	return *user, nil
}

// GetAccountInfoService - сервис для получения основной информации об аккаунте.
// Теперь он возвращает только данные пользователя, без CSRF-токена.
func (s *UsersServiceImp) GetAccountInfoService(ctx context.Context, r *http.Request) (User, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return User{}, fmt.Errorf("не удалось получить сессионные данные: %v", err)
	}

	user, err := s.UsersRepo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return User{}, fmt.Errorf("не удалось получить данные пользователя: %v", err)
	}

	return *user, nil
}

// GetUserPersonalDataService - сервис для получения персональных данных пользователя.
func (s *UsersServiceImp) GetUserPersonalDataService(ctx context.Context, r *http.Request) (Response, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	personalData, err := s.UsersRepo.GetUserPersonalData(ctx, sess.UserID)
	if err != nil {
		return Response{}, fmt.Errorf("ошибка загрузки сообщений: %v", err)
	}

	response := Response{
		StatusCode: http.StatusOK,
		Body:       personalData,
	}
	return response, nil
}

// HandleGetService - сервис для получения данных текущего пользователя.
func (s *UsersServiceImp) HandleGetService(ctx context.Context, r *http.Request) (User, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return User{}, fmt.Errorf("ошибка при получении сессии: %w", err)
	}

	user, err := s.UsersRepo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return User{}, fmt.Errorf("ошибка загрузки данных пользователя: %w", err)
	}

	return *user, nil
}

// CheckUserService - сервис для проверки существования пользователя.
func (s *UsersServiceImp) CheckUserService(ctx context.Context, req CheckUserRequest) (bool, error) {
	if req.Username == nil || len(*req.Username) == 0 {
		return false, fmt.Errorf("поле Username обязательно для заполнения")
	}

	user, err := s.UsersRepo.GetByLoginOrEmail(ctx, *req.Username, *req.Username)
	if err != nil {
		return false, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	return user != nil, nil
}

// HandleGettingOrderExecutorsService - сервис для получения списка исполнителей заказа.
func (s *UsersServiceImp) HandleGettingOrderExecutorsService(ctx context.Context, limitStr, offsetStr, proStr, onlineStr, categories, location string) ([]User, int, error) {
	limit, offset := 3, 0
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, 0, fmt.Errorf("некорректный параметр limit: %w", err)
		}
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return nil, 0, fmt.Errorf("некорректный параметр offset: %w", err)
		}
	}

	users, count, err := s.UsersRepo.FetchUsers(ctx, limit, offset, proStr, onlineStr, categories, location)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка при извлечении пользователей: %w", err)
	}
	return users, count, nil
}

// HandleAccountUpdateEmailService - сервис для обновления email.
func (s *UsersServiceImp) HandleAccountUpdateEmailService(ctx context.Context, r *http.Request, userEmail string) error {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить данные сессии: %w", err)
	}

	// if err := .ValidateEmail(userEmail); err != nil {
	// 	return fmt.Errorf("адрес электронной почты некорректен: %v", err)
	// }

	if err := s.UsersRepo.UpdateEmail(ctx, sess.UserID, userEmail); err != nil {
		return fmt.Errorf("не удалось обновить адрес электронной почты: %v", err)
	}
	return nil
}

// HandleBlockUserService - сервис для блокировки пользователя.
func (s *UsersServiceImp) HandleBlockUserService(ctx context.Context, r *http.Request, blockedID int64) (string, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("не удалось получить сессию: %v", err)
	}
	blockerID := sess.UserID

	exists, err := s.UsersRepo.UserExists(ctx, blockedID)
	if err != nil {
		return "", fmt.Errorf("не удалось проверить существование пользователя: %v", err)
	}
	if !exists {
		return "", fmt.Errorf("пользователь с указанным идентификатором не найден")
	}

	err = s.UsersRepo.BlockUser(ctx, blockerID, blockedID)
	if err != nil {
		return "", fmt.Errorf("не удалось заблокировать пользователя: %v", err)
	}
	return "Пользователь успешно заблокирован", nil
}

// UpdateUserSkills - обновляет навыки пользователя.
func (s *UsersServiceImp) UpdateUserSkills(ctx context.Context, userID int64, skills []int) error {
	if len(skills) == 0 {
		return fmt.Errorf("список навыков пуст")
	}

	for _, skill := range skills {
		if skill < 0 {
			return fmt.Errorf("некорректный индекс навыка: %d", skill)
		}
	}

	err := s.UsersRepo.UpdateUserSkills(ctx, userID, skills)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении навыков пользователя с ID %d: %w", userID, err)
	}

	return nil
}

// HandlePostService - сервис для обновления социальных ссылок.
func (s *UsersServiceImp) HandlePostService(ctx context.Context, r *http.Request) error {
	// sess, err := session.SessionFromContext(ctx)
	// if err != nil {
	// 	return fmt.Errorf("ошибка при получении сессии: %w", err)
	// }

	// userID := sess.UserID
	// r.ParseForm() // Парсим данные формы
	// vk := r.FormValue("vk")
	// telegram := r.FormValue("telegram")
	// whatsapp := r.FormValue("whatsapp")
	// web := r.FormValue("web")
	// twitter := r.FormValue("twitter")

	// if err := s.UsersRepo.UpdateUserLinks(ctx, userID, vk, telegram, whatsapp, web, twitter); err != nil {
	// 	return fmt.Errorf("ошибка сохранения данных: %v", err)
	// }
	return nil
}

// SubscribeHandlerService - сервис для подписки на пользователя.
func (s *UsersServiceImp) SubscribeHandlerService(ctx context.Context, r *http.Request, followedId int64) (Request, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}
	currentUserId := sess.UserID

	exists, err := s.UsersRepo.UserExists(ctx, followedId)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка проверки наличия пользователя: %v", err)
	}
	if !exists {
		return Request{}, fmt.Errorf("указанного пользователя не существует")
	}

	isFollowing, err := s.UsersRepo.IsFollowing(ctx, currentUserId, followedId)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка проверки состояния подписки: %v", err)
	}
	if isFollowing {
		return Request{}, fmt.Errorf("уже подписаны на этого пользователя")
	}

	err = s.UsersRepo.SubscribeUser(ctx, currentUserId, followedId)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка при создании подписки: %v", err)
	}

	response := Request{
		Message: "успешно подписались",
		Number:  int(followedId),
	}
	return response, nil
}

// UnblockHandlerService - сервис для разблокировки пользователя.
func (s *UsersServiceImp) UnblockHandlerService(ctx context.Context, r *http.Request, blockedID int64) (string, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении сессии: %v", err)
	}
	blockerID := sess.UserID

	isBlocked, err := s.UsersRepo.IsBlocked(ctx, blockerID, blockedID)
	if err != nil {
		return "", fmt.Errorf("ошибка проверки блокировки: %v", err)
	}
	if !isBlocked {
		return "", fmt.Errorf("этого пользователя нельзя разблокировать, поскольку он не был заблокирован")
	}

	err = s.UsersRepo.UnblockUser(ctx, blockerID, blockedID)
	if err != nil {
		return "", fmt.Errorf("ошибка снятия блокировки: %v", err)
	}
	return "Пользователь успешно разблокирован", nil
}

// UnsubscribeHandlerService - сервис для отписки от пользователя.
func (s *UsersServiceImp) UnsubscribeHandlerService(ctx context.Context, r *http.Request, followedId int64) (Request, error) {
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}
	currentUserId := sess.UserID

	err = s.UsersRepo.UnsubscribeUser(ctx, currentUserId, followedId)
	if err != nil {
		return Request{}, fmt.Errorf("ошибка отмены подписки: %v", err)
	}

	response := Request{
		Message: "успешно отписались",
		Number:  int(followedId),
	}
	return response, nil
}

// GetUserProfileDataService - функция, которая собирает все данные для профиля.
func (s *UsersServiceImp) GetUserProfileDataService(ctx context.Context, r *http.Request, username string) (Response, error) {
	// Шаг 1: Получаем основной профиль по имени пользователя
	profile, err := s.UsersRepo.GetUserByUsername(ctx, username)
	if err != nil {
		// Ошибка будет обработана вызывающей функцией (ProfileUsername)
		return Response{}, err
	}

	// Шаг 2: Получаем ID текущего пользователя из сессии
	currentUserID := int64(-1)
	sessionCookie, _ := r.Cookie("session_id")
	if sessionCookie != nil {
		currentUserID, err = s.UsersRepo.GetSessionUserID(ctx, sessionCookie.Value)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				// Обрабатываем другие ошибки, связанные с сессией
				return Response{}, fmt.Errorf("ошибка обработки сессии: %w", err)
			}
			// Если сессия не найдена (pgx.ErrNoRows), currentUserID останется -1,
			// и это будет обработано ниже. Мы не возвращаем ошибку, т.к. пользователь просто не авторизован.
		}
	}

	// Шаг 3: Получаем новые сообщения (если пользователь авторизован)
	var messages []Message
	if currentUserID != -1 {
		messages, err = s.UsersRepo.GetNewMessages(ctx, currentUserID)
		if err != nil {
			return Response{}, fmt.Errorf("не удалось получить новые сообщения: %v", err)
		}
	}

	// Шаг 4: Проверяем статус блокировки (если пользователь авторизован)
	isBlocked := false
	if currentUserID != -1 {
		isBlocked, err = s.UsersRepo.IsBlocked(ctx, currentUserID, profile.ID)
		if err != nil {
			return Response{}, fmt.Errorf("не удалось проверить статус блокировки: %v", err)
		}
	}

	// Шаг 5: Проверяем статус подписки (если пользователь авторизован)
	isFollowing := false
	if currentUserID != -1 {
		isFollowing, err = s.UsersRepo.IsFollowing(ctx, currentUserID, profile.ID)
		if err != nil {
			return Response{}, fmt.Errorf("не удалось проверить статус подписки: %v", err)
		}
	}

	// Шаг 6: Получаем количество подписок и подписчиков
	subscriptionsCount, err := s.UsersRepo.GetSubscriptionsCount(ctx, profile.ID)
	if err != nil {
		return Response{}, fmt.Errorf("не удалось получить количество подписок: %v", err)
	}

	followersCount, err := s.UsersRepo.GetFollowersCount(ctx, profile.ID)
	if err != nil {
		return Response{}, fmt.Errorf("не удалось получить количество подписчиков: %v", err)
	}

	// Шаг 7: Собираем итоговый ответ
	profile.IsFollowing = isFollowing
	profile.FollowersCount = followersCount

	response := Response{
		StatusCode: http.StatusOK,
		Body: struct {
			Profile            User      `json:"profile"`
			Messages           []Message `json:"messages"`
			SubscriptionsCount int64     `json:"subscriptions_count"`
			IsBlocked          bool      `json:"is_blocked"`
			IsFollowing        bool      `json:"is_following"`
		}{
			Profile:            *profile,
			Messages:           messages,
			SubscriptionsCount: subscriptionsCount,
			IsBlocked:          isBlocked,
			IsFollowing:        isFollowing,
		},
	}
	return response, nil
}

// GetUserProfileService - функция, которая собирает все данные для профиля.
func (s *UsersServiceImp) GetUserProfileService(ctx context.Context, r *http.Request, profileID int64) (ProfileResponse, error) {
	// Шаг 1: Получаем ID текущего пользователя из сессии.
	currentUserID := int64(-1)
	sess, err := session.SessionFromContext(r.Context())
	if err == nil {
		currentUserID = sess.UserID
	}

	// Шаг 2: Получаем основной профиль и количество подписок.
	profile, subscriptionsCount, err := s.UsersRepo.GetUserProfile(ctx, profileID, currentUserID)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("ошибка получения профиля пользователя: %v", err)
	}

	// Шаг 3: Проверяем статус блокировки.
	isBlocked := false
	if currentUserID != -1 {
		isBlocked, err = s.UsersRepo.IsBlocked(ctx, currentUserID, profileID)
		if err != nil {
			return ProfileResponse{}, fmt.Errorf("ошибка проверки блокировки: %v", err)
		}
	}

	// Шаг 4: Проверяем статус подписки.
	isFollowing := false
	if currentUserID != -1 {
		isFollowing, err = s.UsersRepo.IsFollowing(ctx, currentUserID, profileID)
		if err != nil {
			return ProfileResponse{}, fmt.Errorf("ошибка проверки подписки: %v", err)
		}
	}

	// Шаг 5: Собираем итоговый ответ.
	response := ProfileResponse{
		Profile:            profile,
		SubscriptionsCount: subscriptionsCount,
		IsBlocked:          isBlocked,
		IsFollowing:        isFollowing,
	}

	return response, nil
}

// SetUserSubcategoriesService - сервис для установки подкатегорий пользователя.
func (s *UsersServiceImp) SetUserSubcategoriesService(ctx context.Context, r *http.Request, userServices []UserUslugy) ([]int64, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return nil, fmt.Errorf("отсутствует авторизация: %v", err)
	}

	now := time.Now()
	for i := range userServices {
		userServices[i].UserID = sess.UserID
		userServices[i].CreatedAt = now
		userServices[i].UpdatedAt = now
	}

	if len(userServices) == 0 {
		err := s.UsersRepo.DeleteUserServices(ctx, sess.UserID)
		if err != nil {
			return nil, fmt.Errorf("ошибка удаления услуг: %v", err)
		}
		return nil, nil
	}

	newIDs, err := s.UsersRepo.InsertUserServices(ctx, userServices)
	if err != nil {
		return nil, fmt.Errorf("ошибка вставки услуг: %v", err)
	}

	return newIDs, nil
}

// GetUserSubcategoriesService - сервис для получения подкатегорий пользователя.
func (s *UsersServiceImp) GetUserSubcategoriesService(ctx context.Context, r *http.Request) (UserSpecialtyResponse, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return UserSpecialtyResponse{}, fmt.Errorf("отсутствует авторизация: %v", err)
	}

	userSpecialtyResponse, err := s.UsersRepo.GetAllCategoriesAndSubcategories(ctx, sess.UserID)
	if err != nil {
		return UserSpecialtyResponse{}, fmt.Errorf("ошибка получения специальностей: %v", err)
	}

	return userSpecialtyResponse, nil
}

// UserSkillsGetService - сервис для получения навыков пользователя.
func (s *UsersServiceImp) UserSkillsGetService(ctx context.Context, r *http.Request) (UserSkillsResponse, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return UserSkillsResponse{}, fmt.Errorf("отсутствует авторизация: %v", err)
	}

	skills, err := s.UsersRepo.GetUserSkills(ctx, sess.UserID)
	if err != nil {
		return UserSkillsResponse{}, fmt.Errorf("ошибка получения навыков: %v", err)
	}

	if skills == nil {
		skills = []Skill{}
	}

	response := UserSkillsResponse{
		UserID: sess.UserID,
		Skills: skills,
	}

	return response, nil
}

// UserSkillsPostService - сервис для обновления навыков пользователя.
func (s *UsersServiceImp) UserSkillsPostService(ctx context.Context, r *http.Request, skills []int) error {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return fmt.Errorf("отсутствует авторизация: %v", err)
	}

	if err := s.UsersRepo.UpdateUserSkills(ctx, sess.UserID, skills); err != nil {
		return fmt.Errorf("ошибка обновления навыков: %v", err)
	}

	return nil
}

// ClearUserSubcategoriesService - сервис для удаления подкатегорий пользователя.
func (s *UsersServiceImp) ClearUserSubcategoriesService(ctx context.Context, r *http.Request, request DeleteSubcategoryRequest) (string, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return "", fmt.Errorf("отсутствует авторизация: %v", err)
	}

	err = s.UsersRepo.RemoveSubcategoryFromUserService(ctx, sess.UserID, request.SubcategoryID)
	if err != nil {
		return "", fmt.Errorf("ошибка удаления подкатегории: %v", err)
	}

	return "Подкатегория успешно удалена.", nil
}

// RequestExportService - сервис для запроса экспорта данных.
func (s *UsersServiceImp) RequestExportService(ctx context.Context, r *http.Request) error {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	user, err := s.UsersRepo.GetUserData(ctx, sess.UserID)
	if err != nil {
		return fmt.Errorf("не удалось получить данные пользователя: %v", err)
	}

	userEmail, err := s.UsersRepo.GetEmailByUserID(ctx, sess.UserID)
	if err != nil {
		return fmt.Errorf("не удалось получить email пользователя: %v", err)
	}

	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			log.Printf("Error closing zip writer: %v", err)
		}
	}()
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("ошибка преобразования данных в JSON: %v", err)
	}

	jsonFile, err := zipWriter.Create("user_data.json")
	if err != nil {
		return fmt.Errorf("ошибка создания JSON-файла в ZIP: %v", err)
	}
	if _, err := jsonFile.Write(jsonData); err != nil {
		return fmt.Errorf("ошибка записи данных в JSON-файл: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия ZIP-писателя: %v", err)
	}

	attachment := ports.Attachment{
		Filename:    fmt.Sprintf("export_%d.zip", sess.UserID),
		ContentType: "application/zip",
		Data:        zipBuffer.Bytes(),
	}

	templatePath := filepath.Join("../../web/templates/emails", "data_export.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("не удалось загрузить шаблон письма: %w", err)
	}

	var body bytes.Buffer
	data := struct {
		Username string
	}{
		Username: *user.Username,
	}
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("ошибка при подготовке тела письма: %w", err)
	}

	subject := "Экспорт ваших данных"
	err = s.EmailSender.SendEmailWithAttachments(userEmail, subject, &body, []ports.Attachment{attachment})
	if err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}

	return nil
}

// UpdateUserPersonalDataService - сервис для обновления персональных данных пользователя.
func (s *UsersServiceImp) UpdateUserPersonalDataService(ctx context.Context, r *http.Request, request User) (Response, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	if err := s.UsersRepo.UpdateUserPersonalData(ctx, sess.UserID, request); err != nil {
		return Response{}, err
	}

	msg := SuccessResponse{Message: "Данные успешно обновлены"}
	response := Response{StatusCode: http.StatusOK, Body: msg}
	return response, nil
}

// GetUserPhoneNumberService - сервис для получения номера телефона пользователя.
func (s *UsersServiceImp) GetUserPhoneNumberService(ctx context.Context, r *http.Request) (Response, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	phoneNumber, err := s.UsersRepo.GetUserPhoneNumber(ctx, sess.UserID)
	if err != nil {
		return Response{}, err
	}

	phone := Phone{Phone: phoneNumber}
	response := Response{StatusCode: http.StatusOK, Body: phone}
	return response, nil
}

// UpdateUserPhoneNumberService - сервис для обновления номера телефона пользователя.
func (s *UsersServiceImp) UpdateUserPhoneNumberService(ctx context.Context, r *http.Request, request PhoneNumberUpdateRequest) (Response, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	if err := s.UsersRepo.UpdateUserPhoneNumber(ctx, sess.UserID, request.Phone); err != nil {
		return Response{}, err
	}

	msg := SuccessResponse{Message: "Номер телефона успешно изменён"}
	response := Response{StatusCode: http.StatusOK, Body: msg}
	return response, nil
}

// GetUserBiographyService - сервис для получения биографии пользователя.
func (s *UsersServiceImp) GetUserBiographyService(ctx context.Context, r *http.Request) (Response, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	bio, err := s.UsersRepo.GetUserBio(ctx, sess.UserID)
	if err != nil {
		return Response{}, err
	}

	response := Response{StatusCode: http.StatusOK, Body: Bio{Bio: bio}}
	return response, nil
}

// UpdateUserBiographyService - сервис для обновления биографии пользователя.
func (s *UsersServiceImp) UpdateUserBiographyService(ctx context.Context, r *http.Request, request Bio) (Response, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return Response{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	if err := s.UsersRepo.UpdateUserBio(ctx, sess.UserID, request.Bio); err != nil {
		return Response{}, err
	}

	msg := SuccessResponse{Message: "Биография успешно обновлена"}
	response := Response{StatusCode: http.StatusOK, Body: msg}
	return response, nil
}

// DeleteAccountService - сервис для удаления аккаунта.
func (s *UsersServiceImp) DeleteAccountService(ctx context.Context, r *http.Request) (map[string]interface{}, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сессию: %v", err)
	}

	if err = s.UsersRepo.DeleteUserByID(ctx, sess.UserID); err != nil {
		return nil, fmt.Errorf("не удалось удалить пользователя из базы данных: %v", err)
	}

	// if err = utils.DeleteUserFiles(sess.UserID); err != nil {
	// 	return nil, fmt.Errorf("не удалось удалить файлы пользователя: %v", err)
	// }

	response := map[string]interface{}{
		"status":  "success",
		"message": "Ваш аккаунт был успешно удален.",
	}
	return response, nil
}

// DeleteConfirmationService - сервис для подтверждения удаления аккаунта.
func (s *UsersServiceImp) DeleteConfirmationService(status string) map[string]interface{} {
	var message string
	if status == "success" {
		message = "Ваш аккаунт был успешно удален."
	} else {
		message = "Произошла ошибка при удалении вашего аккаунта."
	}

	return map[string]interface{}{"status": "info", "message": message}
}

// DownloadExportDateService - сервис для скачивания экспортированных данных.
// Функция теперь возвращает буфер с данными и имя файла.
func (s *UsersServiceImp) DownloadExportDateService(ctx context.Context, r *http.Request) (*bytes.Buffer, string, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить сессию: %v", err)
	}

	user, err := s.UsersRepo.GetUserData(ctx, sess.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить данные пользователя: %v", err)
	}

	// Обновляем дату экспорта в базе данных.
	if err := s.UsersRepo.UpdateExportDate(ctx, sess.UserID); err != nil {
		return nil, "", fmt.Errorf("не удалось обновить дату экспорта: %v", err)
	}

	// Создаем буфер в памяти.
	var buffer bytes.Buffer

	// Формируем новый комментарий в буфере, как ты и просил.
	comment := fmt.Sprintf("// Привет!\n// Ты %s в %s скачал информацию о себе.\n\n",
		time.Now().Format("02 января 2006"), time.Now().Format("15:04"))
	buffer.WriteString(comment)

	// Создаем структуру, которая будет содержать только нужные поля для JSON.
	dataToExport := struct {
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
	}{
		Username:  *user.Username, // <--- И здесь нужно разыменовать указатель
		CreatedAt: user.CreatedAt,
	}

	// Кодируем данные в JSON и записываем в буфер.
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ") // Делаем JSON красивым (с отступами)
	if err := encoder.Encode(dataToExport); err != nil {
		return nil, "", fmt.Errorf("не удалось закодировать данные пользователя в JSON: %v", err)
	}

	// Получаем название компании из конфигурации.
	companyName := s.Config.App.Name
	fmt.Println(companyName)
	// Теперь мы корректно формируем строку для записи в буфер.
	// Здесь мы используем fmt.Sprintf для подстановки.
	buffer.WriteString(fmt.Sprintf("\n// С заботой,\n// Команда %s.", companyName)) // <--- Исправлено

	// Формируем имя файла на основе имени пользователя и текущей даты.
	filename := fmt.Sprintf("export_%s_%s.json", *user.Username, time.Now().Format("2006-01-02")) // <--- Исправлено

	return &buffer, filename, nil
}

// ExportDateService - сервис для получения даты последнего экспорта.
func (s *UsersServiceImp) ExportDateService(ctx context.Context, r *http.Request) (map[string]interface{}, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сессию: %v", err)
	}

	email, err := s.UsersRepo.GetEmailByUserID(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить email пользователя: %v", err)
	}

	exportDate, err := s.UsersRepo.GetExportDate(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить дату последнего экспорта: %v", err)
	}

	response := map[string]interface{}{
		"Email":      email,
		"ExportDate": exportDate,
	}
	return response, nil
}

// ProfileGetService - сервис для получения данных профиля.
func (s *UsersServiceImp) ProfileGetService(ctx context.Context, r *http.Request) (User, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return User{}, fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	u, err := s.UsersRepo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return User{}, fmt.Errorf("ошибка получения данных пользователя: %v", err)
	}

	response := User{
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Location:   u.Location,
		Bio:        u.Bio,
		NoAds:      u.NoAds,
	}
	return response, nil
}

// ProfilePostService - сервис для обновления данных профиля.
func (s *UsersServiceImp) ProfilePostService(ctx context.Context, r *http.Request, update User) error {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	if update.FirstName == nil || update.LastName == nil {
		return fmt.Errorf("обязательно заполнить имя и фамилию")
	}

	if err := s.UsersRepo.UpdateUserPersonalData(ctx, sess.UserID, update); err != nil {
		return fmt.Errorf("ошибка обновления профиля: %v", err)
	}

	return nil
}
