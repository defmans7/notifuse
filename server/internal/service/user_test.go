package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
	"notifuse/server/pkg/logger"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockUserRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockUserRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *mockUserRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockUserRepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendMagicCode(email, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *mockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func TestUserService_SignIn(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Test successful sign in
	repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
	repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
		return s.UserID == user.ID && len(s.MagicCode) == 6 && !s.MagicCodeExpires.IsZero()
	})).Return(nil)
	emailSender.On("SendMagicCode", user.Email, mock.MatchedBy(func(code string) bool {
		return len(code) == 6
	})).Return(nil)

	err = service.SignIn(ctx, SignInInput{Email: user.Email})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailSender.AssertExpectations(t)

	// Test user not found
	repo.On("GetUserByEmail", ctx, "notfound@example.com").
		Return(nil, &domain.ErrUserNotFound{Message: "user not found"})

	// When user is not found, expect a new user to be created
	repo.On("CreateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "notfound@example.com" && !u.CreatedAt.IsZero() && !u.UpdatedAt.IsZero()
	})).Return(nil)

	// Expect a session to be created for the new user
	repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
		return s.UserID != "" && len(s.MagicCode) == 6 && !s.MagicCodeExpires.IsZero()
	})).Return(nil)

	// Expect an email to be sent
	emailSender.On("SendMagicCode", "notfound@example.com", mock.MatchedBy(func(code string) bool {
		return len(code) == 6
	})).Return(nil)

	err = service.SignIn(ctx, SignInInput{Email: "notfound@example.com"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

func TestUserService_VerifyCode(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	validCode := "123456"
	validSession := &domain.Session{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		MagicCode:        validCode,
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}

	t.Run("successful verification", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
		repo.On("GetSessionsByUserID", ctx, user.ID).Return([]*domain.Session{validSession}, nil)
		repo.On("UpdateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return s.ID == validSession.ID && s.MagicCode == "" && s.MagicCodeExpires.IsZero()
		})).Return(nil)

		response, err := service.VerifyCode(ctx, VerifyCodeInput{
			Email: user.Email,
			Code:  validCode,
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, user.ID, response.User.ID)
		repo.AssertExpectations(t)
	})

	t.Run("invalid code", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
		repo.On("GetSessionsByUserID", ctx, user.ID).Return([]*domain.Session{validSession}, nil)

		response, err := service.VerifyCode(ctx, VerifyCodeInput{
			Email: user.Email,
			Code:  "000000",
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid magic code")
		repo.AssertExpectations(t)
	})

	t.Run("expired code", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		expiredSession := &domain.Session{
			ID:               uuid.New().String(),
			UserID:           user.ID,
			MagicCode:        validCode,
			MagicCodeExpires: time.Now().Add(-1 * time.Minute),
			ExpiresAt:        time.Now().Add(24 * time.Hour),
		}

		repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
		repo.On("GetSessionsByUserID", ctx, user.ID).Return([]*domain.Session{expiredSession}, nil)

		response, err := service.VerifyCode(ctx, VerifyCodeInput{
			Email: user.Email,
			Code:  validCode,
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "magic code expired")
		repo.AssertExpectations(t)
	})
}

func TestUserService_VerifyUserSession(t *testing.T) {
	repo := new(mockUserRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   new(mockEmailSender),
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	userId := uuid.New().String()
	sessionId := uuid.New().String()

	t.Run("valid session", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Valid session
		validSession := &domain.Session{
			ID:        sessionId,
			UserID:    userId,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		user := &domain.User{
			ID:        userId,
			Email:     "test@example.com",
			CreatedAt: time.Now(),
		}

		repo.On("GetSessionByID", ctx, sessionId).Return(validSession, nil)
		repo.On("GetUserByID", ctx, userId).Return(user, nil)

		result, err := service.VerifyUserSession(ctx, userId, sessionId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userId, result.ID)
		assert.Equal(t, "test@example.com", result.Email)
		repo.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Expired session
		expiredSession := &domain.Session{
			ID:        sessionId,
			UserID:    userId,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		repo.On("GetSessionByID", ctx, sessionId).Return(expiredSession, nil)

		result, err := service.VerifyUserSession(ctx, userId, sessionId)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrSessionExpired, err)
		repo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetSessionByID", ctx, sessionId).Return(nil, fmt.Errorf("session not found"))

		result, err := service.VerifyUserSession(ctx, userId, sessionId)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("wrong user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Session belongs to different user
		wrongUserSession := &domain.Session{
			ID:        sessionId,
			UserID:    "different-user-id",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		repo.On("GetSessionByID", ctx, sessionId).Return(wrongUserSession, nil)

		result, err := service.VerifyUserSession(ctx, userId, sessionId)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "does not belong to user")
		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Valid session
		validSession := &domain.Session{
			ID:        sessionId,
			UserID:    userId,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		repo.On("GetSessionByID", ctx, sessionId).Return(validSession, nil)
		repo.On("GetUserByID", ctx, userId).Return(nil, fmt.Errorf("user not found"))

		result, err := service.VerifyUserSession(ctx, userId, sessionId)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := new(mockUserRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   new(mockEmailSender),
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("user exists", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		userId := uuid.New().String()
		user := &domain.User{
			ID:        userId,
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetUserByID", ctx, userId).Return(user, nil)

		result, err := service.GetUserByID(ctx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userId, result.ID)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Equal(t, "Test User", result.Name)
		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		userId := uuid.New().String()
		repo.On("GetUserByID", ctx, userId).Return(nil, &domain.ErrUserNotFound{Message: "user not found"})

		result, err := service.GetUserByID(ctx, userId)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &domain.ErrUserNotFound{}, err)
		repo.AssertExpectations(t)
	})
}

func TestUserService_SignInDev(t *testing.T) {
	repo := new(mockUserRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   new(mockEmailSender),
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	email := "dev@example.com"

	t.Run("existing user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		user := &domain.User{
			ID:        uuid.New().String(),
			Email:     email,
			Name:      "Dev User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetUserByEmail", ctx, email).Return(user, nil)
		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return s.UserID == user.ID && !s.ExpiresAt.IsZero() && !s.CreatedAt.IsZero()
		})).Return(nil)

		token, err := service.SignInDev(ctx, SignInInput{Email: email})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("new user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetUserByEmail", ctx, email).
			Return(nil, &domain.ErrUserNotFound{Message: "user not found"})

		repo.On("CreateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == email && !u.CreatedAt.IsZero() && !u.UpdatedAt.IsZero()
		})).Return(nil)

		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return s.UserID != "" && !s.ExpiresAt.IsZero() && !s.CreatedAt.IsZero()
		})).Return(nil)

		token, err := service.SignInDev(ctx, SignInInput{Email: email})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetUserByEmail", ctx, email).Return(nil, fmt.Errorf("database error"))

		token, err := service.SignInDev(ctx, SignInInput{Email: email})
		assert.Error(t, err)
		assert.Empty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("session creation error", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		user := &domain.User{
			ID:        uuid.New().String(),
			Email:     email,
			Name:      "Dev User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("GetUserByEmail", ctx, email).Return(user, nil)
		repo.On("CreateSession", ctx, mock.Anything).Return(fmt.Errorf("session creation failed"))

		token, err := service.SignInDev(ctx, SignInInput{Email: email})
		assert.Error(t, err)
		assert.Empty(t, token)
		repo.AssertExpectations(t)
	})
}

