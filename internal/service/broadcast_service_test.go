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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

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

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		// Create an invalid request (missing required fields)
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			// Missing Name and other required fields
		}

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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

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

		notFoundErr := errors.New("broadcast not found")
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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

	t.Run("ScheduleForLater", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"
		scheduledTime := time.Now().Add(24 * time.Hour).UTC()

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			ScheduledAt: scheduledTime,
			SendNow:     false,
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
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			ScheduledAt: time.Now().Add(time.Hour),
			SendNow:     false,
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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

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
