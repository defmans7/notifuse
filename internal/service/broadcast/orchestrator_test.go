package broadcast

import (
	"context"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	bmocks "github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestOrchestratorCanProcess(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockTemplateLoader := bmocks.NewMockTemplateLoader(ctrl)
	mockRecipientFetcher := bmocks.NewMockRecipientFetcher(ctrl)
	mockMessageSender := bmocks.NewMockMessageSender(ctrl)
	mockProgressTracker := bmocks.NewMockProgressTracker(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create the orchestrator
	orchestrator := NewBroadcastOrchestrator(
		mockTemplateLoader,
		mockRecipientFetcher,
		mockMessageSender,
		mockProgressTracker,
		mockLogger,
		nil, // Use default config
	)

	// Test CanProcess with valid task type
	assert.True(t, orchestrator.CanProcess("send_broadcast"), "Should be able to process 'send_broadcast'")

	// Test CanProcess with invalid task type
	assert.False(t, orchestrator.CanProcess("other_task"), "Should not process 'other_task'")
}

func TestOrchestrator_Process_EmptyState(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockTemplateLoader := bmocks.NewMockTemplateLoader(ctrl)
	mockRecipientFetcher := bmocks.NewMockRecipientFetcher(ctrl)
	mockMessageSender := bmocks.NewMockMessageSender(ctrl)
	mockProgressTracker := bmocks.NewMockProgressTracker(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Create test task with a broadcast ID
	broadcastID := "broadcast-123"
	task := &domain.Task{
		ID:          "task-123",
		WorkspaceID: "workspace-123",
		BroadcastID: &broadcastID,
		Type:        "send_broadcast",
		State:       nil, // Empty state to test initialization
	}

	// Set up expectations
	// Expect call to get total recipient count
	mockRecipientFetcher.EXPECT().
		GetTotalRecipientCount(gomock.Any(), task.WorkspaceID, broadcastID).
		Return(100, nil)

	// Create the orchestrator
	orchestrator := NewBroadcastOrchestrator(
		mockTemplateLoader,
		mockRecipientFetcher,
		mockMessageSender,
		mockProgressTracker,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	done, err := orchestrator.Process(context.Background(), task)

	// Verify results
	assert.NoError(t, err)
	assert.False(t, done, "Should not be done after first call with recipients > 0")
	assert.NotNil(t, task.State, "Task state should be initialized")
	assert.NotNil(t, task.State.SendBroadcast, "SendBroadcast state should be initialized")
	assert.Equal(t, 100, task.State.SendBroadcast.TotalRecipients)
	assert.Equal(t, broadcastID, task.State.SendBroadcast.BroadcastID)
}

func TestMockOrchestrator(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock orchestrator
	mockOrchestrator := bmocks.NewMockBroadcastOrchestratorInterface(ctrl)

	// Set up expectations
	mockOrchestrator.EXPECT().
		CanProcess("send_broadcast").
		Return(true)

	mockOrchestrator.EXPECT().
		Process(gomock.Any(), gomock.Any()).
		Return(true, nil)

	// Use the mock
	assert.True(t, mockOrchestrator.CanProcess("send_broadcast"))

	// Create a simple task
	task := &domain.Task{
		ID:   "task-123",
		Type: "send_broadcast",
	}

	// Call process with the task
	done, err := mockOrchestrator.Process(context.Background(), task)

	// Verify results
	assert.True(t, done)
	assert.NoError(t, err)
}
