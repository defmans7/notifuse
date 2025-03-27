package domain_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
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
		{
			name: "valid contact with all custom fields",
			contact: domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "custom1", IsNull: false},
				CustomString2:   &domain.NullableString{String: "custom2", IsNull: false},
				CustomString3:   &domain.NullableString{String: "custom3", IsNull: false},
				CustomString4:   &domain.NullableString{String: "custom4", IsNull: false},
				CustomString5:   &domain.NullableString{String: "custom5", IsNull: false},
				CustomNumber1:   &domain.NullableFloat64{Float64: 1.0, IsNull: false},
				CustomNumber2:   &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				CustomNumber3:   &domain.NullableFloat64{Float64: 3.0, IsNull: false},
				CustomNumber4:   &domain.NullableFloat64{Float64: 4.0, IsNull: false},
				CustomNumber5:   &domain.NullableFloat64{Float64: 5.0, IsNull: false},
				CustomDatetime1: &domain.NullableTime{Time: time.Now(), IsNull: false},
				CustomDatetime2: &domain.NullableTime{Time: time.Now(), IsNull: false},
				CustomDatetime3: &domain.NullableTime{Time: time.Now(), IsNull: false},
				CustomDatetime4: &domain.NullableTime{Time: time.Now(), IsNull: false},
				CustomDatetime5: &domain.NullableTime{Time: time.Now(), IsNull: false},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
				CustomJSON2:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
				CustomJSON3:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
				CustomJSON4:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
				CustomJSON5:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
			},
			wantErr: false,
		},
		{
			name: "valid contact with commerce fields",
			contact: domain.Contact{
				Email:         "test@example.com",
				LifetimeValue: &domain.NullableFloat64{Float64: 100.0, IsNull: false},
				OrdersCount:   &domain.NullableFloat64{Float64: 5.0, IsNull: false},
				LastOrderAt:   &domain.NullableTime{Time: time.Now(), IsNull: false},
			},
			wantErr: false,
		},
		{
			name: "valid contact with address fields",
			contact: domain.Contact{
				Email:        "test@example.com",
				AddressLine1: &domain.NullableString{String: "123 Main St", IsNull: false},
				AddressLine2: &domain.NullableString{String: "Apt 4B", IsNull: false},
				Country:      &domain.NullableString{String: "USA", IsNull: false},
				Postcode:     &domain.NullableString{String: "12345", IsNull: false},
				State:        &domain.NullableString{String: "CA", IsNull: false},
			},
			wantErr: false,
		},
		{
			name: "valid contact with contact info fields",
			contact: domain.Contact{
				Email:     "test@example.com",
				Phone:     &domain.NullableString{String: "+1234567890", IsNull: false},
				FirstName: &domain.NullableString{String: "John", IsNull: false},
				LastName:  &domain.NullableString{String: "Doe", IsNull: false},
				JobTitle:  &domain.NullableString{String: "Developer", IsNull: false},
			},
			wantErr: false,
		},
		{
			name: "valid contact with locale fields",
			contact: domain.Contact{
				Email:    "test@example.com",
				Timezone: &domain.NullableString{String: "America/New_York", IsNull: false},
				Language: &domain.NullableString{String: "en-US", IsNull: false},
			},
			wantErr: false,
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

	// Test scanning with null values
	t.Run("should handle null values", func(t *testing.T) {
		scanner := &contactMockScanner{
			data: []interface{}{
				"test@example.com",                            // Email
				sql.NullString{String: "", Valid: false},      // ExternalID
				sql.NullString{String: "", Valid: false},      // Timezone
				sql.NullString{String: "", Valid: false},      // Language
				sql.NullString{String: "", Valid: false},      // FirstName
				sql.NullString{String: "", Valid: false},      // LastName
				sql.NullString{String: "", Valid: false},      // Phone
				sql.NullString{String: "", Valid: false},      // AddressLine1
				sql.NullString{String: "", Valid: false},      // AddressLine2
				sql.NullString{String: "", Valid: false},      // Country
				sql.NullString{String: "", Valid: false},      // Postcode
				sql.NullString{String: "", Valid: false},      // State
				sql.NullString{String: "", Valid: false},      // JobTitle
				sql.NullFloat64{Float64: 0, Valid: false},     // LifetimeValue
				sql.NullFloat64{Float64: 0, Valid: false},     // OrdersCount
				sql.NullTime{Time: time.Time{}, Valid: false}, // LastOrderAt
				sql.NullString{String: "", Valid: false},      // CustomString1
				sql.NullString{String: "", Valid: false},      // CustomString2
				sql.NullString{String: "", Valid: false},      // CustomString3
				sql.NullString{String: "", Valid: false},      // CustomString4
				sql.NullString{String: "", Valid: false},      // CustomString5
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber1
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber2
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber3
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber4
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber5
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime1
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime2
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime3
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime4
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime5
				[]byte("null"), // CustomJSON1
				[]byte("null"), // CustomJSON2
				[]byte("null"), // CustomJSON3
				[]byte("null"), // CustomJSON4
				[]byte("null"), // CustomJSON5
				time.Now(),     // CreatedAt
				time.Now(),     // UpdatedAt
			},
		}

		contact, err := domain.ScanContact(scanner)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", contact.Email)
		assert.True(t, contact.ExternalID.IsNull)
		assert.True(t, contact.Timezone.IsNull)
		assert.True(t, contact.Language.IsNull)
		assert.True(t, contact.FirstName.IsNull)
		assert.True(t, contact.LastName.IsNull)
		assert.True(t, contact.Phone.IsNull)
		assert.True(t, contact.AddressLine1.IsNull)
		assert.True(t, contact.AddressLine2.IsNull)
		assert.True(t, contact.Country.IsNull)
		assert.True(t, contact.Postcode.IsNull)
		assert.True(t, contact.State.IsNull)
		assert.True(t, contact.JobTitle.IsNull)
		assert.True(t, contact.LifetimeValue.IsNull)
		assert.True(t, contact.OrdersCount.IsNull)
		assert.True(t, contact.LastOrderAt.IsNull)
		assert.True(t, contact.CustomString1.IsNull)
		assert.True(t, contact.CustomNumber1.IsNull)
		assert.True(t, contact.CustomDatetime1.IsNull)
		assert.True(t, contact.CustomJSON1.IsNull)
		assert.True(t, contact.CustomJSON2.IsNull)
		assert.True(t, contact.CustomJSON3.IsNull)
		assert.True(t, contact.CustomJSON4.IsNull)
		assert.True(t, contact.CustomJSON5.IsNull)
	})

	// Test scanning with invalid JSON data
	t.Run("should handle invalid JSON data", func(t *testing.T) {
		scanner := &contactMockScanner{
			data: []interface{}{
				"test@example.com",                            // Email
				sql.NullString{String: "", Valid: false},      // ExternalID
				sql.NullString{String: "", Valid: false},      // Timezone
				sql.NullString{String: "", Valid: false},      // Language
				sql.NullString{String: "", Valid: false},      // FirstName
				sql.NullString{String: "", Valid: false},      // LastName
				sql.NullString{String: "", Valid: false},      // Phone
				sql.NullString{String: "", Valid: false},      // AddressLine1
				sql.NullString{String: "", Valid: false},      // AddressLine2
				sql.NullString{String: "", Valid: false},      // Country
				sql.NullString{String: "", Valid: false},      // Postcode
				sql.NullString{String: "", Valid: false},      // State
				sql.NullString{String: "", Valid: false},      // JobTitle
				sql.NullFloat64{Float64: 0, Valid: false},     // LifetimeValue
				sql.NullFloat64{Float64: 0, Valid: false},     // OrdersCount
				sql.NullTime{Time: time.Time{}, Valid: false}, // LastOrderAt
				sql.NullString{String: "", Valid: false},      // CustomString1
				sql.NullString{String: "", Valid: false},      // CustomString2
				sql.NullString{String: "", Valid: false},      // CustomString3
				sql.NullString{String: "", Valid: false},      // CustomString4
				sql.NullString{String: "", Valid: false},      // CustomString5
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber1
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber2
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber3
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber4
				sql.NullFloat64{Float64: 0, Valid: false},     // CustomNumber5
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime1
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime2
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime3
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime4
				sql.NullTime{Time: time.Time{}, Valid: false}, // CustomDatetime5
				[]byte(`{invalid json}`),                      // CustomJSON1
				[]byte("null"),                                // CustomJSON2
				[]byte("null"),                                // CustomJSON3
				[]byte("null"),                                // CustomJSON4
				[]byte("null"),                                // CustomJSON5
				time.Now(),                                    // CreatedAt
				time.Now(),                                    // UpdatedAt
			},
		}

		contact, err := domain.ScanContact(scanner)
		assert.NoError(t, err)
		assert.True(t, contact.CustomJSON1.IsNull)
		assert.True(t, contact.CustomJSON2.IsNull)
		assert.True(t, contact.CustomJSON3.IsNull)
		assert.True(t, contact.CustomJSON4.IsNull)
		assert.True(t, contact.CustomJSON5.IsNull)
	})
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
		{
			name: "Merge with all custom fields",
			base: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "old1", IsNull: false},
				CustomString2:   &domain.NullableString{String: "old2", IsNull: false},
				CustomString3:   &domain.NullableString{String: "old3", IsNull: false},
				CustomString4:   &domain.NullableString{String: "old4", IsNull: false},
				CustomString5:   &domain.NullableString{String: "old5", IsNull: false},
				CustomNumber1:   &domain.NullableFloat64{Float64: 1.0, IsNull: false},
				CustomNumber2:   &domain.NullableFloat64{Float64: 2.0, IsNull: false},
				CustomNumber3:   &domain.NullableFloat64{Float64: 3.0, IsNull: false},
				CustomNumber4:   &domain.NullableFloat64{Float64: 4.0, IsNull: false},
				CustomNumber5:   &domain.NullableFloat64{Float64: 5.0, IsNull: false},
				CustomDatetime1: &domain.NullableTime{Time: now, IsNull: false},
				CustomDatetime2: &domain.NullableTime{Time: now, IsNull: false},
				CustomDatetime3: &domain.NullableTime{Time: now, IsNull: false},
				CustomDatetime4: &domain.NullableTime{Time: now, IsNull: false},
				CustomDatetime5: &domain.NullableTime{Time: now, IsNull: false},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
				CustomJSON2:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
				CustomJSON3:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
				CustomJSON4:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
				CustomJSON5:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
			},
			other: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "new1", IsNull: false},
				CustomString2:   &domain.NullableString{String: "new2", IsNull: false},
				CustomString3:   &domain.NullableString{String: "new3", IsNull: false},
				CustomString4:   &domain.NullableString{String: "new4", IsNull: false},
				CustomString5:   &domain.NullableString{String: "new5", IsNull: false},
				CustomNumber1:   &domain.NullableFloat64{Float64: 10.0, IsNull: false},
				CustomNumber2:   &domain.NullableFloat64{Float64: 20.0, IsNull: false},
				CustomNumber3:   &domain.NullableFloat64{Float64: 30.0, IsNull: false},
				CustomNumber4:   &domain.NullableFloat64{Float64: 40.0, IsNull: false},
				CustomNumber5:   &domain.NullableFloat64{Float64: 50.0, IsNull: false},
				CustomDatetime1: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime2: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime3: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime4: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime5: &domain.NullableTime{Time: later, IsNull: false},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON2:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON3:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON4:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON5:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
			},
			expected: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "new1", IsNull: false},
				CustomString2:   &domain.NullableString{String: "new2", IsNull: false},
				CustomString3:   &domain.NullableString{String: "new3", IsNull: false},
				CustomString4:   &domain.NullableString{String: "new4", IsNull: false},
				CustomString5:   &domain.NullableString{String: "new5", IsNull: false},
				CustomNumber1:   &domain.NullableFloat64{Float64: 10.0, IsNull: false},
				CustomNumber2:   &domain.NullableFloat64{Float64: 20.0, IsNull: false},
				CustomNumber3:   &domain.NullableFloat64{Float64: 30.0, IsNull: false},
				CustomNumber4:   &domain.NullableFloat64{Float64: 40.0, IsNull: false},
				CustomNumber5:   &domain.NullableFloat64{Float64: 50.0, IsNull: false},
				CustomDatetime1: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime2: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime3: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime4: &domain.NullableTime{Time: later, IsNull: false},
				CustomDatetime5: &domain.NullableTime{Time: later, IsNull: false},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON2:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON3:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON4:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
				CustomJSON5:     &domain.NullableJSON{Data: map[string]interface{}{"new": "value"}, IsNull: false},
			},
		},
		{
			name: "Merge with null custom fields",
			base: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "old1", IsNull: false},
				CustomNumber1:   &domain.NullableFloat64{Float64: 1.0, IsNull: false},
				CustomDatetime1: &domain.NullableTime{Time: now, IsNull: false},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"old": "value"}, IsNull: false},
			},
			other: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "", IsNull: true},
				CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
				CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
				CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
			},
			expected: &domain.Contact{
				Email:           "test@example.com",
				CustomString1:   &domain.NullableString{String: "", IsNull: true},
				CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
				CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
				CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
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

			// Compare all custom fields
			compareCustomFields(t, tt.base, tt.expected)
		})
	}
}

