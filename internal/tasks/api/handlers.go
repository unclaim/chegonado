// internal/tasks/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/unclaim/chegonado.git/internal/shared/common_errors"
	"github.com/unclaim/chegonado.git/internal/shared/utils"
	"github.com/unclaim/chegonado.git/internal/tasks/domain" // Импортируем доменные модели и интерфейсы
	"github.com/unclaim/chegonado.git/pkg/security/session"
	"github.com/unclaim/chegonado.git/pkg/security/token"
)

// TasksHandler отвечает за обработку HTTP-запросов, связанных с задачами.
type TasksHandler struct {
	TasksService domain.TasksService // Зависимость от интерфейса сервиса
	Tokens       token.TokenManager
}

// NewTasksHandler создает новый экземпляр TasksHandler.
func NewTasksHandler(service domain.TasksService, tokens token.TokenManager) *TasksHandler {
	return &TasksHandler{
		TasksService: service,
		Tokens:       tokens,
	}
}

func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common_errors.NewAppError(w, r, fmt.Errorf("метод не разрешен"), http.StatusMethodNotAllowed)
		return
	}

	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании запроса: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}

	userID := sess.UserID

	id, err := h.TasksService.CreateTask(ctx, task, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при создании задачи: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, struct {
		ID int `json:"id"`
	}{
		ID: id,
	})
}

func (h *TasksHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page, err := parsePage(pageStr) // parsePage теперь вспомогательная функция внутри api или service
	if err != nil {
		common_errors.NewAppError(w, r, err, http.StatusBadRequest)
		return
	}

	tasks, message, err := h.TasksService.GetTasks(page)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении задач: %w", err), http.StatusInternalServerError)
		return
	}

	taskResp := &domain.TasksResponse{
		Tasks:   tasks,
		Message: message,
	}

	utils.NewResponse(w, http.StatusOK, taskResp)
}

func (h *TasksHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := r.URL.Query()

	searchStr := params.Get("search")
	if searchStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("параметр 'search' обязателен"), http.StatusBadRequest)
		return
	}

	page, err := parsePage(params.Get("page"))
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный параметр 'page': %w", err), http.StatusBadRequest)
		return
	}

	const pageSize = 10
	offset := page * pageSize

	searchFilters := map[string]interface{}{}
	// Здесь можно добавить парсинг других фильтров из запроса и передать их в searchFilters

	tasks, err := h.TasksService.SearchTasks(ctx, searchStr, searchFilters, pageSize, offset)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при поиске задач: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, &domain.SearchTasksRes{Tasks: tasks})
}

func (h *TasksHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	creatorID := sess.UserID

	var req domain.CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании запроса: %w", err), http.StatusBadRequest)
		return
	}

	contractID, err := h.TasksService.CreateContract(ctx, req, creatorID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при создании контракта: %w", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}{
		Message: "Контракт успешно создан",
		ID:      contractID,
	}

	utils.NewResponse(w, http.StatusOK, response)
}

func (h *TasksHandler) GetTasksResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	tasks, err := h.TasksService.GetTasksResponses(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении задач: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании задач в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var report domain.Report

	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании запроса: %w", err), http.StatusBadRequest) // Изменил статус на 400
		return
	}

	if err := h.TasksService.CreateReport(r.Context(), report); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при создании отчета: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, map[string]string{"message": "Отчет успешно создан"})
}

func (h *TasksHandler) CheckContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 6 || pathParts[1] != "contracts" || pathParts[2] != "check" {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный формат URL"), http.StatusBadRequest)
		return
	}

	taskIDStr := pathParts[3]
	customerIDStr := pathParts[4]
	executorIDStr := pathParts[5]

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования task_id: %w", err), http.StatusBadRequest)
		return
	}

	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования customer_id: %w", err), http.StatusBadRequest)
		return
	}

	executorID, err := strconv.ParseInt(executorIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка преобразования executor_id: %w", err), http.StatusBadRequest)
		return
	}

	contract, err := h.TasksService.CheckContract(ctx, taskID, customerID, executorID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка получения контракта: %w", err), http.StatusInternalServerError)
		return
	}
	if contract == nil {
		utils.NewResponse(w, http.StatusNotFound, map[string]string{"message": "Контракт не найден"})
		return
	}

	utils.NewResponse(w, http.StatusOK, contract)
}

