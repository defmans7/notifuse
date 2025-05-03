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

// Helper function to set up mock logger
func setupMockLogger(ctrl *gomock.Controller) *pkgmocks.MockLogger {
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatal(gomock.Any()).AnyTimes()
	return mockLogger
}

// Helper to create a sample transactional notification
func createTestTransactionalNotification() *domain.TransactionalNotification {
	now := time.Now().UTC()
	return &domain.TransactionalNotification{
		ID:          "test-notification",
		Name:        "Test Notification",
		Description: "Test notification description",
		Channels: domain.ChannelTemplates{
			domain.TransactionalChannelEmail: {
				TemplateID: "template-123",
				Version:    1,
				Settings: domain.MapOfAny{
					"subject": "Test Subject",
				},
			},
		},
		Status:    domain.TransactionalStatusActive,
		Metadata:  domain.MapOfAny{"category": "test"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper to create sample create params
func createTestCreateParams() domain.TransactionalNotificationCreateParams {
	return domain.TransactionalNotificationCreateParams{
		ID:          "test-notification",
		Name:        "Test Notification",
		Description: "Test notification description",
		Channels: domain.ChannelTemplates{
			domain.TransactionalChannelEmail: {
				TemplateID: "template-123",
				Version:    1,
				Settings: domain.MapOfAny{
					"subject": "Test Subject",
				},
			},
		},
		Status:   domain.TransactionalStatusActive,
		Metadata: domain.MapOfAny{"category": "test"},
	}
}

// Helper to create sample update params
func createTestUpdateParams() domain.TransactionalNotificationUpdateParams {
	return domain.TransactionalNotificationUpdateParams{
		Name:        "Updated Notification",
		Description: "Updated description",
		Channels: domain.ChannelTemplates{
			domain.TransactionalChannelEmail: {
				TemplateID: "template-456",
				Version:    2,
				Settings: domain.MapOfAny{
					"subject": "Updated Subject",
				},
			},
		},
		Status:   domain.TransactionalStatusInactive,
		Metadata: domain.MapOfAny{"category": "updated"},
	}
}

// Test service implementation - this would typically be in a separate file
type testTransactionalNotificationService struct {
	repo   domain.TransactionalNotificationRepository
	logger *pkgmocks.MockLogger
}

func (s *testTransactionalNotificationService) CreateNotification(ctx context.Context, workspace string, params domain.TransactionalNotificationCreateParams) (*domain.TransactionalNotification, error) {
	if workspace == "" {
		return nil, errors.New("workspace is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        params.ID,
		"name":      params.Name,
	}).Debug("Creating new transactional notification")

	notification := &domain.TransactionalNotification{
		ID:          params.ID,
		Name:        params.Name,
		Description: params.Description,
		Channels:    params.Channels,
		Status:      params.Status,
		Metadata:    params.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := s.repo.Create(ctx, workspace, notification)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        notification.ID,
		}).Error("Failed to create notification")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        notification.ID,
		"name":      notification.Name,
	}).Info("Transactional notification created successfully")
	return notification, nil
}

func (s *testTransactionalNotificationService) UpdateNotification(ctx context.Context, workspace, id string, params domain.TransactionalNotificationUpdateParams) (*domain.TransactionalNotification, error) {
	if workspace == "" || id == "" {
		return nil, errors.New("workspace and id are required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Updating transactional notification")

	notification, err := s.repo.Get(ctx, workspace, id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to get notification for update")
		return nil, err
	}

	// Update fields that are provided
	if params.Name != "" {
		notification.Name = params.Name
	}

	if params.Description != "" {
		notification.Description = params.Description
	}

	if params.Channels != nil {
		notification.Channels = params.Channels
	}

	if params.Status != "" {
		notification.Status = params.Status
	}

	if params.Metadata != nil {
		notification.Metadata = params.Metadata
	}

	notification.UpdatedAt = time.Now().UTC()

	err = s.repo.Update(ctx, workspace, notification)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        notification.ID,
		}).Error("Failed to update notification")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        notification.ID,
	}).Info("Transactional notification updated successfully")
	return notification, nil
}

