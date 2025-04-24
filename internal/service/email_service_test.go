package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// MockHTTPResponse is a helper to create mock HTTP responses
type MockHTTPResponse struct {
	StatusCode int
	Body       string
}

func (m MockHTTPResponse) ToResponse() *http.Response {
	return &http.Response{
		StatusCode: m.StatusCode,
		Body:       io.NopCloser(strings.NewReader(m.Body)),
		Header:     make(http.Header),
	}
}

func TestEmailService_TestEmailProvider(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockLogger := mocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockMJMLRenderer := mocks.NewMockMJMLRenderer(ctrl)
	mockSESClient := mocks.NewMockSESClient(ctrl)

	// Mock createSESClient function
	createSESClient := func(region, accessKey, secretKey string) domain.SESClient {
		return mockSESClient
	}

	// Create the service with mocked dependencies
	emailService := service.NewEmailServiceWithDependencies(
		mockLogger,
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockHTTPClient,
		mockMJMLRenderer,
		createSESClient,
	)

	// Create a test context
	ctx := context.Background()

	// Test case for SES provider
	t.Run("TestSESProvider", func(t *testing.T) {
		// Define test inputs
		workspaceID := "workspace-1"
		to := "test@example.com"
		sesProvider := domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES: &domain.AmazonSES{
				Region:    "us-west-2",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
		}

		// Mock authentication
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: "user-1"}, nil)

		// Expect SES client to be called with the test email parameters
		mockSESClient.EXPECT().
			SendEmail(gomock.Any()).
			DoAndReturn(func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
				// Verify the input parameters
				assert.Equal(t, *input.Source, "sender@example.com")
				assert.Equal(t, *input.Destination.ToAddresses[0], to)
				assert.Contains(t, *input.Message.Subject.Data, "Test Email Provider")

				// Return success
				return &ses.SendEmailOutput{}, nil
			})

		// Call the method
		err := emailService.TestEmailProvider(ctx, workspaceID, sesProvider, to)

		// Assert no errors
		assert.NoError(t, err)
	})

	// Test case for SparkPost provider
	t.Run("TestSparkPostProvider", func(t *testing.T) {
		// Define test inputs
		workspaceID := "workspace-1"
		to := "test@example.com"
		sparkPostProvider := domain.EmailProvider{
			Kind: domain.EmailProviderKindSparkPost,
			SparkPost: &domain.SparkPostSettings{
				APIKey:      "test-api-key",
				Endpoint:    "https://api.sparkpost.com",
				SandboxMode: false,
			},
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
		}

		// Mock authentication
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: "user-1"}, nil)

		// Expect HTTP client to be called with the SparkPost API request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify the request
				assert.Equal(t, req.URL.String(), "https://api.sparkpost.com/api/v1/transmissions")
				assert.Equal(t, req.Method, "POST")
				assert.Equal(t, req.Header.Get("Authorization"), "test-api-key")

				// Parse the request body
				body, _ := io.ReadAll(req.Body)
				var payload map[string]interface{}
				json.Unmarshal(body, &payload)

				// Verify payload
				content := payload["content"].(map[string]interface{})
				from := content["from"].(map[string]interface{})
				assert.Equal(t, from["email"], "sender@example.com")
				assert.Equal(t, from["name"], "Test Sender")

				// Return success response
				return MockHTTPResponse{
					StatusCode: 200,
					Body:       `{"results": {"total_accepted_recipients": 1}}`,
				}.ToResponse(), nil
			})

		// Call the method
		err := emailService.TestEmailProvider(ctx, workspaceID, sparkPostProvider, to)

		// Assert no errors
		assert.NoError(t, err)
	})

	// Test case with authentication error
	t.Run("AuthenticationError", func(t *testing.T) {
		// Define test inputs
		workspaceID := "workspace-1"
		to := "test@example.com"
		provider := domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES: &domain.AmazonSES{
				Region:    "us-west-2",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
		}

		// Mock authentication error
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, errors.New("authentication failed"))

		// Call the method
		err := emailService.TestEmailProvider(ctx, workspaceID, provider, to)

		// Assert error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})
}

