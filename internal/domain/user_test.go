package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	user := User{
		ID:        "user123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Basic assertion that the struct fields are set correctly
	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
}

func TestWorkspaceUserKey(t *testing.T) {
	// Test with a regular workspace ID
	workspaceID := "workspace123"
	key := WorkspaceUserKey(workspaceID)
	assert.Equal(t, contextKey("workspace_user_workspace123"), key)

	// Test with empty string
	emptyKey := WorkspaceUserKey("")
	assert.Equal(t, contextKey("workspace_user_"), emptyKey)

	// Test with special characters
	specialKey := WorkspaceUserKey("workspace-123_456@example")
	assert.Equal(t, contextKey("workspace_user_workspace-123_456@example"), specialKey)
}

func TestSession(t *testing.T) {
	now := time.Now()
	expiry := now.Add(time.Hour * 24)

	session := Session{
		ID:               "session123",
		UserID:           "user123",
		ExpiresAt:        expiry,
		CreatedAt:        now,
		MagicCode:        "ABCDEF",
		MagicCodeExpires: now.Add(time.Minute * 15),
	}

	// Basic assertion that the struct fields are set correctly
	assert.Equal(t, "session123", session.ID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, expiry, session.ExpiresAt)
	assert.Equal(t, now, session.CreatedAt)
	assert.Equal(t, "ABCDEF", session.MagicCode)
}

func TestErrUserNotFound_Error(t *testing.T) {
	err := &ErrUserNotFound{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestErrUserExists_Error(t *testing.T) {
	err := &ErrUserExists{Message: "user already exists"}
	assert.Equal(t, "user already exists", err.Error())
}

func TestErrSessionNotFound_Error(t *testing.T) {
	err := &ErrSessionNotFound{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}
