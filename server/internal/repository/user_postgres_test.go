package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notifuse/server/internal/domain"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)

	// Verify user was created
	savedUser, err := repo.GetUserByEmail(context.Background(), user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, savedUser.ID)
	assert.Equal(t, user.Email, savedUser.Email)
	assert.Equal(t, user.Name, savedUser.Name)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Test not found
	_, err := repo.GetUserByEmail(context.Background(), "notfound@example.com")
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrUserNotFound{}, err)

	// Create user
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	// Test found
	foundUser, err := repo.GetUserByEmail(context.Background(), user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Name, foundUser.Name)
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Test not found
	_, err := repo.GetUserByID(context.Background(), uuid.New().String())
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrUserNotFound{}, err)

	// Create user
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	// Test found
	foundUser, err := repo.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Name, foundUser.Name)
}

func TestUserRepository_CreateSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create user first
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	// Create session
	session := &domain.Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}

	err = repo.CreateSession(context.Background(), session)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.NotZero(t, session.CreatedAt)

	// Verify session was created
	savedSession, err := repo.GetSessionByID(context.Background(), session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, savedSession.ID)
	assert.Equal(t, session.UserID, savedSession.UserID)
	assert.WithinDuration(t, session.ExpiresAt, savedSession.ExpiresAt, time.Second)
}

func TestUserRepository_DeleteSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Test delete non-existent session
	err := repo.DeleteSession(context.Background(), uuid.New().String())
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)

	// Create user and session
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = repo.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Test delete existing session
	err = repo.DeleteSession(context.Background(), session.ID)
	require.NoError(t, err)

	// Verify session was deleted
	_, err = repo.GetSessionByID(context.Background(), session.ID)
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)
}
