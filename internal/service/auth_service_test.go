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

func TestNewAuthService(t *testing.T) {
	tests := []struct {
		name          string
		config        AuthServiceConfig
		setupMocks    func(*MockLogger)
		expectError   bool
		errorContains string
	}{
		{
			name: "successful creation",
			config: AuthServiceConfig{
				Repository:          &MockAuthRepository{},
				WorkspaceRepository: &MockWorkspaceRepository{},
				PrivateKey:          paseto.NewV4AsymmetricSecretKey().ExportBytes(),
				PublicKey:           paseto.NewV4AsymmetricSecretKey().Public().ExportBytes(),
				Logger:              &MockLogger{},
			},
			expectError: false,
		},
		{
			name: "invalid private key",
			config: AuthServiceConfig{
				Repository:          &MockAuthRepository{},
				WorkspaceRepository: &MockWorkspaceRepository{},
				PrivateKey:          []byte("invalid key"),
				PublicKey:           paseto.NewV4AsymmetricSecretKey().Public().ExportBytes(),
				Logger:              &MockLogger{},
			},
			setupMocks: func(mockLogger *MockLogger) {
				mockLogger.On("WithField", "error", "key length incorrect (11), expected 64").Return(mockLogger)
				mockLogger.On("Error", "Error creating PASETO private key").Return()
			},
			expectError:   true,
			errorContains: "key length incorrect (11), expected 64",
		},
		{
			name: "invalid public key",
			config: AuthServiceConfig{
				Repository:          &MockAuthRepository{},
				WorkspaceRepository: &MockWorkspaceRepository{},
				PrivateKey:          paseto.NewV4AsymmetricSecretKey().ExportBytes(),
				PublicKey:           []byte("invalid key"),
				Logger:              &MockLogger{},
			},
			setupMocks: func(mockLogger *MockLogger) {
				mockLogger.On("WithField", "error", "key length incorrect (11), expected 32").Return(mockLogger)
				mockLogger.On("Error", "Error creating PASETO public key").Return()
			},
			expectError:   true,
			errorContains: "key length incorrect (11), expected 32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks(tt.config.Logger.(*MockLogger))
			}
			service, err := NewAuthService(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, service)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, service)
		})
	}
}

