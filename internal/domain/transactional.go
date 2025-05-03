package domain

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// TransactionalChannel represents supported notification channels
type TransactionalChannel string

const (
	// TransactionalChannelEmail for email notifications
	TransactionalChannelEmail TransactionalChannel = "email"
	// Add other channels in the future (sms, push, etc.)
)

// TransactionalStatus represents the status of a transactional notification
type TransactionalStatus string

const (
	// TransactionalStatusActive indicates the notification is active and can be triggered
	TransactionalStatusActive TransactionalStatus = "active"
	// TransactionalStatusInactive indicates the notification is inactive and cannot be triggered
	TransactionalStatusInactive TransactionalStatus = "inactive"
	// TransactionalStatusDraft indicates the notification is still in draft mode
	TransactionalStatusDraft TransactionalStatus = "draft"
)

// ChannelTemplate represents template configuration for a specific channel
type ChannelTemplate struct {
	TemplateID string   `json:"template_id"`
	Version    int      `json:"version"`
	Settings   MapOfAny `json:"settings,omitempty"`
}

// ChannelTemplates maps channels to their template configurations
type ChannelTemplates map[TransactionalChannel]ChannelTemplate

// Value implements the driver.Valuer interface for database storage
func (ct ChannelTemplates) Value() (driver.Value, error) {
	return json.Marshal(ct)
}

// Scan implements the sql.Scanner interface for database retrieval
func (ct *ChannelTemplates) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(v)
	return json.Unmarshal(cloned, ct)
}

