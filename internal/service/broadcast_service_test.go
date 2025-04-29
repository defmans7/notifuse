package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
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

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

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
			TrackingEnabled: true,
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
				assert.Equal(t, request.TrackingEnabled, broadcast.TrackingEnabled)
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
			TrackingEnabled: true,
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

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

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

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

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
			TrackingEnabled: true,
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
				assert.Equal(t, updateRequest.TrackingEnabled, broadcast.TrackingEnabled)
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
		assert.Equal(t, updateRequest.TrackingEnabled, result.TrackingEnabled)
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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("ScheduleForLater", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		scheduledTime := time.Now().Add(24 * time.Hour).UTC()

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID:          workspaceID,
			ID:                   broadcastID,
			SendNow:              false,
			ScheduledDate:        scheduledTime.Format("2006-01-02"),
			ScheduledTime:        scheduledTime.Format("15:04"),
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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, broadcast *domain.Broadcast) error {
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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, broadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, broadcast.ID)
				assert.Equal(t, workspaceID, broadcast.WorkspaceID)
				assert.Equal(t, domain.BroadcastStatusSending, broadcast.Status)

				// No scheduled time should be set in Schedule
				assert.False(t, broadcast.Schedule.IsScheduled)

				assert.NotNil(t, broadcast.StartedAt) // Should be set when sending now
				return nil
			})

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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No update calls should be made
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only broadcasts with draft status can be scheduled")
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

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update with verification
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, broadcast *domain.Broadcast) error {
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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No update calls expected
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only broadcasts with scheduled or paused status can be cancelled")
	})
}

func TestBroadcastService_ListBroadcasts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		// Create params for listing broadcasts
		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      domain.BroadcastStatusDraft,
			Limit:       10,
			Offset:      0,
		}

		// Create expected broadcasts
		broadcasts := []*domain.Broadcast{
			{
				ID:          "bcast1",
				WorkspaceID: workspaceID,
				Name:        "Broadcast 1",
				Status:      domain.BroadcastStatusDraft,
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
			{
				ID:          "bcast2",
				WorkspaceID: workspaceID,
				Name:        "Broadcast 2",
				Status:      domain.BroadcastStatusDraft,
				CreatedAt:   time.Now().Add(-12 * time.Hour),
				UpdatedAt:   time.Now().Add(-12 * time.Hour),
			},
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: broadcasts,
			TotalCount: 2,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the expected broadcasts
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				// Verify parameters
				assert.Equal(t, workspaceID, p.WorkspaceID)
				assert.Equal(t, domain.BroadcastStatusDraft, p.Status)
				assert.Equal(t, 10, p.Limit)
				assert.Equal(t, 0, p.Offset)

				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, 2, len(result.Broadcasts))
	})

	t.Run("WithTemplates", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		// Create params with templates option
		params := domain.ListBroadcastsParams{
			WorkspaceID:   workspaceID,
			Limit:         10,
			Offset:        0,
			WithTemplates: true,
		}

		// Create broadcasts with test settings and variations
		templateID := "template123"
		broadcasts := []*domain.Broadcast{
			{
				ID:          "bcast1",
				WorkspaceID: workspaceID,
				Name:        "Broadcast with Template",
				Status:      domain.BroadcastStatusDraft,
				TestSettings: domain.BroadcastTestSettings{
					Enabled: true,
					Variations: []domain.BroadcastVariation{
						{
							ID:         "var1",
							TemplateID: templateID,
						},
					},
				},
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-24 * time.Hour),
			},
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: broadcasts,
			TotalCount: 1,
		}

		// Create template that will be returned
		template := &domain.Template{
			ID:       templateID,
			Name:     "Test Template",
			Version:  1,
			Channel:  "email",
			Category: "marketing",
			Email: &domain.EmailTemplate{
				FromAddress:     "test@example.com",
				FromName:        "Test Sender",
				Subject:         "Test Subject",
				CompiledPreview: "<html>Test content</html>",
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the broadcasts
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			Return(expectedResponse, nil)

		// Mock template service to return the template
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(1)).
			Return(template, nil)

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		assert.Equal(t, 1, len(result.Broadcasts))
		assert.Equal(t, templateID, result.Broadcasts[0].TestSettings.Variations[0].TemplateID)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Limit:       10,
			Offset:      0,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return an error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr)

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Same(t, expectedErr, err)
	})

	t.Run("DefaultLimitAndOffset", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		// Create params with zero limit and negative offset
		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Limit:       0,  // Should default to 50
			Offset:      -5, // Should default to 0
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: 0,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to ensure it receives the default values
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				// Verify default parameters were applied
				assert.Equal(t, 50, p.Limit) // Default limit is 50
				assert.Equal(t, 0, p.Offset) // Default offset is 0

				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("MaxLimitEnforced", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		// Create params with limit exceeding maximum (100)
		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Limit:       200, // Should be capped at 100
			Offset:      0,
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: 0,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to ensure it receives the capped limit
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				// Verify limit was capped
				assert.Equal(t, 100, p.Limit) // Maximum limit is 100

				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("TemplateError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"

		// Create params with templates option
		params := domain.ListBroadcastsParams{
			WorkspaceID:   workspaceID,
			Limit:         10,
			Offset:        0,
			WithTemplates: true,
		}

		// Create broadcasts with test settings and variations
		templateID := "template123"
		broadcasts := []*domain.Broadcast{
			{
				ID:          "bcast1",
				WorkspaceID: workspaceID,
				Name:        "Broadcast with Template",
				Status:      domain.BroadcastStatusDraft,
				TestSettings: domain.BroadcastTestSettings{
					Enabled: true,
					Variations: []domain.BroadcastVariation{
						{
							ID:         "var1",
							TemplateID: templateID,
						},
					},
				},
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-24 * time.Hour),
			},
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: broadcasts,
			TotalCount: 1,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the broadcasts
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			Return(expectedResponse, nil)

		// Mock template service to return an error
		templateErr := errors.New("template not found")
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(1)).
			Return(nil, templateErr)

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results - service should continue despite template error
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalCount)
		// The broadcast should still be returned even if template fetch failed
		assert.Equal(t, 1, len(result.Broadcasts))
	})
}

