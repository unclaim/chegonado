package domain

import (
	"errors"
	"time"
)

type PersonalData struct {
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Bdate     *time.Time `json:"bdate,omitempty"`
	Sex       *string    `json:"sex,omitempty"`
	Location  *string    `json:"location,omitempty"`
	Email     string     `json:"email"`
}
type SuccessResponse struct {
	Message string
}
type Phone struct {
	Phone *string `json:"phone"` // Номер телефона пользователя.
}
type PhoneNumberUpdateRequest struct {
	Phone string `json:"phone"`
}
type Bio struct {
	Bio *string `json:"bio"`
}

type CheckUserRequest struct {
	Username *string `json:"username" example:"test_user"`
}

var (
	ErrMessageTooLong = errors.New("содержимое сообщения слишком длинное")
	ErrDatabaseError  = errors.New("ошибка базы данных")
)

// ReviewsBots представляет собой структуру отзыва.
type ReviewsBots struct {
	ReviewID     int       `json:"review_id"`              // Уникальный идентификатор отзыва
	ReviewText   string    `json:"review_text,omitzero"`   // Текст отзыва
	ReviewDate   time.Time `json:"review_date,omitzero"`   // Дата создания отзыва
	Rating       int       `json:"rating,omitzero"`        // Оценка отзыва (например, от 1 до 5)
	CustomerName string    `json:"customer_name,omitzero"` // Имя клиента, оставившего отзыв
}

// DeleteSubcategoryRequest представляет запрос на удаление подкатегории.
type DeleteSubcategoryRequest struct {
	SubcategoryID int64 `json:"subcategory_id,omitzero"` // ID подкатегории для удаления из услуги пользователя
}

// Photo представляет собой структуру для хранения информации о фотографии.
type Photo struct {
	URL string `json:"url,omitzero"` // URL фотографии
}

// AccountRequest представляет запрос на обновление аккаунта пользователя.
type AccountRequest struct {
	CSRFToken string  `json:"csrf_token,omitzero"` // CSRF-токен для защиты от подделки запросов
	Username  *string `json:"username,omitzero"`   // Имя пользователя (может быть nil)
	Email     string  `json:"email,omitzero"`      // Электронная почта пользователя
	NoAds     bool    `json:"no_ads,omitzero"`     // Флаг, указывающий на отсутствие рекламы в аккаунте
}

// CategoryResponse представляет структуру для категорий.
type CategoryResponse struct {
	ID   int    `json:"id,omitzero"`   // Уникальный идентификатор категории
	Name string `json:"name,omitzero"` // Название категории
}

// Company представляет информацию о компании.
type Company struct {
	ID          int       `json:"id,omitzero"`           // Уникальный идентификатор компании
	Name        string    `json:"name,omitzero"`         // Название компании
	LogoURL     string    `json:"logo_url,omitzero"`     // URL логотипа компании
	WebsiteURL  string    `json:"website_url,omitzero"`  // URL веб-сайта компании
	LastUpdated time.Time `json:"last_updated,omitzero"` // Дата последнего обновления информации о компании
}

// ErrorResponse представляет структуру для ошибок API.
type ErrorResponse struct {
	ErrorMessage string   `json:"errorMessage,omitzero"` // Сообщение об ошибке
	ErrorType    string   `json:"errorType,omitzero"`    // Тип ошибки (например, "validation", "not_found" и т.д.)
	StackTrace   []string `json:"stackTrace,omitzero"`   // Стек вызовов (может быть nil или пустым)
}

// MessageReq представляет структуру запроса к API для отметки сообщения как прочитанного.
type MessageReq struct {
	ID int64 `json:"id,omitzero"` // Уникальный идентификатор сообщения
}

// MessageRequest представляет структуру запроса для отправки сообщения.
type MessageRequest struct {
	RecipientID int64  `json:"recipient_id,omitzero"` // Идентификатор получателя сообщения
	Message     string `json:"message,omitzero"`      // Текст сообщения
}

