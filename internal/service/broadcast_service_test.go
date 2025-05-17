package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastService_CreateBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now().UTC()
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			Name:        "Test Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false,
			},
		}

		// Mock auth service to authenticate the user
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), request.WorkspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to expect a broadcast to be created
		mockRepo.EXPECT().
			CreateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, request.WorkspaceID, broadcast.WorkspaceID)
				assert.Equal(t, request.Name, broadcast.Name)
				assert.Equal(t, domain.BroadcastStatusDraft, broadcast.Status)
				assert.Equal(t, request.Audience, broadcast.Audience)
				assert.Equal(t, request.Schedule, broadcast.Schedule)
				assert.Equal(t, request.TestSettings, broadcast.TestSettings)
				assert.NotEmpty(t, broadcast.ID)
				assert.WithinDuration(t, now, broadcast.CreatedAt, 2*time.Second)
				return nil
			})

		// Call the service
		result, err := service.CreateBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.WorkspaceID, result.WorkspaceID)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, domain.BroadcastStatusDraft, result.Status)
	})

	t.Run("AuthenticationError", func(t *testing.T) {
		ctx := context.Background()
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			Name:        "Test Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
		}

		// Mock auth service to return authentication error
		authErr := errors.New("authentication failed")
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), request.WorkspaceID).
			Return(nil, nil, authErr)

		// We expect no repository calls due to authentication failure
		mockRepo.EXPECT().CreateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		result, err := service.CreateBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		// Create an invalid request (missing required fields)
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			// Missing Name and other required fields
		}

		// Mock auth service to authenticate the user
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), request.WorkspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// We expect validation to fail, no repository calls
		mockRepo.EXPECT().CreateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		result, err := service.CreateBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			Name:        "Test Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false,
			},
		}

		// Mock auth service to authenticate the user
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), request.WorkspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return an error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			CreateBroadcast(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		// Call the service
		result, err := service.CreateBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func TestBroadcastService_GetBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		expectedBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(expectedBroadcast, nil)

		// Call the service
		broadcast, err := service.GetBroadcast(ctx, workspaceID, broadcastID)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, expectedBroadcast, broadcast)
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		notFoundErr := &domain.ErrBroadcastNotFound{ID: broadcastID}
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// Call the service
		broadcast, err := service.GetBroadcast(ctx, workspaceID, broadcastID)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, broadcast)
		assert.Equal(t, notFoundErr, err)
	})
}

func TestBroadcastService_UpdateBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		// Create an existing broadcast with fixed timestamps
		createdTime := time.Now().Add(-24 * time.Hour).UTC()
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Original Broadcast",
			Status:      domain.BroadcastStatusDraft,
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false,
			},
			CreatedAt: createdTime,
			UpdatedAt: createdTime,
		}

		// Create update request
		updateRequest := &domain.UpdateBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			Name:        "Updated Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false,
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, broadcast.ID)
				assert.Equal(t, workspaceID, broadcast.WorkspaceID)
				assert.Equal(t, updateRequest.Name, broadcast.Name)
				assert.Equal(t, domain.BroadcastStatusDraft, broadcast.Status)

				// Just verify that updated time isn't zero
				assert.False(t, broadcast.UpdatedAt.IsZero())
				return nil
			})

		// Call the service
		result, err := service.UpdateBroadcast(ctx, updateRequest)

		// Verify results
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updateRequest.Name, result.Name)

		// Verify that updated time is later than creation time
		assert.True(t, result.UpdatedAt.After(result.CreatedAt))
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		updateRequest := &domain.UpdateBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			Name:        "Updated Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
		}

		// Mock repository to return not found error
		notFoundErr := errors.New("broadcast not found")
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// Call the service
		result, err := service.UpdateBroadcast(ctx, updateRequest)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notFoundErr, err)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		// Create an existing broadcast with invalid status
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Original Broadcast",
			Status:      domain.BroadcastStatusSent,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}

		// Create update request
		updateRequest := &domain.UpdateBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			Name:        "Updated Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false,
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No update expected due to validation failure
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		result, err := service.UpdateBroadcast(ctx, updateRequest)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot update broadcast with status")
	})
}

