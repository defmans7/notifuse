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

func TestUserRepository_GetSessionsByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user first
	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}
	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	// Create multiple sessions for the user
	session1 := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
		MagicCode:        "CODE1",
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
	}
	session2 := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now().Add(time.Second), // Ensure different created_at times
		MagicCode:        "CODE2",
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
	}
	session3 := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now().Add(2 * time.Second), // Ensure different created_at times
		MagicCode:        "CODE3",
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
	}

	// Create another user with a session to ensure we're only getting the target user's sessions
	otherUser := &domain.User{
		ID:    uuid.New().String(),
		Email: "other@example.com",
		Name:  "Other User",
	}
	err = repo.CreateUser(context.Background(), otherUser)
	require.NoError(t, err)

	otherSession := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           otherUser.ID,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
		MagicCode:        "OTHER",
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
	}

	// Create all sessions
	ctx := context.Background()
	err = repo.CreateSession(ctx, session1)
	require.NoError(t, err)
	err = repo.CreateSession(ctx, session2)
	require.NoError(t, err)
	err = repo.CreateSession(ctx, session3)
	require.NoError(t, err)
	err = repo.CreateSession(ctx, otherSession)
	require.NoError(t, err)

	// Test retrieving sessions
	sessions, err := repo.GetSessionsByUserID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Verify sessions are returned ordered by created_at DESC
	assert.Equal(t, session3.ID, sessions[0].ID) // Latest session should be first
	assert.Equal(t, session2.ID, sessions[1].ID)
	assert.Equal(t, session1.ID, sessions[2].ID)

	// Test retrieving sessions for a user with no sessions
	nonExistentUserID := uuid.New().String()
	sessions, err = repo.GetSessionsByUserID(ctx, nonExistentUserID)
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestUserRepository_UpdateSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	defer db.Close()

	// Create a test user first
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	// No defer cleanup needed; database is reset between tests

	// Create a session
	now := time.Now().UTC()
	originalExpiry := now.Add(24 * time.Hour)
	originalMagicCode := "ORIGINAL"
	originalMagicCodeExpiry := now.Add(15 * time.Minute)

	session := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		ExpiresAt:        originalExpiry,
		CreatedAt:        now,
		MagicCode:        originalMagicCode,
		MagicCodeExpires: originalMagicCodeExpiry,
	}

	ctx := context.Background()
	err = repo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Verify session was created
	retrievedSession, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, originalMagicCode, retrievedSession.MagicCode)

	// Update the session
	newExpiry := now.Add(48 * time.Hour)
	newMagicCode := "UPDATED"
	newMagicCodeExpiry := now.Add(30 * time.Minute)

	updatedSession := &domain.Session{
		ID:               session.ID,
		UserID:           user.ID,
		ExpiresAt:        newExpiry,
		CreatedAt:        session.CreatedAt, // CreatedAt shouldn't change
		MagicCode:        newMagicCode,
		MagicCodeExpires: newMagicCodeExpiry,
	}

	err = repo.UpdateSession(ctx, updatedSession)
	assert.NoError(t, err)

	// Verify session was updated
	retrievedUpdatedSession, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, newMagicCode, retrievedUpdatedSession.MagicCode)

	// Compare with tolerance for database rounding and using WithinDuration
	assert.WithinDuration(t, newExpiry, retrievedUpdatedSession.ExpiresAt, time.Second)
	assert.WithinDuration(t, newMagicCodeExpiry, retrievedUpdatedSession.MagicCodeExpires, time.Second)
	assert.WithinDuration(t, session.CreatedAt, retrievedUpdatedSession.CreatedAt, time.Second)

	// Test updating a non-existent session
	nonExistentSession := &domain.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err = repo.UpdateSession(ctx, nonExistentSession)
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)
}
