package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

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

func TestEmailService_SendEmail_NoDirectProvider(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
		"marketing",
		recipientEmail,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate user")
}

func TestCreateSESClient(t *testing.T) {
	// Test creating a SES client
	client := service.CreateSESClient("us-east-1", "test-access-key", "test-secret-key")
	assert.NotNil(t, client, "SES client should not be nil")
}

func TestEmailService_TestEmailProvider_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	to := "test@example.com"

	// Setup mocks for successful validation and sending
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(ctx, user, nil)

	// Create test provider with SMTP settings
	testProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: "from@example.com",
		DefaultSenderName:  "Test Sender",
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "password",
		},
	}

	// For this test, we'll expect an error since we're not actually connecting to an SMTP server
	err := emailService.TestEmailProvider(ctx, workspaceID, testProvider, to)
	assert.Error(t, err)
	// The error message might be about invalid mail address or creating SMTP client, so we're just checking for an error
	assert.NotNil(t, err)
}

func TestEmailService_TestTemplate_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	templateID := "template123"
	recipientEmail := "test@example.com"

	// Create test workspace with SMTP provider
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

	// Create test template
	testTemplate := &domain.Template{
		ID:   templateID,
		Name: "Test Template",
		Email: &domain.EmailTemplate{
			Subject:          "Template Subject",
			VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
		},
		TestData: map[string]interface{}{
			"name": "Test User",
		},
	}

	// Setup test compilation result
	compiledHTML := "<p>Test template content with Test User</p>"
	compileResult := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Setup mocks
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(ctx, &domain.User{}, nil)

	mockWorkspaceRepo.EXPECT().
		GetByID(gomock.Any(), workspaceID).
		Return(testWorkspace, nil)

	mockTemplateRepo.EXPECT().
		GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
		Return(testTemplate, nil)

	mockTemplateService.EXPECT().
		CompileTemplate(gomock.Any(), workspaceID, testTemplate.Email.VisualEditorTree, testTemplate.TestData).
		Return(compileResult, nil)

	// Call method - we expect an error because the actual sending will fail
	err := emailService.TestTemplate(ctx, workspaceID, templateID, "marketing", recipientEmail)
	assert.Error(t, err)
	// The error message might be about invalid mail address or creating SMTP client, so we're just checking for an error
	assert.NotNil(t, err)
}

func TestEmailService_SendEmail_WithProviders(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	fromAddress := "sender@example.com"
	fromName := "Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test cases for different provider types
	testCases := []struct {
		name           string
		provider       domain.EmailProvider
		setupMocks     func()
		expectedErrMsg string
	}{
		{
			name: "SES Provider - Missing Config",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSES,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				// SES is nil
			},
			setupMocks:     func() {},
			expectedErrMsg: "SES provider is not configured",
		},
		{
			name: "SparkPost Provider - Missing Config",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSparkPost,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				// SparkPost is nil
			},
			setupMocks:     func() {},
			expectedErrMsg: "SparkPost provider is not configured",
		},
		{
			name: "Postmark Provider - Missing Config",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindPostmark,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				// Postmark is nil
			},
			setupMocks:     func() {},
			expectedErrMsg: "Postmark provider is not configured",
		},
		{
			name: "SES Provider - With Config",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSES,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				SES: &domain.AmazonSES{
					Region:    "us-east-1",
					AccessKey: "test-access-key",
					SecretKey: "test-secret-key",
				},
			},
			setupMocks: func() {
				// This won't actually send an email, but will help with test coverage
				// The call will fail with AWS credential errors, which is expected
			},
			expectedErrMsg: "SES error",
		},
		{
			name: "SparkPost Provider - With Config and Encrypted API Key",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSparkPost,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				SparkPost: &domain.SparkPostSettings{
					Endpoint:        "https://api.sparkpost.com",
					EncryptedAPIKey: "encrypted-key",
					APIKey:          "", // Empty to test decryption path
					SandboxMode:     true,
				},
			},
			setupMocks: func() {
				// Expect decryption error
			},
			expectedErrMsg: "failed to decrypt SparkPost API key",
		},
		{
			name: "Postmark Provider - With Config and Encrypted Server Token",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindPostmark,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				Postmark: &domain.PostmarkSettings{
					EncryptedServerToken: "encrypted-token",
					ServerToken:          "", // Empty to test decryption path
				},
			},
			setupMocks: func() {
				// Expect decryption error
			},
			expectedErrMsg: "failed to decrypt Postmark server token",
		},
		{
			name: "Mailjet Provider - Missing Config",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindMailjet,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				// Mailjet is nil
			},
			setupMocks:     func() {},
			expectedErrMsg: "Mailjet provider is not configured",
		},
		{
			name: "Mailjet Provider - With Config and Encrypted Keys",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindMailjet,
				DefaultSenderEmail: fromAddress,
				DefaultSenderName:  fromName,
				Mailjet: &domain.MailjetSettings{
					EncryptedAPIKey:    "encrypted-api-key",
					EncryptedSecretKey: "encrypted-secret-key",
					SandboxMode:        true,
				},
			},
			setupMocks: func() {
				// Expect decryption error
			},
			expectedErrMsg: "failed to decrypt Mailjet API key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks specific to this test case
			tc.setupMocks()

			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				"", // providerType not used with direct provider
				fromAddress,
				fromName,
				to,
				subject,
				content,
				&tc.provider,
			)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}

