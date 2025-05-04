package domain

import (
	"context"
	"fmt"
	"time"
)

//go:generate mockgen -destination mocks/mock_webhook_event_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain WebhookEventRepository
//go:generate mockgen -destination mocks/mock_webhook_event_service.go -package mocks github.com/Notifuse/notifuse/internal/domain WebhookEventServiceInterface

// EmailEventType defines the type of email webhook event
type EmailEventType string

const (
	// EmailEventDelivered indicates a successful email delivery
	EmailEventDelivered EmailEventType = "delivered"

	// EmailEventBounce indicates a bounced email
	EmailEventBounce EmailEventType = "bounce"

	// EmailEventComplaint indicates a complaint was filed for the email
	EmailEventComplaint EmailEventType = "complaint"
)

// WebhookEvent represents an event received from an email provider webhook
type WebhookEvent struct {
	ID                string            `json:"id"`
	Type              EmailEventType    `json:"type"`
	EmailProviderKind EmailProviderKind `json:"email_provider_kind"`
	IntegrationID     string            `json:"integration_id"`
	RecipientEmail    string            `json:"recipient_email"`
	MessageID         string            `json:"message_id"`
	TransactionalID   string            `json:"transactional_id,omitempty"`
	BroadcastID       string            `json:"broadcast_id,omitempty"`
	Timestamp         time.Time         `json:"timestamp"`
	RawPayload        string            `json:"raw_payload"`

	// Bounce specific fields
	BounceType       string `json:"bounce_type,omitempty"`
	BounceCategory   string `json:"bounce_category,omitempty"`
	BounceDiagnostic string `json:"bounce_diagnostic,omitempty"`

	// Complaint specific fields
	ComplaintFeedbackType string `json:"complaint_feedback_type,omitempty"`
}

// NewWebhookEvent creates a new webhook event
func NewWebhookEvent(
	id string,
	eventType EmailEventType,
	providerKind EmailProviderKind,
	integrationID string,
	recipientEmail string,
	messageID string,
	timestamp time.Time,
	rawPayload string,
) *WebhookEvent {
	return &WebhookEvent{
		ID:                id,
		Type:              eventType,
		EmailProviderKind: providerKind,
		IntegrationID:     integrationID,
		RecipientEmail:    recipientEmail,
		MessageID:         messageID,
		Timestamp:         timestamp,
		RawPayload:        rawPayload,
	}
}

// SetBounceInfo sets bounce-specific information
func (w *WebhookEvent) SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic string) {
	w.BounceType = bounceType
	w.BounceCategory = bounceCategory
	w.BounceDiagnostic = bounceDiagnostic
}

// SetComplaintInfo sets complaint-specific information
func (w *WebhookEvent) SetComplaintInfo(feedbackType string) {
	w.ComplaintFeedbackType = feedbackType
}

// SetTransactionalID sets the transactional ID for this event
func (w *WebhookEvent) SetTransactionalID(transactionalID string) {
	w.TransactionalID = transactionalID
}

// SetBroadcastID sets the broadcast ID for this event
func (w *WebhookEvent) SetBroadcastID(broadcastID string) {
	w.BroadcastID = broadcastID
}

// ErrWebhookEventNotFound is returned when a webhook event is not found
type ErrWebhookEventNotFound struct {
	ID string
}

// Error returns the error message
func (e *ErrWebhookEventNotFound) Error() string {
	return fmt.Sprintf("webhook event with ID %s not found", e.ID)
}

// GetEventsRequest defines the parameters for retrieving webhook events
type GetEventsRequest struct {
	WorkspaceID string         `json:"workspace_id"`
	Type        EmailEventType `json:"type,omitempty"`
	Limit       int            `json:"limit,omitempty"`
	Offset      int            `json:"offset,omitempty"`
}

// GetEventByIDRequest defines the parameters for retrieving a webhook event by ID
type GetEventByIDRequest struct {
	ID string `json:"id"`
}

