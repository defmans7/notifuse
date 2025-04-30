package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// TaskHandler handles HTTP requests related to tasks
type TaskHandler struct {
	taskService domain.TaskService
	publicKey   paseto.V4AsymmetricPublicKey
	logger      logger.Logger
	secretKey   string
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(
	taskService domain.TaskService,
	publicKey paseto.V4AsymmetricPublicKey,
	logger logger.Logger,
	secretKey string,
) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		publicKey:   publicKey,
		logger:      logger,
		secretKey:   secretKey,
	}
}

// RegisterRoutes registers the task-related routes
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/tasks.create", requireAuth(http.HandlerFunc(h.CreateTask)))
	mux.Handle("/api/tasks.list", requireAuth(http.HandlerFunc(h.ListTasks)))
	mux.Handle("/api/tasks.get", requireAuth(http.HandlerFunc(h.GetTask)))
	mux.Handle("/api/tasks.delete", requireAuth(http.HandlerFunc(h.DeleteTask)))
	mux.Handle("/api/tasks.executePending", http.HandlerFunc(h.ExecutePendingTasks))
	mux.Handle("/api/tasks.execute", http.HandlerFunc(h.ExecuteTask))
}

// CreateTask handles creation of a new task
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var createRequest domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := createRequest.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.taskService.CreateTask(r.Context(), createRequest.WorkspaceID, task); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to create task")
		WriteJSONError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"task": task,
	})
}

// GetTask handles retrieval of a task by ID
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var getRequest domain.GetTaskRequest
	if err := getRequest.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTask(r.Context(), getRequest.WorkspaceID, getRequest.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			h.logger.WithField("error", err.Error()).Error("Failed to get task")
			WriteJSONError(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"task": task,
	})
}

// ListTasks handles listing tasks with optional filtering
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var listRequest domain.ListTasksRequest
	if err := listRequest.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := listRequest.ToFilter()

	response, err := h.taskService.ListTasks(r.Context(), listRequest.WorkspaceID, filter)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to list tasks")
		WriteJSONError(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteTask handles deletion of a task
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var deleteRequest domain.DeleteTaskRequest
	if err := deleteRequest.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.taskService.DeleteTask(r.Context(), deleteRequest.WorkspaceID, deleteRequest.ID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			h.logger.WithField("error", err.Error()).Error("Failed to delete task")
			WriteJSONError(w, "Failed to delete task", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// ExecutePendingTasks handles the cron-triggered task execution
func (h *TaskHandler) ExecutePendingTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var executeRequest domain.ExecutePendingTasksRequest
	if err := executeRequest.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute tasks
	if err := h.taskService.ExecutePendingTasks(r.Context(), executeRequest.MaxTasks); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to execute tasks")
		WriteJSONError(w, "Failed to execute tasks", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Task execution initiated",
		"max_tasks": executeRequest.MaxTasks,
	})
}

// ExecuteTask handles execution of a single task
func (h *TaskHandler) ExecuteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var executeRequest domain.ExecuteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&executeRequest); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := executeRequest.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.taskService.ExecuteTask(r.Context(), executeRequest.WorkspaceID, executeRequest.ID); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to execute task")
		WriteJSONError(w, "Failed to execute task", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Task execution initiated",
	})
}
