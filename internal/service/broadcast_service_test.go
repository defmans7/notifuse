package service

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
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
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)

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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

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
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

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
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)
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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

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

func TestBroadcastService_RecordMessageSent(t *testing.T) {
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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Test when repository supports CreateMessageHistory
	t.Run("RepositorySupportsMessageHistory", func(t *testing.T) {
		// Create a mock repository that implements CreateMessageHistory
		mockMessageHistoryRepo := &mockRepositoryWithMessageHistory{
			MockBroadcastRepository: mockRepo,
			ctrl:                    ctrl,
		}

		// Set the repository in the service
		service.repo = mockMessageHistoryRepo

		ctx := context.Background()
		workspaceID := "ws123"
		message := &domain.MessageHistory{
			ID:              "msg123",
			ContactID:       "contact123",
			BroadcastID:     stringPtr("bcast123"),
			TemplateID:      "template123",
			TemplateVersion: 1,
			Channel:         "email",
			Status:          domain.MessageStatusSent,
			SentAt:          time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Expect CreateMessageHistory to be called
		mockMessageHistoryRepo.EXPECT().
			CreateMessageHistory(gomock.Any(), gomock.Eq(workspaceID), gomock.Any()).
			DoAndReturn(func(_ context.Context, wsID string, msg *domain.MessageHistory) error {
				assert.Equal(t, workspaceID, wsID)
				assert.Equal(t, message.ID, msg.ID)
				assert.Equal(t, message.ContactID, msg.ContactID)
				assert.Equal(t, message.Channel, msg.Channel)
				assert.Equal(t, message.Status, msg.Status)
				return nil
			})

		// Call the service
		err := service.RecordMessageSent(ctx, workspaceID, message)

		// Verify results
		require.NoError(t, err)
	})

	// Test when repository does not support CreateMessageHistory
	t.Run("RepositoryDoesNotSupportMessageHistory", func(t *testing.T) {
		// Set the repository back to the standard mock
		service.repo = mockRepo

		ctx := context.Background()
		workspaceID := "ws123"
		message := &domain.MessageHistory{
			ID:              "msg123",
			ContactID:       "contact123",
			BroadcastID:     stringPtr("bcast123"),
			TemplateID:      "template123",
			TemplateVersion: 1,
			Channel:         "email",
			Status:          domain.MessageStatusSent,
			SentAt:          time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Call the service
		err := service.RecordMessageSent(ctx, workspaceID, message)

		// Verify results - should not error even though the repository doesn't support message history
		require.NoError(t, err)
	})

	// Test when repository returns an error
	t.Run("RepositoryReturnsError", func(t *testing.T) {
		// Create a mock repository that implements CreateMessageHistory and returns an error
		mockMessageHistoryRepo := &mockRepositoryWithMessageHistory{
			MockBroadcastRepository: mockRepo,
			ctrl:                    ctrl,
		}

		// Set the repository in the service
		service.repo = mockMessageHistoryRepo

		ctx := context.Background()
		workspaceID := "ws123"
		message := &domain.MessageHistory{
			ID:              "msg123",
			ContactID:       "contact123",
			BroadcastID:     stringPtr("bcast123"),
			TemplateID:      "template123",
			TemplateVersion: 1,
			Channel:         "email",
			Status:          domain.MessageStatusSent,
			SentAt:          time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Expect CreateMessageHistory to be called and return an error
		expectedErr := errors.New("database error")
		mockMessageHistoryRepo.EXPECT().
			CreateMessageHistory(gomock.Any(), gomock.Eq(workspaceID), gomock.Any()).
			Return(expectedErr)

		// Call the service
		err := service.RecordMessageSent(ctx, workspaceID, message)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestBroadcastService_UpdateMessageStatus(t *testing.T) {
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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Test when repository supports UpdateMessageStatus
	t.Run("RepositorySupportsMessageHistoryUpdate", func(t *testing.T) {
		// Create a mock repository that implements UpdateMessageStatus
		mockMessageHistoryRepo := &mockRepositoryWithMessageHistory{
			MockBroadcastRepository: mockRepo,
			ctrl:                    ctrl,
		}

		// Set the repository in the service
		service.repo = mockMessageHistoryRepo

		ctx := context.Background()
		workspaceID := "ws123"
		messageID := "msg123"
		status := domain.MessageStatusDelivered
		timestamp := time.Now()

		// Expect UpdateMessageStatus to be called
		mockMessageHistoryRepo.EXPECT().
			UpdateMessageStatus(gomock.Any(), gomock.Eq(workspaceID), gomock.Eq(messageID), gomock.Eq(status), gomock.Any()).
			DoAndReturn(func(_ context.Context, wsID string, msgID string, sts domain.MessageStatus, ts time.Time) error {
				assert.Equal(t, workspaceID, wsID)
				assert.Equal(t, messageID, msgID)
				assert.Equal(t, status, sts)
				assert.WithinDuration(t, timestamp, ts, time.Second)
				return nil
			})

		// Call the service
		err := service.UpdateMessageStatus(ctx, workspaceID, messageID, status, timestamp)

		// Verify results
		require.NoError(t, err)
	})

	// Test when repository does not support UpdateMessageStatus
	t.Run("RepositoryDoesNotSupportMessageHistoryUpdate", func(t *testing.T) {
		// Set the repository back to the standard mock
		service.repo = mockRepo

		ctx := context.Background()
		workspaceID := "ws123"
		messageID := "msg123"
		status := domain.MessageStatusDelivered
		timestamp := time.Now()

		// Call the service
		err := service.UpdateMessageStatus(ctx, workspaceID, messageID, status, timestamp)

		// Verify results - should not error even though the repository doesn't support message history
		require.NoError(t, err)
	})

	// Test when repository returns an error
	t.Run("RepositoryReturnsError", func(t *testing.T) {
		// Create a mock repository that implements UpdateMessageStatus and returns an error
		mockMessageHistoryRepo := &mockRepositoryWithMessageHistory{
			MockBroadcastRepository: mockRepo,
			ctrl:                    ctrl,
		}

		// Set the repository in the service
		service.repo = mockMessageHistoryRepo

		ctx := context.Background()
		workspaceID := "ws123"
		messageID := "msg123"
		status := domain.MessageStatusDelivered
		timestamp := time.Now()

		// Expect UpdateMessageStatus to be called and return an error
		expectedErr := errors.New("database error")
		mockMessageHistoryRepo.EXPECT().
			UpdateMessageStatus(gomock.Any(), gomock.Eq(workspaceID), gomock.Eq(messageID), gomock.Eq(status), gomock.Any()).
			Return(expectedErr)

		// Call the service
		err := service.UpdateMessageStatus(ctx, workspaceID, messageID, status, timestamp)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestBroadcastService_GetAPIEndpoint(t *testing.T) {
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

	expectedEndpoint := "https://api.example.com"

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, expectedEndpoint)

	// Test GetAPIEndpoint
	t.Run("ReturnsConfiguredEndpoint", func(t *testing.T) {
		endpoint := service.GetAPIEndpoint()
		assert.Equal(t, expectedEndpoint, endpoint)
	})
}

func TestBroadcastService_GetTemplateByID(t *testing.T) {
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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		templateID := "template123"

		expectedTemplate := &domain.Template{
			ID:      templateID,
			Name:    "Test Template",
			Version: 1,
			Email: &domain.EmailTemplate{
				Subject:     "Test Subject",
				FromName:    "Test Sender",
				FromAddress: "test@example.com",
			},
		}

		// Expect GetTemplateByID to be called on the template service
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), gomock.Eq(workspaceID), gomock.Eq(templateID), gomock.Eq(int64(1))).
			Return(expectedTemplate, nil)

		// Call the service
		template, err := service.GetTemplateByID(ctx, workspaceID, templateID)

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, expectedTemplate, template)
	})

	t.Run("TemplateNotFound", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		templateID := "nonexistent"

		expectedErr := errors.New("template not found")

		// Expect GetTemplateByID to be called on the template service and return an error
		mockTemplateSvc.EXPECT().
			GetTemplateByID(gomock.Any(), gomock.Eq(workspaceID), gomock.Eq(templateID), gomock.Eq(int64(1))).
			Return(nil, expectedErr)

		// Call the service
		template, err := service.GetTemplateByID(ctx, workspaceID, templateID)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, template)
		assert.Equal(t, expectedErr, err)
	})
}

func TestBroadcastService_SetTaskService(t *testing.T) {
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

	// Create a new task service to set
	newTaskService := mocks.NewMockTaskService(ctrl)

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	// Verify that the task service is initially what we set
	assert.Equal(t, mockTaskService, service.taskService)

	// Set the new task service
	service.SetTaskService(newTaskService)

	// Verify that the task service has been updated
	assert.Equal(t, newTaskService, service.taskService)
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
	mockTaskService := mocks.NewMockTaskService(ctrl)
	mockEventBus := mocks.NewMockEventBus(ctrl)

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

	service := NewBroadcastService(mockLogger, mockRepo, mockEmailSvc, mockContactRepo, mockTemplateSvc, mockTaskService, mockAuthSvc, mockEventBus, "https://api.example.com")

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a broadcast with an allowed status (not 'sending')
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock getting the broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock the delete operation
		mockRepo.EXPECT().
			DeleteBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.NoError(t, err)
	})

	t.Run("AuthError", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Mock authentication failure
		expectedErr := errors.New("authentication failed")
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, nil, expectedErr)

		// No other calls should be made
		mockRepo.EXPECT().GetBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().DeleteBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		ctx := context.Background()
		workspaceID := "ws123"
		broadcastID := "bcast123"

		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: workspaceID,
			ID:          broadcastID,
		}

		// Create a broadcast with an invalid status (sending)
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusSending,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock getting the broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// No delete should be called
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

		// Create a broadcast with an allowed status
		existingBroadcast := &domain.Broadcast{
			ID:          broadcastID,
			WorkspaceID: workspaceID,
			Name:        "Test Broadcast",
			Status:      domain.BroadcastStatusDraft,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}

		// Mock authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user123"}, nil)

		// Mock getting the broadcast
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(existingBroadcast, nil)

		// Mock repository error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			DeleteBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(expectedErr)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
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

		// Mock broadcast not found
		expectedErr := errors.New("broadcast not found")
		mockRepo.EXPECT().
			GetBroadcast(gomock.Any(), workspaceID, broadcastID).
			Return(nil, expectedErr)

		// No delete should be called
		mockRepo.EXPECT().DeleteBroadcast(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// Call the service
		err := service.DeleteBroadcast(ctx, request)

		// Verify results
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// Helper types for testing
type mockRepositoryWithMessageHistory struct {
	*mocks.MockBroadcastRepository
	ctrl *gomock.Controller
}

func (m *mockRepositoryWithMessageHistory) CreateMessageHistory(ctx context.Context, workspaceID string, message *domain.MessageHistory) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMessageHistory", ctx, workspaceID, message)
	ret0, _ := ret[0].(error)
	return ret0
}

