package service

import (
	"context"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

// MockContactService is a mock implementation of domain.ContactService for testing
type MockContactService struct {
	Contacts                   map[string]*domain.Contact
	ErrToReturn                error
	ErrContactNotFoundToReturn bool

	GetContactsCalled            bool
	GetContactByEmailCalled      bool
	LastContactEmail             string
	GetContactByExternalIDCalled bool
	LastContactExternalID        string
	DeleteContactCalled          bool
	BatchImportContactsCalled    bool
	LastContactsBatchImported    []*domain.Contact
	UpsertContactCalled          bool
	LastContactUpserted          *domain.Contact
	UpsertIsNewToReturn          bool
}

// NewMockContactService creates a new mock contact service for testing
func NewMockContactService() *MockContactService {
	return &MockContactService{
		Contacts: make(map[string]*domain.Contact),
	}
}

func (m *MockContactService) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	m.GetContactsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Convert map to slice
	contacts := make([]*domain.Contact, 0, len(m.Contacts))
	for _, contact := range m.Contacts {
		contacts = append(contacts, contact)
	}

	// For testing purposes, we'll just return all contacts
	// In a real implementation, we would handle pagination and filtering
	return &domain.GetContactsResponse{
		Contacts:   contacts,
		NextCursor: "", // For testing, we don't implement cursor pagination
	}, nil
}

func (m *MockContactService) GetContactByEmail(ctx context.Context, workspaceID string, email string) (*domain.Contact, error) {
	m.GetContactByEmailCalled = true
	m.LastContactEmail = email
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return nil, &domain.ErrContactNotFound{}
	}

	contact, exists := m.Contacts[email]
	if !exists {
		return nil, &domain.ErrContactNotFound{}
	}
	return contact, nil
}

func (m *MockContactService) GetContactByExternalID(ctx context.Context, workspaceID string, externalID string) (*domain.Contact, error) {
	m.GetContactByExternalIDCalled = true
	m.LastContactExternalID = externalID
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return nil, &domain.ErrContactNotFound{}
	}

	for _, contact := range m.Contacts {
		if contact.ExternalID != nil && !contact.ExternalID.IsNull && contact.ExternalID.String == externalID {
			return contact, nil
		}
	}
	return nil, &domain.ErrContactNotFound{}
}

func (m *MockContactService) DeleteContact(ctx context.Context, workspaceID string, email string) error {
	m.DeleteContactCalled = true
	m.LastContactEmail = email
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return &domain.ErrContactNotFound{}
	}

	if _, exists := m.Contacts[email]; !exists {
		return &domain.ErrContactNotFound{}
	}
	delete(m.Contacts, email)
	return nil
}

func (m *MockContactService) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {
	m.BatchImportContactsCalled = true
	m.LastContactsBatchImported = contacts
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps for all contacts
	now := time.Now()
	for _, contact := range contacts {
		if contact.CreatedAt.IsZero() {
			contact.CreatedAt = now
		}
		contact.UpdatedAt = now

		// Store in the map
		m.Contacts[contact.Email] = contact
	}

	return nil
}

func (m *MockContactService) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
	m.UpsertContactCalled = true
	m.LastContactUpserted = contact
	if m.ErrToReturn != nil {
		return false, m.ErrToReturn
	}

	// Check if contact exists
	existingContact, exists := m.Contacts[contact.Email]
	isNew := !exists

	now := time.Now()
	if isNew {
		// Set timestamps for new contact
		if contact.CreatedAt.IsZero() {
			contact.CreatedAt = now
		}
		contact.UpdatedAt = now
		m.Contacts[contact.Email] = contact
	} else {
		// Update existing contact fields
		if contact.ExternalID != nil {
			existingContact.ExternalID = contact.ExternalID
		}
		if contact.Timezone != nil {
			existingContact.Timezone = contact.Timezone
		}
		if contact.Language != nil {
			existingContact.Language = contact.Language
		}
		if contact.FirstName != nil {
			existingContact.FirstName = contact.FirstName
		}
		if contact.LastName != nil {
			existingContact.LastName = contact.LastName
		}
		if contact.Phone != nil {
			existingContact.Phone = contact.Phone
		}
		if contact.AddressLine1 != nil {
			existingContact.AddressLine1 = contact.AddressLine1
		}
		if contact.AddressLine2 != nil {
			existingContact.AddressLine2 = contact.AddressLine2
		}
		if contact.Country != nil {
			existingContact.Country = contact.Country
		}
		if contact.Postcode != nil {
			existingContact.Postcode = contact.Postcode
		}
		if contact.State != nil {
			existingContact.State = contact.State
		}
		if contact.JobTitle != nil {
			existingContact.JobTitle = contact.JobTitle
		}
		if contact.LifetimeValue != nil {
			existingContact.LifetimeValue = contact.LifetimeValue
		}
		if contact.OrdersCount != nil {
			existingContact.OrdersCount = contact.OrdersCount
		}
		if contact.LastOrderAt != nil {
			existingContact.LastOrderAt = contact.LastOrderAt
		}
		if contact.CustomString1 != nil {
			existingContact.CustomString1 = contact.CustomString1
		}
		if contact.CustomString2 != nil {
			existingContact.CustomString2 = contact.CustomString2
		}
		if contact.CustomString3 != nil {
			existingContact.CustomString3 = contact.CustomString3
		}
		if contact.CustomString4 != nil {
			existingContact.CustomString4 = contact.CustomString4
		}
		if contact.CustomString5 != nil {
			existingContact.CustomString5 = contact.CustomString5
		}
		if contact.CustomNumber1 != nil {
			existingContact.CustomNumber1 = contact.CustomNumber1
		}
		if contact.CustomNumber2 != nil {
			existingContact.CustomNumber2 = contact.CustomNumber2
		}
		if contact.CustomNumber3 != nil {
			existingContact.CustomNumber3 = contact.CustomNumber3
		}
		if contact.CustomNumber4 != nil {
			existingContact.CustomNumber4 = contact.CustomNumber4
		}
		if contact.CustomNumber5 != nil {
			existingContact.CustomNumber5 = contact.CustomNumber5
		}
		if contact.CustomDatetime1 != nil {
			existingContact.CustomDatetime1 = contact.CustomDatetime1
		}
		if contact.CustomDatetime2 != nil {
			existingContact.CustomDatetime2 = contact.CustomDatetime2
		}
		if contact.CustomDatetime3 != nil {
			existingContact.CustomDatetime3 = contact.CustomDatetime3
		}
		if contact.CustomDatetime4 != nil {
			existingContact.CustomDatetime4 = contact.CustomDatetime4
		}
		if contact.CustomDatetime5 != nil {
			existingContact.CustomDatetime5 = contact.CustomDatetime5
		}
		if contact.CustomJSON1 != nil {
			existingContact.CustomJSON1 = contact.CustomJSON1
		}
		if contact.CustomJSON2 != nil {
			existingContact.CustomJSON2 = contact.CustomJSON2
		}
		if contact.CustomJSON3 != nil {
			existingContact.CustomJSON3 = contact.CustomJSON3
		}
		if contact.CustomJSON4 != nil {
			existingContact.CustomJSON4 = contact.CustomJSON4
		}
		if contact.CustomJSON5 != nil {
			existingContact.CustomJSON5 = contact.CustomJSON5
		}

		existingContact.UpdatedAt = now
		m.Contacts[contact.Email] = existingContact
	}

	// Return the configured value if set, otherwise return the actual isNew value
	if m.UpsertIsNewToReturn {
		return true, nil
	}
	return isNew, nil
}
