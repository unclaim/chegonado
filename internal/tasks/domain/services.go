// internal/tasks/domain/services.go
package domain

import (
	"context"
	"fmt" // Для parsePage
	"time"
	// Если нужны кастомные ошибки
)

// ServiceError - Кастомная ошибка для слоя сервиса
type ServiceError struct {
	Msg  string
	Code int // HTTP-код ошибки, если применимо
	Err  error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

// TasksServiceImp implements the TasksService interface.
type TasksServiceImp struct {
	tasksRepo TasksRepository
}

// NewTasksService creates a new instance of TasksServiceImp.
func NewTasksService(repo TasksRepository) *TasksServiceImp {
	return &TasksServiceImp{
		tasksRepo: repo,
	}
}

// CreateTask создает новую задачу.
func (s *TasksServiceImp) CreateTask(ctx context.Context, task Task, userID int64) (int, error) {
	// Здесь можно добавить бизнес-логику/валидацию перед сохранением в БД
	if task.Title == "" {
		return 0, &ServiceError{Msg: "название задачи не может быть пустым", Code: 400}
	}
	if task.Description == "" {
		return 0, &ServiceError{Msg: "описание задачи не может быть пустым", Code: 400}
	}

	id, err := s.tasksRepo.InsertTaskIntoDB(ctx, task, userID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при вставке задачи в базу данных: %w", err)
	}
	return id, nil
}

// GetTasks получает список задач с пагинацией.
func (s *TasksServiceImp) GetTasks(page int) ([]Task, string, error) {
	if page < 0 {
		page = 0 // Убедимся, что страница не отрицательная
	}

	tasks, err := s.tasksRepo.FetchTasksFromDB(page)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при получении задач из БД: %w", err)
	}

	message := ""
	if len(tasks) < 5 { // Предполагаем лимит в 5, как в репозитории
		message = "Конец списка"
	}
	return tasks, message, nil
}

// SearchTasks ищет задачи.
func (s *TasksServiceImp) SearchTasks(ctx context.Context, searchStr string, filterParams map[string]interface{}, pageSize, offset int) ([]Task, error) {
	if searchStr == "" {
		return nil, &ServiceError{Msg: "параметр 'search' обязателен", Code: 400}
	}

	tasks, err := s.tasksRepo.SearchTasks(ctx, searchStr, filterParams, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске задач: %w", err)
	}
	return tasks, nil
}

// CreateContract создает контракт.
func (s *TasksServiceImp) CreateContract(ctx context.Context, req CreateContractRequest, creatorID int64) (int64, error) {
	// Проверки перед созданием контракта
	ownerID, err := s.tasksRepo.GetTaskOwner(ctx, req.TaskID)
	if err != nil {
		return 0, &ServiceError{Msg: fmt.Sprintf("не удалось найти владельца задачи: %v", err), Code: 404, Err: err}
	}

	if ownerID != creatorID {
		return 0, &ServiceError{Msg: "недостаточно прав для создания контракта", Code: 403}
	}

	// Дополнительная проверка: существует ли уже активный контракт для этой задачи и исполнителя?
	exists, err := s.tasksRepo.GetContractExists(ctx, req.TaskID, req.ExecutorID, creatorID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при проверке существования контракта: %w", err)
	}
	if exists {
		return 0, &ServiceError{Msg: "контракт для этой задачи и исполнителя уже существует", Code: 409}
	}

	statusID, err := s.tasksRepo.GetActiveStatusID(ctx)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении статуса контракта: %w", err)
	}

	contractID, err := s.tasksRepo.CreateContractInDB(ctx, req.TaskID, req.ExecutorID, creatorID, currentTime(), statusID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при создании контракта в базе данных: %w", err)
	}
	return contractID, nil
}

// GetTasksResponses получает задачи, на которые пользователь откликнулся.
func (s *TasksServiceImp) GetTasksResponses(ctx context.Context, userID int64) ([]Task, error) {
	tasks, err := s.tasksRepo.GetTasksByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении задач для userID: %w", err)
	}
	return tasks, nil
}

// CreateReport создает отчет.
func (s *TasksServiceImp) CreateReport(ctx context.Context, report Report) error {
	if report.ExecutorComments == nil || *report.ExecutorComments == "" {
		return &ServiceError{Msg: "комментарии исполнителя обязательны", Code: 400}
	}

	err := s.tasksRepo.CreateReport(ctx, report.ContractID, report.TaskID, *report.ExecutorComments, report.ExecutionStatus)
	if err != nil {
		return fmt.Errorf("ошибка при создании отчета: %w", err)
	}
	return nil
}

