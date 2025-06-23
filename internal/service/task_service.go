package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/tracing"
)

// Maximum time a task can run before timing out
const defaultMaxTaskRuntime = 55 // 55 seconds

// TaskService manages task execution and state
type TaskService struct {
	repo        domain.TaskRepository
	logger      logger.Logger
	authService *AuthService
	processors  map[string]domain.TaskProcessor
	lock        sync.RWMutex
	apiEndpoint string
}

// WithTransaction executes a function within a transaction
func (s *TaskService) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	// The repository should handle the transaction
	repo, ok := s.repo.(interface {
		WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error
	})
	if !ok {
		return fmt.Errorf("repository does not support transactions")
	}

	return repo.WithTransaction(ctx, fn)
}

// NewTaskService creates a new task service instance
func NewTaskService(repository domain.TaskRepository, logger logger.Logger, authService *AuthService, apiEndpoint string) *TaskService {

	return &TaskService{
		repo:        repository,
		logger:      logger,
		authService: authService,
		processors:  make(map[string]domain.TaskProcessor),
		apiEndpoint: apiEndpoint,
	}
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
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "CreateTask")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", workspace)
	if task.BroadcastID != nil {
		tracing.AddAttribute(ctx, "broadcast_id", *task.BroadcastID)
	}
	tracing.AddAttribute(ctx, "task_type", task.Type)

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

	err := s.repo.Create(ctx, workspace, task)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
	}
	return err
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, workspace, id string) (*domain.Task, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "GetTask")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", workspace)
	tracing.AddAttribute(ctx, "task_id", id)

	task, err := s.repo.Get(ctx, workspace, id)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
	} else if task != nil && task.BroadcastID != nil {
		tracing.AddAttribute(ctx, "broadcast_id", *task.BroadcastID)
	}

	return task, err
}

// ListTasks lists tasks based on filter criteria
func (s *TaskService) ListTasks(ctx context.Context, workspace string, filter domain.TaskFilter) (*domain.TaskListResponse, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "ListTasks")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", workspace)
	tracing.AddAttribute(ctx, "limit", filter.Limit)
	tracing.AddAttribute(ctx, "offset", filter.Offset)

	// Removed Status and Type tracing attributes to fix compilation issues

	tasks, totalCount, err := s.repo.List(ctx, workspace, filter)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return nil, err
	}

	// Calculate if there are more results
	hasMore := (filter.Offset + len(tasks)) < totalCount
	tracing.AddAttribute(ctx, "total_count", totalCount)
	tracing.AddAttribute(ctx, "result_count", len(tasks))
	tracing.AddAttribute(ctx, "has_more", hasMore)

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
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "DeleteTask")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", workspace)
	tracing.AddAttribute(ctx, "task_id", id)

	err := s.repo.Delete(ctx, workspace, id)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
	}
	return err
}

