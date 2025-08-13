package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"github.com/unclaim/chegonado/internal/shared/utils"
	"github.com/unclaim/chegonado/internal/users/domain"
	"golang.org/x/crypto/argon2"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetUserPersonalData извлекает персональные данные пользователя по его ID.
// Если пользователь не найден, возвращает nil.
func (r *UserRepository) GetUserPersonalData(ctx context.Context, userID int64) (*domain.PersonalData, error) {
	query := `SELECT first_name, last_name, bdate, sex, location, email FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, userID)
	var firstName sql.NullString
	var lastName sql.NullString
	var bdate sql.NullTime
	var sex sql.NullString
	var location sql.NullString
	var email string
	err := row.Scan(
		&firstName,
		&lastName,
		&bdate,
		&sex,
		&location,
		&email,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("ошибка получения персональных данных пользователя: %v", err)
	}
	result := &domain.PersonalData{
		FirstName: func(ns sql.NullString) *string {
			if ns.Valid {
				return &ns.String
			}
			return nil
		}(firstName),
		LastName: func(ns sql.NullString) *string {
			if ns.Valid {
				return &ns.String
			}
			return nil
		}(lastName),
		Bdate: func(nt sql.NullTime) *time.Time {
			if nt.Valid {
				return &nt.Time
			}
			return nil
		}(bdate),
		Sex: func(ns sql.NullString) *string {
			if ns.Valid {
				return &ns.String
			}
			return nil
		}(sex),
		Location: func(ns sql.NullString) *string {
			if ns.Valid {
				return &ns.String
			}
			return nil
		}(location),
		Email: email,
	}
	return result, nil
}

// UpdateUserPersonalData обновляет персональные данные пользователя.
func (r *UserRepository) UpdateUserPersonalData(ctx context.Context, userID int64, data domain.User) error {
	updateQuery := `UPDATE users SET first_name = COALESCE($1, first_name), last_name = COALESCE($2, last_name), bdate = COALESCE($3, bdate), sex = COALESCE($4, sex), location = COALESCE($5, location), email = $6 WHERE id = $7`
	_, err := r.db.Exec(ctx, updateQuery,
		data.FirstName,
		data.LastName,
		data.Bdate,
		data.Sex,
		data.Location,
		data.Email,
		userID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления данных пользователя: %v", err)
	}
	return nil
}

// GetUserPhoneNumber извлекает номер телефона пользователя по его ID.
// Возвращает nil, если номер не установлен.
func (r *UserRepository) GetUserPhoneNumber(ctx context.Context, userID int64) (*string, error) {
	query := `SELECT phone FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, userID)
	var phone sql.NullString
	err := row.Scan(&phone)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("ошибка получения номера телефона пользователя: %v", err)
	}
	if phone.Valid {
		return &phone.String, nil
	}
	return nil, nil
}

// UpdateUserPhoneNumber обновляет номер телефона пользователя.
func (r *UserRepository) UpdateUserPhoneNumber(ctx context.Context, userID int64, newPhone string) error {
	updateQuery := `UPDATE users SET phone = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, updateQuery, newPhone, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления номера телефона пользователя: %v", err)
	}
	return nil
}

// GetUserBio извлекает биографию пользователя по его ID.
// Возвращает nil, если биография не установлена.
func (r *UserRepository) GetUserBio(ctx context.Context, userID int64) (*string, error) {
	query := `SELECT bio FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, userID)
	var bio sql.NullString
	err := row.Scan(&bio)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("ошибка получения биографии пользователя: %v", err)
	}
	if bio.Valid {
		return &bio.String, nil
	}
	return nil, nil
}

// UpdateUserBio обновляет биографию пользователя.
func (r *UserRepository) UpdateUserBio(ctx context.Context, userID int64, newBio *string) error {
	updateQuery := `UPDATE users SET bio = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, updateQuery, newBio, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления биографии пользователя: %v", err)
	}
	return nil
}

// GetNewMessages извлекает новые непрочитанные сообщения для текущего пользователя.
func (r *UserRepository) GetNewMessages(ctx context.Context, currentUserId int64) ([]domain.Message, error) {
	query := `SELECT m.id, m.sender_id, m.content, m.created_at, m.is_read, u.id AS user_id, u.username, u.avatar_url FROM messages m JOIN users u ON m.sender_id = u.id WHERE m.recipient_id = $1 AND m.is_read = FALSE ORDER BY m.created_at ASC`
	rows, err := r.db.Query(context.Background(), query, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса к базе данных: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var user domain.User
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.IsRead, &user.ID, &user.Username, &user.AvatarURL); err != nil {
			return nil, fmt.Errorf("ошибка при разборе данных сообщения: %w", err)
		}
		msg.Sender = user
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при закрытии соединений: %w", err)
	}
	return messages, nil
}

// GetByLoginOrEmail ищет пользователя по логину или email.
// Возвращает nil, если пользователь не найден.
func (r *UserRepository) GetByLoginOrEmail(ctx context.Context, username string, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, username, email, password_hash FROM users WHERE username = $1 OR email = $2 LIMIT 1`, username, email)
	user, err := utils.ParseRowToUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при парсинге строки пользователя: %w", err)
	}
	return user, nil
}

