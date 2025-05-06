package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockEmailService is a simple mock for EmailService used in tests
type mockEmailService struct {
	logger logger.Logger
}

func (m *mockEmailService) SendEmail(ctx context.Context, workspaceID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, optionalProvider ...*domain.EmailProvider) error {
	return nil
}

func (m *mockEmailService) TestEmailProvider(ctx context.Context, workspaceID string, provider domain.EmailProvider, to string) error {
	return nil
}

func (m *mockEmailService) TestTemplate(ctx context.Context, workspaceID string, templateID string, integrationID string, recipientEmail string) error {
	return nil
}

func TestTransactionalNotificationService_CreateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name           string
		input          domain.TransactionalNotificationCreateParams
		mockSetup      func()
		expectedError  bool
		expectedResult *domain.TransactionalNotification
	}

	ctx := context.Background()
	workspace := "test-workspace"
	templateID := uuid.New().String()

	tests := []testCase{
		{
			name: "Success_CreateNotification",
			input: domain.TransactionalNotificationCreateParams{
				ID:          uuid.New().String(),
				Name:        "Test Notification",
				Description: "This is a test notification",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: templateID,
					},
				},
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
			mockSetup: func() {
				// Expect template service to validate the template exists
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{ID: templateID}, nil)

				// Expect repo to create notification
				mockRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, notif *domain.TransactionalNotification) error {
						assert.Equal(t, "Test Notification", notif.Name)
						return nil
					})
			},
			expectedError: false,
			expectedResult: &domain.TransactionalNotification{
				ID:          gomock.Any().String(),
				Name:        "Test Notification",
				Description: "This is a test notification",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: templateID,
					},
				},
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name: "Error_TemplateNotFound",
			input: domain.TransactionalNotificationCreateParams{
				ID:   uuid.New().String(),
				Name: "Test Notification",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: templateID,
					},
				},
			},
			mockSetup: func() {
				// Expect template service to fail finding the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(nil, errors.New("template not found"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error_RepositoryCreateFailed",
			input: domain.TransactionalNotificationCreateParams{
				ID:   uuid.New().String(),
				Name: "Test Notification",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: templateID,
					},
				},
			},
			mockSetup: func() {
				// Template exists
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{ID: templateID}, nil)

				// But repo create fails
				mockRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					Return(errors.New("repository error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks for this test case
			tc.mockSetup()

			// Create service with mocked dependencies
			service := &TransactionalNotificationService{
				transactionalRepo:  mockRepo,
				messageHistoryRepo: mockMsgHistoryRepo,
				templateService:    mockTemplateService,
				contactService:     mockContactService,
				emailService:       nil, // Not used in this test
				logger:             mockLogger,
			}

			// Call the method being tested
			result, err := service.CreateNotification(ctx, workspace, tc.input)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.input.Name, result.Name)
				assert.Equal(t, tc.input.Description, result.Description)
				assert.Equal(t, tc.input.Channels, result.Channels)
				assert.Equal(t, tc.input.Metadata, result.Metadata)
			}
		})
	}
}

