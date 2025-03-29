package domain

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_contact_list_service.go -package mocks github.com/Notifuse/notifuse/internal/domain ContactListService
//go:generate mockgen -destination mocks/mock_contact_list_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain ContactListRepository

// ContactListStatus represents the status of a contact's subscription to a list
type ContactListStatus string

const (
	// ContactListStatusActive indicates an active subscription
	ContactListStatusActive ContactListStatus = "active"
	// ContactListStatusPending indicates a pending subscription (e.g., waiting for double opt-in)
	ContactListStatusPending ContactListStatus = "pending"
	// ContactListStatusUnsubscribed indicates an unsubscribed status
	ContactListStatusUnsubscribed ContactListStatus = "unsubscribed"
	// ContactListStatusBounced indicates the contact's email has bounced
	ContactListStatusBounced ContactListStatus = "bounced"
	// ContactListStatusComplained indicates the contact has complained (e.g., marked as spam)
	ContactListStatusComplained ContactListStatus = "complained"
)

// ContactList represents the relationship between a contact and a list
type ContactList struct {
	Email     string            `json:"email" valid:"required,email"`
	ListID    string            `json:"list_id" valid:"required,alphanum,stringlength(1|20)"`
	Status    ContactListStatus `json:"status" valid:"required,in(active|pending|unsubscribed|bounced|complained)"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Validate performs validation on the contact list fields
func (cl *ContactList) Validate() error {
	// Check required fields first
	if cl.Email == "" {
		return fmt.Errorf("email is required")
	}
	if cl.ListID == "" {
		return fmt.Errorf("list_id is required")
	}
	if cl.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Then use govalidator for additional validation
	if _, err := govalidator.ValidateStruct(cl); err != nil {
		return fmt.Errorf("invalid contact list: %w", err)
	}
	return nil
}

// For database scanning
type dbContactList struct {
	Email     string
	ListID    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ScanContactList scans a contact list from the database
func ScanContactList(scanner interface {
	Scan(dest ...interface{}) error
}) (*ContactList, error) {
	var dbcl dbContactList
	if err := scanner.Scan(
		&dbcl.Email,
		&dbcl.ListID,
		&dbcl.Status,
		&dbcl.CreatedAt,
		&dbcl.UpdatedAt,
	); err != nil {
		return nil, err
	}

	cl := &ContactList{
		Email:     dbcl.Email,
		ListID:    dbcl.ListID,
		Status:    ContactListStatus(dbcl.Status),
		CreatedAt: dbcl.CreatedAt,
		UpdatedAt: dbcl.UpdatedAt,
	}

	return cl, nil
}

// Request/Response types
type AddContactToListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	Email       string `json:"email" valid:"required,email"`
	ListID      string `json:"list_id" valid:"required"`
	Status      string `json:"status" valid:"required,in(active|pending|unsubscribed|blacklisted)"`
}

func (r *AddContactToListRequest) Validate() (contactList *ContactList, workspaceID string, err error) {
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return nil, "", fmt.Errorf("invalid add contact to list request: %w", err)
	}
	return &ContactList{
		Email:  r.Email,
		ListID: r.ListID,
		Status: ContactListStatus(r.Status),
	}, r.WorkspaceID, nil
}

type GetContactListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	Email       string `json:"email" valid:"required,email"`
	ListID      string `json:"list_id" valid:"required"`
}

func (r *GetContactListRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.Email = queryParams.Get("email")
	r.ListID = queryParams.Get("list_id")

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid get contact list request: %w", err)
	}

	return nil
}

type GetContactsByListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	ListID      string `json:"list_id" valid:"required,alphanum"`
}

func (r *GetContactsByListRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.ListID = queryParams.Get("list_id")

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid get contacts by list request: %w", err)
	}

	return nil
}

type GetListsByContactRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	Email       string `json:"email" valid:"required,email"`
}

func (r *GetListsByContactRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.Email = queryParams.Get("email")

	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid get lists by contact request: %w", err)
	}

	return nil
}

type UpdateContactListStatusRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	Email       string `json:"email" valid:"required,email"`
	ListID      string `json:"list_id" valid:"required"`
	Status      string `json:"status" valid:"required,in(active|pending|unsubscribed|blacklisted)"`
}

func (r *UpdateContactListStatusRequest) Validate() (workspaceID string, list *ContactList, err error) {
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return "", nil, fmt.Errorf("invalid update contact list status request: %w", err)
	}
	return r.WorkspaceID, &ContactList{
		Email:  r.Email,
		ListID: r.ListID,
		Status: ContactListStatus(r.Status),
	}, nil
}

type RemoveContactFromListRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required"`
	Email       string `json:"email" valid:"required,email"`
	ListID      string `json:"list_id" valid:"required"`
}

func (r *RemoveContactFromListRequest) Validate() (err error) {
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return fmt.Errorf("invalid remove contact from list request: %w", err)
	}
	return nil
}

// ContactListService provides operations for managing contact list relationships
type ContactListService interface {
	// AddContactToList adds a contact to a list
	AddContactToList(ctx context.Context, workspaceID string, contactList *ContactList) error

	// GetContactListByIDs retrieves a contact list by email and list ID
	GetContactListByIDs(ctx context.Context, workspaceID string, email, listID string) (*ContactList, error)

	// GetContactsByListID retrieves all contacts for a list
	GetContactsByListID(ctx context.Context, workspaceID string, listID string) ([]*ContactList, error)

	// GetListsByEmail retrieves all lists for a contact
	GetListsByEmail(ctx context.Context, workspaceID string, email string) ([]*ContactList, error)

	// UpdateContactListStatus updates the status of a contact on a list
	UpdateContactListStatus(ctx context.Context, workspaceID string, email, listID string, status ContactListStatus) error

	// RemoveContactFromList removes a contact from a list
	RemoveContactFromList(ctx context.Context, workspaceID string, email, listID string) error
}

type ContactListRepository interface {
	// AddContactToList adds a contact to a list
	AddContactToList(ctx context.Context, workspaceID string, contactList *ContactList) error

	// GetContactListByIDs retrieves a contact list by email and list ID
	GetContactListByIDs(ctx context.Context, workspaceID string, email, listID string) (*ContactList, error)

	// GetContactsByListID retrieves all contacts for a list
	GetContactsByListID(ctx context.Context, workspaceID string, listID string) ([]*ContactList, error)

	// GetListsByEmail retrieves all lists for a contact
	GetListsByEmail(ctx context.Context, workspaceID string, email string) ([]*ContactList, error)

	// UpdateContactListStatus updates the status of a contact on a list
	UpdateContactListStatus(ctx context.Context, workspaceID string, email, listID string, status ContactListStatus) error

	// RemoveContactFromList removes a contact from a list
	RemoveContactFromList(ctx context.Context, workspaceID string, email, listID string) error
}

// ErrContactListNotFound is returned when a contact list is not found
type ErrContactListNotFound struct {
	Message string
}

func (e *ErrContactListNotFound) Error() string {
	return e.Message
}
