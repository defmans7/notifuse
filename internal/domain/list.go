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
	ID                string    `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name              string    `json:"name" valid:"required,stringlength(1|255)"`
	IsDoubleOptin     bool      `json:"is_double_optin" db:"is_double_optin"`
	IsPublic          bool      `json:"is_public" db:"is_public"`
	Description       string    `json:"description,omitempty" valid:"optional"`
	TotalActive       int       `json:"total_active" db:"total_active"`
	TotalPending      int       `json:"total_pending" db:"total_pending"`
	TotalUnsubscribed int       `json:"total_unsubscribed" db:"total_unsubscribed"`
	TotalBounced      int       `json:"total_bounced" db:"total_bounced"`
	TotalComplained   int       `json:"total_complained" db:"total_complained"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Validate performs validation on the list fields
func (l *List) Validate() error {
	if _, err := govalidator.ValidateStruct(l); err != nil {
		return fmt.Errorf("invalid list: %w", err)
	}
	return nil
}

// For database scanning
type dbList struct {
	ID            string
	Name          string
	IsDoubleOptin bool
	IsPublic      bool
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
		&dbl.CreatedAt,
		&dbl.UpdatedAt,
	); err != nil {
		return nil, err
	}

	l := &List{
		ID:            dbl.ID,
		Name:          dbl.Name,
		IsDoubleOptin: dbl.IsDoubleOptin,
		IsPublic:      dbl.IsPublic,
		Description:   dbl.Description,
		CreatedAt:     dbl.CreatedAt,
		UpdatedAt:     dbl.UpdatedAt,
	}

	return l, nil
}

// Request/Response types
type CreateListRequest struct {
	WorkspaceID   string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
	ID            string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name          string `json:"name" valid:"required,stringlength(1|255)"`
	IsDoubleOptin bool   `json:"is_double_optin"`
	IsPublic      bool   `json:"is_public"`
	Description   string `json:"description,omitempty"`
}

func (r *CreateListRequest) Validate() (list *List, workspaceID string, err error) {

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return nil, "", fmt.Errorf("invalid create list request: %w", err)
	}

	return &List{
		ID:            r.ID,
		Name:          r.Name,
		IsDoubleOptin: r.IsDoubleOptin,
		IsPublic:      r.IsPublic,
		Description:   r.Description,
	}, r.WorkspaceID, nil
}

type GetListsRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
}

func (r *GetListsRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid get lists request: %w", err)
	}
	return nil
}

type GetListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
	ID          string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
}

func (r *GetListRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.ID = queryParams.Get("id")

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid get list request: %w", err)
	}
	return nil
}

type UpdateListRequest struct {
	WorkspaceID   string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
	ID            string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name          string `json:"name" valid:"required,stringlength(1|255)"`
	IsDoubleOptin bool   `json:"is_double_optin"`
	IsPublic      bool   `json:"is_public"`
	Description   string `json:"description,omitempty"`
}

func (r *UpdateListRequest) Validate() (list *List, workspaceID string, err error) {
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return nil, "", fmt.Errorf("invalid update list request: %w", err)
	}

	return &List{
		ID:            r.ID,
		Name:          r.Name,
		IsDoubleOptin: r.IsDoubleOptin,
		IsPublic:      r.IsPublic,
		Description:   r.Description,
	}, r.WorkspaceID, nil
}

type DeleteListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
	ID          string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
}

func (r *DeleteListRequest) Validate() (workspaceID string, err error) {
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return "", fmt.Errorf("invalid delete list request: %w", err)
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
