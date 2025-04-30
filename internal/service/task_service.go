package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
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

// ExecuteTask executes a specific task
func (s *TaskService) ExecuteTask(ctx context.Context, workspace, taskID string) error {
	// First check if the context is already cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Get the task
	var task *domain.Task
	var processor domain.TaskProcessor
	var err error

	// Wrap the initial setup operations in a transaction
	err = s.WithTransaction(ctx, func(tx *sql.Tx) error {

		var taskErr error
		task, taskErr = s.repo.GetTx(ctx, tx, workspace, taskID)
		if taskErr != nil {
			return fmt.Errorf("failed to get task: %w", taskErr)
		}

		// Get the processor for this task type
		var procErr error
		processor, procErr = s.GetProcessor(task.Type)
		if procErr != nil {
			return fmt.Errorf("failed to get processor: %w", procErr)
		}

		// Set timeout
		timeoutAt := time.Now().Add(time.Duration(task.MaxRuntime) * time.Second)

		// Mark task as running within the same transaction
		return s.repo.MarkAsRunningTx(ctx, tx, workspace, taskID, timeoutAt)
	})

	if err != nil {
		return fmt.Errorf("failed to prepare task execution: %w", err)
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

// RegisterDefaultProcessors registers the default set of task processors
func (s *TaskService) RegisterDefaultProcessors(broadcastService domain.BroadcastSender) {
	// Register send broadcast processor
	broadcastProcessor := NewSendBroadcastProcessor(broadcastService, s.logger)
	s.RegisterProcessor(broadcastProcessor)
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
	broadcastID := payload.EntityID

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast scheduled event")

	// Use a transaction for checking and potentially updating/creating task
	err := s.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Try to find the task for this broadcast ID directly
		existingTask, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
		if err != nil {
			// If no task exists, we'll create one later
			s.logger.WithField("broadcast_id", broadcastID).
				WithField("error", err.Error()).
				Debug("No existing task found for broadcast, will create new one")
		}

		// Update existing task if found
		if existingTask != nil {
			s.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcastID,
				"task_id":      existingTask.ID,
			}).Info("Task already exists for broadcast, updating status")

			// Update task state if needed
			sendNow, _ := payload.Data["send_now"].(bool)
			status, _ := payload.Data["status"].(string)

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

				if updateErr := s.repo.Update(ctx, payload.WorkspaceID, existingTask); updateErr != nil {
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
		}

		if createErr := s.CreateTask(ctx, payload.WorkspaceID, task); createErr != nil {
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
		s.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": payload.WorkspaceID,
			"error":        err.Error(),
		}).Error("Failed to handle broadcast scheduled event")
	}
}

func (s *TaskService) handleBroadcastPaused(ctx context.Context, payload domain.EventPayload) {
	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		s.logger.Error("Failed to handle broadcast paused event: missing or invalid broadcast_id")
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast paused event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		s.logger.WithField("error", err.Error()).Debug("No task found for paused broadcast")
		return
	}

	// Pause the task
	nextRunAfter := time.Now().Add(24 * time.Hour) // Pause for 24 hours
	if err := s.repo.MarkAsPaused(ctx, payload.WorkspaceID, task.ID, nextRunAfter); err != nil {
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
	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		s.logger.Error("Failed to handle broadcast resumed event: missing or invalid broadcast_id")
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast resumed event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		s.logger.WithField("error", err.Error()).Debug("No task found for resumed broadcast")
		return
	}

	// Resume the task
	nextRunAfter := time.Now()
	task.NextRunAfter = &nextRunAfter
	task.Status = domain.TaskStatusPending
	if err := s.repo.Update(ctx, payload.WorkspaceID, task); err != nil {
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
	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		s.logger.Error("Failed to handle broadcast sent event: missing or invalid broadcast_id")
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast sent event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		s.logger.WithField("error", err.Error()).Debug("No task found for sent broadcast")
		return
	}

	// Mark the task as completed
	if err := s.repo.MarkAsCompleted(ctx, payload.WorkspaceID, task.ID); err != nil {
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
	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		s.logger.Error("Failed to handle broadcast failed event: missing or invalid broadcast_id")
		return
	}

	reason, _ := payload.Data["reason"].(string)
	if reason == "" {
		reason = "Broadcast failed"
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
		"reason":       reason,
	}).Info("Handling broadcast failed event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		s.logger.WithField("error", err.Error()).Debug("No task found for failed broadcast")
		return
	}

	// Mark the task as failed
	if err := s.repo.MarkAsFailed(ctx, payload.WorkspaceID, task.ID, reason); err != nil {
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
	broadcastID, ok := payload.Data["broadcast_id"].(string)
	if !ok || broadcastID == "" {
		s.logger.Error("Failed to handle broadcast cancelled event: missing or invalid broadcast_id")
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"workspace_id": payload.WorkspaceID,
	}).Info("Handling broadcast cancelled event")

	// Find associated task by broadcast ID
	task, err := s.repo.GetTaskByBroadcastID(ctx, payload.WorkspaceID, broadcastID)
	if err != nil {
		s.logger.WithField("error", err.Error()).Debug("No task found for cancelled broadcast")
		return
	}

	// Mark the task as failed with cancellation reason
	if err := s.repo.MarkAsFailed(ctx, payload.WorkspaceID, task.ID, "Broadcast was cancelled"); err != nil {
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
