package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// Maximum time a task can run before timing out
const defaultMaxTaskRuntime = 55 // 55 seconds

// TaskServiceConfig contains the configuration options for the task service
type TaskServiceConfig struct {
	Repository             domain.TaskRepository
	Logger                 logger.Logger
	AuthService            *AuthService
	SubtaskEndpointBaseURL string // Base URL for HTTP subtask execution
}

// TaskService manages task execution and state
type TaskService struct {
	repo                   domain.TaskRepository
	logger                 logger.Logger
	authService            *AuthService
	processors             map[string]domain.TaskProcessor
	lock                   sync.RWMutex
	subtaskEndpointBaseURL string
}

// NewTaskService creates a new task service instance
func NewTaskService(config TaskServiceConfig) (*TaskService, error) {
	if config.Repository == nil {
		return nil, fmt.Errorf("task repository is required")
	}
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Set default endpoint URL if not provided
	subtaskEndpointBaseURL := config.SubtaskEndpointBaseURL
	if subtaskEndpointBaseURL == "" {
		subtaskEndpointBaseURL = "http://localhost:8080"
	}

	return &TaskService{
		repo:                   config.Repository,
		logger:                 config.Logger,
		authService:            config.AuthService,
		processors:             make(map[string]domain.TaskProcessor),
		subtaskEndpointBaseURL: subtaskEndpointBaseURL,
	}, nil
}

// RegisterProcessor registers a task processor for a specific task type
func (s *TaskService) RegisterProcessor(processor domain.TaskProcessor) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Determine which task types this processor can handle
	for _, taskType := range getTaskTypes() {
		if processor.CanProcess(taskType) {
			s.processors[taskType] = processor
			s.logger.WithField("task_type", taskType).Info("Registered task processor")
		}
	}
}

// getTaskTypes returns all supported task types
func getTaskTypes() []string {
	// This could be expanded with more task types as needed
	return []string{
		"import_contacts",
		"export_contacts",
		"send_broadcast",
		"generate_report",
	}
}