// ExecutePendingTasks processes a batch of pending tasks
func (s *TaskService) ExecutePendingTasks(ctx context.Context, maxTasks int) error {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "ExecutePendingTasks")
	defer tracing.EndSpan(span, nil)

	// Get the next batch of tasks
	if maxTasks <= 0 {
		maxTasks = 10 // Default value
	}

	tracing.AddAttribute(ctx, "max_tasks", maxTasks)

	tasks, err := s.repo.GetNextBatch(ctx, maxTasks)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to get next batch of tasks: %w", err)
	}

	tracing.AddAttribute(ctx, "task_count", len(tasks))
	s.logger.WithField("task_count", len(tasks)).Info("Retrieved batch of tasks to process")

	if s.apiEndpoint == "" {
		tracing.AddAttribute(ctx, "execution_mode", "direct")
		s.logger.Warn("API endpoint not configured, falling back to direct execution")
		return s.executeTasksDirectly(ctx, tasks)
	}

	tracing.AddAttribute(ctx, "execution_mode", "http")
	// Execute tasks using HTTP roundtrips
	for _, task := range tasks {
		go func(t *domain.Task) {
			taskCtx, taskSpan := tracing.StartServiceSpan(ctx, "TaskService", "DispatchTaskExecution")
			defer tracing.EndSpan(taskSpan, nil)

			tracing.AddAttribute(taskCtx, "task_id", t.ID)
			tracing.AddAttribute(taskCtx, "workspace_id", t.WorkspaceID)
			tracing.AddAttribute(taskCtx, "task_type", t.Type)
			if t.BroadcastID != nil {
				tracing.AddAttribute(taskCtx, "broadcast_id", *t.BroadcastID)
			}

			s.logger.WithField("task_id", t.ID).
				WithField("workspace_id", t.WorkspaceID).
				Info("Dispatching task execution via HTTP")

			// Create request payload
			reqBody, err := json.Marshal(domain.ExecuteTaskRequest{
				WorkspaceID: t.WorkspaceID,
				ID:          t.ID,
			})
			if err != nil {
				tracing.MarkSpanError(taskCtx, err)
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("error", err.Error()).
					Error("Failed to marshal task execution request")
				return
			}

			// Create HTTP client with timeout
			httpClient := &http.Client{
				Timeout: 53 * time.Second, // 53 seconds timeout as requested
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // Skip TLS verification
					},
				},
			}

			// Wrap with OpenCensus tracing
			httpClient = tracing.WrapHTTPClient(httpClient)

			// Create request with tracing context
			endpoint := fmt.Sprintf("%s/api/tasks.execute", s.apiEndpoint)
			req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
			if err != nil {
				tracing.MarkSpanError(taskCtx, err)
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("error", err.Error()).
					Error("Failed to create HTTP request for task execution")
				return
			}

			// Set content type
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Task-ID", t.ID) // Add task ID for tracing

			// Execute request
			resp, err := httpClient.Do(req)
			if err != nil {
				tracing.MarkSpanError(taskCtx, err)
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("error", err.Error()).
					Error("HTTP request for task execution failed")
				return
			}
			defer resp.Body.Close()

			// Check response
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				err := fmt.Errorf("non-OK status: %d, response: %s", resp.StatusCode, string(body))
				tracing.MarkSpanError(taskCtx, err)
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("status_code", resp.StatusCode).
					WithField("response", string(body)).
					Error("HTTP request for task execution returned non-OK status")
				return
			}

			s.logger.WithField("task_id", t.ID).
				WithField("workspace_id", t.WorkspaceID).
				Info("Task execution request dispatched successfully")
		}(task)
	}

	return nil
}

// executeTasksDirectly processes tasks directly without HTTP roundtrips
// This is used as a fallback when API endpoint is not configured
func (s *TaskService) executeTasksDirectly(ctx context.Context, tasks []*domain.Task) error {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "executeTasksDirectly")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "task_count", len(tasks))
	bgCtx := context.Background()

	for _, task := range tasks {
		// Create a separate context for each task with a timeout
		taskCtxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(task.MaxRuntime)*time.Second)

		// Handle the task in a goroutine
		go func(t *domain.Task, ctxWithTimeout context.Context, cancelFunc context.CancelFunc) {
			execCtxWithTimeout, execSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "executeTaskDirectly")

			// Set task attributes
			tracing.AddAttribute(execCtxWithTimeout, "task_id", t.ID)
			tracing.AddAttribute(execCtxWithTimeout, "workspace_id", t.WorkspaceID)
			tracing.AddAttribute(execCtxWithTimeout, "task_type", t.Type)
			if t.BroadcastID != nil {
				tracing.AddAttribute(execCtxWithTimeout, "broadcast_id", *t.BroadcastID)
			}

			// Ensure we clean up the context and handle timeout
			defer func() {
				// Check if the context expired (timed out)
				if ctxWithTimeout.Err() == context.DeadlineExceeded {
					timeoutError := fmt.Errorf("task execution timed out after %d seconds", t.MaxRuntime)
					tracing.MarkSpanError(execCtxWithTimeout, timeoutError)
					s.logger.WithField("task_id", t.ID).
						WithField("workspace_id", t.WorkspaceID).
						WithField("error", timeoutError.Error()).
						Warn("Task timed out")

					// Mark the task as failed due to timeout
					if err := s.repo.MarkAsFailed(bgCtx, t.WorkspaceID, t.ID, timeoutError.Error()); err != nil {
						tracing.MarkSpanError(execCtxWithTimeout, err)
						s.logger.WithField("task_id", t.ID).
							WithField("workspace_id", t.WorkspaceID).
							WithField("error", err.Error()).
							Error("Failed to mark timed out task as failed")
					}
				}
				tracing.EndSpan(execSpan, ctxWithTimeout.Err())
				cancelFunc()
			}()

			if err := s.ExecuteTask(execCtxWithTimeout, t.WorkspaceID, t.ID); err != nil {
				tracing.MarkSpanError(execCtxWithTimeout, err)
				s.logger.WithField("task_id", t.ID).
					WithField("workspace_id", t.WorkspaceID).
					WithField("error", err.Error()).
					Error("Failed to execute task")
			}
		}(task, taskCtxWithTimeout, cancel)
	}

	return nil
}

