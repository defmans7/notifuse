package domain

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_list_service.go -package mocks github.com/Notifuse/notifuse/internal/domain ListService
//go:generate mockgen -destination mocks/mock_list_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain ListRepository

// List represents a subscription list
type List struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	IsDoubleOptin       bool               `json:"is_double_optin" db:"is_double_optin"`
	IsPublic            bool               `json:"is_public" db:"is_public"`
	Description         string             `json:"description,omitempty"`
	TotalActive         int                `json:"total_active" db:"total_active"`
	TotalPending        int                `json:"total_pending" db:"total_pending"`
	TotalUnsubscribed   int                `json:"total_unsubscribed" db:"total_unsubscribed"`
	TotalBounced        int                `json:"total_bounced" db:"total_bounced"`
	TotalComplained     int                `json:"total_complained" db:"total_complained"`
	DoubleOptInTemplate *TemplateReference `json:"double_optin_template,omitempty"`
	WelcomeTemplate     *TemplateReference `json:"welcome_template,omitempty"`
	UnsubscribeTemplate *TemplateReference `json:"unsubscribe_template,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// Validate performs validation on the list fields
func (l *List) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("invalid list: id is required")
	}
	if !govalidator.IsAlphanumeric(l.ID) {
		return fmt.Errorf("invalid list: id must be alphanumeric")
	}
	if len(l.ID) > 20 {
		return fmt.Errorf("invalid list: id length must be between 1 and 20")
	}

	if l.Name == "" {
		return fmt.Errorf("invalid list: name is required")
	}
	if len(l.Name) > 255 {
		return fmt.Errorf("invalid list: name length must be between 1 and 255")
	}

	// Validate optional template references if they exist
	if l.DoubleOptInTemplate != nil {
		if err := l.DoubleOptInTemplate.Validate(); err != nil {
			return fmt.Errorf("invalid list: double opt-in template: %w", err)
		}
	}

	if l.WelcomeTemplate != nil {
		if err := l.WelcomeTemplate.Validate(); err != nil {
			return fmt.Errorf("invalid list: welcome template: %w", err)
		}
	}

	if l.UnsubscribeTemplate != nil {
		if err := l.UnsubscribeTemplate.Validate(); err != nil {
			return fmt.Errorf("invalid list: unsubscribe template: %w", err)
		}
	}

	return nil
}

// For database scanning
type dbList struct {
	ID                  string
	Name                string
	IsDoubleOptin       bool
	IsPublic            bool
	Description         string
	TotalActive         int
	TotalPending        int
	TotalUnsubscribed   int
	TotalBounced        int
	TotalComplained     int
	DoubleOptInTemplate *TemplateReference
	WelcomeTemplate     *TemplateReference
	UnsubscribeTemplate *TemplateReference
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ScanList scans a list from the database
func ScanList(scanner interface {
	Scan(dest ...interface{}) error
}) (*List, error) {
	var dbl dbList
	if err := scanner.Scan(
		&dbl.ID,
		&dbl.Name,
		&dbl.IsDoubleOptin,
		&dbl.IsPublic,
		&dbl.Description,
		&dbl.TotalActive,
		&dbl.TotalPending,
		&dbl.TotalUnsubscribed,
		&dbl.TotalBounced,
		&dbl.TotalComplained,
		&dbl.DoubleOptInTemplate,
		&dbl.WelcomeTemplate,
		&dbl.UnsubscribeTemplate,
		&dbl.CreatedAt,
		&dbl.UpdatedAt,
	); err != nil {
		return nil, err
	}

	l := &List{
		ID:                  dbl.ID,
		Name:                dbl.Name,
		IsDoubleOptin:       dbl.IsDoubleOptin,
		IsPublic:            dbl.IsPublic,
		Description:         dbl.Description,
		TotalActive:         dbl.TotalActive,
		TotalPending:        dbl.TotalPending,
		TotalUnsubscribed:   dbl.TotalUnsubscribed,
		TotalBounced:        dbl.TotalBounced,
		TotalComplained:     dbl.TotalComplained,
		DoubleOptInTemplate: dbl.DoubleOptInTemplate,
		WelcomeTemplate:     dbl.WelcomeTemplate,
		UnsubscribeTemplate: dbl.UnsubscribeTemplate,
		CreatedAt:           dbl.CreatedAt,
		UpdatedAt:           dbl.UpdatedAt,
	}

	return l, nil
}

// Request/Response types
type CreateListRequest struct {
	WorkspaceID         string             `json:"workspace_id"`
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	IsDoubleOptin       bool               `json:"is_double_optin"`
	IsPublic            bool               `json:"is_public"`
	Description         string             `json:"description,omitempty"`
	DoubleOptInTemplate *TemplateReference `json:"double_optin_template,omitempty"`
	WelcomeTemplate     *TemplateReference `json:"welcome_template,omitempty"`
	UnsubscribeTemplate *TemplateReference `json:"unsubscribe_template,omitempty"`
}

func (r *CreateListRequest) Validate() (list *List, workspaceID string, err error) {
	if r.WorkspaceID == "" {
		return nil, "", fmt.Errorf("invalid create list request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return nil, "", fmt.Errorf("invalid create list request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 20 {
		return nil, "", fmt.Errorf("invalid create list request: workspace_id length must be between 1 and 20")
	}

	if r.ID == "" {
		return nil, "", fmt.Errorf("invalid create list request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return nil, "", fmt.Errorf("invalid create list request: id must be alphanumeric")
	}
	if len(r.ID) > 20 {
		return nil, "", fmt.Errorf("invalid create list request: id length must be between 1 and 20")
	}

	if r.Name == "" {
		return nil, "", fmt.Errorf("invalid create list request: name is required")
	}
	if len(r.Name) > 255 {
		return nil, "", fmt.Errorf("invalid create list request: name length must be between 1 and 255")
	}

	// Validate optional template references if they exist
	if r.DoubleOptInTemplate != nil {
		if err := r.DoubleOptInTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid create list request: double opt-in template: %w", err)
		}
	}

	if r.WelcomeTemplate != nil {
		if err := r.WelcomeTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid create list request: welcome template: %w", err)
		}
	}

	if r.UnsubscribeTemplate != nil {
		if err := r.UnsubscribeTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid create list request: unsubscribe template: %w", err)
		}
	}

	return &List{
		ID:                  r.ID,
		Name:                r.Name,
		IsDoubleOptin:       r.IsDoubleOptin,
		IsPublic:            r.IsPublic,
		Description:         r.Description,
		DoubleOptInTemplate: r.DoubleOptInTemplate,
		WelcomeTemplate:     r.WelcomeTemplate,
		UnsubscribeTemplate: r.UnsubscribeTemplate,
	}, r.WorkspaceID, nil
}

type GetListsRequest struct {
	WorkspaceID string `json:"workspace_id"`
}

func (r *GetListsRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")

	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid get lists request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return fmt.Errorf("invalid get lists request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 20 {
		return fmt.Errorf("invalid get lists request: workspace_id length must be between 1 and 20")
	}

	return nil
}

type GetListRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

func (r *GetListRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.ID = queryParams.Get("id")

	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid get list request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return fmt.Errorf("invalid get list request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 20 {
		return fmt.Errorf("invalid get list request: workspace_id length must be between 1 and 20")
	}

	if r.ID == "" {
		return fmt.Errorf("invalid get list request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return fmt.Errorf("invalid get list request: id must be alphanumeric")
	}
	if len(r.ID) > 20 {
		return fmt.Errorf("invalid get list request: id length must be between 1 and 20")
	}

	return nil
}

type UpdateListRequest struct {
	WorkspaceID         string             `json:"workspace_id"`
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	IsDoubleOptin       bool               `json:"is_double_optin"`
	IsPublic            bool               `json:"is_public"`
	Description         string             `json:"description,omitempty"`
	DoubleOptInTemplate *TemplateReference `json:"double_optin_template,omitempty"`
	WelcomeTemplate     *TemplateReference `json:"welcome_template,omitempty"`
	UnsubscribeTemplate *TemplateReference `json:"unsubscribe_template,omitempty"`
}

func (r *UpdateListRequest) Validate() (list *List, workspaceID string, err error) {
	if r.WorkspaceID == "" {
		return nil, "", fmt.Errorf("invalid update list request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return nil, "", fmt.Errorf("invalid update list request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 20 {
		return nil, "", fmt.Errorf("invalid update list request: workspace_id length must be between 1 and 20")
	}

	if r.ID == "" {
		return nil, "", fmt.Errorf("invalid update list request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return nil, "", fmt.Errorf("invalid update list request: id must be alphanumeric")
	}
	if len(r.ID) > 20 {
		return nil, "", fmt.Errorf("invalid update list request: id length must be between 1 and 20")
	}

	if r.Name == "" {
		return nil, "", fmt.Errorf("invalid update list request: name is required")
	}
	if len(r.Name) > 255 {
		return nil, "", fmt.Errorf("invalid update list request: name length must be between 1 and 255")
	}

	// Validate optional template references if they exist
	if r.DoubleOptInTemplate != nil {
		if err := r.DoubleOptInTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid update list request: double opt-in template: %w", err)
		}
	}

	if r.WelcomeTemplate != nil {
		if err := r.WelcomeTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid update list request: welcome template: %w", err)
		}
	}

	if r.UnsubscribeTemplate != nil {
		if err := r.UnsubscribeTemplate.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid update list request: unsubscribe template: %w", err)
		}
	}

	return &List{
		ID:                  r.ID,
		Name:                r.Name,
		IsDoubleOptin:       r.IsDoubleOptin,
		IsPublic:            r.IsPublic,
		Description:         r.Description,
		DoubleOptInTemplate: r.DoubleOptInTemplate,
		WelcomeTemplate:     r.WelcomeTemplate,
		UnsubscribeTemplate: r.UnsubscribeTemplate,
	}, r.WorkspaceID, nil
}

type DeleteListRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

func (r *DeleteListRequest) Validate() (workspaceID string, err error) {
	if r.WorkspaceID == "" {
		return "", fmt.Errorf("invalid delete list request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return "", fmt.Errorf("invalid delete list request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 20 {
		return "", fmt.Errorf("invalid delete list request: workspace_id length must be between 1 and 20")
	}

	if r.ID == "" {
		return "", fmt.Errorf("invalid delete list request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return "", fmt.Errorf("invalid delete list request: id must be alphanumeric")
	}
	if len(r.ID) > 20 {
		return "", fmt.Errorf("invalid delete list request: id length must be between 1 and 20")
	}

	return r.WorkspaceID, nil
}

// ContactListTotalType represents the type of total to increment/decrement
type ContactListTotalType string

const (
	TotalTypePending      ContactListTotalType = "pending"
	TotalTypeUnsubscribed ContactListTotalType = "unsubscribed"
	TotalTypeBounced      ContactListTotalType = "bounced"
	TotalTypeComplained   ContactListTotalType = "complained"
	TotalTypeActive       ContactListTotalType = "active"
)

// Validate checks if the total type is valid
func (ct ContactListTotalType) Validate() error {
	switch ct {
	case TotalTypePending, TotalTypeUnsubscribed, TotalTypeBounced, TotalTypeComplained, TotalTypeActive:
		return nil
	default:
		return fmt.Errorf("invalid total type: %s", ct)
	}
}

// ListService provides operations for managing lists
type ListService interface {
	// CreateList creates a new list
	CreateList(ctx context.Context, workspaceID string, list *List) error

	// GetListByID retrieves a list by ID
	GetListByID(ctx context.Context, workspaceID string, id string) (*List, error)

	// GetLists retrieves all lists
	GetLists(ctx context.Context, workspaceID string) ([]*List, error)

	// UpdateList updates an existing list
	UpdateList(ctx context.Context, workspaceID string, list *List) error

	// DeleteList deletes a list by ID
	DeleteList(ctx context.Context, workspaceID string, id string) error
}

type ListRepository interface {
	// CreateList creates a new list in the database
	CreateList(ctx context.Context, workspaceID string, list *List) error

	// GetListByID retrieves a list by its ID
	GetListByID(ctx context.Context, workspaceID string, id string) (*List, error)

	// GetLists retrieves all lists
	GetLists(ctx context.Context, workspaceID string) ([]*List, error)

	// UpdateList updates an existing list
	UpdateList(ctx context.Context, workspaceID string, list *List) error

	// DeleteList deletes a list
	DeleteList(ctx context.Context, workspaceID string, id string) error

	// IncrementTotal increments the specified total type for a list
	IncrementTotal(ctx context.Context, workspaceID string, listID string, totalType ContactListTotalType) error

	// DecrementTotal decrements the specified total type for a list
	DecrementTotal(ctx context.Context, workspaceID string, listID string, totalType ContactListTotalType) error
}

// ErrListNotFound is returned when a list is not found
type ErrListNotFound struct {
	Message string
}

func (e *ErrListNotFound) Error() string {
	return e.Message
}
