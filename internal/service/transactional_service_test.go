package service

import (
	"context"
	"errors"
	"testing"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockEmailService is a simplified mock for testing
type mockEmailService struct {
	logger logger.Logger
}

func (m *mockEmailService) SendEmail(ctx context.Context, workspaceID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, optionalProvider *domain.EmailProvider, replyTo string, cc []string, bcc []string) error {
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
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

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
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
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
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

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
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
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
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

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
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
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
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

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
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
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
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

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
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
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

func TestNewTransactionalNotificationService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	apiEndpoint := "https://api.example.com"

	service := NewTransactionalNotificationService(
		mockRepo,
		mockMsgHistoryRepo,
		mockTemplateService,
		mockContactService,
		mockEmailService,
		mockLogger,
		mockWorkspaceRepo,
		apiEndpoint,
	)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.transactionalRepo)
	assert.Equal(t, mockMsgHistoryRepo, service.messageHistoryRepo)
	assert.Equal(t, mockTemplateService, service.templateService)
	assert.Equal(t, mockContactService, service.contactService)
	assert.Equal(t, mockEmailService, service.emailService)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, mockWorkspaceRepo, service.workspaceRepo)
	assert.Equal(t, apiEndpoint, service.apiEndpoint)
}

func TestTransactionalNotificationService_SendNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name           string
		params         domain.TransactionalNotificationSendParams
		mockSetup      func()
		expectedError  bool
		expectedResult string
	}

	ctx := context.Background()
	workspace := "test-workspace"
	notificationID := uuid.New().String()
	templateID := uuid.New().String()

	// Create a sample notification and contact for tests
	notification := &domain.TransactionalNotification{
		ID:          notificationID,
		Name:        "Test Notification",
		Description: "Test Description",
		Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
			domain.TransactionalChannelEmail: {
				TemplateID: templateID,
			},
		},
	}

	workspaceObj := &domain.Workspace{
		ID:   workspace,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			TransactionalEmailProviderID: "integration-1",
		},
		Integrations: []domain.Integration{
			{
				ID:   "integration-1",
				Name: "Test Integration",
				Type: "email",
				EmailProvider: domain.EmailProvider{
					Kind:               domain.EmailProviderKindSparkPost,
					DefaultSenderEmail: "test@example.com",
					DefaultSenderName:  "Test Sender",
					SparkPost: &domain.SparkPostSettings{
						EncryptedAPIKey: "encrypted-api-key",
					},
				},
			},
		},
	}

	contact := &domain.Contact{
		Email: "test@example.com",
		FirstName: &domain.NullableString{
			String: "John",
			IsNull: false,
		},
		LastName: &domain.NullableString{
			String: "Doe",
			IsNull: false,
		},
	}

	tests := []testCase{
		{
			name: "Success_SendNotification",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
				Data: map[string]interface{}{
					"product_name": "Test Product",
					"order_id":     "12345",
				},
				Metadata: map[string]interface{}{
					"source": "api",
				},
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert the contact
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// Get the contact after upsert
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(contact, nil)

				// Expect call to DoSendEmailNotification which will be mocked in a special test version
				template := notification.Channels[domain.TransactionalChannelEmail]

				// Since we're testing a method on the real service (not a mocked one),
				// we'll need to mock the templateService and emailService that DoSendEmailNotification calls

				// First get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, template.TemplateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "test@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Then the template is compiled
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    aws.String("<html>Test content</html>"),
					}, nil)

				// Then the email is sent - ensure provider is non-nil
				mockEmailService.EXPECT().
					SendEmail(
						gomock.Any(),
						workspace,
						false,
						"test@example.com",
						"Test Sender",
						contact.Email,
						"Test Subject",
						"<html>Test content</html>",
						gomock.Not(gomock.Nil()), // Ensure we expect a non-nil provider
						gomock.Any(),             // replyTo
						gomock.Any(),             // cc
						gomock.Any(),             // bcc
					).Return(nil)

				// Finally, message history is created
				mockMsgHistoryRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					Return(nil)
			},
			expectedError:  false,
			expectedResult: gomock.Any().String(), // We expect a non-empty message ID
		},
		{
			name: "Error_NotificationNotFound",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Notification not found
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(nil, errors.New("notification not found"))
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_ContactMissing",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: nil,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_UpsertContactFailed",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert contact fails
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationError,
						Error:  "contact validation failed",
					})
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_GetContactAfterUpsertFailed",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert contact succeeds
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// But getting the contact after upsert fails
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(nil, errors.New("contact not found after upsert"))
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_InvalidChannel",
			params: domain.TransactionalNotificationSendParams{
				ID:       notificationID,
				Contact:  contact,
				Channels: []domain.TransactionalChannel{"invalid-channel"},
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert the contact
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// Get the contact after upsert
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(contact, nil)

				// No further calls as the channel validation will fail
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_TemplateGetFailed",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert the contact
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// Get the contact after upsert
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(contact, nil)

				// Then getting the template fails
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, gomock.Any(), int64(0)).
					Return(nil, errors.New("template not found"))
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_TemplateCompilationFailed",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert the contact
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// Get the contact after upsert
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(contact, nil)

				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, gomock.Any(), int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "test@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Then compilation fails
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: false,
						Error: &mjmlgo.Error{
							Message: "Compilation error",
						},
					}, nil)
			},
			expectedError:  true,
			expectedResult: "",
		},
		{
			name: "Error_EmailSendFailed",
			params: domain.TransactionalNotificationSendParams{
				ID:      notificationID,
				Contact: contact,
			},
			mockSetup: func() {
				// Get the workspace
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspace).
					Return(workspaceObj, nil)

				// Get the notification
				mockRepo.EXPECT().
					Get(gomock.Any(), workspace, notificationID).
					Return(notification, nil)

				// Upsert the contact
				mockContactService.EXPECT().
					UpsertContact(gomock.Any(), workspace, contact).
					Return(domain.UpsertContactOperation{
						Email:  contact.Email,
						Action: domain.UpsertContactOperationUpdate,
					})

				// Get the contact after upsert
				mockContactService.EXPECT().
					GetContactByEmail(gomock.Any(), workspace, contact.Email).
					Return(contact, nil)

				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, gomock.Any(), int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "test@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compilation succeeds
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    aws.String("<html>Test content</html>"),
					}, nil)

				// First, expect message history to be created with "sent" status initially
				mockMsgHistoryRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, message *domain.MessageHistory) error {
						assert.Equal(t, domain.MessageStatusSent, message.Status)
						return nil
					})

				// Then, email sending fails - ensure provider is non-nil
				mockEmailService.EXPECT().
					SendEmail(
						gomock.Any(),
						workspace,
						false,
						"test@example.com",
						"Test Sender",
						contact.Email,
						"Test Subject",
						"<html>Test content</html>",
						gomock.Not(gomock.Nil()), // Ensure we expect a non-nil provider
						gomock.Any(),             // replyTo
						gomock.Any(),             // cc
						gomock.Any(),             // bcc
					).Return(errors.New("email sending failed"))

				// Finally, expect message history to be updated with "failed" status
				mockMsgHistoryRepo.EXPECT().
					Update(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, message *domain.MessageHistory) error {
						assert.Equal(t, domain.MessageStatusFailed, message.Status)
						assert.NotNil(t, message.Error)
						return nil
					})
			},
			expectedError:  true,
			expectedResult: "",
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
				emailService:       mockEmailService,
				logger:             mockLogger,
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
			}

			// Call the method being tested
			result, err := service.SendNotification(ctx, workspace, tc.params)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedResult, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result) // Should be a UUID
			}
		})
	}
}

