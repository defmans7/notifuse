package broadcast_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service/broadcast"
	"github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/Notifuse/notifuse/pkg/notifuse_mjml"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastOrchestrator_CanProcess(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
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

func TestBroadcastOrchestrator_LoadTemplates(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	templateIDs := []string{"template-1", "template-2"}

	// Mock template responses
	template1 := &domain.Template{
		ID: "template-1",
		Email: &domain.EmailTemplate{
			Subject:  "Test Subject 1",
			SenderID: "sender-123",
			VisualEditorTree: &notifuse_mjml.MJMLBlock{
				BaseBlock: notifuse_mjml.BaseBlock{
					ID:   "root1",
					Type: notifuse_mjml.MJMLComponentMjml,
					Attributes: map[string]interface{}{
						"version": "4.0.0",
					},
				},
			},
		},
	}
	template2 := &domain.Template{
		ID: "template-2",
		Email: &domain.EmailTemplate{
			Subject:  "Test Subject 2",
			SenderID: "sender-123",
			VisualEditorTree: &notifuse_mjml.MJMLBlock{
				BaseBlock: notifuse_mjml.BaseBlock{
					ID:   "root2",
					Type: notifuse_mjml.MJMLComponentMjml,
					Attributes: map[string]interface{}{
						"version": "4.0.0",
					},
				},
			},
		},
	}

	// Setup expectations
	mockTemplateRepo.EXPECT().
		GetTemplateByID(ctx, workspaceID, "template-1", int64(0)).
		Return(template1, nil)

	mockTemplateRepo.EXPECT().
		GetTemplateByID(ctx, workspaceID, "template-2", int64(0)).
		Return(template2, nil)

	// Execute
	templates, err := orchestrator.LoadTemplates(ctx, workspaceID, templateIDs)

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
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations - ensure all possible calls are mocked
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
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
						Subject:  "Test Subject",
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
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
			name: "Missing subject",
			templates: map[string]*domain.Template{
				"template-1": {
					ID: "template-1",
					Email: &domain.EmailTemplate{
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
						},
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
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"

	// Mock broadcast
	testBroadcast := &domain.Broadcast{
		Audience: domain.AudienceSettings{
			Lists: []string{"list-1", "list-2"},
		},
	}

	// Setup expectations
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(testBroadcast, nil)

	mockContactRepo.EXPECT().
		CountContactsForBroadcast(ctx, workspaceID, testBroadcast.Audience).
		Return(100, nil)

	// Execute
	count, err := orchestrator.GetTotalRecipientCount(ctx, workspaceID, broadcastID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 100, count)
}

func TestBroadcastOrchestrator_FetchBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	offset := 0
	limit := 50

	// Mock broadcast
	testBroadcast := &domain.Broadcast{
		Audience: domain.AudienceSettings{
			Lists: []string{"list-1"},
		},
	}

	// Mock contacts
	expectedContacts := []*domain.ContactWithList{
		{Contact: &domain.Contact{Email: "user1@example.com"}, ListID: "list-1"},
		{Contact: &domain.Contact{Email: "user2@example.com"}, ListID: "list-1"},
	}

	// Setup expectations
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(testBroadcast, nil)

	mockContactRepo.EXPECT().
		GetContactsForBroadcast(ctx, workspaceID, testBroadcast.Audience, limit, offset).
		Return(expectedContacts, nil)

	// Execute
	contacts, err := orchestrator.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, expectedContacts, contacts)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"minutes_and_seconds", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"hours_and_minutes", 1*time.Hour + 30*time.Minute, "1h 30m"},
		{"zero_duration", 0, "0s"},
		{"large_duration", 25*time.Hour + 90*time.Minute, "26h 30m"},
		{"exact_seconds", 60 * time.Second, "1m 0s"},
		{"exact_minute", 60 * time.Minute, "1h 0m"},
		{"exact_hour", 1 * time.Hour, "1h 0m"},
		{"milliseconds_rounded", 1500 * time.Millisecond, "1s"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := broadcast.FormatDuration(tc.duration)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		name      string
		processed int
		total     int
		expected  float64
	}{
		{"zero_total", 0, 0, 100.0},
		{"zero_processed", 0, 100, 0.0},
		{"half_processed", 50, 100, 50.0},
		{"fully_processed", 100, 100, 100.0},
		{"more_than_total_processed", 150, 100, 100.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := broadcast.CalculateProgress(tc.processed, tc.total)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatProgressMessage(t *testing.T) {
	tests := []struct {
		name      string
		processed int
		total     int
		elapsed   time.Duration
		expected  string
	}{
		{"initial_progress_no_ETA", 0, 100, 5 * time.Second, "Processed 0/100 recipients (0.0%)"},
		{"progress_with_ETA", 25, 100, 1 * time.Minute, "Processed 25/100 recipients (25.0%), ETA: 3m 0s"},
		{"completed", 100, 100, 2 * time.Minute, "Processed 100/100 recipients (100.0%)"},
		{"zero_total", 0, 0, 30 * time.Second, "Processed 0/0 recipients (100.0%)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := broadcast.FormatProgressMessage(tc.processed, tc.total, tc.elapsed)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSaveProgressState(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastRepository := domainmocks.NewMockBroadcastRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastRepository,
		mockTemplateRepo,
		mockContactRepo,
		mockTaskRepo,
		mockWorkspaceRepo,
		nil, // abTestEvaluator not needed for tests,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
	)

	ctx := context.Background()
	workspaceID := "workspace-123"
	taskID := "task-123"
	broadcastID := "broadcast-123"
	totalRecipients := 100
	sentCount := 25
	failedCount := 5
	processedCount := 30
	lastSaveTime := time.Now().Add(-10 * time.Second)
	startTime := time.Now().Add(-1 * time.Minute)

	// Mock time provider
	currentTime := time.Now()
	mockTimeProvider.EXPECT().Now().Return(currentTime).AnyTimes()
	mockTimeProvider.EXPECT().Since(gomock.Any()).Return(5 * time.Second).AnyTimes()

	// Setup expectations
	mockTaskRepo.EXPECT().
		SaveState(ctx, workspaceID, taskID, gomock.Any(), gomock.Any()).
		Return(nil)

	// Execute
	newSaveTime, err := orchestrator.SaveProgressState(
		ctx, workspaceID, taskID, broadcastID,
		totalRecipients, sentCount, failedCount, processedCount,
		lastSaveTime, startTime,
	)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, currentTime, newSaveTime)
}

// TestBroadcastOrchestrator_Process tests the main Process method covering lines 594-795
func TestBroadcastOrchestrator_Process(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider)
		task          *domain.Task
		expectedDone  bool
		expectedError bool
		errorContains string
	}{
		{
			name: "successful_process_with_recipients",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()
				mockTimeProvider.EXPECT().Since(gomock.Any()).Return(10 * time.Second).AnyTimes()

				// Mock workspace with email provider (this is called first in lines 594-795)
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:                "secret-key",
						EmailTrackingEnabled:     true,
						MarketingEmailProviderID: "marketing-provider-id",
					},
					Integrations: []domain.Integration{
						{
							ID:   "marketing-provider-id",
							Name: "Marketing Provider",
							Type: domain.IntegrationTypeEmail,
							EmailProvider: domain.EmailProvider{
								Kind: domain.EmailProviderKindSES,
								SES: &domain.AmazonSESSettings{
									AccessKey: "access-key",
									SecretKey: "secret-key",
									Region:    "us-east-1",
								},
							},
						},
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				// Mock broadcast for template loading
				broadcast := &domain.Broadcast{
					ID: "broadcast-123",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{TemplateID: "template-1"},
						},
					},
					Audience: domain.AudienceSettings{
						Lists: []string{"list-1"},
					},
					Status: domain.BroadcastStatusSending,
				}
				// First call for template loading, second call for status update on completion
				mockBroadcastRepo.EXPECT().GetBroadcast(gomock.Any(), "workspace-123", "broadcast-123").Return(broadcast, nil).Times(2)

				// Mock broadcast status update on completion
				mockBroadcastRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, b *domain.Broadcast) error {
					// Verify the broadcast status was updated to sent
					assert.Equal(t, domain.BroadcastStatusSent, b.Status)
					assert.NotNil(t, b.CompletedAt)
					return nil
				})

				// Mock template
				template := &domain.Template{
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:  "Test Subject",
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
						},
					},
				}
				mockTemplateRepo.EXPECT().GetTemplateByID(gomock.Any(), "workspace-123", "template-1", int64(0)).Return(template, nil)

				// Mock recipients - return fewer than batch size to indicate completion
				recipients := []*domain.ContactWithList{
					{Contact: &domain.Contact{Email: "user1@example.com"}, ListID: "list-1"},
					{Contact: &domain.Contact{Email: "user2@example.com"}, ListID: "list-1"},
				}
				// Expect batch size of 2 because remainingInPhase (2) < FetchBatchSize (100)
				mockContactRepo.EXPECT().GetContactsForBroadcast(gomock.Any(), "workspace-123", broadcast.Audience, 2, 0).Return(recipients, nil)

				// Mock message sending
				mockMessageSender.EXPECT().SendBatch(
					gomock.Any(),
					"workspace-123",
					"secret-key",
					true,
					"broadcast-123",
					recipients,
					gomock.Any(),
					gomock.Any(),
				).Return(2, 0, nil)

				// Mock task state saving
				mockTaskRepo.EXPECT().SaveState(gomock.Any(), "workspace-123", "task-123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 2, // Set to non-zero to skip recipient counting phase
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  true,
			expectedError: false,
		},
		{
			name: "broadcast_status_update_failure",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()
				mockTimeProvider.EXPECT().Since(gomock.Any()).Return(10 * time.Second).AnyTimes()

				// Mock workspace with email provider
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:                "secret-key",
						EmailTrackingEnabled:     true,
						MarketingEmailProviderID: "marketing-provider-id",
					},
					Integrations: []domain.Integration{
						{
							ID:   "marketing-provider-id",
							Name: "Marketing Provider",
							Type: domain.IntegrationTypeEmail,
							EmailProvider: domain.EmailProvider{
								Kind: domain.EmailProviderKindSES,
								SES: &domain.AmazonSESSettings{
									AccessKey: "access-key",
									SecretKey: "secret-key",
									Region:    "us-east-1",
								},
							},
						},
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				// Mock broadcast for template loading
				broadcast := &domain.Broadcast{
					ID: "broadcast-123",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{TemplateID: "template-1"},
						},
					},
					Audience: domain.AudienceSettings{
						Lists: []string{"list-1"},
					},
					Status: domain.BroadcastStatusSending,
				}
				// First call for template loading, second call for status update on completion
				mockBroadcastRepo.EXPECT().GetBroadcast(gomock.Any(), "workspace-123", "broadcast-123").Return(broadcast, nil).Times(2)

				// Mock broadcast status update failure on completion
				mockBroadcastRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

				// Mock template
				template := &domain.Template{
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:  "Test Subject",
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
						},
					},
				}
				mockTemplateRepo.EXPECT().GetTemplateByID(gomock.Any(), "workspace-123", "template-1", int64(0)).Return(template, nil)

				// Mock recipients - return fewer than batch size to indicate completion
				recipients := []*domain.ContactWithList{
					{Contact: &domain.Contact{Email: "user1@example.com"}, ListID: "list-1"},
				}
				// Expect batch size of 1 because remainingInPhase (1) < FetchBatchSize (100)
				mockContactRepo.EXPECT().GetContactsForBroadcast(gomock.Any(), "workspace-123", broadcast.Audience, 1, 0).Return(recipients, nil)

				// Mock message sending
				mockMessageSender.EXPECT().SendBatch(
					gomock.Any(),
					"workspace-123",
					"secret-key",
					true,
					"broadcast-123",
					recipients,
					gomock.Any(),
					gomock.Any(),
				).Return(1, 0, nil)

				// Mock task state saving
				mockTaskRepo.EXPECT().SaveState(gomock.Any(), "workspace-123", "task-123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 1, // Set to non-zero to skip recipient counting phase
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  false,
			expectedError: true,
			errorContains: "failed to update broadcast status to sent",
		},
		{
			name: "workspace_not_found",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()

				// Workspace not found
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(nil, fmt.Errorf("workspace not found"))

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 2,
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  false,
			expectedError: true,
			errorContains: "failed to get workspace",
		},
		{
			name: "no_email_provider_configured",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()

				// Mock workspace without email provider
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:            "secret-key",
						EmailTrackingEnabled: true,
						// No MarketingEmailProviderID configured
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 2,
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  false,
			expectedError: true,
			errorContains: "no email provider configured for marketing emails",
		},
		{
			name: "template_loading_failure",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()

				// Mock workspace with email provider
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:                "secret-key",
						EmailTrackingEnabled:     true,
						MarketingEmailProviderID: "marketing-provider-id",
					},
					Integrations: []domain.Integration{
						{
							ID:   "marketing-provider-id",
							Name: "Marketing Provider",
							Type: domain.IntegrationTypeEmail,
							EmailProvider: domain.EmailProvider{
								Kind: domain.EmailProviderKindSES,
								SES: &domain.AmazonSESSettings{
									AccessKey: "access-key",
									SecretKey: "secret-key",
									Region:    "us-east-1",
								},
							},
						},
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				// Mock broadcast for template loading
				broadcast := &domain.Broadcast{
					ID: "broadcast-123",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{TemplateID: "template-1"},
						},
					},
					Audience: domain.AudienceSettings{
						Lists: []string{"list-1"},
					},
				}
				mockBroadcastRepo.EXPECT().GetBroadcast(gomock.Any(), "workspace-123", "broadcast-123").Return(broadcast, nil)

				// Template loading failure
				mockTemplateRepo.EXPECT().GetTemplateByID(gomock.Any(), "workspace-123", "template-1", int64(0)).Return(nil, fmt.Errorf("template not found"))

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 2,
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  false,
			expectedError: true,
			errorContains: "no valid templates found for broadcast",
		},
		{
			name: "recipient_fetch_failure",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()

				// Mock workspace with email provider
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:                "secret-key",
						EmailTrackingEnabled:     true,
						MarketingEmailProviderID: "marketing-provider-id",
					},
					Integrations: []domain.Integration{
						{
							ID:   "marketing-provider-id",
							Name: "Marketing Provider",
							Type: domain.IntegrationTypeEmail,
							EmailProvider: domain.EmailProvider{
								Kind: domain.EmailProviderKindSES,
								SES: &domain.AmazonSESSettings{
									AccessKey: "access-key",
									SecretKey: "secret-key",
									Region:    "us-east-1",
								},
							},
						},
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				// Mock broadcast for template loading
				broadcast := &domain.Broadcast{
					ID: "broadcast-123",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{TemplateID: "template-1"},
						},
					},
					Audience: domain.AudienceSettings{
						Lists: []string{"list-1"},
					},
				}
				mockBroadcastRepo.EXPECT().GetBroadcast(gomock.Any(), "workspace-123", "broadcast-123").Return(broadcast, nil).AnyTimes()

				// Mock template
				template := &domain.Template{
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:  "Test Subject",
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
						},
					},
				}
				mockTemplateRepo.EXPECT().GetTemplateByID(gomock.Any(), "workspace-123", "template-1", int64(0)).Return(template, nil)

				// Recipient fetch failure - expect batch size of 2 because remainingInPhase (2) < FetchBatchSize (100)
				mockContactRepo.EXPECT().GetContactsForBroadcast(gomock.Any(), "workspace-123", broadcast.Audience, 2, 0).Return(nil, fmt.Errorf("database error"))

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-123"),
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "broadcast-123",
						TotalRecipients: 2,
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  false,
			expectedError: true,
			errorContains: "failed to fetch recipients",
		},
		{
			name: "broadcast_id_from_task_field",
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockMessageSender, *domainmocks.MockBroadcastRepository, *domainmocks.MockTemplateRepository, *domainmocks.MockContactRepository, *domainmocks.MockTaskRepository, *domainmocks.MockWorkspaceRepository, *pkgmocks.MockLogger, *mocks.MockTimeProvider) {
				mockMessageSender := mocks.NewMockMessageSender(ctrl)
				mockBroadcastRepo := domainmocks.NewMockBroadcastRepository(ctrl)
				mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
				mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
				mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
				mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
				mockLogger := pkgmocks.NewMockLogger(ctrl)
				mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

				// Setup logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				// Setup time provider
				baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				mockTimeProvider.EXPECT().Now().Return(baseTime).AnyTimes()
				mockTimeProvider.EXPECT().Since(gomock.Any()).Return(10 * time.Second).AnyTimes()

				// Mock workspace with email provider
				workspace := &domain.Workspace{
					ID: "workspace-123",
					Settings: domain.WorkspaceSettings{
						SecretKey:                "secret-key",
						EmailTrackingEnabled:     true,
						MarketingEmailProviderID: "marketing-provider-id",
					},
					Integrations: []domain.Integration{
						{
							ID:   "marketing-provider-id",
							Name: "Marketing Provider",
							Type: domain.IntegrationTypeEmail,
							EmailProvider: domain.EmailProvider{
								Kind: domain.EmailProviderKindSES,
								SES: &domain.AmazonSESSettings{
									AccessKey: "access-key",
									SecretKey: "secret-key",
									Region:    "us-east-1",
								},
							},
						},
					},
				}
				mockWorkspaceRepo.EXPECT().GetByID(gomock.Any(), "workspace-123").Return(workspace, nil)

				// Mock broadcast for template loading
				broadcast := &domain.Broadcast{
					ID: "broadcast-456",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{TemplateID: "template-1"},
						},
					},
					Audience: domain.AudienceSettings{
						Lists: []string{"list-1"},
					},
					Status: domain.BroadcastStatusSending,
				}
				// First call for template loading, second call for status update on completion
				mockBroadcastRepo.EXPECT().GetBroadcast(gomock.Any(), "workspace-123", "broadcast-456").Return(broadcast, nil).Times(2)

				// Mock broadcast status update on completion
				mockBroadcastRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, b *domain.Broadcast) error {
					// Verify the broadcast status was updated to sent
					assert.Equal(t, domain.BroadcastStatusSent, b.Status)
					assert.NotNil(t, b.CompletedAt)
					return nil
				})

				// Mock template
				template := &domain.Template{
					ID: "template-1",
					Email: &domain.EmailTemplate{
						Subject:  "Test Subject",
						SenderID: "sender-123",
						VisualEditorTree: &notifuse_mjml.MJMLBlock{
							BaseBlock: notifuse_mjml.BaseBlock{
								ID:   "root1",
								Type: notifuse_mjml.MJMLComponentMjml,
								Attributes: map[string]interface{}{
									"version": "4.0.0",
								},
							},
						},
					},
				}
				mockTemplateRepo.EXPECT().GetTemplateByID(gomock.Any(), "workspace-123", "template-1", int64(0)).Return(template, nil)

				// Mock recipients - return fewer than batch size to indicate completion
				recipients := []*domain.ContactWithList{
					{Contact: &domain.Contact{Email: "user1@example.com"}, ListID: "list-1"},
				}
				// Expect batch size of 1 because remainingInPhase (1) < FetchBatchSize (100)
				mockContactRepo.EXPECT().GetContactsForBroadcast(gomock.Any(), "workspace-123", broadcast.Audience, 1, 0).Return(recipients, nil)

				// Mock message sending
				mockMessageSender.EXPECT().SendBatch(
					gomock.Any(),
					"workspace-123",
					"secret-key",
					true,
					"broadcast-456",
					recipients,
					gomock.Any(),
					gomock.Any(),
				).Return(1, 0, nil)

				// Mock task state saving
				mockTaskRepo.EXPECT().SaveState(gomock.Any(), "workspace-123", "task-123", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				return mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider
			},
			task: &domain.Task{
				ID:          "task-123",
				WorkspaceID: "workspace-123",
				Type:        "send_broadcast",
				BroadcastID: stringPtr("broadcast-456"), // Broadcast ID in task field
				State: &domain.TaskState{
					Progress: 0,
					Message:  "Starting broadcast",
					SendBroadcast: &domain.SendBroadcastState{
						BroadcastID:     "", // Empty broadcast ID in state
						TotalRecipients: 1,
						SentCount:       0,
						FailedCount:     0,
						RecipientOffset: 0,
					},
				},
				RetryCount: 0,
				MaxRetries: 3,
			},
			expectedDone:  true,
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMessageSender, mockBroadcastRepo, mockTemplateRepo, mockContactRepo, mockTaskRepo, mockWorkspaceRepo, mockLogger, mockTimeProvider := tc.setupMocks(ctrl)

			config := &broadcast.Config{
				FetchBatchSize:      100,
				MaxProcessTime:      30 * time.Second,
				ProgressLogInterval: 5 * time.Second,
			}

			orchestrator := broadcast.NewBroadcastOrchestrator(
				mockMessageSender,
				mockBroadcastRepo,
				mockTemplateRepo,
				mockContactRepo,
				mockTaskRepo,
				mockWorkspaceRepo,
				nil, // abTestEvaluator not needed for tests,
				mockLogger,
				config,
				mockTimeProvider,
			)

			// Execute
			ctx := context.Background()
			done, err := orchestrator.Process(ctx, tc.task)

			// Verify
			assert.Equal(t, tc.expectedDone, done)
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