// TransactionalNotification represents a transactional notification configuration
type TransactionalNotification struct {
	ID          string              `json:"id"` // Unique identifier for the notification, also used for API triggering
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Channels    ChannelTemplates    `json:"channels"`
	Status      TransactionalStatus `json:"status"`
	IsPublic    bool                `json:"is_public"` // Indicates if the notification is publicly accessible
	Metadata    MapOfAny            `json:"metadata,omitempty"`

	// System timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// TransactionalNotificationRepository defines methods for transactional notification persistence
type TransactionalNotificationRepository interface {
	// Create adds a new transactional notification
	Create(ctx context.Context, workspace string, notification *TransactionalNotification) error

	// Update updates an existing transactional notification
	Update(ctx context.Context, workspace string, notification *TransactionalNotification) error

	// Get retrieves a transactional notification by ID
	Get(ctx context.Context, workspace, id string) (*TransactionalNotification, error)

	// List retrieves all transactional notifications with optional filtering
	List(ctx context.Context, workspace string, filter map[string]interface{}, limit, offset int) ([]*TransactionalNotification, int, error)

	// Delete soft-deletes a transactional notification
	Delete(ctx context.Context, workspace, id string) error
}

// TransactionalNotificationCreateParams contains the parameters for creating a new transactional notification
type TransactionalNotificationCreateParams struct {
	ID          string              `json:"id" validate:"required"` // Unique identifier for API triggering
	Name        string              `json:"name" validate:"required"`
	Description string              `json:"description"`
	Channels    ChannelTemplates    `json:"channels" validate:"required,min=1"`
	Status      TransactionalStatus `json:"status" validate:"required"`
	IsPublic    bool                `json:"is_public"`
	Metadata    MapOfAny            `json:"metadata,omitempty"`
}

// TransactionalNotificationUpdateParams contains the parameters for updating an existing transactional notification
type TransactionalNotificationUpdateParams struct {
	Name        string              `json:"name,omitempty"`
	Description string              `json:"description,omitempty"`
	Channels    ChannelTemplates    `json:"channels,omitempty"`
	Status      TransactionalStatus `json:"status,omitempty"`
	IsPublic    *bool               `json:"is_public,omitempty"`
	Metadata    MapOfAny            `json:"metadata,omitempty"`
}

// TransactionalNotificationSendParams contains the parameters for sending a transactional notification
type TransactionalNotificationSendParams struct {
	ID       string                 `json:"id" validate:"required"`      // ID of the notification to send
	Contact  *Contact               `json:"contact" validate:"required"` // Contact to send the notification to
	Channels []TransactionalChannel `json:"channels,omitempty"`          // Specific channels to send through (if empty, use all configured channels)
	Data     MapOfAny               `json:"data,omitempty"`              // Data to populate the template with
	Metadata MapOfAny               `json:"metadata,omitempty"`          // Additional metadata for tracking
}

// TransactionalNotificationService defines the interface for transactional notification operations
type TransactionalNotificationService interface {
	// CreateNotification creates a new transactional notification
	CreateNotification(ctx context.Context, workspace string, params TransactionalNotificationCreateParams) (*TransactionalNotification, error)

	// UpdateNotification updates an existing transactional notification
	UpdateNotification(ctx context.Context, workspace, id string, params TransactionalNotificationUpdateParams) (*TransactionalNotification, error)

	// GetNotification retrieves a transactional notification by ID
	GetNotification(ctx context.Context, workspace, id string) (*TransactionalNotification, error)

	// ListNotifications retrieves all transactional notifications with optional filtering
	ListNotifications(ctx context.Context, workspace string, filter map[string]interface{}, limit, offset int) ([]*TransactionalNotification, int, error)

	// DeleteNotification soft-deletes a transactional notification
	DeleteNotification(ctx context.Context, workspace, id string) error

	// SendNotification sends a transactional notification to a contact
	SendNotification(ctx context.Context, workspace string, params TransactionalNotificationSendParams) (string, error)
}

// Request and response types for transactional notifications

// ListTransactionalRequest represents a request to list transactional notifications
type ListTransactionalRequest struct {
	WorkspaceID string                 `json:"workspace_id"`
	Status      string                 `json:"status,omitempty"`
	Search      string                 `json:"search,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

// FromURLParams populates the request from URL query parameters
func (req *ListTransactionalRequest) FromURLParams(values map[string][]string) error {
	req.WorkspaceID = getFirstValue(values, "workspace_id")
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	req.Status = getFirstValue(values, "status")
	req.Search = getFirstValue(values, "search")

	if limitStr := getFirstValue(values, "limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if offsetStr := getFirstValue(values, "offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	// Convert status and search to filter if provided
	if req.Filter == nil {
		req.Filter = make(map[string]interface{})
	}
	if req.Status != "" {
		req.Filter["status"] = req.Status
	}
	if req.Search != "" {
		req.Filter["search"] = req.Search
	}

	return nil
}

// GetTransactionalRequest represents a request to get a transactional notification
type GetTransactionalRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// FromURLParams populates the request from URL query parameters
func (req *GetTransactionalRequest) FromURLParams(values map[string][]string) error {
	req.WorkspaceID = getFirstValue(values, "workspace_id")
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	req.ID = getFirstValue(values, "id")
	if req.ID == "" {
		return NewValidationError("id is required")
	}

	return nil
}

// CreateTransactionalRequest represents a request to create a transactional notification
type CreateTransactionalRequest struct {
	WorkspaceID  string                                `json:"workspace_id"`
	Notification TransactionalNotificationCreateParams `json:"notification"`
}

// Validate validates the create request
func (req *CreateTransactionalRequest) Validate() error {
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	if req.Notification.ID == "" {
		return NewValidationError("notification.id is required")
	}

	if req.Notification.Name == "" {
		return NewValidationError("notification.name is required")
	}

	if len(req.Notification.Channels) == 0 {
		return NewValidationError("notification must have at least one channel")
	}

	if req.Notification.Status == "" {
		return NewValidationError("notification.status is required")
	}

	return nil
}

// UpdateTransactionalRequest represents a request to update a transactional notification
type UpdateTransactionalRequest struct {
	WorkspaceID string                                `json:"workspace_id"`
	ID          string                                `json:"id"`
	Updates     TransactionalNotificationUpdateParams `json:"updates"`
}

// Validate validates the update request
func (req *UpdateTransactionalRequest) Validate() error {
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	if req.ID == "" {
		return NewValidationError("id is required")
	}

	// At least one field must be updated
	if req.Updates.Name == "" &&
		req.Updates.Description == "" &&
		req.Updates.Status == "" &&
		req.Updates.Channels == nil &&
		req.Updates.Metadata == nil {
		return NewValidationError("at least one field must be updated")
	}

	return nil
}

// DeleteTransactionalRequest represents a request to delete a transactional notification
type DeleteTransactionalRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// Validate validates the delete request
func (req *DeleteTransactionalRequest) Validate() error {
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	if req.ID == "" {
		return NewValidationError("id is required")
	}

	return nil
}

// SendTransactionalRequest represents a request to send a transactional notification
type SendTransactionalRequest struct {
	WorkspaceID  string                              `json:"workspace_id"`
	Notification TransactionalNotificationSendParams `json:"notification"`
}

// Validate validates the send request
func (req *SendTransactionalRequest) Validate() error {
	if req.WorkspaceID == "" {
		return NewValidationError("workspace_id is required")
	}

	if req.Notification.ID == "" {
		return NewValidationError("notification.id is required")
	}

	if req.Notification.Contact == nil {
		return NewValidationError("notification.contact is required")
	}

	if req.Notification.Contact.Validate() != nil {
		return NewValidationError("notification.contact is invalid")
	}

	return nil
}

// Helper function to get the first value from a map of string slices
func getFirstValue(values map[string][]string, key string) string {
	if vals, ok := values[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}