func TestBroadcastService_ScheduleBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	// Set up logger mock to return itself for chaining
	mockLoggerWithField := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLoggerWithField).AnyTimes()
	mockLoggerWithField.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("ScheduleForLater", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID:          workspaceID,
			ID:                   broadcastID,
			SendNow:              false,
			ScheduledDate:        time.Now().Add(time.Hour).Format("2006-01-02"),
			ScheduledTime:        time.Now().Add(time.Hour).Format("15:04"),
			Timezone:             "UTC",
			UseRecipientTimezone: false,
		}

		// Create a draft broadcast
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Create a workspace with a marketing email provider
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				MarketingEmailProviderID: "email-provider-123",
			},
			Integrations: []domain.Integration{
				{
					ID:   "email-provider-123",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSMTP,
						DefaultSenderEmail: "test@example.com",
						DefaultSenderName:  "Test Sender",
					},
				},
			},
		}

		// Mock workspace repository to return workspace with email provider
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Mock the WithTransaction call and execute the provided function with nil
		mockRepo.EXPECT().
			WithTransaction(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
				return fn(nil) // Pass nil as the transaction, it won't be used in the test
			})

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcastTx(gomock.Any(), gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcastTx(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *sql.Tx, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, broadcast.ID)
				assert.Equal(t, workspaceID, broadcast.WorkspaceID)
				assert.Equal(t, domain.BroadcastStatusScheduled, broadcast.Status)

				// Verify scheduled time using Schedule struct
				assert.True(t, broadcast.Schedule.IsScheduled)
				assert.NotEmpty(t, broadcast.Schedule.ScheduledDate)
				assert.NotEmpty(t, broadcast.Schedule.ScheduledTime)

				assert.Nil(t, broadcast.StartedAt) // Should not be set when scheduling for later
				return nil
			})

		// In the TestBroadcastService_ScheduleBroadcast test, find all test cases and add mockEventBus expectation before the service call
		mockEventBus.EXPECT().
			PublishWithAck(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, payload domain.EventPayload, callback domain.EventAckCallback) {
				// Call the callback with success
				callback(nil)
			}).
			AnyTimes()

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("SendNow", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			SendNow:     true,
		}

		// Create a draft broadcast
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Create a workspace with a marketing email provider
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				MarketingEmailProviderID: "email-provider-123",
			},
			Integrations: []domain.Integration{
				{
					ID:   "email-provider-123",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSMTP,
						DefaultSenderEmail: "test@example.com",
						DefaultSenderName:  "Test Sender",
					},
				},
			},
		}

		// Mock workspace repository to return workspace with email provider
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Mock the WithTransaction call and execute the provided function with nil
		mockRepo.EXPECT().
			WithTransaction(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
				return fn(nil) // Pass nil as the transaction, it won't be used in the test
			})

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcastTx(gomock.Any(), gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcastTx(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *sql.Tx, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, broadcast.ID)
				assert.Equal(t, workspaceID, broadcast.WorkspaceID)
				assert.Equal(t, domain.BroadcastStatusSending, broadcast.Status)

				// No scheduled time should be set in Schedule
				assert.False(t, broadcast.Schedule.IsScheduled)

				assert.NotNil(t, broadcast.StartedAt) // Should be set when sending now
				return nil
			})

		// In the TestBroadcastService_ScheduleBroadcast test, find all test cases and add mockEventBus expectation before the service call
		mockEventBus.EXPECT().
			PublishWithAck(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, payload domain.EventPayload, callback domain.EventAckCallback) {
				// Call the callback with success
				callback(nil)
			}).
			AnyTimes()

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("NonDraftStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID:          workspaceID,
			ID:                   broadcastID,
			SendNow:              false,
			ScheduledDate:        time.Now().Add(time.Hour).Format("2006-01-02"),
			ScheduledTime:        time.Now().Add(time.Hour).Format("15:04"),
			Timezone:             "UTC",
			UseRecipientTimezone: false,
		}

		// Create a broadcast with non-draft status
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSent, // Already sent
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Create a workspace with a marketing email provider
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				MarketingEmailProviderID: "email-provider-123",
			},
			Integrations: []domain.Integration{
				{
					ID:   "email-provider-123",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSMTP,
						DefaultSenderEmail: "test@example.com",
						DefaultSenderName:  "Test Sender",
					},
				},
			},
		}

		// Mock workspace repository to return workspace with email provider
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Mock the WithTransaction call and execute the provided function with nil
		mockRepo.EXPECT().
			WithTransaction(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
				return fn(nil) // Pass nil as the transaction, it won't be used in the test
			})

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcastTx(gomock.Any(), gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No update calls should be made
		mockRepo.EXPECT().UpdateBroadcastTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only broadcasts with draft status can be scheduled")
	})

	t.Run("NoMarketingEmailProvider", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			SendNow:     true,
		}

		// Create a workspace with no marketing email provider set
		workspace := &domain.Workspace{
			ID:        workspaceID,
			Name:      "Test Workspace",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Settings: domain.WorkspaceSettings{
				// No marketing email provider ID set
				MarketingEmailProviderID: "",
			},
			Integrations: []domain.Integration{}, // Empty integrations
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock workspace repository to return workspace with no email provider
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// No transaction should be started
		mockRepo.EXPECT().WithTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no marketing email provider configured")
	})
}

