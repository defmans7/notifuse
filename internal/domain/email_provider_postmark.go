package domain

import (
	"context"
)

// PostmarkWebhookPayload represents the base webhook payload from Postmark
type PostmarkWebhookPayload struct {
	RecordType    string            `json:"RecordType"`
	MessageStream string            `json:"MessageStream"`
	ID            string            `json:"ID"`
	MessageID     string            `json:"MessageID"`
	ServerID      int               `json:"ServerID"`
	Metadata      map[string]string `json:"Metadata,omitempty"`
	Tag           string            `json:"Tag,omitempty"`

	// Delivered event specific fields
	DeliveredFields *PostmarkDeliveredFields `json:"-"`

	// Bounce event specific fields
	BounceFields *PostmarkBounceFields `json:"-"`

	// Complaint event specific fields
	ComplaintFields *PostmarkComplaintFields `json:"-"`
}

// PostmarkDeliveredFields contains fields specific to delivery events
type PostmarkDeliveredFields struct {
	RecipientEmail string `json:"Recipient"`
	DeliveredAt    string `json:"DeliveredAt"`
	Details        string `json:"Details"`
}

// PostmarkBounceFields contains fields specific to bounce events
type PostmarkBounceFields struct {
	RecipientEmail string `json:"Email"`
	BouncedAt      string `json:"BouncedAt"`
	Type           string `json:"Type"`
	TypeCode       int    `json:"TypeCode"`
	Name           string `json:"Name"`
	Description    string `json:"Description,omitempty"`
	Details        string `json:"Details,omitempty"`
	DumpAvailable  bool   `json:"DumpAvailable"`
	CanActivate    bool   `json:"CanActivate"`
	Subject        string `json:"Subject"`
	Content        string `json:"Content,omitempty"`
}

// PostmarkComplaintFields contains fields specific to complaint events
type PostmarkComplaintFields struct {
	RecipientEmail string `json:"Email"`
	ComplainedAt   string `json:"ComplainedAt"`
	Type           string `json:"Type"`
	UserAgent      string `json:"UserAgent,omitempty"`
	Subject        string `json:"Subject"`
}

// PostmarkWebhookConfig represents a webhook configuration in Postmark
type PostmarkWebhookConfig struct {
	ID            int                   `json:"ID,omitempty"`
	URL           string                `json:"Url"`
	MessageStream string                `json:"MessageStream"`
	HttpAuth      *HttpAuth             `json:"HttpAuth,omitempty"`
	HttpHeaders   map[string]string     `json:"HttpHeaders,omitempty"`
	TriggerRules  []PostmarkTriggerRule `json:"Triggers"`
}

// HttpAuth represents HTTP authentication for webhooks
type HttpAuth struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

// PostmarkTriggerRule represents a trigger for webhooks
type PostmarkTriggerRule struct {
	Key   string `json:"Key"`
	Match string `json:"Match"`
	Value string `json:"Value"`
}

// PostmarkWebhookResponse represents the response from Postmark API for webhook operations
type PostmarkWebhookResponse struct {
	ID            int               `json:"ID"`
	URL           string            `json:"Url"`
	MessageStream string            `json:"MessageStream"`
	Triggers      []PostmarkTrigger `json:"Triggers"`
}

// PostmarkTrigger represents a webhook trigger in the response
type PostmarkTrigger struct {
	Key   string `json:"Key"`
	Match string `json:"Match"`
	Value string `json:"Value"`
}

// PostmarkListWebhooksResponse represents the response for listing webhooks
type PostmarkListWebhooksResponse struct {
	TotalCount int                       `json:"TotalCount"`
	Webhooks   []PostmarkWebhookResponse `json:"Webhooks"`
}

// PostmarkConfig contains configuration for Postmark API
type PostmarkConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	ServerToken string `json:"server_token"`
}

//go:generate mockgen -destination mocks/mock_postmark_service.go -package mocks github.com/Notifuse/notifuse/internal/domain PostmarkServiceInterface

// PostmarkServiceInterface defines operations for managing Postmark webhooks
type PostmarkServiceInterface interface {
	// ListWebhooks retrieves all registered webhooks
	ListWebhooks(ctx context.Context, config PostmarkConfig) (*PostmarkListWebhooksResponse, error)

	// RegisterWebhook registers a new webhook
	RegisterWebhook(ctx context.Context, config PostmarkConfig, webhook PostmarkWebhookConfig) (*PostmarkWebhookResponse, error)

	// UnregisterWebhook removes a webhook by ID
	UnregisterWebhook(ctx context.Context, config PostmarkConfig, webhookID int) error

	// GetWebhook retrieves a specific webhook by ID
	GetWebhook(ctx context.Context, config PostmarkConfig, webhookID int) (*PostmarkWebhookResponse, error)

	// UpdateWebhook updates an existing webhook
	UpdateWebhook(ctx context.Context, config PostmarkConfig, webhookID int, webhook PostmarkWebhookConfig) (*PostmarkWebhookResponse, error)

	// TestWebhook sends a test event to the webhook
	TestWebhook(ctx context.Context, config PostmarkConfig, webhookID int, eventType EmailEventType) error
}