// CheckContract проверяет контракт.
func (s *TasksServiceImp) CheckContract(ctx context.Context, taskID, customerID, executorID int64) (*Contract, error) {
	contract, err := s.tasksRepo.GetContractByDetails(ctx, taskID, executorID, customerID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения контракта: %w", err)
	}
	// Если контракт не найден, репозиторий возвращает nil, nil. Это не ошибка на уровне сервиса.
	return contract, nil
}

// CreateReview создает отзыв.
func (s *TasksServiceImp) CreateReview(ctx context.Context, review Review) (int, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return 0, &ServiceError{Msg: "рейтинг должен быть от 1 до 5", Code: 400}
	}
	if review.Comment == "" {
		return 0, &ServiceError{Msg: "комментарий не может быть пустым", Code: 400}
	}

	reviewID, err := s.tasksRepo.InsertReviewInDB(ctx, review)
	if err != nil {
		return 0, fmt.Errorf("ошибка при попытке добавить отзыв в базу данных: %w", err)
	}
	return reviewID, nil
}

// GetContractReportExists проверяет существование отчета по контракту.
func (s *TasksServiceImp) GetContractReportExists(ctx context.Context, contractID int) (bool, *Report, error) {
	// Репозиторий должен уметь получить отчет, если он существует.
	// Здесь логика проверки существования отчета переносится из хендлера.
	report, err := s.tasksRepo.GetReportByContractID(ctx, int64(contractID))
	if err != nil {
		// Если отчет не найден, это не ошибка, а просто 'не существует'
		if err.Error() == fmt.Sprintf("отчет с contract ID %d не найден", contractID) { // Ищем конкретную строку ошибки от репозитория
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("ошибка при проверке существования отчета: %w", err)
	}
	return report != nil, report, nil
}

// GetReviewsByUser получает отзывы пользователя.
func (s *TasksServiceImp) GetReviewsByUser(ctx context.Context, userID string) ([]Review, error) {
	reviews, err := s.tasksRepo.FetchReviewsByUserFromDB(userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить отзывы: %w", err)
	}
	return reviews, nil
}

// UpdateReport обновляет отчет.
func (s *TasksServiceImp) UpdateReport(ctx context.Context, contractID int64, feedback *string, confirmation *bool) error {
	// Получаем отчет для обновления
	report, err := s.tasksRepo.GetReportByContractID(ctx, contractID)
	if err != nil {
		return &ServiceError{Msg: fmt.Sprintf("не удалось получить отчет для contract ID: %v", err), Code: 404, Err: err}
	}

	// Обновление полей отчета
	if feedback != nil {
		report.CustomerFeedback = feedback
	}
	if confirmation != nil {
		report.CustomerConfirmation = confirmation
	}

	// Валидация обновленных данных
	if report.CustomerFeedback != nil && *report.CustomerFeedback == "" {
		return &ServiceError{Msg: "отзыв заказчика не может быть пустым", Code: 400}
	}

	err = s.tasksRepo.UpdateReport(ctx, report.ID, *report.CustomerFeedback, report.CustomerConfirmation)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении отчета: %w", err)
	}
	return nil
}

// CancelTask отменяет задачу.
func (s *TasksServiceImp) CancelTask(ctx context.Context, taskID, userID int64) error {
	ownsTask, err := s.tasksRepo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("ошибка проверки владения задачей: %w", err)
	}
	if !ownsTask {
		return &ServiceError{Msg: "у вас нет прав для отмены этого задания", Code: 403}
	}

	err = s.tasksRepo.CancelTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("ошибка отмены задания: %w", err)
	}
	return nil
}

// GetAllCategories получает все категории с подкатегориями.
func (s *TasksServiceImp) GetAllCategories(ctx context.Context) ([]Category, error) {
	categories, err := s.tasksRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить категории: %w", err)
	}
	return categories, nil
}

// GetCategoryByID получает название категории по ID.
func (s *TasksServiceImp) GetCategoryByID(ctx context.Context, id int) (string, error) {
	categoryName, err := s.tasksRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("не удалось получить категорию: %w", err)
	}
	return categoryName, nil
}

