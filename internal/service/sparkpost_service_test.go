package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

// mockHTTPResponse creates a mock HTTP response
func mockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}

func TestSparkPostService_ListWebhooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test configuration
	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Mock successful response
		webhookListResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{
				{
					ID:     "webhook-1",
					Name:   "Test Webhook",
					Target: "https://example.com/webhook",
					Events: []string{"delivery", "bounce"},
					Active: true,
				},
			},
		}
		responseJSON, _ := json.Marshal(webhookListResponse)

		// Expect HTTP request and return mocked response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))

				return mockHTTPResponse(http.StatusOK, string(responseJSON)), nil
			})

		// Call the service method
		result, err := sparkPostService.ListWebhooks(ctx, config)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "webhook-1", result.Results[0].ID)
		assert.Equal(t, "Test Webhook", result.Results[0].Name)
	})

	t.Run("HTTP request error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Mock HTTP client to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, expectedErr)

		// Call the service method
		result, err := sparkPostService.ListWebhooks(ctx, config)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to execute request")
	})

	t.Run("Non-OK status code", func(t *testing.T) {
		ctx := context.Background()

		// Mock error response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusUnauthorized, `{"errors":[{"message":"Unauthorized"}]}`), nil)

		// Call the service method
		result, err := sparkPostService.ListWebhooks(ctx, config)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "API returned non-OK status code 401")
	})

	t.Run("Invalid response body", func(t *testing.T) {
		ctx := context.Background()

		// Mock invalid JSON response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusOK, `invalid json`), nil)

		// Call the service method
		result, err := sparkPostService.ListWebhooks(ctx, config)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to decode response")
	})
}

func TestSparkPostService_CreateWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test configuration
	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}

	webhook := domain.SparkPostWebhook{
		Name:     "Test Webhook",
		Target:   "https://example.com/webhook",
		Events:   []string{"delivery", "bounce"},
		Active:   true,
		AuthType: "none",
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Mock successful response
		webhookResponse := domain.SparkPostWebhookResponse{
			Results: domain.SparkPostWebhook{
				ID:     "webhook-123",
				Name:   webhook.Name,
				Target: webhook.Target,
				Events: webhook.Events,
				Active: webhook.Active,
			},
		}
		responseJSON, _ := json.Marshal(webhookResponse)

		// Expect HTTP request and return mocked response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))

				// Verify request body
				var requestBody domain.SparkPostWebhook
				body, _ := io.ReadAll(req.Body)
				_ = json.Unmarshal(body, &requestBody)
				assert.Equal(t, webhook.Name, requestBody.Name)
				assert.Equal(t, webhook.Target, requestBody.Target)

				return mockHTTPResponse(http.StatusOK, string(responseJSON)), nil
			})

		// Call the service method
		result, err := sparkPostService.CreateWebhook(ctx, config, webhook)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "webhook-123", result.Results.ID)
		assert.Equal(t, webhook.Name, result.Results.Name)
	})
}

func TestSparkPostService_GetWebhookStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test configuration
	providerConfig := &domain.EmailProvider{
		SparkPost: &domain.SparkPostSettings{
			Endpoint: "https://api.sparkpost.test",
			APIKey:   "test-api-key",
		},
	}

	workspaceID := "workspace-123"
	integrationID := "integration-123"

	t.Run("Webhook found", func(t *testing.T) {
		ctx := context.Background()

		// Mock list webhooks response with a matching webhook
		webhookTarget := "https://api.notifuse.com/webhook?provider=sparkpost&workspace_id=workspace-123&integration_id=integration-123"
		listResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{
				{
					ID:     "webhook-123",
					Name:   "Notifuse Webhook",
					Target: webhookTarget,
					Events: []string{"delivery", "bounce"},
					Active: true,
				},
			},
		}
		listResponseJSON, _ := json.Marshal(listResponse)

		// Expect list webhooks request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				return mockHTTPResponse(http.StatusOK, string(listResponseJSON)), nil
			})

		// Call the service method
		result, err := sparkPostService.GetWebhookStatus(ctx, workspaceID, integrationID, providerConfig)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.EmailProviderKindSparkPost, result.EmailProviderKind)
		assert.True(t, result.IsRegistered)
		assert.NotEmpty(t, result.Endpoints)
	})

	t.Run("Webhook not found", func(t *testing.T) {
		ctx := context.Background()

		// Mock list webhooks response with no matching webhook
		listResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{
				{
					ID:     "webhook-456",
					Name:   "Other Webhook",
					Target: "https://other-service.com/webhook",
					Events: []string{"delivery"},
					Active: true,
				},
			},
		}
		listResponseJSON, _ := json.Marshal(listResponse)

		// Expect list webhooks request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				return mockHTTPResponse(http.StatusOK, string(listResponseJSON)), nil
			})

		// Call the service method
		result, err := sparkPostService.GetWebhookStatus(ctx, workspaceID, integrationID, providerConfig)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.EmailProviderKindSparkPost, result.EmailProviderKind)
		assert.False(t, result.IsRegistered)
		assert.Empty(t, result.Endpoints)
	})

	t.Run("Sandbox mode", func(t *testing.T) {
		ctx := context.Background()

		// Create a provider config with sandbox mode enabled
		sandboxConfig := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint:    "https://api.sparkpost.test",
				APIKey:      "test-api-key",
				SandboxMode: true,
			},
		}

		// Call the service method
		result, err := sparkPostService.GetWebhookStatus(ctx, workspaceID, integrationID, sandboxConfig)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.EmailProviderKindSparkPost, result.EmailProviderKind)
		assert.True(t, result.IsRegistered)
		assert.NotEmpty(t, result.Endpoints)
		assert.Equal(t, true, result.ProviderDetails["sandbox_mode"])
	})
}

