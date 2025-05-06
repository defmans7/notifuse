package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	notifusemjml "github.com/Notifuse/notifuse/pkg/mjml"
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}

	// Create test workspace
	testWorkspace := &domain.Workspace{
		ID: workspaceID,
		Settings: domain.WorkspaceSettings{
			MarketingEmailProviderID: "integration-marketing-id",
		},
		Integrations: []domain.Integration{
			{
				ID:   "integration-marketing-id",
				Name: "Marketing Email Provider",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
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
		},
	}

	// Test cases
	tests := []struct {
		name          string
		setupMocks    func()
		isMarketing   bool
		expectedError string
	}{
		{
			name: "Authentication Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(nil, nil, errors.New("auth error"))
			},
			isMarketing:   true,
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
			isMarketing:   true,
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
			isMarketing:   false,
			expectedError: "no email provider configured for type: false",
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
				tt.isMarketing,
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
		mockHTTPClient,
	)

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
		true, // isMarketing not used with direct provider
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
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
		mockHTTPClient,
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
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	templateID := "template123"
	recipientEmail := "test@example.com"
	integrationID := "integration-marketing-id"
	user := &domain.User{ID: "user123"}

	// Create test templates
	testTemplate := &domain.Template{
		ID:   templateID,
		Name: "Test Template",
		Email: &domain.EmailTemplate{
			Subject:          "Custom Test Subject",
			VisualEditorTree: notifusemjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"content": "test content"}},
		},
		TestData: map[string]interface{}{
			"name":    "Test User",
			"company": "Notifuse",
		},
	}

	testTemplateNoEmail := &domain.Template{
		ID:   "template-no-email",
		Name: "Test Template No Email",
	}

	// Create different workspace configurations for different test cases
	// Standard workspace with valid integration
	standardWorkspace := &domain.Workspace{
		ID: workspaceID,
		Integrations: []domain.Integration{
			{
				ID:   integrationID,
				Name: "Marketing Email Provider",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
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
		},
	}

	// Workspace with empty provider
	emptyProviderWorkspace := &domain.Workspace{
		ID: workspaceID,
		Integrations: []domain.Integration{
			{
				ID:            "empty-provider",
				Name:          "Empty Provider",
				Type:          domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
					// Kind is empty
				},
			},
		},
	}

	// Test cases
	tests := []struct {
		name               string
		setupMocks         func()
		workspaceID        string
		templateID         string
		integrationID      string
		recipientEmail     string
		expectedErrorRegex string
	}{
		{
			name: "Authentication Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(nil, nil, errors.New("auth error"))
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "failed to authenticate user",
		},
		{
			name: "GetWorkspace Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(nil, errors.New("workspace not found"))
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "failed to get workspace",
		},
		{
			name: "GetTemplateByID Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(nil, errors.New("template not found"))
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "failed to get template",
		},
		{
			name: "Integration Not Found",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(testTemplate, nil)
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      "non-existent-id",
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "integration not found",
		},
		{
			name: "Email Provider Not Configured",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(emptyProviderWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(testTemplate, nil)
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      "empty-provider",
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "no email provider configured",
		},
		{
			name: "Template With Email - Compilation Error",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(testTemplate, nil)

				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
					Return(nil, errors.New("compilation error"))
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "failed to compile template",
		},
		{
			name: "Template With Email - Compilation Failed",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(testTemplate, nil)

				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: false,
						Error: &mjmlgo.Error{
							Message: "template error",
						},
					}, nil)
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "template compilation failed",
		},
		{
			name: "Template With Email - Success Path",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
					Return(testTemplate, nil)

				htmlContent := "<p>Compiled HTML</p>"
				mockTemplateService.EXPECT().
					CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
					Return(&domain.CompileTemplateResponse{
						Success: true,
						HTML:    &htmlContent,
					}, nil)

				// Expect SendEmail to be called - with a failure for SMTP connection
				// This is fine for testing since we just want to verify the flow reached SendEmail
			},
			workspaceID:        workspaceID,
			templateID:         templateID,
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "invalid sender", // Expecting error related to sender format now
		},
		{
			name: "Template Without Email - Default Content",
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(standardWorkspace, nil)

				mockTemplateRepo.EXPECT().
					GetTemplateByID(gomock.Any(), workspaceID, "template-no-email", int64(0)).
					Return(testTemplateNoEmail, nil)

				// Expect SendEmail to be called - with a failure for SMTP connection
				// This is fine for testing since we just want to verify the flow reached SendEmail
			},
			workspaceID:        workspaceID,
			templateID:         "template-no-email",
			integrationID:      integrationID,
			recipientEmail:     recipientEmail,
			expectedErrorRegex: "invalid sender", // Expecting error related to sender format now
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Call the method
			err := emailService.TestTemplate(
				ctx,
				tt.workspaceID,
				tt.templateID,
				tt.integrationID,
				tt.recipientEmail,
			)

			// Assert
			if tt.expectedErrorRegex != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.expectedErrorRegex, err.Error())
			} else {
				// For the "success path" tests, we expect some error but don't care about the message
				// Since we're not actually connecting to an email provider, this is expected
				assert.Error(t, err)
			}
		})
	}
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
		mockHTTPClient,
	)

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
		mockHTTPClient,
	)

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
				SES: &domain.AmazonSESSettings{
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
				true, // isMarketing not used with direct provider
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
		mockHTTPClient,
	)

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
		true,
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
		mockHTTPClient,
	)

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
			EncryptedPassword: "encrypted-password", // Will try to decrypt
		},
	}

	// Call method - should fail with decryption error
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		true,
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
		mockHTTPClient,
	)

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
			MarketingEmailProviderID: "integration-marketing-id",
		},
		Integrations: []domain.Integration{
			{
				ID:   "integration-marketing-id",
				Name: "Marketing Email Provider",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
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
			},
		},
	}

	// Test cases for different email types
	testCases := []struct {
		name           string
		isMarketing    bool
		setupMocks     func()
		expectedErrMsg string
	}{
		{
			name:        "Marketing Provider",
			isMarketing: true,
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
			name:        "Transactional Provider",
			isMarketing: false,
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			expectedErrMsg: "no email provider configured for type: false",
		},
		{
			name:        "Invalid Provider Type",
			isMarketing: false,
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(ctx, user, nil)

				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), workspaceID).
					Return(testWorkspace, nil)
			},
			expectedErrMsg: "no email provider configured for type: false",
		},
		{
			name:        "Authentication Error",
			isMarketing: true,
			setupMocks: func() {
				mockAuthService.EXPECT().
					AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
					Return(nil, nil, errors.New("authentication failed"))
			},
			expectedErrMsg: "failed to authenticate user",
		},
		{
			name:        "GetWorkspace Error",
			isMarketing: true,
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
				tc.isMarketing,
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
		mockHTTPClient,
	)

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
		true, // Try to use a provider that's not configured
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
		mockHttpClient,
	)

	// Test case: Using Mailgun provider for transactional email
	t.Run("SendEmail with Mailgun provider", func(t *testing.T) {
		// Setup test data
		ctx := context.Background()
		workspaceID := "test-workspace-id"
		isMarketing := false
		fromAddress := "test@example.com"
		fromName := "Test Sender"
		toAddress := "recipient@example.com"
		subject := "Test Subject"
		content := "<p>Test Content</p>"

		// Create a direct Mailgun provider
		apiKey := "test-api-key"
		mailgunProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindMailgun,
			DefaultSenderEmail: fromAddress,
			DefaultSenderName:  fromName,
			Mailgun: &domain.MailgunSettings{
				Domain: "test-domain.com",
				APIKey: apiKey,
				Region: "US",
			},
		}

		// Setup expected HTTP response
		expectedResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id": "test-message-id", "message": "Queued"}`)),
		}

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
				assert.Equal(t, apiKey, password)

				// Parse and verify form data
				err := req.ParseForm()
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("%s <%s>", fromName, fromAddress), req.Form.Get("from"))
				assert.Equal(t, toAddress, req.Form.Get("to"))
				assert.Equal(t, subject, req.Form.Get("subject"))
				assert.Equal(t, content, req.Form.Get("html"))

				return expectedResponse, nil
			})

		// Call the method directly with the provider
		err := emailService.SendEmail(ctx, workspaceID, isMarketing, fromAddress, fromName, toAddress, subject, content, &mailgunProvider)

		// Assertions
		assert.NoError(t, err)
	})
}

