package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notifuse/server/internal/domain"
)

func TestUserRepository_CreateUser_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Setup mock expectations
	mock.ExpectExec(`INSERT INTO users \(id, email, name, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`).
		WithArgs(sqlmock.AnyArg(), user.Email, user.Name, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	err := repo.CreateUser(context.Background(), user)

	// Assert expectations
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}

func TestUserRepository_GetUserByEmail_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("user found", func(t *testing.T) {
		// Setup expected data
		expectedUser := &domain.User{
			ID:        uuid.NewString(),
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now().Round(time.Second),
			UpdatedAt: time.Now().Round(time.Second),
		}

		// Setup mock expectations
		rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE email = \$1`).
			WithArgs(expectedUser.Email).
			WillReturnRows(rows)

		// Call the method under test
		foundUser, err := repo.GetUserByEmail(context.Background(), expectedUser.Email)

		// Assert expectations
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, foundUser.ID)
		assert.Equal(t, expectedUser.Email, foundUser.Email)
		assert.Equal(t, expectedUser.Name, foundUser.Name)
		assert.Equal(t, expectedUser.CreatedAt.Unix(), foundUser.CreatedAt.Unix())
		assert.Equal(t, expectedUser.UpdatedAt.Unix(), foundUser.UpdatedAt.Unix())
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("user not found", func(t *testing.T) {
		email := "notfound@example.com"

		// Setup mock expectations
		mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Call the method under test
		foundUser, err := repo.GetUserByEmail(context.Background(), email)

		// Assert expectations
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.IsType(t, &domain.ErrUserNotFound{}, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})
}

func TestUserRepository_GetUserByID_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("user found", func(t *testing.T) {
		// Setup expected data
		expectedUser := &domain.User{
			ID:        uuid.NewString(),
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now().Round(time.Second),
			UpdatedAt: time.Now().Round(time.Second),
		}

		// Setup mock expectations
		rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).
			WithArgs(expectedUser.ID).
			WillReturnRows(rows)

		// Call the method under test
		foundUser, err := repo.GetUserByID(context.Background(), expectedUser.ID)

		// Assert expectations
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, foundUser.ID)
		assert.Equal(t, expectedUser.Email, foundUser.Email)
		assert.Equal(t, expectedUser.Name, foundUser.Name)
		assert.Equal(t, expectedUser.CreatedAt.Unix(), foundUser.CreatedAt.Unix())
		assert.Equal(t, expectedUser.UpdatedAt.Unix(), foundUser.UpdatedAt.Unix())
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("user not found", func(t *testing.T) {
		id := uuid.NewString()

		// Setup mock expectations
		mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		// Call the method under test
		foundUser, err := repo.GetUserByID(context.Background(), id)

		// Assert expectations
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.IsType(t, &domain.ErrUserNotFound{}, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})
}

func TestUserRepository_CreateSession_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create test data
	userID := uuid.NewString()
	session := &domain.Session{
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Round(time.Second),
	}

	// Setup mock expectations with the correct table name (user_sessions)
	mock.ExpectExec("INSERT INTO user_sessions").
		WithArgs(
			sqlmock.AnyArg(), // ID will be generated
			session.UserID,
			session.ExpiresAt,
			sqlmock.AnyArg(), // CreatedAt will be set
			sqlmock.AnyArg(), // MagicCode (can be null)
			sqlmock.AnyArg(), // MagicCodeExpires (can be null)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	err := repo.CreateSession(context.Background(), session)

	// Assert expectations
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.NotZero(t, session.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
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

func TestUserRepository_GetSessionByID_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Setup test data
	sessionID := "session123"
	now := time.Now().UTC().Round(time.Second)
	expiryTime := now.Add(24 * time.Hour)

	expectedSession := &domain.Session{
		ID:               sessionID,
		UserID:           "user123",
		ExpiresAt:        expiryTime,
		CreatedAt:        now,
		MagicCode:        "MAGICCODE",
		MagicCodeExpires: now.Add(15 * time.Minute),
	}

	// Setup mock expectations with the correct table name (user_sessions)
	rows := sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "magic_code", "magic_code_expires_at"}).
		AddRow(
			expectedSession.ID,
			expectedSession.UserID,
			expectedSession.ExpiresAt,
			expectedSession.CreatedAt,
			expectedSession.MagicCode,
			expectedSession.MagicCodeExpires,
		)

	mock.ExpectQuery("SELECT .* FROM user_sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnRows(rows)

	// Call the method under test
	session, err := repo.GetSessionByID(context.Background(), sessionID)

	// Assert expectations
	require.NoError(t, err)
	assert.Equal(t, expectedSession.ID, session.ID)
	assert.Equal(t, expectedSession.UserID, session.UserID)
	assert.Equal(t, expectedSession.ExpiresAt.Unix(), session.ExpiresAt.Unix())
	assert.Equal(t, expectedSession.CreatedAt.Unix(), session.CreatedAt.Unix())
	assert.Equal(t, expectedSession.MagicCode, session.MagicCode)
	assert.Equal(t, expectedSession.MagicCodeExpires.Unix(), session.MagicCodeExpires.Unix())
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")

	// Test session not found
	nonExistentID := "nonexistent"

	mock.ExpectQuery("SELECT .* FROM user_sessions WHERE id = \\$1").
		WithArgs(nonExistentID).
		WillReturnError(sql.ErrNoRows)

	// Call the method under test
	session, err = repo.GetSessionByID(context.Background(), nonExistentID)

	// Assert expectations
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}

func TestUserRepository_GetSessionsByUserID_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Setup test data
	userID := "user123"
	now := time.Now().UTC().Round(time.Second)

	// Setup mock expectations with the correct table name (user_sessions)
	rows := sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "magic_code", "magic_code_expires_at"}).
		// Latest session first (created_at DESC order)
		AddRow("session3", userID, now.Add(24*time.Hour), now.Add(2*time.Second), "CODE3", now.Add(15*time.Minute)).
		AddRow("session2", userID, now.Add(24*time.Hour), now.Add(time.Second), "CODE2", now.Add(15*time.Minute)).
		AddRow("session1", userID, now.Add(24*time.Hour), now, "CODE1", now.Add(15*time.Minute))

	mock.ExpectQuery("SELECT .* FROM user_sessions WHERE user_id = \\$1 ORDER BY created_at DESC").
		WithArgs(userID).
		WillReturnRows(rows)

	// Call the method under test
	sessions, err := repo.GetSessionsByUserID(context.Background(), userID)

	// Assert expectations
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Latest session should be first
	assert.Equal(t, "session3", sessions[0].ID)
	assert.Equal(t, "session2", sessions[1].ID)
	assert.Equal(t, "session1", sessions[2].ID)

	// All sessions should have the same user ID
	for _, session := range sessions {
		assert.Equal(t, userID, session.UserID)
	}

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")

	// Test for user with no sessions
	userIDNoSessions := "user456"

	mock.ExpectQuery("SELECT .* FROM user_sessions WHERE user_id = \\$1 ORDER BY created_at DESC").
		WithArgs(userIDNoSessions).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "magic_code", "magic_code_expires_at"}))

	// Call the method under test
	emptySessions, err := repo.GetSessionsByUserID(context.Background(), userIDNoSessions)

	// Assert expectations
	require.NoError(t, err)
	assert.Empty(t, emptySessions)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}

func TestUserRepository_DeleteSession_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Setup test data
	sessionID := "session123"

	// Setup mock expectations with the correct table name (user_sessions)
	mock.ExpectExec("DELETE FROM user_sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method under test
	err := repo.DeleteSession(context.Background(), sessionID)

	// Assert expectations
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")

	// Test deleting non-existent session
	nonExistentID := "nonexistent"

	mock.ExpectExec("DELETE FROM user_sessions WHERE id = \\$1").
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Zero rows affected

	// Call the method under test
	err = repo.DeleteSession(context.Background(), nonExistentID)

	// Assert expectations
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}

func TestUserRepository_UpdateSession_WithMock(t *testing.T) {
	// Setup mock database
	db, mock := setupMockTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Setup test data
	now := time.Now().UTC().Round(time.Second)
	session := &domain.Session{
		ID:               "session123",
		UserID:           "user123",
		ExpiresAt:        now.Add(48 * time.Hour),
		CreatedAt:        now.Add(-24 * time.Hour), // Created yesterday
		MagicCode:        "NEWCODE",
		MagicCodeExpires: now.Add(30 * time.Minute),
	}

	// Setup mock expectations with the correct table name (user_sessions)
	mock.ExpectExec("UPDATE user_sessions SET").
		WithArgs(
			session.ExpiresAt,
			session.MagicCode,
			session.MagicCodeExpires,
			session.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method under test
	err := repo.UpdateSession(context.Background(), session)

	// Assert expectations
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")

	// Test updating non-existent session
	nonExistentSession := &domain.Session{
		ID:               "nonexistent",
		UserID:           "user123",
		ExpiresAt:        now.Add(24 * time.Hour),
		MagicCode:        "CODE",
		MagicCodeExpires: now.Add(15 * time.Minute),
	}

	mock.ExpectExec("UPDATE user_sessions SET").
		WithArgs(
			nonExistentSession.ExpiresAt,
			nonExistentSession.MagicCode,
			nonExistentSession.MagicCodeExpires,
			nonExistentSession.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Zero rows affected

	// Call the method under test
	err = repo.UpdateSession(context.Background(), nonExistentSession)

	// Assert expectations
	assert.Error(t, err)
	assert.IsType(t, &domain.ErrSessionNotFound{}, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}
