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

	"github.com/Notifuse/notifuse/internal/domain"
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

// Setup a real AuthService for testing
func setupAuthService() *AuthService {
	// Create key pair for testing
	key := paseto.NewV4AsymmetricSecretKey()
	privateKey := key.ExportBytes()
	publicKey := key.Public().ExportBytes()

	mockRepo := new(MockAuthRepository)
	mockLogger := new(MockLogger)

	service, _ := NewAuthService(AuthServiceConfig{
		Repository: mockRepo,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Logger:     mockLogger,
	})

	return service
}

func TestUserService_SignIn(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)
	authService := setupAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		AuthService:   authService,
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
	authService := setupAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		AuthService:   authService,
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
			Code:  "invalid",
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
			MagicCodeExpires: time.Now().Add(-1 * time.Minute), // Expired
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
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)
	authService := setupAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		AuthService:   authService,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	userID := uuid.New().String()
	sessionID := uuid.New().String()

	t.Run("valid session", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Expected session and user for this test
		expectedSession := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
		}

		expectedUser := &domain.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "Test User",
		}

		// Set up expectations
		repo.On("GetSessionByID", ctx, sessionID).Return(expectedSession, nil)
		repo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		repo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Session not found
		repo.On("GetSessionByID", ctx, sessionID).Return(nil, fmt.Errorf("session not found"))

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "session not found")
		repo.AssertExpectations(t)
	})

	t.Run("session does not belong to user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Session belongs to another user
		session := &domain.Session{
			ID:        sessionID,
			UserID:    "another-user-id",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
		}

		repo.On("GetSessionByID", ctx, sessionID).Return(session, nil)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "session does not belong to user")
		repo.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Session is expired
		expiredSession := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt: time.Now(),
		}

		repo.On("GetSessionByID", ctx, sessionID).Return(expiredSession, nil)

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrSessionExpired, err)
		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Session is valid but user not found
		validSession := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
		}

		repo.On("GetSessionByID", ctx, sessionID).Return(validSession, nil)
		repo.On("GetUserByID", ctx, userID).Return(nil, fmt.Errorf("user not found"))

		user, err := service.VerifyUserSession(ctx, userID, sessionID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		repo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)
	authService := setupAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		AuthService:   authService,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
	})
	require.NoError(t, err)

	ctx := context.Background()
	userID := uuid.New().String()

	t.Run("successful retrieval", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "Test User",
		}

		repo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		repo.On("GetUserByID", ctx, userID).Return(nil, fmt.Errorf("user not found"))

		user, err := service.GetUserByID(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		repo.AssertExpectations(t)
	})
}

func TestUserService_SignInDev(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)
	mockLogger := new(MockLogger)
	authService := setupAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		AuthService:   authService,
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

	// Test successful sign in dev
	t.Run("existing user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// Set up expectations
		repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return s.UserID == user.ID && s.ExpiresAt.After(time.Now())
		})).Return(nil)

		token, err := service.SignInDev(ctx, SignInInput{Email: user.Email})
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("new user", func(t *testing.T) {
		repo.Mock = mock.Mock{}

		// User not found, should create a new one
		repo.On("GetUserByEmail", ctx, "new@example.com").Return(nil, &domain.ErrUserNotFound{Message: "user not found"})

		// Should create a new user
		repo.On("CreateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == "new@example.com" && !u.CreatedAt.IsZero() && !u.UpdatedAt.IsZero()
		})).Return(nil)

		// Should create a session for the new user
		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return s.UserID != "" && s.ExpiresAt.After(time.Now())
		})).Return(nil)

		token, err := service.SignInDev(ctx, SignInInput{Email: "new@example.com"})
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})
}