func TestSparkPostService_RegisterWebhooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test configuration
	providerConfig := &domain.EmailProvider{
		SparkPost: &domain.SparkPostSettings{
			Endpoint: "https://api.sparkpost.test",
			APIKey:   "test-api-key",
		},
	}

	workspaceID := "workspace-123"
	integrationID := "integration-123"
	baseURL := "https://api.notifuse.com/webhook"
	eventTypes := []domain.EmailEventType{domain.EmailEventDelivered, domain.EmailEventBounce}

	t.Run("Success - Create new webhook", func(t *testing.T) {
		ctx := context.Background()

		// Mock list webhooks response (empty list)
		listResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{},
		}
		listResponseJSON, _ := json.Marshal(listResponse)

		// Mock create webhook response
		createResponse := domain.SparkPostWebhookResponse{
			Results: domain.SparkPostWebhook{
				ID:     "webhook-123",
				Name:   "Notifuse-integration-123",
				Target: domain.GenerateWebhookCallbackURL(baseURL, domain.EmailProviderKindSparkPost, workspaceID, integrationID),
				Events: []string{"delivery", "bounce"},
				Active: true,
			},
		}
		createResponseJSON, _ := json.Marshal(createResponse)

		// Expect list webhooks request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				return mockHTTPResponse(http.StatusOK, string(listResponseJSON)), nil
			})

		// Expect create webhook request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				return mockHTTPResponse(http.StatusOK, string(createResponseJSON)), nil
			})

		// Call the service method
		result, err := sparkPostService.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, providerConfig)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.EmailProviderKindSparkPost, result.EmailProviderKind)
		assert.True(t, result.IsRegistered)
		assert.Len(t, result.Endpoints, 2) // One for each event type
	})

	t.Run("Success - Sandbox mode", func(t *testing.T) {
		ctx := context.Background()

		// Create a provider config with sandbox mode enabled
		sandboxConfig := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint:    "https://api.sparkpost.test",
				APIKey:      "test-api-key",
				SandboxMode: true,
			},
		}

		// Call the service method
		result, err := sparkPostService.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, sandboxConfig)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.EmailProviderKindSparkPost, result.EmailProviderKind)
		assert.True(t, result.IsRegistered)
		assert.Len(t, result.Endpoints, 2) // One for each event type
		assert.Equal(t, true, result.ProviderDetails["sandbox_mode"])
	})

	t.Run("Invalid provider config", func(t *testing.T) {
		ctx := context.Background()

		// Call with nil provider config
		result, err := sparkPostService.RegisterWebhooks(ctx, workspaceID, integrationID, baseURL, eventTypes, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "configuration is missing or invalid")
	})
}