// BlockUser блокирует одного пользователя другим.
func (r *UserRepository) BlockUser(ctx context.Context, blockerID, blockedID int64) error {
	if blockerID == blockedID {
		return fmt.Errorf("пользователь не может заблокировать самого себя")
	}
	_, err := r.db.Exec(
		context.Background(),
		"INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1, $2)",
		blockerID, blockedID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("пользователь с ID %d уже заблокирован пользователем с ID %d",
				blockedID, blockerID)
		}
		return fmt.Errorf("ошибка блокировки пользователя: %v", err)
	}
	return nil
}

// CreateAccountVerificationsCode создает или обновляет код подтверждения для email.
func (r *UserRepository) CreateAccountVerificationsCode(ctx context.Context, email string, code int64) error {
	result, err := r.db.Exec(ctx, `INSERT INTO account_verifications (email, verification_code) VALUES ($1, $2) ON CONFLICT ON CONSTRAINT unique_email DO UPDATE SET verification_code = EXCLUDED.verification_code;`, email, code)
	if err != nil {
		return fmt.Errorf("ошибка при создании кода подтверждения: %w", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected <= 0 {
		return fmt.Errorf("не удалось обновить или создать код подтверждения для email '%s'", email)
	}
	return nil
}

// CreateHashPass хеширует пароль с использованием Argon2.
func (r *UserRepository) CreateHashPass(ctx context.Context, plainPassword, salt string) ([]byte, error) {
	hashBytes := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	finalHash := make([]byte, len(salt)+len(hashBytes))
	copy(finalHash[:len(salt)], []byte(salt))
	copy(finalHash[len(salt):], hashBytes)
	return finalHash, nil
}

// DeleteUserByID удаляет пользователя по ID.
func (r *UserRepository) DeleteUserByID(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("недопустимый идентификатор пользователя: %d", userID)
	}
	var exists bool
	err := r.db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя с ID %d: %w", userID, err)
	}
	if !exists {
		return fmt.Errorf("пользователь с ID %d не найден в базе данных", userID)
	}
	_, err = r.db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("не удалось выполнить запрос на удаление пользователя с ID %d: %w", userID, err)
	}
	return nil
}

// FetchUsers получает список пользователей с фильтрацией и пагинацией.
// **Внимание:** Запрос переписан с использованием параметризации для защиты от SQL-инъекций.
func (r *UserRepository) FetchUsers(ctx context.Context, limit, offset int, proStr, onlineStr, categories, location string) ([]domain.User, int, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	baseQuery := `
		SELECT DISTINCT u.id, u.pro, u.type, u.username, u.avatar_url, u.first_name, u.last_name, u.bio, u.location
		FROM users u
		LEFT JOIN sessions s ON u.id = s.user_id AND s.status = 'active'
		LEFT JOIN user_skills us ON u.id = us.user_id
		WHERE u.blacklisted = false AND u.type IN ('USER', 'BOT')
	`

	countQuery := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		LEFT JOIN sessions s ON u.id = s.user_id AND s.status = 'active'
		LEFT JOIN user_skills us ON u.id = us.user_id
		WHERE u.blacklisted = false AND u.type IN ('USER', 'BOT')
	`

	if proStr != "" {
		switch proStr {
		case "true":
			conditions = append(conditions, fmt.Sprintf("u.pro = $%d", argCount))
			args = append(args, true)
			argCount++
		case "false":
			conditions = append(conditions, fmt.Sprintf("u.pro = $%d", argCount))
			args = append(args, false)
			argCount++
		}
	}

	if onlineStr == "true" {
		conditions = append(conditions, "s.id IS NOT NULL")
	}

	if categories != "" {
		categoryIDs := strings.Split(categories, ",")
		conditions = append(conditions, fmt.Sprintf("us.category_id = ANY($%d)", argCount))
		var intIDs []int
		for _, idStr := range categoryIDs {
			var id int
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
				intIDs = append(intIDs, id)
			}
		}
		args = append(args, intIDs)
		argCount++
	}

	if location != "" {
		conditions = append(conditions, fmt.Sprintf("u.location = $%d", argCount))
		args = append(args, location)
		argCount++
	}

	if len(conditions) > 0 {
		filterClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += filterClause
		countQuery += filterClause
	}

	rows, err := r.db.Query(context.Background(), baseQuery+fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1), append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Pro, &user.Type, &user.Username, &user.AvatarURL, &user.FirstName, &user.LastName, &user.Bio, &user.Location); err != nil {
			return nil, 0, fmt.Errorf("ошибка чтения строки пользователя: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка во время итерации результатов: %v", err)
	}

	var count int
	err = r.db.QueryRow(context.Background(), countQuery, args...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка подсчета количества пользователей: %v", err)
	}

	return users, count, nil
}

// GetByEmail ищет пользователя по email.
func (r *UserRepository) GetByEmail(ctx context.Context, Email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, ver, username, email FROM users WHERE email = $1;`, Email)
	return utils.ParseRowToUserReset(row)
}

