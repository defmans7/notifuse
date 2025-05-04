package domain

import (
	"context"
)

//go:generate mockgen -destination mocks/mock_mailgun_service.go -package mocks github.com/Notifuse/notifuse/internal/domain MailgunServiceInterface

// MailgunWebhookPayload represents a Mailgun webhook payload
type MailgunWebhookPayload struct {
	Signature MailgunSignature `json:"signature"`
	EventData MailgunEventData `json:"event-data"`
}

// MailgunSignature contains signature information for webhook authentication
type MailgunSignature struct {
	Timestamp string `json:"timestamp"`
	Token     string `json:"token"`
	Signature string `json:"signature"`
}

// MailgunEventData contains the main event data from Mailgun
type MailgunEventData struct {
	Event         string                 `json:"event"`
	Timestamp     float64                `json:"timestamp"`
	ID            string                 `json:"id"`
	Recipient     string                 `json:"recipient"`
	Tags          []string               `json:"tags"`
	Message       MailgunMessage         `json:"message"`
	Delivery      MailgunDelivery        `json:"delivery,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
	Severity      string                 `json:"severity,omitempty"`
	Storage       map[string]interface{} `json:"storage,omitempty"`
	UserVariables map[string]interface{} `json:"user-variables,omitempty"`
	Flags         map[string]interface{} `json:"flags,omitempty"`
}

// MailgunMessage contains information about the email message
type MailgunMessage struct {
	Headers     MailgunHeaders `json:"headers"`
	Attachments []interface{}  `json:"attachments"`
	Size        int            `json:"size"`
}

// MailgunHeaders contains email headers
type MailgunHeaders struct {
	To        string `json:"to"`
	MessageID string `json:"message-id"`
	From      string `json:"from"`
	Subject   string `json:"subject"`
}

// MailgunDelivery contains delivery information
type MailgunDelivery struct {
	Status           string                 `json:"status,omitempty"`
	Code             int                    `json:"code,omitempty"`
	Message          string                 `json:"message,omitempty"`
	AttemptNo        int                    `json:"attempt-no,omitempty"`
	Description      string                 `json:"description,omitempty"`
	SessionSeconds   float64                `json:"session-seconds,omitempty"`
	Certificate      bool                   `json:"certificate,omitempty"`
	TLS              bool                   `json:"tls,omitempty"`
	MXHost           string                 `json:"mx-host,omitempty"`
	DelvDataFeedback []interface{}          `json:"delivery-status,omitempty"`
	SMTP             map[string]interface{} `json:"smtp,omitempty"`
}

// MailgunWebhook represents a webhook configuration in Mailgun
type MailgunWebhook struct {
	ID     string   `json:"id,omitempty"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

// MailgunWebhookListResponse represents the response from listing webhooks
type MailgunWebhookListResponse struct {
	Items []MailgunWebhook `json:"items"`
	Total int              `json:"total"`
}

// MailgunConfig represents configuration for Mailgun API
type MailgunConfig struct {
	APIKey  string `json:"api_key"`
	Domain  string `json:"domain"`
	BaseURL string `json:"base_url"`
	Region  string `json:"region,omitempty"` // "US" or "EU"
}

//go:generate mockgen -destination mocks/mock_mailgun_service.go -package mocks github.com/Notifuse/notifuse/internal/domain MailgunServiceInterface

// MailgunServiceInterface defines operations for managing Mailgun webhooks
type MailgunServiceInterface interface {
	// ListWebhooks retrieves all registered webhooks for a domain
	ListWebhooks(ctx context.Context, config MailgunConfig) (*MailgunWebhookListResponse, error)

	// CreateWebhook creates a new webhook
	CreateWebhook(ctx context.Context, config MailgunConfig, webhook MailgunWebhook) (*MailgunWebhook, error)

	// GetWebhook retrieves a webhook by ID
	GetWebhook(ctx context.Context, config MailgunConfig, webhookID string) (*MailgunWebhook, error)

	// UpdateWebhook updates an existing webhook
	UpdateWebhook(ctx context.Context, config MailgunConfig, webhookID string, webhook MailgunWebhook) (*MailgunWebhook, error)

	// DeleteWebhook deletes a webhook by ID
	DeleteWebhook(ctx context.Context, config MailgunConfig, webhookID string) error

	// TestWebhook sends a test event to a webhook
	TestWebhook(ctx context.Context, config MailgunConfig, webhookID string, eventType string) error
}