func compareCustomFields(t *testing.T, base, expected *domain.Contact) {
	// Compare CustomString fields
	for i := 1; i <= 5; i++ {
		field := fmt.Sprintf("CustomString%d", i)
		baseField := reflect.ValueOf(base).Elem().FieldByName(field)
		expectedField := reflect.ValueOf(expected).Elem().FieldByName(field)

		if !expectedField.IsNil() {
			if baseField.IsNil() {
				t.Errorf("%s is nil, want non-nil", field)
			} else {
				baseValue := baseField.Interface().(*domain.NullableString)
				expectedValue := expectedField.Interface().(*domain.NullableString)
				if baseValue.String != expectedValue.String || baseValue.IsNull != expectedValue.IsNull {
					t.Errorf("%s = %+v, want %+v", field, baseValue, expectedValue)
				}
			}
		}
	}

	// Compare CustomNumber fields
	for i := 1; i <= 5; i++ {
		field := fmt.Sprintf("CustomNumber%d", i)
		baseField := reflect.ValueOf(base).Elem().FieldByName(field)
		expectedField := reflect.ValueOf(expected).Elem().FieldByName(field)

		if !expectedField.IsNil() {
			if baseField.IsNil() {
				t.Errorf("%s is nil, want non-nil", field)
			} else {
				baseValue := baseField.Interface().(*domain.NullableFloat64)
				expectedValue := expectedField.Interface().(*domain.NullableFloat64)
				if baseValue.Float64 != expectedValue.Float64 || baseValue.IsNull != expectedValue.IsNull {
					t.Errorf("%s = %+v, want %+v", field, baseValue, expectedValue)
				}
			}
		}
	}

	// Compare CustomDatetime fields
	for i := 1; i <= 5; i++ {
		field := fmt.Sprintf("CustomDatetime%d", i)
		baseField := reflect.ValueOf(base).Elem().FieldByName(field)
		expectedField := reflect.ValueOf(expected).Elem().FieldByName(field)

		if !expectedField.IsNil() {
			if baseField.IsNil() {
				t.Errorf("%s is nil, want non-nil", field)
			} else {
				baseValue := baseField.Interface().(*domain.NullableTime)
				expectedValue := expectedField.Interface().(*domain.NullableTime)
				if !baseValue.Time.Equal(expectedValue.Time) || baseValue.IsNull != expectedValue.IsNull {
					t.Errorf("%s = %+v, want %+v", field, baseValue, expectedValue)
				}
			}
		}
	}

	// Compare CustomJSON fields
	for i := 1; i <= 5; i++ {
		field := fmt.Sprintf("CustomJSON%d", i)
		baseField := reflect.ValueOf(base).Elem().FieldByName(field)
		expectedField := reflect.ValueOf(expected).Elem().FieldByName(field)

		if !expectedField.IsNil() {
			if baseField.IsNil() {
				t.Errorf("%s is nil, want non-nil", field)
			} else {
				baseValue := baseField.Interface().(*domain.NullableJSON)
				expectedValue := expectedField.Interface().(*domain.NullableJSON)
				if !reflect.DeepEqual(baseValue.Data, expectedValue.Data) || baseValue.IsNull != expectedValue.IsNull {
					t.Errorf("%s = %+v, want %+v", field, baseValue, expectedValue)
				}
			}
		}
	}
}

func TestFromJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	validJSON := `{
		"email": "test@example.com",
		"external_id": "ext123",
		"timezone": "Europe/Paris",
		"language": "en",
		"first_name": "John",
		"last_name": "Doe",
		"phone": "1234567890",
		"address_line_1": "123 Main St",
		"address_line_2": "Apt 4B",
		"country": "US",
		"postcode": "12345",
		"state": "NY",
		"job_title": "Engineer",
		"lifetime_value": 1000.50,
		"orders_count": 5,
		"last_order_at": "` + now.Format(time.RFC3339) + `",
		"custom_string_1": "custom1",
		"custom_string_2": null,
		"custom_number_1": 42.5,
		"custom_number_2": null,
		"custom_datetime_1": "` + now.Format(time.RFC3339) + `",
		"custom_datetime_2": null,
		"custom_json_1": {"key": "value"},
		"custom_json_2": null
	}`

	tests := []struct {
		name    string
		input   interface{}
		want    *domain.Contact
		wantErr bool
	}{
		{
			name:  "valid JSON as []byte",
			input: []byte(validJSON),
			want: &domain.Contact{
				Email:           "test@example.com",
				ExternalID:      &domain.NullableString{String: "ext123", IsNull: false},
				Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
				Language:        &domain.NullableString{String: "en", IsNull: false},
				FirstName:       &domain.NullableString{String: "John", IsNull: false},
				LastName:        &domain.NullableString{String: "Doe", IsNull: false},
				Phone:           &domain.NullableString{String: "1234567890", IsNull: false},
				AddressLine1:    &domain.NullableString{String: "123 Main St", IsNull: false},
				AddressLine2:    &domain.NullableString{String: "Apt 4B", IsNull: false},
				Country:         &domain.NullableString{String: "US", IsNull: false},
				Postcode:        &domain.NullableString{String: "12345", IsNull: false},
				State:           &domain.NullableString{String: "NY", IsNull: false},
				JobTitle:        &domain.NullableString{String: "Engineer", IsNull: false},
				LifetimeValue:   &domain.NullableFloat64{Float64: 1000.50, IsNull: false},
				OrdersCount:     &domain.NullableFloat64{Float64: 5, IsNull: false},
				LastOrderAt:     &domain.NullableTime{Time: now, IsNull: false},
				CustomString1:   &domain.NullableString{String: "custom1", IsNull: false},
				CustomString2:   &domain.NullableString{String: "", IsNull: true},
				CustomNumber1:   &domain.NullableFloat64{Float64: 42.5, IsNull: false},
				CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
				CustomDatetime1: &domain.NullableTime{Time: now, IsNull: false},
				CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
				CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"key": "value"}, IsNull: false},
				CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			},
			wantErr: false,
		},
		{
			name:  "valid JSON as string",
			input: validJSON,
			want: &domain.Contact{
				Email:      "test@example.com",
				ExternalID: &domain.NullableString{String: "ext123", IsNull: false},
				// ... other fields same as above ...
			},
			wantErr: false,
		},
		{
			name:    "invalid input type",
			input:   42,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing required email",
			input:   `{"external_id": "ext123"}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid email format",
			input:   `{"email": "invalid-email"}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON format for nullable string",
			input: `{
				"email": "test@example.com",
				"external_id": 123
			}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON format for nullable float",
			input: `{
				"email": "test@example.com",
				"lifetime_value": "not-a-number"
			}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON format for nullable time",
			input: `{
				"email": "test@example.com",
				"last_order_at": "invalid-time"
			}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON format for custom JSON",
			input: `{
				"email": "test@example.com",
				"custom_json_1": "not-a-json-object"
			}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "complex custom JSON fields",
			input: `{
				"email": "test@example.com",
				"custom_json_1": {
					"nested": {
						"array": [1, 2, 3],
						"object": {"key": "value"}
					}
				},
				"custom_json_2": [
					{"id": 1, "name": "item1"},
					{"id": 2, "name": "item2"}
				]
			}`,
			want: &domain.Contact{
				Email: "test@example.com",
				CustomJSON1: &domain.NullableJSON{
					Data: map[string]interface{}{
						"nested": map[string]interface{}{
							"array":  []interface{}{float64(1), float64(2), float64(3)},
							"object": map[string]interface{}{"key": "value"},
						},
					},
					IsNull: false,
				},
				CustomJSON2: &domain.NullableJSON{
					Data: []interface{}{
						map[string]interface{}{"id": float64(1), "name": "item1"},
						map[string]interface{}{"id": float64(2), "name": "item2"},
					},
					IsNull: false,
				},
			},
			wantErr: false,
		},
		{
			name: "custom JSON fields with null values",
			input: `{
				"email": "test@example.com",
				"custom_json_1": null,
				"custom_json_2": null,
				"custom_json_3": null,
				"custom_json_4": null,
				"custom_json_5": null
			}`,
			want: &domain.Contact{
				Email:       "test@example.com",
				CustomJSON1: &domain.NullableJSON{Data: nil, IsNull: true},
				CustomJSON2: &domain.NullableJSON{Data: nil, IsNull: true},
				CustomJSON3: &domain.NullableJSON{Data: nil, IsNull: true},
				CustomJSON4: &domain.NullableJSON{Data: nil, IsNull: true},
				CustomJSON5: &domain.NullableJSON{Data: nil, IsNull: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.FromJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Compare specific fields that we want to verify
			if tt.want != nil {
				assert.Equal(t, tt.want.Email, got.Email)

				if tt.want.ExternalID != nil {
					assert.Equal(t, tt.want.ExternalID.String, got.ExternalID.String)
					assert.Equal(t, tt.want.ExternalID.IsNull, got.ExternalID.IsNull)
				}

				if tt.want.CustomJSON1 != nil {
					assert.Equal(t, tt.want.CustomJSON1.IsNull, got.CustomJSON1.IsNull)
					assert.Equal(t, tt.want.CustomJSON1.Data, got.CustomJSON1.Data)
				}

				if tt.want.CustomJSON2 != nil {
					assert.Equal(t, tt.want.CustomJSON2.IsNull, got.CustomJSON2.IsNull)
					assert.Equal(t, tt.want.CustomJSON2.Data, got.CustomJSON2.Data)
				}

				if tt.want.LastOrderAt != nil {
					assert.Equal(t, tt.want.LastOrderAt.Time.Unix(), got.LastOrderAt.Time.Unix())
					assert.Equal(t, tt.want.LastOrderAt.IsNull, got.LastOrderAt.IsNull)
				}
			}
		})
	}
}