// GetUserByUsername получает полную информацию о пользователе по его имени пользователя.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
        SELECT id, ver, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, first_name, last_name, middle_name, username, bdate, phone, email, avatar_url, bio, location, created_at, updated_at, links
        FROM users
        WHERE username = $1
    `, username).Scan(
		&user.ID, &user.Version, &user.Blacklisted, &user.Sex, &user.FollowersCount, &user.Verified, &user.NoAds, &user.CanUploadShot, &user.Pro, &user.Type, &user.FirstName, &user.LastName, &user.MiddleName, &user.Username, &user.Bdate, &user.Phone, &user.Email, &user.AvatarURL, &user.Bio, &user.Location, &user.CreatedAt, &user.UpdatedAt, &user.Links,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки данных пользователя: %w", err)
	}
	return &user, nil
}

// GetCompanyInfo получает информацию о компании.
func (r *UserRepository) GetCompanyInfo(ctx context.Context) (domain.Company, error) {
	var company domain.Company
	err := r.db.QueryRow(context.Background(), "SELECT id, name, logo_url, website_url, last_updated FROM company LIMIT 1").Scan(&company.ID, &company.Name, &company.LogoURL, &company.WebsiteURL, &company.LastUpdated)
	return company, err
}

// GetEmailByUserID извлекает email пользователя по его ID.
func (r *UserRepository) GetEmailByUserID(ctx context.Context, userID int64) (string, error) {
	var email string
	query := "SELECT email FROM users WHERE id = $1 LIMIT 1;"
	err := r.db.QueryRow(ctx, query, userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("пользователь с идентификатором %d не найден", userID)
		}
		return "", fmt.Errorf("ошибка при попытке получить email пользователя с идентификатором %d: %w", userID, err)
	}
	return email, nil
}

// GetExportDate получает дату последнего экспорта данных пользователя.
func (r *UserRepository) GetExportDate(ctx context.Context, userID int64) (time.Time, error) {
	var exportDate time.Time
	query := "SELECT export_date FROM exports WHERE user_id = $1 LIMIT 1;"
	err := r.db.QueryRow(ctx, query, userID).Scan(&exportDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, fmt.Errorf("экспорт данных для пользователя с идентификатором %d не обнаружен", userID)
		}
		return time.Time{}, fmt.Errorf("ошибка при попытке получить дату экспорта для пользователя с идентификатором %d: %w", userID, err)
	}
	return exportDate, nil
}

// GetUserCategories получает категории навыков пользователя.
func (r *UserRepository) GetUserCategories(ctx context.Context, userId int) ([]domain.CategoryResponse, error) {
	rows, err := r.db.Query(context.Background(), `SELECT category_id FROM user_skills WHERE user_id = $1;`, userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к таблице user_skills: %v", err)
	}
	defer rows.Close()

	var categoryIDs []int
	for rows.Next() {
		var categoryID int
		if err := rows.Scan(&categoryID); err != nil {
			return nil, fmt.Errorf("ошибка считывания id категории: %v", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после завершения чтения строк: %v", err)
	}

	if len(categoryIDs) == 0 {
		return []domain.CategoryResponse{}, nil
	}

	categoryRows, err := r.db.Query(context.Background(), `SELECT id, name FROM categories WHERE id = ANY($1);`, categoryIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к таблице categories: %v", err)
	}
	defer categoryRows.Close()

	var categories []domain.CategoryResponse
	for categoryRows.Next() {
		var category domain.CategoryResponse
		if err := categoryRows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("ошибка считывания названия категории: %v", err)
		}
		categories = append(categories, category)
	}

	if err := categoryRows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после завершения чтения строк: %v", err)
	}

	return categories, nil
}

// GetUserData получает базовые данные пользователя.
func (r *UserRepository) GetUserData(ctx context.Context, userID int64) (domain.User, error) {
	var user domain.User
	query := `SELECT username, pro, type, created_at FROM users WHERE id = $1;`
	row := r.db.QueryRow(ctx, query, userID)
	err := row.Scan(
		&user.Username, &user.Pro, &user.Type, &user.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user, fmt.Errorf("пользователь с идентификатором %d не найден", userID)
		}
		return user, fmt.Errorf("ошибка при сканировании данных пользователя с идентификатором %d: %w", userID, err)
	}
	return user, nil
}

// GetUserProfile получает полный профиль пользователя, включая информацию о подписках.
func (r *UserRepository) GetUserProfile(ctx context.Context, userId int64, currentUserId int64) (domain.User, int64, error) {
	var profile domain.User
	err := r.db.QueryRow(context.Background(),
		"SELECT id, ver, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, first_name, last_name, middle_name, username, password_hash, bdate, phone, email, html_url, avatar_url, bio, location, created_at, updated_at FROM users WHERE id = $1", userId).
		Scan(&profile.ID, &profile.Version, &profile.Blacklisted, &profile.Sex, &profile.FollowersCount, &profile.Verified, &profile.NoAds, &profile.CanUploadShot, &profile.Pro, &profile.Type, &profile.FirstName, &profile.LastName, &profile.MiddleName, &profile.Username, &profile.PasswordHash, &profile.Bdate, &profile.Phone, &profile.Email, &profile.HTMLURL, &profile.AvatarURL, &profile.Bio, &profile.Location, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return profile, 0, fmt.Errorf("пользователь с id %d не найден", userId)
		}
		return profile, 0, err
	}
	err = r.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE follower_id = $1 AND followed_id = $2)",
		currentUserId, userId).Scan(&profile.IsFollowing)
	if err != nil {
		return profile, 0, err
	}
	err = r.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM subscriptions WHERE followed_id = $1",
		userId).Scan(&profile.FollowersCount)
	if err != nil {
		return profile, 0, err
	}
	var subscriptionsCount int64
	err = r.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM subscriptions WHERE follower_id = $1",
		userId).Scan(&subscriptionsCount)
	if err != nil {
		return profile, 0, err
	}
	return profile, subscriptionsCount, nil
}

// GetUserSkills получает навыки пользователя.
func (r *UserRepository) GetUserSkills(ctx context.Context, userID int64) ([]domain.Skill, error) {
	rows, err := r.db.Query(context.Background(), `SELECT category_id, id FROM user_skills WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения навыков пользователя: %v", err)
	}
	defer rows.Close()

	var skills []domain.Skill
	for rows.Next() {
		var skill domain.Skill
		if err := rows.Scan(&skill.CategoryID, &skill.SkillID); err != nil {
			return nil, fmt.Errorf("ошибка сканирования навыка: %v", err)
		}
		if skills == nil {
			skills = []domain.Skill{}
		}
		skills = append(skills, skill)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %v", err)
	}
	return skills, nil
}