func TestBroadcastService_DeleteBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)

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

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a broadcast that can be deleted (with draft status)
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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository to delete the broadcast
		mockRepo.EXPECT().
			DeleteBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()

		// Create invalid request with missing fields
		request := &domain.DeleteBroadcastRequest{
			// Missing WorkspaceID
			ID: "bcast123",
		}

		// Set up logger mock for this test case
		mockLogger.EXPECT().
			WithField("broadcast_id", request.ID).
			Return(mockLoggerWithFields).
			AnyTimes()

		// Authentication will be called with an empty workspace ID, which should fail
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), "").
			Return(nil, nil, errors.New("workspace ID is required")).Times(1)

		// No repository calls expected
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().DeleteBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authenticate user")
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return not found error
		notFoundErr := &domain.ErrBroadcastNotFound{ID: broadcastID}
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// No delete call expected
		mockRepo.EXPECT().DeleteBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, notFoundErr, err)
	})

	t.Run("SendingStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a broadcast with 'sending' status which cannot be deleted
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No delete call expected
		mockRepo.EXPECT().DeleteBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "broadcasts in 'sending' status cannot be deleted")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a broadcast that can be deleted (with draft status)
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

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository to return an error on delete
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			DeleteBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(expectedErr)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, expectedErr, err)
	})
}

func TestBroadcastService_SendToIndividual(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)

	// Setup logger mock
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		templateID := "template123"
		recipientEmail := "user@example.com"

		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			VariationID:    variationID,
			RecipientEmail: recipientEmail,
		}

		// Create a broadcast with test variation
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: templateID,
					},
				},
			},
		}

		// Create template
		template := &domain.Template{
			ID:       templateID,
			Name:     "Test Template",
			Version:  1,
			Channel:  "email",
			Category: "marketing",
			Email: &domain.EmailTemplate{
				FromAddress: "sender@example.com",
				FromName:    "Sender Name",
				Subject:     "Test Subject",
			},
		}

		// Create contact
		contact := &domain.Contact{
			Email:     recipientEmail,
			FirstName: &domain.NullableString{String: "Test", IsNull: false},
			LastName:  &domain.NullableString{String: "User", IsNull: false},
		}

		// Create compiled template result
		compiledHTML := "<html><body>Hello Test!</body></html>"
		compiledTemplate := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    &compiledHTML,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), workspaceID, recipientEmail).
			Return(contact, nil)

		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(1)).
			Return(template, nil)

		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
			Return(compiledTemplate, nil)

		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				workspaceID,
				"marketing",
				template.Email.FromAddress,
				template.Email.FromName,
				recipientEmail,
				template.Email.Subject,
				compiledHTML,
			).
			Return(nil)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"
		recipientEmail := "user@example.com"

		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			RecipientEmail: recipientEmail,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return not found error
		notFoundErr := errors.New("broadcast not found")
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, notFoundErr, err)
	})

	t.Run("NoVariationSpecified", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		templateID := "template123"
		recipientEmail := "user@example.com"

		// No variation ID specified in request
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			RecipientEmail: recipientEmail,
		}

		// Create a broadcast with variation
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: templateID,
					},
				},
			},
		}

		// Create template
		template := &domain.Template{
			ID:       templateID,
			Name:     "Test Template",
			Version:  1,
			Channel:  "email",
			Category: "marketing",
			Email: &domain.EmailTemplate{
				FromAddress: "sender@example.com",
				FromName:    "Sender Name",
				Subject:     "Test Subject",
			},
		}

		// Create compiled template result
		compiledHTML := "<html><body>Hello!</body></html>"
		compiledTemplate := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    &compiledHTML,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations - should use first variation
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), workspaceID, recipientEmail).
			Return(nil, errors.New("contact not found"))

		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(1)).
			Return(template, nil)

		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
			Return(compiledTemplate, nil)

		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				workspaceID,
				"marketing",
				template.Email.FromAddress,
				template.Email.FromName,
				recipientEmail,
				template.Email.Subject,
				compiledHTML,
			).
			Return(nil)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("NoVariations", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		recipientEmail := "user@example.com"

		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			RecipientEmail: recipientEmail,
		}

		// Create a broadcast with NO variations
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Enabled:    true,
				Variations: []domain.BroadcastVariation{}, // Empty variations
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "broadcast has no variations")
	})

	t.Run("VariationNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		nonExistentVariationID := "var999"
		templateID := "template123"
		recipientEmail := "user@example.com"

		request := &domain.SendToIndividualRequest{
			WorkspaceID:    workspaceID,
			BroadcastID:    broadcastID,
			VariationID:    nonExistentVariationID, // Non-existent variation
			RecipientEmail: recipientEmail,
		}

		// Create a broadcast with different variation
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID, // Different ID
						TemplateID: templateID,
					},
				},
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Call the service
		err := service.SendToIndividual(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variation with ID var999 not found in broadcast")
	})
}