func TestEmailService_SendTemplateEmail(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockLogger := mocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockMJMLRenderer := mocks.NewMockMJMLRenderer(ctrl)
	mockSESClient := mocks.NewMockSESClient(ctrl)

	// Mock createSESClient function
	createSESClient := func(region, accessKey, secretKey string) domain.SESClient {
		return mockSESClient
	}

	// Create the service with mocked dependencies
	emailService := service.NewEmailServiceWithDependencies(
		mockLogger,
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockHTTPClient,
		mockMJMLRenderer,
		createSESClient,
	)

	// Create a test context
	ctx := context.Background()

	// Test case for TestTemplate with SES provider
	t.Run("TestTemplateWithSES", func(t *testing.T) {
		// Define test inputs
		workspaceID := "workspace-1"
		templateID := "template-1"
		providerType := "transactional"
		recipientEmail := "test@example.com"

		// Create test workspace with SES provider
		workspace := &domain.Workspace{
			ID: workspaceID,
			Settings: domain.WorkspaceSettings{
				EmailTransactionalProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSES,
					SES: &domain.AmazonSES{
						Region:    "us-west-2",
						AccessKey: "test-access-key",
						SecretKey: "test-secret-key",
					},
					DefaultSenderEmail: "sender@example.com",
					DefaultSenderName:  "Test Sender",
				},
			},
		}

		// Create test template
		template := &domain.Template{
			ID:   templateID,
			Name: "Test Template",
			Email: &domain.EmailTemplate{
				Subject:         "Test Subject",
				CompiledPreview: "<html><body><h1>Test Template</h1></body></html>",
			},
		}

		// Mock authentication
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: "user-1"}, nil)

		// Mock workspace repository
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Mock template repository
		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, gomock.Any()).
			Return(template, nil)

		// Expect SES client to be called with the template content
		mockSESClient.EXPECT().
			SendEmail(gomock.Any()).
			DoAndReturn(func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
				// Verify the input parameters
				assert.Equal(t, *input.Source, "sender@example.com")
				assert.Equal(t, *input.Destination.ToAddresses[0], recipientEmail)
				assert.Equal(t, *input.Message.Subject.Data, "Test Subject")
				assert.Equal(t, *input.Message.Body.Html.Data, template.Email.CompiledPreview)

				// Return success
				return &ses.SendEmailOutput{}, nil
			})

		// Call the method
		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)

		// Assert no errors
		assert.NoError(t, err)
	})

	// Add more test cases as needed
}

func TestEmailService_SendEmail(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockLogger := mocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockMJMLRenderer := mocks.NewMockMJMLRenderer(ctrl)
	mockSESClient := mocks.NewMockSESClient(ctrl)

	// Mock createSESClient function
	createSESClient := func(region, accessKey, secretKey string) domain.SESClient {
		return mockSESClient
	}

	// Create the service with mocked dependencies
	emailService := service.NewEmailServiceWithDependencies(
		mockLogger,
		mockAuthService,
		"test-secret-key",
		mockWorkspaceRepo,
		mockTemplateRepo,
		mockHTTPClient,
		mockMJMLRenderer,
		createSESClient,
	)

	// Create a test context
	ctx := context.Background()

	// Test case for SendEmail with Postmark provider
	t.Run("SendEmailWithPostmark", func(t *testing.T) {
		// Define test inputs
		workspaceID := "workspace-1"
		providerType := "transactional"
		from := "sender@example.com"
		to := "recipient@example.com"
		subject := "Test Email"
		content := "<h1>Test Email Content</h1>"

		// Create test workspace with Postmark provider
		workspace := &domain.Workspace{
			ID: workspaceID,
			Settings: domain.WorkspaceSettings{
				EmailTransactionalProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindPostmark,
					Postmark: &domain.PostmarkSettings{
						ServerToken: "test-server-token",
					},
					DefaultSenderEmail: "default@example.com",
					DefaultSenderName:  "Default Sender",
				},
			},
		}

		// Mock authentication
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: "user-1"}, nil)

		// Mock workspace repository
		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		// Expect HTTP client to be called with the Postmark API request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify the request
				assert.Equal(t, req.URL.String(), "https://api.postmarkapp.com/email")
				assert.Equal(t, req.Method, "POST")
				assert.Equal(t, req.Header.Get("X-Postmark-Server-Token"), "test-server-token")

				// Parse the request body
				body, _ := io.ReadAll(req.Body)
				var payload map[string]interface{}
				json.Unmarshal(body, &payload)

				// Verify payload
				assert.Equal(t, payload["From"], from)
				assert.Equal(t, payload["To"], to)
				assert.Equal(t, payload["Subject"], subject)
				assert.Equal(t, payload["HtmlBody"], content)

				// Return success response
				return MockHTTPResponse{
					StatusCode: 200,
					Body:       `{"MessageID": "test-message-id"}`,
				}.ToResponse(), nil
			})

		// Call the method
		err := emailService.SendEmail(ctx, workspaceID, providerType, from, to, subject, content)

		// Assert no errors
		assert.NoError(t, err)
	})

	// Add more test cases as needed
}