func (h *TasksHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common_errors.NewAppError(w, r, fmt.Errorf("метод не разрешен"), http.StatusMethodNotAllowed)
		return
	}

	var review domain.Review
	log.Println(review)
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при декодировании тела запроса: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	reviewID, err := h.TasksService.CreateReview(ctx, review)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при попытке добавить отзыв: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]int{"id": reviewID}); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при формировании JSON-ответа: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) GetContractReportExists(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := "/reports/check/"

	if !strings.HasPrefix(path, prefix) {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный путь"), http.StatusBadRequest)
		return
	}

	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("отсутствует contract_id в пути"), http.StatusBadRequest)
		return
	}

	contractID, err := strconv.Atoi(idStr)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор контракта: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	exists, report, err := h.TasksService.GetContractReportExists(ctx, contractID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при проверке существования отчета: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, domain.ReportResponse{
		Exists: exists,
		Report: report,
	})
}

func (h *TasksHandler) GetReviewsByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id") // Предполагаем, что router парсит PathValue
	if userID == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("идентификатор пользователя не предоставлен"), http.StatusBadRequest) // Изменил на 400 Bad Request
		return
	}

	reviews, err := h.TasksService.GetReviewsByUser(r.Context(), userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить отзывы: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, reviews)
}

