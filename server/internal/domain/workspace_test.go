package domain

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		workspace Workspace
		expectErr bool
	}{
		{
			name: "valid workspace",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name: "missing ID",
			workspace: Workspace{
				ID:   "",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid ID with special characters",
			workspace: Workspace{
				ID:   "test-123", // Contains hyphen
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "missing name",
			workspace: Workspace{
				ID:   "test123",
				Name: "",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid timezone",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "InvalidTimezone",
				},
			},
			expectErr: true,
		},
		{
			name: "missing timezone",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid website URL",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "not-a-url",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid logo URL",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "not-a-url",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.workspace.Validate()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserWorkspace_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		userWorkspace UserWorkspace
		expectErr     bool
	}{
		{
			name: "valid owner",
			userWorkspace: UserWorkspace{
				UserID:      "user123",
				WorkspaceID: "workspace123",
				Role:        "owner",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectErr: false,
		},
		{
			name: "valid member",
			userWorkspace: UserWorkspace{
				UserID:      "user123",
				WorkspaceID: "workspace123",
				Role:        "member",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectErr: false,
		},
		{
			name: "invalid role",
			userWorkspace: UserWorkspace{
				UserID:      "user123",
				WorkspaceID: "workspace123",
				Role:        "admin", // Invalid role
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectErr: true,
		},
		{
			name: "missing role",
			userWorkspace: UserWorkspace{
				UserID:      "user123",
				WorkspaceID: "workspace123",
				Role:        "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.userWorkspace.Validate()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock scanner for ScanWorkspace tests
type mockScanner struct {
	values []interface{}
	err    error
}

func (m *mockScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			*v = m.values[i].(string)
		case *[]byte:
			*v = m.values[i].([]byte)
		case *time.Time:
			*v = m.values[i].(time.Time)
		}
	}

	return nil
}

func TestScanWorkspace(t *testing.T) {
	now := time.Now()
	settingsJSON, _ := json.Marshal(WorkspaceSettings{
		WebsiteURL: "https://example.com",
		LogoURL:    "https://example.com/logo.png",
		Timezone:   "UTC",
	})

	t.Run("successful scan", func(t *testing.T) {
		scanner := &mockScanner{
			values: []interface{}{
				"workspace123",
				"Test Workspace",
				settingsJSON,
				now,
				now,
			},
		}

		workspace, err := ScanWorkspace(scanner)
		require.NoError(t, err)
		assert.Equal(t, "workspace123", workspace.ID)
		assert.Equal(t, "Test Workspace", workspace.Name)
		assert.Equal(t, "https://example.com", workspace.Settings.WebsiteURL)
		assert.Equal(t, "https://example.com/logo.png", workspace.Settings.LogoURL)
		assert.Equal(t, "UTC", workspace.Settings.Timezone)
		assert.Equal(t, now, workspace.CreatedAt)
		assert.Equal(t, now, workspace.UpdatedAt)
	})

	t.Run("scan error", func(t *testing.T) {
		scanner := &mockScanner{
			err: sql.ErrNoRows,
		}

		workspace, err := ScanWorkspace(scanner)
		assert.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("invalid settings JSON", func(t *testing.T) {
		scanner := &mockScanner{
			values: []interface{}{
				"workspace123",
				"Test Workspace",
				[]byte("invalid json"),
				now,
				now,
			},
		}

		workspace, err := ScanWorkspace(scanner)
		assert.Error(t, err)
		assert.Nil(t, workspace)
	})
}

func TestErrUnauthorized_Error(t *testing.T) {
	err := &ErrUnauthorized{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}
