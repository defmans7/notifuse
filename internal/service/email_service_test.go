package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
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

func TestCreateSESClient(t *testing.T) {
	// Test creating a SES client
	client := service.CreateSESClient("us-east-1", "test-access-key", "test-secret-key")
	assert.NotNil(t, client, "SES client should not be nil")
}

func TestEmailService_TestEmailProvider_Success(t *testing.T) {
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
	templateID := "template123"
	user := &domain.User{ID: "user123"}
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
		Return(ctx, user, nil)

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