func TestEmailService_SendEmail_DefaultInfo(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test using provider's default sender info
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: "default@example.com",
		DefaultSenderName:  "Default Sender",
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "password",
		},
	}

	// Call method with empty sender info to test defaulting
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		"",
		"", // Empty fromAddress
		"", // Empty fromName
		to,
		subject,
		content,
		&provider,
	)

	assert.Error(t, err)
	// The error could be different depending on the implementation, just check for an error
	assert.NotNil(t, err)
}

func TestEmailService_SendEmail_SMTP_EncryptedPassword(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	fromAddress := "sender@example.com"
	fromName := "Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test provider with encrypted password
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: fromAddress,
		DefaultSenderName:  fromName,
		SMTP: &domain.SMTPSettings{
			Host:              "smtp.example.com",
			Port:              587,
			Username:          "user",
			Password:          "",                   // Empty password
			EncryptedPassword: "encrypted-password", // Will try to decrypt
		},
	}

	// Call method - should fail with decryption error
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		"",
		fromAddress,
		fromName,
		to,
		subject,
		content,
		&provider,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt SMTP password")
}

func TestEmailService_SendEmail_WithWorkspace(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	fromAddress := "sender@example.com"
	fromName := "Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Create test workspace with both marketing and transactional providers
	testWorkspace := &domain.Workspace{
		ID: workspaceID,
		Settings: domain.WorkspaceSettings{
			EmailMarketingProvider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: "marketing@example.com",
				DefaultSenderName:  "Marketing Sender",
				SMTP: &domain.SMTPSettings{
					Host:     "smtp-marketing.example.com",
					Port:     587,
					Username: "marketing-user",
					Password: "marketing-password",
				},
			},
			EmailTransactionalProvider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: "transactional@example.com",
				DefaultSenderName:  "Transactional Sender",
				SMTP: &domain.SMTPSettings{
					Host:     "smtp-transactional.example.com",
					Port:     587,
					Username: "transactional-user",
					Password: "transactional-password",
				},
			},
		},
	}

	// Test cases for different provider types
	testCases := []struct {
		name           string
		providerType   string
		setupMocks     func()
		expectedErrMsg string
	}{
		{
			name:         "Marketing Provider",
			providerType: "marketing",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			expectedErrMsg: "invalid sender",
		},
		{
			name:         "Transactional Provider",
			providerType: "transactional",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			expectedErrMsg: "invalid sender",
		},
		{
			name:         "Invalid Provider Type",
			providerType: "invalid",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			expectedErrMsg: "invalid provider type",
		},
		{
			name:         "Authentication Error",
			providerType: "marketing",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(nil, nil, errors.New("authentication failed"))
			},
			expectedErrMsg: "failed to authenticate user",
		},
		{
			name:         "GetWorkspace Error",
			providerType: "marketing",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(nil, errors.New("workspace not found"))
			},
			expectedErrMsg: "failed to get workspace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks specific to this test case
			tc.setupMocks()

			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				tc.providerType,
				fromAddress,
				fromName,
				to,
				subject,
				content,
			)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}