// GetProcessor returns the processor for a given task type
func (s *TaskService) GetProcessor(taskType string) (domain.TaskProcessor, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	processor, ok := s.processors[taskType]
	if !ok {
		return nil, fmt.Errorf("no processor registered for task type: %s", taskType)
	}

	return processor, nil
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(ctx context.Context, workspace string, task *domain.Task) error {
	if task.MaxRuntime <= 0 {
		task.MaxRuntime = defaultMaxTaskRuntime
	}

	// Set default retry settings if not provided
	if task.MaxRetries <= 0 {
		task.MaxRetries = 3
	}
	if task.RetryInterval <= 0 {
		task.RetryInterval = 60 // Default to 1 minute between retries
	}

	return s.repo.Create(ctx, workspace, task)
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, workspace, id string) (*domain.Task, error) {
	return s.repo.Get(ctx, workspace, id)
}

// ListTasks lists tasks based on filter criteria
func (s *TaskService) ListTasks(ctx context.Context, workspace string, filter domain.TaskFilter) (*domain.TaskListResponse, error) {
	tasks, totalCount, err := s.repo.List(ctx, workspace, filter)
	if err != nil {
		return nil, err
	}

	// Calculate if there are more results
	hasMore := (filter.Offset + len(tasks)) < totalCount

	return &domain.TaskListResponse{
		Tasks:      tasks,
		TotalCount: totalCount,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		HasMore:    hasMore,
	}, nil
}

// DeleteTask removes a task
func (s *TaskService) DeleteTask(ctx context.Context, workspace, id string) error {
	return s.repo.Delete(ctx, workspace, id)
}

// ExecuteTasks processes a batch of pending tasks
func (s *TaskService) ExecuteTasks(ctx context.Context, maxTasks int) error {
	// Get the next batch of tasks
	if maxTasks <= 0 {
		maxTasks = 10 // Default value
	}

	tasks, err := s.repo.GetNextBatch(ctx, maxTasks)
	if err != nil {
		return fmt.Errorf("failed to get next batch of tasks: %w", err)
	}

	s.logger.WithField("task_count", len(tasks)).Info("Retrieved batch of tasks to process")

	for _, task := range tasks {
		// Create a separate context for each task with a timeout
		taskCtx, cancel := context.WithTimeout(ctx, time.Duration(task.MaxRuntime)*time.Second)

		// Handle the task in a goroutine
		go func(t *domain.Task, ctx context.Context, cancelFunc context.CancelFunc) {
			// Ensure we clean up the context and handle timeout
			defer func() {
				// Check if the context expired (timed out)
				if ctx.Err() == context.DeadlineExceeded {
					timeoutError := fmt.Errorf("task execution timed out after %d seconds", t.MaxRuntime)
					s.logger.WithField("task_id", t.ID).
						WithField("workspace_id", t.WorkspaceID).
						WithField("error", timeoutError.Error()).
						Warn("Task timed out")

					// Mark the task as failed due to timeout
					if err := s.repo.MarkAsFailed(ctx, t.WorkspaceID, t.ID, timeoutError.Error()); err != nil {
						s.logger.WithField("task_id", t.ID).
							WithField("workspace_id", t.WorkspaceID).
							WithField("error", err.Error()).
							Error("Failed to mark timed out task as failed")
					}
				}
				cancelFunc()
			}()

			if err := s.ExecuteTask(ctx, t.WorkspaceID, t.ID); err != nil {
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("error", err.Error()).
					Error("Failed to execute task")
			}
		}(task, taskCtx, cancel)
	}

	return nil
}

// createSubtasksAndTriggerHTTP creates subtasks for a task and triggers their execution via HTTP
func (s *TaskService) createSubtasksAndTriggerHTTP(ctx context.Context, task *domain.Task) error {
	processor, err := s.GetProcessor(task.Type)
	if err != nil {
		return fmt.Errorf("failed to get processor: %w", err)
	}

	// Check if the processor supports parallelization
	if !processor.SupportsParallelization() {
		return fmt.Errorf("task type %s does not support parallelization", task.Type)
	}

	// Determine how many subtasks to create - this could be dynamic based on the task or processor
	subtaskCount := 5 // Default value, could be determined by the processor or task data
	if task.State != nil && task.State.Progress > 0 {
		// Use the subtask count from the task itself
		if task.SubtaskCount > 0 {
			subtaskCount = task.SubtaskCount
		}
	}

	// Create the subtasks
	subtasks, err := s.repo.CreateSubtasks(ctx, task.WorkspaceID, task.ID, subtaskCount)
	if err != nil {
		return fmt.Errorf("failed to create subtasks: %w", err)
	}

	task.ParallelSubtasks = true
	task.SubtaskCount = len(subtasks)
	task.Subtasks = subtasks

	// Update the task to indicate it has subtasks
	if err := s.repo.Update(ctx, task.WorkspaceID, task); err != nil {
		return fmt.Errorf("failed to update task with subtasks: %w", err)
	}

	// Trigger HTTP execution of each subtask asynchronously
	for _, subtask := range subtasks {
		// Use a closure to capture subtask
		go func(st *domain.Subtask) {
			s.triggerSubtaskExecution(st.ID)
		}(subtask)
	}

	return nil
}

// triggerSubtaskExecution makes an HTTP request to execute a subtask
func (s *TaskService) triggerSubtaskExecution(subtaskID string) {
	// Create request body
	requestBody, err := json.Marshal(domain.SubtaskRequest{
		SubtaskID: subtaskID,
	})
	if err != nil {
		s.logger.WithField("subtask_id", subtaskID).
			WithField("error", err.Error()).
			Error("Failed to marshal subtask request")
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.subtaskEndpointBaseURL+"/api/tasks.executeSubtask", bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.WithField("subtask_id", subtaskID).
			WithField("error", err.Error()).
			Error("Failed to create HTTP request for subtask execution")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{
		Timeout: 10 * time.Second, // Short timeout as this is just to initiate the execution
	}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.WithField("subtask_id", subtaskID).
			WithField("error", err.Error()).
			Error("Failed to send HTTP request for subtask execution")
		return
	}
	defer resp.Body.Close()

	// Log response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.WithField("subtask_id", subtaskID).
			WithField("status_code", resp.StatusCode).
			WithField("response", string(body)).
			Error("Non-OK response from subtask execution endpoint")
		return
	}

	s.logger.WithField("subtask_id", subtaskID).Info("Successfully triggered subtask execution")
}

