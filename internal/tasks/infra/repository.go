// internal/tasks/infra/repository.go
package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog" // Для GetTaskByID и SearchTasks, если там были strconv.ParseInt
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"

	"github.com/unclaim/chegonado/internal/tasks/domain" // Обновленный импорт моделей
)

// TasksRepository представляет реализацию репозитория задач для PostgreSQL.
type TasksRepository struct {
	db *pgxpool.Pool
}

// NewTasksRepository создает новый экземпляр TasksRepository.
func NewTasksRepository(db *pgxpool.Pool) *TasksRepository {
	return &TasksRepository{
		db: db,
	}
}

// GetReportByContractID получает отчет по идентификатору контракта.
func (r *TasksRepository) GetReportByContractID(ctx context.Context, contractID int64) (*domain.Report, error) {
	query := `
        SELECT id, contract_id, task_id, executor_comments, customer_feedback, execution_status, customer_confirmation, created_at, updated_at
        FROM reports
        WHERE contract_id = $1;`
	var report domain.Report
	var executorComments, customerFeedback sql.NullString
	var customerConfirmation sql.NullBool

	err := r.db.QueryRow(ctx, query, contractID).Scan(
		&report.ID,
		&report.ContractID,
		&report.TaskID,
		&executorComments,
		&customerFeedback,
		&report.ExecutionStatus,
		&customerConfirmation,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows { // Используем pgx.ErrNoRows, так как используется pgxpool
			return nil, fmt.Errorf("отчет с contract ID %d не найден", contractID)
		}
		return nil, fmt.Errorf("ошибка при получении отчета: %w", err)
	}
	if executorComments.Valid {
		report.ExecutorComments = &executorComments.String
	}
	if customerFeedback.Valid {
		report.CustomerFeedback = &customerFeedback.String
	}
	if customerConfirmation.Valid {
		report.CustomerConfirmation = &customerConfirmation.Bool
	}

	return &report, nil
}

// CheckResponseView проверяет, был ли просмотрен отклик.
func (r *TasksRepository) CheckResponseView(ctx context.Context, responseID int64, userID int64) (bool, error) {
	var viewed bool
	query := `SELECT viewed FROM response_views WHERE response_id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, query, responseID, userID).Scan(&viewed)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return viewed, nil
}

// RecordTaskView записывает просмотр задачи.
func (r *TasksRepository) RecordTaskView(ctx context.Context, taskID int64, userID int64) error {
	query := `INSERT INTO task_views (task_id, user_id) VALUES ($1, $2)`

	_, err := r.db.Exec(ctx, query, taskID, userID)
	if err != nil {
		// Проверяем на ошибку нарушения уникального ограничения, если пользователь уже просмотрел задачу
		// if pgErr, ok := err.(*pgx.PgError); ok && pgErr.Code == "23505" { // 23505 - unique_violation
		// 	return nil // Игнорируем, если запись уже существует
		// }
		return fmt.Errorf("ошибка при сохранении просмотра задачи: %w", err)
	}

	return nil
}

// InsertTaskIntoDB вставляет новую задачу в базу данных.
func (r *TasksRepository) InsertTaskIntoDB(ctx context.Context, task domain.Task, userID int64) (int, error) {
	query := `
        INSERT INTO tasks (title, description, user_id, category_id, subcategory_id, cost, addresses, service_location, period_type, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id`

	var id int
	err := r.db.QueryRow(ctx, query,
		task.Title,
		task.Description,
		userID,
		task.CategoryID,
		task.SubcategoryID,
		task.Cost,
		pq.Array(task.Addresses), // Использование pq.Array для []string
		task.ServiceLocation,
		task.PeriodType,
		task.StartDate,
		task.EndDate).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("не удалось создать задачу для пользователя с ID %d: %w", userID, err)
	}

	return id, nil
}

// InsertResponseIntoDB вставляет новый отклик в базу данных.
func (r *TasksRepository) InsertResponseIntoDB(newResponse domain.ProposedResponse) (domain.ProposedResponse, error) {
	query := `
        INSERT INTO responses (task_id, user_id, proposed_price, response_text)
        VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := r.db.QueryRow(context.Background(), query, newResponse.TaskID, newResponse.UserID, newResponse.ProposedPrice, newResponse.ResponseText).
		Scan(&newResponse.ID, &newResponse.CreatedAt)

	if err != nil {
		return domain.ProposedResponse{}, fmt.Errorf("не удалось вставить отклик для задачи с ID %d и пользователя с ID %d: %w", newResponse.TaskID, newResponse.UserID, err)
	}

	return newResponse, nil
}

// GetUserCreatedTasksCount получает количество задач, созданных пользователем.
func (r *TasksRepository) GetUserCreatedTasksCount(ctx context.Context, userID int64) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM tasks
        WHERE user_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить запрос для подсчета задач пользователя с ID %d: %w", userID, err)
	}
	return count, nil
}