// GetWorkPreferences получает рабочие предпочтения пользователя.
func (r *UserRepository) GetWorkPreferences(ctx context.Context, userId int64) (*domain.WorkPreferences, error) {
	var wp domain.WorkPreferences
	row := r.db.QueryRow(ctx, `SELECT id, user_id, availability, location, specialties, skills FROM user_work_preferences WHERE user_id = $1`, userId)
	var specialtiesJSON, skillsJSON []byte
	err := row.Scan(&wp.UserID, &wp.UserID, &wp.Availability, &wp.Location, &specialtiesJSON, &skillsJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("не удалось получить рабочие предпочтения: %v", err)
	}
	if err := json.Unmarshal(specialtiesJSON, &wp.Specializations); err != nil {
		return nil, fmt.Errorf("не удалось преобразовать специальности из JSON: %v", err)
	}
	if err := json.Unmarshal(skillsJSON, &wp.Skills); err != nil {
		return nil, fmt.Errorf("не удалось преобразовать навыки из JSON: %v", err)
	}
	return &wp, nil
}

// IsBlocked проверяет, заблокировал ли один пользователь другого.
func (r *UserRepository) IsBlocked(ctx context.Context, blockerID, blockedID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)",
		blockerID, blockedID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки блокировки пользователя: %v", err)
	}
	return exists, nil
}

// IsFollowing проверяет, подписан ли один пользователь на другого.
func (r *UserRepository) IsFollowing(ctx context.Context, followerId, followedId int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE follower_id = $1 AND followed_id = $2)",
		followerId, followedId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки блокировки пользователя: %v", err)
	}
	return exists, nil
}

