package domain

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

// UsersService defines the interface for user-related services.
type UsersService interface {
	UpdateUserSkills(ctx context.Context, userID int64, skills []int) error
	UpdateAccountService(ctx context.Context, r *http.Request, accountRequest AccountRequest) error
	AdminRemoveUserTypeService(ctx context.Context, userID string) error
	AdminFetchUserTypesService(ctx context.Context) ([]User, error)
	HandleGetBotsService(ctx context.Context) ([]User, error)
	ReviewsBotsService(ctx context.Context) ([]ReviewsBots, error)
	GetUserCategoriesService(ctx context.Context, userIdInt int) ([]CategoryResponse, error)
	OrdersHandlerGetService(ctx context.Context) ([]Order, error)
	OrdersHandlerPostService(ctx context.Context, newOrder Order) (Order, error)
	HandleCompanyInfoService(ctx context.Context) (Company, error)
	SaveCompanyInfoService(ctx context.Context, name, logoURL, websiteURL string) error
	AdminListUserTypesService(ctx context.Context) (UserResponse, error)
	GetUserByQueryService(ctx context.Context, userID int64) (User, error)
	GetAccountInfoService(ctx context.Context, r *http.Request) (User, error)
	GetUserPersonalDataService(ctx context.Context, r *http.Request) (Response, error)
	HandleGetService(ctx context.Context, r *http.Request) (User, error)
	CheckUserService(ctx context.Context, req CheckUserRequest) (bool, error)
	HandleGettingOrderExecutorsService(ctx context.Context, limitStr, offsetStr, proStr, onlineStr, categories, location string) ([]User, int, error)
	HandleAccountUpdateEmailService(ctx context.Context, r *http.Request, userEmail string) error
	HandleBlockUserService(ctx context.Context, r *http.Request, blockedID int64) (string, error)
	HandlePostService(ctx context.Context, r *http.Request) error
	SubscribeHandlerService(ctx context.Context, r *http.Request, followedId int64) (Request, error)
	UnblockHandlerService(ctx context.Context, r *http.Request, blockedID int64) (string, error)
	UnsubscribeHandlerService(ctx context.Context, r *http.Request, followedId int64) (Request, error)
	GetUserProfileDataService(ctx context.Context, r *http.Request, username string) (Response, error)
	GetUserProfileService(ctx context.Context, r *http.Request, profileID int64) (ProfileResponse, error)
	SetUserSubcategoriesService(ctx context.Context, r *http.Request, userServices []UserUslugy) ([]int64, error)
	GetUserSubcategoriesService(ctx context.Context, r *http.Request) (UserSpecialtyResponse, error)
	UserSkillsGetService(ctx context.Context, r *http.Request) (UserSkillsResponse, error)
	UserSkillsPostService(ctx context.Context, r *http.Request, skills []int) error
	ClearUserSubcategoriesService(ctx context.Context, r *http.Request, request DeleteSubcategoryRequest) (string, error)
	RequestExportService(ctx context.Context, r *http.Request) error
	UpdateUserPersonalDataService(ctx context.Context, r *http.Request, request User) (Response, error)
	GetUserPhoneNumberService(ctx context.Context, r *http.Request) (Response, error)
	UpdateUserPhoneNumberService(ctx context.Context, r *http.Request, request PhoneNumberUpdateRequest) (Response, error)
	GetUserBiographyService(ctx context.Context, r *http.Request) (Response, error)
	UpdateUserBiographyService(ctx context.Context, r *http.Request, request Bio) (Response, error)
	DeleteAccountService(ctx context.Context, r *http.Request) (map[string]interface{}, error)
	DeleteConfirmationService(status string) map[string]interface{}
	DownloadExportDateService(ctx context.Context, r *http.Request) (*bytes.Buffer, string, error)
	ExportDateService(ctx context.Context, r *http.Request) (map[string]interface{}, error)
	ProfileGetService(ctx context.Context, r *http.Request) (User, error)
	ProfilePostService(ctx context.Context, r *http.Request, update User) error
}