// MessageResponse представляет структуру ответа API на запрос с сообщением.
type MessageResponse struct {
	StatusCode int         `json:"statusCode,omitzero"` // Код статуса HTTP-ответа
	Body       interface{} `json:"body,omitzero"`       // Тело ответа (может содержать любые данные)
}
type OrderExecutorsResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       []User `json:"body"`
	TotalCount int    `json:"totalCount"`
}
type UserSkillsResponse struct {
	UserID int64   `json:"user_id"`
	Skills []Skill `json:"skills"`
}

type ProfileResponse struct {
	Profile            User      `json:"profile"`
	Messages           []Message `json:"messages"`
	SubscriptionsCount int64     `json:"subscriptions_count"`
	IsBlocked          bool      `json:"is_blocked"`
	IsFollowing        bool      `json:"is_following"`
}

// Order представляет структуру заказа.
type Order struct {
	ID     int    `json:"id,omitzero"`     // Уникальный идентификатор заказа
	Item   string `json:"item,omitzero"`   // Название товара или услуги
	Amount int    `json:"amount,omitzero"` // Количество товара или сумма заказа
}

// PasswordRequest представляет структуру запроса для изменения пароля пользователя.
type PasswordRequest struct {
	OldPassword  string `json:"old_password,omitzero"` // Текущий пароль пользователя
	NewPassword1 string `json:"pass1,omitzero"`        // Новый пароль (первый ввод)
	NewPassword2 string `json:"pass2,omitzero"`        // Новый пароль (второй ввод для подтверждения)
}

// PasswordResetRequest представляет структуру запроса для сброса пароля по email.
type PasswordResetRequest struct {
	Email string `json:"email,omitzero"` // Электронная почта пользователя для сброса пароля
}

// Request представляет тело общего запроса с сообщением и числовым параметром.
type Request struct {
	Message string `json:"message,omitzero"` // Текст сообщения
	Number  int    `json:"number,omitzero"`  // Числовой параметр запроса
}

// ResetPasswordParams содержит параметры для сброса пароля.
type ResetPasswordParams struct {
	Token string `json:"token,omitzero"` // Токен сброса пароля, полученный пользователем
	Email string `json:"email,omitzero"` // Электронная почта пользователя, связанная с токеном
}

// ResetPasswordRequest представляет структуру запроса для сброса пароля.
type ResetPasswordRequest struct {
	Token       string `json:"token,omitzero"`        // Токен для сброса пароля
	Email       string `json:"email,omitzero"`        // Электронная почта пользователя
	NewPassword string `json:"new_password,omitzero"` // Новый пароль
}

// ResetPasswordResponse представляет структуру ответа на запрос сброса пароля.
type ResetPasswordResponse struct {
	Token string `json:"token,omitzero"` // Токен, подтверждающий успешный сброс пароля
	Email string `json:"email,omitzero"` // Электронная почта пользователя
}

// Response представляет собой тело ответа с кодом статуса и данными.
type Response struct {
	StatusCode int         `json:"statusCode,omitzero"` // Код статуса HTTP-ответа
	Body       interface{} `json:"body,omitzero"`       // Тело ответа (может содержать любые данные)
}

// SessionRequest представляет структуру запроса для работы с сессией.
type SessionRequest struct {
	Message   string `json:"message,omitzero"`    // Сообщение, связанное с сессией
	SessionID string `json:"session_id,omitzero"` // Уникальный идентификатор сессии
}

// SigninRequest представляет структуру запроса для входа пользователя.
type SigninRequest struct {
	Username     *string `json:"username,omitzero"`      // Имя пользователя (может быть nil)
	PasswordHash string  `json:"password_hash,omitzero"` // Хэш пароля пользователя
}

// SignupResponse представляет ответ на запрос регистрации нового пользователя.
type SignupResponse struct {
	Message  string `json:"message,omitzero"`  // Сообщение об успешной регистрации или ошибке (опционально)
	Error    string `json:"error,omitzero"`    // Сообщение об ошибке (опционально)
	Redirect string `json:"redirect,omitzero"` // URL для перенаправления после успешной регистрации (опционально)
}

// Skill представляет структуру навыка с идентификатором категории и навыка.
type Skill struct {
	CategoryID int `json:"category_id,omitzero"` // Идентификатор категории навыка
	SkillID    int `json:"skill_id,omitzero"`    // Идентификатор самого навыка
}

