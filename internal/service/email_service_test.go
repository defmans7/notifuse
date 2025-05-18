package service

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func (m *mockEmailProviderService) SendEmail(ctx context.Context, workspaceID string, messageID string, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider, replyTo string, cc []string, bcc []string) error {
	// Check if an expectation is set
	key := fmt.Sprintf("SendEmail-%s-%s-%s-%s-%s-%s-%s", workspaceID, messageID, fromAddress, fromName, to, subject, replyTo)
	if m.calls == nil {
		m.ctrl.T.Fatalf("No expectations set for SendEmail")
		return nil
	}

	call, exists := m.calls[key]
	if !exists {
		// Print the keys that exist in the map for debugging
		availableKeys := make([]string, 0, len(m.calls))
		for k := range m.calls {
			availableKeys = append(availableKeys, k)
		}
		m.ctrl.T.Fatalf("Unexpected call to SendEmail with calculated key: %s, available keys: %v, args: %v, %v, %v, %v, %v, %v, %v",
			key, availableKeys, ctx, workspaceID, messageID, fromAddress, fromName, to, subject)
		return nil
	}

	// Return the error from the expectation
	if call[0] != nil {
		return call[0].(error)
	}
	return nil
}

func (m *mockEmailProviderService) expectSendEmailWithOptions(ctx context.Context, workspaceID string, messageID string, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider, replyTo string, cc []string, bcc []string, err error) {
	// Initialize the calls map if needed
	if m.calls == nil {
		m.calls = make(map[string][]interface{})
	}

	// Store the expectation using the messageID in the key
	key := fmt.Sprintf("SendEmail-%s-%s-%s-%s-%s-%s-%s", workspaceID, messageID, fromAddress, fromName, to, subject, replyTo)
	m.calls[key] = []interface{}{err}
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
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		// Provider should send an email
		testEmailContent := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			"test-message-id",
			provider.DefaultSenderEmail,
			provider.DefaultSenderName,
			toEmail,
			"Notifuse: Test Email Provider",
			testEmailContent,
			&provider,
			"",
			nil,
			nil,
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
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
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
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
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
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, &domain.User{ID: "user-123"}, nil)

		testEmailContent := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			"test-message-id",
			provider.DefaultSenderEmail,
			provider.DefaultSenderName,
			toEmail,
			"Notifuse: Test Email Provider",
			testEmailContent,
			&provider,
			"",
			nil,
			nil,
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
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
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
			GetByID(gomock.Any(), workspaceID).
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
				ReplyTo:          "reply-to@example.com",
			},
			TestData: map[string]interface{}{
				"name": "Test User",
			},
		}

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(&template, nil)

		// Setup the mock to return the compilation result
		htmlResult := "<html><body>Test content for Test User</body></html>"
		compilationResult := &domain.CompileTemplateResponse{
			Success: true,
			HTML:    aws.String(htmlResult),
			Error:   nil,
		}

		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compilationResult, nil)

		// Initialize the calls map for mockSESService if it's nil
		if mockSESService.calls == nil {
			mockSESService.calls = make(map[string][]interface{})
		}

		// Provider should send an email
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			"test-message-id",
			workspace.Integrations[0].EmailProvider.DefaultSenderEmail,
			workspace.Integrations[0].EmailProvider.DefaultSenderName,
			recipientEmail,
			template.Email.Subject,
			htmlResult,
			&workspace.Integrations[0].EmailProvider,
			template.Email.ReplyTo,
			nil,
			nil,
			nil,
		)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail, nil, nil, template.Email.ReplyTo)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(ctx, nil, assert.AnError)

		// Call method under test
		err := emailService.TestTemplate(ctx, workspaceID, templateID, integrationID, recipientEmail, nil, nil, "")

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
		replyTo         string
		cc              []string
		bcc             []string
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
		{
			name:            "With ReplyTo and CC/BCC",
			providerKind:    domain.EmailProviderKindSES,
			mockService:     mockSESService,
			isDefaultSender: false,
			replyTo:         "reply@example.com",
			cc:              []string{"cc1@example.com", "cc2@example.com"},
			bcc:             []string{"bcc@example.com"},
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
			messageID := "test-message-id"

			if tc.isDefaultSender {
				testFromAddress = ""
				testFromName = ""
				expectedFromAddress = provider.DefaultSenderEmail
				expectedFromName = provider.DefaultSenderName
			}

			// Set expectation for the mock service
			tc.mockService.expectSendEmailWithOptions(
				ctx,
				workspaceID,
				messageID,
				expectedFromAddress,
				expectedFromName,
				toEmail,
				subject,
				content,
				&provider,
				tc.replyTo,
				tc.cc,
				tc.bcc,
				nil,
			)

			// Call method under test
			err := emailService.SendEmail(ctx, workspaceID, messageID, false, testFromAddress, testFromName, toEmail, subject, content, &provider, tc.replyTo, tc.cc, tc.bcc)

			// Assertions
			require.NoError(t, err)
		})
	}

	t.Run("Unsupported provider kind", func(t *testing.T) {
		provider := domain.EmailProvider{
			Kind: "unsupported",
		}
		messageID := "test-message-id"

		// Call method under test
		err := emailService.SendEmail(ctx, workspaceID, messageID, false, fromAddress, fromName, toEmail, subject, content, &provider, "", nil, nil)

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
		messageID := "test-message-id"

		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			messageID,
			fromAddress,
			fromName,
			toEmail,
			subject,
			content,
			&provider,
			"",
			nil,
			nil,
			assert.AnError,
		)

		// Call method under test
		err := emailService.SendEmail(ctx, workspaceID, messageID, false, fromAddress, fromName, toEmail, subject, content, &provider, "", nil, nil)

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

func TestEmailService_VisitLink(t *testing.T) {
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
	mockMessageRepo := mocks.NewMockMessageHistoryRepository(ctrl)

	// Create the email service
	emailService := EmailService{
		logger:          mockLogger,
		authService:     mockAuthService,
		workspaceRepo:   mockWorkspaceRepo,
		templateRepo:    mockTemplateRepo,
		templateService: mockTemplateService,
		httpClient:      mockHTTPClient,
		messageRepo:     mockMessageRepo,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "message-456"

	t.Run("Successfully sets message as clicked", func(t *testing.T) {
		// Setup message repository mock to expect SetClicked
		mockMessageRepo.EXPECT().
			SetClicked(ctx, workspaceID, messageID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, timestamp time.Time) error {
				// Verify the timestamp is close to now
				assert.True(t, time.Now().Sub(timestamp) < time.Second)
				return nil
			})

		// No logger error expected

		// Call method under test
		err := emailService.VisitLink(ctx, messageID, workspaceID)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Error setting clicked status", func(t *testing.T) {
		// Setup message repository mock to return an error
		mockMessageRepo.EXPECT().
			SetClicked(ctx, workspaceID, messageID, gomock.Any()).
			Return(assert.AnError)

		// Should log the error
		mockLogger.EXPECT().Error(gomock.Any())

		// Call method under test
		err := emailService.VisitLink(ctx, messageID, workspaceID)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set clicked")
	})
}

func TestEmailService_OpenEmail(t *testing.T) {
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
	mockMessageRepo := mocks.NewMockMessageHistoryRepository(ctrl)

	// Create the email service
	emailService := EmailService{
		logger:          mockLogger,
		authService:     mockAuthService,
		workspaceRepo:   mockWorkspaceRepo,
		templateRepo:    mockTemplateRepo,
		templateService: mockTemplateService,
		httpClient:      mockHTTPClient,
		messageRepo:     mockMessageRepo,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "message-456"

	t.Run("Successfully sets message as opened", func(t *testing.T) {
		// Setup message repository mock to expect SetOpened
		mockMessageRepo.EXPECT().
			SetOpened(ctx, workspaceID, messageID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, timestamp time.Time) error {
				// Verify the timestamp is close to now
				assert.True(t, time.Now().Sub(timestamp) < time.Second)
				return nil
			})

		// No logger error expected

		// Call method under test
		err := emailService.OpenEmail(ctx, messageID, workspaceID)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Error setting opened status", func(t *testing.T) {
		// Setup message repository mock to return an error
		mockMessageRepo.EXPECT().
			SetOpened(ctx, workspaceID, messageID, gomock.Any()).
			Return(assert.AnError)

		// Setup logger mock to expect Error call
		mockLoggerWithFields := pkgmocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().
			WithFields(gomock.Any()).
			Return(mockLoggerWithFields).
			AnyTimes()
		mockLoggerWithFields.EXPECT().
			Error(gomock.Any()).
			AnyTimes()

		// Call method under test
		err := emailService.OpenEmail(ctx, messageID, workspaceID)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update message opened")
	})
}

func TestEmailService_SendEmailForTemplate(t *testing.T) {
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
	mockMessageRepo := mocks.NewMockMessageHistoryRepository(ctrl)

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
		messageRepo:      mockMessageRepo,
		webhookEndpoint:  "https://webhook.test",
		smtpService:      mockSMTPService,
		sesService:       mockSESService,
		sparkPostService: mockSparkPostService,
		postmarkService:  mockPostmarkService,
		mailgunService:   mockMailgunService,
		mailjetService:   mockMailjetService,
	}

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "message-456"

	// Create a contact
	contact := &domain.Contact{
		Email:     "test@example.com",
		FirstName: &domain.NullableString{String: "Test", IsNull: false},
		LastName:  &domain.NullableString{String: "User", IsNull: false},
	}

	// Create template config
	templateConfig := domain.ChannelTemplate{
		TemplateID: "template-789",
	}

	// Create message data
	messageData := domain.MessageData{
		Data: map[string]interface{}{
			"name": "Test User",
			"link": "https://example.com/test",
		},
	}

	// Create tracking settings
	trackingSettings := mjml.TrackingSettings{
		Endpoint:       "https://track.example.com",
		EnableTracking: true,
		UTMSource:      "newsletter",
		UTMMedium:      "email",
		UTMCampaign:    "welcome",
		UTMContent:     "template-789",
		UTMTerm:        "new-user",
	}

	// Create email provider
	emailProvider := &domain.EmailProvider{
		Kind:               domain.EmailProviderKindSES,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Sender Name",
		SES: &domain.AmazonSESSettings{
			Region:    "us-east-1",
			AccessKey: "access-key",
			SecretKey: "secret-key",
		},
	}

	// Set up common mock expectations for logger
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create email template
	emailTemplate := &domain.Template{
		ID:   "template-789",
		Name: "Welcome Email",
		Email: &domain.EmailTemplate{
			Subject:          "Welcome to Our Service",
			FromAddress:      "welcome@example.com",
			FromName:         "Welcome Team",
			ReplyTo:          "support@example.com",
			VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
		},
	}

	// Create compile template result
	compiledHTML := "<h1>Welcome!</h1><p>Hello Test User, welcome to our service!</p>"
	compileResult := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	t.Run("Successfully sends email template", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Setup compile template mock
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compileResult, nil)

		// Setup message repository mock
		mockMessageRepo.EXPECT().
			Create(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(_ context.Context, wsID string, msgHistory *domain.MessageHistory) error {
				// Verify message history properties
				assert.Equal(t, messageID, msgHistory.ID)
				assert.Equal(t, contact.Email, msgHistory.ContactEmail)
				assert.Equal(t, templateConfig.TemplateID, msgHistory.TemplateID)
				assert.Equal(t, "email", msgHistory.Channel)
				assert.Equal(t, domain.MessageStatusSent, msgHistory.Status)
				assert.Equal(t, messageData, msgHistory.MessageData)

				return nil
			})

		// Setup email provider mock
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			messageID,
			emailTemplate.Email.FromAddress,
			emailTemplate.Email.FromName,
			contact.Email,
			emailTemplate.Email.Subject,
			compiledHTML,
			emailProvider,
			emailTemplate.Email.ReplyTo,
			nil, // cc
			nil, // bcc
			nil, // no error
		)

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Error getting template", func(t *testing.T) {
		// Setup template service mock to return an error
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(nil, assert.AnError)

		// Logger should log the error
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get template")
	})

	t.Run("Error compiling template", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Setup compile template mock to return an error
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		// Logger should log the error
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile template")
	})

	t.Run("Template compilation unsuccessful", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Create unsuccessful compile result
		unsuccessfulResult := &domain.CompileTemplateResponse{
			Success: false,
			Error: &mjmlgo.Error{
				Message: "Template compilation error",
			},
		}

		// Setup compile template mock to return unsuccessful result
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(unsuccessfulResult, nil)

		// Logger should log the error
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template compilation failed")
	})

	t.Run("Error creating message history", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Setup compile template mock
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compileResult, nil)

		// Setup message repository mock to return an error
		mockMessageRepo.EXPECT().
			Create(gomock.Any(), workspaceID, gomock.Any()).
			Return(assert.AnError)

		// Logger should log the error
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create message history")
	})

	t.Run("Error sending email", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Setup compile template mock
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compileResult, nil)

		// Setup message repository mock
		mockMessageRepo.EXPECT().
			Create(gomock.Any(), workspaceID, gomock.Any()).
			Return(nil)

		// Setup email provider mock to return an error
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			messageID,
			emailTemplate.Email.FromAddress,
			emailTemplate.Email.FromName,
			contact.Email,
			emailTemplate.Email.Subject,
			compiledHTML,
			emailProvider,
			emailTemplate.Email.ReplyTo,
			nil, // cc
			nil, // bcc
			assert.AnError,
		)

		// Setup message repository mock to update with error status
		mockMessageRepo.EXPECT().
			Update(gomock.Any(), workspaceID, gomock.Any()).
			DoAndReturn(func(_ context.Context, wsID string, msgHistory *domain.MessageHistory) error {
				// Verify message history error properties
				assert.Equal(t, messageID, msgHistory.ID)
				assert.Equal(t, domain.MessageStatusFailed, msgHistory.Status)
				assert.NotNil(t, msgHistory.Error)

				return nil
			})

		// Logger should log the error
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})

	t.Run("Error updating message history after failed email", func(t *testing.T) {
		// Setup template service mock
		mockTemplateService.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateConfig.TemplateID, int64(0)).
			Return(emailTemplate, nil)

		// Setup compile template mock
		mockTemplateService.EXPECT().
			CompileTemplate(gomock.Any(), gomock.Any()).
			Return(compileResult, nil)

		// Setup message repository mock
		mockMessageRepo.EXPECT().
			Create(gomock.Any(), workspaceID, gomock.Any()).
			Return(nil)

		// Setup email provider mock to return an error
		mockSESService.expectSendEmailWithOptions(
			ctx,
			workspaceID,
			messageID,
			emailTemplate.Email.FromAddress,
			emailTemplate.Email.FromName,
			contact.Email,
			emailTemplate.Email.Subject,
			compiledHTML,
			emailProvider,
			emailTemplate.Email.ReplyTo,
			nil, // cc
			nil, // bcc
			assert.AnError,
		)

		// Setup message repository mock to fail updating with error status
		mockMessageRepo.EXPECT().
			Update(gomock.Any(), workspaceID, gomock.Any()).
			Return(assert.AnError)

		// Logger should log both errors
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Call method under test
		err := emailService.SendEmailForTemplate(
			ctx,
			workspaceID,
			messageID,
			contact,
			templateConfig,
			messageData,
			trackingSettings,
			emailProvider,
			nil, // cc
			nil, // bcc
		)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})
}