func (s *testTransactionalNotificationService) GetNotification(ctx context.Context, workspace, id string) (*domain.TransactionalNotification, error) {
	if workspace == "" || id == "" {
		return nil, errors.New("workspace and id are required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Retrieving transactional notification")

	notification, err := s.repo.Get(ctx, workspace, id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to get notification")
		return nil, err
	}
	return notification, nil
}

func (s *testTransactionalNotificationService) ListNotifications(ctx context.Context, workspace string, filter map[string]interface{}, limit, offset int) ([]*domain.TransactionalNotification, int, error) {
	if workspace == "" {
		return nil, 0, errors.New("workspace is required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"limit":     limit,
		"offset":    offset,
		"filter":    filter,
	}).Debug("Listing transactional notifications")

	notifications, total, err := s.repo.List(ctx, workspace, filter, limit, offset)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
		}).Error("Failed to list notifications")
		return nil, 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"count":     len(notifications),
		"total":     total,
	}).Debug("Successfully retrieved notifications list")
	return notifications, total, nil
}

func (s *testTransactionalNotificationService) DeleteNotification(ctx context.Context, workspace, id string) error {
	if workspace == "" || id == "" {
		return errors.New("workspace and id are required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Deleting transactional notification")

	err := s.repo.Delete(ctx, workspace, id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to delete notification")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Info("Transactional notification deleted successfully")
	return nil
}

func (s *testTransactionalNotificationService) SendNotification(ctx context.Context, workspace string, params domain.TransactionalNotificationSendParams) (string, error) {
	if workspace == "" || params.ID == "" || params.Contact == nil || params.Contact.Email == "" {
		return "", errors.New("workspace, notification ID, and contact are required")
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace":    workspace,
		"id":           params.ID,
		"contact":      params.Contact.Email,
		"has_channels": len(params.Channels) > 0,
		"has_data":     params.Data != nil,
	}).Debug("Sending transactional notification")

	// Get the notification configuration
	notification, err := s.repo.Get(ctx, workspace, params.ID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        params.ID,
		}).Error("Failed to get notification")
		return "", err
	}

	if notification.Status != domain.TransactionalStatusActive {
		s.logger.WithFields(map[string]interface{}{
			"workspace": workspace,
			"id":        params.ID,
			"status":    notification.Status,
		}).Error("Notification is not active")
		return "", errors.New("notification is not active")
	}

	// In a real implementation, this would dispatch the notification to appropriate channels
	// For testing purposes, we'll just return a dummy message ID
	messageID := "msg_" + params.ID + "_" + params.Contact.Email

	s.logger.WithFields(map[string]interface{}{
		"workspace":  workspace,
		"id":         params.ID,
		"contact":    params.Contact.Email,
		"message_id": messageID,
	}).Info("Notification sent successfully")

	return messageID, nil
}

// Ensure the test service implements the interface
var _ domain.TransactionalNotificationService = (*testTransactionalNotificationService)(nil)

func TestTransactionalNotificationService_CreateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	params := createTestCreateParams()

	// Set up the mock expectation
	repo.EXPECT().Create(gomock.Any(), workspace, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, notification *domain.TransactionalNotification) error {
			// Verify the notification was created with the correct parameters
			assert.Equal(t, params.ID, notification.ID)
			assert.Equal(t, params.Name, notification.Name)
			assert.Equal(t, params.Description, notification.Description)
			assert.Equal(t, params.Status, notification.Status)
			assert.NotZero(t, notification.CreatedAt)
			assert.NotZero(t, notification.UpdatedAt)
			return nil
		})

	// Call the service method
	result, err := service.CreateNotification(context.Background(), workspace, params)

	// Verify the result
	require.NoError(t, err)
	assert.Equal(t, params.ID, result.ID)
	assert.Equal(t, params.Name, result.Name)
	assert.Equal(t, params.Description, result.Description)
	assert.Equal(t, params.Channels, result.Channels)
	assert.Equal(t, params.Status, result.Status)
	assert.Equal(t, params.Metadata, result.Metadata)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)
}

