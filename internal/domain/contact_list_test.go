package domain_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestContactList_Validate(t *testing.T) {
	tests := []struct {
		name        string
		contactList domain.ContactList
		wantErr     bool
	}{
		{
			name: "valid contact list",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				ListID:    "list123",
				Status:    domain.ContactListStatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid contact list with pending status",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				ListID:    "list123",
				Status:    domain.ContactListStatusPending,
			},
			wantErr: false,
		},
		{
			name: "missing contact ID",
			contactList: domain.ContactList{
				ListID: "list123",
				Status: domain.ContactListStatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid contact ID format",
			contactList: domain.ContactList{
				ContactID: "not-a-uuid",
				ListID:    "list123",
				Status:    domain.ContactListStatusActive,
			},
			wantErr: true,
		},
		{
			name: "missing list ID",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				Status:    domain.ContactListStatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid list ID format",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				ListID:    "invalid@list&id",
				Status:    domain.ContactListStatusActive,
			},
			wantErr: true,
		},
		{
			name: "missing status",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				ListID:    "list123",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			contactList: domain.ContactList{
				ContactID: "123e4567-e89b-12d3-a456-426614174000",
				ListID:    "list123",
				Status:    "invalid-status",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contactList.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScanContactList(t *testing.T) {
	now := time.Now()

	// Test cases for different status values
	statuses := []string{
		string(domain.ContactListStatusActive),
		string(domain.ContactListStatusPending),
		string(domain.ContactListStatusUnsubscribed),
		string(domain.ContactListStatusBounced),
		string(domain.ContactListStatusComplained),
	}

	for _, status := range statuses {
		t.Run("scan with "+status+" status", func(t *testing.T) {
			// Create mock scanner
			scanner := &contactListMockScanner{
				data: []interface{}{
					"123e4567-e89b-12d3-a456-426614174000", // ContactID
					"list123",                              // ListID
					status,                                 // Status
					now,                                    // CreatedAt
					now,                                    // UpdatedAt
				},
			}

			// Test successful scan
			contactList, err := domain.ScanContactList(scanner)
			assert.NoError(t, err)
			assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", contactList.ContactID)
			assert.Equal(t, "list123", contactList.ListID)
			assert.Equal(t, domain.ContactListStatus(status), contactList.Status)
			assert.Equal(t, now, contactList.CreatedAt)
			assert.Equal(t, now, contactList.UpdatedAt)
		})
	}

	// Test scan error
	t.Run("scan error", func(t *testing.T) {
		scanner := &contactListMockScanner{
			err: sql.ErrNoRows,
		}
		_, err := domain.ScanContactList(scanner)
		assert.Error(t, err)
	})
}

// ContactListStatus constants test
func TestContactListStatusConstants(t *testing.T) {
	assert.Equal(t, domain.ContactListStatus("active"), domain.ContactListStatusActive)
	assert.Equal(t, domain.ContactListStatus("pending"), domain.ContactListStatusPending)
	assert.Equal(t, domain.ContactListStatus("unsubscribed"), domain.ContactListStatusUnsubscribed)
	assert.Equal(t, domain.ContactListStatus("bounced"), domain.ContactListStatusBounced)
	assert.Equal(t, domain.ContactListStatus("complained"), domain.ContactListStatusComplained)
}

// Mock scanner for testing
type contactListMockScanner struct {
	data []interface{}
	err  error
}

func (m *contactListMockScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			if s, ok := m.data[i].(string); ok {
				*v = s
			}
		case *time.Time:
			if t, ok := m.data[i].(time.Time); ok {
				*v = t
			}
		}
	}

	return nil
}