// GetTasks получает задачи, созданные определенным пользователем.
func (tr *TasksRepository) GetTasks(ctx context.Context, userID int64) ([]*domain.Task, error) {
	query := ` SELECT id, title, description, created_at, user_id, category_id, subcategory_id, cost, addresses, service_location, period_type, start_date, end_date, status_code FROM tasks WHERE user_id = $1`

	rows, err := tr.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		var categoryID, subcategoryID, cost sql.NullInt64
		var startDate, endDate sql.NullTime
		var addresses pq.StringArray // Используем pq.StringArray
		var statusCode sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.CreatedAt,
			&task.UserID, &categoryID, &subcategoryID, &cost,
			&addresses, &task.ServiceLocation, // Сканируем в pq.StringArray
			&task.PeriodType, &startDate, &endDate,
			&statusCode,
		)
		if err != nil {
			return nil, err
		}
		task.Addresses = []string(addresses) // Преобразуем pq.StringArray обратно в []string

		if categoryID.Valid {
			task.CategoryID = new(int)
			*task.CategoryID = int(categoryID.Int64)
		}
		if subcategoryID.Valid {
			task.SubcategoryID = new(int)
			*task.SubcategoryID = int(subcategoryID.Int64)
		}
		if cost.Valid {
			task.Cost = new(int)
			*task.Cost = int(cost.Int64)
		}
		if startDate.Valid {
			task.StartDate = new(time.Time)
			*task.StartDate = startDate.Time
		}
		if endDate.Valid {
			task.EndDate = new(time.Time)
			*task.EndDate = endDate.Time
		}
		if statusCode.Valid {
			task.StatusCode = int(statusCode.Int64)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// GetTasksByUserID получает задачи, на которые пользователь откликнулся.
func (r *TasksRepository) GetTasksByUserID(ctx context.Context, userID int64) ([]domain.Task, error) {
	query := `
        SELECT t.id, t.title, t.description, t.created_at, t.user_id,
               t.category_id, t.subcategory_id, t.cost, t.addresses,
               t.service_location, t.period_type, t.start_date, t.end_date, t.status_code
        FROM tasks t
        JOIN responses r ON t.id = r.task_id
        WHERE r.user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var categoryID, subcategoryID, cost sql.NullInt64
		var startDate, endDate sql.NullTime
		var statusCode sql.NullInt64
		var addresses pq.StringArray

		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.CreatedAt, &task.UserID,
			&categoryID, &subcategoryID, &cost, &addresses,
			&task.ServiceLocation, &task.PeriodType, &startDate, &endDate, &statusCode); err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}

		task.Addresses = []string(addresses) // Преобразуем pq.StringArray обратно в []string

		if categoryID.Valid {
			task.CategoryID = new(int)
			*task.CategoryID = int(categoryID.Int64)
		}

		if subcategoryID.Valid {
			task.SubcategoryID = new(int)
			*task.SubcategoryID = int(subcategoryID.Int64)
		}

		if cost.Valid {
			task.Cost = new(int)
			*task.Cost = int(cost.Int64)
		}

		if startDate.Valid {
			task.StartDate = new(time.Time)
			*task.StartDate = startDate.Time
		}

		if endDate.Valid {
			task.EndDate = new(time.Time)
			*task.EndDate = endDate.Time
		}

		if statusCode.Valid {
			task.StatusCode = int(statusCode.Int64)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("произошла ошибка во время итерации по строкам: %w", err)
	}

	return tasks, nil
}

// GetTasksUserID получает задачи, созданные пользователем (дублирует GetTasks, но сохраним для ясности).
func (r *TasksRepository) GetTasksUserID(ctx context.Context, userID int64) ([]domain.Task, error) {
	query := `
        SELECT id, title, description, created_at, user_id, category_id,
               subcategory_id, cost, addresses, service_location,
               period_type, start_date, end_date, status_code
        FROM tasks
        WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var categoryID, subcategoryID, cost sql.NullInt64
		var startDate, endDate sql.NullTime
		var statusCode sql.NullInt64
		var addresses pq.StringArray

		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.CreatedAt,
			&task.UserID, &categoryID, &subcategoryID,
			&cost, &addresses, &task.ServiceLocation,
			&task.PeriodType, &startDate, &endDate, &statusCode); err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}

		task.Addresses = []string(addresses)

		if categoryID.Valid {
			task.CategoryID = new(int)
			*task.CategoryID = int(categoryID.Int64)
		}

		if subcategoryID.Valid {
			task.SubcategoryID = new(int)
			*task.SubcategoryID = int(subcategoryID.Int64)
		}

		if cost.Valid {
			task.Cost = new(int)
			*task.Cost = int(cost.Int64)
		}

		if startDate.Valid {
			task.StartDate = new(time.Time)
			*task.StartDate = startDate.Time
		}

		if endDate.Valid {
			task.EndDate = new(time.Time)
			*task.EndDate = endDate.Time
		}

		if statusCode.Valid {
			task.StatusCode = int(statusCode.Int64)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("произошла ошибка во время итерации по строкам: %w", err)
	}

	return tasks, nil
}

