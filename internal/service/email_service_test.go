package service

import (
	"context"
	"fmt"
	"testing"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create our own mock of EmailProviderService instead of using gomock
type mockEmailProviderService struct {
	ctrl  *gomock.Controller
	calls map[string][]interface{}
}

func (m *mockEmailProviderService) SendEmail(ctx context.Context, workspaceID string, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider) error {
	// Check if an expectation is set
	key := fmt.Sprintf("SendEmail-%s-%s-%s-%s-%s", workspaceID, fromAddress, fromName, to, subject)
	if m.calls == nil {
		m.ctrl.T.Fatalf("No expectations set for SendEmail")
		return nil
	}

	call, exists := m.calls[key]
	if !exists {
		m.ctrl.T.Fatalf("Unexpected call to SendEmail with args: %v, %v, %v, %v, %v, %v, %v",
			ctx, workspaceID, fromAddress, fromName, to, subject, content)
		return nil
	}

	// Return the error from the expectation
	if call[0] != nil {
		return call[0].(error)
	}
	return nil
}

func (m *mockEmailProviderService) expectSendEmail(ctx context.Context, workspaceID string, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider, err error) {
	// Initialize the calls map if needed
	if m.calls == nil {
		m.calls = make(map[string][]interface{})
	}

	// Store the expectation
	key := fmt.Sprintf("SendEmail-%s-%s-%s-%s-%s", workspaceID, fromAddress, fromName, to, subject)
	m.calls[key] = []interface{}{err}
}

// Helper function to create string pointer
func createStringPtr(s string) *string {
	return &s
}

func TestEmailService_TestEmailProvider(t *testing.T) {
	// Setup the controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	// Email provider services
	mockSMTPService := &mockEmailProviderService{ctrl: ctrl}
	mockSESService := &mockEmailProviderService{ctrl: ctrl}
	mockSparkPostService := &mockEmailProviderService{ctrl: ctrl}
	mockPostmarkService := &mockEmailProviderService{ctrl: ctrl}
	mockMailgunService := &mockEmailProviderService{ctrl: ctrl}
	mockMailjetService := &mockEmailProviderService{ctrl: ctrl}

	secretKey := "test-secret-key"
	webhookEndpoint := "https://webhook.test"

	// Create the email service
	emailService := EmailService{
		logger:           mockLogger,
		authService:      mockAuthService,
		secretKey:        secretKey,
		workspaceRepo:    mockWorkspaceRepo,
		templateRepo:     mockTemplateRepo,
		templateService:  mockTemplateService,
		httpClient:       mockHTTPClient,
		webhookEndpoint:  webhookEndpoint,
		smtpService:      mockSMTPService,
		sesService:       mockSESService,
		sparkPostService: mockSparkPostService,
		postmarkService:  mockPostmarkService,
		mailgunService:   mockMailgunService,
		mailjetService:   mockMailjetService,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	toEmail := "test@example.com"

	t.Run("Success with SES provider", func(t *testing.T) {
		// Create a provider for testing
		provider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SES: &domain.AmazonSESSettings{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
		}

		// Set up authentication mock
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		// Provider should send an email
		testEmailContent := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"
		mockSESService.expectSendEmail(
			ctx,
			workspaceID,
			provider.DefaultSenderEmail,
			provider.DefaultSenderName,
			toEmail,
			"Notifuse: Test Email Provider",
			testEmailContent,
			&provider,
			nil,
		)

		// Call method under test
		err := emailService.TestEmailProvider(ctx, workspaceID, provider, toEmail)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Authentication failure", func(t *testing.T) {
		provider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SES: &domain.AmazonSESSettings{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
		}

		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, nil, assert.AnError)

		// Call method under test
		err := emailService.TestEmailProvider(ctx, workspaceID, provider, toEmail)

		// Assertions
		require.Error(t, err)
	})

	t.Run("Provider validation failure", func(t *testing.T) {
		// Create an invalid provider
		provider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "", // Invalid - empty sender email
			DefaultSenderName:  "Test Sender",
			SES: &domain.AmazonSESSettings{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
		}

		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		// Call method under test
		err := emailService.TestEmailProvider(ctx, workspaceID, provider, toEmail)

		// Assertions
		require.Error(t, err)
	})

	t.Run("Email sending failure", func(t *testing.T) {
		provider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SES: &domain.AmazonSESSettings{
				Region:    "us-east-1",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
		}

		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		testEmailContent := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"
		mockSESService.expectSendEmail(
			ctx,
			workspaceID,
			provider.DefaultSenderEmail,
			provider.DefaultSenderName,
			toEmail,
			"Notifuse: Test Email Provider",
			testEmailContent,
			&provider,
			assert.AnError,
		)

		// Call method under test
		err := emailService.TestEmailProvider(ctx, workspaceID, provider, toEmail)

		// Assertions
		require.Error(t, err)
	})
}

func TestEmailService_TestTemplate(t *testing.T) {
	// Setup the controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	// Email provider services
	mockSMTPService := &mockEmailProviderService{ctrl: ctrl}
	mockSESService := &mockEmailProviderService{ctrl: ctrl}
	mockSparkPostService := &mockEmailProviderService{ctrl: ctrl}
	mockPostmarkService := &mockEmailProviderService{ctrl: ctrl}
	mockMailgunService := &mockEmailProviderService{ctrl: ctrl}
	mockMailjetService := &mockEmailProviderService{ctrl: ctrl}

	secretKey := "test-secret-key"
	webhookEndpoint := "https://webhook.test"

	// Create the email service
	emailService := EmailService{
		logger:           mockLogger,
		authService:      mockAuthService,
		secretKey:        secretKey,
		workspaceRepo:    mockWorkspaceRepo,
		templateRepo:     mockTemplateRepo,
		templateService:  mockTemplateService,
		httpClient:       mockHTTPClient,
		webhookEndpoint:  webhookEndpoint,
		smtpService:      mockSMTPService,
		sesService:       mockSESService,
		sparkPostService: mockSparkPostService,
		postmarkService:  mockPostmarkService,
		mailgunService:   mockMailgunService,
		mailjetService:   mockMailjetService,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	templateID := "template-123"
	integrationID := "integration-123"
	recipientEmail := "recipient@example.com"

	t.Run("Success with existing template", func(t *testing.T) {
		// Set up authentication mock
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		// Setup workspace with email provider
		workspace := domain.Workspace{
			ID: workspaceID,
			Integrations: []domain.Integration{
				{
					ID: integrationID,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSES,
						DefaultSenderEmail: "sender@example.com",
						DefaultSenderName:  "Test Sender",
						SES: &domain.AmazonSESSettings{
							Region:    "us-east-1",
							AccessKey: "test-access-key",
							SecretKey: "test-secret-key",
						},
					},
				},
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&workspace, nil)

		// Setup template with the correct EmailBlock structure
		editorTree := mjml.EmailBlock{
			Kind: "root",
			Data: map[string]interface{}{
				"styles": map[string]interface{}{
					"backgroundColor": "#ffffff",
				},
			},
			Children: []mjml.EmailBlock{
				{
					Kind: "paragraph",
					Children: []mjml.EmailBlock{
						{
							Kind: "text",
							Data: map[string]interface{}{
								"content": "Test content",
							},
						},
					},
				},
			},
		}

		template := domain.Template{
			ID:   templateID,
			Name: "Test Template",
			Email: &domain.EmailTemplate{
				Subject:          "Test Subject",
				VisualEditorTree: editorTree,
			},
			TestData: map[string]interface{}{
				"name": "Test User",
			},
		}

		mockTemplateRepo.EXPECT().
			GetTemplateByID(ctx, workspaceID, templateID, int64(0)).
			Return(&template, nil)

		// Setup template compilation
		htmlResult := "<html><body>Test content for Test User</body></html>"

		// Setup the mock to return the compilation result
		compilationResult := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    aws.String(htmlResult),
			Error:   nil,
		}

		mockTemplateService.EXPECT().
			CompileTemplate(
				ctx,
				workspaceID,
				editorTree,
				template.TestData,
			).Return(compilationResult, nil)

		// Provider should send an email
		mockSESService.expectSendEmail(
			ctx,
			workspaceID,
			workspace.Integrations[0].EmailProvider.DefaultSenderEmail,
			workspace.Integrations[0].EmailProvider.DefaultSenderName,
			recipientEmail,
			template.Email.Subject,
			htmlResult,
			&workspace.Integrations[0].EmailProvider,
			nil,
		)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, nil, assert.AnError)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.Error(t, err)
	})

	t.Run("Workspace not found", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(nil, assert.AnError)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.Error(t, err)
	})

	t.Run("Template not found", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		workspace := domain.Workspace{
			ID: workspaceID,
			Integrations: []domain.Integration{
				{
					ID: integrationID,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSES,
						DefaultSenderEmail: "sender@example.com",
						DefaultSenderName:  "Test Sender",
						SES: &domain.AmazonSESSettings{
							Region:    "us-east-1",
							AccessKey: "test-access-key",
							SecretKey: "test-secret-key",
						},
					},
				},
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&workspace, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(ctx, workspaceID, templateID, int64(0)).
			Return(nil, assert.AnError)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.Error(t, err)
	})

	t.Run("Integration not found", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		// Workspace without the requested integration
		workspace := domain.Workspace{
			ID: workspaceID,
			Integrations: []domain.Integration{
				{
					ID:            "different-integration",
					EmailProvider: domain.EmailProvider{},
				},
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&workspace, nil)

		template := domain.Template{
			ID:   templateID,
			Name: "Test Template",
			Email: &domain.EmailTemplate{
				Subject: "Test Subject",
			},
		}

		mockTemplateRepo.EXPECT().
			GetTemplateByID(ctx, workspaceID, templateID, int64(0)).
			Return(&template, nil)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "integration not found")
	})

	t.Run("Template compilation failure", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		workspace := domain.Workspace{
			ID: workspaceID,
			Integrations: []domain.Integration{
				{
					ID: integrationID,
					EmailProvider: domain.EmailProvider{
						Kind:               domain.EmailProviderKindSES,
						DefaultSenderEmail: "sender@example.com",
						DefaultSenderName:  "Test Sender",
						SES: &domain.AmazonSESSettings{
							Region:    "us-east-1",
							AccessKey: "test-access-key",
							SecretKey: "test-secret-key",
						},
					},
				},
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&workspace, nil)

		// Setup template with the correct EmailBlock structure
		editorTree := mjml.EmailBlock{
			Kind: "root",
			Data: map[string]interface{}{
				"styles": map[string]interface{}{
					"backgroundColor": "#ffffff",
				},
			},
			Children: []mjml.EmailBlock{
				{
					Kind: "paragraph",
					Children: []mjml.EmailBlock{
						{
							Kind: "text",
							Data: map[string]interface{}{
								"content": "Test content",
							},
						},
					},
				},
			},
		}

		template := domain.Template{
			ID:   templateID,
			Name: "Test Template",
			Email: &domain.EmailTemplate{
				Subject:          "Test Subject",
				VisualEditorTree: editorTree,
			},
			TestData: map[string]interface{}{
				"name": "Test User",
			},
		}

		mockTemplateRepo.EXPECT().
			GetTemplateByID(ctx, workspaceID, templateID, int64(0)).
			Return(&template, nil)

		// Setup the mock to return the compilation result with an error
		mockTemplateService.EXPECT().
			CompileTemplate(
				ctx,
				workspaceID,
				editorTree,
				template.TestData,
			).Return(&domain.CompileTemplateResponse{
			Success: false,
			HTML:    nil,
			Error: &mjmlgo.Error{
				Message: "Compilation failed",
			},
		}, assert.AnError)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail)

		// Assertions
		require.Error(t, err)
	})
}