// GetEventsByMessageIDRequest defines the parameters for retrieving webhook events by message ID
type GetEventsByMessageIDRequest struct {
	MessageID string `json:"message_id"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// GetEventsByTransactionalIDRequest defines the parameters for retrieving webhook events by transactional ID
type GetEventsByTransactionalIDRequest struct {
	WorkspaceID     string `json:"workspace_id"`
	TransactionalID string `json:"transactional_id"`
	Limit           int    `json:"limit,omitempty"`
	Offset          int    `json:"offset,omitempty"`
}

// GetEventsByBroadcastIDRequest defines the parameters for retrieving webhook events by broadcast ID
type GetEventsByBroadcastIDRequest struct {
	WorkspaceID string `json:"workspace_id"`
	BroadcastID string `json:"broadcast_id"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// Validate validates the GetEventsRequest
func (r *GetEventsRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Max limit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

// Validate validates the GetEventByIDRequest
func (r *GetEventByIDRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}

// Validate validates the GetEventsByMessageIDRequest
func (r *GetEventsByMessageIDRequest) Validate() error {
	if r.MessageID == "" {
		return fmt.Errorf("message_id is required")
	}
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Max limit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

// Validate validates the GetEventsByTransactionalIDRequest
func (r *GetEventsByTransactionalIDRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}
	if r.TransactionalID == "" {
		return fmt.Errorf("transactional_id is required")
	}
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Max limit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

// Validate validates the GetEventsByBroadcastIDRequest
func (r *GetEventsByBroadcastIDRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}
	if r.BroadcastID == "" {
		return fmt.Errorf("broadcast_id is required")
	}
	if r.Limit <= 0 {
		r.Limit = 20 // Default limit
	}
	if r.Limit > 100 {
		r.Limit = 100 // Max limit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

//go:generate mockgen -destination mocks/mock_webhook_event_service.go -package mocks github.com/Notifuse/notifuse/internal/domain WebhookEventServiceInterface

// WebhookEventServiceInterface defines the interface for webhook event service
type WebhookEventServiceInterface interface {
	// ProcessWebhook processes a webhook event from an email provider
	ProcessWebhook(ctx context.Context, workspaceID, integrationID string, rawPayload []byte) error

	// GetEventByID retrieves a webhook event by its ID
	GetEventByID(ctx context.Context, id string) (*WebhookEvent, error)

	// GetEventsByType retrieves webhook events by type for a workspace
	GetEventsByType(ctx context.Context, workspaceID string, eventType EmailEventType, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByMessageID retrieves all webhook events associated with a message ID
	GetEventsByMessageID(ctx context.Context, messageID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByTransactionalID retrieves all webhook events associated with a transactional ID
	GetEventsByTransactionalID(ctx context.Context, transactionalID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByBroadcastID retrieves all webhook events associated with a broadcast ID
	GetEventsByBroadcastID(ctx context.Context, broadcastID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventCount retrieves the count of events by type for a workspace
	GetEventCount(ctx context.Context, workspaceID string, eventType EmailEventType) (int, error)
}

// WebhookEventRepository is the interface for webhook event operations
type WebhookEventRepository interface {
	// StoreEvent stores a webhook event in the database
	StoreEvent(ctx context.Context, event *WebhookEvent) error

	// GetEventByID retrieves a webhook event by its ID
	GetEventByID(ctx context.Context, id string) (*WebhookEvent, error)

	// GetEventsByMessageID retrieves all webhook events associated with a message ID
	GetEventsByMessageID(ctx context.Context, messageID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByTransactionalID retrieves all webhook events associated with a transactional ID
	GetEventsByTransactionalID(ctx context.Context, transactionalID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByBroadcastID retrieves all webhook events associated with a broadcast ID
	GetEventsByBroadcastID(ctx context.Context, broadcastID string, limit, offset int) ([]*WebhookEvent, error)

	// GetEventsByType retrieves webhook events by type (delivered, bounce, complaint)
	GetEventsByType(ctx context.Context, workspaceID string, eventType EmailEventType, limit, offset int) ([]*WebhookEvent, error)

	// GetEventCount retrieves the count of events by type for a workspace
	GetEventCount(ctx context.Context, workspaceID string, eventType EmailEventType) (int, error)
}
