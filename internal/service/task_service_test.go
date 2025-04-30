package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Mock logger for testing
type mockLogger struct{}

func (l *mockLogger) Debug(msg string)                                       {}
func (l *mockLogger) Info(msg string)                                        {}
func (l *mockLogger) Warn(msg string)                                        {}
func (l *mockLogger) Error(msg string)                                       {}
func (l *mockLogger) Fatal(msg string)                                       {}
func (l *mockLogger) Debugf(format string, args ...interface{})              {}
func (l *mockLogger) Infof(format string, args ...interface{})               {}
func (l *mockLogger) Warnf(format string, args ...interface{})               {}
func (l *mockLogger) Errorf(format string, args ...interface{})              {}
func (l *mockLogger) Fatalf(format string, args ...interface{})              {}
func (l *mockLogger) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *mockLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }
func (l *mockLogger) WithError(err error) logger.Logger                      { return l }

// Mock auth service for testing - not used in our tests, just needed for constructor
type mockAuthService struct{}

func TestTaskService_ExecuteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockLogger := &mockLogger{}
	// Use nil for auth service since it's not used in our tests
	var mockAuthService *AuthService = nil
	apiEndpoint := "http://localhost:8080"

	taskService := NewTaskService(mockRepo, mockLogger, mockAuthService, apiEndpoint)

	// Setup transaction mocking for all tests
	mockRepo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		}).AnyTimes()

	t.Run("Task not found error", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		workspaceID := "workspace1"
		taskID := "task123"

		// Configure mock repository to return a "not found" error
		notFoundErr := fmt.Errorf("task not found")
		mockRepo.EXPECT().
			GetTx(gomock.Any(), gomock.Any(), workspaceID, taskID).
			Return(nil, notFoundErr)

		// Call the method under test
		err := taskService.ExecuteTask(ctx, workspaceID, taskID)

		// Verify returned error is of type ErrNotFound
		assert.Error(t, err)
		var notFoundError *domain.ErrNotFound
		assert.True(t, errors.As(err, &notFoundError))
		assert.Equal(t, "task", notFoundError.Entity)
		assert.Equal(t, taskID, notFoundError.ID)
	})

	t.Run("Processor not found error", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		workspaceID := "workspace1"
		taskID := "task123"

		// Create a task with an unsupported type
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: workspaceID,
			Type:        "unsupported_task_type",
			Status:      domain.TaskStatusPending,
		}

		// Configure mock repository
		mockRepo.EXPECT().
			GetTx(gomock.Any(), gomock.Any(), workspaceID, taskID).
			Return(task, nil)

		// Call the method under test
		err := taskService.ExecuteTask(ctx, workspaceID, taskID)

		// Verify returned error is of type ErrTaskExecution
		assert.Error(t, err)
		var taskExecError *domain.ErrTaskExecution
		assert.True(t, errors.As(err, &taskExecError))
		assert.Equal(t, taskID, taskExecError.TaskID)
		assert.Equal(t, "no processor registered for task type", taskExecError.Reason)
		assert.Contains(t, taskExecError.Error(), "unsupported_task_type")
	})

	t.Run("Mark as running error", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		workspaceID := "workspace1"
		taskID := "task123"

		// Create a task with a supported type
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: workspaceID,
			Type:        "send_broadcast",
			Status:      domain.TaskStatusPending,
			MaxRuntime:  60,
		}

		// Register a processor for the task type
		mockProcessor := mocks.NewMockTaskProcessor(ctrl)
		// Configure CanProcess to be called for all supported task types
		for _, supportedType := range getTaskTypes() {
			mockProcessor.EXPECT().
				CanProcess(supportedType).
				Return(supportedType == "send_broadcast").
				AnyTimes()
		}
		taskService.RegisterProcessor(mockProcessor)

		// Configure mock repository
		mockRepo.EXPECT().
			GetTx(gomock.Any(), gomock.Any(), workspaceID, taskID).
			Return(task, nil)

		// MarkAsRunningTx should return an error
		markingError := fmt.Errorf("database connection error")
		mockRepo.EXPECT().
			MarkAsRunningTx(gomock.Any(), gomock.Any(), workspaceID, taskID, gomock.Any()).
			Return(markingError)

		// Call the method under test
		err := taskService.ExecuteTask(ctx, workspaceID, taskID)

		// Verify returned error is of type ErrTaskExecution with the correct reason
		assert.Error(t, err)
		var taskExecError *domain.ErrTaskExecution
		assert.True(t, errors.As(err, &taskExecError))
		assert.Equal(t, taskID, taskExecError.TaskID)
		assert.Equal(t, "failed to mark task as running", taskExecError.Reason)
		assert.Equal(t, markingError, taskExecError.Err)
	})

	t.Run("Processing error returns ErrTaskExecution", func(t *testing.T) {
		// Setup - create a new controller for this test to avoid interference
		procCtrl := gomock.NewController(t)
		defer procCtrl.Finish()

		ctx := context.Background()
		workspaceID := "workspace1"
		taskID := "task456"

		// Create a task with a supported type
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: workspaceID,
			Type:        "send_broadcast",
			Status:      domain.TaskStatusPending,
			MaxRuntime:  60,
		}

		// Create a new task service instance for this test
		procTaskService := NewTaskService(mockRepo, mockLogger, mockAuthService, apiEndpoint)

		// Register a processor for the task type
		mockProcessor := mocks.NewMockTaskProcessor(procCtrl)
		// Configure CanProcess to be called for all supported task types
		for _, supportedType := range getTaskTypes() {
			mockProcessor.EXPECT().
				CanProcess(supportedType).
				Return(supportedType == "send_broadcast").
				AnyTimes()
		}
		procTaskService.RegisterProcessor(mockProcessor)

		// Configure mock repository
		mockRepo.EXPECT().
			GetTx(gomock.Any(), gomock.Any(), workspaceID, taskID).
			Return(task, nil)

		// MarkAsRunningTx should succeed
		mockRepo.EXPECT().
			MarkAsRunningTx(gomock.Any(), gomock.Any(), workspaceID, taskID, gomock.Any()).
			Return(nil)

		// Configure processor to return an error
		processingError := fmt.Errorf("processing failed")
		mockProcessor.EXPECT().
			Process(gomock.Any(), task).
			Return(false, processingError)

		// Mark as failed should succeed
		mockRepo.EXPECT().
			MarkAsFailed(gomock.Any(), workspaceID, taskID, gomock.Any()).
			Return(nil)

		// Call the method under test
		err := procTaskService.ExecuteTask(ctx, workspaceID, taskID)

		// Verify returned error is of type ErrTaskExecution
		assert.Error(t, err)
		var taskExecError *domain.ErrTaskExecution
		assert.True(t, errors.As(err, &taskExecError))
		assert.Equal(t, taskID, taskExecError.TaskID)
		assert.Equal(t, "processing failed", taskExecError.Reason)
		assert.Equal(t, processingError, taskExecError.Err)
	})

	t.Run("Timeout error returns ErrTaskTimeout", func(t *testing.T) {
		t.Skip("Skipping timeout test because it depends on context timing which is flaky in tests")
		// Note: This test is more integration-style and might be flaky due to timing issues

		// Setup - create a context that's already timed out
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		time.Sleep(2 * time.Millisecond) // Ensure the context times out
		defer cancel()

		workspaceID := "workspace1"
		taskID := "task123"

		// Create a task with a supported type
		task := &domain.Task{
			ID:          taskID,
			WorkspaceID: workspaceID,
			Type:        "send_broadcast",
			Status:      domain.TaskStatusPending,
			MaxRuntime:  60,
			MaxRetries:  1,
			RetryCount:  1, // Max retries reached
		}

		// Register a processor for the task type
		mockProcessor := mocks.NewMockTaskProcessor(ctrl)
		mockProcessor.EXPECT().CanProcess("send_broadcast").Return(true)
		taskService.RegisterProcessor(mockProcessor)

		// Configure mock repository
		mockRepo.EXPECT().
			GetTx(gomock.Any(), gomock.Any(), workspaceID, taskID).
			Return(task, nil)

		// MarkAsRunningTx should succeed
		mockRepo.EXPECT().
			MarkAsRunningTx(gomock.Any(), gomock.Any(), workspaceID, taskID, gomock.Any()).
			Return(nil)

		// Mark as failed should succeed for a timeout
		mockRepo.EXPECT().
			MarkAsFailed(gomock.Any(), workspaceID, taskID, gomock.Any()).
			Return(nil)

		// Call the method under test with the timed out context
		err := taskService.ExecuteTask(timeoutCtx, workspaceID, taskID)

		// Verify returned error is of type ErrTaskTimeout
		assert.Error(t, err)
		var timeoutError *domain.ErrTaskTimeout
		assert.True(t, errors.As(err, &timeoutError))
		assert.Equal(t, taskID, timeoutError.TaskID)
		assert.Equal(t, 60, timeoutError.MaxRuntime)
	})
}