func TestBroadcastService_CancelBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	// Set up logger mock to return itself for chaining
	mockLoggerWithField := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLoggerWithField).AnyTimes()
	mockLoggerWithField.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLoggerWithField.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("CancelScheduledBroadcast", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.CancelBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a scheduled broadcast with future time
		futureTime := time.Now().Add(24 * time.Hour).UTC()
		scheduledDate := futureTime.Format("2006-01-02")
		scheduledTimeStr := futureTime.Format("15:04")

		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusScheduled,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			Schedule: domain.ScheduleSettings{
				IsScheduled:   true,
				ScheduledDate: scheduledDate,
				ScheduledTime: scheduledTimeStr,
				Timezone:      "UTC",
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock the WithTransaction call and execute the provided function with nil
		mockRepo.EXPECT().
			WithTransaction(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
				return fn(nil) // Pass nil as the transaction, it won't be used in the test
			})

		// Mock repository to return the existing broadcast for the transaction
		mockRepo.EXPECT().
			GetBroadcastTx(gomock.Any(), gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcastTx(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *sql.Tx, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, broadcast.ID)
				assert.Equal(t, workspaceID, broadcast.WorkspaceID)
				assert.Equal(t, domain.BroadcastStatusCancelled, broadcast.Status)
				assert.NotNil(t, broadcast.CancelledAt)

				// Schedule settings should remain the same
				assert.Equal(t, existingBroadcast.Schedule.ScheduledDate, broadcast.Schedule.ScheduledDate)
				assert.Equal(t, existingBroadcast.Schedule.ScheduledTime, broadcast.Schedule.ScheduledTime)
				assert.Equal(t, existingBroadcast.Schedule.Timezone, broadcast.Schedule.Timezone)

				return nil
			})

		// In the TestBroadcastService_CancelBroadcast test, find all test cases and add mockEventBus expectation before the service call
		mockEventBus.EXPECT().
			PublishWithAck(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, payload domain.EventPayload, callback domain.EventAckCallback) {
				// Call the callback with success
				callback(nil)
			}).
			AnyTimes()

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.CancelBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a draft broadcast - can't cancel a draft
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock the WithTransaction call and execute the provided function with nil
		mockRepo.EXPECT().
			WithTransaction(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
				return fn(nil) // Pass nil as the transaction, it won't be used in the test
			})

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcastTx(gomock.Any(), gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No update calls expected
		mockRepo.EXPECT().UpdateBroadcastTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only broadcasts with scheduled or paused status can be cancelled")
	})
}

