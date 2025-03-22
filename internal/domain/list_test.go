package domain_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestList_Validate(t *testing.T) {
	tests := []struct {
		name    string
		list    domain.List
		wantErr bool
	}{
		{
			name: "valid list",
			list: domain.List{
				ID:            "list123",
				Name:          "My List",
				Type:          "public",
				IsDoubleOptin: true,
				Description:   "This is a description",
			},
			wantErr: false,
		},
		{
			name: "valid list without description",
			list: domain.List{
				ID:            "list123",
				Name:          "My List",
				Type:          "private",
				IsDoubleOptin: false,
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			list: domain.List{
				ID:            "",
				Name:          "My List",
				Type:          "public",
				IsDoubleOptin: true,
			},
			wantErr: true,
		},
		{
			name: "invalid name",
			list: domain.List{
				ID:            "list123",
				Name:          "",
				Type:          "public",
				IsDoubleOptin: true,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			list: domain.List{
				ID:            "list123",
				Name:          "My List",
				Type:          "invalid",
				IsDoubleOptin: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.list.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScanList(t *testing.T) {
	// Create mock scanner
	scanner := &mockScanner{
		data: []interface{}{
			"list123",        // ID
			"My List",        // Name
			"public",         // Type
			true,             // IsDoubleOptin
			"This is a list", // Description
			time.Now(),       // CreatedAt
			time.Now(),       // UpdatedAt
		},
	}

	// Test successful scan
	list, err := domain.ScanList(scanner)
	assert.NoError(t, err)
	assert.Equal(t, "list123", list.ID)
	assert.Equal(t, "My List", list.Name)
	assert.Equal(t, "public", list.Type)
	assert.Equal(t, true, list.IsDoubleOptin)
	assert.Equal(t, "This is a list", list.Description)

	// Test scan error
	scanner.err = sql.ErrNoRows
	_, err = domain.ScanList(scanner)
	assert.Error(t, err)
}

// Mock scanner for testing
type mockScanner struct {
	data []interface{}
	err  error
}

func (m *mockScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			*v = m.data[i].(string)
		case *bool:
			*v = m.data[i].(bool)
		case *time.Time:
			*v = m.data[i].(time.Time)
		}
	}
	return nil
}