// ExecuteTask executes a specific task
func (s *TaskService) ExecuteTask(ctxWithTimeout context.Context, workspace, taskID string) error {
	ctxWithTimeout, span := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "ExecuteTask")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctxWithTimeout, "workspace_id", workspace)
	tracing.AddAttribute(ctxWithTimeout, "task_id", taskID)

	// First check if the context is already cancelled
	if ctxWithTimeout.Err() != nil {
		tracing.MarkSpanError(ctxWithTimeout, ctxWithTimeout.Err())
		return ctxWithTimeout.Err()
	}

	// Get the task
	var task *domain.Task
	var processor domain.TaskProcessor
	var err error

	// Wrap the initial setup operations in a transaction
	err = s.WithTransaction(ctxWithTimeout, func(tx *sql.Tx) error {
		txCtx, txSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "ExecuteTaskTransaction")
		defer tracing.EndSpan(txSpan, nil)

		var taskErr error
		task, taskErr = s.repo.GetTx(txCtx, tx, workspace, taskID)
		if taskErr != nil {
			tracing.MarkSpanError(txCtx, taskErr)
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"error":        taskErr.Error(),
			}).Error("Failed to get task for execution")
			return &domain.ErrNotFound{
				Entity: "task",
				ID:     taskID,
			}
		}

		if task != nil {
			tracing.AddAttribute(txCtx, "task_type", task.Type)
			if task.BroadcastID != nil {
				tracing.AddAttribute(txCtx, "broadcast_id", *task.BroadcastID)
			}
		}

		// Get the processor for this task type
		var procErr error
		processor, procErr = s.GetProcessor(task.Type)
		if procErr != nil {
			tracing.MarkSpanError(txCtx, procErr)
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"task_type":    task.Type,
				"error":        procErr.Error(),
			}).Error("Failed to get processor for task type")
			return &domain.ErrTaskExecution{
				TaskID: taskID,
				Reason: "no processor registered for task type",
				Err:    procErr,
			}
		}

		// Set timeout
		timeoutAt := time.Now().Add(time.Duration(task.MaxRuntime) * time.Second)
		tracing.AddAttribute(txCtx, "timeout_at", timeoutAt.Format(time.RFC3339))

		// Mark task as running within the same transaction
		if markErr := s.repo.MarkAsRunningTx(txCtx, tx, workspace, taskID, timeoutAt); markErr != nil {
			tracing.MarkSpanError(txCtx, markErr)
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"error":        markErr.Error(),
			}).Error("Failed to mark task as running")
			return &domain.ErrTaskExecution{
				TaskID: taskID,
				Reason: "failed to mark task as running",
				Err:    markErr,
			}
		}

		return nil
	})

	if err != nil {
		tracing.MarkSpanError(ctxWithTimeout, err)
		return err
	}

	// For non-parallel tasks, use the standard execution flow
	// Set up completion channel and context handling
	done := make(chan bool, 1)
	processErr := make(chan error, 1)
	bgCtx := context.Background()

	// Process the task in a goroutine
	go func() {
		procCtxWithTimeout, procSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "ProcessTask")
		defer tracing.EndSpan(procSpan, nil)

		tracing.AddAttribute(procCtxWithTimeout, "task_id", taskID)
		tracing.AddAttribute(procCtxWithTimeout, "workspace_id", workspace)
		tracing.AddAttribute(procCtxWithTimeout, "task_type", task.Type)
		if task.BroadcastID != nil {
			tracing.AddAttribute(procCtxWithTimeout, "broadcast_id", *task.BroadcastID)
		}

		// Check if context was cancelled before we even start
		if procCtxWithTimeout.Err() != nil {
			tracing.MarkSpanError(procCtxWithTimeout, procCtxWithTimeout.Err())
			processErr <- procCtxWithTimeout.Err()
			return
		}

		// Track task execution time
		startTime := time.Now()

		// Call the processor
		completed, err := processor.Process(procCtxWithTimeout, task)

		// Calculate elapsed time
		elapsed := time.Since(startTime)
		tracing.AddAttribute(procCtxWithTimeout, "elapsed_time_ms", elapsed.Milliseconds())
		tracing.AddAttribute(procCtxWithTimeout, "task_completed", completed)

		if err != nil {
			tracing.MarkSpanError(procCtxWithTimeout, err)
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"elapsed_time": elapsed,
				"error":        err.Error(),
			}).Error("Task processing failed")
			processErr <- &domain.ErrTaskExecution{
				TaskID: taskID,
				Reason: "processing failed",
				Err:    err,
			}
			return
		}

		done <- completed
	}()

	// Wait for completion, error, or timeout
	select {
	case completed := <-done:
		if completed {
			// Task was completed successfully
			completeCtxWithTimeout, completeSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "MarkTaskCompleted")
			defer tracing.EndSpan(completeSpan, nil)

			tracing.AddAttribute(completeCtxWithTimeout, "task_id", taskID)
			tracing.AddAttribute(completeCtxWithTimeout, "workspace_id", workspace)

			if err := s.repo.MarkAsCompleted(bgCtx, workspace, taskID); err != nil {
				tracing.MarkSpanError(completeCtxWithTimeout, err)
				s.logger.WithFields(map[string]interface{}{
					"task_id":      taskID,
					"workspace_id": workspace,
					"error":        err.Error(),
				}).Error("Failed to mark task as completed")
				return &domain.ErrTaskExecution{
					TaskID: taskID,
					Reason: "failed to mark task as completed",
					Err:    err,
				}
			}
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
			}).Info("Task completed successfully")
		} else {
			// Mark task as paused for next run
			pauseCtxWithTimeout, pauseSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "MarkTaskPaused")
			defer tracing.EndSpan(pauseSpan, nil)

			tracing.AddAttribute(pauseCtxWithTimeout, "task_id", taskID)
			tracing.AddAttribute(pauseCtxWithTimeout, "workspace_id", workspace)

			nextRun := time.Now()
			tracing.AddAttribute(pauseCtxWithTimeout, "next_run", nextRun.Format(time.RFC3339))
			tracing.AddAttribute(pauseCtxWithTimeout, "progress", task.Progress)

			if err := s.repo.MarkAsPaused(bgCtx, task.WorkspaceID, task.ID, nextRun, task.Progress, task.State); err != nil {
				tracing.MarkSpanError(pauseCtxWithTimeout, err)
				s.logger.WithFields(map[string]interface{}{
					"task_id":      taskID,
					"workspace_id": workspace,
					"error":        err.Error(),
				}).Error("Failed to mark task as paused")
				return &domain.ErrTaskExecution{
					TaskID: taskID,
					Reason: "failed to mark task as paused",
					Err:    err,
				}
			}
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"next_run":     nextRun,
			}).Info("Task paused and will continue in next run")
		}
	case err := <-processErr:
		// Task failed with an error
		failCtxWithTimeout, failSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "MarkTaskFailed")
		defer tracing.EndSpan(failSpan, nil)

		tracing.AddAttribute(failCtxWithTimeout, "task_id", taskID)
		tracing.AddAttribute(failCtxWithTimeout, "workspace_id", workspace)
		tracing.AddAttribute(failCtxWithTimeout, "error", err.Error())

		if markErr := s.repo.MarkAsFailed(bgCtx, workspace, taskID, err.Error()); markErr != nil {
			tracing.MarkSpanError(failCtxWithTimeout, markErr)
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"error":        markErr.Error(),
				"process_err":  err.Error(),
			}).Error("Failed to mark task as failed")
			return fmt.Errorf("failed to mark task as failed: %w", markErr)
		}
		return err
	case <-ctxWithTimeout.Done():
		// Task timed out or context was cancelled
		timeoutCtx, timeoutSpan := tracing.StartServiceSpan(ctxWithTimeout, "TaskService", "HandleTaskTimeout")

		tracing.AddAttribute(timeoutCtx, "task_id", taskID)
		tracing.AddAttribute(timeoutCtx, "workspace_id", workspace)
		tracing.AddAttribute(timeoutCtx, "context_error", ctxWithTimeout.Err().Error())

		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			// This is a timeout
			timeoutErr := &domain.ErrTaskTimeout{
				TaskID:     taskID,
				MaxRuntime: task.MaxRuntime,
			}
			tracing.MarkSpanError(timeoutCtx, timeoutErr)

			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
				"max_runtime":  task.MaxRuntime,
			}).Warn(timeoutErr.Error())

			// Try to reschedule if retries are available
			if task.RetryCount < task.MaxRetries {
				// MarkAsPaused now includes state and progress parameters
				nextRun := time.Now().Add(time.Duration(task.RetryInterval) * time.Second)
				tracing.AddAttribute(timeoutCtx, "retry_count", task.RetryCount+1)
				tracing.AddAttribute(timeoutCtx, "max_retries", task.MaxRetries)
				tracing.AddAttribute(timeoutCtx, "next_run", nextRun.Format(time.RFC3339))

				if err := s.repo.MarkAsPaused(bgCtx, task.WorkspaceID, task.ID, nextRun, task.Progress, task.State); err != nil {
					tracing.MarkSpanError(timeoutCtx, err)
					s.logger.WithFields(map[string]interface{}{
						"task_id":      taskID,
						"workspace_id": workspace,
						"error":        err.Error(),
					}).Error("Failed to mark task as paused after timeout")
					tracing.EndSpan(timeoutSpan, err)
					return &domain.ErrTaskExecution{
						TaskID: taskID,
						Reason: "failed to mark task as paused after timeout",
						Err:    err,
					}
				}
				s.logger.WithFields(map[string]interface{}{
					"task_id":        taskID,
					"workspace_id":   workspace,
					"retry_count":    task.RetryCount + 1,
					"max_retries":    task.MaxRetries,
					"next_run":       nextRun,
					"retry_interval": task.RetryInterval,
				}).Warn("Task timed out and will retry")
			} else {
				// No retries left, mark as failed
				tracing.AddAttribute(timeoutCtx, "retries_exhausted", true)

				if err := s.repo.MarkAsFailed(bgCtx, workspace, taskID, timeoutErr.Error()); err != nil {
					tracing.MarkSpanError(timeoutCtx, err)
					s.logger.WithFields(map[string]interface{}{
						"task_id":      taskID,
						"workspace_id": workspace,
						"error":        err.Error(),
					}).Error("Failed to mark task as failed after timeout")
					tracing.EndSpan(timeoutSpan, err)
					return &domain.ErrTaskExecution{
						TaskID: taskID,
						Reason: "failed to mark task as failed after timeout",
						Err:    err,
					}
				}
				s.logger.WithFields(map[string]interface{}{
					"task_id":      taskID,
					"workspace_id": workspace,
					"max_retries":  task.MaxRetries,
				}).Warn("Task timed out and has no retries left")
			}

			tracing.EndSpan(timeoutSpan, timeoutErr)
			return timeoutErr
		} else {
			// This is a cancellation
			cancelErr := &domain.ErrTaskExecution{
				TaskID: taskID,
				Reason: "task execution was cancelled",
			}
			tracing.MarkSpanError(timeoutCtx, cancelErr)

			if err := s.repo.MarkAsFailed(bgCtx, workspace, taskID, cancelErr.Error()); err != nil {
				tracing.MarkSpanError(timeoutCtx, err)
				s.logger.WithFields(map[string]interface{}{
					"task_id":      taskID,
					"workspace_id": workspace,
					"error":        err.Error(),
				}).Error("Failed to mark task as failed after cancellation")
				tracing.EndSpan(timeoutSpan, err)
				return &domain.ErrTaskExecution{
					TaskID: taskID,
					Reason: "failed to mark task as failed after cancellation",
					Err:    err,
				}
			}
			s.logger.WithFields(map[string]interface{}{
				"task_id":      taskID,
				"workspace_id": workspace,
			}).Warn("Task execution was cancelled")

			tracing.EndSpan(timeoutSpan, cancelErr)
			return cancelErr
		}
	}

	return nil
}