func TestBroadcastService_SendWinningVariation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)

	// Setup logger mock
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Add direct logger method expectations
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	service, err := NewBroadcastService(BroadcastServiceConfig{
		Logger:      mockLogger,
		AuthService: mockAuthSvc,
	})
	require.NoError(t, err)

	// Then manually set the repositories needed for testing
	service.repo = mockRepo
	service.emailSvc = mockEmailSvc
	service.contactRepo = mockContactRepo
	service.templateSvc = mockTemplateSvc

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		templateID := "template123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     variationID,
			TrackingEnabled: true,
		}

		// Create a broadcast with A/B testing enabled
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: templateID,
					},
				},
			},
			TrackingEnabled: false, // Different from request
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Should update broadcast with winning variation
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, updatedBroadcast *domain.Broadcast) error {
				// Verify important properties
				assert.Equal(t, broadcastID, updatedBroadcast.ID)
				assert.Equal(t, workspaceID, updatedBroadcast.WorkspaceID)
				assert.Equal(t, variationID, updatedBroadcast.WinningVariation)
				assert.True(t, updatedBroadcast.TrackingEnabled)
				assert.NotNil(t, updatedBroadcast.WinnerSentAt)
				return nil
			})

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"
		variationID := "var123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     variationID,
			TrackingEnabled: true,
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock repository to return not found error
		notFoundErr := errors.New("broadcast not found")
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, notFoundErr, err)
	})

	t.Run("ABTestingNotEnabled", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     variationID,
			TrackingEnabled: true,
		}

		// Create a broadcast WITHOUT A/B testing enabled
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: false, // A/B testing disabled
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "broadcast does not have A/B testing enabled")
	})

	t.Run("VariationNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		nonExistentVariationID := "var999"
		templateID := "template123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     nonExistentVariationID, // Non-existent variation
			TrackingEnabled: true,
		}

		// Create a broadcast with different variation
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID, // Different ID
						TemplateID: templateID,
					},
				},
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variation with ID var999 not found in broadcast")
	})

	t.Run("UpdateError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		templateID := "template123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     variationID,
			TrackingEnabled: true,
		}

		// Create a broadcast with A/B testing enabled
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: templateID,
					},
				},
			},
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Mock repository to return an error on update
		updateErr := errors.New("database error")
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			Return(updateErr)

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Same(t, updateErr, err)
	})

	t.Run("UseExistingTrackingSettings", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		variationID := "var123"
		templateID := "template123"

		request := &domain.SendWinningVariationRequest{
			WorkspaceID:     workspaceID,
			BroadcastID:     broadcastID,
			VariationID:     variationID,
			TrackingEnabled: false, // Not enabling tracking in request
		}

		// Create a broadcast with tracking enabled
		broadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			TestSettings: domain.BroadcastTestSettings{
				Enabled: true,
				Variations: []domain.BroadcastVariation{
					{
						ID:         variationID,
						TemplateID: templateID,
					},
				},
			},
			TrackingEnabled: true, // Enabled in broadcast
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(broadcast, nil)

		// Should update broadcast with winning variation and preserve tracking settings
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, updatedBroadcast *domain.Broadcast) error {
				// Verify tracking settings preserved from original broadcast
				assert.True(t, updatedBroadcast.TrackingEnabled)
				assert.Equal(t, variationID, updatedBroadcast.WinningVariation)
				assert.NotNil(t, updatedBroadcast.WinnerSentAt)
				return nil
			})

		// Call the service
		err := service.SendWinningVariation(ctx, request)

		// Verify results
		require.NoError(t, err)
	})
}
