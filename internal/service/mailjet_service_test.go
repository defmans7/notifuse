package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPResponse creates a mock HTTP response
func mockHTTPResponse(t *testing.T, statusCode int, body interface{}) *http.Response {
	var bodyReader io.ReadCloser

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		bodyReader = io.NopCloser(bytes.NewReader([]byte{}))
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       bodyReader,
	}
}

func TestMailjetService_ListWebhooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	testLogger := logger.NewLogger()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, testLogger)

	// Test data
	config := domain.MailjetSettings{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
	}

	ctx := context.Background()

	t.Run("Successfully list webhooks", func(t *testing.T) {
		// Expected response from Mailjet API
		expectedResponse := domain.MailjetWebhookResponse{
			Count: 2,
			Data: []domain.MailjetWebhook{
				{
					ID:        123,
					EventType: string(domain.MailjetEventBounce),
					Endpoint:  "https://example.com/webhook1",
					Status:    "active",
				},
				{
					ID:        456,
					EventType: string(domain.MailjetEventClick),
					Endpoint:  "https://example.com/webhook2",
					Status:    "active",
				},
			},
			Total: 2,
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3/eventcallback", req.URL.String())

				// Verify auth header
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, config.APIKey, username)
				assert.Equal(t, config.SecretKey, password)

				return mockHTTPResponse(t, http.StatusOK, expectedResponse), nil
			})

		// Call the service method
		response, err := service.ListWebhooks(ctx, config)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 2, response.Count)
		assert.Equal(t, 2, len(response.Data))
		assert.Equal(t, int64(123), response.Data[0].ID)
		assert.Equal(t, string(domain.MailjetEventBounce), response.Data[0].EventType)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		response, err := service.ListWebhooks(ctx, config)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
		assert.Nil(t, response)
	})

	t.Run("API returns error status code", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusUnauthorized, nil), nil)

		// Call the service method
		response, err := service.ListWebhooks(ctx, config)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code")
		assert.Nil(t, response)
	})

	t.Run("Malformed response", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusOK, "not a valid json"), nil)

		// Call the service method
		response, err := service.ListWebhooks(ctx, config)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
		assert.Nil(t, response)
	})
}

func TestMailjetService_CreateWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	testLogger := logger.NewLogger()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, testLogger)

	// Test data
	config := domain.MailjetSettings{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
	}

	webhookToCreate := domain.MailjetWebhook{
		EventType: string(domain.MailjetEventBounce),
		Endpoint:  "https://example.com/webhook",
		Status:    "active",
	}

	ctx := context.Background()

	t.Run("Successfully create webhook", func(t *testing.T) {
		// Expected response from Mailjet API
		expectedResponse := domain.MailjetWebhook{
			ID:        123,
			EventType: string(domain.MailjetEventBounce),
			Endpoint:  "https://example.com/webhook",
			Status:    "active",
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3/eventcallback", req.URL.String())

				// Verify auth header
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, config.APIKey, username)
				assert.Equal(t, config.SecretKey, password)

				// Verify request body
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var sentWebhook domain.MailjetWebhook
				err = json.Unmarshal(body, &sentWebhook)
				require.NoError(t, err)

				assert.Equal(t, webhookToCreate.EventType, sentWebhook.EventType)
				assert.Equal(t, webhookToCreate.Endpoint, sentWebhook.Endpoint)
				assert.Equal(t, webhookToCreate.Status, sentWebhook.Status)

				return mockHTTPResponse(t, http.StatusCreated, expectedResponse), nil
			})

		// Call the service method
		response, err := service.CreateWebhook(ctx, config, webhookToCreate)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, int64(123), response.ID)
		assert.Equal(t, webhookToCreate.EventType, response.EventType)
		assert.Equal(t, webhookToCreate.Endpoint, response.Endpoint)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		response, err := service.CreateWebhook(ctx, config, webhookToCreate)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
		assert.Nil(t, response)
	})

	t.Run("API returns error status code", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusBadRequest, nil), nil)

		// Call the service method
		response, err := service.CreateWebhook(ctx, config, webhookToCreate)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code")
		assert.Nil(t, response)
	})
}

