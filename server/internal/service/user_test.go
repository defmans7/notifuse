package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
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

func TestUserService_SignIn(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
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

	err = service.SignIn(ctx, SignInInput{Email: "notfound@example.com"})
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_SignUp(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
	})
	require.NoError(t, err)

	ctx := context.Background()
	input := SignUpInput{
		Email: "test@example.com",
		Name:  "Test User",
	}

	t.Run("successful sign up", func(t *testing.T) {
		// Reset mocks before test
		repo.Mock = mock.Mock{}
		emailSender.Mock = mock.Mock{}

		// Test successful sign up
		repo.On("GetUserByEmail", ctx, input.Email).
			Return(nil, &domain.ErrUserNotFound{Message: "user not found"})
		repo.On("CreateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == input.Email && u.Name == input.Name
		})).Return(nil)
		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.Session) bool {
			return len(s.MagicCode) == 6 && !s.MagicCodeExpires.IsZero()
		})).Return(nil)
		emailSender.On("SendMagicCode", input.Email, mock.MatchedBy(func(code string) bool {
			return len(code) == 6
		})).Return(nil)

		err = service.SignUp(ctx, input)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		emailSender.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		// Reset mocks before test
		repo.Mock = mock.Mock{}
		emailSender.Mock = mock.Mock{}

		existingUser := &domain.User{
			ID:    uuid.New().String(),
			Email: input.Email,
			Name:  input.Name,
		}
		repo.On("GetUserByEmail", ctx, input.Email).Return(existingUser, nil)

		err = service.SignUp(ctx, input)
		assert.EqualError(t, err, "user already exists")
		repo.AssertExpectations(t)
		emailSender.AssertExpectations(t)
	})
}

func TestUserService_VerifyCode(t *testing.T) {
	repo := new(mockUserRepository)
	emailSender := new(mockEmailSender)

	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	service, err := NewUserService(UserServiceConfig{
		Repository:    repo,
		PrivateKey:    cfg.Security.PasetoPrivateKey,
		PublicKey:     cfg.Security.PasetoPublicKey,
		EmailSender:   emailSender,
		SessionExpiry: 24 * time.Hour,
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
		assert.Contains(t, err.Error(), "invalid or expired code")
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
		assert.Contains(t, err.Error(), "invalid or expired code")
		repo.AssertExpectations(t)
	})
}