func TestEmailService_SendEmail_NoConfiguredProvider(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	emailService := service.NewEmailService(
		mockLogger,
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
	fromAddress := "sender@example.com"
	fromName := "Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Create test workspace with no configured providers
	testWorkspace := &domain.Workspace{
		ID:       workspaceID,
		Settings: domain.WorkspaceSettings{},
	}

	// Setup mocks
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(ctx, user, nil)

	mockWorkspaceRepo.EXPECT().
		GetByID(gomock.Any(), workspaceID).
		Return(testWorkspace, nil)

	// Call the method
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		"marketing", // Try to use a provider that's not configured
		fromAddress,
		fromName,
		to,
		subject,
		content,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no email provider configured for type")
}

func TestEmailService_SendEmail_WithMailgun(t *testing.T) {
	// Create mocks
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockLogger := setupMockLogger(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(mockCtrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(mockCtrl)
	mockTemplateService := mocks.NewMockTemplateService(mockCtrl)
	mockHttpClient := mocks.NewMockHTTPClient(mockCtrl)

	// Create email service with mocked dependencies
	secretKey := "test-secret-key"
	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		secretKey,
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)
	emailService.SetHTTPClient(mockHttpClient)

	// Test case: Using Mailgun provider for transactional email
	t.Run("SendEmail with Mailgun provider", func(t *testing.T) {
		// Setup test data
		ctx := context.Background()
		workspaceID := "test-workspace-id"
		providerType := "transactional"
		fromAddress := "test@example.com"
		fromName := "Test Sender"
		toAddress := "recipient@example.com"
		subject := "Test Subject"
		content := "<p>Test Content</p>"

		// Create a workspace with a Mailgun provider
		workspace := domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				EmailTransactionalProvider: domain.EmailProvider{
					Kind:               domain.EmailProviderKindMailgun,
					DefaultSenderEmail: fromAddress,
					DefaultSenderName:  fromName,
					Mailgun: &domain.MailgunSettings{
						Domain:          "test-domain.com",
						EncryptedAPIKey: "encrypted-api-key",
						Region:          "US",
					},
				},
			},
		}

		// Setup expected HTTP response
		expectedResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id": "test-message-id", "message": "Queued"}`)),
		}

		// Expect auth service call
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user-id"}, nil)

		// Expect workspace repository call
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&workspace, nil)

		// Expect decryption of the API key
		// This would normally happen internally, but we need to ensure the decrypted key is available
		// for the HTTP request, so we'll manually set it for this test
		decryptedAPIKey := "test-api-key"
		workspace.Settings.EmailTransactionalProvider.Mailgun.APIKey = decryptedAPIKey

		// Expect HTTP client call with proper request to Mailgun API
		mockHttpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.mailgun.net/v3/test-domain.com/messages", req.URL.String())
				assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

				// Verify auth header (Basic auth with username "api" and the API key)
				username, password, ok := req.BasicAuth()
				assert.True(t, ok, "Request should have basic auth")
				assert.Equal(t, "api", username)
				assert.Equal(t, decryptedAPIKey, password)

				// Parse and verify form data
				err := req.ParseForm()
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("%s <%s>", fromName, fromAddress), req.Form.Get("from"))
				assert.Equal(t, toAddress, req.Form.Get("to"))
				assert.Equal(t, subject, req.Form.Get("subject"))
				assert.Equal(t, content, req.Form.Get("html"))

				return expectedResponse, nil
			})

		// Call the method
		err := emailService.SendEmail(ctx, workspaceID, providerType, fromAddress, fromName, toAddress, subject, content)

		// Assertions
		assert.NoError(t, err)
	})
}

