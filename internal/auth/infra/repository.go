package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	FileStorage "github.com/unclaim/chegonado.git/internal/filestorage/domain"
	"github.com/unclaim/chegonado.git/internal/users/domain"
	"github.com/unclaim/chegonado.git/pkg/security/session"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
	fs FileStorage.FileStorageService
}

func NewAuthRepository(db *pgxpool.Pool, fs FileStorage.FileStorageService) *AuthRepository {
	return &AuthRepository{
		db: db,
		fs: fs,
	}
}

// DeleteVerificationCode удаляет код верификации из базы данных.
func (r *AuthRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	query := `DELETE FROM verification_codes WHERE email = $1`
	_, err := r.db.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("ошибка при удалении кода верификации: %w", err)
	}
	return nil
}

// Реализация метода SaveVerificationCode
func (r *AuthRepository) SaveVerificationCode(ctx context.Context, email string, code int64, expiresAt time.Time) error {
	query := `
        INSERT INTO verification_codes (email, code, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (email) DO UPDATE SET
            code = EXCLUDED.code,
            expires_at = EXCLUDED.expires_at
    `
	// Используем EXCLUED.code для обновления, так как это более надёжный способ
	_, err := r.db.Exec(ctx, query, email, code, expiresAt)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении кода верификации: %w", err)
	}
	return nil
}

func (r *AuthRepository) CleanUpVerificationCodes(ctx context.Context) error {
	query := `DELETE FROM verification_codes WHERE expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка при очистке просроченных кодов верификации: %w", err)
	}
	log.Println("Устаревшие коды верификации очищены.")
	return nil
}

func (r *AuthRepository) CreateUser(
	ctx context.Context,
	firstname, lastname, username, email, password string,
) (*domain.User, error) {
	query := `
	INSERT INTO users (first_name, last_name, username, email, password_hash)
	VALUES($1, $2, $3, $4, crypt($5, gen_salt('bf'))) 
	RETURNING id, first_name, last_name, username, email, ver, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, bdate, phone, html_url, avatar_url, bio, location, created_at, updated_at;
	`
	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, firstname, lastname, username, email, password).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Version,
		&user.Blacklisted,
		&user.Sex,
		&user.FollowersCount,
		&user.Verified,
		&user.NoAds,
		&user.CanUploadShot,
		&user.Pro,
		&user.Type,
		&user.Bdate,
		&user.Phone,
		&user.HTMLURL,
		&user.AvatarURL,
		&user.Bio,
		&user.Location,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	// Создаем аватар
	// if err := createDefaultAvatar(email, user.ID); err != nil {
	// 	log.Printf("Не удалось создать аватар для пользователя %d: %v", user.ID, err)
	// 	// Обработка ошибки, но не блокируем создание пользователя
	// }
	if _, err := r.fs.CreateDefaultAvatar(ctx, email, user.ID); err != nil {
		log.Printf("Не удалось создать аватар для пользователя %d: %v", user.ID, err)
		// Обработка ошибки, но не блокируем создание пользователя
	}
	return user, nil
}

func (r *AuthRepository) CheckPasswordByLoginOrEmail(ctx context.Context, login, email, pass string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, ver, blacklisted, sex, followers_count, verified, no_ads, can_upload_shot, pro, type, first_name, last_name, middle_name, username, bdate, phone, email, html_url, avatar_url, bio, location, created_at, updated_at FROM users WHERE ((username = $1 OR email = $2) AND password_hash = crypt($3, password_hash))`, login, email, pass)
	return r.parseRowToUser(row)
}

// parseRowToUser считывает данные из строки базы данных и заполняет структуру User.
func (r *AuthRepository) parseRowToUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}

	err := row.Scan(
		&u.ID,
		&u.Version,
		&u.Blacklisted,
		&u.Sex,
		&u.FollowersCount,
		&u.Verified,
		&u.NoAds,
		&u.CanUploadShot,
		&u.Pro,
		&u.Type,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.Username,
		&u.Bdate,
		&u.Phone,
		&u.Email,
		&u.HTMLURL,
		&u.AvatarURL,
		&u.Bio,
		&u.Location,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования строки пользователя: %w", err)
	}

	return u, nil
}