// UserService представляет услугу, предоставляемую пользователем.
type UserUslugy struct {
	ID             int64     `json:"id,omitzero"`              // Уникальный идентификатор услуги
	UserID         int64     `json:"user_id,omitzero"`         // Идентификатор пользователя, предоставляющего услугу
	CategoryID     int       `json:"category_id,omitzero"`     // Идентификатор категории услуги
	CategoryName   string    `json:"category_name,omitzero"`   // Название категории услуги
	SubcategoryIDs []int64   `json:"subcategory_ids,omitzero"` // Массив идентификаторов подкатегорий услуги
	CreatedAt      time.Time `json:"created_at,omitzero"`      // Дата и время создания услуги
	UpdatedAt      time.Time `json:"updated_at,omitzero"`      // Дата и время последнего обновления услуги
}

// UserSpecialtyResponse представляет ответ с услугами пользователя.
type UserSpecialtyResponse struct {
	ID          int64        `json:"id,omitzero"`            // Уникальный идентификатор ответа
	UserID      int64        `json:"user_id,omitzero"`       // Идентификатор пользователя, чьи услуги возвращаются в ответе
	UserUslugys []UserUslugy `json:"user_services,omitzero"` // Список услуг, предоставляемых пользователем
}

// TotalUserCount представляет общее количество пользователей.
type TotalUserCount struct {
	Total int64 `json:"total,omitzero"` // Общее количество пользователей
}

// UserCount представляет количество пользователей по типу.
type UserCount struct {
	Type  string `json:"type,omitzero"`  // Тип пользователей (например, активные, неактивные)
	Count int64  `json:"count,omitzero"` // Количество пользователей данного типа
}

// UserExport представляет информацию о пользователе и дате экспорта данных.
type UserExport struct {
	UserID     int       `json:"user_id,omitzero"`     // Уникальный идентификатор пользователя
	ExportDate time.Time `json:"export_date,omitzero"` // Дата экспорта данных пользователя
}

// UserNotFoundError представляет ошибку, когда пользователь не найден.
type UserNotFoundError struct {
	UserID int64 `json:"user_id,omitzero"` // Идентификатор пользователя, который не был найден
}

// UserResponse представляет ответ с информацией о пользователях.
type UserResponse struct {
	UserCounts     []UserCount    `json:"user_counts,omitzero"`      // Список количеств пользователей по типам
	TotalUserCount TotalUserCount `json:"total_user_count,omitzero"` // Общее количество пользователей
}

// UserSkills представляет информацию о навыках пользователя.
type UserSkills struct {
	UserID int64 `json:"user_id,omitzero"` // Уникальный идентификатор пользователя
	Skills []int `json:"skills,omitzero"`  // Список идентификаторов навыков пользователя
}

// WorkPreferences представляет предпочтения пользователя в работе.
type WorkPreferences struct {
	UserID                   int64    `json:"user_id,omitzero"`                   // Уникальный идентификатор пользователя
	Availability             string   `json:"availability,omitzero"`              // Доступность пользователя для работы
	Location                 string   `json:"location,omitzero"`                  // Местоположение пользователя
	Specializations          []string `json:"specialties,omitzero"`               // Список специализаций пользователя
	Skills                   []string `json:"skills,omitzero"`                    // Список навыков пользователя
	AvailableSpecializations []string `json:"available_specializations,omitzero"` // Доступные специализации для работы
}

// VerifyCode представляет объект, используемый для проверки кода верификации.
type VerifyCode struct {
	ID        int64     `json:"id,omitzero"`         // Уникальный идентификатор записи проверки кода.
	Email     string    `json:"email,omitzero"`      // Адрес электронной почты пользователя.
	Code      int64     `json:"code,omitzero"`       // Код верификации, отправленный пользователю.
	ExpiresAt time.Time `json:"expires_at,omitzero"` // Дата и время истечения срока действия кода верификации.
}

