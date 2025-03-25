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
			name: "valid contact with required email field",
			contact: domain.Contact{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid contact with all fields including custom JSON",
			contact: domain.Contact{
				Email:      "test@example.com",
				ExternalID: "ext123",
				Timezone:   "Europe/Paris",
				FirstName:  domain.NullableString{String: "John", IsNull: false},
				LastName:   domain.NullableString{String: "Doe", IsNull: false},
				CustomJSON1: domain.NullableJSON{
					Data:  map[string]interface{}{"preferences": map[string]interface{}{"theme": "dark"}},
					Valid: true,
				},
				CustomJSON2: domain.NullableJSON{
					Data:  []interface{}{"tag1", "tag2"},
					Valid: true,
				},
				CustomJSON3: domain.NullableJSON{
					Data:  42.5,
					Valid: true,
				},
				CustomJSON4: domain.NullableJSON{
					Data:  "string value",
					Valid: true,
				},
				CustomJSON5: domain.NullableJSON{
					Data:  true,
					Valid: true,
				},
			},
			wantErr: false,
		},
		{
			name: "missing email",
			contact: domain.Contact{
				ExternalID: "ext123",
				Timezone:   "Europe/Paris",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			contact: domain.Contact{
				Email:      "invalid-email",
				ExternalID: "ext123",
				Timezone:   "Europe/Paris",
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
			"ext123",           // ExternalID
			"Europe/Paris",     // Timezone
			sql.NullString{String: "John", Valid: true},        // FirstName
			sql.NullString{String: "Doe", Valid: true},         // LastName
			sql.NullString{String: "+1234567890", Valid: true}, // Phone
			sql.NullString{String: "123 Main St", Valid: true}, // AddressLine1
			sql.NullString{String: "Apt 4B", Valid: true},      // AddressLine2
			sql.NullString{String: "USA", Valid: true},         // Country
			sql.NullString{String: "12345", Valid: true},       // Postcode
			sql.NullString{String: "CA", Valid: true},          // State
			sql.NullString{String: "Developer", Valid: true},   // JobTitle
			sql.NullFloat64{Float64: 100.50, Valid: true},      // LifetimeValue
			sql.NullFloat64{Float64: 5, Valid: true},           // OrdersCount
			sql.NullTime{Time: now, Valid: true},               // LastOrderAt
			sql.NullString{String: "Custom 1", Valid: true},    // CustomString1
			sql.NullString{String: "Custom 2", Valid: true},    // CustomString2
			sql.NullString{String: "Custom 3", Valid: true},    // CustomString3
			sql.NullString{String: "Custom 4", Valid: true},    // CustomString4
			sql.NullString{String: "Custom 5", Valid: true},    // CustomString5
			sql.NullFloat64{Float64: 42.0, Valid: true},        // CustomNumber1
			sql.NullFloat64{Float64: 43.0, Valid: true},        // CustomNumber2
			sql.NullFloat64{Float64: 44.0, Valid: true},        // CustomNumber3
			sql.NullFloat64{Float64: 45.0, Valid: true},        // CustomNumber4
			sql.NullFloat64{Float64: 46.0, Valid: true},        // CustomNumber5
			sql.NullTime{Time: now, Valid: true},               // CustomDatetime1
			sql.NullTime{Time: now, Valid: true},               // CustomDatetime2
			sql.NullTime{Time: now, Valid: true},               // CustomDatetime3
			sql.NullTime{Time: now, Valid: true},               // CustomDatetime4
			sql.NullTime{Time: now, Valid: true},               // CustomDatetime5
			jsonData1,                                          // CustomJSON1
			jsonData2,                                          // CustomJSON2
			jsonData3,                                          // CustomJSON3
			jsonData4,                                          // CustomJSON4
			jsonData5,                                          // CustomJSON5
			now,                                                // CreatedAt
			now,                                                // UpdatedAt
		},
	}

	// Test successful scan
	contact, err := domain.ScanContact(scanner)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", contact.Email)
	assert.Equal(t, "ext123", contact.ExternalID)
	assert.Equal(t, "Europe/Paris", contact.Timezone)
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
	assert.True(t, contact.CustomJSON1.Valid)
	preferences, ok := contact.CustomJSON1.Data.(map[string]interface{})
	assert.True(t, ok)
	theme, ok := preferences["preferences"].(map[string]interface{})["theme"].(string)
	assert.True(t, ok)
	assert.Equal(t, "dark", theme)

	assert.True(t, contact.CustomJSON2.Valid)
	tags, ok := contact.CustomJSON2.Data.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "tag1", tags[0])
	assert.Equal(t, "tag2", tags[1])

	assert.True(t, contact.CustomJSON3.Valid)
	assert.Equal(t, 42.5, contact.CustomJSON3.Data)

	assert.True(t, contact.CustomJSON4.Valid)
	assert.Equal(t, "string value", contact.CustomJSON4.Data)

	assert.True(t, contact.CustomJSON5.Valid)
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
