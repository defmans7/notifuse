package domain

import (
	"context"
)

// ContactService provides operations for managing contacts
type ContactService interface {
	// CreateContact creates a new contact
	CreateContact(ctx context.Context, contact *Contact) error

	// GetContactByUUID retrieves a contact by its UUID
	GetContactByUUID(ctx context.Context, uuid string) (*Contact, error)

	// GetContactByEmail retrieves a contact by email
	GetContactByEmail(ctx context.Context, email string) (*Contact, error)

	// GetContactByExternalID retrieves a contact by external ID
	GetContactByExternalID(ctx context.Context, externalID string) (*Contact, error)

	// GetContacts retrieves all contacts
	GetContacts(ctx context.Context) ([]*Contact, error)

	// UpdateContact updates an existing contact
	UpdateContact(ctx context.Context, contact *Contact) error

	// DeleteContact deletes a contact by UUID
	DeleteContact(ctx context.Context, uuid string) error
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

// ContactListService provides operations for managing contact list relationships
type ContactListService interface {
	// AddContactToList adds a contact to a list
	AddContactToList(ctx context.Context, contactList *ContactList) error

	// GetContactListByIDs retrieves a contact list by contact ID and list ID
	GetContactListByIDs(ctx context.Context, contactID, listID string) (*ContactList, error)

	// GetContactsByListID retrieves all contacts for a list
	GetContactsByListID(ctx context.Context, listID string) ([]*ContactList, error)

	// GetListsByContactID retrieves all lists for a contact
	GetListsByContactID(ctx context.Context, contactID string) ([]*ContactList, error)

	// UpdateContactListStatus updates the status of a contact on a list
	UpdateContactListStatus(ctx context.Context, contactID, listID string, status ContactListStatus) error

	// RemoveContactFromList removes a contact from a list
	RemoveContactFromList(ctx context.Context, contactID, listID string) error
}
