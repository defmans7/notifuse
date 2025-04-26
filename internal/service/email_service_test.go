package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// mockLogger is a simple mock implementation of the logger.Logger interface for testing
type mockLogger struct{}

func (l *mockLogger) Info(message string)                                    {}
func (l *mockLogger) Error(message string)                                   {}
func (l *mockLogger) Debug(message string)                                   {}
func (l *mockLogger) Warn(message string)                                    {}
func (l *mockLogger) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *mockLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }
func (l *mockLogger) Fatal(message string)                                   {}

func TestEmailService_SendEmail_NoDirectProvider(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		&mockLogger{},
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)
	emailService.SetHTTPClient(mockHTTPClient)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}

	// Create test workspace
	testWorkspace := &domain.Workspace{
		ID: workspaceID,
		Settings: domain.WorkspaceSettings{
			EmailMarketingProvider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: "from@example.com",
				DefaultSenderName:  "Test Sender",
				SMTP: &domain.SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	// Test cases
	tests := []struct {
		name          string
		setupMocks    func()
		providerType  string
		expectedError string
	}{
		{
			name: "Authentication Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(nil, nil, errors.New("auth error"))
			},
			providerType:  "marketing",
			expectedError: "failed to authenticate user",
		},
		{
			name: "Get Workspace Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(nil, errors.New("workspace not found"))
			},
			providerType:  "marketing",
			expectedError: "failed to get workspace",
		},
		{
			name: "Invalid Provider Type",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			providerType:  "invalid",
			expectedError: "invalid provider type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				tt.providerType,
				"sender@example.com",
				"Sender",
				"recipient@example.com",
				"Test Subject",
				"<p>Test Content</p>",
			)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_SendEmail_DirectProvider(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		&mockLogger{},
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)
	emailService.SetHTTPClient(mockHTTPClient)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"

	// Test case with direct provider
	provider := domain.EmailProvider{
		Kind: domain.EmailProviderKindSMTP,
		// No SMTP config, should fail with "SMTP settings required"
	}

	// Test
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		"", // providerType not used with direct provider
		"sender@example.com",
		"Sender",
		"recipient@example.com",
		"Test Subject",
		"<p>Test Content</p>",
		&provider,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP settings required")
}

func TestEmailService_TestEmailProvider(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)

	emailService := service.NewEmailService(
		&mockLogger{},
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	to := "test@example.com"

	// Test case for authentication error
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(nil, nil, errors.New("auth error"))

	// Test
	err := emailService.TestEmailProvider(
		ctx,
		workspaceID,
		domain.EmailProvider{Kind: domain.EmailProviderKindSMTP},
		to,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
}

func TestEmailService_TestTemplate(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)

	emailService := service.NewEmailService(
		&mockLogger{},
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	templateID := "template123"
	providerType := "marketing"
	recipientEmail := "test@example.com"

	// Test case for authentication error
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(nil, nil, errors.New("auth error"))

	// Test
	err := emailService.TestTemplate(
		ctx,
		workspaceID,
		templateID,
		providerType,
		recipientEmail,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate user")
}