// ExecuteTask executes a specific task
func (s *TaskService) ExecuteTask(ctx context.Context, workspace, taskID string) error {
	// First check if the context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Get the task
	task, err := s.repo.Get(ctx, workspace, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get the processor for this task type
	processor, err := s.GetProcessor(task.Type)
	if err != nil {
		return fmt.Errorf("failed to get processor: %w", err)
	}

	// Set timeout
	timeoutAt := time.Now().Add(time.Duration(task.MaxRuntime) * time.Second)

	// Mark task as running
	if err := s.repo.MarkAsRunning(ctx, workspace, taskID, timeoutAt); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Check if this task should use parallel processing with subtasks
	if task.ParallelSubtasks || (processor.SupportsParallelization() && task.State != nil && task.State.Progress > 0) {
		// If we already have subtasks, check their status
		if task.SubtaskCount > 0 {
			// Just update progress from existing subtasks
			if err := s.repo.UpdateTaskProgressFromSubtasks(ctx, workspace, taskID); err != nil {
				s.logger.WithField("task_id", taskID).
					WithField("workspace_id", workspace).
					WithField("error", err.Error()).
					Error("Failed to update task progress from subtasks")
			}

			return nil
		}

		// Check if we should use HTTP executor
		if task.UseHTTPExecutor {
			// Create subtasks and trigger their execution via HTTP
			if err := s.createSubtasksAndTriggerHTTP(ctx, task); err != nil {
				s.logger.WithField("task_id", taskID).
					WithField("workspace_id", workspace).
					WithField("error", err.Error()).
					Error("Failed to create subtasks and trigger HTTP execution")

				if err := s.repo.MarkAsFailed(ctx, workspace, taskID, fmt.Sprintf("Failed to create subtasks: %s", err.Error())); err != nil {
					s.logger.WithField("task_id", taskID).
						WithField("workspace_id", workspace).
						WithField("error", err.Error()).
						Error("Failed to mark task as failed")
				}

				return fmt.Errorf("failed to create subtasks: %w", err)
			}

			// Return after triggering subtasks - they'll run independently via HTTP
			return nil
		} else {
			// Use in-process subtask execution
			s.logger.WithField("task_id", taskID).
				WithField("workspace_id", workspace).
				Info("Creating subtasks for in-process execution")

			// Create the subtasks
			subtaskCount := 5 // Default value
			if task.SubtaskCount > 0 {
				subtaskCount = task.SubtaskCount
			}

			subtasks, err := s.repo.CreateSubtasks(ctx, workspace, taskID, subtaskCount)
			if err != nil {
				s.logger.WithField("task_id", taskID).
					WithField("workspace_id", workspace).
					WithField("error", err.Error()).
					Error("Failed to create subtasks")

				if err := s.repo.MarkAsFailed(ctx, workspace, taskID, fmt.Sprintf("Failed to create subtasks: %s", err.Error())); err != nil {
					s.logger.WithField("task_id", taskID).
						WithField("workspace_id", workspace).
						WithField("error", err.Error()).
						Error("Failed to mark task as failed")
				}

				return fmt.Errorf("failed to create subtasks: %w", err)
			}

			// Process each subtask directly
			for _, subtask := range subtasks {
				// Process in separate goroutine
				go func(st *domain.Subtask) {
					if err := s.ExecuteSubtask(context.Background(), st.ID); err != nil {
						s.logger.WithField("subtask_id", st.ID).
							WithField("task_id", taskID).
							WithField("error", err.Error()).
							Error("Failed to execute subtask")
					}
				}(subtask)
			}

			return nil
		}
	}

	// For non-parallel tasks, use the standard execution flow
	// Set up completion channel and context handling
	done := make(chan bool, 1)
	processErr := make(chan error, 1)

	// Process the task in a goroutine
	go func() {
		// Check if context was cancelled before we even start
		if ctx.Err() != nil {
			processErr <- ctx.Err()
			return
		}

		// Track task execution time
		startTime := time.Now()

		// Call the processor
		completed, err := processor.Process(ctx, task)

		// Calculate elapsed time
		elapsed := time.Since(startTime)

		if err != nil {
			s.logger.WithField("task_id", taskID).
				WithField("elapsed_time", elapsed).
				WithField("error", err.Error()).
				Error("Task processing failed")
			processErr <- err
			return
		}

		done <- completed
	}()

	// Wait for completion, error, or timeout
	select {
	case completed := <-done:
		if completed {
			// Task was completed successfully
			if err := s.repo.MarkAsCompleted(ctx, workspace, taskID); err != nil {
				return fmt.Errorf("failed to mark task as completed: %w", err)
			}
			s.logger.WithField("task_id", taskID).Info("Task completed successfully")
		} else {
			// Task needs more time, schedule it to run again
			nextRun := time.Now().Add(1 * time.Minute)
			if err := s.repo.MarkAsPaused(ctx, workspace, taskID, nextRun); err != nil {
				return fmt.Errorf("failed to mark task as paused: %w", err)
			}
			s.logger.WithField("task_id", taskID).Info("Task paused and will continue in next run")
		}
	case err := <-processErr:
		// Task failed with an error
		if err := s.repo.MarkAsFailed(ctx, workspace, taskID, err.Error()); err != nil {
			return fmt.Errorf("failed to mark task as failed: %w", err)
		}
		return fmt.Errorf("task processing error: %w", err)
	case <-ctx.Done():
		// Task timed out or context was cancelled
		if ctx.Err() == context.DeadlineExceeded {
			// This is a timeout
			errorMsg := fmt.Sprintf("Task execution timed out after %d seconds", task.MaxRuntime)
			s.logger.WithField("task_id", taskID).Warn(errorMsg)

			// Try to reschedule if retries are available
			if task.RetryCount < task.MaxRetries {
				nextRun := time.Now().Add(time.Duration(task.RetryInterval) * time.Second)
				if err := s.repo.MarkAsPaused(ctx, workspace, taskID, nextRun); err != nil {
					return fmt.Errorf("failed to mark task as paused after timeout: %w", err)
				}
				s.logger.WithField("task_id", taskID).
					WithField("retry_count", task.RetryCount+1).
					WithField("max_retries", task.MaxRetries).
					Warn("Task timed out and will retry")
			} else {
				// No retries left, mark as failed
				if err := s.repo.MarkAsFailed(ctx, workspace, taskID, errorMsg); err != nil {
					return fmt.Errorf("failed to mark task as failed after timeout: %w", err)
				}
				s.logger.WithField("task_id", taskID).
					WithField("max_retries", task.MaxRetries).
					Warn("Task timed out and has no retries left")
			}
		} else {
			// This is a cancellation
			if err := s.repo.MarkAsFailed(ctx, workspace, taskID, "Task execution was cancelled"); err != nil {
				return fmt.Errorf("failed to mark task as failed after cancellation: %w", err)
			}
			s.logger.WithField("task_id", taskID).Warn("Task execution was cancelled")
		}

		return ctx.Err()
	}

	return nil
}

// SaveTaskProgress updates the progress and state of a running task
func (s *TaskService) SaveTaskProgress(ctx context.Context, workspace, taskID string, progress float64, state *domain.TaskState) error {
	return s.repo.SaveState(ctx, workspace, taskID, progress, state)
}

// ExecuteSubtask executes a specific subtask
func (s *TaskService) ExecuteSubtask(ctx context.Context, subtaskID string) error {
	// Get the subtask details
	subtask, err := s.repo.GetSubtask(ctx, subtaskID)
	if err != nil {
		return fmt.Errorf("failed to get subtask: %w", err)
	}

	// Get the parent task
	parentTask, err := s.repo.Get(ctx, "", subtask.ParentTaskID)
	if err != nil {
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	// Get the appropriate processor
	processor, err := s.GetProcessor(parentTask.Type)
	if err != nil {
		return fmt.Errorf("failed to get processor: %w", err)
	}

	// Check if the processor supports parallelization
	if !processor.SupportsParallelization() {
		return fmt.Errorf("processor does not support parallelization")
	}

	// Mark subtask as running
	now := time.Now()
	subtask.Status = domain.SubtaskStatusRunning
	subtask.StartedAt = &now
	subtask.UpdatedAt = now

	// Process the subtask
	completed, err := processor.ProcessSubtask(ctx, subtask, parentTask)
	if err != nil {
		// Mark the subtask as failed
		if failErr := s.repo.FailSubtask(ctx, subtaskID, err.Error()); failErr != nil {
			s.logger.WithField("subtask_id", subtaskID).
				WithField("error", failErr.Error()).
				Error("Failed to mark subtask as failed")
		}
		return fmt.Errorf("failed to process subtask: %w", err)
	}

	// Mark subtask as completed or update progress
	if completed {
		if err := s.repo.CompleteSubtask(ctx, subtaskID); err != nil {
			return fmt.Errorf("failed to mark subtask as completed: %w", err)
		}
	} else {
		// Just update progress
		if err := s.repo.UpdateSubtaskProgress(ctx, subtaskID, subtask.Progress, subtask.State); err != nil {
			return fmt.Errorf("failed to update subtask progress: %w", err)
		}
	}

	// Update parent task progress
	if err := s.repo.UpdateTaskProgressFromSubtasks(ctx, parentTask.WorkspaceID, parentTask.ID); err != nil {
		s.logger.WithField("task_id", parentTask.ID).
			WithField("error", err.Error()).
			Error("Failed to update parent task progress")
	}

	return nil
}

// RegisterDefaultProcessors registers the default set of task processors
func (s *TaskService) RegisterDefaultProcessors(broadcastService *BroadcastService) {
	// Register send broadcast processor
	broadcastProcessor := NewSendBroadcastProcessor(broadcastService, s.logger)
	s.RegisterProcessor(broadcastProcessor)
}
