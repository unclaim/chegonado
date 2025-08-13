// internal/tasks/domain/ports.go
package domain

import (
	"context"
	"time"
)

// TasksService определяет интерфейс для бизнес-логики задач.
type TasksService interface {
	CreateTask(ctx context.Context, task Task, userID int64) (int, error)
	GetTasks(page int) ([]Task, string, error)
	SearchTasks(ctx context.Context, searchStr string, filterParams map[string]interface{}, pageSize, offset int) ([]Task, error)
	CreateContract(ctx context.Context, req CreateContractRequest, creatorID int64) (int64, error)
	GetTasksResponses(ctx context.Context, userID int64) ([]Task, error)
	CreateReport(ctx context.Context, report Report) error
	CheckContract(ctx context.Context, taskID, customerID, executorID int64) (*Contract, error)
	CreateReview(ctx context.Context, review Review) (int, error)
	GetContractReportExists(ctx context.Context, contractID int) (bool, *Report, error)
	GetReviewsByUser(ctx context.Context, userID string) ([]Review, error)
	UpdateReport(ctx context.Context, contractID int64, feedback *string, confirmation *bool) error
	CancelTask(ctx context.Context, taskID, userID int64) error
	GetAllCategories(ctx context.Context) ([]Category, error)
	GetCategoryByID(ctx context.Context, id int) (string, error)
	CreateResponse(ctx context.Context, newResponse ProposedResponse) (ProposedResponse, error)
	DeleteResponse(ctx context.Context, responseID, userID int64) error
	DeleteTaskByID(ctx context.Context, userID, taskID int64) error
	TotalResponses(taskID int64) (int, error)
	GetResponsesHandler(ctx context.Context, taskID, userID int64) ([]ResponseWithUser, error)
	GetSubcategories(ctx context.Context, categoryID string) ([]Subcategory, error)
	GetTaskViewsCount(ctx context.Context, taskID int64) (int64, error)
	GetTaskHandler(ctx context.Context, taskID int64) (*Task, error)
	GetTasksByUserID(ctx context.Context, userID int64) ([]*Task, error)              // Возвращает задачи, созданные пользователем
	GetTasksCountResponses(ctx context.Context, userID int64) ([]TaskResponse, error) // Задачи, на которые пользователь откликнулся, с количеством откликов
	GetTasksCount(ctx context.Context, userID int64) (int, []TaskResponse, error)     // Общее количество созданных задач и откликов
	GetUserReviewsStats(ctx context.Context, userID string) (ReviewStats, error)
	RecordTaskView(ctx context.Context, taskID, userID int64) error
	GetResponseByTaskAndUser(ctx context.Context, taskID, userID int64) (ProposedResponse, error)
	CheckResponseView(ctx context.Context, responseID, userID int64) (bool, error)
}

// TasksRepository определяет интерфейс для доступа к данным задач.
type TasksRepository interface {
	InsertTaskIntoDB(ctx context.Context, task Task, userID int64) (int, error)
	FetchTasksFromDB(page int) ([]Task, error) // Убрали message, так как это логика сервиса
	SearchTasks(ctx context.Context, queryStr string, filterParams map[string]interface{}, limit, offset int) ([]Task, error)
	GetTaskOwner(ctx context.Context, taskID int64) (int64, error)
	GetActiveStatusID(ctx context.Context) (int64, error)
	CreateContractInDB(ctx context.Context, taskID, executorID, customerID int64, createdAt time.Time, statusID int64) (int64, error)
	GetTasksByUserID(ctx context.Context, userID int64) ([]Task, error) // Задачи, на которые пользователь откликнулся
	GetTasksUserID(ctx context.Context, userID int64) ([]Task, error)   // Задачи, созданные пользователем
	CreateReport(ctx context.Context, contractID, taskID int64, executorComments string, executionStatus bool) error
	GetContractByDetails(ctx context.Context, taskID, executorID, customerID int64) (*Contract, error)
	InsertReviewInDB(ctx context.Context, review Review) (int, error)
	CheckResponseView(ctx context.Context, responseID, userID int64) (bool, error)
	GetReportByContractID(ctx context.Context, contractID int64) (*Report, error)
	FetchReviewsByUserFromDB(userID string) ([]Review, error)
	UpdateReport(ctx context.Context, reportID int64, customerFeedback string, customerConfirmation *bool) error
	CheckTaskOwnership(ctx context.Context, taskID, userID int64) (bool, error)
	CancelTask(ctx context.Context, taskID int64) error
	GetAllCategories(ctx context.Context) ([]Category, error)
	FetchCategoriesFromDB() ([]Category, error) // Возможно, этот метод будет объединен с GetAllCategories
	GetCategoryByID(ctx context.Context, id int) (string, error)
	InsertResponseIntoDB(newResponse ProposedResponse) (ProposedResponse, error)
	DeleteResponse(ctx context.Context, responseID, userID int64) error
	DeleteTask(ctx context.Context, userID, taskID int64) error
	GetContractExists(ctx context.Context, taskID int64, executorID int64, customerID int64) (bool, error)
	TotalResponses(taskID int64) (int, error)
	GetTasks(ctx context.Context, userID int64) ([]*Task, error)

	GetResponsesByTaskID(taskID, userID int64) ([]ResponseWithUser, error)
	FetchSubcategoriesByCategoryIDFromDB(categoryID string) ([]Subcategory, error)
	CountTaskViews(ctx context.Context, taskID int64) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*Task, error)
	GetUserCreatedTasksCount(ctx context.Context, userID int64) (int, error)
	GetTaskResponsesCountByUserID(ctx context.Context, userID int64) (map[int64]int, error) // Для задач, созданных пользователем
	GetTaskResponsesCountUserID(ctx context.Context, userID int64) (map[int64]int, error)   // Для задач, на которые пользователь откликнулся
	FetchUserReviewsStatsFromDB(ctx context.Context, userID string) (ReviewStats, error)
	RecordTaskView(ctx context.Context, taskID, userID int64) error
	GetResponseByTaskAndUser(ctx context.Context, taskID, userID int64) (ProposedResponse, error)
	ResponseExists(ctx context.Context, taskID, userID int64) (bool, error) // Добавлен, чтобы сервис мог использовать
}
