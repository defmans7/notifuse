package domain

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

//go:generate mockgen -destination=mocks/mock_message_history_repository.go -package=mocks -source=message.go MessageHistoryRepository

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
	ContactID       string        `json:"contact_id"`
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

// MessageHistoryRepository defines methods for message history persistence
type MessageHistoryRepository interface {
	// Create adds a new message history record
	Create(ctx context.Context, workspace string, message *MessageHistory) error

	// Update updates an existing message history record
	Update(ctx context.Context, workspace string, message *MessageHistory) error

	// Get retrieves a message history by ID
	Get(ctx context.Context, workspace, id string) (*MessageHistory, error)

	// GetByContact retrieves message history for a specific contact
	GetByContact(ctx context.Context, workspace, contactID string, limit, offset int) ([]*MessageHistory, int, error)

	// GetByBroadcast retrieves message history for a specific broadcast
	GetByBroadcast(ctx context.Context, workspace, broadcastID string, limit, offset int) ([]*MessageHistory, int, error)

	// UpdateStatus updates the status of a message and sets the corresponding timestamp
	UpdateStatus(ctx context.Context, workspace, id string, status MessageStatus, timestamp time.Time) error
}