// User представляет пользователя с информацией о его профиле.
type User struct {
	ID             int64      `json:"id,omitzero"`              // Уникальный идентификатор пользователя в системе.
	Version        int64      `json:"ver,omitzero"`             // Версия профиля пользователя.
	Blacklisted    bool       `json:"blacklisted,omitzero"`     // Статус черного списка: true, если пользователь в черном списке, иначе false.
	Sex            *string    `json:"sex,omitzero"`             // Пол пользователя, представленный как строка (ENUM).
	FollowersCount int64      `json:"followers_count,omitzero"` // Количество подписчиков пользователя.
	Verified       bool       `json:"verified,omitzero"`        // Статус подтверждения профиля (true - подтвержден, false - не подтвержден).
	NoAds          bool       `json:"no_ads,omitzero"`          // Флаг отключения рекламы (true - реклама отключена).
	CanUploadShot  bool       `json:"can_upload_shot,omitzero"` // Флаг, указывающий, может ли пользователь загружать работы на платформу.
	Pro            bool       `json:"pro,omitzero"`             // Флаг, указывающий, является ли пользователь профессионалом (true - да).
	Type           string     `json:"type,omitzero"`            // Тип пользователя, например, "обычный" или "профессионал".
	FirstName      *string    `json:"first_name,omitzero"`      // Имя пользователя.
	LastName       *string    `json:"last_name,omitzero"`       // Фамилия пользователя.
	MiddleName     *string    `json:"middle_name,omitzero"`     // Отчество пользователя (если есть).
	Username       *string    `json:"username,omitzero"`        // Уникальное имя пользователя.
	PasswordHash   string     `json:"password_hash,omitzero"`   // Хэш пароля пользователя.
	Bdate          *time.Time `json:"bdate,omitzero"`           // Дата рождения пользователя.
	Phone          *string    `json:"phone,omitzero"`           // Номер телефона пользователя.
	Email          string     `json:"email,omitzero"`           // Электронная почта пользователя.
	HTMLURL        *string    `json:"html_url,omitzero"`        // URL-адрес профиля пользователя в формате HTML.
	AvatarURL      *string    `json:"avatar_url,omitzero"`      // URL-адрес аватара пользователя.
	Bio            *string    `json:"bio,omitzero"`             // Краткая информация о пользователе.
	Location       *string    `json:"location,omitzero"`        // Местоположение пользователя.
	CreatedAt      time.Time  `json:"created_at,omitzero"`      // Дата создания профиля пользователя.
	UpdatedAt      time.Time  `json:"updated_at,omitzero"`      // Дата последнего обновления профиля пользователя.
	Links          UserLinks  `json:"links,omitzero"`           // Внешние ссылки пользователя (веб-сайт, Twitter и др.).
	Teams          []Team     `json:"teams,omitzero"`           // Список команд, в которых состоит пользователь.
	IsFollowing    bool       `json:"is_following,omitzero"`    // Указывает, подписан ли текущий пользователь на данного пользователя.
	IsBlocked      bool       `json:"is_blocked,omitzero"`      // Указывает, заблокирован ли текущий пользователь данным пользователем.
}

// UserLinks представляет ссылки пользователя на внешние ресурсы.
type UserLinks struct {
	Web      string `json:"web,omitzero"`      // URL веб-сайта пользователя.
	Twitter  string `json:"twitter,omitzero"`  // URL профиля пользователя в Twitter.
	VK       string `json:"vk,omitzero"`       // URL профиля пользователя Вконтакте.
	Telegram string `json:"telegram,omitzero"` // URL профиля в Telegram.
	WhatsApp string `json:"whatsapp,omitzero"` // URL профиля в WhatsApp.
}

// Team представляет команду, в которой состоит пользователь.
type Team struct {
	ID        int       `json:"id,omitzero"`         // Уникальный идентификатор команды.
	Name      string    `json:"name,omitzero"`       // Название команды.
	Login     string    `json:"login,omitzero"`      // Логин команды для доступа.
	HTMLURL   string    `json:"html_url,omitzero"`   // URL профиля команды на Dribbble.
	AvatarURL string    `json:"avatar_url,omitzero"` // URL аватара команды.
	Bio       string    `json:"bio,omitzero"`        // Биография команды в формате HTML.
	Location  string    `json:"location,omitzero"`   // Местоположение команды.
	Links     UserLinks `json:"links,omitzero"`      // Внешние ссылки (например, веб-сайт и Twitter команды).
	Type      string    `json:"type,omitzero"`       // Тип команды (например, "Team").
	CreatedAt time.Time `json:"created_at,omitzero"` // Дата создания команды.
	UpdatedAt time.Time `json:"updated_at,omitzero"` // Дата последнего обновления команды.
}