func TestTransactionalNotificationService_DoSendEmailNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
	mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Create a stub logger that simply returns itself for chaining calls
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	type testCase struct {
		name           string
		templateConfig domain.ChannelTemplate
		messageData    domain.MessageData
		mockSetup      func()
		expectedError  bool
	}

	ctx := context.Background()
	workspace := "test-workspace"
	messageID := uuid.New().String()
	templateID := uuid.New().String()

	contact := &domain.Contact{
		Email: "test@example.com",
		FirstName: &domain.NullableString{
			String: "John",
			IsNull: false,
		},
		LastName: &domain.NullableString{
			String: "Doe",
			IsNull: false,
		},
	}

	templateConfig := domain.ChannelTemplate{
		TemplateID: templateID,
	}

	messageData := domain.MessageData{
		Data: map[string]interface{}{
			"contact":  contact,
			"order_id": "12345",
		},
		Metadata: map[string]interface{}{
			"source": "api",
		},
	}

	tests := []testCase{
		{
			name:           "Success_SendEmail",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "sender@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compile the template
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    aws.String("<html>Test Email</html>"),
					}, nil)

				// Send the email
				mockEmailService.EXPECT().
					SendEmail(
						gomock.Any(),
						workspace,
						false,
						"sender@example.com",      // FromAddress
						"Test Sender",             // FromName
						contact.Email,             // To
						"Test Subject",            // Subject
						"<html>Test Email</html>", // Content
						gomock.Not(gomock.Nil()),  // Expect non-nil email provider
						gomock.Any(),              // replyTo
						gomock.Any(),              // cc
						gomock.Any(),              // bcc
					).Return(nil)

				// Record the message history
				mockMsgHistoryRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, message *domain.MessageHistory) error {
						assert.Equal(t, messageID, message.ID)
						assert.Equal(t, contact.Email, message.ContactID)
						assert.Equal(t, templateID, message.TemplateID)
						assert.Equal(t, "email", message.Channel)
						assert.Equal(t, domain.MessageStatusSent, message.Status)
						assert.NotNil(t, message.MessageData)
						return nil
					})
			},
			expectedError: false,
		},
		{
			name:           "Error_TemplateNotFound",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Template not found
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(nil, errors.New("template not found"))
			},
			expectedError: true,
		},
		{
			name:           "Error_TemplateCompilationFailed",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "sender@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compilation fails
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: false,
						Error: &mjmlgo.Error{
							Message: "compilation error",
						},
					}, nil)
			},
			expectedError: true,
		},
		{
			name:           "Error_CompilationWithNoHTMLContent",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "sender@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compilation succeeds but without HTML content
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    nil, // Missing HTML content
					}, nil)
			},
			expectedError: true,
		},
		{
			name:           "Error_EmailSendFailed",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "sender@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compile the template
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    aws.String("<html>Test Email</html>"),
					}, nil)

				// First, expect message history to be created with "sent" status initially
				mockMsgHistoryRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, message *domain.MessageHistory) error {
						assert.Equal(t, messageID, message.ID)
						assert.Equal(t, domain.MessageStatusSent, message.Status)
						return nil
					})

				// Then, email sending fails - ensure provider is non-nil
				mockEmailService.EXPECT().
					SendEmail(
						gomock.Any(),
						workspace,
						false,
						"sender@example.com",      // FromAddress
						"Test Sender",             // FromName
						contact.Email,             // To
						"Test Subject",            // Subject
						"<html>Test Email</html>", // Content
						gomock.Not(gomock.Nil()),  // Ensure we expect a non-nil provider
						gomock.Any(),              // replyTo
						gomock.Any(),              // cc
						gomock.Any(),              // bcc
					).Return(errors.New("email sending failed"))

				// Finally, expect message history to be updated with "failed" status
				mockMsgHistoryRepo.EXPECT().
					Update(gomock.Any(), workspace, gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, message *domain.MessageHistory) error {
						assert.Equal(t, messageID, message.ID)
						assert.Equal(t, domain.MessageStatusFailed, message.Status)
						assert.NotNil(t, message.Error)
						return nil
					})
			},
			expectedError: true,
		},
		{
			name:           "Error_MessageHistoryCreationFailed",
			templateConfig: templateConfig,
			messageData:    messageData,
			mockSetup: func() {
				// Get the template
				mockTemplateService.EXPECT().
					GetTemplateByID(gomock.Any(), workspace, templateID, int64(0)).
					Return(&domain.Template{
						ID: templateID,
						Email: &domain.EmailTemplate{
							Subject:     "Test Subject",
							FromAddress: "sender@example.com",
							FromName:    "Test Sender",
						},
					}, nil)

				// Compile the template
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    aws.String("<html>Test Email</html>"),
					}, nil)

				// Message history creation fails
				mockMsgHistoryRepo.EXPECT().
					Create(gomock.Any(), workspace, gomock.Any()).
					Return(errors.New("message history creation failed"))

				// The email should never be sent if message history creation fails
				// We don't expect a call to SendEmail here
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
				emailService:       mockEmailService,
				logger:             mockLogger,
				workspaceRepo:      mockWorkspaceRepo,
				apiEndpoint:        "https://api.example.com",
			}

			// Call the method being tested
			err := service.DoSendEmailNotification(
				ctx,
				workspace,
				messageID,
				contact,
				templateConfig,
				messageData,
				mjml.TrackingSettings{
					EnableTracking: true,
					UTMSource:      "test",
					UTMMedium:      "email",
					UTMCampaign:    "test_campaign",
				},
				&domain.EmailProvider{
					Kind:               domain.EmailProviderKindSparkPost,
					DefaultSenderEmail: "test@example.com",
					DefaultSenderName:  "Test Sender",
					SparkPost: &domain.SparkPostSettings{
						EncryptedAPIKey: "encrypted-api-key",
					},
				},
				[]string{}, // cc
				[]string{}, // bcc
			)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
