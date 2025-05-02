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
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastOrchestrator_CanProcess(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
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
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
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

func TestBroadcastOrchestrator_LoadTemplatesForBroadcast(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
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
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
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
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
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
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
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
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Setup logger expectations - ensure all possible calls are mocked
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	config := &broadcast.Config{
		FetchBatchSize: 50,
	}

	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		config,
		mockTimeProvider,
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

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 45*time.Second,
			expected: "2m 45s",
		},
		{
			name:     "hours and minutes",
			duration: 3*time.Hour + 25*time.Minute,
			expected: "3h 25m",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "large duration",
			duration: 24*time.Hour + 30*time.Minute + 15*time.Second,
			expected: "24h 30m",
		},
		{
			name:     "exact seconds",
			duration: 1 * time.Second,
			expected: "1s",
		},
		{
			name:     "exact minute",
			duration: 1 * time.Minute,
			expected: "1m 0s",
		},
		{
			name:     "exact hour",
			duration: 1 * time.Hour,
			expected: "1h 0m",
		},
		{
			name:     "milliseconds rounded",
			duration: 1500 * time.Millisecond,
			expected: "1s",
		},
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
		{
			name:      "zero total",
			processed: 10,
			total:     0,
			expected:  100.0, // Avoid division by zero
		},
		{
			name:      "zero processed",
			processed: 0,
			total:     100,
			expected:  0.0,
		},
		{
			name:      "half processed",
			processed: 50,
			total:     100,
			expected:  50.0,
		},
		{
			name:      "fully processed",
			processed: 100,
			total:     100,
			expected:  100.0,
		},
		{
			name:      "more than total processed",
			processed: 110,
			total:     100,
			expected:  100.0, // Cap at 100%
		},
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
		contains  string // We test for contains rather than exact match due to ETA variations
	}{
		{
			name:      "initial progress no ETA",
			processed: 5,
			total:     100,
			elapsed:   10 * time.Second,
			contains:  "Processed 5/100 recipients (5.0%)",
		},
		{
			name:      "progress with ETA",
			processed: 10,
			total:     100,
			elapsed:   20 * time.Second,
			contains:  "Processed 10/100 recipients (10.0%), ETA:", // ETA will vary
		},
		{
			name:      "completed",
			processed: 100,
			total:     100,
			elapsed:   2 * time.Minute,
			contains:  "Processed 100/100 recipients (100.0%)",
		},
		{
			name:      "zero total",
			processed: 0,
			total:     0,
			elapsed:   5 * time.Second,
			contains:  "Processed 0/0 recipients (100.0%)", // Avoid division by zero
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := broadcast.FormatProgressMessage(tc.processed, tc.total, tc.elapsed)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestSaveProgressState(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageSender := mocks.NewMockMessageSender(ctrl)
	mockBroadcastSender := domainmocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	mockContactRepo := domainmocks.NewMockContactRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTimeProvider := mocks.NewMockTimeProvider(ctrl)

	// Set fixed times for testing
	testStartTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	later := testStartTime.Add(10 * time.Second)

	// Setup time provider expectations
	mockTimeProvider.EXPECT().Now().Return(later).AnyTimes()

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Use the concrete type instead of the interface
	orchestrator := broadcast.NewBroadcastOrchestrator(
		mockMessageSender,
		mockBroadcastSender,
		mockTemplateService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		nil, // Use default config
		mockTimeProvider,
	).(*broadcast.BroadcastOrchestrator)

	ctx := context.Background()
	workspaceID := "workspace-123"
	taskID := "task-456"
	broadcastID := "broadcast-789"
	totalRecipients := 100
	sentCount := 50
	failedCount := 10
	processedCount := 60
	lastSaveTime := testStartTime.Add(-10 * time.Second) // Ensure enough time has passed to save

	// Test case 1: Successful state save
	mockTaskRepo.EXPECT().
		SaveState(
			ctx,
			workspaceID,
			taskID,
			gomock.Any(), // Progress percentage
			gomock.Any(), // State object
		).
		Return(nil).
		Times(1)

	// Execute
	result, err := orchestrator.SaveProgressState(
		ctx,
		workspaceID,
		taskID,
		broadcastID,
		totalRecipients,
		sentCount,
		failedCount,
		processedCount,
		lastSaveTime,
		testStartTime,
	)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, later, result)

	// Test case 2: State save with error
	lastSaveTime = testStartTime.Add(-10 * time.Second) // Reset

	mockTaskRepo.EXPECT().
		SaveState(
			ctx,
			workspaceID,
			taskID,
			gomock.Any(), // Progress percentage
			gomock.Any(), // State object
		).
		Return(fmt.Errorf("database error")).
		Times(1)

	// Execute
	result, err = orchestrator.SaveProgressState(
		ctx,
		workspaceID,
		taskID,
		broadcastID,
		totalRecipients,
		sentCount,
		failedCount,
		processedCount,
		lastSaveTime,
		testStartTime,
	)

	// Verify
	require.Error(t, err)
	assert.Equal(t, lastSaveTime, result) // Should return the original lastSaveTime on error

	// Test case 3: Not enough time passed to save
	recentSaveTime := later.Add(-2 * time.Second) // Only 2 seconds passed (< 5 required)

	// Don't expect SaveState to be called

	// Execute
	result, err = orchestrator.SaveProgressState(
		ctx,
		workspaceID,
		taskID,
		broadcastID,
		totalRecipients,
		sentCount,
		failedCount,
		processedCount,
		recentSaveTime,
		testStartTime,
	)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, recentSaveTime, result) // Should return the same time since no save happened
}
