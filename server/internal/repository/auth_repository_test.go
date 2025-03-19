package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/pkg/logger"
)

// MockLogger for repository tests
type MockLogger struct {
	testifyMock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func TestSQLAuthRepository_GetSessionByID_WithMock(t *testing.T) {
	// Create sqlmock using our helper
	db, mock := setupMockTestDB(t)
	defer db.Close()

	mockLogger := new(MockLogger)
	mockLogger.On("WithField", testifyMock.Anything, testifyMock.Anything).Return(mockLogger)

	repo := NewSQLAuthRepository(db, mockLogger)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		sessionID := "test-session"
		userID := "test-user"
		expectedTime := time.Now().Add(time.Hour)

		// Set up mock to return the expected expiry time
		rows := sqlmock.NewRows([]string{"expires_at"}).AddRow(expectedTime)
		mock.ExpectQuery("SELECT expires_at FROM sessions WHERE").
			WithArgs(sessionID, userID).
			WillReturnRows(rows)

		// Call the method
		expiresAt, err := repo.GetSessionByID(ctx, sessionID, userID)

		// Assert that no error occurred
		require.NoError(t, err)
		assert.NotNil(t, expiresAt)
		assert.Equal(t, expectedTime.Unix(), expiresAt.Unix())

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("session not found", func(t *testing.T) {
		sessionID := "nonexistent-session"
		userID := "test-user"

		// Set up mock to return no rows
		mock.ExpectQuery("SELECT expires_at FROM sessions WHERE").
			WithArgs(sessionID, userID).
			WillReturnError(sql.ErrNoRows)

		// Call the method
		expiresAt, err := repo.GetSessionByID(ctx, sessionID, userID)

		// Assert that the expected error occurred
		require.Error(t, err)
		assert.Nil(t, expiresAt)
		assert.Equal(t, sql.ErrNoRows, err)

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("database error", func(t *testing.T) {
		sessionID := "test-session"
		userID := "test-user"

		// Set up mock to return a database error
		mock.ExpectQuery("SELECT expires_at FROM sessions WHERE").
			WithArgs(sessionID, userID).
			WillReturnError(sql.ErrConnDone)

		// Call the method
		expiresAt, err := repo.GetSessionByID(ctx, sessionID, userID)

		// Assert that the expected error occurred
		require.Error(t, err)
		assert.Nil(t, expiresAt)
		assert.Equal(t, sql.ErrConnDone, err)

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})
}

func TestSQLAuthRepository_GetUserByID_WithMock(t *testing.T) {
	// Create sqlmock using our helper
	db, mock := setupMockTestDB(t)
	defer db.Close()

	mockLogger := new(MockLogger)
	mockLogger.On("WithField", testifyMock.Anything, testifyMock.Anything).Return(mockLogger)

	repo := NewSQLAuthRepository(db, mockLogger)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		userID := "test-user"
		expectedEmail := "test@example.com"
		expectedCreatedAt := time.Now()

		// Set up mock to return the expected user data
		rows := sqlmock.NewRows([]string{"id", "email", "created_at"}).
			AddRow(userID, expectedEmail, expectedCreatedAt)
		mock.ExpectQuery("SELECT id, email, created_at FROM users WHERE").
			WithArgs(userID).
			WillReturnRows(rows)

		// Call the method
		user, err := repo.GetUserByID(ctx, userID)

		// Assert that no error occurred
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, expectedEmail, user.Email)
		assert.Equal(t, expectedCreatedAt.Unix(), user.CreatedAt.Unix())

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent-user"

		// Set up mock to return no rows
		mock.ExpectQuery("SELECT id, email, created_at FROM users WHERE").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		// Call the method
		user, err := repo.GetUserByID(ctx, userID)

		// Assert that the expected error occurred
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, sql.ErrNoRows, err)

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})

	t.Run("database error", func(t *testing.T) {
		userID := "test-user"

		// Set up mock to return a database error
		mock.ExpectQuery("SELECT id, email, created_at FROM users WHERE").
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		// Call the method
		user, err := repo.GetUserByID(ctx, userID)

		// Assert that the expected error occurred
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, sql.ErrConnDone, err)

		// Assert that all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
	})
}