func (h *TasksHandler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contractIDStr := r.URL.Query().Get("id")
	if contractIDStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("отсутствует параметр id в запросе"), http.StatusBadRequest)
		return
	}

	contractID, err := strconv.ParseInt(contractIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректный формат contract ID: %w", err), http.StatusBadRequest)
		return
	}

	var updateData struct {
		CustomerFeedback     *string `json:"customer_feedback,omitempty"`
		CustomerConfirmation *bool   `json:"customer_confirmation,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректная структура тела запроса: %w", err), http.StatusBadRequest)
		return
	}

	err = h.TasksService.UpdateReport(ctx, contractID, updateData.CustomerFeedback, updateData.CustomerConfirmation)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при обновлении отчета: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, "Report successfully updated")
}

func (h *TasksHandler) CancelTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common_errors.NewAppError(w, r, fmt.Errorf("метод не разрешен"), http.StatusMethodNotAllowed)
		return
	}

	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка декодирования JSON: %w", err), http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	if err := h.TasksService.CancelTask(r.Context(), task.ID, userID); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка отмены задания: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TasksHandler) Categories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.TasksService.GetAllCategories(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить категории: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, categories)
}

func (h *TasksHandler) CategoryId(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	categoryId, err := strconv.Atoi(id)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор категории: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	categoryName, err := h.TasksService.GetCategoryByID(ctx, categoryId)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить категорию: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, categoryName)
}

func (h *TasksHandler) CreateResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common_errors.NewAppError(w, r, fmt.Errorf("метод не разрешен"), http.StatusMethodNotAllowed)
		return
	}

	var newResponse domain.ProposedResponse
	if err := json.NewDecoder(r.Body).Decode(&newResponse); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый формат запроса: %w", err), http.StatusBadRequest)
		return
	}

	// Предполагается, что UserID должен быть извлечен из сессии, а не из тела запроса для безопасности.
	// Если newResponse.UserID приходит из клиента, это уязвимость.
	// Для примера, возьмем его из сессии, как в других методах.
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	newResponse.UserID = sess.UserID

	createdResponse, err := h.TasksService.CreateResponse(r.Context(), newResponse)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось сохранить ответ: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdResponse); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) DeleteResponseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	responseIDStr := r.URL.Query().Get("response_id")
	if responseIDStr == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("отсутствует идентификатор ответа"), http.StatusBadRequest)
		return
	}

	responseID, err := strconv.ParseInt(responseIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор ответа: %w", err), http.StatusBadRequest)
		return
	}

	err = h.TasksService.DeleteResponse(ctx, responseID, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось удалить ответ: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TasksHandler) DeleteTaskByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	taskIDStr := r.PathValue("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор задачи: %w", err), http.StatusBadRequest)
		return
	}

	if err := h.TasksService.DeleteTaskByID(ctx, userID, taskID); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось удалить задачу: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TasksHandler) TotalResponses(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный параметр taskID: %w", err), http.StatusBadRequest)
		return
	}

	total, err := h.TasksService.TotalResponses(taskID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении общего количества откликов: %w", err), http.StatusInternalServerError)
		return
	}

	response := map[string]int64{"total_responses": int64(total)}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

// GetCategories теперь просто заглушка или может быть удален, так как Categories() делает то же самое, но с подкатегориями
func (h TasksHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.TasksService.GetAllCategories(r.Context()) // Этот метод не определен в интерфейсе
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить категории: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(categories); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) GetResponsesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	taskIDStr := r.PathValue("task_id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор задачи: %s", taskIDStr), http.StatusBadRequest)
		return
	}

	responses, err := h.TasksService.GetResponsesHandler(ctx, taskID, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить ответы: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, responses)
}

func (h *TasksHandler) GetSubcategories(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("category_id")

	subcategories, err := h.TasksService.GetSubcategories(r.Context(), categoryID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить подкатегории: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(subcategories); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) GetTaskViewsCountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := r.PathValue("task_id")
	if taskID == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("идентификатор задания обязателен"), http.StatusBadRequest)
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор задачи: %s", taskID), http.StatusBadRequest)
		return
	}

	count, err := h.TasksService.GetTaskViewsCount(ctx, taskIDInt)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить количество просмотров: %w", err), http.StatusInternalServerError)
		return
	}

	responseBody := make(map[string]interface{})
	responseBody["views_count"] = count

	utils.NewResponse(w, http.StatusOK, responseBody)
}

func (h *TasksHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор задачи: %w", err), http.StatusBadRequest) // Изменил на 400 Bad Request
		return
	}

	task, err := h.TasksService.GetTaskHandler(ctx, taskID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить задание: %w", err), http.StatusInternalServerError) // Изменил на 500
		return
	}

	utils.NewResponse(w, http.StatusOK, task)
}

func (h *TasksHandler) GetTasksByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	tasks, err := h.TasksService.GetTasksByUserID(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить задачи: %w", err), http.StatusInternalServerError)
		return
	}

	utils.NewResponse(w, http.StatusOK, tasks)
}

func (h *TasksHandler) GetTasksCountResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	taskResponses, err := h.TasksService.GetTasksCountResponses(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить количество откликов: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(taskResponses); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) GetTasksCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	createdTasksCount, taskResponses, err := h.TasksService.GetTasksCount(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить статистику задач: %w", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		CreatedTasksCount int                   `json:"created_tasks_count"`
		TaskResponses     []domain.TaskResponse `json:"task_responses"`
	}{
		CreatedTasksCount: createdTasksCount,
		TaskResponses:     taskResponses,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) GetUserReviewsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("user_id")
	if userID == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("идентификатор пользователя обязателен"), http.StatusBadRequest)
		return
	}

	stats, err := h.TasksService.GetUserReviewsStats(ctx, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось получить статистику отзывов: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) RecordTaskViewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := r.PathValue("task_id")
	if taskID == "" {
		common_errors.NewAppError(w, r, fmt.Errorf("идентификатор задания обязателен"), http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("недопустимый идентификатор задачи: %s", taskID), http.StatusBadRequest)
		return
	}

	if err := h.TasksService.RecordTaskView(ctx, taskIDInt, userID); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("не удалось записать просмотр задачи: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h *TasksHandler) ResponseHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	ctx := r.Context()

	sess, err := session.SessionFromContext(ctx)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении сессии: %w", err), http.StatusUnauthorized)
		return
	}
	userID := sess.UserID

	intTaskID, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("неверный идентификатор задачи: %s", taskID), http.StatusBadRequest)
		return
	}

	response, err := h.TasksService.GetResponseByTaskAndUser(ctx, intTaskID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no response found") { // Проверяем на конкретную ошибку
			w.WriteHeader(http.StatusNoContent) // Возвращаем 204 No Content, если отклик не найден
			return
		}
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при получении отклика: %w", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

func (h *TasksHandler) ResponseViews(w http.ResponseWriter, r *http.Request) {
	responseIDStr := r.URL.Query().Get("response_id")
	userIDStr := r.URL.Query().Get("user_id")

	responseID, err := strconv.ParseInt(responseIDStr, 10, 64)
	if err != nil || responseID <= 0 {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректное значение параметра 'response_id': %w", err), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		common_errors.NewAppError(w, r, fmt.Errorf("некорректное значение параметра 'user_id': %w", err), http.StatusBadRequest)
		return
	}

	viewed, err := h.TasksService.CheckResponseView(r.Context(), responseID, userID)
	if err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка обработки запроса: %w", err), http.StatusInternalServerError)
		return
	}

	response := domain.ResponseViewResponse{ // Используем модель из domain
		Viewed: viewed,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common_errors.NewAppError(w, r, fmt.Errorf("ошибка при кодировании ответа в JSON: %w", err), http.StatusInternalServerError)
		return
	}
}

// parsePage вспомогательная функция для API слоя
func parsePage(pageStr string) (int, error) {
	if pageStr == "" {
		return 0, nil // по умолчанию первая страница
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		return 0, fmt.Errorf("неправильный номер страницы '%s'", pageStr)
	}
	return page, nil
}
