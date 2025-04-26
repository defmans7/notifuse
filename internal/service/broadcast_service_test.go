package service

import (
	"context"
	"errors"
	"fmt"
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

// Helper function for time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

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

	t.Run("ScheduledTimeSetting", func(t *testing.T) {
		ctx := context.Background()
		scheduledTime := time.Now().UTC().Add(24 * time.Hour)
		request := &domain.CreateBroadcastRequest{
			WorkspaceID: "ws123",
			Name:        "Test Scheduled Broadcast",
			Audience: domain.AudienceSettings{
				Lists:               []string{"list123"},
				ExcludeUnsubscribed: true,
				SkipDuplicateEmails: true,
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled:   true,
				ScheduledTime: scheduledTime,
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
				// Verify scheduled time is set
				assert.NotNil(t, broadcast.ScheduledAt)
				assert.Equal(t, scheduledTime.Unix(), broadcast.ScheduledAt.Unix())
				return nil
			})

		// Call the service
		result, err := service.CreateBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ScheduledAt)
		assert.Equal(t, scheduledTime.Unix(), result.ScheduledAt.Unix())
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

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		expectedErr := errors.New("database error")

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, expectedErr)

		// Call the service
		broadcast, err := service.GetBroadcast(ctx, workspaceID, broadcastID)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, broadcast)
		assert.Equal(t, expectedErr, err)
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
		// Just verify that updated time isn't zero
		assert.False(t, result.UpdatedAt.IsZero())
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

		notFoundErr := &domain.ErrBroadcastNotFound{ID: broadcastID}

		// Mock repository to return not found error
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

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		invalidStatuses := []domain.BroadcastStatus{
			domain.BroadcastStatusSent,
			domain.BroadcastStatusSending,
			domain.BroadcastStatusCancelled,
			domain.BroadcastStatusFailed,
		}

		for _, status := range invalidStatuses {
			// Create an existing broadcast with invalid status
			existingBroadcast := &domain.Broadcast{
				ID:          broadcastID,
				WorkspaceID: workspaceID,
				Name:        "Original Broadcast",
				Status:      status,
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
		}
	})

	t.Run("AllowedStatuses", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		allowedStatuses := []domain.BroadcastStatus{
			domain.BroadcastStatusDraft,
			domain.BroadcastStatusScheduled,
			domain.BroadcastStatusPaused,
		}

		for _, status := range allowedStatuses {
			// Create an existing broadcast with valid status
			existingBroadcast := &domain.Broadcast{
				ID:          broadcastID,
				WorkspaceID: workspaceID,
				Name:        "Original Broadcast",
				Status:      status,
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
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-24 * time.Hour),
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

			// Update should be allowed for these statuses
			mockRepo.EXPECT().
				UpdateBroadcast(gomock.Any(), gomock.Any()).
				Return(nil)

			// Call the service
			result, err := service.UpdateBroadcast(ctx, updateRequest)

			// Verify results
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, updateRequest.Name, result.Name)
			assert.Equal(t, status, result.Status, "Status should not be changed during update")
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		// Create an existing broadcast
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
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
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

		// Mock repository update to return error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		// Call the service
		result, err := service.UpdateBroadcast(ctx, updateRequest)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusDraft
		limit := 10
		offset := 20
		totalCount := 45 // Total number of broadcasts that match the criteria

		// Create sample broadcasts
		broadcasts := []*domain.Broadcast{
			{
				ID:          "bcast1",
				WorkspaceID: workspaceID,
				Name:        "Broadcast 1",
				Status:      status,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "bcast2",
				WorkspaceID: workspaceID,
				Name:        "Broadcast 2",
				Status:      status,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			Limit:       limit,
			Offset:      offset,
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: broadcasts,
			TotalCount: totalCount,
		}

		// Mock repository to return broadcasts
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				assert.Equal(t, workspaceID, p.WorkspaceID)
				assert.Equal(t, status, p.Status)
				assert.Equal(t, limit, p.Limit)
				assert.Equal(t, offset, p.Offset)
				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, len(broadcasts), len(result.Broadcasts))
		assert.Equal(t, broadcasts, result.Broadcasts)
		assert.Equal(t, totalCount, result.TotalCount)
	})

	t.Run("DefaultPaginationValues", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusDraft
		totalCount := 5

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			// Omitting limit and offset to test defaults
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: totalCount,
		}

		// Mock repository to return broadcasts and check default pagination values
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				assert.Equal(t, workspaceID, p.WorkspaceID)
				assert.Equal(t, status, p.Status)
				assert.Equal(t, 50, p.Limit) // Check default limit
				assert.Equal(t, 0, p.Offset) // Check default offset
				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, totalCount, result.TotalCount)
	})

	t.Run("LimitCapReached", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusDraft
		totalCount := 150

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			Limit:       200, // Over the max limit
			Offset:      0,
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: totalCount,
		}

		// Mock repository to verify limit is capped
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				assert.Equal(t, workspaceID, p.WorkspaceID)
				assert.Equal(t, status, p.Status)
				assert.Equal(t, 100, p.Limit) // Check limit was capped
				assert.Equal(t, 0, p.Offset)
				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, totalCount, result.TotalCount)
	})

	t.Run("NegativeOffset", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusDraft
		totalCount := 25

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			Limit:       10,
			Offset:      -5, // Negative offset should be corrected
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: totalCount,
		}

		// Mock repository to verify offset is corrected
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
				assert.Equal(t, workspaceID, p.WorkspaceID)
				assert.Equal(t, status, p.Status)
				assert.Equal(t, 10, p.Limit)
				assert.Equal(t, 0, p.Offset) // Check offset was corrected
				return expectedResponse, nil
			})

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, totalCount, result.TotalCount)
	})

	t.Run("EmptyList", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusSent
		totalCount := 0

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			Limit:       10,
			Offset:      0,
		}

		expectedResponse := &domain.BroadcastListResponse{
			Broadcasts: []*domain.Broadcast{},
			TotalCount: totalCount,
		}

		// Return empty list
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), params).
			Return(expectedResponse, nil)

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.NoError(t, err)
		assert.Empty(t, result.Broadcasts)
		assert.Equal(t, totalCount, result.TotalCount)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		status := domain.BroadcastStatusDraft

		params := domain.ListBroadcastsParams{
			WorkspaceID: workspaceID,
			Status:      status,
			Limit:       10,
			Offset:      0,
		}

		expectedErr := errors.New("database error")

		// Mock repository to return error
		mockRepo.EXPECT().
			ListBroadcasts(gomock.Any(), params).
			Return(nil, expectedErr)

		// Call the service
		result, err := service.ListBroadcasts(ctx, params)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
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
				assert.NotNil(t, broadcast.ScheduledAt)
				assert.Equal(t, scheduledTime.Unix(), broadcast.ScheduledAt.Unix())
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
				assert.Nil(t, broadcast.ScheduledAt)  // Should not be set when sending now
				assert.NotNil(t, broadcast.StartedAt) // Should be set when sending now
				return nil
			})

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		// Create an invalid request (missing scheduled time when not sending now)
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			SendNow:     false,
			// Missing ScheduledAt
		}

		// No repository calls should be made
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scheduled_at is required")
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			ScheduledAt: time.Now().Add(time.Hour),
			SendNow:     false,
		}

		notFoundErr := &domain.ErrBroadcastNotFound{ID: broadcastID}

		// Mock repository to return not found error
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, notFoundErr)

		// No update calls should be made
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, notFoundErr, err)
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

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
			ScheduledAt: time.Now().Add(time.Hour),
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

		// Mock repository update to return error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			Return(expectedErr)

		// Call the service
		err := service.ScheduleBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// TestBroadcastService_CancelBroadcast tests the CancelBroadcast method
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

		// Create a scheduled broadcast
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusScheduled,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			ScheduledAt: timePtr(time.Now().Add(24 * time.Hour)),
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
				assert.Equal(t, existingBroadcast.ScheduledAt, broadcast.ScheduledAt) // Scheduled time should remain
				return nil
			})

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("CancelPausedBroadcast", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.CancelBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a paused broadcast
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusPaused,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			ScheduledAt: timePtr(time.Now().Add(24 * time.Hour)),
			PausedAt:    timePtr(time.Now().Add(-30 * time.Minute)),
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
				assert.Equal(t, existingBroadcast.PausedAt, broadcast.PausedAt) // Paused time should remain
				return nil
			})

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		request := &domain.CancelBroadcastRequest{
			// Missing required fields
		}

		// No repository calls expected
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		request := &domain.CancelBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Mock repository to return not found error
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, &domain.ErrBroadcastNotFound{ID: broadcastID})

		// No update calls expected
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		_, ok := err.(*domain.ErrBroadcastNotFound)
		assert.True(t, ok, "Expected ErrBroadcastNotFound error")
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

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.CancelBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a scheduled broadcast
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusScheduled,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			ScheduledAt: timePtr(time.Now().Add(24 * time.Hour)),
		}

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update to return an error
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		// Call the service
		err := service.CancelBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

