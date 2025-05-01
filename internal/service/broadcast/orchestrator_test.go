package broadcast_test

import (
	"context"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service/broadcast"
	"github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a compatible logger implementation
type testLoggerAdapter struct{}

func (l *testLoggerAdapter) Debug(msg string)                                       {}
func (l *testLoggerAdapter) Info(msg string)                                        {}
func (l *testLoggerAdapter) Warn(msg string)                                        {}
func (l *testLoggerAdapter) Error(msg string)                                       {}
func (l *testLoggerAdapter) Fatal(msg string)                                       {}
func (l *testLoggerAdapter) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *testLoggerAdapter) WithFields(fields map[string]interface{}) logger.Logger { return l }

func TestBroadcastOrchestrator_CanProcess(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		nil, // Use default config
	)

	// Test cases
	tests := []struct {
		taskType string
		expected bool
	}{
		{"send_broadcast", true},
		{"other_task", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.taskType, func(t *testing.T) {
			result := orchestrator.CanProcess(tc.taskType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBroadcastOrchestrator_LoadTemplatesForBroadcast(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		nil, // Use default config
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"

	// Mock a broadcast with template variations
	testBroadcast := &domain.Broadcast{
		TestSettings: domain.BroadcastTestSettings{
			Variations: []domain.BroadcastVariation{
				{TemplateID: "template-1"},
				{TemplateID: "template-2"},
			},
		},
	}

	// Mock template responses
	template1 := &domain.Template{
		ID: "template-1",
		Email: &domain.EmailTemplate{
			Subject:     "Test Subject 1",
			FromAddress: "test@example.com",
			VisualEditorTree: mjml.EmailBlock{
				Kind: "container",
				Data: map[string]interface{}{
					"styles": map[string]interface{}{},
				},
			},
		},
	}
	template2 := &domain.Template{
		ID: "template-2",
		Email: &domain.EmailTemplate{
			Subject:     "Test Subject 2",
			FromAddress: "test@example.com",
			VisualEditorTree: mjml.EmailBlock{
				Kind: "container",
				Data: map[string]interface{}{
					"styles": map[string]interface{}{},
				},
			},
		},
	}

	// Setup expectations
	mockBroadcastSender.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(testBroadcast, nil)

	mockTemplateService.EXPECT().
		GetTemplateByID(ctx, workspaceID, "template-1", int64(1)).
		Return(template1, nil)

	mockTemplateService.EXPECT().
		GetTemplateByID(ctx, workspaceID, "template-2", int64(1)).
		Return(template2, nil)

	// Execute
	templates, err := orchestrator.LoadTemplatesForBroadcast(ctx, workspaceID, broadcastID)

	// Verify
	require.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Equal(t, template1, templates["template-1"])
	assert.Equal(t, template2, templates["template-2"])
}

func TestBroadcastOrchestrator_ValidateTemplates(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		nil, // Use default config
	)

	// Test cases
	tests := []struct {
		name        string
		templates   map[string]*domain.Template
		expectError bool
	}{
		{
			name: "Valid templates",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:     "Test Subject",
						FromAddress: "test@example.com",
						VisualEditorTree: mjml.EmailBlock{
							Kind: "container",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:        "Empty templates",
			templates:   map[string]*domain.Template{},
			expectError: true,
		},
		{
			name: "Missing email config",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
				},
			},
			expectError: true,
		},
		{
			name: "Missing from address",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject: "Test Subject",
						VisualEditorTree: mjml.EmailBlock{
							Kind: "container",
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "Missing subject",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
					Email: &domain.EmailTemplate{
						FromAddress: "test@example.com",
						VisualEditorTree: mjml.EmailBlock{
							Kind: "container",
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "Missing content",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:          "Test Subject",
						FromAddress:      "test@example.com",
						VisualEditorTree: mjml.EmailBlock{},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := orchestrator.ValidateTemplates(tc.templates)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBroadcastOrchestrator_GetTotalRecipientCount(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		nil, // Use default config
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"

	// Mock broadcast with audience
	audience := domain.AudienceSettings{
		Lists:    []string{"list-1", "list-2"},
		Segments: []string{"segment-1"},
	}
	testBroadcast := &domain.Broadcast{
		Audience: audience,
	}

	// Setup expectations
	mockBroadcastSender.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(testBroadcast, nil)

	mockContactRepo.EXPECT().
		CountContactsForBroadcast(ctx, workspaceID, audience).
		Return(150, nil)

	// Execute
	count, err := orchestrator.GetTotalRecipientCount(ctx, workspaceID, broadcastID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 150, count)
}

func TestBroadcastOrchestrator_FetchBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	config := &broadcast.Config{
		FetchBatchSize: 50,
	}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		config,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	offset := 0
	limit := 100

	// Mock broadcast with audience
	audience := domain.AudienceSettings{
		Lists:    []string{"list-1", "list-2"},
		Segments: []string{"segment-1"},
	}
	testBroadcast := &domain.Broadcast{
		Audience: audience,
	}

	// Create mock contacts
	mockContacts := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{Email: "user1@example.com"},
			ListID:  "list-1",
		},
		{
			Contact: &domain.Contact{Email: "user2@example.com"},
			ListID:  "list-2",
		},
	}

	// Setup expectations
	mockBroadcastSender.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(testBroadcast, nil)

	mockContactRepo.EXPECT().
		GetContactsForBroadcast(ctx, workspaceID, audience, limit, offset).
		Return(mockContacts, nil)

	// Execute
	contacts, err := orchestrator.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, mockContacts, contacts)
	assert.Len(t, contacts, 2)
}

func TestBroadcastOrchestrator_Process(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	testLogger := &testLoggerAdapter{}

	config := &broadcast.Config{
		FetchBatchSize:      50,
		MaxProcessTime:      1 * time.Second,
		ProgressLogInterval: 500 * time.Millisecond,
	}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		testLogger,
		config,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"

	// Create a task with existing state, but we'll modify the orchestrator test
	// to handle if the task state is reset and the orchestrator tries to rebuild it
	task := &domain.Task{
		ID:          "task-123",
		WorkspaceID: workspaceID,
		Type:        "send_broadcast",
		Status:      domain.TaskStatusRunning,
		State: &domain.TaskState{
			SendBroadcast: &domain.SendBroadcastState{
				BroadcastID:     broadcastID,
				TotalRecipients: 150, // Setting a value > 0 to test the regular processing path
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0,
			},
		},
	}

	// Mock broadcast with audience for the GetBroadcast call
	audience := domain.AudienceSettings{
		Lists:    []string{"list-1", "list-2"},
		Segments: []string{"segment-1"},
	}

	testBroadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audience,
		TestSettings: domain.BroadcastTestSettings{
			Variations: []domain.BroadcastVariation{
				{TemplateID: "template-1"},
			},
		},
	}

	// Create mock templates
	mockTemplates := map[string]*domain.Template{
		"template-1": {
			ID: "template-1",
			Email: &domain.EmailTemplate{
				Subject:     "Test Subject",
				FromAddress: "test@example.com",
				VisualEditorTree: mjml.EmailBlock{
					Kind: "container",
					Data: map[string]interface{}{
						"styles": map[string]interface{}{},
					},
				},
			},
		},
	}

	// Create mock contacts (empty to simulate completion)
	emptyContacts := []*domain.ContactWithList{}

	// Setup expectations for the GetAPIEndpoint method
	mockBroadcastSender.EXPECT().
		GetAPIEndpoint().
		Return("https://api.example.com").
		AnyTimes()

	// Expect GetBroadcast to be called
	mockBroadcastSender.EXPECT().
		GetBroadcast(gomock.Any(), workspaceID, broadcastID).
		Return(testBroadcast, nil).
		AnyTimes()

	// Expect GetTemplateByID to be called
	mockTemplateService.EXPECT().
		GetTemplateByID(gomock.Any(), workspaceID, "template-1", int64(1)).
		Return(mockTemplates["template-1"], nil).
		AnyTimes()

	// Expect CountContactsForBroadcast to be called
	mockContactRepo.EXPECT().
		CountContactsForBroadcast(gomock.Any(), workspaceID, audience).
		Return(0, nil).
		AnyTimes()

	// For the batch fetching, return empty contacts to signal completion
	mockContactRepo.EXPECT().
		GetContactsForBroadcast(gomock.Any(), workspaceID, audience, 50, 0).
		Return(emptyContacts, nil).
		AnyTimes()

	// For state saving
	mockTaskRepo.EXPECT().
		SaveState(gomock.Any(), workspaceID, task.ID, gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Execute
	completed, err := orchestrator.Process(ctx, task)

	// Verify
	require.NoError(t, err)
	assert.True(t, completed)
	// More specific assertions depend on the implementation
	// so we'll just verify that no error occurred and the task completed
}