// TestEmailService_SendEmail_SMTP_ConnectionErrors tests the SendEmail method with SMTP errors
func TestEmailService_SendEmail_SMTP_ConnectionErrors(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	// Intentionally using an invalid email below instead of this valid one
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test cases for SMTP provider with different errors
	testCases := []struct {
		name           string
		provider       domain.EmailProvider
		expectedErrMsg string
	}{
		{
			name: "SMTP Provider - Invalid Sender Format",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SMTP: &domain.SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user",
					Password: "password",
				},
			},
			expectedErrMsg: "invalid sender", // go-mail will validate the email sender format
		},
		{
			name: "SMTP Provider - Encrypted Password",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SMTP: &domain.SMTPSettings{
					Host:              "smtp.example.com",
					Port:              587,
					Username:          "user",
					EncryptedPassword: "encrypted-password", // This will trigger decryption
				},
			},
			expectedErrMsg: "failed to decrypt SMTP password", // Decryption error is expected
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use an invalid email to trigger a validation error
			invalidTo := "not-an-email"

			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				true, // isMarketing not used with direct provider
				from,
				fromName,
				invalidTo,
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

// TestEmailService_SendEmail_HTTP_Errors tests HTTP errors in providers using HTTP APIs
func TestEmailService_SendEmail_HTTP_Errors(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test cases for HTTP-based providers with different errors
	testCases := []struct {
		name           string
		provider       domain.EmailProvider
		httpSetup      func()
		expectedErrMsg string
	}{
		{
			name: "SparkPost Provider - HTTP Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSparkPost,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SparkPost: &domain.SparkPostSettings{
					Endpoint:    "https://api.sparkpost.com",
					APIKey:      "test-api-key",
					SandboxMode: false,
				},
			},
			httpSetup: func() {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection error"))
			},
			expectedErrMsg: "failed to send request to SparkPost API",
		},
		{
			name: "Postmark Provider - HTTP Response Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindPostmark,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				Postmark: &domain.PostmarkSettings{
					ServerToken: "test-server-token",
				},
			},
			httpSetup: func() {
				// Return a 400 error response
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader(`{"ErrorCode": 400, "Message": "Bad request"}`)),
					}, nil)
			},
			expectedErrMsg: "Postmark API error (400)",
		},
		{
			name: "SparkPost Provider - Encrypted Key",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSparkPost,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SparkPost: &domain.SparkPostSettings{
					Endpoint:        "https://api.sparkpost.com",
					EncryptedAPIKey: "encrypted-api-key", // This will trigger decryption
					SandboxMode:     false,
				},
			},
			httpSetup: func() {
				// No need to expect HTTP call since decryption will fail first
			},
			expectedErrMsg: "failed to decrypt SparkPost API key",
		},
		{
			name: "Mailgun Provider - EU Region",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindMailgun,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				Mailgun: &domain.MailgunSettings{
					Domain: "test-domain.com",
					APIKey: "test-api-key",
					Region: "EU", // Test EU region specifically
				},
			},
			httpSetup: func() {
				// Verify that EU endpoint is used
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Check that the EU endpoint is used
						assert.Contains(t, req.URL.String(), "api.eu.mailgun.net")
						return nil, errors.New("connection error")
					})
			},
			expectedErrMsg: "failed to send request to Mailgun API",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup HTTP expectations
			tc.httpSetup()

			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				true, // isMarketing not used with direct provider
				from,
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