func (m *mockRepositoryWithMessageHistory) EXPECT() *mockRepositoryWithMessageHistoryRecorder {
	return &mockRepositoryWithMessageHistoryRecorder{m}
}

type mockRepositoryWithMessageHistoryRecorder struct {
	mock *mockRepositoryWithMessageHistory
}

func (mr *mockRepositoryWithMessageHistoryRecorder) CreateMessageHistory(ctx, workspaceID, message interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"CreateMessageHistory",
		reflect.TypeOf((*mockRepositoryWithMessageHistory)(nil).CreateMessageHistory),
		ctx, workspaceID, message,
	)
}

func (m *mockRepositoryWithMessageHistory) UpdateMessageStatus(ctx context.Context, workspaceID string, messageID string, status domain.MessageStatus, timestamp time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMessageStatus", ctx, workspaceID, messageID, status, timestamp)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *mockRepositoryWithMessageHistoryRecorder) UpdateMessageStatus(ctx, workspaceID, messageID, status, timestamp interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"UpdateMessageStatus",
		reflect.TypeOf((*mockRepositoryWithMessageHistory)(nil).UpdateMessageStatus),
		ctx, workspaceID, messageID, status, timestamp,
	)
}

// Helper function to get a string pointer
func stringPtr(s string) *string {
	return &s
}