func TestEmailService_SendEmail(t *testing.T) {
	// Setup the controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	// Email provider services
	mockSMTPService := &mockEmailProviderService{ctrl: ctrl}
	mockSESService := &mockEmailProviderService{ctrl: ctrl}
	mockSparkPostService := &mockEmailProviderService{ctrl: ctrl}
	mockPostmarkService := &mockEmailProviderService{ctrl: ctrl}
	mockMailgunService := &mockEmailProviderService{ctrl: ctrl}
	mockMailjetService := &mockEmailProviderService{ctrl: ctrl}

	secretKey := "test-secret-key"
	webhookEndpoint := "https://webhook.test"

	// Create the email service
	emailService := EmailService{
		logger:           mockLogger,
		authService:      mockAuthService,
		secretKey:        secretKey,
		workspaceRepo:    mockWorkspaceRepo,
		templateRepo:     mockTemplateRepo,
		templateService:  mockTemplateService,
		httpClient:       mockHTTPClient,
		webhookEndpoint:  webhookEndpoint,
		smtpService:      mockSMTPService,
		sesService:       mockSESService,
		sparkPostService: mockSparkPostService,
		postmarkService:  mockPostmarkService,
		mailgunService:   mockMailgunService,
		mailjetService:   mockMailjetService,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	toEmail := "recipient@example.com"
	subject := "Test Subject"
	content := "<html><body>Test content</body></html>"

	testCases := []struct {
		name            string
		providerKind    domain.EmailProviderKind
		mockService     *mockEmailProviderService
		isDefaultSender bool
	}{
		{
			name:            "SMTP provider",
			providerKind:    domain.EmailProviderKindSMTP,
			mockService:     mockSMTPService,
			isDefaultSender: false,
		},
		{
			name:            "SES provider",
			providerKind:    domain.EmailProviderKindSES,
			mockService:     mockSESService,
			isDefaultSender: false,
		},
		{
			name:            "SparkPost provider",
			providerKind:    domain.EmailProviderKindSparkPost,
			mockService:     mockSparkPostService,
			isDefaultSender: false,
		},
		{
			name:            "Postmark provider",
			providerKind:    domain.EmailProviderKindPostmark,
			mockService:     mockPostmarkService,
			isDefaultSender: false,
		},
		{
			name:            "Mailgun provider",
			providerKind:    domain.EmailProviderKindMailgun,
			mockService:     mockMailgunService,
			isDefaultSender: false,
		},
		{
			name:            "Mailjet provider",
			providerKind:    domain.EmailProviderKindMailjet,
			mockService:     mockMailjetService,
			isDefaultSender: false,
		},
		{
			name:            "Default sender",
			providerKind:    domain.EmailProviderKindSES,
			mockService:     mockSESService,
			isDefaultSender: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := domain.EmailProvider{
				Kind:               tc.providerKind,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
			}

			// If testing default sender, use empty values for from
			testFromAddress := fromAddress
			testFromName := fromName
			expectedFromAddress := fromAddress
			expectedFromName := fromName

			if tc.isDefaultSender {
				testFromAddress = ""
				testFromName = ""
				expectedFromAddress = provider.DefaultSenderEmail
				expectedFromName = provider.DefaultSenderName
			}

			// Set expectation for the mock service
			tc.mockService.expectSendEmail(
				ctx,
				workspaceID,
				expectedFromAddress,
				expectedFromName,
				toEmail,
				subject,
				content,
				&provider,
				nil,
			)

			// Call method under test
			err := emailService.SendEmail(ctx, workspaceID, false, testFromAddress, testFromName, toEmail, subject, content, &provider)

			// Assertions
			require.NoError(t, err)
		})
	}

	t.Run("Unsupported provider kind", func(t *testing.T) {
		provider := domain.EmailProvider{
			Kind: "unsupported",
		}

		// Call method under test
		err := emailService.SendEmail(ctx, workspaceID, false, fromAddress, fromName, toEmail, subject, content, &provider)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider kind")
	})

	t.Run("Provider service returns error", func(t *testing.T) {
		provider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "default@example.com",
			DefaultSenderName:  "Default Sender",
		}

		mockSESService.expectSendEmail(
			ctx,
			workspaceID,
			fromAddress,
			fromName,
			toEmail,
			subject,
			content,
			&provider,
			assert.AnError,
		)

		// Call method under test
		err := emailService.SendEmail(ctx, workspaceID, false, fromAddress, fromName, toEmail, subject, content, &provider)

		// Assertions
		require.Error(t, err)
	})
}

