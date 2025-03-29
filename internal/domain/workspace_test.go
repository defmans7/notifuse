package domain

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/asaskevich/govalidator"
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
		{
			name: "invalid cover URL",
			workspace: Workspace{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					CoverURL:   "not-a-url",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "name too long",
			workspace: Workspace{
				ID:   "test123",
				Name: string(make([]byte, 256)), // 256 chars
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			expectErr: true,
		},
		{
			name: "ID too long",
			workspace: Workspace{
				ID:   string(make([]byte, 21)), // 21 chars
				Name: "Test Workspace",
				Settings: WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
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

func TestCreateWorkspaceRequest_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		request CreateWorkspaceRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			request: CreateWorkspaceRequest{
				ID:   "",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid ID with special characters",
			request: CreateWorkspaceRequest{
				ID:   "test-123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "missing name",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timezone",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "InvalidTimezone",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid website URL",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "not-a-url",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid logo URL",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "not-a-url",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "missing settings name",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: "Test Workspace",
				Settings: WorkspaceSettingsData{
					Name:       "",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "name too long",
			request: CreateWorkspaceRequest{
				ID:   "test123",
				Name: string(make([]byte, 33)), // 33 chars
				Settings: WorkspaceSettingsData{
					Name:       "Test Workspace",
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateWorkspaceRequest_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		request UpdateWorkspaceRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: UpdateWorkspaceRequest{
				ID:         "test123",
				Name:       "Test Workspace",
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "UTC",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			request: UpdateWorkspaceRequest{
				ID:         "",
				Name:       "Test Workspace",
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
		{
			name: "invalid ID with special characters",
			request: UpdateWorkspaceRequest{
				ID:         "test-123",
				Name:       "Test Workspace",
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			request: UpdateWorkspaceRequest{
				ID:         "test123",
				Name:       "",
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
		{
			name: "invalid timezone",
			request: UpdateWorkspaceRequest{
				ID:         "test123",
				Name:       "Test Workspace",
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "InvalidTimezone",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteWorkspaceRequest_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		request DeleteWorkspaceRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: DeleteWorkspaceRequest{
				ID: "test123",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			request: DeleteWorkspaceRequest{
				ID: "",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInviteMemberRequest_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		request InviteMemberRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: InviteMemberRequest{
				WorkspaceID: "test123",
				Email:       "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: InviteMemberRequest{
				WorkspaceID: "",
				Email:       "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			request: InviteMemberRequest{
				WorkspaceID: "test123",
				Email:       "",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			request: InviteMemberRequest{
				WorkspaceID: "test123",
				Email:       "invalid-email",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkspace_Validate_TimezoneValidatorRegistration(t *testing.T) {
	// Save the original timezone validator
	originalTimezoneValidator, exists := govalidator.TagMap["timezone"]

	// Remove the timezone validator to test registration
	delete(govalidator.TagMap, "timezone")

	workspace := Workspace{
		ID:   "test123",
		Name: "Test Workspace",
		Settings: WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC", // Use a valid timezone
		},
	}

	err := workspace.Validate()
	assert.NoError(t, err) // Should pass as the validator will be registered

	// Restore the original validator
	if exists {
		govalidator.TagMap["timezone"] = originalTimezoneValidator
	}
}

func TestCreateWorkspaceRequest_Validate_TimezoneValidatorRegistration(t *testing.T) {
	// Save the original timezone validator
	originalTimezoneValidator, exists := govalidator.TagMap["timezone"]

	// Remove the timezone validator to test registration
	delete(govalidator.TagMap, "timezone")

	request := CreateWorkspaceRequest{
		ID:   "test123",
		Name: "Test Workspace",
		Settings: WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC", // Use a valid timezone
		},
	}

	err := request.Validate()
	assert.NoError(t, err) // Should pass as the validator will be registered

	// Restore the original validator
	if exists {
		govalidator.TagMap["timezone"] = originalTimezoneValidator
	}
}

func TestWorkspace_Validate_FirstValidationFails(t *testing.T) {
	workspace := Workspace{
		ID:   "", // Invalid ID to fail first validation
		Name: "Test Workspace",
		Settings: WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	err := workspace.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workspace")
}

func TestCreateWorkspaceRequest_Validate_FirstValidationFails(t *testing.T) {
	request := CreateWorkspaceRequest{
		ID:   "", // Invalid ID to fail first validation
		Name: "Test Workspace",
		Settings: WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	err := request.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid create workspace request")
}
