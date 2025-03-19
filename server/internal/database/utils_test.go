package database

import (
	"testing"

	"notifuse/server/config"

	"github.com/stretchr/testify/assert"
)

func TestGetSystemDSN(t *testing.T) {
	testCases := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "standard config",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "notifuse",
			},
			expected: "postgres://postgres:password@localhost:5432/notifuse?sslmode=disable",
		},
		{
			name: "custom port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5433,
				User:     "postgres",
				Password: "password",
				DBName:   "notifuse",
			},
			expected: "postgres://postgres:password@localhost:5433/notifuse?sslmode=disable",
		},
		{
			name: "remote host",
			config: &config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5432,
				User:     "app_user",
				Password: "secure_password",
				DBName:   "notifuse_prod",
			},
			expected: "postgres://app_user:secure_password@db.example.com:5432/notifuse_prod?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetSystemDSN(tc.config)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetWorkspaceDSN(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.DatabaseConfig
		workspaceID string
		expected    string
	}{
		{
			name: "standard workspace",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Prefix:   "nf",
			},
			workspaceID: "workspace123",
			expected:    "postgres://postgres:password@localhost:5432/nf_ws_workspace123?sslmode=disable",
		},
		{
			name: "workspace with hyphens",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Prefix:   "nf",
			},
			workspaceID: "workspace-123",
			expected:    "postgres://postgres:password@localhost:5432/nf_ws_workspace_123?sslmode=disable",
		},
		{
			name: "custom configuration",
			config: &config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "app_user",
				Password: "secure_password",
				Prefix:   "notifuse",
			},
			workspaceID: "client456",
			expected:    "postgres://app_user:secure_password@db.example.com:5433/notifuse_ws_client456?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetWorkspaceDSN(tc.config, tc.workspaceID)
			assert.Equal(t, tc.expected, result)
		})
	}
}