// SubscribeToBroadcastEvents registers handlers for broadcast-related events
func (s *TaskService) SubscribeToBroadcastEvents(eventBus domain.EventBus) {
	// Subscribe to broadcast events
	eventBus.Subscribe(domain.EventBroadcastScheduled, s.handleBroadcastScheduled)
	eventBus.Subscribe(domain.EventBroadcastPaused, s.handleBroadcastPaused)
	eventBus.Subscribe(domain.EventBroadcastResumed, s.handleBroadcastResumed)
	eventBus.Subscribe(domain.EventBroadcastSent, s.handleBroadcastSent)
	eventBus.Subscribe(domain.EventBroadcastFailed, s.handleBroadcastFailed)
	eventBus.Subscribe(domain.EventBroadcastCancelled, s.handleBroadcastCancelled)

	s.logger.Info("TaskService subscribed to broadcast events")
}

// Event handlers for broadcast events
func (s *TaskService) handleBroadcastScheduled(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastScheduled")
	defer tracing.EndSpan(span, nil)

	broadcastID := payload.EntityID
	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)
	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast scheduled event")

	// Use a transaction for checking and potentially updating/creating task
	err := s.WithTransaction(ctx, func(tx *sql.Tx) error {
		txCtx, txSpan := tracing.StartServiceSpan(ctx, "TaskService", "BroadcastScheduledTransaction")
		defer tracing.EndSpan(txSpan, nil)

		// Try to find the task for this broadcast ID directly
		existingTask, err := s.repo.GetTaskByBroadcastID(txCtx, payload.WorkspaceID, broadcastID)
		if err != nil {
			// If no task exists, we'll create one later
			tracing.AddAttribute(txCtx, "task_exists", false)
			s.logger.WithField("broadcast_id", broadcastID).
				WithField("error", err.Error()).
				Debug("No existing task found for broadcast, will create new one")
		}

		// Update existing task if found
		if existingTask != nil {
			tracing.AddAttribute(txCtx, "task_exists", true)
			tracing.AddAttribute(txCtx, "task_id", existingTask.ID)
			tracing.AddAttribute(txCtx, "current_status", string(existingTask.Status))

			s.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcastID,
				"task_id":      existingTask.ID,
			}).Info("Task already exists for broadcast, updating status")

			// Update task state if needed
			sendNow, _ := payload.Data["send_now"].(bool)
			status, _ := payload.Data["status"].(string)

			tracing.AddAttribute(txCtx, "send_now", sendNow)
			tracing.AddAttribute(txCtx, "broadcast_status", status)

			if sendNow && status == string(domain.BroadcastStatusSending) {
				// If broadcast is being sent immediately, mark task as pending and set next run to now
				nextRunAfter := time.Now()
				existingTask.NextRunAfter = &nextRunAfter
				existingTask.Status = domain.TaskStatusPending

				// Ensure BroadcastID is set
				if existingTask.BroadcastID == nil {
					broadcastIDCopy := broadcastID
					existingTask.BroadcastID = &broadcastIDCopy
				}

				if updateErr := s.repo.Update(txCtx, payload.WorkspaceID, existingTask); updateErr != nil {
					tracing.MarkSpanError(txCtx, updateErr)
					s.logger.WithFields(map[string]interface{}{
						"broadcast_id": broadcastID,
						"task_id":      existingTask.ID,
						"error":        updateErr.Error(),
					}).Error("Failed to update task for scheduled broadcast")
					return updateErr
				}
			}

			return nil
		}

		// If no task exists, create one
		s.logger.WithField("broadcast_id", broadcastID).Info("Creating new task for scheduled broadcast")

		// Set up task state
		sendNow, _ := payload.Data["send_now"].(bool)
		status, _ := payload.Data["status"].(string)

		tracing.AddAttribute(txCtx, "send_now", sendNow)
		tracing.AddAttribute(txCtx, "broadcast_status", status)
		tracing.AddAttribute(txCtx, "creating_new_task", true)

		// Create a copy of the broadcast ID for the pointer
		broadcastIDCopy := broadcastID

		task := &domain.Task{
			WorkspaceID: payload.WorkspaceID,
			Type:        "send_broadcast",
			Status:      domain.TaskStatusPending,
			BroadcastID: &broadcastIDCopy,
			State: &domain.TaskState{
				Progress: 0,
				Message:  "Starting broadcast",
				SendBroadcast: &domain.SendBroadcastState{
					BroadcastID:     broadcastID,
					ChannelType:     "email",
					SentCount:       0,
					FailedCount:     0,
					RecipientOffset: 0,
				},
			},
			MaxRuntime:    600, // 10 minutes
			MaxRetries:    3,
			RetryInterval: 300, // 5 minutes
		}

		// If the broadcast is set to send immediately, we don't need to set NextRunAfter
		// If it's scheduled for the future, we should set NextRunAfter based on the schedule
		if !sendNow && status == string(domain.BroadcastStatusScheduled) {
			// Schedule the task to run in the future
			nextRunAfter := time.Now().Add(1 * time.Minute)
			task.NextRunAfter = &nextRunAfter
			tracing.AddAttribute(txCtx, "next_run_after", nextRunAfter.Format(time.RFC3339))
		}

		if createErr := s.CreateTask(txCtx, payload.WorkspaceID, task); createErr != nil {
			tracing.MarkSpanError(txCtx, createErr)
			s.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcastID,
				"error":        createErr.Error(),
			}).Error("Failed to create task for scheduled broadcast")
			return createErr
		}

		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": payload.WorkspaceID,
		}).Info("Successfully created task for scheduled broadcast")

		return nil
	})

	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": payload.WorkspaceID,
			"error":        err.Error(),
		}).Error("Failed to handle broadcast scheduled event")
	}
}