// UserRepositoryPort определяет интерфейс для работы с данными пользователей.
// Это помогает абстрагировать логику работы с БД от бизнес-логики.
type UserRepositoryPort interface {
	GetUserPersonalData(ctx context.Context, userID int64) (*PersonalData, error)
	UpdateUserPersonalData(ctx context.Context, userID int64, data User) error
	GetUserPhoneNumber(ctx context.Context, userID int64) (*string, error)
	UpdateUserPhoneNumber(ctx context.Context, userID int64, newPhone string) error
	GetUserBio(ctx context.Context, userID int64) (*string, error)
	UpdateUserBio(ctx context.Context, userID int64, newBio *string) error
	GetNewMessages(ctx context.Context, currentUserId int64) ([]Message, error)
	GetByLoginOrEmail(ctx context.Context, username string, email string) (*User, error)
	BlockUser(ctx context.Context, blockerID, blockedID int64) error
	CreateAccountVerificationsCode(ctx context.Context, email string, code int64) error
	CreateHashPass(ctx context.Context, plainPassword, salt string) ([]byte, error)
	DeleteUserByID(ctx context.Context, userID int64) error
	FetchUsers(ctx context.Context, limit, offset int, proStr, onlineStr, categories, location string) ([]User, int, error)
	GetByEmail(ctx context.Context, Email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetCompanyInfo(ctx context.Context) (Company, error)
	GetEmailByUserID(ctx context.Context, userID int64) (string, error)
	GetExportDate(ctx context.Context, userID int64) (time.Time, error)
	GetUserCategories(ctx context.Context, userId int) ([]CategoryResponse, error)
	GetUserData(ctx context.Context, userID int64) (User, error)
	GetUserProfile(ctx context.Context, userId int64, currentUserId int64) (User, int64, error)
	GetUserSkills(ctx context.Context, userID int64) ([]Skill, error)
	GetWorkPreferences(ctx context.Context, userId int64) (*WorkPreferences, error)
	IsBlocked(ctx context.Context, blockerID, blockedID int64) (bool, error)
	IsFollowing(ctx context.Context, followerId, followedId int64) (bool, error)
	SaveCompanyInfo(ctx context.Context, name, logoURL, websiteURL string) error
	SaveWorkPreferences(ctx context.Context, wp WorkPreferences) error
	DeleteUserServices(ctx context.Context, userID int64) error
	InsertUserServices(ctx context.Context, UserUslugys []UserUslugy) ([]int64, error)
	FetchUserServices(ctx context.Context, userID int64) ([]UserUslugy, error)
	GetAllCategoriesAndSubcategories(ctx context.Context, userID int64) (UserSpecialtyResponse, error)
	RemoveSubcategoryFromUserService(ctx context.Context, userID int64, subcategoryID int64) error
	SubscribeUser(ctx context.Context, followerId, followedId int64) error
	UnblockUser(ctx context.Context, blockerID, blockedID int64) error
	UnsubscribeUser(ctx context.Context, currentUserId int64, followedId int64) error
	UpdateEmail(ctx context.Context, userID int64, userEmail string) error
	UpdateExportDate(ctx context.Context, userID int64) error
	UpdateThePasswordInTheSettings(ctx context.Context, userID int64, newPassword string) error
	UpdateProfile(ctx context.Context, userID int64, firstName, lastName, middleName, location, bio string, noAds bool) error
	UpdateUserLinks(ctx context.Context, userID int64, vk, telegram, whatsapp, web, twitter string) error
	UpdateUserSkills(ctx context.Context, userID int64, skills []int) error
	UpdateUser(ctx context.Context, userID int64, username *string, email string, noAds bool) error
	UserExists(ctx context.Context, userID int64) (bool, error)
	Users(ctx context.Context) ([]*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	RemoveUserByID(ctx context.Context, userID string) error
	FetchUsersByType(ctx context.Context, userType string) ([]User, error)
	ListUserCounts(ctx context.Context) ([]UserCount, int64, error)
	GetBots(ctx context.Context) ([]User, error)
	GetReviewsBots(ctx context.Context) ([]ReviewsBots, error)
	GetSessionUserID(ctx context.Context, sessionID string) (int64, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	GetSubscriptionsCount(ctx context.Context, userID int64) (int64, error)
	GetFollowersCount(ctx context.Context, userID int64) (int64, error)
}