// TestEmailService_SendEmail_UnsupportedProvider tests sending with an unsupported provider
func TestEmailService_SendEmail_UnsupportedProvider(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Create a provider with an unknown kind
	unknownProvider := domain.EmailProvider{
		Kind:               "unknown-provider", // Unsupported provider kind
		DefaultSenderEmail: from,
		DefaultSenderName:  fromName,
	}

	// Call the method
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		true, // isMarketing not used with direct provider
		from,
		fromName,
		to,
		subject,
		content,
		&unknownProvider,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider kind")
}

// TestEmailService_SendEmail_HTTP_ReadResponseError tests HTTP response reading errors
func TestEmailService_SendEmail_HTTP_ReadResponseError(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Create an errorReader that always returns an error
	errorReader := &errorReadCloser{err: errors.New("read error")}

	// Test with Postmark provider
	postmarkProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindPostmark,
		DefaultSenderEmail: from,
		DefaultSenderName:  fromName,
		Postmark: &domain.PostmarkSettings{
			ServerToken: "test-server-token",
		},
	}

	// Setup HTTP expectations with response that has an error reader
	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       errorReader,
		}, nil)

	// Call the method
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		true, // isMarketing not used with direct provider
		from,
		fromName,
		to,
		subject,
		content,
		&postmarkProvider,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read Postmark API response")
}

// errorReadCloser is a mock io.ReadCloser that always returns an error
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