// GetTaskStatuses получает все статусы задач.
func (r *TasksRepository) GetTaskStatuses(ctx context.Context) ([]domain.TaskStatus, error) {
	query := `
        SELECT id, status_code, status_name, description
        FROM task_statuses
    `

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []domain.TaskStatus
	for rows.Next() {
		var status domain.TaskStatus
		if err := rows.Scan(&status.ID, &status.StatusCode, &status.StatusName, &status.Description); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// GetTaskResponsesCountUserID получает количество откликов на задачи, созданные пользователем.
func (r *TasksRepository) GetTaskResponsesCountUserID(ctx context.Context, userID int64) (map[int64]int, error) {
	query := `
        SELECT r.task_id, COUNT(r.id) as response_count
        FROM responses r
        JOIN tasks t ON r.task_id = t.id
        WHERE t.user_id = $1
        GROUP BY r.task_id` // Изменено: теперь считает отклики для задач, созданных пользователем

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var taskID int64
		var count int
		if err := rows.Scan(&taskID, &count); err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}
		counts[taskID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("произошла ошибка во время итерации по строкам: %w", err)
	}

	return counts, nil
}

// GetTaskResponsesCountByUserID получает количество откликов на задачи, на которые пользователь откликнулся.
func (r *TasksRepository) GetTaskResponsesCountByUserID(ctx context.Context, userID int64) (map[int64]int, error) {
	query := `
      SELECT task_id, COUNT(*) as response_count
      FROM responses
      WHERE user_id = $1
      GROUP BY task_id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var taskID int64
		var count int
		if err := rows.Scan(&taskID, &count); err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}
		counts[taskID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("произошла ошибка во время итерации по строкам: %w", err)
	}

	return counts, nil
}

// GetTaskOwner получает владельца задачи по ее ID.
func (r *TasksRepository) GetTaskOwner(ctx context.Context, taskID int64) (int64, error) {
	var ownerID int64
	err := r.db.QueryRow(ctx, `SELECT user_id FROM tasks WHERE id = $1`, taskID).Scan(&ownerID)
	if err != nil {
		if err == pgx.ErrNoRows { // Используем pgx.ErrNoRows
			return 0, fmt.Errorf("задача с ID %d не найдена", taskID)
		}
		return 0, fmt.Errorf("ошибка при выполнении запроса для получения владельца задачи: %w", err)
	}

	return ownerID, nil
}

// GetResponsesByTaskID получает отклики на задачу по ее ID, включая информацию о пользователях.
func (r *TasksRepository) GetResponsesByTaskID(taskID, userID int64) ([]domain.ResponseWithUser, error) {
	var customerID int64
	err := r.db.QueryRow(context.Background(), "SELECT user_id FROM tasks WHERE id = $1", taskID).Scan(&customerID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказчика задачи: %w", err)
	}

	rows, err := r.db.Query(context.Background(),
		`SELECT
        r.id,
        r.task_id,
        r.user_id,
        r.proposed_price,
        r.response_text,
        r.created_at,
        u.first_name,
        u.last_name,
        u.username,
        u.avatar_url
        FROM responses r
        JOIN users u ON r.user_id = u.id
        WHERE r.task_id = $1`, taskID)

	if err != nil {
		return nil, fmt.Errorf("не удалось получить отклики: %w", err)
	}
	defer rows.Close()

	var responses []domain.ResponseWithUser
	for rows.Next() {
		var response domain.ResponseWithUser
		var firstName, lastName, username, avatarURL sql.NullString

		err := rows.Scan(&response.ID, &response.TaskID, &response.UserID, &response.ProposedPrice,
			&response.ResponseText, &response.CreatedAt, &firstName, &lastName, &username, &avatarURL)
		if err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}

		if firstName.Valid {
			response.UserInfo.FirstName = firstName.String
		}
		if lastName.Valid {
			response.UserInfo.LastName = lastName.String
		}
		if username.Valid {
			response.UserInfo.Username = username.String
		}
		if avatarURL.Valid {
			response.UserInfo.AvatarURL = avatarURL.String
		}

		var hasContract bool
		err = r.db.QueryRow(context.Background(),
			`SELECT EXISTS(
        SELECT 1
        FROM contracts
        WHERE
        task_id = $1 AND
        executor_id = $2 AND
        customer_id = $3 AND
        is_active = TRUE AND
        status_id = (SELECT id FROM contract_statuses WHERE status = 'Active')
        )`, response.TaskID, response.UserID, customerID).Scan(&hasContract)
		if err != nil {
			return nil, fmt.Errorf("не удалось проверить существование контракта: %w", err)
		}
		response.HasContract = hasContract
		responses = append(responses, response)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам откликов: %w", err)
	}

	return responses, nil
}

// GetResponseByTaskAndUser получает отклик по ID задачи и ID пользователя.
func (r *TasksRepository) GetResponseByTaskAndUser(ctx context.Context, taskID int64, userID int64) (domain.ProposedResponse, error) {
	var response domain.ProposedResponse

	query := `SELECT id, task_id, user_id, proposed_price, response_text, created_at
              FROM responses
              WHERE task_id = $1 AND user_id = $2`

	err := r.db.QueryRow(ctx, query, taskID, userID).Scan(&response.ID, &response.TaskID, &response.UserID,
		&response.ProposedPrice, &response.ResponseText, &response.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ProposedResponse{}, fmt.Errorf("no response found for task ID: %d and user ID: %d", taskID, userID)
		}
		return domain.ProposedResponse{}, err
	}

	return response, nil
}

// GetContractByDetails получает контракт по деталям задачи, исполнителя и заказчика.
func (r *TasksRepository) GetContractByDetails(ctx context.Context, taskID int64, executorID int64, customerID int64) (*domain.Contract, error) {
	var contract domain.Contract
	var startDate, endDate sql.NullTime

	err := r.db.QueryRow(ctx, `
        SELECT id, task_id, executor_id, customer_id, created_at, updated_at, is_active, status_id, start_date, end_date
        FROM contracts
        WHERE task_id = $1 AND executor_id = $2 AND customer_id = $3`, taskID, executorID, customerID).Scan(
		&contract.ID,
		&contract.TaskID,
		&contract.ExecutorID,
		&contract.CustomerID,
		&contract.CreatedAt,
		&contract.UpdatedAt,
		&contract.IsActive,
		&contract.StatusID,
		&startDate,
		&endDate,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Если контракт не найден, возвращаем nil, nil
		}
		return nil, fmt.Errorf("ошибка при выполнении запроса для получения информации о контракте: %w", err)
	}

	if startDate.Valid {
		contract.StartDate = &startDate.Time
	}
	if endDate.Valid {
		contract.EndDate = &endDate.Time
	}

	return &contract, nil
}

// GetAllCategories получает все категории с подкатегориями.
func (r *TasksRepository) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category

	query := `
SELECT c.id, c.name,
       COALESCE(json_agg(json_build_object('id', s.id, 'name', s.name, 'category_id', s.category_id)) FILTER (WHERE s.id IS NOT NULL), '[]') AS subcategories
FROM categories c
LEFT JOIN subcategories s ON s.category_id = c.id
GROUP BY c.id, c.name
ORDER BY c.id;` // Добавил category_id в json_build_object для подкатегорий

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение категорий: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category domain.Category
		var subcategoriesJSON string

		if err := rows.Scan(&category.ID, &category.Name, &subcategoriesJSON); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки результата: %w", err)
		}

		var subcategories []domain.Subcategory
		if err := json.Unmarshal([]byte(subcategoriesJSON), &subcategories); err != nil {
			return nil, fmt.Errorf("ошибка при преобразовании JSON в подкатегории: %w", err)
		}

		category.Subcategories = subcategories
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов запроса на получение категорий: %w", err)
	}

	return categories, nil
}

// GetActiveStatusID получает ID активного статуса контракта.
func (r *TasksRepository) GetActiveStatusID(ctx context.Context) (int64, error) {
	var statusID int64

	err := r.db.QueryRow(ctx, `SELECT id FROM contract_statuses WHERE status = $1`, "Active").Scan(&statusID)

	if err != nil {
		if err == pgx.ErrNoRows { // Используем pgx.ErrNoRows
			return 0, fmt.Errorf("статус 'Active' не найден в таблице contract_statuses")
		}
		return 0, fmt.Errorf("ошибка при извлечении идентификатора статуса 'Active': %w", err)
	}

	return statusID, nil
}

// FetchUserReviewsStatsFromDB получает статистику отзывов пользователя.
func (r *TasksRepository) FetchUserReviewsStatsFromDB(ctx context.Context, userID string) (domain.ReviewStats, error) {
	query := `
      SELECT
        COUNT(*) AS total_reviews,
        COALESCE(AVG(rating), 0) AS average_rating
      FROM reviews
      WHERE user_id = $1
    `

	var stats domain.ReviewStats
	err := r.db.QueryRow(ctx, query, userID).Scan(&stats.TotalReviews, &stats.AverageRating)
	if err != nil {
		return domain.ReviewStats{}, err
	}

	return stats, nil
}

// FetchTasksFromDB получает список задач с пагинацией.
func (r *TasksRepository) FetchTasksFromDB(page int) ([]domain.Task, error) {
	limit := 5
	offset := page * limit

	rows, err := r.db.Query(context.Background(), `
    SELECT
        id,
        title,
        description,
        created_at,
        user_id,
        category_id,
        subcategory_id,
        cost,
        addresses,
        service_location,
        period_type,
        start_date,
        end_date,
        status_code
    FROM
        tasks
    LIMIT $1 OFFSET $2`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение задач: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var startDate, endDate sql.NullTime
		var categoryID, subcategoryID, cost sql.NullInt64
		var statusCode sql.NullInt64
		var addresses pq.StringArray

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.CreatedAt,
			&task.UserID,
			&categoryID,
			&subcategoryID,
			&cost,
			&addresses,
			&task.ServiceLocation,
			&task.PeriodType,
			&startDate,
			&endDate,
			&statusCode,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании задачи: %w", err)
		}

		task.Addresses = []string(addresses) // Преобразуем pq.StringArray обратно в []string

		if categoryID.Valid {
			task.CategoryID = new(int)
			*task.CategoryID = int(categoryID.Int64)
		}
		if subcategoryID.Valid {
			task.SubcategoryID = new(int)
			*task.SubcategoryID = int(subcategoryID.Int64)
		}
		if cost.Valid {
			task.Cost = new(int)
			*task.Cost = int(cost.Int64)
		}
		if startDate.Valid {
			task.StartDate = new(time.Time)
			*task.StartDate = startDate.Time
		}
		if endDate.Valid {
			task.EndDate = new(time.Time)
			*task.EndDate = endDate.Time
		}
		if statusCode.Valid {
			task.StatusCode = int(statusCode.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// FetchSubcategoriesByCategoryIDFromDB получает подкатегории по ID категории.
func (r *TasksRepository) FetchSubcategoriesByCategoryIDFromDB(categoryID string) ([]domain.Subcategory, error) {
	rows, err := r.db.Query(context.Background(), "SELECT id, name, category_id FROM subcategories WHERE category_id = $1", categoryID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение подкатегорий для категории с ID %s: %w", categoryID, err)
	}
	defer rows.Close()

	var subcategories []domain.Subcategory

	for rows.Next() {
		var subcategory domain.Subcategory
		if err := rows.Scan(&subcategory.ID, &subcategory.Name, &subcategory.CategoryID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании подкатегории: %w", err)
		}
		subcategories = append(subcategories, subcategory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк результата: %w", err)
	}

	return subcategories, nil
}

// FetchReviewsByUserFromDB получает отзывы пользователя по его ID.
func (r *TasksRepository) FetchReviewsByUserFromDB(userID string) ([]domain.Review, error) {
	query := "SELECT id, contract_id, user_id, rating, comment, created_at FROM reviews WHERE user_id = $1;"

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение отзывов пользователя: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review

	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(&review.ID, &review.ContractID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании отзыва: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк результата: %w", err)
	}

	return reviews, nil
}

// FetchCategoriesFromDB получает все категории.
func (r *TasksRepository) FetchCategoriesFromDB() ([]domain.Category, error) {
	rows, err := r.db.Query(context.Background(), "SELECT id, name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение категорий: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category

	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результата: %w", err)
		}

		subcategories, err := r.FetchSubcategoriesByCategoryID(category.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении подкатегорий для категории %d: %w", category.ID, err)
		}

		category.Subcategories = subcategories

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк результата: %w", err)
	}

	return categories, nil
}

// FetchSubcategoriesByCategoryID получает подкатегории по ID категории (вспомогательный метод).
func (r *TasksRepository) FetchSubcategoriesByCategoryID(categoryID int) ([]domain.Subcategory, error) {
	rows, err := r.db.Query(context.Background(), "SELECT id, name FROM subcategories WHERE category_id = $1", categoryID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение подкатегорий: %w", err)
	}
	defer rows.Close()

	var subcategories []domain.Subcategory

	for rows.Next() {
		var subcategory domain.Subcategory
		if err := rows.Scan(&subcategory.ID, &subcategory.Name); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результата: %w", err)
		}

		subcategory.CategoryID = categoryID
		subcategories = append(subcategories, subcategory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк результата: %w", err)
	}

	return subcategories, nil
}

// DeleteTask удаляет задачу из базы данных.
func (r *TasksRepository) DeleteTask(ctx context.Context, userID int64, taskID int64) error {
	query := `
        DELETE FROM tasks
        WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, taskID, userID)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса на удаление задачи: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена с ид: %d для пользователя с ид: %d", taskID, userID)
	}

	return nil
}