// Message представляет сообщение в системе.
type Message struct {
	ID        int64     `json:"id,omitzero"`         // Уникальный идентификатор сообщения.
	SenderID  int64     `json:"sender_id,omitzero"`  // Идентификатор отправителя сообщения.
	Content   string    `json:"content,omitzero"`    // Содержимое сообщения.
	CreatedAt time.Time `json:"created_at,omitzero"` // Дата и время отправки сообщения.
	IsRead    bool      `json:"is_read,omitzero"`    // Статус прочтения сообщения (true - прочитано, false - не прочитано).
	Sender    User      `json:"sender,omitzero"`     // Добавляем информацию о пользователе, отправившем сообщение
}

// GetID возвращает уникальный идентификатор пользователя.
func (u *User) GetID() int64 {
	return u.ID
}

// GetUsrVersion возвращает версию профиля пользователя.
func (u *User) GetUsrVersion() int64 {
	return u.Version
}

// IsBlacklisted возвращает статус черного списка пользователя.
func (u *User) IsBlacklisted() bool {
	return u.Blacklisted
}

// GetSex возвращает пол пользователя.
func (u *User) GetSex() *string {
	return u.Sex
}

// GetFollowersCount возвращает количество подписчиков пользователя.
func (u *User) GetFollowersCount() int64 {
	return u.FollowersCount
}

// IsVerified возвращает статус подтверждения профиля пользователя.
func (u *User) IsVerified() bool {
	return u.Verified
}

// IsNoAds возвращает флаг отключения рекламы.
func (u *User) IsNoAds() bool {
	return u.NoAds
}

// CanUpload возвращает, может ли пользователь загружать работы на платформу.
func (u *User) CanUpload() bool {
	return u.CanUploadShot
}

// IsPro возвращает, является ли пользователь профессионалом.
func (u *User) IsPro() bool {
	return u.Pro
}

// GetType возвращает тип пользователя.
func (u *User) GetType() string {
	return u.Type
}

// GetFirstName возвращает имя пользователя.
func (u *User) GetFirstName() *string {
	return u.FirstName
}

// GetLastName возвращает фамилию пользователя.
func (u *User) GetLastName() *string {
	return u.LastName
}

// GetMiddleName возвращает отчество пользователя.
func (u *User) GetMiddleName() *string {
	return u.MiddleName
}

// GetUsername возвращает уникальное имя пользователя.
func (u *User) GetUsername() *string {
	return u.Username
}

// GetPasswordHash возвращает хэш пароля пользователя.
func (u *User) GetPasswordHash() string {
	return u.PasswordHash
}

// GetBdate возвращает дату рождения пользователя.
func (u *User) GetBdate() *time.Time {
	return u.Bdate
}

// GetPhone возвращает номер телефона пользователя.
func (u *User) GetPhone() *string {
	return u.Phone
}

// GetEmail возвращает электронную почту пользователя.
func (u *User) GetEmail() string {
	return u.Email
}

// GetHTMLURL возвращает URL-адрес профиля пользователя в формате HTML.
func (u *User) GetHTMLURL() *string {
	return u.HTMLURL
}

// GetAvatarURL возвращает URL-адрес аватара пользователя.
func (u *User) GetAvatarURL() *string {
	return u.AvatarURL
}

// GetBio возвращает краткую информацию о пользователе.
func (u *User) GetBio() *string {
	return u.Bio
}

// GetLocation возвращает местоположение пользователя.
func (u *User) GetLocation() *string {
	return u.Location
}

// GetCreatedAt возвращает дату и время создания профиля пользователя.
func (u *User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// GetUpdatedAt возвращает дату и время последнего обновления профиля пользователя.
func (u *User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

// GetLinks возвращает внешние ссылки пользователя.
func (u *User) GetLinks() UserLinks {
	return u.Links
}

// GetTeams возвращает список команд, в которых состоит пользователь.
func (u *User) GetTeams() []Team {
	return u.Teams
}
