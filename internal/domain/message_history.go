package domain

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_message_history_service.go -package mocks github.com/Notifuse/notifuse/internal/domain MessageHistoryService
//go:generate mockgen -destination mocks/mock_message_history_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain MessageHistoryRepository

// MessageStatus represents the current status of a message
type MessageStatus string

const (
	// Message status constants
	MessageStatusSent         MessageStatus = "sent"
	MessageStatusDelivered    MessageStatus = "delivered"
	MessageStatusFailed       MessageStatus = "failed"
	MessageStatusOpened       MessageStatus = "opened"
	MessageStatusClicked      MessageStatus = "clicked"
	MessageStatusBounced      MessageStatus = "bounced"
	MessageStatusComplained   MessageStatus = "complained"
	MessageStatusUnsubscribed MessageStatus = "unsubscribed"
)

// MessageData represents the JSON data used to compile a template
type MessageData struct {
	// Custom fields used in template compilation
	Data map[string]interface{} `json:"data"`
	// Optional metadata for tracking
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (d MessageData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface for database retrieval
func (d *MessageData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return sql.ErrNoRows
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, &d)
}

// MessageHistory represents a record of a message sent to a contact
type MessageHistory struct {
	ID              string        `json:"id"`
	ContactEmail    string        `json:"contact_email"`
	BroadcastID     *string       `json:"broadcast_id,omitempty"`
	TemplateID      string        `json:"template_id"`
	TemplateVersion int           `json:"template_version"`
	Channel         string        `json:"channel"` // email, sms, push, etc.
	Status          MessageStatus `json:"status"`
	Error           *string       `json:"error,omitempty"`
	MessageData     MessageData   `json:"message_data"`

	// Event timestamps
	SentAt         time.Time  `json:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClickedAt      *time.Time `json:"clicked_at,omitempty"`
	BouncedAt      *time.Time `json:"bounced_at,omitempty"`
	ComplainedAt   *time.Time `json:"complained_at,omitempty"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`

	// System timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageHistoryStatusSum struct {
	TotalSent         int `json:"total_sent"`
	TotalDelivered    int `json:"total_delivered"`
	TotalBounced      int `json:"total_bounced"`
	TotalComplained   int `json:"total_complained"`
	TotalFailed       int `json:"total_failed"`
	TotalOpened       int `json:"total_opened"`
	TotalClicked      int `json:"total_clicked"`
	TotalUnsubscribed int `json:"total_unsubscribed"`
}

// MessageHistoryRepository defines methods for message history persistence
type MessageHistoryRepository interface {
	// Create adds a new message history record
	Create(ctx context.Context, workspaceID string, message *MessageHistory) error

	// Update updates an existing message history record
	Update(ctx context.Context, workspaceID string, message *MessageHistory) error

	// Get retrieves a message history by ID
	Get(ctx context.Context, workspaceID, id string) (*MessageHistory, error)

	// GetByContact retrieves message history for a specific contact
	GetByContact(ctx context.Context, workspaceID, contactEmail string, limit, offset int) ([]*MessageHistory, int, error)

	// GetByBroadcast retrieves message history for a specific broadcast
	GetByBroadcast(ctx context.Context, workspaceID, broadcastID string, limit, offset int) ([]*MessageHistory, int, error)

	// ListMessages retrieves message history with cursor-based pagination and filtering
	ListMessages(ctx context.Context, workspaceID string, params MessageListParams) ([]*MessageHistory, string, error)

	// UpdateStatus updates the status of a message and sets the corresponding timestamp
	UpdateStatus(ctx context.Context, workspaceID, id string, status MessageStatus, timestamp time.Time) error

	// SetStatusIfNotSet sets a status only if it hasn't been set before (the field is NULL)
	SetStatusIfNotSet(ctx context.Context, workspaceID, id string, status MessageStatus, timestamp time.Time) error

	// SetClicked sets the clicked_at timestamp and ensures opened_at is also set
	SetClicked(ctx context.Context, workspaceID, id string, timestamp time.Time) error

	// SetOpened sets the opened_at timestamp if not already set
	SetOpened(ctx context.Context, workspaceID, id string, timestamp time.Time) error

	// GetBroadcastStats retrieves statistics for a broadcast
	GetBroadcastStats(ctx context.Context, workspaceID, broadcastID string) (*MessageHistoryStatusSum, error)

	// GetBroadcastVariationStats retrieves statistics for a specific variation of a broadcast
	GetBroadcastVariationStats(ctx context.Context, workspaceID, broadcastID, variationID string) (*MessageHistoryStatusSum, error)
}

// MessageHistoryService defines methods for interacting with message history
type MessageHistoryService interface {
	// ListMessages retrieves messages for a workspace with cursor-based pagination and filters
	ListMessages(ctx context.Context, workspaceID string, params MessageListParams) (*MessageListResult, error)

	// GetBroadcastStats retrieves statistics for a broadcast
	GetBroadcastStats(ctx context.Context, workspaceID, broadcastID string) (*MessageHistoryStatusSum, error)

	// GetBroadcastVariationStats retrieves statistics for a specific variation of a broadcast
	GetBroadcastVariationStats(ctx context.Context, workspaceID, broadcastID, variationID string) (*MessageHistoryStatusSum, error)
}

// MessageListParams contains parameters for listing messages with pagination and filtering
type MessageListParams struct {
	// Cursor-based pagination
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`

	// Filters
	Channel      string        `json:"channel,omitempty"`       // email, sms, push, etc.
	Status       MessageStatus `json:"status,omitempty"`        // message status filter
	ContactEmail string        `json:"contact_email,omitempty"` // filter by contact
	BroadcastID  string        `json:"broadcast_id,omitempty"`  // filter by broadcast
	TemplateID   string        `json:"template_id,omitempty"`   // filter by template
	HasError     *bool         `json:"has_error,omitempty"`     // filter messages with/without errors

	// Time range filters
	SentAfter     *time.Time `json:"sent_after,omitempty"`
	SentBefore    *time.Time `json:"sent_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
}

// FromQuery creates MessageListParams from HTTP query parameters
func (p *MessageListParams) FromQuery(query url.Values) error {
	// Parse cursor and basic string filters
	p.Cursor = query.Get("cursor")
	p.Channel = query.Get("channel")
	p.Status = MessageStatus(query.Get("status"))
	p.ContactEmail = query.Get("contact_email")
	p.BroadcastID = query.Get("broadcast_id")
	p.TemplateID = query.Get("template_id")

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		var limit int
		if err := json.Unmarshal([]byte(limitStr), &limit); err != nil {
			return fmt.Errorf("invalid limit value: %s", limitStr)
		}
		p.Limit = limit
	}

	// Parse hasError if provided
	if hasErrorStr := query.Get("has_error"); hasErrorStr != "" {
		var hasError bool
		if err := json.Unmarshal([]byte(hasErrorStr), &hasError); err != nil {
			return fmt.Errorf("invalid has_error value: %s", hasErrorStr)
		}
		p.HasError = &hasError
	}

	// Parse time filters if provided
	if err := parseTimeParam(query, "sent_after", &p.SentAfter); err != nil {
		return err
	}
	if err := parseTimeParam(query, "sent_before", &p.SentBefore); err != nil {
		return err
	}
	if err := parseTimeParam(query, "updated_after", &p.UpdatedAfter); err != nil {
		return err
	}
	if err := parseTimeParam(query, "updated_before", &p.UpdatedBefore); err != nil {
		return err
	}

	// Validate all parameters
	return p.Validate()
}

// Helper function to parse time parameters
func parseTimeParam(query url.Values, paramName string, target **time.Time) error {
	if paramStr := query.Get(paramName); paramStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, paramStr)
		if err != nil {
			return fmt.Errorf("invalid %s time format, expected RFC3339: %v", paramName, err)
		}
		*target = &parsedTime
	}
	return nil
}

func (p *MessageListParams) Validate() error {
	// Validate limit
	if p.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if p.Limit > 100 {
		p.Limit = 100 // Cap at maximum 100 items
	}
	if p.Limit == 0 {
		p.Limit = 20 // Default limit
	}

	// Validate channel
	if p.Channel != "" {
		// Use govalidator to check if channel is valid
		if !govalidator.IsIn(p.Channel, "email", "sms", "push") {
			return fmt.Errorf("invalid channel type: %s", p.Channel)
		}
	}

	// Validate status
	if p.Status != "" {
		validStatuses := []string{
			string(MessageStatusSent),
			string(MessageStatusDelivered),
			string(MessageStatusFailed),
			string(MessageStatusOpened),
			string(MessageStatusClicked),
			string(MessageStatusBounced),
			string(MessageStatusComplained),
			string(MessageStatusUnsubscribed),
		}
		if !govalidator.IsIn(string(p.Status), validStatuses...) {
			return fmt.Errorf("invalid message status: %s", p.Status)
		}
	}

	// Validate contact email if provided
	if p.ContactEmail != "" && !govalidator.IsEmail(p.ContactEmail) {
		return fmt.Errorf("invalid contact email format")
	}

	// Validate broadcast ID if provided
	if p.BroadcastID != "" && !govalidator.IsUUID(p.BroadcastID) {
		return fmt.Errorf("invalid broadcast ID format")
	}

	// Validate template ID if provided
	if p.TemplateID != "" && !govalidator.IsUUID(p.TemplateID) {
		return fmt.Errorf("invalid template ID format")
	}

	// Validate time ranges
	if p.SentAfter != nil && p.SentBefore != nil {
		if p.SentAfter.After(*p.SentBefore) {
			return fmt.Errorf("sent_after must be before sent_before")
		}
	}

	if p.UpdatedAfter != nil && p.UpdatedBefore != nil {
		if p.UpdatedAfter.After(*p.UpdatedBefore) {
			return fmt.Errorf("updated_after must be before updated_before")
		}
	}

	return nil
}

// MessageListResult contains the result of a ListMessages operation
type MessageListResult struct {
	Messages   []*MessageHistory `json:"messages"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}
