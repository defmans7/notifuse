package service

import (
	"context"
	"database/sql"
	"errors"
	"notifuse/server/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthRepository is a mock implementation of the AuthRepository interface
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetSessionByID(ctx context.Context, sessionID string, userID string) (*time.Time, error) {
	args := m.Called(ctx, sessionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthService_VerifyUserSession(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewAuthService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user123"
	sessionID := "session456"

	t.Run("valid session", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// Future expiration time
		futureTime := time.Now().Add(1 * time.Hour)
		expectedUser := &domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: time.Now(),
		}

		// Setup expectations
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(&futureTime, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// Session not found
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(nil, sql.ErrNoRows)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrSessionExpired, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("session query error", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// General error retrieving session
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(nil, errors.New("database error"))

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// Past expiration time
		pastTime := time.Now().Add(-1 * time.Hour)
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(&pastTime, nil)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrSessionExpired, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// Session valid but user not found
		futureTime := time.Now().Add(1 * time.Hour)
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(&futureTime, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, sql.ErrNoRows)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user query error", func(t *testing.T) {
		// Reset mock
		mockRepo.Mock = mock.Mock{}

		// Session valid but error retrieving user
		futureTime := time.Now().Add(1 * time.Hour)
		mockRepo.On("GetSessionByID", ctx, sessionID, userID).Return(&futureTime, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("database error"))

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

// Simple constructor test
func TestAuthService_NewAuthService(t *testing.T) {
	mockLogger := new(MockLogger)
	mockRepo := new(MockAuthRepository)

	service := NewAuthService(mockRepo, mockLogger)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, mockLogger, service.logger)
}