func (s *TaskService) handleBroadcastPaused(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastPaused")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		err := fmt.Errorf("missing or invalid broadcast_id")
		tracing.MarkSpanError(ctx, err)
		s.logger.Error("Failed to handle broadcast paused event: missing or invalid broadcast_id")
		return
	}

	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast paused event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithField("error", err.Error()).Debug("No task found for paused broadcast")
		return
	}

	tracing.AddAttribute(ctx, "task_id", task.ID)
	tracing.AddAttribute(ctx, "current_status", string(task.Status))

	// Pause the task
	nextRunAfter := time.Now().Add(24 * time.Hour) // Pause for 24 hours
	tracing.AddAttribute(ctx, "next_run_after", nextRunAfter.Format(time.RFC3339))

	if err := s.repo.MarkAsPaused(ctx, payload.WorkspaceID, task.ID, nextRunAfter, task.Progress, task.State); err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
			"error":        err.Error(),
		}).Error("Failed to pause task for paused broadcast")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
		}).Info("Successfully paused task for paused broadcast")
	}
}

func (s *TaskService) handleBroadcastResumed(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastResumed")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		err := fmt.Errorf("missing or invalid broadcast_id")
		tracing.MarkSpanError(ctx, err)
		s.logger.Error("Failed to handle broadcast resumed event: missing or invalid broadcast_id")
		return
	}

	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast resumed event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithField("error", err.Error()).Debug("No task found for resumed broadcast")
		return
	}

	tracing.AddAttribute(ctx, "task_id", task.ID)
	tracing.AddAttribute(ctx, "current_status", string(task.Status))

	// Resume the task
	nextRunAfter := time.Now()
	task.NextRunAfter = &nextRunAfter
	task.Status = domain.TaskStatusPending

	tracing.AddAttribute(ctx, "next_run_after", nextRunAfter.Format(time.RFC3339))
	tracing.AddAttribute(ctx, "new_status", string(task.Status))

	if err := s.repo.Update(ctx, payload.WorkspaceID, task); err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
			"error":        err.Error(),
		}).Error("Failed to resume task for resumed broadcast")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
		}).Info("Successfully resumed task for resumed broadcast")
	}
}

