package domain_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestContact_Validate(t *testing.T) {
	tests := []struct {
		name    string
		contact domain.Contact
		wantErr bool
	}{
		{
			name: "valid contact with all required fields",
			contact: domain.Contact{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				ExternalID: "ext123",
				Email:      "test@example.com",
				Timezone:   "Europe/Paris",
			},
			wantErr: false,
		},
		{
			name: "valid contact with all fields",
			contact: domain.Contact{
				UUID:            "123e4567-e89b-12d3-a456-426614174000",
				ExternalID:      "ext123",
				Email:           "test@example.com",
				Timezone:        "Europe/Paris",
				FirstName:       "John",
				LastName:        "Doe",
				Phone:           "+1234567890",
				AddressLine1:    "123 Main St",
				AddressLine2:    "Apt 4B",
				Country:         "USA",
				Postcode:        "12345",
				State:           "CA",
				JobTitle:        "Developer",
				LifetimeValue:   100.50,
				OrdersCount:     5,
				LastOrderAt:     time.Now(),
				CustomString1:   "Custom 1",
				CustomNumber1:   42.0,
				CustomDatetime1: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			contact: domain.Contact{
				ExternalID: "ext123",
				Email:      "test@example.com",
				Timezone:   "Europe/Paris",
			},
			wantErr: true,
		},
		{
			name: "missing external ID",
			contact: domain.Contact{
				UUID:     "123e4567-e89b-12d3-a456-426614174000",
				Email:    "test@example.com",
				Timezone: "Europe/Paris",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			contact: domain.Contact{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				ExternalID: "ext123",
				Timezone:   "Europe/Paris",
			},
			wantErr: true,
		},
		{
			name: "missing timezone",
			contact: domain.Contact{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				ExternalID: "ext123",
				Email:      "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			contact: domain.Contact{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				ExternalID: "ext123",
				Email:      "invalid-email",
				Timezone:   "Europe/Paris",
			},
			wantErr: true,
		},
		{
			name: "invalid timezone",
			contact: domain.Contact{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				ExternalID: "ext123",
				Email:      "test@example.com",
				Timezone:   "InvalidTimezone",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScanContact(t *testing.T) {
	now := time.Now()

	// Create mock scanner
	scanner := &contactMockScanner{
		data: []interface{}{
			"123e4567-e89b-12d3-a456-426614174000", // UUID
			"ext123",                               // ExternalID
			"test@example.com",                     // Email
			"Europe/Paris",                         // Timezone
			"John",                                 // FirstName
			"Doe",                                  // LastName
			"+1234567890",                          // Phone
			"123 Main St",                          // AddressLine1
			"Apt 4B",                               // AddressLine2
			"USA",                                  // Country
			"12345",                                // Postcode
			"CA",                                   // State
			"Developer",                            // JobTitle
			100.50,                                 // LifetimeValue
			5,                                      // OrdersCount
			now,                                    // LastOrderAt
			"Custom 1",                             // CustomString1
			"Custom 2",                             // CustomString2
			"Custom 3",                             // CustomString3
			"Custom 4",                             // CustomString4
			"Custom 5",                             // CustomString5
			42.0,                                   // CustomNumber1
			43.0,                                   // CustomNumber2
			44.0,                                   // CustomNumber3
			45.0,                                   // CustomNumber4
			46.0,                                   // CustomNumber5
			now,                                    // CustomDatetime1
			now,                                    // CustomDatetime2
			now,                                    // CustomDatetime3
			now,                                    // CustomDatetime4
			now,                                    // CustomDatetime5
			now,                                    // CreatedAt
			now,                                    // UpdatedAt
		},
	}

	// Test successful scan
	contact, err := domain.ScanContact(scanner)
	assert.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", contact.UUID)
	assert.Equal(t, "ext123", contact.ExternalID)
	assert.Equal(t, "test@example.com", contact.Email)
	assert.Equal(t, "Europe/Paris", contact.Timezone)
	assert.Equal(t, "John", contact.FirstName)
	assert.Equal(t, "Doe", contact.LastName)
	assert.Equal(t, "+1234567890", contact.Phone)
	assert.Equal(t, "123 Main St", contact.AddressLine1)
	assert.Equal(t, "Apt 4B", contact.AddressLine2)
	assert.Equal(t, "USA", contact.Country)
	assert.Equal(t, "12345", contact.Postcode)
	assert.Equal(t, "CA", contact.State)
	assert.Equal(t, "Developer", contact.JobTitle)
	assert.Equal(t, 100.50, contact.LifetimeValue)
	assert.Equal(t, 5, contact.OrdersCount)
	assert.Equal(t, now, contact.LastOrderAt)
	assert.Equal(t, "Custom 1", contact.CustomString1)
	assert.Equal(t, 42.0, contact.CustomNumber1)
	assert.Equal(t, now, contact.CustomDatetime1)

	// Test scan error
	scanner.err = sql.ErrNoRows
	_, err = domain.ScanContact(scanner)
	assert.Error(t, err)
}

// Mock scanner for testing
type contactMockScanner struct {
	data []interface{}
	err  error
}

func (m *contactMockScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			if s, ok := m.data[i].(string); ok {
				*v = s
			}
		case *int:
			if n, ok := m.data[i].(int); ok {
				*v = n
			}
		case *float64:
			if f, ok := m.data[i].(float64); ok {
				*v = f
			}
		case *time.Time:
			if t, ok := m.data[i].(time.Time); ok {
				*v = t
			}
		}
	}

	return nil
}