// SaveCompanyInfo сохраняет информацию о компании.
func (r *UserRepository) SaveCompanyInfo(ctx context.Context, name, logoURL, websiteURL string) error {
	_, err := r.db.Exec(context.Background(), "INSERT INTO company (name, logo_url, website_url, last_updated) VALUES ($1, $2, $3, CURRENT_TIMESTAMP) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, logo_url = EXCLUDED.logo_url, website_url = EXCLUDED.website_url, last_updated = CURRENT_TIMESTAMP", name, logoURL, websiteURL)
	return err
}

// SaveWorkPreferences сохраняет рабочие предпочтения пользователя.
func (r *UserRepository) SaveWorkPreferences(ctx context.Context, wp domain.WorkPreferences) error {
	if wp.UserID <= 0 {
		return fmt.Errorf("недопустимый идентификатор пользователя")
	}
	if wp.Availability == "" || wp.Location == "" {
		return fmt.Errorf("доступность и местоположение не могут быть пустыми")
	}
	_, err := r.db.Exec(ctx, `INSERT INTO user_work_preferences (user_id, availability, location, specialties, skills) VALUES ($1, $2, $3, $4::jsonb, $5::jsonb) ON CONFLICT (user_id) DO UPDATE SET availability = EXCLUDED.availability, location = EXCLUDED.location, specialties = COALESCE(EXCLUDED.specialties, user_work_preferences.specialties), skills = COALESCE(EXCLUDED.skills, user_work_preferences.skills)`, wp.UserID, wp.Availability, wp.Location, wp.Specializations, wp.Skills)
	if err != nil {
		return fmt.Errorf("не удалось сохранить рабочие предпочтения: %w", err)
	}
	return nil
}

// DeleteUserServices удаляет все услуги пользователя.
func (r *UserRepository) DeleteUserServices(ctx context.Context, userID int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM user_services WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления услуг пользователя с ID %d: %v", userID, err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("услуги для пользователя с ID %d не найдены", userID)
	}
	return nil
}

// InsertUserServices вставляет новые услуги пользователя.
// Сначала удаляет старые услуги, затем вставляет новые в одной транзакции.
func (r *UserRepository) InsertUserServices(ctx context.Context, UserUslugys []domain.UserUslugy) ([]int64, error) {
	var newIDs []int64
	if len(UserUslugys) == 0 {
		return newIDs, nil
	}
	userID := UserUslugys[0].UserID
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				err = fmt.Errorf("%w, rollback error: %v", err, rollbackErr)
			}
		} else {
			commitErr := tx.Commit(ctx)
			if commitErr != nil {
				err = fmt.Errorf("ошибка фиксации изменений: %w", commitErr)
			}
		}
	}()
	if _, err = tx.Exec(ctx, `DELETE FROM user_services WHERE user_id = $1`, userID); err != nil {
		return nil, fmt.Errorf("ошибка удаления существующих услуг пользователя с ID %d: %w", userID, err)
	}
	for _, service := range UserUslugys {
		var newID int64
		err = tx.QueryRow(ctx, `INSERT INTO user_services (user_id, category_id, subcategory_ids, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`, service.UserID, service.CategoryID, service.SubcategoryIDs, service.CreatedAt, service.UpdatedAt).Scan(&newID)
		if err != nil {
			return nil, fmt.Errorf("ошибка вставки услуги для пользователя с ID %d: %w", service.UserID, err)
		}
		newIDs = append(newIDs, newID)
	}
	return newIDs, nil
}