// DeleteResponse удаляет отклик из базы данных.
func (r *TasksRepository) DeleteResponse(ctx context.Context, responseID int64, userID int64) error {
	query := `DELETE FROM responses WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, responseID, userID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении отклика из базы данных: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("отклик не найден для идентификатора отклика: %d и идентификатора пользователя: %d", responseID, userID)
	}

	return nil
}

// CreateReport создает отчет в базе данных.
func (r *TasksRepository) CreateReport(ctx context.Context, contractID, taskID int64, executorComments string, executionStatus bool) error {
	query := `INSERT INTO reports (contract_id, task_id, executor_comments, execution_status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, contractID, taskID, executorComments, executionStatus)
	return err
}

// CreateContractInDB создает контракт в базе данных.
func (r *TasksRepository) CreateContractInDB(ctx context.Context, taskID, executorID, customerID int64, createdAt time.Time, statusID int64) (int64, error) {
	var contractID int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO contracts (task_id, executor_id, customer_id, created_at, updated_at, is_active, status_id, start_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		taskID, executorID, customerID, createdAt, createdAt, true, statusID, createdAt).Scan(&contractID)

	if err != nil {
		return 0, fmt.Errorf("ошибка при создании контракта в базе данных: %w", err)
	}

	return contractID, nil
}