func TestTransactionalNotificationService_UpdateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name           string
		id             string
		input          domain.TransactionalNotificationUpdateParams
		mockSetup      func()
		expectedError  bool
		expectedResult *domain.TransactionalNotification
	}

	ctx := context.Background()
	workspace := "test-workspace"
	notificationID := uuid.New().String()
	templateID := uuid.New().String()
	newTemplateID := uuid.New().String()

	existingNotification := &domain.TransactionalNotification{
		ID:          notificationID,
		Name:        "Original Name",
		Description: "Original Description",
		Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
			domain.TransactionalChannelEmail: {
				TemplateID: templateID,
			},
		},
		Metadata: map[string]interface{}{
			"original": "value",
		},
	}

	tests := []testCase{
		{
			name: "Success_UpdateName",
			id:   notificationID,
			input: domain.TransactionalNotificationUpdateParams{
				Name: "Updated Name",
			},
			mockSetup: func() {
				// Get existing notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(existingNotification, nil)

				// Update notification
				mockRepo.EXPECT().
					Update(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, notif *domain.TransactionalNotification) error {
						assert.Equal(t, "Updated Name", notif.Name)
						assert.Equal(t, existingNotification.Description, notif.Description)
						assert.Equal(t, existingNotification.Channels, notif.Channels)
						assert.Equal(t, existingNotification.Metadata, notif.Metadata)
						return nil
					})
			},
			expectedError: false,
			expectedResult: &domain.TransactionalNotification{
				ID:          notificationID,
				Name:        "Updated Name",
				Description: "Original Description",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: templateID,
					},
				},
				Metadata: map[string]interface{}{
					"original": "value",
				},
			},
		},
		{
			name: "Success_UpdateAllFields",
			id:   notificationID,
			input: domain.TransactionalNotificationUpdateParams{
				Name:        "Completely Updated",
				Description: "New Description",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: newTemplateID,
					},
				},
				Metadata: map[string]interface{}{
					"new": "metadata",
				},
			},
			mockSetup: func() {
				// Get existing notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(existingNotification, nil)

				// Expect template service to validate the template exists
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, newTemplateID, int64(0)).
					Return(&domain.Template{ID: newTemplateID}, nil)

				// Update notification
				mockRepo.EXPECT().
					Update(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, notif *domain.TransactionalNotification) error {
						assert.Equal(t, "Completely Updated", notif.Name)
						assert.Equal(t, "New Description", notif.Description)
						assert.Equal(t, newTemplateID, notif.Channels[domain.TransactionalChannelEmail].TemplateID)
						assert.Equal(t, "metadata", notif.Metadata["new"])
						return nil
					})
			},
			expectedError: false,
			expectedResult: &domain.TransactionalNotification{
				ID:          notificationID,
				Name:        "Completely Updated",
				Description: "New Description",
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: newTemplateID,
					},
				},
				Metadata: map[string]interface{}{
					"new": "metadata",
				},
			},
		},
		{
			name: "Error_NotificationNotFound",
			id:   notificationID,
			input: domain.TransactionalNotificationUpdateParams{
				Name: "Updated Name",
			},
			mockSetup: func() {
				// Get existing notification fails
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(nil, errors.New("notification not found"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error_TemplateNotFound",
			id:   notificationID,
			input: domain.TransactionalNotificationUpdateParams{
				Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
					domain.TransactionalChannelEmail: {
						TemplateID: newTemplateID,
					},
				},
			},
			mockSetup: func() {
				// Get existing notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(existingNotification, nil)

				// Template validation fails
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, newTemplateID, int64(0)).
					Return(nil, errors.New("template not found"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error_UpdateFailed",
			id:   notificationID,
			input: domain.TransactionalNotificationUpdateParams{
				Name: "Updated Name",
			},
			mockSetup: func() {
				// Get existing notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(existingNotification, nil)

				// Update notification fails
				mockRepo.EXPECT().
					Update(gomock.Any(), workspace, gomock.Any()).
					Return(errors.New("update failed"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks for this test case
			tc.mockSetup()

			// Create service with mocked dependencies
			service := &TransactionalNotificationService{
				transactionalRepo:  mockRepo,
				messageHistoryRepo: mockMsgHistoryRepo,
				templateService:    mockTemplateService,
				contactService:     mockContactService,
				emailService:       nil, // Not used in this test
				logger:             mockLogger,
			}

			// Call the method being tested
			result, err := service.UpdateNotification(ctx, workspace, tc.id, tc.input)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.input.Name != "" {
					assert.Equal(t, tc.input.Name, result.Name)
				}
				if tc.input.Description != "" {
					assert.Equal(t, tc.input.Description, result.Description)
				}
				if tc.input.Channels != nil {
					assert.Equal(t, tc.input.Channels, result.Channels)
				}
				if tc.input.Metadata != nil {
					assert.Equal(t, tc.input.Metadata, result.Metadata)
				}
			}
		})
	}
}

func TestTransactionalNotificationService_GetNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name           string
		id             string
		mockSetup      func()
		expectedError  bool
		expectedResult *domain.TransactionalNotification
	}

	ctx := context.Background()
	workspace := "test-workspace"
	notificationID := uuid.New().String()

	existingNotification := &domain.TransactionalNotification{
		ID:          notificationID,
		Name:        "Test Notification",
		Description: "Test Description",
	}

	tests := []testCase{
		{
			name: "Success_GetNotification",
			id:   notificationID,
			mockSetup: func() {
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(existingNotification, nil)
			},
			expectedError:  false,
			expectedResult: existingNotification,
		},
		{
			name: "Error_NotificationNotFound",
			id:   notificationID,
			mockSetup: func() {
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(nil, errors.New("notification not found"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks for this test case
			tc.mockSetup()

			// Create service with mocked dependencies
			service := &TransactionalNotificationService{
				transactionalRepo:  mockRepo,
				messageHistoryRepo: mockMsgHistoryRepo,
				templateService:    mockTemplateService,
				contactService:     mockContactService,
				emailService:       nil, // Not used in this test
				logger:             mockLogger,
			}

			// Call the method being tested
			result, err := service.GetNotification(ctx, workspace, tc.id)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestTransactionalNotificationService_ListNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name              string
		filter            map[string]interface{}
		limit             int
		offset            int
		mockSetup         func()
		expectedError     bool
		expectedResults   []*domain.TransactionalNotification
		expectedTotalRows int
	}

	ctx := context.Background()
	workspace := "test-workspace"

	notifications := []*domain.TransactionalNotification{
		{
			ID:   uuid.New().String(),
			Name: "Notification 1",
		},
		{
			ID:   uuid.New().String(),
			Name: "Notification 2",
		},
	}

	tests := []testCase{
		{
			name:   "Success_ListNotifications",
			filter: map[string]interface{}{"name": "Test"},
			limit:  10,
			offset: 0,
			mockSetup: func() {
				mockRepo.EXPECT().
					List(gomock.Any(), workspace, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(notifications, 2, nil)
			},
			expectedError:     false,
			expectedResults:   notifications,
			expectedTotalRows: 2,
		},
		{
			name:   "Success_EmptyResults",
			filter: map[string]interface{}{"name": "NonExistent"},
			limit:  10,
			offset: 0,
			mockSetup: func() {
				mockRepo.EXPECT().
					List(gomock.Any(), workspace, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*domain.TransactionalNotification{}, 0, nil)
			},
			expectedError:     false,
			expectedResults:   []*domain.TransactionalNotification{},
			expectedTotalRows: 0,
		},
		{
			name:   "Error_RepositoryListFailed",
			filter: map[string]interface{}{},
			limit:  10,
			offset: 0,
			mockSetup: func() {
				mockRepo.EXPECT().
					List(gomock.Any(), workspace, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, errors.New("repository error"))
			},
			expectedError:     true,
			expectedResults:   nil,
			expectedTotalRows: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks for this test case
			tc.mockSetup()

			// Create service with mocked dependencies
			service := &TransactionalNotificationService{
				transactionalRepo:  mockRepo,
				messageHistoryRepo: mockMsgHistoryRepo,
				templateService:    mockTemplateService,
				contactService:     mockContactService,
				emailService:       nil, // Not used in this test
				logger:             mockLogger,
			}

			// Call the method being tested
			results, total, err := service.ListNotifications(ctx, workspace, tc.filter, tc.limit, tc.offset)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResults, results)
				assert.Equal(t, tc.expectedTotalRows, total)
			}
		})
	}
}

func TestTransactionalNotificationService_DeleteNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name          string
		id            string
		mockSetup     func()
		expectedError bool
	}

	ctx := context.Background()
	workspace := "test-workspace"
	notificationID := uuid.New().String()

	tests := []testCase{
		{
			name: "Success_DeleteNotification",
			id:   notificationID,
			mockSetup: func() {
				mockRepo.EXPECT().
					Delete(gomock.Any(), workspace, notificationID).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Error_DeleteFailed",
			id:   notificationID,
			mockSetup: func() {
				mockRepo.EXPECT().
					Delete(gomock.Any(), workspace, notificationID).
					Return(errors.New("delete failed"))
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks for this test case
			tc.mockSetup()

			// Create service with mocked dependencies
			service := &TransactionalNotificationService{
				transactionalRepo:  mockRepo,
				messageHistoryRepo: mockMsgHistoryRepo,
				templateService:    mockTemplateService,
				contactService:     mockContactService,
				emailService:       nil, // Not used in this test
				logger:             mockLogger,
			}

			// Call the method being tested
			err := service.DeleteNotification(ctx, workspace, tc.id)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