func TestMailjetService_GetWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	testLogger := logger.NewLogger()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, testLogger)

	// Test data
	config := domain.MailjetSettings{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
	}

	webhookID := int64(123)
	ctx := context.Background()

	t.Run("Successfully get webhook", func(t *testing.T) {
		// Expected response from Mailjet API
		responseData := struct {
			Data []domain.MailjetWebhook `json:"Data"`
		}{
			Data: []domain.MailjetWebhook{
				{
					ID:        123,
					EventType: string(domain.MailjetEventBounce),
					Endpoint:  "https://example.com/webhook",
					Status:    "active",
				},
			},
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3/eventcallback/123", req.URL.String())

				// Verify auth header
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, config.APIKey, username)
				assert.Equal(t, config.SecretKey, password)

				return mockHTTPResponse(t, http.StatusOK, responseData), nil
			})

		// Call the service method
		response, err := service.GetWebhook(ctx, config, webhookID)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, webhookID, response.ID)
		assert.Equal(t, string(domain.MailjetEventBounce), response.EventType)
		assert.Equal(t, "https://example.com/webhook", response.Endpoint)
	})

	t.Run("Webhook not found", func(t *testing.T) {
		// Empty response data
		responseData := struct {
			Data []domain.MailjetWebhook `json:"Data"`
		}{
			Data: []domain.MailjetWebhook{},
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusOK, responseData), nil)

		// Call the service method
		response, err := service.GetWebhook(ctx, config, webhookID)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "webhook with ID 123 not found")
		assert.Nil(t, response)
	})
}

func TestMailjetService_DeleteWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	testLogger := logger.NewLogger()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, testLogger)

	// Test data
	config := domain.MailjetSettings{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
	}

	webhookID := int64(123)
	ctx := context.Background()

	t.Run("Successfully delete webhook", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "DELETE", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3/eventcallback/123", req.URL.String())

				// Verify auth header
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, config.APIKey, username)
				assert.Equal(t, config.SecretKey, password)

				return mockHTTPResponse(t, http.StatusNoContent, nil), nil
			})

		// Call the service method
		err := service.DeleteWebhook(ctx, config, webhookID)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("API returns error status code", func(t *testing.T) {
		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusNotFound, nil), nil)

		// Call the service method
		err := service.DeleteWebhook(ctx, config, webhookID)

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code")
	})
}

func TestMailjetService_SendEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	testLogger := logger.NewLogger()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, testLogger)

	// Test data
	workspaceID := "workspace-123"
	messageID := "message-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Email Content</p>"

	t.Run("Successfully send email", func(t *testing.T) {
		ctx := context.Background()

		// Create provider config
		provider := &domain.EmailProvider{
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Expected response from Mailjet API
		expectedResponse := map[string]interface{}{
			"Messages": []map[string]interface{}{
				{
					"Status": "success",
					"To": []map[string]interface{}{
						{
							"Email":       to,
							"MessageID":   "message-id-123",
							"MessageUUID": "uuid-123",
						},
					},
					"CustomID": messageID,
				},
			},
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.mailjet.com/v3.1/send", req.URL.String())

				// Verify auth header
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, provider.Mailjet.APIKey, username)
				assert.Equal(t, provider.Mailjet.SecretKey, password)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				// Verify request body
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				var emailReq map[string]interface{}
				err = json.Unmarshal(body, &emailReq)
				require.NoError(t, err)

				messages, ok := emailReq["Messages"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, messages, 1)

				message := messages[0].(map[string]interface{})
				from := message["From"].(map[string]interface{})
				assert.Equal(t, fromAddress, from["Email"])
				assert.Equal(t, fromName, from["Name"])

				recipients := message["To"].([]interface{})
				assert.Len(t, recipients, 1)
				assert.Equal(t, to, recipients[0].(map[string]interface{})["Email"])

				assert.Equal(t, subject, message["Subject"])
				assert.Equal(t, content, message["HTMLPart"])
				assert.Equal(t, messageID, message["CustomID"])

				return mockHTTPResponse(t, http.StatusOK, expectedResponse), nil
			})

		// Call the service method
		err := service.SendEmail(ctx, workspaceID, messageID, fromAddress, fromName, to, subject, content, provider, domain.EmailOptions{})

		// Assertions
		require.NoError(t, err)
	})

	t.Run("Missing Mailjet configuration", func(t *testing.T) {
		ctx := context.Background()

		// Create provider without Mailjet config
		provider := &domain.EmailProvider{}

		// Call the service method
		err := service.SendEmail(ctx, workspaceID, "test-message-id", fromAddress, fromName, to, subject, content, provider, domain.EmailOptions{})

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet provider is not configured")
	})

	t.Run("HTTP client error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("network error")

		// Create provider config
		provider := &domain.EmailProvider{
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, expectedErr)

		// Call the service method
		err := service.SendEmail(ctx, workspaceID, "test-message-id", fromAddress, fromName, to, subject, content, provider, domain.EmailOptions{})

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
	})

	t.Run("API returns error status code", func(t *testing.T) {
		ctx := context.Background()

		// Create provider config
		provider := &domain.EmailProvider{
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Error response
		errorResp := map[string]interface{}{
			"ErrorMessage": "Invalid recipient",
			"ErrorCode":    "mj-0002",
		}

		// Setup mock expectations
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(t, http.StatusBadRequest, errorResp), nil)

		// Call the service method
		err := service.SendEmail(ctx, workspaceID, "test-message-id", fromAddress, fromName, to, subject, content, provider, domain.EmailOptions{})

		// Assertions
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code 400")
	})
}