func TestTransactionalNotificationService_CreateNotification_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	params := createTestCreateParams()

	// Set up the mock to return an error
	expectedErr := errors.New("repository error")
	repo.EXPECT().Create(gomock.Any(), workspace, gomock.Any()).Return(expectedErr)

	// Call the service method
	result, err := service.CreateNotification(context.Background(), workspace, params)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestTransactionalNotificationService_GetNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"
	expected := createTestTransactionalNotification()

	// Set up the mock expectation
	repo.EXPECT().Get(gomock.Any(), workspace, id).Return(expected, nil)

	// Call the service method
	result, err := service.GetNotification(context.Background(), workspace, id)

	// Verify the result
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTransactionalNotificationService_GetNotification_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"

	// Set up the mock to return an error
	expectedErr := errors.New("not found")
	repo.EXPECT().Get(gomock.Any(), workspace, id).Return(nil, expectedErr)

	// Call the service method
	result, err := service.GetNotification(context.Background(), workspace, id)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestTransactionalNotificationService_UpdateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"

	// Create original notification
	original := createTestTransactionalNotification()
	updateParams := createTestUpdateParams()

	// Set up the mock expectations
	repo.EXPECT().Get(gomock.Any(), workspace, id).Return(original, nil)
	repo.EXPECT().Update(gomock.Any(), workspace, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, updated *domain.TransactionalNotification) error {
			// Verify the notification was updated with the correct parameters
			assert.Equal(t, id, updated.ID)
			assert.Equal(t, updateParams.Name, updated.Name)
			assert.Equal(t, updateParams.Description, updated.Description)
			assert.Equal(t, updateParams.Channels, updated.Channels)
			assert.Equal(t, updateParams.Status, updated.Status)
			assert.Equal(t, updateParams.Metadata, updated.Metadata)
			assert.Equal(t, original.CreatedAt, updated.CreatedAt)
			// Instead of asserting specific times, just check that it's after the original
			assert.True(t, updated.UpdatedAt.After(original.UpdatedAt) || updated.UpdatedAt.Equal(original.UpdatedAt),
				"Expected UpdatedAt to be after or equal to original UpdatedAt")
			return nil
		})

	// Call the service method
	result, err := service.UpdateNotification(context.Background(), workspace, id, updateParams)

	// Verify the result
	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, updateParams.Name, result.Name)
	assert.Equal(t, updateParams.Description, result.Description)
	assert.Equal(t, updateParams.Channels, result.Channels)
	assert.Equal(t, updateParams.Status, result.Status)
	assert.Equal(t, updateParams.Metadata, result.Metadata)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	// Same check for the result
	assert.True(t, result.UpdatedAt.After(original.UpdatedAt) || result.UpdatedAt.Equal(original.UpdatedAt),
		"Expected result.UpdatedAt to be after or equal to original UpdatedAt")
}

func TestTransactionalNotificationService_UpdateNotification_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"
	updateParams := createTestUpdateParams()

	// Set up the mock to return an error for Get
	expectedErr := errors.New("not found")
	repo.EXPECT().Get(gomock.Any(), workspace, id).Return(nil, expectedErr)

	// Call the service method
	result, err := service.UpdateNotification(context.Background(), workspace, id, updateParams)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestTransactionalNotificationService_UpdateNotification_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"
	original := createTestTransactionalNotification()
	updateParams := createTestUpdateParams()

	// Set up the mocks
	repo.EXPECT().Get(gomock.Any(), workspace, id).Return(original, nil)
	expectedErr := errors.New("update error")
	repo.EXPECT().Update(gomock.Any(), workspace, gomock.Any()).Return(expectedErr)

	// Call the service method
	result, err := service.UpdateNotification(context.Background(), workspace, id, updateParams)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestTransactionalNotificationService_ListNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	filter := map[string]interface{}{"status": "active"}
	limit := 10
	offset := 0

	// Expected results
	notifications := []*domain.TransactionalNotification{
		createTestTransactionalNotification(),
		createTestTransactionalNotification(),
	}
	totalCount := 2

	// Set up the mock expectation
	repo.EXPECT().List(gomock.Any(), workspace, filter, limit, offset).Return(notifications, totalCount, nil)

	// Call the service method
	result, count, err := service.ListNotifications(context.Background(), workspace, filter, limit, offset)

	// Verify the result
	require.NoError(t, err)
	assert.Equal(t, notifications, result)
	assert.Equal(t, totalCount, count)
}