func TestBroadcastService_SendToIndividual(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockBroadcastRepository(ctrl)
		mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockContactRepo := mocks.NewMockContactRepository(ctrl)
		mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockTaskService := mocks.NewMockTaskService(ctrl)
		mockEventBus := mocks.NewMockEventBus(ctrl)
		mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

		// Set up logger mock to return itself for chaining
		mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
		mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
		mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

		// Add direct logger method expectations
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

		apiEndpoint := "https://api.example.com"
		service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, apiEndpoint)

		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		recipientEmail := "test@example.com"
		variationID := "variation123"

		// Create the request
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			RecipientEmail: recipientEmail,
			VariationID:    variationID,
		}

		// Mock auth service to authenticate the user
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Create a broadcast with the test variation
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: "template123",
					},
				},
			},
			UTMParameters: &domain.UTMParameters{
				Source:   "test_source",
				Medium:   "test_medium",
				Campaign: "test_campaign",
				Content:  "test_content",
			},
		}

		// Mock repository to return the broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Mock workspace repository to return a workspace with marketing email provider
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				TransactionalEmailProviderID: "email-provider-123",
				MarketingEmailProviderID:     "email-provider-123",
				EmailTrackingEnabled:         true,
				SecretKey:                    "test-secret-key",
			},
			Integrations: []domain.Integration{
				{
					ID:   "email-provider-123",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSMTP,
						DefaultSenderEmail: "test@example.com",
						DefaultSenderName:  "Test Sender",
					},
				},
			},
		}
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Mock contact repository to return a contact
		contact := &domain.Contact{
			Email: recipientEmail,
			FirstName: &domain.NullableString{
				String: "Test",
				IsNull: false,
			},
			LastName: &domain.NullableString{
				String: "User",
				IsNull: false,
			},
		}
		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), workspaceID, recipientEmail).
			Return(contact, nil)

		// Mock template service to return a template
		emailBlock := getTestEmailBlock()
		template := &domain.Template{
			ID:   "template123",
			Name: "Test Template",
			Email: &domain.EmailTemplate{
				Subject:          "Test Subject",
				FromName:         "Test Sender",
				FromAddress:      "sender@example.com",
				VisualEditorTree: emailBlock,
			},
		}
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, "template123", int64(0)).
			Return(template, nil)

		// Mock template service to compile template
		compiledHTML := "<html><body>Test Content</body></html>"
		compiledResult := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    &compiledHTML,
		}
		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compiledResult, nil)

		// Mock email service to send email
		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				workspaceID,
				true, // isMarketing
				template.Email.FromAddress,
				template.Email.FromName,
				recipientEmail,
				template.Email.Subject,
				*compiledResult.HTML,
				nil,
				"",
				nil,
				nil,
			).
			Return(nil)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("NoMarketingEmailProvider", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockBroadcastRepository(ctrl)
		mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockContactRepo := mocks.NewMockContactRepository(ctrl)
		mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockTaskService := mocks.NewMockTaskService(ctrl)
		mockEventBus := mocks.NewMockEventBus(ctrl)
		mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

		// Set up logger mock to return itself for chaining
		mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
		mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
		mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
		mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

		// Add direct logger method expectations
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

		apiEndpoint := "https://api.example.com"
		service := NewBroadcastService(mockLogger, mockRepo, mockWorkspaceRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, apiEndpoint)

		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		recipientEmail := "test@example.com"
		variationID := "variation123"

		// Create the request
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			RecipientEmail: recipientEmail,
			VariationID:    variationID,
		}

		// Mock auth service to authenticate the user
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Create a workspace with no marketing email provider
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				// No marketing email provider ID set
				TransactionalEmailProviderID: "email-provider-123",
				MarketingEmailProviderID:     "",
			},
			Integrations: []domain.Integration{
				{
					ID:   "email-provider-123",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSMTP,
						DefaultSenderEmail: "test@example.com",
						DefaultSenderName:  "Test Sender",
					},
				},
			},
		}

		// Mock workspace repository to return a workspace without a marketing email provider
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// No calls to repository expected after workspace check
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no marketing email provider configured")
	})
}

func getTestEmailBlock() mjml.EmailBlock {
	return mjml.EmailBlock{
		Kind: "root",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#ffffff",
			},
		},
	}
}