func TestEmailService_getProviderService(t *testing.T) {
	// Setup the controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

	// Email provider services
	mockSMTPService := &mockEmailProviderService{ctrl: ctrl}
	mockSESService := &mockEmailProviderService{ctrl: ctrl}
	mockSparkPostService := &mockEmailProviderService{ctrl: ctrl}
	mockPostmarkService := &mockEmailProviderService{ctrl: ctrl}
	mockMailgunService := &mockEmailProviderService{ctrl: ctrl}
	mockMailjetService := &mockEmailProviderService{ctrl: ctrl}

	// Create the email service
	emailService := EmailService{
		logger:           mockLogger,
		authService:      mockAuthService,
		workspaceRepo:    mockWorkspaceRepo,
		templateRepo:     mockTemplateRepo,
		templateService:  mockTemplateService,
		httpClient:       mockHTTPClient,
		smtpService:      mockSMTPService,
		sesService:       mockSESService,
		sparkPostService: mockSparkPostService,
		postmarkService:  mockPostmarkService,
		mailgunService:   mockMailgunService,
		mailjetService:   mockMailjetService,
	}

	tests := []struct {
		name         string
		providerKind domain.EmailProviderKind
		expected     domain.EmailProviderService
		expectError  bool
	}{
		{
			name:         "SMTP provider",
			providerKind: domain.EmailProviderKindSMTP,
			expected:     mockSMTPService,
			expectError:  false,
		},
		{
			name:         "SES provider",
			providerKind: domain.EmailProviderKindSES,
			expected:     mockSESService,
			expectError:  false,
		},
		{
			name:         "SparkPost provider",
			providerKind: domain.EmailProviderKindSparkPost,
			expected:     mockSparkPostService,
			expectError:  false,
		},
		{
			name:         "Postmark provider",
			providerKind: domain.EmailProviderKindPostmark,
			expected:     mockPostmarkService,
			expectError:  false,
		},
		{
			name:         "Mailgun provider",
			providerKind: domain.EmailProviderKindMailgun,
			expected:     mockMailgunService,
			expectError:  false,
		},
		{
			name:         "Mailjet provider",
			providerKind: domain.EmailProviderKindMailjet,
			expected:     mockMailjetService,
			expectError:  false,
		},
		{
			name:         "Unsupported provider",
			providerKind: "unsupported",
			expected:     nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			providerService, err := emailService.getProviderService(tc.providerKind)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, providerService)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, providerService)
			}
		})
	}
}