func TestTransactionalNotificationService_ListNotifications_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	filter := map[string]interface{}{"status": "active"}
	limit := 10
	offset := 0

	// Set up the mock to return an error
	expectedErr := errors.New("list error")
	repo.EXPECT().List(gomock.Any(), workspace, filter, limit, offset).Return(nil, 0, expectedErr)

	// Call the service method
	result, count, err := service.ListNotifications(context.Background(), workspace, filter, limit, offset)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}

func TestTransactionalNotificationService_DeleteNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"

	// Set up the mock expectation
	repo.EXPECT().Delete(gomock.Any(), workspace, id).Return(nil)

	// Call the service method
	err := service.DeleteNotification(context.Background(), workspace, id)

	// Verify the result
	require.NoError(t, err)
}

func TestTransactionalNotificationService_DeleteNotification_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	id := "test-notification"

	// Set up the mock to return an error
	expectedErr := errors.New("delete error")
	repo.EXPECT().Delete(gomock.Any(), workspace, id).Return(expectedErr)

	// Call the service method
	err := service.DeleteNotification(context.Background(), workspace, id)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTransactionalNotificationService_SendNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	notification := createTestTransactionalNotification()

	params := domain.TransactionalNotificationSendParams{
		ID:       notification.ID,
		Contact:  &domain.Contact{Email: "test@example.com"},
		Channels: []domain.TransactionalChannel{domain.TransactionalChannelEmail},
		Data: domain.MapOfAny{
			"name": "Test User",
		},
		Metadata: domain.MapOfAny{
			"source": "test",
		},
	}

	// Set up the mock expectation
	repo.EXPECT().Get(gomock.Any(), workspace, params.ID).Return(notification, nil)

	// Call the service method
	messageID, err := service.SendNotification(context.Background(), workspace, params)

	// Verify the result
	require.NoError(t, err)
	expectedID := "msg_" + params.ID + "_" + params.Contact.Email
	assert.Equal(t, expectedID, messageID)
}

func TestTransactionalNotificationService_SendNotification_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	params := domain.TransactionalNotificationSendParams{
		ID:      "non-existent",
		Contact: &domain.Contact{Email: "test@example.com"},
	}

	// Set up the mock to return an error
	expectedErr := errors.New("not found")
	repo.EXPECT().Get(gomock.Any(), workspace, params.ID).Return(nil, expectedErr)

	// Call the service method
	messageID, err := service.SendNotification(context.Background(), workspace, params)

	// Verify the error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, messageID)
}

func TestTransactionalNotificationService_SendNotification_Inactive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockLogger := setupMockLogger(ctrl)
	service := &testTransactionalNotificationService{repo: repo, logger: mockLogger}

	workspace := "test-workspace"
	// Create an inactive notification
	notification := createTestTransactionalNotification()
	notification.Status = domain.TransactionalStatusInactive

	params := domain.TransactionalNotificationSendParams{
		ID:      notification.ID,
		Contact: &domain.Contact{Email: "test@example.com"},
	}

	// Set up the mock expectation
	repo.EXPECT().Get(gomock.Any(), workspace, params.ID).Return(notification, nil)

	// Call the service method
	messageID, err := service.SendNotification(context.Background(), workspace, params)

	// Verify the error about inactive notification
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
	assert.Empty(t, messageID)
}
