package service

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
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMailjetResponse creates an HTTP response for mailjet tests
func mockMailjetResponse(t *testing.T, statusCode int, body interface{}) *http.Response {
	var responseBody io.ReadCloser
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		responseBody = io.NopCloser(strings.NewReader(string(jsonData)))
	} else {
		responseBody = io.NopCloser(strings.NewReader(""))
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       responseBody,
	}
}

// TestMailjetService_TestWebhook tests the TestWebhook method
func TestMailjetService_TestWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, mockLogger)

	// Test data
	ctx := context.Background()
	config := domain.MailjetSettings{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
	}
	webhookID := "123"
	eventType := "sent"

	// Call the method - need to implement TestWebhook in mailjet_service.go if it doesn't exist
	err := service.TestWebhook(ctx, config, webhookID, eventType)

	// Assertion - this should return an error since Mailjet doesn't support testing webhooks
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestMailjetService_RegisterWebhooksProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, mockLogger)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	integrationID := "integration-456"
	baseURL := "https://api.notifuse.com"
	eventTypes := []domain.EmailEventType{
		domain.EmailEventDelivered,
		domain.EmailEventBounce,
		domain.EmailEventComplaint,
	}

	t.Run("successful registration", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Expected webhook URL
		expectedWebhookURL := domain.GenerateWebhookCallbackURL(baseURL, domain.EmailProviderKindMailjet, workspaceID, integrationID)

		// Response for ListWebhooks - empty list
		emptyResponse := domain.MailjetWebhookResponse{
			Count: 0,
			Data:  []domain.MailjetWebhook{},
			Total: 0,
		}

		// Response for created webhook
		createdWebhook := domain.MailjetWebhook{
			ID:        1001,
			EventType: string(domain.MailjetEventSent),
			Endpoint:  expectedWebhookURL,
			Status:    "active",
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, emptyResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Setup mock for CreateWebhook
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "POST" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusCreated, createdWebhook), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		status, err := service.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, providerConfig)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, domain.EmailProviderKindMailjet, status.EmailProviderKind)
		assert.True(t, status.IsRegistered)
		assert.NotEmpty(t, status.Endpoints)
		assert.Equal(t, workspaceID, status.ProviderDetails["workspace_id"])
		assert.Equal(t, integrationID, status.ProviderDetails["integration_id"])

		// Check for registered event types
		var eventTypesCovered = make(map[domain.EmailEventType]bool)
		for _, endpoint := range status.Endpoints {
			eventTypesCovered[endpoint.EventType] = true
		}

		assert.True(t, eventTypesCovered[domain.EmailEventDelivered], "Should have an endpoint for delivered events")
	})

	t.Run("missing configuration", func(t *testing.T) {
		// Call with nil provider config
		status, err := service.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, nil)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")
		assert.Nil(t, status)

		// Call with empty Mailjet config
		emptyConfig := &domain.EmailProvider{
			Kind:    domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{},
		}

		status, err = service.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, emptyConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")
		assert.Nil(t, status)
	})

	t.Run("list webhooks error", func(t *testing.T) {
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Setup mock for ListWebhooks to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		status, err := service.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, providerConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list Mailjet webhooks")
		assert.Nil(t, status)
	})

	t.Run("create webhook error", func(t *testing.T) {
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Response for ListWebhooks - empty list
		emptyResponse := domain.MailjetWebhookResponse{
			Count: 0,
			Data:  []domain.MailjetWebhook{},
			Total: 0,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, emptyResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Setup mock for CreateWebhook to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "POST" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusBadRequest, nil), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		status, err := service.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, providerConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create Mailjet webhook")
		assert.Nil(t, status)
	})
}