func (s *TaskService) handleBroadcastSent(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastSent")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		err := fmt.Errorf("missing or invalid broadcast_id")
		tracing.MarkSpanError(ctx, err)
		s.logger.Error("Failed to handle broadcast sent event: missing or invalid broadcast_id")
		return
	}

	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast sent event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithField("error", err.Error()).Debug("No task found for sent broadcast")
		return
	}

	tracing.AddAttribute(ctx, "task_id", task.ID)
	tracing.AddAttribute(ctx, "current_status", string(task.Status))

	// Mark the task as completed
	if err := s.repo.MarkAsCompleted(ctx, payload.WorkspaceID, task.ID); err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
			"error":        err.Error(),
		}).Error("Failed to complete task for sent broadcast")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
		}).Info("Successfully completed task for sent broadcast")
	}
}

func (s *TaskService) handleBroadcastFailed(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastFailed")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		err := fmt.Errorf("missing or invalid broadcast_id")
		tracing.MarkSpanError(ctx, err)
		s.logger.Error("Failed to handle broadcast failed event: missing or invalid broadcast_id")
		return
	}

	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)

	reason, _ := payload.Data["reason"].(string)
	if reason == "" {
		reason = "Broadcast failed"
	}

	tracing.AddAttribute(ctx, "failure_reason", reason)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
		"reason":       reason,
	}).Info("Handling broadcast failed event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithField("error", err.Error()).Debug("No task found for failed broadcast")
		return
	}

	tracing.AddAttribute(ctx, "task_id", task.ID)
	tracing.AddAttribute(ctx, "current_status", string(task.Status))

	// Mark the task as failed
	if err := s.repo.MarkAsFailed(ctx, payload.WorkspaceID, task.ID, reason); err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
			"error":        err.Error(),
		}).Error("Failed to mark task as failed for failed broadcast")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
		}).Info("Successfully marked task as failed for failed broadcast")
	}
}

