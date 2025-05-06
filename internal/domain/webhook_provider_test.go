package domain_test

import (
	"context"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookProviderInterface(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	workspaceID := "workspace-123"
	integrationID := "integration-456"
	baseURL := "https://api.example.com/webhooks"
	eventTypes := []domain.EmailEventType{domain.EmailEventDelivered, domain.EmailEventBounce}

	// Create provider config for testing
	providerConfig := &domain.EmailProvider{
		Kind: domain.EmailProviderKindSparkPost,
		SparkPost: &domain.SparkPostSettings{
			APIKey:   "test-api-key",
			Endpoint: "https://api.sparkpost.com",
		},
	}

	t.Run("RegisterWebhooks", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Define the expected webhook status to be returned
		expectedStatus := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSparkPost,
			IsRegistered:      true,
			Endpoints: []domain.WebhookEndpointStatus{
				{
					WebhookID: "webhook-123",
					URL:       baseURL,
					EventType: domain.EmailEventDelivered,
					Active:    true,
				},
			},
		}

		// Setup expectations
		mockProvider.EXPECT().RegisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			baseURL,
			eventTypes,
			providerConfig,
		).Return(expectedStatus, nil)

		// Call method on the mock
		status, err := mockProvider.RegisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			baseURL,
			eventTypes,
			providerConfig,
		)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("GetWebhookStatus", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Define the expected webhook status to be returned
		expectedStatus := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSparkPost,
			IsRegistered:      true,
			Endpoints: []domain.WebhookEndpointStatus{
				{
					WebhookID: "webhook-123",
					URL:       baseURL,
					EventType: domain.EmailEventDelivered,
					Active:    true,
				},
			},
		}

		// Setup expectations
		mockProvider.EXPECT().GetWebhookStatus(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		).Return(expectedStatus, nil)

		// Call method on the mock
		status, err := mockProvider.GetWebhookStatus(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("UnregisterWebhooks", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Setup expectations
		mockProvider.EXPECT().UnregisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		).Return(nil)

		// Call method on the mock
		err := mockProvider.UnregisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("RegisterWebhooks_Error", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Setup error expectations
		mockProvider.EXPECT().RegisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			baseURL,
			eventTypes,
			providerConfig,
		).Return(nil, assert.AnError)

		// Call method on the mock
		status, err := mockProvider.RegisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			baseURL,
			eventTypes,
			providerConfig,
		)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("GetWebhookStatus_Error", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Setup error expectations
		mockProvider.EXPECT().GetWebhookStatus(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		).Return(nil, assert.AnError)

		// Call method on the mock
		status, err := mockProvider.GetWebhookStatus(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("UnregisterWebhooks_Error", func(t *testing.T) {
		// Create mock implementation
		mockProvider := mocks.NewMockWebhookProvider(ctrl)

		// Setup error expectations
		mockProvider.EXPECT().UnregisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		).Return(assert.AnError)

		// Call method on the mock
		err := mockProvider.UnregisterWebhooks(
			ctx,
			workspaceID,
			integrationID,
			providerConfig,
		)

		// Assertions
		require.Error(t, err)
	})
}