// CountTaskViews считает количество просмотров задачи.
func (r *TasksRepository) CountTaskViews(ctx context.Context, taskID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM task_views WHERE task_id = $1`
	var count int64

	err := r.db.QueryRow(ctx, query, taskID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении количества просмотров: %w", err)
	}

	return count, nil
}

// ContractReport создает отчет о контракте (заметка: дублирует CreateReport, возможно, стоит объединить или переименовать).
func (r *TasksRepository) ContractReport(ctx context.Context, contractID, taskID int64, executorComments string, executionStatus bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Debug("Ошибка отката транзакции: %v", slog.Any("проблемы с транзакцией контракта", rollbackErr))
		}
	}()
	_, err = tx.Exec(ctx, `
        INSERT INTO reports (contract_id, task_id, executor_comments, execution_status)
        VALUES ($1, $2, $3, $4)
    `, contractID, taskID, executorComments, executionStatus)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// CheckTaskOwnership проверяет владение задачей пользователем.
func (r *TasksRepository) CheckTaskOwnership(ctx context.Context, taskID int64, userID int64) (bool, error) {
	query := `
        SELECT COUNT(*)
        FROM tasks
        WHERE id = $1 AND user_id = $2
    `

	var count int
	err := r.db.QueryRow(ctx, query, taskID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("не удалось проверить принадлежность задачи с ID %d: %w", taskID, err)
	}

	return count > 0, nil
}

// CancelTask отменяет задачу.
func (r *TasksRepository) CancelTask(ctx context.Context, taskID int64) error {
	query := `
        UPDATE tasks
        SET status_code = 108
        WHERE id = $1
    `

	_, err := r.db.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("не удалось отменить задачу с ID %d: %w", taskID, err)
	}

	return nil
}

// AddResponse добавляет отклик на задачу.
func (r *TasksRepository) AddResponse(ctx context.Context, taskID int64, userID int64, proposedPrice int, responseText string) (domain.ProposedResponse, error) {
	// Эта проверка теперь должна быть в сервисе
	// exists, err := r.ResponseExists(ctx, taskID, userID)
	// if err != nil {
	// 	return domain.ProposedResponse{}, fmt.Errorf("ошибка при проверке существования отклика: %w", err)
	// }

	// if exists {
	// 	return domain.ProposedResponse{}, fmt.Errorf("пользователь уже ответил на эту задачу")
	// }

	var response domain.ProposedResponse

	query := `INSERT INTO responses (task_id, user_id, proposed_price, response_text)
        VALUES ($1, $2, $3, $4) RETURNING id, task_id, user_id, proposed_price, response_text, created_at`

	err := r.db.QueryRow(ctx, query, taskID, userID, proposedPrice, responseText).
		Scan(&response.ID, &response.TaskID, &response.UserID, &response.ProposedPrice, &response.ResponseText, &response.CreatedAt)
	if err != nil {
		return domain.ProposedResponse{}, fmt.Errorf("ошибка при добавлении отклика в базу данных: %w", err)
	}

	return response, nil
}

// GetCategoryByID получает категорию по ID.
func (r *TasksRepository) GetCategoryByID(ctx context.Context, id int) (string, error) {
	var category domain.Category

	err := r.db.QueryRow(ctx, "SELECT id, name FROM categories WHERE id = $1", id).Scan(&category.ID, &category.Name)
	if err != nil {
		if err == pgx.ErrNoRows { // Используем pgx.ErrNoRows
			return "", fmt.Errorf("категория с id %d не найдена", id)
		}
		return "", fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}

	return category.Name, nil
}

// GetContractExists проверяет существование контракта.
func (r *TasksRepository) GetContractExists(ctx context.Context, taskID int64, executorID int64, customerID int64) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM contracts
            WHERE task_id = $1 AND executor_id = $2 AND customer_id = $3
        )`, taskID, executorID, customerID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("ошибка при выполнении запроса для проверки существования контракта: %w", err)
	}

	return exists, nil
}

