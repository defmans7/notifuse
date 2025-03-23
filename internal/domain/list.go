package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
)

// List represents a subscription list
type List struct {
	ID            string    `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name          string    `json:"name" valid:"required,stringlength(1|255)"`
	Type          string    `json:"type" valid:"required,in(public|private)"`
	IsDoubleOptin bool      `json:"is_double_optin" db:"is_double_optin"`
	Description   string    `json:"description,omitempty" valid:"optional"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
	Type          string
	IsDoubleOptin bool
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
		&dbl.Type,
		&dbl.IsDoubleOptin,
		&dbl.Description,
		&dbl.CreatedAt,
		&dbl.UpdatedAt,
	); err != nil {
		return nil, err
	}

	l := &List{
		ID:            dbl.ID,
		Name:          dbl.Name,
		Type:          dbl.Type,
		IsDoubleOptin: dbl.IsDoubleOptin,
		Description:   dbl.Description,
		CreatedAt:     dbl.CreatedAt,
		UpdatedAt:     dbl.UpdatedAt,
	}

	return l, nil
}

// ListService provides operations for managing lists
type ListService interface {
	// CreateList creates a new list
	CreateList(ctx context.Context, list *List) error

	// GetListByID retrieves a list by ID
	GetListByID(ctx context.Context, id string) (*List, error)

	// GetLists retrieves all lists
	GetLists(ctx context.Context) ([]*List, error)

	// UpdateList updates an existing list
	UpdateList(ctx context.Context, list *List) error

	// DeleteList deletes a list by ID
	DeleteList(ctx context.Context, id string) error
}

type ListRepository interface {
	// CreateList creates a new list in the database
	CreateList(ctx context.Context, list *List) error

	// GetListByID retrieves a list by its ID
	GetListByID(ctx context.Context, id string) (*List, error)

	// GetLists retrieves all lists
	GetLists(ctx context.Context) ([]*List, error)

	// UpdateList updates an existing list
	UpdateList(ctx context.Context, list *List) error

	// DeleteList deletes a list
	DeleteList(ctx context.Context, id string) error
}

// ErrListNotFound is returned when a list is not found
type ErrListNotFound struct {
	Message string
}

func (e *ErrListNotFound) Error() string {
	return e.Message
}