// FetchUserServices получает все услуги пользователя.
func (r *UserRepository) FetchUserServices(ctx context.Context, userID int64) ([]domain.UserUslugy, error) {
	rows, err := r.db.Query(ctx, `SELECT us.id, us.user_id, us.category_id, c.name AS category_name, us.subcategory_ids FROM user_services us JOIN categories c ON us.category_id = c.id WHERE us.user_id = $1 ORDER BY c.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var userServices []domain.UserUslugy
	for rows.Next() {
		var userService domain.UserUslugy
		var subcategoryIDs pq.Int64Array
		if err = rows.Scan(&userService.ID, &userService.UserID, &userService.CategoryID, &userService.CategoryName, &subcategoryIDs); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
		}
		userService.SubcategoryIDs = subcategoryIDs
		userServices = append(userServices, userService)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе строк: %v", err)
	}
	return userServices, nil
}

// GetAllCategoriesAndSubcategories получает все категории и подкатегории услуг пользователя.
func (r *UserRepository) GetAllCategoriesAndSubcategories(ctx context.Context, userID int64) (domain.UserSpecialtyResponse, error) {
	userServices, err := r.FetchUserServices(ctx, userID)
	if err != nil {
		return domain.UserSpecialtyResponse{}, fmt.Errorf("ошибка получения услуг пользователя: %v", err)
	}
	response := domain.UserSpecialtyResponse{
		UserID:      userID,
		UserUslugys: []domain.UserUslugy{},
	}
	categoryMap := make(map[int]domain.UserUslugy)
	for _, userService := range userServices {
		if existingService, exists := categoryMap[userService.CategoryID]; exists {
			existingService.SubcategoryIDs = append(existingService.SubcategoryIDs, userService.SubcategoryIDs...)
			categoryMap[userService.CategoryID] = existingService
		} else {
			categoryMap[userService.CategoryID] = userService
		}
	}
	for _, service := range categoryMap {
		response.UserUslugys = append(response.UserUslugys, service)
	}
	return response, nil
}

// RemoveSubcategoryFromUserService удаляет указанную подкатегорию из услуг пользователя.
// Если после удаления не останется подкатегорий, удаляет и саму услугу.
func (r *UserRepository) RemoveSubcategoryFromUserService(ctx context.Context, userID int64, subcategoryID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE user_services SET subcategory_ids = array_remove(subcategory_ids, $1) WHERE user_id = $2 AND $1 = ANY(subcategory_ids)`, subcategoryID, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления подкатегории: %v", err)
	}
	var categoryID int64
	err = r.db.QueryRow(ctx, `SELECT category_id FROM subcategories WHERE id = $1`, subcategoryID).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("ошибка получения категории для подкатегории: %v", err)
	}
	var count int64
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_services us WHERE us.user_id = $1 AND us.category_id = $2 AND array_length(us.subcategory_ids, 1) > 0`, userID, categoryID).Scan(&count)
	if err != nil {
		return fmt.Errorf("ошибка проверки наличия подкатегорий: %v", err)
	}
	if count == 0 {
		_, err = r.db.Exec(ctx, `DELETE FROM user_services WHERE user_id = $1 AND category_id = $2`, userID, categoryID)
		if err != nil {
			return fmt.Errorf("ошибка удаления записи из user_services: %v", err)
		}
	}
	return nil
}

// SubscribeUser добавляет подписку.
func (r *UserRepository) SubscribeUser(ctx context.Context, followerId, followedId int64) error {
	_, err := r.db.Exec(ctx, "INSERT INTO subscriptions (follower_id, followed_id) VALUES ($1, $2)", followerId, followedId)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return fmt.Errorf("подписка уже существует для follower_id: %d и followed_id: %d", followerId, followedId)
		}
		return fmt.Errorf("не удалось добавить подписку: %w", err)
	}
	return nil
}

// UnblockUser разблокирует пользователя.
func (r *UserRepository) UnblockUser(ctx context.Context, blockerID, blockedID int64) error {
	_, err := r.db.Exec(context.Background(), "DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2", blockerID, blockedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь с ID %d не был заблокирован пользователем с ID %d", blockedID, blockerID)
		} else if errors.Is(err, context.Canceled) {
			return fmt.Errorf("операция была отменена: %w", err)
		} else if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("время выполнения запроса истекло: %w", err)
		}
		return fmt.Errorf("не удалось разблокировать пользователя: %w", err)
	}
	return nil
}

// UnsubscribeUser отменяет подписку пользователя.
func (r *UserRepository) UnsubscribeUser(ctx context.Context, currentUserId int64, followedId int64) error {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE follower_id = $1 AND followed_id = $2)",
		currentUserId, followedId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("не удалось проверить существование подписки: %w", err)
	}
	if !exists {
		return fmt.Errorf("подписка не найдена для пользователя с ID %d на пользователя с ID %d", currentUserId, followedId)
	}
	result, err := r.db.Exec(ctx,
		"DELETE FROM subscriptions WHERE follower_id = $1 AND followed_id = $2",
		currentUserId, followedId)
	if err != nil {
		return fmt.Errorf("не удалось отменить подписку: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("не удалось отменить подписку, возможно она уже была отменена")
	}
	return nil
}

// UpdateEmail обновляет email пользователя.
func (r *UserRepository) UpdateEmail(ctx context.Context, userID int64, userEmail string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET email = $1 WHERE id = $2", userEmail, userID)
	return err
}

// UpdateExportDate обновляет дату экспорта данных пользователя.
func (r *UserRepository) UpdateExportDate(ctx context.Context, userID int64) error {
	query := `INSERT INTO exports (user_id, export_date) VALUES ($1, CURRENT_TIMESTAMP) ON CONFLICT (user_id) DO UPDATE SET export_date = CURRENT_TIMESTAMP;`
	if _, err := r.db.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("ошибка при обновлении даты экспорта: %w", err)
	}
	return nil
}

// UpdateThePasswordInTheSettings обновляет пароль пользователя.
// Перед обновлением сверяет новый пароль с текущим.
func (r *UserRepository) UpdateThePasswordInTheSettings(ctx context.Context, userID int64, newPassword string) error {
	var currentPasswordHash string
	err := r.db.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentPasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("пользователь с ID %d не найден", userID)
		}
		return fmt.Errorf("ошибка при получении текущего пароля пользователя: %w", err)
	}
	if err := utils.ComparePasswords(currentPasswordHash, newPassword); err != nil {
		return err
	}
	var updatedID int64
	err = r.db.QueryRow(ctx,
		"UPDATE users SET password_hash = crypt($1, gen_salt('bf')), ver = ver + 1 WHERE id = $2 RETURNING id",
		newPassword, userID).Scan(&updatedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь с ID %d не найден", userID)
		}
		return fmt.Errorf("ошибка обновления пароля в базе данных: %w", err)
	}
	return nil
}

// UpdateProfile обновляет основные данные профиля пользователя.
func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, middleName, location, bio string, noAds bool) error {
	_, err := r.db.Exec(context.Background(), `UPDATE users SET first_name = $1, last_name = $2, middle_name = $3, location = $4, bio = $5, no_ads = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $7`, firstName, lastName, middleName, location, bio, noAds, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления профиля пользователя: %v", err)
	}
	return nil
}

// userLinksCache - кэш для хранения ссылок пользователей.
var userLinksCache sync.Map

// UpdateUserLinks обновляет ссылки пользователя и сохраняет их в кэше.
func (r *UserRepository) UpdateUserLinks(ctx context.Context, userID int64, vk, telegram, whatsapp, web, twitter string) error {
	links := domain.UserLinks{
		VK:       vk,
		Telegram: telegram,
		WhatsApp: whatsapp,
		Web:      web,
		Twitter:  twitter,
	}
	userLinksCache.Store(userID, links)
	if _, err := r.db.Exec(ctx, `UPDATE users SET links = $1 WHERE id = $2`, links, userID); err != nil {
		return fmt.Errorf("ошибка обновления ссылок пользователя: %w", err)
	}
	return nil
}

// UpdateUserSkills обновляет навыки пользователя, сначала удаляя старые.
func (r *UserRepository) UpdateUserSkills(ctx context.Context, userID int64, skills []int) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM user_skills WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления старых навыков пользователя: %v", err)
	}
	if len(skills) == 0 {
		return nil
	}
	for _, skillID := range skills {
		_, err := r.db.Exec(context.Background(), `INSERT INTO user_skills (user_id, category_id) VALUES ($1, $2) ON CONFLICT (user_id, category_id) DO NOTHING`, userID, skillID)
		if err != nil {
			return fmt.Errorf("ошибка добавления навыка пользователя: %v", err)
		}
	}
	return nil
}

// UpdateUser обновляет имя пользователя, email и флаг no_ads.
func (r *UserRepository) UpdateUser(ctx context.Context, userID int64, username *string, email string, noAds bool) error {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}
	if !exists {
		return fmt.Errorf("пользователь с ID %d не найден", userID)
	}
	if username != nil {
		_, err = r.db.Exec(ctx, "UPDATE users SET username=$1, email=$2, no_ads=$3 WHERE id=$4", *username, email, noAds, userID)
	} else {
		_, err = r.db.Exec(ctx, "UPDATE users SET email=$1, no_ads=$2 WHERE id=$3", email, noAds, userID)
	}
	if err != nil {
		return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}
	return nil
}

// UserExists проверяет, существует ли пользователь с заданным ID.
func (r *UserRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования пользователя: %v", err)
	}
	return exists, nil
}

// Users получает список всех пользователей.
func (r *UserRepository) Users(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, ver, blacklisted, followers_count, sex, username, first_name, last_name, middle_name, password_hash, location, bio, bdate, phone, email, avatar_url, verified, no_ads, can_upload_shot, pro, created_at, updated_at FROM users ORDER BY random();`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Version,
			&user.Blacklisted,
			&user.FollowersCount,
			&user.Sex,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.MiddleName,
			&user.PasswordHash,
			&user.Location,
			&user.Bio,
			&user.Bdate,
			&user.Phone,
			&user.Email,
			&user.AvatarURL,
			&user.Verified,
			&user.NoAds,
			&user.CanUploadShot,
			&user.Pro,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных пользователя: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}
	return users, nil
}

