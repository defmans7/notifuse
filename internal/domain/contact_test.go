package domain_test

import (
	"database/sql"
	"encoding/json"
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
			name: "valid contact with required email field only",
			contact: domain.Contact{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid contact with all optional fields",
			contact: domain.Contact{
				Email:      "test@example.com",
				ExternalID: &domain.NullableString{String: "ext123", IsNull: false},
				Timezone:   &domain.NullableString{String: "Europe/Paris", IsNull: false},
				Language:   &domain.NullableString{String: "en", IsNull: false},
				FirstName:  &domain.NullableString{String: "John", IsNull: false},
				LastName:   &domain.NullableString{String: "Doe", IsNull: false},
				CustomJSON1: &domain.NullableJSON{
					Data:   map[string]interface{}{"preferences": map[string]interface{}{"theme": "dark"}},
					IsNull: false,
				},
			},
			wantErr: false,
		},
		{
			name: "missing email",
			contact: domain.Contact{
				ExternalID: &domain.NullableString{String: "ext123", IsNull: false},
				Timezone:   &domain.NullableString{String: "Europe/Paris", IsNull: false},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			contact: domain.Contact{
				Email:      "invalid-email",
				ExternalID: &domain.NullableString{String: "ext123", IsNull: false},
				Timezone:   &domain.NullableString{String: "Europe/Paris", IsNull: false},
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

	// Create JSON test data
	jsonData1, _ := json.Marshal(map[string]interface{}{"preferences": map[string]interface{}{"theme": "dark"}})
	jsonData2, _ := json.Marshal([]interface{}{"tag1", "tag2"})
	jsonData3, _ := json.Marshal(42.5)
	jsonData4, _ := json.Marshal("string value")
	jsonData5, _ := json.Marshal(true)

	// Create mock scanner
	scanner := &contactMockScanner{
		data: []interface{}{
			"test@example.com", // Email
			sql.NullString{String: "ext123", Valid: true},       // ExternalID
			sql.NullString{String: "Europe/Paris", Valid: true}, // Timezone
			sql.NullString{String: "en-US", Valid: true},        // Language
			sql.NullString{String: "John", Valid: true},         // FirstName
			sql.NullString{String: "Doe", Valid: true},          // LastName
			sql.NullString{String: "+1234567890", Valid: true},  // Phone
			sql.NullString{String: "123 Main St", Valid: true},  // AddressLine1
			sql.NullString{String: "Apt 4B", Valid: true},       // AddressLine2
			sql.NullString{String: "USA", Valid: true},          // Country
			sql.NullString{String: "12345", Valid: true},        // Postcode
			sql.NullString{String: "CA", Valid: true},           // State
			sql.NullString{String: "Developer", Valid: true},    // JobTitle
			sql.NullFloat64{Float64: 100.50, Valid: true},       // LifetimeValue
			sql.NullFloat64{Float64: 5, Valid: true},            // OrdersCount
			sql.NullTime{Time: now, Valid: true},                // LastOrderAt
			sql.NullString{String: "Custom 1", Valid: true},     // CustomString1
			sql.NullString{String: "Custom 2", Valid: true},     // CustomString2
			sql.NullString{String: "Custom 3", Valid: true},     // CustomString3
			sql.NullString{String: "Custom 4", Valid: true},     // CustomString4
			sql.NullString{String: "Custom 5", Valid: true},     // CustomString5
			sql.NullFloat64{Float64: 42.0, Valid: true},         // CustomNumber1
			sql.NullFloat64{Float64: 43.0, Valid: true},         // CustomNumber2
			sql.NullFloat64{Float64: 44.0, Valid: true},         // CustomNumber3
			sql.NullFloat64{Float64: 45.0, Valid: true},         // CustomNumber4
			sql.NullFloat64{Float64: 46.0, Valid: true},         // CustomNumber5
			sql.NullTime{Time: now, Valid: true},                // CustomDatetime1
			sql.NullTime{Time: now, Valid: true},                // CustomDatetime2
			sql.NullTime{Time: now, Valid: true},                // CustomDatetime3
			sql.NullTime{Time: now, Valid: true},                // CustomDatetime4
			sql.NullTime{Time: now, Valid: true},                // CustomDatetime5
			jsonData1,                                           // CustomJSON1
			jsonData2,                                           // CustomJSON2
			jsonData3,                                           // CustomJSON3
			jsonData4,                                           // CustomJSON4
			jsonData5,                                           // CustomJSON5
			now,                                                 // CreatedAt
			now,                                                 // UpdatedAt
		},
	}

	// Test successful scan
	contact, err := domain.ScanContact(scanner)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", contact.Email)
	assert.Equal(t, "ext123", contact.ExternalID.String)
	assert.Equal(t, "Europe/Paris", contact.Timezone.String)
	assert.Equal(t, "en-US", contact.Language.String)
	assert.False(t, contact.Language.IsNull)
	assert.Equal(t, "John", contact.FirstName.String)
	assert.False(t, contact.FirstName.IsNull)
	assert.Equal(t, "Doe", contact.LastName.String)
	assert.False(t, contact.LastName.IsNull)
	assert.Equal(t, "+1234567890", contact.Phone.String)
	assert.False(t, contact.Phone.IsNull)
	assert.Equal(t, "123 Main St", contact.AddressLine1.String)
	assert.False(t, contact.AddressLine1.IsNull)
	assert.Equal(t, "Apt 4B", contact.AddressLine2.String)
	assert.False(t, contact.AddressLine2.IsNull)
	assert.Equal(t, "USA", contact.Country.String)
	assert.False(t, contact.Country.IsNull)
	assert.Equal(t, "12345", contact.Postcode.String)
	assert.False(t, contact.Postcode.IsNull)
	assert.Equal(t, "CA", contact.State.String)
	assert.False(t, contact.State.IsNull)
	assert.Equal(t, "Developer", contact.JobTitle.String)
	assert.False(t, contact.JobTitle.IsNull)
	assert.Equal(t, 100.50, contact.LifetimeValue.Float64)
	assert.False(t, contact.LifetimeValue.IsNull)
	assert.Equal(t, float64(5), contact.OrdersCount.Float64)
	assert.False(t, contact.OrdersCount.IsNull)
	assert.Equal(t, now, contact.LastOrderAt.Time)
	assert.False(t, contact.LastOrderAt.IsNull)
	assert.Equal(t, "Custom 1", contact.CustomString1.String)
	assert.False(t, contact.CustomString1.IsNull)
	assert.Equal(t, 42.0, contact.CustomNumber1.Float64)
	assert.False(t, contact.CustomNumber1.IsNull)
	assert.Equal(t, now, contact.CustomDatetime1.Time)
	assert.False(t, contact.CustomDatetime1.IsNull)

	// Test custom JSON fields
	assert.False(t, contact.CustomJSON1.IsNull)
	preferences, ok := contact.CustomJSON1.Data.(map[string]interface{})
	assert.True(t, ok)
	theme, ok := preferences["preferences"].(map[string]interface{})["theme"].(string)
	assert.True(t, ok)
	assert.Equal(t, "dark", theme)

	assert.False(t, contact.CustomJSON2.IsNull)
	tags, ok := contact.CustomJSON2.Data.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "tag1", tags[0])
	assert.Equal(t, "tag2", tags[1])

	assert.False(t, contact.CustomJSON3.IsNull)
	assert.Equal(t, 42.5, contact.CustomJSON3.Data)

	assert.False(t, contact.CustomJSON4.IsNull)
	assert.Equal(t, "string value", contact.CustomJSON4.Data)

	assert.False(t, contact.CustomJSON5.IsNull)
	assert.Equal(t, true, contact.CustomJSON5.Data)

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
		if i >= len(m.data) {
			continue
		}

		switch v := d.(type) {
		case *string:
			if s, ok := m.data[i].(string); ok {
				*v = s
			}
		case *sql.NullString:
			if s, ok := m.data[i].(sql.NullString); ok {
				*v = s
			}
		case *sql.NullFloat64:
			if f, ok := m.data[i].(sql.NullFloat64); ok {
				*v = f
			}
		case *sql.NullTime:
			if t, ok := m.data[i].(sql.NullTime); ok {
				*v = t
			}
		case *[]byte:
			switch data := m.data[i].(type) {
			case []byte:
				*v = data
			case string:
				*v = []byte(data)
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

func TestContact_Merge(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		base     *domain.Contact
		other    *domain.Contact
		expected *domain.Contact
	}{
		{
			name: "Merge with nil contact",
			base: &domain.Contact{
				Email:     "test@example.com",
				FirstName: &domain.NullableString{String: "Original", IsNull: false},
			},
			other: nil,
			expected: &domain.Contact{
				Email:     "test@example.com",
				FirstName: &domain.NullableString{String: "Original", IsNull: false},
			},
		},
		{
			name: "Merge basic fields",
			base: &domain.Contact{
				Email:     "old@example.com",
				FirstName: &domain.NullableString{String: "Old", IsNull: false},
				LastName:  &domain.NullableString{String: "Name", IsNull: false},
			},
			other: &domain.Contact{
				Email:     "new@example.com",
				FirstName: &domain.NullableString{String: "New", IsNull: false},
			},
			expected: &domain.Contact{
				Email:     "new@example.com",
				FirstName: &domain.NullableString{String: "New", IsNull: false},
				LastName:  &domain.NullableString{String: "Name", IsNull: false},
			},
		},
		{
			name: "Merge with null fields",
			base: &domain.Contact{
				Email:     "test@example.com",
				FirstName: &domain.NullableString{String: "Original", IsNull: false},
				LastName:  &domain.NullableString{String: "Name", IsNull: false},
			},
			other: &domain.Contact{
				Email:     "test@example.com",
				FirstName: &domain.NullableString{String: "", IsNull: true},
			},
			expected: &domain.Contact{
				Email:     "test@example.com",
				FirstName: &domain.NullableString{String: "", IsNull: true},
				LastName:  &domain.NullableString{String: "Name", IsNull: false},
			},
		},
		{
			name: "Merge timestamps",
			base: &domain.Contact{
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			other: &domain.Contact{
				Email:     "test@example.com",
				CreatedAt: later,
				UpdatedAt: later,
			},
			expected: &domain.Contact{
				Email:     "test@example.com",
				CreatedAt: later,
				UpdatedAt: later,
			},
		},
		{
			name: "Merge custom fields",
			base: &domain.Contact{
				Email:         "test@example.com",
				CustomString1: &domain.NullableString{String: "Old String", IsNull: false},
				CustomNumber1: &domain.NullableFloat64{Float64: 1.0, IsNull: false},
				CustomJSON1:   &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
			},
			other: &domain.Contact{
				Email:         "test@example.com",
				CustomString1: &domain.NullableString{String: "New String", IsNull: false},
				CustomNumber1: &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				CustomJSON1:   &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
			},
			expected: &domain.Contact{
				Email:         "test@example.com",
				CustomString1: &domain.NullableString{String: "New String", IsNull: false},
				CustomNumber1: &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				CustomJSON1:   &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
			},
		},
		{
			name: "Merge commerce fields",
			base: &domain.Contact{
				Email:         "test@example.com",
				LifetimeValue: &domain.NullableFloat64{Float64: 100.0, IsNull: false},
				OrdersCount:   &domain.NullableFloat64{Float64: 1.0, IsNull: false},
				LastOrderAt:   &domain.NullableTime{Time: now, IsNull: false},
			},
			other: &domain.Contact{
				Email:         "test@example.com",
				LifetimeValue: &domain.NullableFloat64{Float64: 200.0, IsNull: false},
				OrdersCount:   &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				LastOrderAt:   &domain.NullableTime{Time: later, IsNull: false},
			},
			expected: &domain.Contact{
				Email:         "test@example.com",
				LifetimeValue: &domain.NullableFloat64{Float64: 200.0, IsNull: false},
				OrdersCount:   &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				LastOrderAt:   &domain.NullableTime{Time: later, IsNull: false},
			},
		},
		{
			name: "Merge address fields",
			base: &domain.Contact{
				Email:        "test@example.com",
				AddressLine1: &domain.NullableString{String: "123 Old St", IsNull: false},
				AddressLine2: &domain.NullableString{String: "Apt 1", IsNull: false},
				Country:      &domain.NullableString{String: "USA", IsNull: false},
				State:        &domain.NullableString{String: "CA", IsNull: false},
				Postcode:     &domain.NullableString{String: "12345", IsNull: false},
			},
			other: &domain.Contact{
				Email:        "test@example.com",
				AddressLine1: &domain.NullableString{String: "456 New St", IsNull: false},
				Country:      &domain.NullableString{String: "Canada", IsNull: false},
			},
			expected: &domain.Contact{
				Email:        "test@example.com",
				AddressLine1: &domain.NullableString{String: "456 New St", IsNull: false},
				AddressLine2: &domain.NullableString{String: "Apt 1", IsNull: false},
				Country:      &domain.NullableString{String: "Canada", IsNull: false},
				State:        &domain.NullableString{String: "CA", IsNull: false},
				Postcode:     &domain.NullableString{String: "12345", IsNull: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.other)

			// Compare Email
			if tt.base.Email != tt.expected.Email {
				t.Errorf("Email = %v, want %v", tt.base.Email, tt.expected.Email)
			}

			// Compare FirstName if present
			if tt.expected.FirstName != nil {
				if tt.base.FirstName == nil {
					t.Error("FirstName is nil, want non-nil")
				} else if tt.base.FirstName.String != tt.expected.FirstName.String || tt.base.FirstName.IsNull != tt.expected.FirstName.IsNull {
					t.Errorf("FirstName = %+v, want %+v", tt.base.FirstName, tt.expected.FirstName)
				}
			}

			// Compare LastName if present
			if tt.expected.LastName != nil {
				if tt.base.LastName == nil {
					t.Error("LastName is nil, want non-nil")
				} else if tt.base.LastName.String != tt.expected.LastName.String || tt.base.LastName.IsNull != tt.expected.LastName.IsNull {
					t.Errorf("LastName = %+v, want %+v", tt.base.LastName, tt.expected.LastName)
				}
			}

			// Compare timestamps
			if !tt.base.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", tt.base.CreatedAt, tt.expected.CreatedAt)
			}
			if !tt.base.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("UpdatedAt = %v, want %v", tt.base.UpdatedAt, tt.expected.UpdatedAt)
			}

			// Compare custom fields if present
			if tt.expected.CustomString1 != nil {
				if tt.base.CustomString1 == nil {
					t.Error("CustomString1 is nil, want non-nil")
				} else if tt.base.CustomString1.String != tt.expected.CustomString1.String || tt.base.CustomString1.IsNull != tt.expected.CustomString1.IsNull {
					t.Errorf("CustomString1 = %+v, want %+v", tt.base.CustomString1, tt.expected.CustomString1)
				}
			}

			if tt.expected.CustomNumber1 != nil {
				if tt.base.CustomNumber1 == nil {
					t.Error("CustomNumber1 is nil, want non-nil")
				} else if tt.base.CustomNumber1.Float64 != tt.expected.CustomNumber1.Float64 || tt.base.CustomNumber1.IsNull != tt.expected.CustomNumber1.IsNull {
					t.Errorf("CustomNumber1 = %+v, want %+v", tt.base.CustomNumber1, tt.expected.CustomNumber1)
				}
			}

			if tt.expected.CustomJSON1 != nil {
				if tt.base.CustomJSON1 == nil {
					t.Error("CustomJSON1 is nil, want non-nil")
				} else if tt.base.CustomJSON1.IsNull != tt.expected.CustomJSON1.IsNull {
					t.Errorf("CustomJSON1.IsNull = %v, want %v", tt.base.CustomJSON1.IsNull, tt.expected.CustomJSON1.IsNull)
				}
			}

			// Compare commerce fields if present
			if tt.expected.LifetimeValue != nil {
				if tt.base.LifetimeValue == nil {
					t.Error("LifetimeValue is nil, want non-nil")
				} else if tt.base.LifetimeValue.Float64 != tt.expected.LifetimeValue.Float64 || tt.base.LifetimeValue.IsNull != tt.expected.LifetimeValue.IsNull {
					t.Errorf("LifetimeValue = %+v, want %+v", tt.base.LifetimeValue, tt.expected.LifetimeValue)
				}
			}

			// Compare address fields if present
			if tt.expected.AddressLine1 != nil {
				if tt.base.AddressLine1 == nil {
					t.Error("AddressLine1 is nil, want non-nil")
				} else if tt.base.AddressLine1.String != tt.expected.AddressLine1.String || tt.base.AddressLine1.IsNull != tt.expected.AddressLine1.IsNull {
					t.Errorf("AddressLine1 = %+v, want %+v", tt.base.AddressLine1, tt.expected.AddressLine1)
				}
			}
		})
	}
}