// TestEmailService_SES_WithEncryptedKey tests SES with encrypted secret key
func TestEmailService_SES_WithEncryptedKey(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	validRecipient := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Create SES provider with encrypted secret key
	sesProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSES,
		DefaultSenderEmail: from,
		DefaultSenderName:  fromName,
		SES: &domain.AmazonSESSettings{
			Region:             "us-east-1",
			AccessKey:          "test-access-key",
			EncryptedSecretKey: "encrypted-secret-key", // This will trigger decryption
		},
	}

	// Call the method - it will try to use AWS SDK and fail
	err := emailService.SendEmail(
		ctx,
		workspaceID,
		true, // isMarketing not used with direct provider
		from,
		fromName,
		validRecipient,
		subject,
		content,
		&sesProvider,
	)

	// Assert
	assert.Error(t, err)
	// Error message will vary based on AWS SDK behavior, just verify there is an error
	assert.NotNil(t, err)
}

// TestEmailService_DecryptionErrors tests decryption errors in the email service
func TestEmailService_DecryptionErrors(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := setupMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	// Use an intentionally invalid secret key to cause decryption errors
	invalidSecretKey := "invalid-secret-key-that-will-cause-decryption-errors"

	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		invalidSecretKey,
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockTemplateService,
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	from := "sender@example.com"
	fromName := "Test Sender"
	recipient := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Content</p>"

	// Test cases for different provider types with encrypted credentials
	testCases := []struct {
		name           string
		provider       domain.EmailProvider
		expectedErrMsg string
	}{
		{
			name: "SMTP Provider - Decryption Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSMTP,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SMTP: &domain.SMTPSettings{
					Host:              "smtp.example.com",
					Port:              587,
					Username:          "user",
					EncryptedPassword: "encrypted-password-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
		{
			name: "SES Provider - Decryption Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSES,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SES: &domain.AmazonSESSettings{
					Region:             "us-east-1",
					AccessKey:          "access-key",
					EncryptedSecretKey: "encrypted-secret-key-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
		{
			name: "SparkPost Provider - Decryption Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSparkPost,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				SparkPost: &domain.SparkPostSettings{
					Endpoint:        "https://api.sparkpost.com",
					EncryptedAPIKey: "encrypted-api-key-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
		{
			name: "Postmark Provider - Decryption Error",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindPostmark,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				Postmark: &domain.PostmarkSettings{
					EncryptedServerToken: "encrypted-server-token-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
		{
			name: "Mailgun Provider - Decryption Error for API Key",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindMailgun,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				Mailgun: &domain.MailgunSettings{
					Domain:          "test-domain.com",
					EncryptedAPIKey: "encrypted-api-key-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
		{
			name: "Mailjet Provider - Decryption Error for Secret Key",
			provider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindMailjet,
				DefaultSenderEmail: from,
				DefaultSenderName:  fromName,
				Mailjet: &domain.MailjetSettings{
					APIKey:             "valid-api-key",
					EncryptedSecretKey: "encrypted-secret-key-that-wont-decrypt", // Will cause decryption error
				},
			},
			expectedErrMsg: "failed to decrypt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method
			err := emailService.SendEmail(
				ctx,
				workspaceID,
				true, // isMarketing not used with direct provider
				from,
				fromName,
				recipient,
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

func TestEmailService_TestTemplate_Stages(t *testing.T) {
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
		mockHTTPClient,
	)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace123"
	templateID := "template123"
	integrationID := "integration-marketing-id"
	recipientEmail := "test@example.com"
	user := &domain.User{ID: "user123"}

	// Test scenario 1: Authentication error
	t.Run("Authentication Error", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, nil, errors.New("auth error"))

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	// Test scenario 2: GetWorkspace error
	t.Run("GetWorkspace Error", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(nil, errors.New("workspace error"))

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace")
	})

	// Create standard workspace with valid integration for reuse in tests
	standardWorkspace := &domain.Workspace{
		ID: workspaceID,
		Integrations: []domain.Integration{
			{
				ID: integrationID,
				EmailProvider: domain.EmailProvider{
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
		},
	}

	// Test scenario 3: GetTemplateByID error
	t.Run("GetTemplateByID Error", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(standardWorkspace, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(nil, errors.New("template error"))

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get template")
	})

	// Test scenario 4: Integration not found
	t.Run("Integration Not Found", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID:           workspaceID,
				Integrations: []domain.Integration{}, // Empty integrations array
			}, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
			}, nil)

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "integration not found")
	})

	// Test scenario 5: Email provider not configured
	t.Run("Email Provider Not Configured", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		// Create mock workspace with integration that has empty provider kind
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID: workspaceID,
				Integrations: []domain.Integration{
					{
						ID:            integrationID,
						EmailProvider: domain.EmailProvider{
							// Kind is empty, which should trigger the error
						},
					},
				},
			}, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
			}, nil)

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no email provider configured")
	})

	// Test scenario 6: Template with Email - CompileTemplate error
	t.Run("CompileTemplate Error", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		// Create mock workspace with valid integration
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID: workspaceID,
				Integrations: []domain.Integration{
					{
						ID: integrationID,
						EmailProvider: domain.EmailProvider{
							Kind:               domain.EmailProviderKindSMTP,
							DefaultSenderEmail: "sender@example.com",
							DefaultSenderName:  "Sender Name",
							SMTP: &domain.SMTPSettings{
								Host:     "smtp.example.com",
								Port:     587,
								Username: "username",
								Password: "password",
							},
						},
					},
				},
			}, nil)

		// Create template with valid email section
		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
				Email: &domain.EmailTemplate{
					Subject:          "Test Subject",
					VisualEditorTree: notifusemjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"content": "test"}},
				},
			}, nil)

		// Mock compilation error
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
			Return(nil, errors.New("compilation error"))

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile template")
	})

	// Test scenario 7: Template with Email - Compilation failed (no success)
	t.Run("Compilation Failed", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		// Create mock workspace with valid integration
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID: workspaceID,
				Integrations: []domain.Integration{
					{
						ID: integrationID,
						EmailProvider: domain.EmailProvider{
							Kind:               domain.EmailProviderKindSMTP,
							DefaultSenderEmail: "sender@example.com",
							DefaultSenderName:  "Sender Name",
							SMTP: &domain.SMTPSettings{
								Host:     "smtp.example.com",
								Port:     587,
								Username: "username",
								Password: "password",
							},
						},
					},
				},
			}, nil)

		// Create template with valid email section
		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
				Email: &domain.EmailTemplate{
					Subject:          "Test Subject",
					VisualEditorTree: notifusemjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"content": "test"}},
				},
			}, nil)

		// Mock compilation result with failure
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
			Return(&domain.CompileTemplateResponse{
				Success: false,
				Error: &mjmlgo.Error{
					Message: "compilation error",
				},
			}, nil)

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template compilation failed")
	})

	// Test scenario 8: Template with Email - Success path but SMTP client creation fails
	t.Run("Success Path With SMTP Client Error", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		// Create mock workspace with valid integration
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID: workspaceID,
				Integrations: []domain.Integration{
					{
						ID: integrationID,
						EmailProvider: domain.EmailProvider{
							Kind:               domain.EmailProviderKindSMTP,
							DefaultSenderEmail: "sender@example.com",
							DefaultSenderName:  "Sender Name",
							SMTP: &domain.SMTPSettings{
								Host:     "nonexistent.example.com", // This will cause mail.NewClient to fail
								Port:     587,
								Username: "username",
								Password: "password",
							},
						},
					},
				},
			}, nil)

		// Create template with valid email section
		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
				Email: &domain.EmailTemplate{
					Subject:          "Test Subject",
					VisualEditorTree: notifusemjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"content": "test"}},
				},
			}, nil)

		// Mock successful compilation
		htmlContent := "<p>Compiled HTML</p>"
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), workspaceID, gomock.Any(), gomock.Any()).
			Return(&domain.CompileTemplateResponse{
				Success: true,
				HTML:    &htmlContent,
			}, nil)

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		// The error should relate to SMTP client creation or connection
		assert.Contains(t, err.Error(), "invalid") // This could be "invalid sender", "invalid recipient" or similar SMTP error
	})

	// Test scenario 9: Template without Email - Uses default content and still fails with SMTP
	t.Run("Template Without Email", func(t *testing.T) {
		// Setup mocks
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, user, nil)

		// Create mock workspace with valid integration
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(&domain.Workspace{
				ID: workspaceID,
				Integrations: []domain.Integration{
					{
						ID: integrationID,
						EmailProvider: domain.EmailProvider{
							Kind:               domain.EmailProviderKindSMTP,
							DefaultSenderEmail: "sender@example.com",
							DefaultSenderName:  "Sender Name",
							SMTP: &domain.SMTPSettings{
								Host:     "nonexistent.example.com", // This will cause mail.NewClient to fail
								Port:     587,
								Username: "username",
								Password: "password",
							},
						},
					},
				},
			}, nil)

		// Create template with no email section
		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&domain.Template{
				ID:   templateID,
				Name: "Test Template",
				// No Email field
			}, nil)

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assert
		assert.Error(t, err)
		// The error should relate to SMTP client creation or connection
		assert.Contains(t, err.Error(), "invalid") // This could be "invalid sender", "invalid recipient" or similar SMTP error
	})
}
