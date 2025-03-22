package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_VerifyUserSession(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Create key pair for testing
	key := paseto.NewV4AsymmetricSecretKey()
	privateKey := key.ExportBytes()
	publicKey := key.Public().ExportBytes()

	// Create service with config
	service, err := NewAuthService(AuthServiceConfig{
		Repository: mockRepo,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Logger:     mockLogger,
	})
	require.NoError(t, err)

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

// Test the constructor with config
func TestAuthService_NewAuthService(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	// Create key pair for testing
	key := paseto.NewV4AsymmetricSecretKey()
	privateKey := key.ExportBytes()
	publicKey := key.Public().ExportBytes()

	service, err := NewAuthService(AuthServiceConfig{
		Repository: mockRepo,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Logger:     mockLogger,
	})

	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, mockLogger, service.logger)
	// Cannot directly compare paseto keys as they are interfaces
	assert.NotNil(t, service.privateKey)
	assert.NotNil(t, service.publicKey)
}

// Test the GenerateAuthToken method
func TestAuthService_GenerateAuthToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Create key pair for testing
	key := paseto.NewV4AsymmetricSecretKey()
	privateKey := key.ExportBytes()
	publicKey := key.Public().ExportBytes()

	// Create service with config
	service, err := NewAuthService(AuthServiceConfig{
		Repository: mockRepo,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Logger:     mockLogger,
	})
	require.NoError(t, err)

	// Create a user and session for testing
	user := &domain.User{
		ID:        "test-user-id",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sessionID := "test-session-id"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Generate token
	token := service.GenerateAuthToken(user, sessionID, expiresAt)

	// Verify the token is not empty
	assert.NotEmpty(t, token)

	// Parse the token to verify its contents
	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Public(key.Public(), token, nil)
	require.NoError(t, err)

	// Verify token claims
	userId, err := parsedToken.GetString("user_id")
	require.NoError(t, err)
	assert.Equal(t, user.ID, userId)

	email, err := parsedToken.GetString("email")
	require.NoError(t, err)
	assert.Equal(t, user.Email, email)

	sessionIdFromToken, err := parsedToken.GetString("session_id")
	require.NoError(t, err)
	assert.Equal(t, sessionID, sessionIdFromToken)
}

// Test the GenerateInvitationToken method
func TestAuthService_GenerateInvitationToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Create key pair for testing
	key := paseto.NewV4AsymmetricSecretKey()
	privateKey := key.ExportBytes()
	publicKey := key.Public().ExportBytes()

	// Create service with config
	service, err := NewAuthService(AuthServiceConfig{
		Repository: mockRepo,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Logger:     mockLogger,
	})
	require.NoError(t, err)

	// Create a workspace invitation for testing
	invitationID := "test-invitation-id"
	workspaceID := "test-workspace-id"
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	invitation := &domain.WorkspaceInvitation{
		ID:          invitationID,
		WorkspaceID: workspaceID,
		Email:       email,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	// Generate token
	token := service.GenerateInvitationToken(invitation)

	// Verify the token is not empty
	assert.NotEmpty(t, token)

	// Parse the token to verify its contents
	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Public(key.Public(), token, nil)
	require.NoError(t, err)

	// Verify token claims
	invitationIdFromToken, err := parsedToken.GetString("invitation_id")
	require.NoError(t, err)
	assert.Equal(t, invitationID, invitationIdFromToken)

	workspaceIdFromToken, err := parsedToken.GetString("workspace_id")
	require.NoError(t, err)
	assert.Equal(t, workspaceID, workspaceIdFromToken)

	emailFromToken, err := parsedToken.GetString("email")
	require.NoError(t, err)
	assert.Equal(t, email, emailFromToken)
}