func (s *TaskService) handleBroadcastCancelled(ctx context.Context, payload domain.EventPayload) {
	ctx, span := tracing.StartServiceSpan(ctx, "TaskService", "handleBroadcastCancelled")
	defer tracing.EndSpan(span, nil)

	tracing.AddAttribute(ctx, "workspace_id", payload.WorkspaceID)
	tracing.AddAttribute(ctx, "event_type", payload.Type)

	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		err := fmt.Errorf("missing or invalid broadcast_id")
		tracing.MarkSpanError(ctx, err)
		s.logger.Error("Failed to handle broadcast cancelled event: missing or invalid broadcast_id")
		return
	}

	tracing.AddAttribute(ctx, "broadcast_id", broadcastID)

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast cancelled event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithField("error", err.Error()).Debug("No task found for cancelled broadcast")
		return
	}

	tracing.AddAttribute(ctx, "task_id", task.ID)
	tracing.AddAttribute(ctx, "current_status", string(task.Status))

	// Mark the task as failed with cancellation reason
	cancelReason := "Broadcast was cancelled"
	tracing.AddAttribute(ctx, "cancel_reason", cancelReason)

	if err := s.repo.MarkAsFailed(ctx, payload.WorkspaceID, task.ID, cancelReason); err != nil {
		tracing.MarkSpanError(ctx, err)
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
			"error":        err.Error(),
		}).Error("Failed to mark task as failed for cancelled broadcast")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"task_id":      task.ID,
		}).Info("Successfully marked task as failed for cancelled broadcast")
	}
}