// GetUserByID получает полную информацию о пользователе по его ID.
func (r *UserRepository) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
        SELECT id, ver, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, first_name, last_name, middle_name, username, bdate, phone, email, avatar_url, bio, location, created_at, updated_at, links
        FROM users
        WHERE id = $1
    `, userID).Scan(
		&user.ID, &user.Version, &user.Blacklisted, &user.Sex, &user.FollowersCount, &user.Verified, &user.NoAds, &user.CanUploadShot, &user.Pro, &user.Type, &user.FirstName, &user.LastName, &user.MiddleName, &user.Username, &user.Bdate, &user.Phone, &user.Email, &user.AvatarURL, &user.Bio, &user.Location, &user.CreatedAt, &user.UpdatedAt, &user.Links,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки данных пользователя: %w", err)
	}
	return &user, nil
}

// RemoveUserByID удаляет пользователя по ID.
func (r *UserRepository) RemoveUserByID(ctx context.Context, userID string) error {
	result, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// FetchUsersByType получает всех пользователей определенного типа.
func (r *UserRepository) FetchUsersByType(ctx context.Context, userType string) ([]domain.User, error) {
	query := ` SELECT id, first_name, last_name, username, email, created_at FROM users WHERE type = $1; `

	rows, err := r.db.Query(ctx, query, userType)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("не удалось считать данные пользователя: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return users, nil
}

// ListUserCounts получает количество пользователей по типам и общее количество.
func (r *UserRepository) ListUserCounts(ctx context.Context) ([]domain.UserCount, int64, error) {
	var userCounts []domain.UserCount
	var totalCount int64

	query := ` SELECT type, COUNT(*) AS count FROM users GROUP BY type ORDER BY type ASC; `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userCount domain.UserCount
		if err := rows.Scan(&userCount.Type, &userCount.Count); err != nil {
			return nil, 0, fmt.Errorf("не удалось считать данные: %w", err)
		}
		userCounts = append(userCounts, userCount)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка при считывании данных: %w", err)
	}

	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось получить общее количество пользователей: %w", err)
	}

	return userCounts, totalCount, nil
}

// GetBots получает список всех ботов.
func (r *UserRepository) GetBots(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, ` SELECT id, version, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, first_name, last_name, middle_name, username, password_hash, bdate, phone, email, html_url, avatar_url, bio, location, created_at, updated_at FROM users WHERE type = $1`, "BOT")
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Version, &user.Blacklisted, &user.Sex, &user.FollowersCount,
			&user.Verified, &user.NoAds, &user.CanUploadShot, &user.Pro, &user.Type,
			&user.FirstName, &user.LastName, &user.MiddleName, &user.Username,
			&user.PasswordHash, &user.Bdate, &user.Phone, &user.Email, &user.HTMLURL,
			&user.AvatarURL, &user.Bio, &user.Location, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("не удалось просканировать данные: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("не удалось завершить обработку: %w", err)
	}

	return users, nil
}

// GetReviewsBots получает отзывы для ботов.
func (r *UserRepository) GetReviewsBots(ctx context.Context) ([]domain.ReviewsBots, error) {
	rows, err := r.db.Query(ctx, "SELECT ReviewID, ReviewText, ReviewDate, Rating, CustomerName FROM reviews_bots")
	if err != nil {
		return nil, fmt.Errorf("ошибка при обращении к базе данных: %w", err)
	}
	defer rows.Close()

	var reviews []domain.ReviewsBots
	for rows.Next() {
		var review domain.ReviewsBots
		if err := rows.Scan(&review.ReviewID, &review.ReviewText, &review.ReviewDate, &review.Rating, &review.CustomerName); err != nil {
			return nil, fmt.Errorf("ошибка сканирования отзыва: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обращения к базе данных: %w", err)
	}

	return reviews, nil
}

// GetSessionUserID получает ID пользователя по ID сессии.
func (r *UserRepository) GetSessionUserID(ctx context.Context, sessionID string) (int64, error) {
	var userID int64
	err := r.db.QueryRow(ctx, `SELECT user_id FROM sessions WHERE id = $1`, sessionID).Scan(&userID)
	if err != nil {
		return -1, err
	}
	return userID, nil
}

// UpdatePassword - функция для обновления пароля, которая принимает готовый хеш.
// Логика хеширования должна быть вне репозитория.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET password_hash = $1, ver = ver + 1 WHERE id = $2",
		passwordHash, userID)

	if err != nil {
		return fmt.Errorf("ошибка обновления пароля в базе данных: %w", err)
	}
	return nil
}

// GetSubscriptionsCount получает количество подписок пользователя.
func (r *UserRepository) GetSubscriptionsCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE follower_id = $1",
		userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества подписок: %w", err)
	}
	return count, nil
}

// GetFollowersCount получает количество подписчиков пользователя.
func (r *UserRepository) GetFollowersCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE followed_id = $1",
		userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества подписчиков: %w", err)
	}
	return count, nil
}
