package domain

import (
	"context"
)

//go:generate mockgen -destination mocks/mock_webhook_registration_service.go -package mocks github.com/Notifuse/notifuse/internal/domain WebhookRegistrationService

// WebhookRegistrationService defines the interface for registering webhooks with email providers
type WebhookRegistrationService interface {
	// RegisterWebhooks registers the webhook URLs with the email provider
	RegisterWebhooks(ctx context.Context, workspaceID string, config *WebhookRegistrationConfig) (*WebhookRegistrationStatus, error)

	// GetWebhookStatus gets the current status of webhooks for the email provider
	GetWebhookStatus(ctx context.Context, workspaceID string, integrationID string) (*WebhookRegistrationStatus, error)
}

// WebhookRegistrationConfig defines the configuration for registering webhooks
type WebhookRegistrationConfig struct {
	BaseURL       string           `json:"base_url"`
	IntegrationID string           `json:"integration_id"`
	EventTypes    []EmailEventType `json:"event_types"`
}

// WebhookRegistrationStatus represents the current status of webhooks for a provider
type WebhookRegistrationStatus struct {
	EmailProviderKind EmailProviderKind       `json:"email_provider_kind"`
	IsRegistered      bool                    `json:"is_registered"`
	RegisteredEvents  []EmailEventType        `json:"registered_events,omitempty"`
	Endpoints         []WebhookEndpointStatus `json:"endpoints,omitempty"`
	Error             string                  `json:"error,omitempty"`
	ProviderDetails   map[string]interface{}  `json:"provider_details,omitempty"`
}

// WebhookEndpointStatus represents the status of a single webhook endpoint
type WebhookEndpointStatus struct {
	URL       string         `json:"url"`
	EventType EmailEventType `json:"event_type"`
	Active    bool           `json:"active"`
}

// RegisterWebhookRequest defines the request to register webhooks
type RegisterWebhookRequest struct {
	WorkspaceID   string           `json:"workspace_id"`
	IntegrationID string           `json:"integration_id"`
	BaseURL       string           `json:"base_url"`
	EventTypes    []EmailEventType `json:"event_types"`
}

// Validate validates the RegisterWebhookRequest
func (r *RegisterWebhookRequest) Validate() error {
	if r.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}
	if r.IntegrationID == "" {
		return NewValidationError("integration_id is required")
	}
	if r.BaseURL == "" {
		return NewValidationError("base_url is required")
	}
	if len(r.EventTypes) == 0 {
		// Default to all event types if not specified
		r.EventTypes = []EmailEventType{
			EmailEventDelivered,
			EmailEventBounce,
			EmailEventComplaint,
		}
	}

	return nil
}

// GetWebhookStatusRequest defines the request to get webhook status
type GetWebhookStatusRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	IntegrationID string `json:"integration_id"`
}

// Validate validates the GetWebhookStatusRequest
func (r *GetWebhookStatusRequest) Validate() error {
	if r.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}
	if r.IntegrationID == "" {
		return NewValidationError("integration_id is required")
	}
	return nil
}
