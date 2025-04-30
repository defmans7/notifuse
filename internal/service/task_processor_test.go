package service

import (
	"context"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
)

func TestSendBroadcastProcessor_CanProcess(t *testing.T) {
	// Create a mock broadcast service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockLogger := &testLogger{}

	processor := NewSendBroadcastProcessor(mockBroadcastService, mockLogger)

	// Test that it can process send_broadcast tasks
	if !processor.CanProcess("send_broadcast") {
		t.Error("Expected processor to handle send_broadcast tasks")
	}

	// Test that it cannot process other task types
	if processor.CanProcess("some_other_task") {
		t.Error("Expected processor to reject non-send_broadcast tasks")
	}
}

func TestSendBroadcastProcessor_Process_MissingBroadcastID(t *testing.T) {
	// Create a mock broadcast service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockLogger := &testLogger{}

	processor := NewSendBroadcastProcessor(mockBroadcastService, mockLogger)

	// Create a task with no broadcast ID
	task := &domain.Task{
		ID:          "task123",
		WorkspaceID: "workspace1",
		Type:        "send_broadcast",
		Status:      domain.TaskStatusPending,
		State: &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0,
				// No BroadcastID set
			},
		},
	}

	// Process should return an error
	completed, err := processor.Process(context.Background(), task)

	// Check that it did not complete
	if completed {
		t.Error("Expected process to return not completed")
	}

	// Verify error type and message
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	// Check for the specific error type
	taskErr, ok := err.(*domain.ErrTaskExecution)
	if !ok {
		t.Fatalf("Expected ErrTaskExecution, got %T", err)
	}

	if taskErr.TaskID != "task123" {
		t.Errorf("Expected task ID task123, got %s", taskErr.TaskID)
	}

	if taskErr.Reason != "broadcast ID is missing in task state" {
		t.Errorf("Expected reason 'broadcast ID is missing in task state', got '%s'", taskErr.Reason)
	}
}

func TestSendBroadcastProcessor_Process_WithBroadcastID_FromTask(t *testing.T) {
	// Create a mock broadcast service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockLogger := &testLogger{}

	processor := NewSendBroadcastProcessor(mockBroadcastService, mockLogger)

	// Create a broadcast ID
	broadcastID := "broadcast123"

	// Create a task with broadcast ID in the task but not in the state
	task := &domain.Task{
		ID:          "task123",
		WorkspaceID: "workspace1",
		Type:        "send_broadcast",
		Status:      domain.TaskStatusPending,
		BroadcastID: &broadcastID,
		State: &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0,
				// No BroadcastID set here, should be copied from task
			},
		},
	}

	// Set up mock to expect GetRecipientCount call with the correct broadcast ID
	mockBroadcastService.EXPECT().
		GetRecipientCount(gomock.Any(), "workspace1", "broadcast123").
		Return(10, nil)

	// Process should not return an error
	completed, err := processor.Process(context.Background(), task)

	// Check that it did not complete yet (this is first run)
	if completed {
		t.Error("Expected process to return not completed on first run")
	}

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that broadcast ID was copied to the state
	if task.State.SendBroadcast.BroadcastID != "broadcast123" {
		t.Errorf("Expected broadcast ID to be copied to state, got %s", task.State.SendBroadcast.BroadcastID)
	}
}

func TestSendBroadcastProcessor_Process_BroadcastNotFound(t *testing.T) {
	// Create a mock broadcast service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockLogger := &testLogger{}

	processor := NewSendBroadcastProcessor(mockBroadcastService, mockLogger)

	// Create a task with broadcast ID
	task := &domain.Task{
		ID:          "task123",
		WorkspaceID: "workspace1",
		Type:        "send_broadcast",
		Status:      domain.TaskStatusPending,
		State: &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				BroadcastID:     "broadcast123",
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0,
			},
		},
	}

	// Return a "not found" error from GetRecipientCount
	notFoundErr := &domain.ErrNotFound{
		Entity: "broadcast",
		ID:     "broadcast123",
	}
	mockBroadcastService.EXPECT().
		GetRecipientCount(gomock.Any(), "workspace1", "broadcast123").
		Return(0, notFoundErr)

	// Process should return an error
	completed, err := processor.Process(context.Background(), task)

	// Check that it did not complete
	if completed {
		t.Error("Expected process to return not completed")
	}

	// Verify error type and message
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	// Check for the specific error type
	taskErr, ok := err.(*domain.ErrTaskExecution)
	if !ok {
		t.Fatalf("Expected ErrTaskExecution, got %T", err)
	}

	if taskErr.TaskID != "task123" {
		t.Errorf("Expected task ID task123, got %s", taskErr.TaskID)
	}

	if taskErr.Reason != "broadcast not found" {
		t.Errorf("Expected reason 'broadcast not found', got '%s'", taskErr.Reason)
	}

	// Check that original error is wrapped
	if taskErr.Err != notFoundErr {
		t.Errorf("Expected original error to be wrapped, got %v", taskErr.Err)
	}
}

func TestSendBroadcastProcessor_Process_NoRecipients(t *testing.T) {
	// Create a mock broadcast service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockLogger := &testLogger{}

	processor := NewSendBroadcastProcessor(mockBroadcastService, mockLogger)

	// Create a task with broadcast ID
	task := &domain.Task{
		ID:          "task123",
		WorkspaceID: "workspace1",
		Type:        "send_broadcast",
		Status:      domain.TaskStatusPending,
		State: &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				BroadcastID:     "broadcast123",
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0,
			},
		},
	}

	// Return 0 recipients
	mockBroadcastService.EXPECT().
		GetRecipientCount(gomock.Any(), "workspace1", "broadcast123").
		Return(0, nil)

	// Process should return completed=true
	completed, err := processor.Process(context.Background(), task)

	// Check that it completed immediately
	if !completed {
		t.Error("Expected process to return completed for zero recipients")
	}

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify progress is 100%
	if task.Progress != 100.0 {
		t.Errorf("Expected progress 100.0, got %f", task.Progress)
	}

	// Verify status message
	expectedMsg := "Broadcast completed: No recipients found"
	if task.State.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, task.State.Message)
	}
}

// Minimal test logger implementation
type testLogger struct{}

func (l *testLogger) Debug(msg string)                                       {}
func (l *testLogger) Info(msg string)                                        {}
func (l *testLogger) Warn(msg string)                                        {}
func (l *testLogger) Error(msg string)                                       {}
func (l *testLogger) Fatal(msg string)                                       {}
func (l *testLogger) Debugf(format string, args ...interface{})              {}
func (l *testLogger) Infof(format string, args ...interface{})               {}
func (l *testLogger) Warnf(format string, args ...interface{})               {}
func (l *testLogger) Errorf(format string, args ...interface{})              {}
func (l *testLogger) Fatalf(format string, args ...interface{})              {}
func (l *testLogger) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *testLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }
func (l *testLogger) WithError(err error) logger.Logger                      { return l }