func (r *AuthRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	sql := `
 SELECT
        id,
        ver,
        blacklisted,
        sex,
		followers_count,   
		verified,
	    no_ads,
        can_upload_shot,    
        pro,
        type, 
        first_name,
        last_name,
        middle_name,
		username,
        bdate,
        phone,
        email,
		html_url,
        avatar_url,
        bio,
		location,       
        created_at,
        updated_at              
    FROM
        users
    WHERE
        id = $1;
`
	row := r.db.QueryRow(ctx, sql, id)
	return r.parseRowToUser(row)
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// Запрос SELECT * более надёжен, чем SELECT email, так как он берёт все поля
	err := r.db.QueryRow(ctx, "SELECT id, email, password_hash, created_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		// Явно проверяем ошибку pgx.ErrNoRows, которая является обёрткой над sql.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows // Возвращаем стандартную ошибку sql
		}
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	return &user, nil
}
func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// Запрос SELECT * более надёжен, чем SELECT email, так как он берёт все поля
	err := r.db.QueryRow(ctx, "SELECT id, email FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email)
	if err != nil {
		// Явно проверяем ошибку pgx.ErrNoRows, которая является обёрткой над sql.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows // Возвращаем стандартную ошибку sql
		}
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) GetSessByID(ctx context.Context, sessionID string) (int64, error) {
	var userID int64
	query := "SELECT user_id FROM sessions WHERE id = $1"
	err := r.db.QueryRow(ctx, query, sessionID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("сессия с идентификатором %s не найдена", sessionID)
		}
		return 0, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	return userID, nil
}
func (r *AuthRepository) GetSessionsByUserID(ctx context.Context, userID int64) ([]session.Session, error) {
	query := `SELECT id, ip, browser, operating_system, created_at, first_login FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке сеансов пользователя с идентификатором %d: %w", userID, err)
	}
	defer rows.Close()
	var sessions []session.Session
	for rows.Next() {
		var sess session.Session
		if err := rows.Scan(&sess.ID, &sess.IP, &sess.Browser, &sess.OperatingSystem, &sess.CreatedAt, &sess.FirstLogin); err != nil {
			return nil, fmt.Errorf("ошибка при разбора данных о сеансах пользователя с идентификатором %d: %w", userID, err)
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при завершении выборки данных о сеансах пользователя с идентификатором %d: %w", userID, err)
	}
	return sessions, nil
}
func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID string, userID int64) (int64, error) {
	result, err := r.db.Exec(ctx, `
        DELETE FROM sessions 
        WHERE id = $1 AND user_id = $2`,
		sessionID, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
func (r *AuthRepository) CheckPasswordByUserID(ctx context.Context, uid int64, pass string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, ver, password_hash FROM users WHERE id = $1 AND password_hash = crypt($2, password_hash)`, uid, pass)
	return r.passwordIsValid(row)
}

func (r *AuthRepository) passwordIsValid(row pgx.Row) (*domain.User, error) {
	var user domain.User
	var passwordHash string
	err := row.Scan(&user.ID, &user.Version, &passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("пользователь с идентификатором %d не найден", user.ID)
		}
		return nil, fmt.Errorf("ошибка при сканировании данных пользователя: %w", err)
	}

	// Здесь логика проверки пароля уже сделана в SQL через crypt.
	// Поэтому просто возвращаем пользователя.
	return &user, nil
}

func (r *AuthRepository) ReadVerificationCode(ctx context.Context, email string) (*domain.VerifyCode, error) {
	code := &domain.VerifyCode{}
	query := `SELECT id, email, code, expires_at FROM verification_codes WHERE email = $1 LIMIT 1`
	err := r.db.QueryRow(ctx, query, email).Scan(&code.ID, &code.Email, &code.Code, &code.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("проверочный код для электронного адреса %s не найден", email)
		}
		return nil, fmt.Errorf("ошибка при получении проверочного кода: %w", err)
	}

	// Проверяем срок действия кода
	if time.Now().After(code.ExpiresAt) {
		return nil, fmt.Errorf("проверочный код для электронного адреса %s истёк", email)
	}

	return code, nil
}
func (r *AuthRepository) CheckVerifiedByUserID(ctx context.Context, uid int64) (bool, error) {
	var verified bool
	err := r.db.QueryRow(ctx, `SELECT verified FROM users WHERE id = $1 LIMIT 1`, uid).Scan(&verified)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("пользователь не найден")
		}
		return false, fmt.Errorf("ошибка при получении статуса верификации пользователя: %w", err)
	}
	return verified, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, email, newPassword string) error {
	query := `UPDATE users SET password_hash = crypt($1, gen_salt('bf')) WHERE email = $2`
	result, err := r.db.Exec(context.Background(), query, newPassword, email)
	if err != nil {
		return fmt.Errorf("ошибка обновления пароля: %w", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("пользователь с указанной электронной почтой (%s) не найден", email)
	}
	return nil
}

func (r *AuthRepository) Verified(ctx context.Context, email string) error {
	if r == nil {
		return fmt.Errorf("указатель на репозиторий равен нулю")
	}
	if r.db == nil {
		return fmt.Errorf("соединение с базой данных равно нулю")
	}
	_, err := r.db.Exec(ctx, `update users set verified = true where email = $1;`, email)
	if err != nil {
		errFmt := fmt.Errorf("не удалось обновить статус верификации: %v", err.Error())
		fmt.Println(errFmt.Error())
		return errFmt
	}
	return nil
}

func (r *AuthRepository) UpdateThePasswordInTheSettings(ctx context.Context, userID int64, newPassword string) error {
	var currentPasswordHash string
	err := r.db.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentPasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("пользователь с ID %d не найден", userID)
		}
		return fmt.Errorf("ошибка при получении текущего пароля пользователя: %w", err)
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