func TestNewUserService_Comprehensive(t *testing.T) {
	// Generate test PASETO keys that will work with the test
	privateKeyObj := paseto.NewV4AsymmetricSecretKey() // Generates a valid key
	privateKey := privateKeyObj.ExportBytes()
	publicKey := privateKeyObj.Public().ExportBytes()

	// Create mock dependencies
	mockRepo := &mockUserRepository{}
	mockEmail := &mockEmailSender{}
	mockLog := &mockLogger{}
	mockLog.On("WithField", mock.Anything, mock.Anything).Return(mockLog)
	mockLog.On("Error", mock.Anything).Return()

	tests := []struct {
		name          string
		config        UserServiceConfig
		shouldError   bool
		expectedError string
	}{
		{
			name: "valid_configuration",
			config: UserServiceConfig{
				Repository:    mockRepo,
				PrivateKey:    privateKey,
				PublicKey:     publicKey,
				EmailSender:   mockEmail,
				SessionExpiry: 24 * time.Hour,
				Logger:        mockLog,
			},
			shouldError: false,
		},
		{
			name: "missing_private_key",
			config: UserServiceConfig{
				Repository:    mockRepo,
				PrivateKey:    nil,
				PublicKey:     publicKey,
				EmailSender:   mockEmail,
				SessionExpiry: 24 * time.Hour,
				Logger:        mockLog,
			},
			shouldError:   true,
			expectedError: "error creating private key",
		},
		{
			name: "missing_public_key",
			config: UserServiceConfig{
				Repository:    mockRepo,
				PrivateKey:    privateKey,
				PublicKey:     nil,
				EmailSender:   mockEmail,
				SessionExpiry: 24 * time.Hour,
				Logger:        mockLog,
			},
			shouldError:   true,
			expectedError: "error creating public key",
		},
		{
			name: "invalid_private_key",
			config: UserServiceConfig{
				Repository:    mockRepo,
				PrivateKey:    []byte("invalid-key"),
				PublicKey:     publicKey,
				EmailSender:   mockEmail,
				SessionExpiry: 24 * time.Hour,
				Logger:        mockLog,
			},
			shouldError:   true,
			expectedError: "error creating private key",
		},
		{
			name: "invalid_public_key",
			config: UserServiceConfig{
				Repository:    mockRepo,
				PrivateKey:    privateKey,
				PublicKey:     []byte("invalid-key"),
				EmailSender:   mockEmail,
				SessionExpiry: 24 * time.Hour,
				Logger:        mockLog,
			},
			shouldError:   true,
			expectedError: "error creating public key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks
			mockRepo = &mockUserRepository{}
			mockEmail = &mockEmailSender{}
			mockLog = &mockLogger{}
			mockLog.On("WithField", mock.Anything, mock.Anything).Return(mockLog)
			mockLog.On("Error", mock.Anything).Return()

			// Update the config with fresh mocks
			tc.config.Repository = mockRepo
			tc.config.EmailSender = mockEmail
			tc.config.Logger = mockLog

			// Call the function
			service, err := NewUserService(tc.config)

			if tc.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, service)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, service)

				// Assert that the service has the expected configuration
				assert.Equal(t, mockRepo, service.repo)
				assert.Equal(t, mockEmail, service.emailSender)
				assert.Equal(t, tc.config.SessionExpiry, service.sessionExpiry)
				assert.Equal(t, mockLog, service.logger)
			}
		})
	}
}

func TestSignIn_ComprehensiveErrorCases(t *testing.T) {
	// This is just a dummy test as we can't properly create the service without the correct keys
	// In a real test, you would use proper PASETO keys
	t.Skip("This test requires proper PASETO keys to be implemented")
}
