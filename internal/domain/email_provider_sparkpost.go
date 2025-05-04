package domain

import (
	"context"
)

//go:generate mockgen -destination mocks/mock_sparkpost_service.go -package mocks github.com/Notifuse/notifuse/internal/domain SparkPostServiceInterface

// SparkPostWebhookPayload represents the webhook payload from SparkPost
type SparkPostWebhookPayload struct {
	MSys SparkPostMSys `json:"msys"`
}

// SparkPostMSys contains the event data from SparkPost
type SparkPostMSys struct {
	// Different event types that can be present in the payload
	DeliveryEvent  *SparkPostDeliveryEvent  `json:"delivery_event,omitempty"`
	BounceEvent    *SparkPostBounceEvent    `json:"bounce_event,omitempty"`
	DelayEvent     *SparkPostDelayEvent     `json:"delay_event,omitempty"`
	InjectionEvent *SparkPostInjectionEvent `json:"injection_event,omitempty"`
	SpamComplaint  *SparkPostSpamComplaint  `json:"spam_complaint,omitempty"`
}

// SparkPostDeliveryEvent represents a delivery event from SparkPost
type SparkPostDeliveryEvent struct {
	Type          string                 `json:"type"`
	CampaignID    string                 `json:"campaign_id"`
	MessageID     string                 `json:"message_id"`
	Timestamp     string                 `json:"timestamp"`
	RecipientTo   string                 `json:"rcpt_to"`
	RecipientMeta map[string]interface{} `json:"rcpt_meta,omitempty"`
	RawReason     string                 `json:"raw_reason,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Transmission  string                 `json:"transmission_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	GeoIP         *SparkPostGeoIP        `json:"geo_ip,omitempty"`
	MessageFrom   string                 `json:"msg_from,omitempty"`
	QueueTime     string                 `json:"queue_time,omitempty"`
}

// SparkPostBounceEvent represents a bounce event from SparkPost
type SparkPostBounceEvent struct {
	Type          string                 `json:"type"`
	CampaignID    string                 `json:"campaign_id"`
	MessageID     string                 `json:"message_id"`
	Timestamp     string                 `json:"timestamp"`
	RecipientTo   string                 `json:"rcpt_to"`
	RecipientMeta map[string]interface{} `json:"rcpt_meta,omitempty"`
	RawReason     string                 `json:"raw_reason,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Transmission  string                 `json:"transmission_id,omitempty"`
	BounceClass   string                 `json:"bounce_class,omitempty"`
	Error         string                 `json:"error_code,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
}

// SparkPostDelayEvent represents a delay event from SparkPost
type SparkPostDelayEvent struct {
	Type          string                 `json:"type"`
	CampaignID    string                 `json:"campaign_id"`
	MessageID     string                 `json:"message_id"`
	Timestamp     string                 `json:"timestamp"`
	RecipientTo   string                 `json:"rcpt_to"`
	RecipientMeta map[string]interface{} `json:"rcpt_meta,omitempty"`
	RawReason     string                 `json:"raw_reason,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Transmission  string                 `json:"transmission_id,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
}

// SparkPostInjectionEvent represents an injection event from SparkPost
type SparkPostInjectionEvent struct {
	Type          string                 `json:"type"`
	CampaignID    string                 `json:"campaign_id"`
	MessageID     string                 `json:"message_id"`
	Timestamp     string                 `json:"timestamp"`
	RecipientTo   string                 `json:"rcpt_to"`
	RecipientMeta map[string]interface{} `json:"rcpt_meta,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Transmission  string                 `json:"transmission_id,omitempty"`
	MessageSize   string                 `json:"msg_size,omitempty"`
}

// SparkPostSpamComplaint represents a spam complaint from SparkPost
type SparkPostSpamComplaint struct {
	Type          string                 `json:"type"`
	CampaignID    string                 `json:"campaign_id"`
	MessageID     string                 `json:"message_id"`
	Timestamp     string                 `json:"timestamp"`
	RecipientTo   string                 `json:"rcpt_to"`
	RecipientMeta map[string]interface{} `json:"rcpt_meta,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Transmission  string                 `json:"transmission_id,omitempty"`
	FeedbackType  string                 `json:"fbtype,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
}

// SparkPostGeoIP represents geographic IP information
type SparkPostGeoIP struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// SparkPostWebhook represents a webhook configuration in SparkPost
type SparkPostWebhook struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name"`
	Target        string                 `json:"target"`
	Events        []string               `json:"events"`
	Active        bool                   `json:"active"`
	AuthType      string                 `json:"auth_type,omitempty"` // "none", "basic", "oauth2"
	AuthToken     string                 `json:"auth_token,omitempty"`
	AuthRequest   map[string]interface{} `json:"auth_request,omitempty"`
	CustomHeaders map[string]string      `json:"custom_headers,omitempty"`
}

// SparkPostWebhookListResponse represents the response for listing webhooks
type SparkPostWebhookListResponse struct {
	Results []SparkPostWebhook `json:"results"`
}

// SparkPostWebhookResponse represents a response for webhook operations
type SparkPostWebhookResponse struct {
	Results SparkPostWebhook `json:"results"`
}

// SparkPostConfig represents configuration for SparkPost API
type SparkPostConfig struct {
	APIKey      string `json:"api_key"`
	APIEndpoint string `json:"api_endpoint"`
}

//go:generate mockgen -destination mocks/mock_sparkpost_service.go -package mocks github.com/Notifuse/notifuse/internal/domain SparkPostServiceInterface

// SparkPostServiceInterface defines operations for managing SparkPost webhooks
type SparkPostServiceInterface interface {
	// ListWebhooks retrieves all registered webhooks
	ListWebhooks(ctx context.Context, config SparkPostConfig) (*SparkPostWebhookListResponse, error)

	// CreateWebhook creates a new webhook
	CreateWebhook(ctx context.Context, config SparkPostConfig, webhook SparkPostWebhook) (*SparkPostWebhookResponse, error)

	// GetWebhook retrieves a webhook by ID
	GetWebhook(ctx context.Context, config SparkPostConfig, webhookID string) (*SparkPostWebhookResponse, error)

	// UpdateWebhook updates an existing webhook
	UpdateWebhook(ctx context.Context, config SparkPostConfig, webhookID string, webhook SparkPostWebhook) (*SparkPostWebhookResponse, error)

	// DeleteWebhook deletes a webhook by ID
	DeleteWebhook(ctx context.Context, config SparkPostConfig, webhookID string) error

	// TestWebhook sends a test event to a webhook
	TestWebhook(ctx context.Context, config SparkPostConfig, webhookID string) error

	// ValidateWebhook validates a webhook's configuration
	ValidateWebhook(ctx context.Context, config SparkPostConfig, webhook SparkPostWebhook) (bool, error)
}
