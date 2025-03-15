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

func (m *mockUserRepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendMagicLink(email, token string) error {
	args := m.Called(email, token)
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
	emailSender.On("SendMagicLink", user.Email, mock.AnythingOfType("string")).Return(nil)

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
		emailSender.On("SendMagicLink", input.Email, mock.AnythingOfType("string")).Return(nil)

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

func TestUserService_VerifyToken(t *testing.T) {
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

	// Generate a valid token
	token := service.generateMagicLinkToken(user.Email)

	// Test successful verification
	repo.On("GetUserByEmail", ctx, user.Email).Return(user, nil)
	repo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)

	response, err := service.VerifyToken(ctx, VerifyTokenInput{Token: token})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.Name, response.User.Name)
	assert.NotEmpty(t, response.Token)
	repo.AssertExpectations(t)

	// Test invalid token
	response, err = service.VerifyToken(ctx, VerifyTokenInput{Token: "invalid-token"})
	assert.Error(t, err)
	assert.Nil(t, response)
}