// CreateResponse создает отклик на задачу.
func (s *TasksServiceImp) CreateResponse(ctx context.Context, newResponse ProposedResponse) (ProposedResponse, error) {
	// Проверка на существование отклика уже на уровне сервиса
	exists, err := s.tasksRepo.ResponseExists(ctx, newResponse.TaskID, newResponse.UserID)
	if err != nil {
		return ProposedResponse{}, fmt.Errorf("ошибка при проверке существования отклика: %w", err)
	}
	if exists {
		return ProposedResponse{}, &ServiceError{Msg: "пользователь уже ответил на эту задачу", Code: 409}
	}

	createdResponse, err := s.tasksRepo.InsertResponseIntoDB(newResponse)
	if err != nil {
		return ProposedResponse{}, fmt.Errorf("не удалось сохранить ответ: %w", err)
	}
	return createdResponse, nil
}

// DeleteResponse удаляет отклик.
func (s *TasksServiceImp) DeleteResponse(ctx context.Context, responseID, userID int64) error {
	err := s.tasksRepo.DeleteResponse(ctx, responseID, userID)
	if err != nil {
		if err.Error() == fmt.Sprintf("отклик не найден для идентификатора отклика: %d и идентификатора пользователя: %d", responseID, userID) {
			return &ServiceError{Msg: fmt.Sprintf("отклик не найден: %v", err), Code: 404, Err: err}
		}
		return fmt.Errorf("не удалось удалить ответ: %w", err)
	}
	return nil
}

// DeleteTaskByID удаляет задачу по ID.
func (s *TasksServiceImp) DeleteTaskByID(ctx context.Context, userID, taskID int64) error {
	err := s.tasksRepo.DeleteTask(ctx, userID, taskID)
	if err != nil {
		if err.Error() == fmt.Sprintf("задача не найдена с ид: %d для пользователя с ид: %d", taskID, userID) {
			return &ServiceError{Msg: fmt.Sprintf("задача не найдена: %v", err), Code: 404, Err: err}
		}
		return fmt.Errorf("не удалось удалить задачу: %w", err)
	}
	return nil
}

// TotalResponses получает общее количество откликов на задачу.
func (s *TasksServiceImp) TotalResponses(taskID int64) (int, error) {
	total, err := s.tasksRepo.TotalResponses(taskID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении общего количества откликов: %w", err)
	}
	return total, nil
}

// GetResponsesHandler получает отклики на задачу.
func (s *TasksServiceImp) GetResponsesHandler(ctx context.Context, taskID, userID int64) ([]ResponseWithUser, error) {
	responses, err := s.tasksRepo.GetResponsesByTaskID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить ответы: %w", err)
	}
	return responses, nil
}

// GetSubcategories получает подкатегории по ID категории.
func (s *TasksServiceImp) GetSubcategories(ctx context.Context, categoryID string) ([]Subcategory, error) {
	if categoryID == "" {
		return nil, &ServiceError{Msg: "идентификатор категории обязателен", Code: 400}
	}
	subcategories, err := s.tasksRepo.FetchSubcategoriesByCategoryIDFromDB(categoryID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить подкатегории: %w", err)
	}
	return subcategories, nil
}

// GetTaskViewsCount получает количество просмотров задачи.
func (s *TasksServiceImp) GetTaskViewsCount(ctx context.Context, taskID int64) (int64, error) {
	count, err := s.tasksRepo.CountTaskViews(ctx, taskID)
	if err != nil {
		return 0, fmt.Errorf("не удалось получить количество просмотров: %w", err)
	}
	return count, nil
}

// GetTaskHandler получает задачу по ID.
func (s *TasksServiceImp) GetTaskHandler(ctx context.Context, taskID int64) (*Task, error) {
	task, err := s.tasksRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задание: %w", err)
	}
	return task, nil
}

// GetTasksByUserID получает задачи, созданные пользователем.
func (s *TasksServiceImp) GetTasksByUserID(ctx context.Context, userID int64) ([]*Task, error) {
	tasks, err := s.tasksRepo.GetTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}
	return tasks, nil
}