func TestMailjetService_GetWebhookStatusProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, mockLogger)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	integrationID := "integration-456"

	t.Run("successful status check with webhooks", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Generate webhook URL
		webhookURL := domain.GenerateWebhookCallbackURL("https://api.notifuse.com", domain.EmailProviderKindMailjet, workspaceID, integrationID)

		// Response for ListWebhooks with registered webhooks
		webhooksResponse := domain.MailjetWebhookResponse{
			Count: 3,
			Data: []domain.MailjetWebhook{
				{
					ID:        101,
					EventType: string(domain.MailjetEventSent),
					Endpoint:  webhookURL,
					Status:    "active",
				},
				{
					ID:        102,
					EventType: string(domain.MailjetEventBounce),
					Endpoint:  webhookURL,
					Status:    "active",
				},
				{
					ID:        103,
					EventType: string(domain.MailjetEventSpam),
					Endpoint:  webhookURL,
					Status:    "active",
				},
			},
			Total: 3,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, webhooksResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		status, err := service.GetWebhookStatus(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, domain.EmailProviderKindMailjet, status.EmailProviderKind)
		assert.True(t, status.IsRegistered)
		assert.NotEmpty(t, status.Endpoints)

		// Verify webhooks are properly mapped to event types
		hasEventTypes := make(map[domain.EmailEventType]bool)
		for _, endpoint := range status.Endpoints {
			hasEventTypes[endpoint.EventType] = true
		}

		assert.True(t, hasEventTypes[domain.EmailEventDelivered])
	})

	t.Run("no webhooks registered", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Response for ListWebhooks with no registered webhooks
		emptyResponse := domain.MailjetWebhookResponse{
			Count: 0,
			Data:  []domain.MailjetWebhook{},
			Total: 0,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, emptyResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		status, err := service.GetWebhookStatus(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, domain.EmailProviderKindMailjet, status.EmailProviderKind)
		assert.False(t, status.IsRegistered)
		assert.Empty(t, status.Endpoints)
	})

	t.Run("missing configuration", func(t *testing.T) {
		// Call with nil provider config
		status, err := service.GetWebhookStatus(ctx, workspaceID, integrationID, nil)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")
		assert.Nil(t, status)

		// Call with empty Mailjet config
		emptyConfig := &domain.EmailProvider{
			Kind:    domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{},
		}

		status, err = service.GetWebhookStatus(ctx, workspaceID, integrationID, emptyConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")
		assert.Nil(t, status)
	})

	t.Run("list webhooks error", func(t *testing.T) {
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Setup mock for ListWebhooks to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		status, err := service.GetWebhookStatus(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list Mailjet webhooks")
		assert.Nil(t, status)
	})
}

func TestMailjetService_UnregisterWebhooksProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Create service with mocks
	service := NewMailjetService(mockHTTPClient, mockAuthService, mockLogger)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	integrationID := "integration-456"

	t.Run("successful unregistration", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Generate webhook URL
		webhookURL := domain.GenerateWebhookCallbackURL("https://api.notifuse.com", domain.EmailProviderKindMailjet, workspaceID, integrationID)

		// Response for ListWebhooks with registered webhooks
		webhooksResponse := domain.MailjetWebhookResponse{
			Count: 2,
			Data: []domain.MailjetWebhook{
				{
					ID:        101,
					EventType: string(domain.MailjetEventSent),
					Endpoint:  webhookURL,
					Status:    "active",
				},
				{
					ID:        102,
					EventType: string(domain.MailjetEventBounce),
					Endpoint:  webhookURL,
					Status:    "active",
				},
			},
			Total: 2,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, webhooksResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Setup mock for DeleteWebhook
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "DELETE" && strings.Contains(req.URL.String(), "eventcallback/101") {
					return mockMailjetResponse(t, http.StatusNoContent, nil), nil
				}
				return nil, errors.New("unexpected request")
			})

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "DELETE" && strings.Contains(req.URL.String(), "eventcallback/102") {
					return mockMailjetResponse(t, http.StatusNoContent, nil), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		err := service.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("no webhooks to unregister", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Response for ListWebhooks with no registered webhooks
		emptyResponse := domain.MailjetWebhookResponse{
			Count: 0,
			Data:  []domain.MailjetWebhook{},
			Total: 0,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, emptyResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Call the service method
		err := service.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Assertions - no error when there are no webhooks to delete
		require.NoError(t, err)
	})

	t.Run("missing configuration", func(t *testing.T) {
		// Call with nil provider config
		err := service.UnregisterWebhooks(ctx, workspaceID, integrationID, nil)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")

		// Call with empty Mailjet config
		emptyConfig := &domain.EmailProvider{
			Kind:    domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{},
		}

		err = service.UnregisterWebhooks(ctx, workspaceID, integrationID, emptyConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mailjet configuration is missing or invalid")
	})

	t.Run("list webhooks error", func(t *testing.T) {
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Setup mock for ListWebhooks to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		err := service.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list Mailjet webhooks")
	})

	t.Run("delete webhook error", func(t *testing.T) {
		// Create email provider config
		providerConfig := &domain.EmailProvider{
			Kind: domain.EmailProviderKindMailjet,
			Mailjet: &domain.MailjetSettings{
				APIKey:    "test-api-key",
				SecretKey: "test-secret-key",
			},
		}

		// Generate webhook URL
		webhookURL := domain.GenerateWebhookCallbackURL("https://api.notifuse.com", domain.EmailProviderKindMailjet, workspaceID, integrationID)

		// Response for ListWebhooks with registered webhooks
		webhooksResponse := domain.MailjetWebhookResponse{
			Count: 1,
			Data: []domain.MailjetWebhook{
				{
					ID:        101,
					EventType: string(domain.MailjetEventSent),
					Endpoint:  webhookURL,
					Status:    "active",
				},
			},
			Total: 1,
		}

		// Setup mock for ListWebhooks
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && strings.Contains(req.URL.String(), "eventcallback") {
					return mockMailjetResponse(t, http.StatusOK, webhooksResponse), nil
				}
				return nil, errors.New("unexpected request")
			})

		// Setup mock for DeleteWebhook to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, errors.New("network error"))

		// Call the service method
		err := service.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete one or more Mailjet webhooks")
	})
}