func TestAuthenticateUserFromContext(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		setupMocks    func(*MockAuthRepository, *MockLogger)
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing user_id in context",
			ctx:           context.Background(),
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:          "missing session_id in context",
			ctx:           context.WithValue(context.Background(), "user_id", "test-user"),
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name: "valid context",
			ctx:  context.WithValue(context.WithValue(context.Background(), "user_id", "test-user"), "session_id", "test-session"),
			setupMocks: func(mockRepo *MockAuthRepository, mockLogger *MockLogger) {
				futureTime := time.Now().Add(1 * time.Hour)
				mockRepo.On("GetSessionByID", mock.Anything, "test-session", "test-user").Return(&futureTime, nil)
				mockRepo.On("GetUserByID", mock.Anything, "test-user").Return(&domain.User{ID: "test-user"}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockAuthRepository{}
			mockLogger := &MockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockLogger)
			}

			service := &AuthService{
				repo:   mockRepo,
				logger: mockLogger,
			}

			user, err := service.AuthenticateUserFromContext(tt.ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, user)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, user)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthenticateUserForWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		workspaceID   string
		setupMocks    func(*MockAuthRepository, *MockWorkspaceRepository, *MockLogger)
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful authentication",
			ctx:         context.WithValue(context.WithValue(context.Background(), "user_id", "test-user"), "session_id", "test-session"),
			workspaceID: "test-workspace",
			setupMocks: func(authRepo *MockAuthRepository, workspaceRepo *MockWorkspaceRepository, mockLogger *MockLogger) {
				futureTime := time.Now().Add(1 * time.Hour)
				authRepo.On("GetSessionByID", mock.Anything, "test-session", "test-user").Return(&futureTime, nil)
				authRepo.On("GetUserByID", mock.Anything, "test-user").Return(&domain.User{ID: "test-user"}, nil)
				workspaceRepo.On("GetUserWorkspace", mock.Anything, "test-user", "test-workspace").Return(&domain.UserWorkspace{}, nil)
			},
			expectError: false,
		},
		{
			name:        "workspace not found",
			ctx:         context.WithValue(context.WithValue(context.Background(), "user_id", "test-user"), "session_id", "test-session"),
			workspaceID: "test-workspace",
			setupMocks: func(authRepo *MockAuthRepository, workspaceRepo *MockWorkspaceRepository, mockLogger *MockLogger) {
				futureTime := time.Now().Add(1 * time.Hour)
				authRepo.On("GetSessionByID", mock.Anything, "test-session", "test-user").Return(&futureTime, nil)
				authRepo.On("GetUserByID", mock.Anything, "test-user").Return(&domain.User{ID: "test-user"}, nil)
				workspaceRepo.On("GetUserWorkspace", mock.Anything, "test-user", "test-workspace").Return(nil, sql.ErrNoRows)
			},
			expectError: true,
		},
		{
			name:        "workspace repository error",
			ctx:         context.WithValue(context.WithValue(context.Background(), "user_id", "test-user"), "session_id", "test-session"),
			workspaceID: "test-workspace",
			setupMocks: func(authRepo *MockAuthRepository, workspaceRepo *MockWorkspaceRepository, mockLogger *MockLogger) {
				futureTime := time.Now().Add(1 * time.Hour)
				authRepo.On("GetSessionByID", mock.Anything, "test-session", "test-user").Return(&futureTime, nil)
				authRepo.On("GetUserByID", mock.Anything, "test-user").Return(&domain.User{ID: "test-user"}, nil)
				workspaceRepo.On("GetUserWorkspace", mock.Anything, "test-user", "test-workspace").Return(nil, assert.AnError)
			},
			expectError: true,
		},
		{
			name:        "authentication error",
			ctx:         context.WithValue(context.WithValue(context.Background(), "user_id", "test-user"), "session_id", "test-session"),
			workspaceID: "test-workspace",
			setupMocks: func(authRepo *MockAuthRepository, workspaceRepo *MockWorkspaceRepository, mockLogger *MockLogger) {
				futureTime := time.Now().Add(1 * time.Hour)
				authRepo.On("GetSessionByID", mock.Anything, "test-session", "test-user").Return(&futureTime, nil)
				authRepo.On("GetUserByID", mock.Anything, "test-user").Return(nil, sql.ErrNoRows)
				mockLogger.On("WithField", "user_id", "test-user").Return(mockLogger)
				mockLogger.On("Error", "User not found").Return()
			},
			expectError:   true,
			errorContains: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := &MockAuthRepository{}
			mockWorkspaceRepo := &MockWorkspaceRepository{}
			mockLogger := &MockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuthRepo, mockWorkspaceRepo, mockLogger)
			}

			service := &AuthService{
				repo:          mockAuthRepo,
				workspaceRepo: mockWorkspaceRepo,
				logger:        mockLogger,
			}

			user, err := service.AuthenticateUserForWorkspace(tt.ctx, tt.workspaceID)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, user)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, user)
			mockAuthRepo.AssertExpectations(t)
			mockWorkspaceRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetPrivateKey(t *testing.T) {
	privateKey := paseto.NewV4AsymmetricSecretKey()
	service := &AuthService{
		privateKey: privateKey,
	}

	result := service.GetPrivateKey()
	assert.Equal(t, privateKey, result)
}

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockAuthRepository, *MockLogger)
		expectError   bool
		errorContains string
	}{
		{
			name:   "successful retrieval",
			userID: "test-user",
			setupMocks: func(repo *MockAuthRepository, mockLogger *MockLogger) {
				repo.On("GetUserByID", mock.Anything, "test-user").Return(&domain.User{ID: "test-user"}, nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: "test-user",
			setupMocks: func(repo *MockAuthRepository, mockLogger *MockLogger) {
				repo.On("GetUserByID", mock.Anything, "test-user").Return(nil, sql.ErrNoRows)
			},
			expectError:   true,
			errorContains: "user not found",
		},
		{
			name:   "repository error",
			userID: "test-user",
			setupMocks: func(repo *MockAuthRepository, mockLogger *MockLogger) {
				repo.On("GetUserByID", mock.Anything, "test-user").Return(nil, assert.AnError)
				mockLogger.On("WithField", "error", assert.AnError.Error()).Return(mockLogger)
				mockLogger.On("WithField", "user_id", "test-user").Return(mockLogger)
				mockLogger.On("Error", "Failed to get user by ID").Return()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockAuthRepository{}
			mockLogger := &MockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockLogger)
			}

			service := &AuthService{
				repo:   mockRepo,
				logger: mockLogger,
			}

			user, err := service.GetUserByID(context.Background(), tt.userID)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, user)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, user)
			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
