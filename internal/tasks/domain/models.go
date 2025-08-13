// internal/tasks/domain/models.go
package domain

import (
	"time"
)

// TaskStatus представляет статус задачи.
type TaskStatus struct {
	ID          int    `json:"id"`
	StatusCode  int    `json:"status_code"`
	StatusName  string `json:"status_name"`
	Description string `json:"description"`
}

// Task представляет собой структуру задачи.
type Task struct {
	ID              int64      `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
	UserID          int64      `json:"user_id"`
	CategoryID      *int       `json:"category_id"`
	SubcategoryID   *int       `json:"subcategory_id"`
	Cost            *int       `json:"cost"`
	Addresses       []string   `json:"addresses"`
	ServiceLocation string     `json:"service_location"`
	PeriodType      string     `json:"period_type"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	StatusCode      int        `json:"status_code"` // 100 - Active, 101 - In Progress, 102 - Completed, 103 - Cancelled
}

// Category представляет собой структуру категории.
type Category struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Subcategories []Subcategory `json:"subcategories"`
}

// Subcategory представляет собой структуру подкатегории.
type Subcategory struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
}

// ProposedResponse представляет отклик на задачу.
type ProposedResponse struct {
	ID            int64     `json:"id"`
	TaskID        int64     `json:"task_id"`
	UserID        int64     `json:"user_id"`
	ProposedPrice int       `json:"proposed_price"`
	ResponseText  string    `json:"response_text"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserInfo представляет информацию о пользователе.
type UserInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

// ResponseWithUser представляет отклик с информацией о пользователе.
type ResponseWithUser struct {
	ID            int64     `json:"id"`
	TaskID        int64     `json:"task_id"`
	UserID        int64     `json:"user_id"`
	ProposedPrice int       `json:"proposed_price"`
	ResponseText  string    `json:"response_text"`
	CreatedAt     time.Time `json:"created_at"`
	UserInfo      UserInfo  `json:"user_info"`
	HasContract   bool      `json:"has_contract"`
}

// Contract представляет собой структуру контракта.
type Contract struct {
	ID         int64      `json:"id"`
	TaskID     int64      `json:"task_id"`
	ExecutorID int64      `json:"executor_id"`
	CustomerID int64      `json:"customer_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	IsActive   bool       `json:"is_active"`
	StatusID   int64      `json:"status_id"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
}

// CreateContractRequest представляет запрос на создание контракта.
type CreateContractRequest struct {
	TaskID     int64 `json:"task_id"`
	ExecutorID int64 `json:"executor_id"`
}

// Report представляет собой структуру отчета.
type Report struct {
	ID                   int64     `json:"id"`
	ContractID           int64     `json:"contract_id"`
	TaskID               int64     `json:"task_id"`
	ExecutorComments     *string   `json:"executor_comments"`
	CustomerFeedback     *string   `json:"customer_feedback"`
	ExecutionStatus      bool      `json:"execution_status"`
	CustomerConfirmation *bool     `json:"customer_confirmation"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Review представляет собой структуру отзыва.
type Review struct {
	ID         int64     `json:"id"`
	ContractID int64     `json:"contract_id"`
	UserID     int64     `json:"user_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReviewStats представляет статистику отзывов.
type ReviewStats struct {
	TotalReviews  int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
}

// TasksResponse - это структура для ответа GetTasks.
type TasksResponse struct {
	Tasks   []Task `json:"tasks"`
	Message string `json:"message"`
}

// SearchTasksRes - это структура для ответа SearchTasks.
type SearchTasksRes struct {
	Tasks []Task `json:"tasks"`
}

// ReportResponse - это структура для ответа GetContractReportExists.
type ReportResponse struct {
	Exists bool    `json:"exists"`
	Report *Report `json:"report"`
}

// ResponseViewResponse - это структура для ответа ResponseViews.
type ResponseViewResponse struct {
	Viewed bool `json:"viewed"`
}

// TaskResponse - это вспомогательная структура для GetTasksCountResponses
type TaskResponse struct {
	Task          Task  `json:"task"`
	ResponseCount int   `json:"response_count"`
	TaskID        int64 `json:"task_id"` // Добавляем, так как используется в GetTasksCount
}