// GetTasksCountResponses получает задачи, на которые пользователь откликнулся, с количеством откликов.
func (s *TasksServiceImp) GetTasksCountResponses(ctx context.Context, userID int64) ([]TaskResponse, error) {
	// Получаем задачи, на которые пользователь откликнулся
	tasks, err := s.tasksRepo.GetTasksByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}

	// Получаем количество откликов на эти задачи (фактически, это будет 1, так как это отклики самого пользователя)
	// Более осмысленно здесь было бы получать отклики на *созданные* пользователем задачи.
	// Исходя из названия метода `GetTasksCountResponses` и его исходного кода, он, похоже, предназначен для задач,
	// на которые *пользователь откликнулся*, и при этом показывал количество откликов (что странно для индивидуального отклика).
	// Если это количество откликов на *созданные* пользователем задачи, то нужно вызвать GetTaskResponsesCountUserID.
	// Предположим, что это отклики на задачи, на которые *пользователь откликнулся*,
	// и хотим показать, сколько откликов на *его* отклики, что не имеет смысла.
	// Скорее всего, здесь есть недопонимание в оригинальной логике.
	// Для ясности, я сделаю, как будто нужно получить количество *своих* откликов на эти задачи (т.е. 1),
	// или, если логика подразумевала подсчет откликов на задачи, *созданные* этим пользователем:

	// Если нужна карта taskID -> количество откликов *для задач, созданных этим пользователем*:
	// responseCounts, err := s.tasksRepo.GetTaskResponsesCountUserID(ctx, userID)
	// Если нужна карта taskID -> количество откликов *для задач, на которые пользователь откликнулся*:
	responseCounts, err := s.tasksRepo.GetTaskResponsesCountByUserID(ctx, userID) // Это ближе к оригинальному использованию `responseCounts`
	if err != nil {
		return nil, fmt.Errorf("не удалось получить количество откликов: %w", err)
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		count := responseCounts[task.ID] // Получаем количество откликов
		taskResponses = append(taskResponses, TaskResponse{
			Task:          task,
			ResponseCount: count,
		})
	}
	return taskResponses, nil
}

// GetTasksCount получает количество созданных задач и откликов.
func (s *TasksServiceImp) GetTasksCount(ctx context.Context, userID int64) (int, []TaskResponse, error) {
	// Получаем количество созданных задач пользователем
	createdTasksCount, err := s.tasksRepo.GetUserCreatedTasksCount(ctx, userID)
	if err != nil {
		return 0, nil, fmt.Errorf("не удалось получить количество созданных задач: %w", err)
	}

	// Получаем количество откликов на задачи *этого пользователя* (т.е. от других пользователей на его задачи)
	responseCounts, err := s.tasksRepo.GetTaskResponsesCountUserID(ctx, userID) // Это правильный метод для подсчета откликов на СОЗДАННЫЕ задачи
	if err != nil {
		return 0, nil, fmt.Errorf("не удалось получить количество откликов: %w", err)
	}

	// Формируем список ответов на задачи
	var taskResponses []TaskResponse
	for taskID, count := range responseCounts {
		taskResponses = append(taskResponses, TaskResponse{
			TaskID:        taskID,
			ResponseCount: count,
		})
	}

	return createdTasksCount, taskResponses, nil
}

// GetUserReviewsStats получает статистику отзывов пользователя.
func (s *TasksServiceImp) GetUserReviewsStats(ctx context.Context, userID string) (ReviewStats, error) {
	stats, err := s.tasksRepo.FetchUserReviewsStatsFromDB(ctx, userID)
	if err != nil {
		return ReviewStats{}, fmt.Errorf("не удалось получить статистику отзывов: %w", err)
	}
	return stats, nil
}

// RecordTaskView записывает просмотр задачи.
func (s *TasksServiceImp) RecordTaskView(ctx context.Context, taskID, userID int64) error {
	err := s.tasksRepo.RecordTaskView(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("не удалось записать просмотр задачи: %w", err)
	}
	return nil
}

// GetResponseByTaskAndUser получает отклик по задаче и пользователю.
func (s *TasksServiceImp) GetResponseByTaskAndUser(ctx context.Context, taskID, userID int64) (ProposedResponse, error) {
	response, err := s.tasksRepo.GetResponseByTaskAndUser(ctx, taskID, userID)
	if err != nil {
		return ProposedResponse{}, fmt.Errorf("ошибка при получении отклика: %w", err)
	}
	return response, nil
}

// CheckResponseView проверяет, был ли просмотрен отклик.
func (s *TasksServiceImp) CheckResponseView(ctx context.Context, responseID, userID int64) (bool, error) {
	viewed, err := s.tasksRepo.CheckResponseView(ctx, responseID, userID)
	if err != nil {
		return false, fmt.Errorf("ошибка обработки запроса: %w", err)
	}
	return viewed, nil
}

func currentTime() time.Time {
	return time.Now()
}