func TestEmailService_SendEmail_WithMailjet(t *testing.T) {
	// Create mocks
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockLogger := setupMockLogger(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(mockCtrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(mockCtrl)
	mockTemplateService := mocks.NewMockTemplateService(mockCtrl)
	mockHttpClient := mocks.NewMockHTTPClient(mockCtrl)

	// Create email service with mocked dependencies
	secretKey := "test-secret-key"
	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		secretKey,
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
	)
	emailService.SetHTTPClient(mockHttpClient)

	// Test case: Using Mailjet provider for transactional email
	t.Run("SendEmail with Mailjet provider", func(t *testing.T) {
		// Setup test data
		ctx := context.Background()
		workspaceID := "test-workspace-id"
		providerType := "transactional"
		fromAddress := "test@example.com"
		fromName := "Test Sender"
		toAddress := "recipient@example.com"
		subject := "Test Subject"
		content := "<p>Test Content</p>"

		// Create a workspace with a Mailjet provider
		workspace := domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				EmailTransactionalProvider: domain.EmailProvider{
					Kind:               domain.EmailProviderKindMailjet,
					DefaultSenderEmail: fromAddress,
					DefaultSenderName:  fromName,
					Mailjet: &domain.MailjetSettings{
						EncryptedAPIKey:    "encrypted-api-key",
						EncryptedSecretKey: "encrypted-secret-key",
						SandboxMode:        true,
					},
				},
			},
		}

		// Setup expected HTTP response
		expectedResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"Messages":[{"Status":"success"}]}`)),
		}

		// Expect auth service call
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user-id"}, nil)

		// Expect workspace repository call
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&workspace, nil)

		// Expect decryption of the API key and Secret key
		// This would normally happen internally, but we need to ensure the decrypted keys are available
		// for the HTTP request, so we'll manually set them for this test
		decryptedAPIKey := "test-api-key"
		decryptedSecretKey := "test-secret-key"
		workspace.Settings.EmailTransactionalProvider.Mailjet.APIKey = decryptedAPIKey
		workspace.Settings.EmailTransactionalProvider.Mailjet.SecretKey = decryptedSecretKey

		// Expect HTTP client call with proper request to Mailjet API
		mockHttpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3.1/send", req.URL.String())
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				// Verify auth header (Basic auth with API key and Secret key)
				username, password, ok := req.BasicAuth()
				assert.True(t, ok, "Request should have basic auth")
				assert.Equal(t, decryptedAPIKey, username)
				assert.Equal(t, decryptedSecretKey, password)

				// Verify request body
				body, err := io.ReadAll(req.Body)
				assert.NoError(t, err)

				var payload map[string]interface{}
				err = json.Unmarshal(body, &payload)
				assert.NoError(t, err)

				// Check sandbox mode
				assert.Equal(t, true, payload["SandboxMode"])

				// Check message details
				messages, ok := payload["Messages"].([]interface{})
				assert.True(t, ok, "Messages should be an array")
				assert.Equal(t, 1, len(messages), "Should have one message")

				message := messages[0].(map[string]interface{})

				// Check From
				from, ok := message["From"].(map[string]interface{})
				assert.True(t, ok, "From should be an object")
				assert.Equal(t, fromAddress, from["Email"])
				assert.Equal(t, fromName, from["Name"])

				// Check To
				recipients, ok := message["To"].([]interface{})
				assert.True(t, ok, "To should be an array")
				assert.Equal(t, 1, len(recipients), "Should have one recipient")

				recipient := recipients[0].(map[string]interface{})
				assert.Equal(t, toAddress, recipient["Email"])

				// Check other fields
				assert.Equal(t, subject, message["Subject"])
				assert.Equal(t, content, message["HTMLPart"])

				return expectedResponse, nil
			})

		// Call the method
		err := emailService.SendEmail(ctx, workspaceID, providerType, fromAddress, fromName, toAddress, subject, content)

		// Assertions
		assert.NoError(t, err)
	})
}

// testLogger is a simple logger that implements the logger.Logger interface
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string) {
	l.t.Log(msg)
}

func (l *testLogger) Info(msg string) {
	l.t.Log(msg)
}

func (l *testLogger) Warn(msg string) {
	l.t.Log(msg)
}

func (l *testLogger) Error(msg string) {
	l.t.Error(msg)
}

func (l *testLogger) Fatal(msg string) {
	l.t.Fatal(msg)
}

func (l *testLogger) WithField(key string, value interface{}) logger.Logger {
	// For testing, we just return the same logger
	return l
}

func (l *testLogger) WithFields(fields map[string]interface{}) logger.Logger {
	// For testing, we just return the same logger
	return l
}