func TestSparkPostService_UnregisterWebhooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test configuration
	providerConfig := &domain.EmailProvider{
		SparkPost: &domain.SparkPostSettings{
			Endpoint: "https://api.sparkpost.test",
			APIKey:   "test-api-key",
		},
	}

	workspaceID := "workspace-123"
	integrationID := "integration-123"

	t.Run("Success - Delete existing webhook", func(t *testing.T) {
		ctx := context.Background()

		// Mock list webhooks response with a matching webhook
		webhookTarget := "https://api.notifuse.com/webhook?provider=sparkpost&workspace_id=workspace-123&integration_id=integration-123"
		listResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{
				{
					ID:     "webhook-123",
					Name:   "Notifuse Webhook",
					Target: webhookTarget,
					Events: []string{"delivery", "bounce"},
					Active: true,
				},
			},
		}
		listResponseJSON, _ := json.Marshal(listResponse)

		// Expect list webhooks request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks", req.URL.String())
				return mockHTTPResponse(http.StatusOK, string(listResponseJSON)), nil
			})

		// Expect delete webhook request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "DELETE", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/webhook-123", req.URL.String())
				return mockHTTPResponse(http.StatusOK, "{}"), nil
			})

		// Call the service method
		err := sparkPostService.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("Success - Sandbox mode", func(t *testing.T) {
		ctx := context.Background()

		// Create a provider config with sandbox mode enabled
		sandboxConfig := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint:    "https://api.sparkpost.test",
				APIKey:      "test-api-key",
				SandboxMode: true,
			},
		}

		// Call the service method - should succeed without making any HTTP calls
		err := sparkPostService.UnregisterWebhooks(ctx, workspaceID, integrationID, sandboxConfig)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("No matching webhooks", func(t *testing.T) {
		ctx := context.Background()

		// Mock list webhooks response with no matching webhooks
		listResponse := domain.SparkPostWebhookListResponse{
			Results: []domain.SparkPostWebhook{
				{
					ID:     "webhook-abc",
					Name:   "Other Webhook",
					Target: "https://other-service.com/webhook",
					Events: []string{"delivery"},
					Active: true,
				},
			},
		}
		listResponseJSON, _ := json.Marshal(listResponse)

		// Expect list webhooks request
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				return mockHTTPResponse(http.StatusOK, string(listResponseJSON)), nil
			})

		// Call the service method
		err := sparkPostService.UnregisterWebhooks(ctx, workspaceID, integrationID, providerConfig)

		// Verify results - should succeed as there's nothing to delete
		assert.NoError(t, err)
	})
}

func TestSparkPostService_GetWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}
	webhookID := "webhook-123"

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		webhookResponse := domain.SparkPostWebhookResponse{
			Results: domain.SparkPostWebhook{
				ID:     webhookID,
				Name:   "Test Webhook",
				Target: "https://example.com/webhook",
				Events: []string{"delivery", "bounce"},
				Active: true,
			},
		}
		responseJSON, _ := json.Marshal(webhookResponse)

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/webhook-123", req.URL.String())
				assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))

				return mockHTTPResponse(http.StatusOK, string(responseJSON)), nil
			})

		result, err := sparkPostService.GetWebhook(ctx, config, webhookID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, webhookID, result.Results.ID)
		assert.Equal(t, "Test Webhook", result.Results.Name)
	})

	t.Run("Error response", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusNotFound, `{"errors":[{"message":"Webhook not found"}]}`), nil)

		result, err := sparkPostService.GetWebhook(ctx, config, webhookID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "API returned non-OK status code 404")
	})
}

func TestSparkPostService_UpdateWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}
	webhookID := "webhook-123"
	webhook := domain.SparkPostWebhook{
		Name:     "Updated Webhook",
		Target:   "https://example.com/webhook",
		Events:   []string{"delivery", "bounce", "spam_complaint"},
		Active:   true,
		AuthType: "none",
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		webhookResponse := domain.SparkPostWebhookResponse{
			Results: domain.SparkPostWebhook{
				ID:     webhookID,
				Name:   webhook.Name,
				Target: webhook.Target,
				Events: webhook.Events,
				Active: webhook.Active,
			},
		}
		responseJSON, _ := json.Marshal(webhookResponse)

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "PUT", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/webhook-123", req.URL.String())

				var requestBody domain.SparkPostWebhook
				body, _ := io.ReadAll(req.Body)
				_ = json.Unmarshal(body, &requestBody)
				assert.Equal(t, webhook.Name, requestBody.Name)
				assert.Equal(t, webhook.Events, requestBody.Events)

				return mockHTTPResponse(http.StatusOK, string(responseJSON)), nil
			})

		result, err := sparkPostService.UpdateWebhook(ctx, config, webhookID, webhook)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, webhookID, result.Results.ID)
		assert.Equal(t, webhook.Name, result.Results.Name)
		assert.Equal(t, webhook.Events, result.Results.Events)
	})
}

func TestSparkPostService_DeleteWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}
	webhookID := "webhook-123"

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "DELETE", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/webhook-123", req.URL.String())

				return mockHTTPResponse(http.StatusOK, "{}"), nil
			})

		err := sparkPostService.DeleteWebhook(ctx, config, webhookID)

		assert.NoError(t, err)
	})

	t.Run("Error response", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusNotFound, `{"errors":[{"message":"Webhook not found"}]}`), nil)

		err := sparkPostService.DeleteWebhook(ctx, config, webhookID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code 404")
	})
}