// InsertReviewInDB вставляет новый отзыв в базу данных.
func (r *TasksRepository) InsertReviewInDB(ctx context.Context, review domain.Review) (int, error) {
	query := `
        INSERT INTO reviews (contract_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`

	var reviewID int
	err := r.db.QueryRow(ctx, query, review.ContractID, review.UserID, review.Rating, review.Comment).Scan(&reviewID)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать отзыв для контракта с ID %d и пользователя с ID %d: %w", review.ContractID, review.UserID, err)
	}

	return reviewID, nil
}

// ResponseExists проверяет, существует ли отклик пользователя на конкретную задачу.
func (r *TasksRepository) ResponseExists(ctx context.Context, taskID int64, userID int64) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (SELECT 1 FROM responses WHERE task_id = $1 AND user_id = $2)`

	err := r.db.QueryRow(ctx, query, taskID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("не удалось проверить существование отклика для задачи с ID %d и пользователя с ID %d: %w", taskID, userID, err)
	}

	return exists, nil
}

// TotalResponses получает общее количество откликов на задачу.
func (r *TasksRepository) TotalResponses(taskID int64) (int, error) {
	var totalResponses int

	query := `SELECT COUNT(*) FROM responses WHERE task_id = $1`

	err := r.db.QueryRow(context.Background(), query, taskID).Scan(&totalResponses)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка при получении общего количества откликов для задачи с ID %d: %w", taskID, err)
	}

	return totalResponses, nil
}

// UpdateReport обновляет отчет в базе данных.
func (r *TasksRepository) UpdateReport(ctx context.Context, reportID int64, customerFeedback string, customerConfirmation *bool) error {
	query := `
        UPDATE reports
        SET customer_feedback = $1, customer_confirmation = $2, updated_at = $3
        WHERE id = $4;`

	// pgx QueryRow/Exec не принимает nil для *bool напрямую. Используем sql.NullBool.
	var nullConfirmation sql.NullBool
	if customerConfirmation != nil {
		nullConfirmation = sql.NullBool{Bool: *customerConfirmation, Valid: true}
	} else {
		nullConfirmation = sql.NullBool{Valid: false}
	}

	_, err := r.db.Exec(ctx, query, customerFeedback, nullConfirmation, time.Now(), reportID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении отчета: %w", err)
	}

	return nil
}

// GetTaskByID получает задачу по ID.
func (r *TasksRepository) GetTaskByID(ctx context.Context, id int64) (*domain.Task, error) {
	var task domain.Task

	query := `SELECT id, title, description, created_at, user_id, category_id, subcategory_id,
                cost, addresses, service_location, period_type, start_date, end_date, status_code
              FROM tasks WHERE id = $1`

	var startDate, endDate sql.NullTime
	var categoryID, subcategoryID, cost sql.NullInt64
	var statusCode sql.NullInt64
	var addresses pq.StringArray // Использование pq.StringArray

	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.CreatedAt,
		&task.UserID,
		&categoryID,
		&subcategoryID,
		&cost,
		&addresses, // Сканируем в pq.StringArray
		&task.ServiceLocation,
		&task.PeriodType,
		&startDate,
		&endDate,
		&statusCode,
	)

	if err != nil {
		if err == pgx.ErrNoRows { // Используем pgx.ErrNoRows
			return nil, fmt.Errorf("задача с ID %d не найдена", id)
		}
		return nil, fmt.Errorf("ошибка при выполнении запроса для получения задачи: %w", err)
	}

	task.Addresses = []string(addresses) // Преобразуем pq.StringArray обратно в []string

	if categoryID.Valid {
		task.CategoryID = new(int)
		*task.CategoryID = int(categoryID.Int64)
	}
	if subcategoryID.Valid {
		task.SubcategoryID = new(int)
		*task.SubcategoryID = int(subcategoryID.Int64)
	}
	if cost.Valid {
		task.Cost = new(int)
		*task.Cost = int(cost.Int64)
	}
	if startDate.Valid {
		task.StartDate = new(time.Time)
		*task.StartDate = startDate.Time
	}
	if endDate.Valid {
		task.EndDate = new(time.Time)
		*task.EndDate = endDate.Time
	}
	if statusCode.Valid {
		task.StatusCode = int(statusCode.Int64)
	}

	return &task, nil
}

// SearchTasks ищет задачи по строке запроса и фильтрам.
func (r *TasksRepository) SearchTasks(
	ctx context.Context,
	queryStr string,
	filterParams map[string]interface{},
	limit, offset int,
) ([]domain.Task, error) {
	// Базовый запрос
	sqlQuery := ` SELECT id, title, description, created_at, user_id, status_code, start_date, end_date, cost, addresses, service_location, period_type, category_id, subcategory_id FROM tasks WHERE to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery($1)`

	args := []interface{}{queryStr}
	argIndex := 2

	// Добавляем фильтры
	for k, v := range filterParams {
		fieldName := strings.ToLower(k)
		switch v := v.(type) {
		case string:
			sqlQuery += fmt.Sprintf(" AND LOWER(%s) LIKE $%d", fieldName, argIndex)
			args = append(args, "%"+strings.ToLower(v)+"%")
			argIndex++
		case time.Time:
			sqlQuery += fmt.Sprintf(" AND %s >= $%d", fieldName, argIndex)
			args = append(args, v) // time.Time будет автоматически преобразована pgx
			argIndex++
		case int: // Принимаем int
			sqlQuery += fmt.Sprintf(" AND %s = $%d", fieldName, argIndex)
			args = append(args, v)
			argIndex++
		case float64: // если есть float64
			sqlQuery += fmt.Sprintf(" AND %s = $%d", fieldName, argIndex)
			args = append(args, v)
			argIndex++
		default:
			// Пропускаем неизвестные типы фильтров
			continue
		}
	}

	sqlQuery += fmt.Sprintf("\nORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var startDate, endDate sql.NullTime
		var categoryID, subcategoryID, cost sql.NullInt64
		var statusCode sql.NullInt64
		var addresses pq.StringArray

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.CreatedAt,
			&task.UserID,
			&statusCode,
			&startDate,
			&endDate,
			&cost,
			&addresses,
			&task.ServiceLocation,
			&task.PeriodType,
			&categoryID,
			&subcategoryID,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}

		task.Addresses = []string(addresses) // Преобразуем pq.StringArray обратно в []string

		if categoryID.Valid {
			task.CategoryID = new(int)
			*task.CategoryID = int(categoryID.Int64)
		}
		if subcategoryID.Valid {
			task.SubcategoryID = new(int)
			*task.SubcategoryID = int(subcategoryID.Int64)
		}
		if cost.Valid {
			task.Cost = new(int)
			*task.Cost = int(cost.Int64)
		}
		if startDate.Valid {
			task.StartDate = new(time.Time)
			*task.StartDate = startDate.Time
		}
		if endDate.Valid {
			task.EndDate = new(time.Time)
			*task.EndDate = endDate.Time
		}
		if statusCode.Valid {
			task.StatusCode = int(statusCode.Int64)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результата: %w", err)
	}

	return tasks, nil
}