// TestBroadcastService_PauseBroadcast tests the PauseBroadcast method
func TestBroadcastService_PauseBroadcast(t *testing.T) {
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

	t.Run("PauseSendingBroadcast", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.PauseBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a sending broadcast
		startedAt := time.Now().Add(-30 * time.Minute).UTC()
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			StartedAt:   &startedAt,
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
				assert.Equal(t, domain.BroadcastStatusPaused, broadcast.Status)
				assert.NotNil(t, broadcast.PausedAt)
				assert.Equal(t, startedAt, *broadcast.StartedAt) // Started time should remain
				assert.True(t, broadcast.UpdatedAt.After(startedAt))
				return nil
			})

		// Call the service
		err := service.PauseBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		request := &domain.PauseBroadcastRequest{
			// Missing required fields
		}

		// No repository calls expected
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.PauseBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "nonexistent"

		request := &domain.PauseBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Mock repository to return not found error
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, &domain.ErrBroadcastNotFound{ID: broadcastID})

		// No update calls expected
		mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.PauseBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		_, ok := err.(*domain.ErrBroadcastNotFound)
		assert.True(t, ok, "Expected ErrBroadcastNotFound error")
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.PauseBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		invalidStatuses := []domain.BroadcastStatus{
			domain.BroadcastStatusDraft,
			domain.BroadcastStatusScheduled,
			domain.BroadcastStatusPaused,
			domain.BroadcastStatusSent,
			domain.BroadcastStatusCancelled,
			domain.BroadcastStatusFailed,
		}

		for _, status := range invalidStatuses {
			// Create a broadcast with invalid status (not sending)
			existingBroadcast := &domain.Broadcast{
				ID:          broadcastID,
				WorkspaceID: workspaceID,
				Name:        "Test Broadcast",
				Status:      status,
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
			err := service.PauseBroadcast(ctx, request)

			// Verify results
			require.Error(t, err)
			assert.Contains(t, err.Error(), "only broadcasts with sending status can be paused")
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.PauseBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a sending broadcast
		startedAt := time.Now().Add(-30 * time.Minute)
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			StartedAt:   &startedAt,
		}

		// Mock repository to return the existing broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository update to return an error
		mockRepo.EXPECT().
			UpdateBroadcast(gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		// Call the service
		err := service.PauseBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
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

	// Set up logger mock to return itself for chaining
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
	mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLoggerWithFields.EXPECT().Debug(gomock.Any()).AnyTimes()

	service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

	variation1 := domain.BroadcastVariation{
		ID:              "var1",
		Name:            "Variation 1",
		TemplateID:      "template1",
		TemplateVersion: 1,
	}

	variation2 := domain.BroadcastVariation{
		ID:              "var2",
		Name:            "Variation 2",
		TemplateID:      "template2",
		TemplateVersion: 1,
	}

	broadcast := &domain.Broadcast{
		ID:          "broadcast1",
		WorkspaceID: "workspace1",
		Name:        "Test Broadcast",
		Status:      domain.BroadcastStatusDraft,
		TestSettings: domain.BroadcastTestSettings{
			Variations: []domain.BroadcastVariation{variation1, variation2},
		},
	}

	t.Run("Success", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast1",
			RecipientEmail: "recipient@example.com",
			VariationID:    "var1",
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast1").
			Return(broadcast, nil)

		// Add missing expectation for GetContactByEmail
		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), "workspace1", "recipient@example.com").
			Return(nil, fmt.Errorf("contact not found"))

		// Add the template service expectations
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), "workspace1", "template1", gomock.Any()).
			Return(&domain.Template{
				Email: &domain.EmailTemplate{
					FromAddress:      "from@example.com",
					FromName:         "Sender Name",
					Subject:          "Test Subject",
					VisualEditorTree: mjml.EmailBlock{},
				},
			}, nil)

		// Add the compile template expectation
		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), "workspace1", gomock.Any(), gomock.Any()).
			Return(&domain.CompileTemplateResponse{
				Success: true,
				HTML: func() *string {
					html := "<html><body>Test</body></html>"
					return &html
				}(),
			}, nil)

		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				"workspace1",
				"marketing",
				"from@example.com",
				"Sender Name",
				"recipient@example.com",
				"Test Subject",
				gomock.Any(),
			).
			Return(nil)

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Info(gomock.Any()).
			AnyTimes()
		mockLogger.EXPECT().
			Debug(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("DefaultVariation", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast1",
			RecipientEmail: "recipient@example.com",
			// No variation ID specified, should use the first one
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast1").
			Return(broadcast, nil)

		// Add missing expectation for GetContactByEmail
		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), "workspace1", "recipient@example.com").
			Return(nil, fmt.Errorf("contact not found"))

		// Add the template service expectations
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), "workspace1", "template1", gomock.Any()).
			Return(&domain.Template{
				Email: &domain.EmailTemplate{
					FromAddress:      "from@example.com",
					FromName:         "Sender Name",
					Subject:          "Test Subject",
					VisualEditorTree: mjml.EmailBlock{},
				},
			}, nil)

		// Add the compile template expectation
		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), "workspace1", gomock.Any(), gomock.Any()).
			Return(&domain.CompileTemplateResponse{
				Success: true,
				HTML: func() *string {
					html := "<html><body>Test</body></html>"
					return &html
				}(),
			}, nil)

		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				"workspace1",
				"marketing",
				"from@example.com",
				"Sender Name",
				"recipient@example.com",
				"Test Subject",
				gomock.Any(),
			).
			Return(nil)

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Info(gomock.Any()).
			AnyTimes()
		mockLogger.EXPECT().
			Debug(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID: "workspace1",
			BroadcastID: "broadcast1",
			// RecipientEmail is missing
			VariationID: "var1",
		}

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Error(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recipient_email is required")
	})

	t.Run("BroadcastNotFound", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "nonexistent",
			RecipientEmail: "recipient@example.com",
			VariationID:    "var1",
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "nonexistent").
			Return(nil, errors.New("broadcast not found"))

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Error(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "broadcast not found")
	})

	t.Run("NoVariations", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast2",
			RecipientEmail: "recipient@example.com",
			// No variation ID specified
		}

		broadcastWithNoVariations := &domain.Broadcast{
			ID:          "broadcast2",
			WorkspaceID: "workspace1",
			Name:        "No Variations Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Variations: []domain.BroadcastVariation{}, // Empty variations
			},
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast2").
			Return(broadcastWithNoVariations, nil)

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Error(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "broadcast has no variations")
	})

	t.Run("VariationNotFound", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast1",
			RecipientEmail: "recipient@example.com",
			VariationID:    "nonexistent",
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast1").
			Return(broadcast, nil)

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Error(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "variation with ID nonexistent not found")
	})

	t.Run("EmailServiceError", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast1",
			RecipientEmail: "recipient@example.com",
			VariationID:    "var1",
		}

		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast1").
			Return(broadcast, nil)

		// Add missing expectation for GetContactByEmail
		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), "workspace1", "recipient@example.com").
			Return(nil, fmt.Errorf("contact not found"))

		// Add the template service expectations
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), "workspace1", "template1", gomock.Any()).
			Return(&domain.Template{
				Email: &domain.EmailTemplate{
					FromAddress:      "from@example.com",
					FromName:         "Sender Name",
					Subject:          "Test Subject",
					VisualEditorTree: mjml.EmailBlock{},
				},
			}, nil)

		// Add the compile template expectation
		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), "workspace1", gomock.Any(), gomock.Any()).
			Return(&domain.CompileTemplateResponse{
				Success: true,
				HTML: func() *string {
					html := "<html><body>Test</body></html>"
					return &html
				}(),
			}, nil)

		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				"workspace1",
				"marketing",
				"from@example.com",
				"Sender Name",
				"recipient@example.com",
				"Test Subject",
				gomock.Any(),
			).
			Return(errors.New("failed to send email"))

		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Error(gomock.Any()).
			AnyTimes()
		mockLogger.EXPECT().
			Debug(gomock.Any()).
			AnyTimes()

		err := service.SendToIndividual(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})

	t.Run("SuccessWithContact", func(t *testing.T) {
		request := &domain.SendToIndividualRequest{
			WorkspaceID:    "workspace1",
			BroadcastID:    "broadcast1",
			RecipientEmail: "contact@example.com",
			VariationID:    "var1",
		}

		// Create a template for the variation
		variation1.TemplateID = "template123"
		variation1.TemplateVersion = 1
		updatedBroadcast := &domain.Broadcast{
			ID:          "broadcast1",
			WorkspaceID: "workspace1",
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			TestSettings: domain.BroadcastTestSettings{
				Variations: []domain.BroadcastVariation{variation1, variation2},
			},
		}

		// Create a mock contact
		now := time.Now().UTC()
		contact := &domain.Contact{
			Email:     "contact@example.com",
			FirstName: &domain.NullableString{String: "John", IsNull: false},
			LastName:  &domain.NullableString{String: "Doe", IsNull: false},
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Create a mock email template
		emailTemplate := &domain.EmailTemplate{
			FromAddress:     "from@example.com",
			FromName:        "Sender Name",
			Subject:         "Test Subject",
			CompiledPreview: "<html><body>Preview</body></html>",
			VisualEditorTree: mjml.EmailBlock{
				ID:   "root",
				Kind: "root",
				Children: []mjml.EmailBlock{
					{
						ID:   "text",
						Kind: "text",
						Data: map[string]interface{}{
							"editorData": []map[string]interface{}{
								{
									"type": "paragraph",
									"children": []map[string]interface{}{
										{
											"text": "Hello {{contact.first_name}}",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Create a mock template
		template := &domain.Template{
			ID:      "template123",
			Version: 1,
			Name:    "Test Template",
			Email:   emailTemplate,
		}

		// Create mock HTML output
		htmlOutput := "<html><body>Hello John</body></html>"
		compiledTemplateResponse := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    &htmlOutput,
		}

		// Set up expectations
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), "workspace1", "broadcast1").
			Return(updatedBroadcast, nil)

		// Expect GetContactByEmail to be called and return a contact
		mockContactRepo.EXPECT().
			GetContactByEmail(gomock.Any(), "workspace1", "contact@example.com").
			Return(contact, nil)

		// Expect GetTemplateByID to be called and return a template
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), "workspace1", "template123", int64(1)).
			Return(template, nil)

		// Expect CompileTemplate to be called with contact data from ToMapOfAny
		mockTemplateSvc.EXPECT().
			CompileTemplate(gomock.Any(), "workspace1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, workspaceID string, tree mjml.EmailBlock, data domain.MapOfAny) (*domain.CompileTemplateResponse, error) {
				// Verify the data passed to CompileTemplate
				contactData, ok := data["contact"].(domain.MapOfAny)
				assert.True(t, ok, "Contact data should be a MapOfAny")
				assert.Equal(t, "contact@example.com", contactData["email"])
				assert.Equal(t, "John", contactData["first_name"])
				assert.Equal(t, "Doe", contactData["last_name"])
				return compiledTemplateResponse, nil
			})

		// Expect SendEmail to be called with the compiled HTML
		mockEmailSvc.EXPECT().
			SendEmail(
				gomock.Any(),
				"workspace1",
				"marketing",
				"from@example.com",
				"Sender Name",
				"contact@example.com",
				"Test Subject",
				htmlOutput,
			).
			Return(nil)

		// Set up logger expectations
		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLogger).
			AnyTimes()
		mockLogger.EXPECT().
			Info(gomock.Any()).
			AnyTimes()
		mockLogger.EXPECT().
			Debug(gomock.Any()).
			AnyTimes()

		// Call the service
		err := service.SendToIndividual(context.Background(), request)
		assert.NoError(t, err)
	})
}

func TestBroadcastService_ResumeBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workspaceID := "ws123"
	broadcastID := "bcast123"

	now := time.Now().UTC()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)

	// Create a template service mock for our tests
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)

	tests := []struct {
		name           string
		request        *domain.ResumeBroadcastRequest
		setupMocks     func(*mocks.MockBroadcastRepository, *pkgmocks.MockLogger)
		expectedError  error
		expectedStatus domain.BroadcastStatus
	}{
		{
			name: "ResumeToScheduled",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

				pausedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusPaused,
					ScheduledAt: &futureTime,
					PausedAt:    nil,
					StartedAt:   nil,
				}

				expectedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusScheduled,
					ScheduledAt: &futureTime,
					PausedAt:    nil,
					StartedAt:   nil,
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(pausedBroadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, broadcast *domain.Broadcast) error {
						assert.Equal(t, expectedBroadcast.Status, broadcast.Status)
						assert.Equal(t, expectedBroadcast.ScheduledAt, broadcast.ScheduledAt)
						assert.Nil(t, broadcast.PausedAt)
						assert.Nil(t, broadcast.StartedAt)
						// Validate UpdatedAt is set and recent
						assert.False(t, broadcast.UpdatedAt.IsZero())
						assert.WithinDuration(t, time.Now(), broadcast.UpdatedAt, 2*time.Second)
						return nil
					})
			},
			expectedError:  nil,
			expectedStatus: domain.BroadcastStatusScheduled,
		},
		{
			name: "ResumeToSending_PastSchedule",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

				pausedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusPaused,
					ScheduledAt: &pastTime,
					PausedAt:    nil,
				}

				expectedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusSending,
					ScheduledAt: &pastTime,
					PausedAt:    nil,
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(pausedBroadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, broadcast *domain.Broadcast) error {
						assert.Equal(t, expectedBroadcast.Status, broadcast.Status)
						assert.Equal(t, expectedBroadcast.ScheduledAt, broadcast.ScheduledAt)
						assert.Nil(t, broadcast.PausedAt)
						assert.NotNil(t, broadcast.StartedAt)
						// Validate UpdatedAt is set and recent
						assert.False(t, broadcast.UpdatedAt.IsZero())
						assert.False(t, broadcast.StartedAt.IsZero())
						assert.WithinDuration(t, time.Now(), broadcast.UpdatedAt, 2*time.Second)
						assert.WithinDuration(t, time.Now(), *broadcast.StartedAt, 2*time.Second)
						return nil
					})
			},
			expectedError:  nil,
			expectedStatus: domain.BroadcastStatusSending,
		},
		{
			name: "ResumeToSending_AlreadyStarted",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

				startTime := pastTime.Add(1 * time.Hour)
				pausedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusPaused,
					ScheduledAt: &pastTime,
					PausedAt:    nil,
					StartedAt:   &startTime,
				}

				expectedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusSending,
					ScheduledAt: &pastTime,
					PausedAt:    nil,
					StartedAt:   &startTime,
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(pausedBroadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, broadcast *domain.Broadcast) error {
						assert.Equal(t, expectedBroadcast.Status, broadcast.Status)
						assert.Equal(t, expectedBroadcast.ScheduledAt, broadcast.ScheduledAt)
						assert.Nil(t, broadcast.PausedAt)
						assert.Equal(t, startTime, *broadcast.StartedAt)
						// Validate UpdatedAt is set and recent
						assert.False(t, broadcast.UpdatedAt.IsZero())
						assert.WithinDuration(t, time.Now(), broadcast.UpdatedAt, 2*time.Second)
						return nil
					})
			},
			expectedError:  nil,
			expectedStatus: domain.BroadcastStatusSending,
		},
		{
			name: "ValidationError",
			request: &domain.ResumeBroadcastRequest{
				// Missing WorkspaceID
				ID: broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
				// No repository calls expected
			},
			expectedError: errors.New("workspace_id is required"),
		},
		{
			name: "BroadcastNotFound",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(nil, errors.New("broadcast not found"))
			},
			expectedError: errors.New("broadcast not found"),
		},
		{
			name: "NonPausedStatus",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

				draftBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusDraft,
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(draftBroadcast, nil)
			},
			expectedError: fmt.Errorf("only broadcasts with paused status can be resumed, current status: %s", domain.BroadcastStatusDraft),
		},
		{
			name: "RepositoryError",
			request: &domain.ResumeBroadcastRequest{
				WorkspaceID: workspaceID,
				ID:          broadcastID,
			},
			setupMocks: func(mockRepo *mocks.MockBroadcastRepository, mockLogger *pkgmocks.MockLogger) {
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

				pausedBroadcast := &domain.Broadcast{
					ID:          broadcastID,
					WorkspaceID: workspaceID,
					Status:      domain.BroadcastStatusPaused,
					ScheduledAt: &futureTime,
					PausedAt:    nil,
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(pausedBroadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			expectedError: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBroadcastRepository(ctrl)
			mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
			mockLogger := pkgmocks.NewMockLogger(ctrl)
			mockContactRepo := mocks.NewMockContactRepository(ctrl)

			service := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockLogger)
			}

			// Call the service
			err := service.ResumeBroadcast(context.Background(), tt.request)

			// Verify results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBroadcastService_SendWinningVariation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
	mockRepo := mocks.NewMockBroadcastRepository(ctrl)
	mockEmailSvc := mocks.NewMockEmailServiceInterface(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTemplateSvc := mocks.NewMockTemplateService(ctrl)

	svc := NewBroadcastService(mockRepo, mockEmailSvc, mockLogger, mockContactRepo, mockTemplateSvc)

	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	variationID := "variation-1"

	tests := []struct {
		name          string
		setupMocks    func()
		request       *domain.SendWinningVariationRequest
		expectedError bool
		errorContains string
	}{
		{
			name: "success - tracking enabled",
			setupMocks: func() {
				broadcast := &domain.Broadcast{
					ID:              broadcastID,
					WorkspaceID:     workspaceID,
					Name:            "Test A/B Broadcast",
					Status:          domain.BroadcastStatusSending,
					TrackingEnabled: true,
					TestSettings: domain.BroadcastTestSettings{
						Enabled:          true,
						SamplePercentage: 50,
						Variations: []domain.BroadcastVariation{
							{
								ID:              variationID,
								Name:            "Variation A",
								TemplateID:      "template-1",
								TemplateVersion: 1,
							},
							{
								ID:              "variation-2",
								Name:            "Variation B",
								TemplateID:      "template-2",
								TemplateVersion: 1,
							},
						},
					},
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(broadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, b *domain.Broadcast) error {
						assert.Equal(t, variationID, b.WinningVariation)
						assert.NotNil(t, b.WinnerSentAt)
						assert.True(t, b.TrackingEnabled)
						return nil
					})

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: true,
			},
			expectedError: false,
		},
		{
			name: "success - tracking not specified but enabled in broadcast",
			setupMocks: func() {
				broadcast := &domain.Broadcast{
					ID:              broadcastID,
					WorkspaceID:     workspaceID,
					Name:            "Test A/B Broadcast",
					Status:          domain.BroadcastStatusSending,
					TrackingEnabled: true,
					TestSettings: domain.BroadcastTestSettings{
						Enabled:          true,
						SamplePercentage: 50,
						Variations: []domain.BroadcastVariation{
							{
								ID:              variationID,
								Name:            "Variation A",
								TemplateID:      "template-1",
								TemplateVersion: 1,
							},
							{
								ID:              "variation-2",
								Name:            "Variation B",
								TemplateID:      "template-2",
								TemplateVersion: 1,
							},
						},
					},
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(broadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, b *domain.Broadcast) error {
						assert.Equal(t, variationID, b.WinningVariation)
						assert.NotNil(t, b.WinnerSentAt)
						assert.True(t, b.TrackingEnabled)
						return nil
					})

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Info(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: false, // Not specified, should use broadcast value
			},
			expectedError: false,
		},
		{
			name: "failure - broadcast not found",
			setupMocks: func() {
				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).
					Return(nil, &domain.ErrBroadcastNotFound{ID: broadcastID})

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: true,
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "failure - A/B testing not enabled",
			setupMocks: func() {
				broadcast := &domain.Broadcast{
					ID:              broadcastID,
					WorkspaceID:     workspaceID,
					Name:            "Test Broadcast",
					Status:          domain.BroadcastStatusSending,
					TrackingEnabled: true,
					TestSettings: domain.BroadcastTestSettings{
						Enabled: false,
					},
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(broadcast, nil)

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: true,
			},
			expectedError: true,
			errorContains: "does not have A/B testing enabled",
		},
		{
			name: "failure - variation not found",
			setupMocks: func() {
				broadcast := &domain.Broadcast{
					ID:              broadcastID,
					WorkspaceID:     workspaceID,
					Name:            "Test A/B Broadcast",
					Status:          domain.BroadcastStatusSending,
					TrackingEnabled: true,
					TestSettings: domain.BroadcastTestSettings{
						Enabled:          true,
						SamplePercentage: 50,
						Variations: []domain.BroadcastVariation{
							{
								ID:              "other-variation",
								Name:            "Other Variation",
								TemplateID:      "template-other",
								TemplateVersion: 1,
							},
						},
					},
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(broadcast, nil)

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: true,
			},
			expectedError: true,
			errorContains: "not found in broadcast",
		},
		{
			name: "failure - update broadcast fails",
			setupMocks: func() {
				broadcast := &domain.Broadcast{
					ID:              broadcastID,
					WorkspaceID:     workspaceID,
					Name:            "Test A/B Broadcast",
					Status:          domain.BroadcastStatusSending,
					TrackingEnabled: true,
					TestSettings: domain.BroadcastTestSettings{
						Enabled:          true,
						SamplePercentage: 50,
						Variations: []domain.BroadcastVariation{
							{
								ID:              variationID,
								Name:            "Variation A",
								TemplateID:      "template-1",
								TemplateVersion: 1,
							},
						},
					},
				}

				mockRepo.EXPECT().GetBroadcast(gomock.Any(), workspaceID, broadcastID).Return(broadcast, nil)
				mockRepo.EXPECT().UpdateBroadcast(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

				// Add logger expectations
				mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLoggerWithFields).AnyTimes()
				mockLoggerWithFields.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			request: &domain.SendWinningVariationRequest{
				WorkspaceID:     workspaceID,
				BroadcastID:     broadcastID,
				VariationID:     variationID,
				TrackingEnabled: true,
			},
			expectedError: true,
			errorContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := svc.SendWinningVariation(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