func TestSparkPostService_TestWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}
	webhookID := "webhook-123"

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/webhook-123/validate", req.URL.String())

				return mockHTTPResponse(http.StatusOK, `{"results":{"message":"Test event sent successfully"}}`), nil
			})

		err := sparkPostService.TestWebhook(ctx, config, webhookID)

		assert.NoError(t, err)
	})
}

func TestSparkPostService_ValidateWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	config := domain.SparkPostSettings{
		Endpoint: "https://api.sparkpost.test",
		APIKey:   "test-api-key",
	}
	webhook := domain.SparkPostWebhook{
		Target: "https://example.com/webhook",
	}

	t.Run("Valid webhook", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/webhooks/validate", req.URL.String())

				// Verify request body
				var requestBody map[string]string
				body, _ := io.ReadAll(req.Body)
				_ = json.Unmarshal(body, &requestBody)
				assert.Equal(t, webhook.Target, requestBody["target"])

				return mockHTTPResponse(http.StatusOK, `{"results":{"valid":true}}`), nil
			})

		isValid, err := sparkPostService.ValidateWebhook(ctx, config, webhook)

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("Invalid webhook", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusOK, `{"results":{"valid":false}}`), nil)

		isValid, err := sparkPostService.ValidateWebhook(ctx, config, webhook)

		assert.NoError(t, err)
		assert.False(t, isValid)
	})

	t.Run("Error decoding response", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusOK, `invalid json`), nil)

		isValid, err := sparkPostService.ValidateWebhook(ctx, config, webhook)

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "failed to decode validation response")
	})
}

func TestSparkPostService_SendEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Allow any log calls
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Initialize service
	sparkPostService := service.NewSparkPostService(mockHTTPClient, mockAuthService, mockLogger)

	// Test data
	workspaceID := "workspace-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Email Content</p>"

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Create provider config
		provider := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint: "https://api.sparkpost.test",
				APIKey:   "test-api-key",
			},
		}

		// Expect HTTP request and return success response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "https://api.sparkpost.test/api/v1/transmissions", req.URL.String())
				assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				// Verify request body
				body, _ := io.ReadAll(req.Body)
				var emailReq map[string]interface{}
				err := json.Unmarshal(body, &emailReq)
				assert.NoError(t, err)

				// Check essential fields
				recipients, ok := emailReq["recipients"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, recipients, 1)
				assert.Equal(t, to, recipients[0].(map[string]interface{})["address"])

				// Check from field
				from, ok := emailReq["from"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, fromAddress, from["email"])
				assert.Equal(t, fromName, from["name"])

				// Check subject
				assert.Equal(t, subject, emailReq["subject"])

				// Check content
				contentMap, ok := emailReq["content"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, content, contentMap["html"])

				return mockHTTPResponse(http.StatusOK, `{"results":{"id":"test-transmission-id"}}`), nil
			})

		// Call the service method
		err := sparkPostService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, "", nil, nil)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("Missing SparkPost configuration", func(t *testing.T) {
		ctx := context.Background()

		// Create provider without SparkPost config
		provider := &domain.EmailProvider{}

		// Call the service method
		err := sparkPostService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, "", nil, nil)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SparkPost provider is not configured")
	})

	t.Run("HTTP request error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Create provider config
		provider := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint: "https://api.sparkpost.test",
				APIKey:   "test-api-key",
			},
		}

		// Mock HTTP client to return error
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(nil, expectedErr)

		// Call the service method
		err := sparkPostService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, "", nil, nil)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
	})

	t.Run("API error response", func(t *testing.T) {
		ctx := context.Background()

		// Create provider config
		provider := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint: "https://api.sparkpost.test",
				APIKey:   "test-api-key",
			},
		}

		// Mock error response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusBadRequest, `{"errors":[{"message":"Invalid recipient address"}]}`), nil)

		// Call the service method
		err := sparkPostService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, "", nil, nil)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned non-OK status code 400")
	})

	t.Run("Sandbox mode", func(t *testing.T) {
		ctx := context.Background()

		// Create provider with sandbox mode enabled
		provider := &domain.EmailProvider{
			SparkPost: &domain.SparkPostSettings{
				Endpoint:    "https://api.sparkpost.test",
				APIKey:      "test-api-key",
				SandboxMode: true,
			},
		}

		// Expect HTTP request and return success response
		mockHTTPClient.EXPECT().
			Do(gomock.Any()).
			Return(mockHTTPResponse(http.StatusOK, `{"results":{"id":"test-transmission-id"}}`), nil)

		// Call the service method
		err := sparkPostService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, "", nil, nil)

		// Verify results - should succeed in sandbox mode
		assert.NoError(t, err)
	})
}
