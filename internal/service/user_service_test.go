package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type mockEmailSender struct {
	shouldError bool
}

func (m *mockEmailSender) SendMagicCode(email, code string) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

func setupUserTest(t *testing.T) (
	*mocks.MockUserRepository,
	*mocks.MockAuthService,
	*mocks.MockLogger,
	*UserService,
	*mockEmailSender,
) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockSender := &mockEmailSender{}

	service, err := NewUserService(UserServiceConfig{
		Repository:    mockRepo,
		AuthService:   mockAuthService,
		EmailSender:   mockSender,
		SessionExpiry: 24 * time.Hour,
		Logger:        mockLogger,
		IsDevelopment: false,
	})
	require.NoError(t, err)

	return mockRepo, mockAuthService, mockLogger, service, mockSender
}

func TestUserService_SignIn(t *testing.T) {
	mockRepo, _, mockLogger, service, mockSender := setupUserTest(t)

	ctx := context.Background()
	email := "test@example.com"

	t.Run("successful sign in - existing user", func(t *testing.T) {
		mockSender.shouldError = true // Force email sending error for logging
		user := &domain.User{
			ID:    "user123",
			Email: email,
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		mockRepo.EXPECT().
			CreateSession(ctx, gomock.Any()).
			Return(nil)

		mockLogger.EXPECT().
			WithField("user_id", user.ID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("email", user.Email).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "mock error").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to send magic code")

		code, err := service.SignIn(ctx, domain.SignInInput{Email: email})
		require.Error(t, err)
		require.Equal(t, "mock error", err.Error())
		require.Empty(t, code)
	})

	t.Run("successful sign in - new user", func(t *testing.T) {
		mockSender.shouldError = true // Force email sending error for logging
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, &domain.ErrUserNotFound{})

		mockRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil)

		mockRepo.EXPECT().
			CreateSession(ctx, gomock.Any()).
			Return(nil)

		mockLogger.EXPECT().
			WithField("user_id", gomock.Any()).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("email", email).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "mock error").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to send magic code")

		code, err := service.SignIn(ctx, domain.SignInInput{Email: email})
		require.Error(t, err)
		require.Equal(t, "mock error", err.Error())
		require.Empty(t, code)
	})

	t.Run("development mode returns code directly", func(t *testing.T) {
		service.isDevelopment = true
		mockSender.shouldError = false // No email sending in dev mode
		user := &domain.User{
			ID:    "user123",
			Email: email,
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		mockRepo.EXPECT().
			CreateSession(ctx, gomock.Any()).
			Return(nil)

		code, err := service.SignIn(ctx, domain.SignInInput{Email: email})
		require.NoError(t, err)
		require.NotEmpty(t, code)
		require.Len(t, code, 6) // Should be 6 digits
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errors.New("db error"))

		mockLogger.EXPECT().
			WithField("email", email).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "db error").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to get user by email")

		code, err := service.SignIn(ctx, domain.SignInInput{Email: email})
		require.Error(t, err)
		require.Empty(t, code)
	})
}

func TestUserService_VerifyCode(t *testing.T) {
	mockRepo, mockAuthService, mockLogger, service, _ := setupUserTest(t)

	ctx := context.Background()
	email := "test@example.com"
	code := "123456"
	userID := "user123"

	t.Run("successful verification", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: email,
		}

		session := &domain.Session{
			ID:               "session123",
			UserID:           userID,
			MagicCode:        code,
			MagicCodeExpires: time.Now().Add(15 * time.Minute),
			ExpiresAt:        time.Now().Add(24 * time.Hour),
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		mockRepo.EXPECT().
			GetSessionsByUserID(ctx, userID).
			Return([]*domain.Session{session}, nil)

		mockRepo.EXPECT().
			UpdateSession(ctx, gomock.Any()).
			Return(nil)

		mockAuthService.EXPECT().
			GenerateAuthToken(user, session.ID, session.ExpiresAt).
			Return("token123")

		result, err := service.VerifyCode(ctx, domain.VerifyCodeInput{
			Email: email,
			Code:  code,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "token123", result.Token)
		require.Equal(t, *user, result.User)
		require.Equal(t, session.ExpiresAt, result.ExpiresAt)
	})

	t.Run("invalid code", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: email,
		}

		session := &domain.Session{
			ID:               "session123",
			UserID:           userID,
			MagicCode:        "654321",
			MagicCodeExpires: time.Now().Add(15 * time.Minute),
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		mockRepo.EXPECT().
			GetSessionsByUserID(ctx, userID).
			Return([]*domain.Session{session}, nil)

		mockLogger.EXPECT().
			WithField("user_id", userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("email", email).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Invalid magic code")

		result, err := service.VerifyCode(ctx, domain.VerifyCodeInput{
			Email: email,
			Code:  code,
		})

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "invalid magic code", err.Error())
	})

	t.Run("expired code", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: email,
		}

		session := &domain.Session{
			ID:               "session123",
			UserID:           userID,
			MagicCode:        code,
			MagicCodeExpires: time.Now().Add(-1 * time.Minute),
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		mockRepo.EXPECT().
			GetSessionsByUserID(ctx, userID).
			Return([]*domain.Session{session}, nil)

		mockLogger.EXPECT().
			WithField("user_id", userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("email", email).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("session_id", session.ID).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Magic code expired")

		result, err := service.VerifyCode(ctx, domain.VerifyCodeInput{
			Email: email,
			Code:  code,
		})

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "magic code expired", err.Error())
	})
}

func TestUserService_VerifyUserSession(t *testing.T) {
	mockRepo, _, mockLogger, service, _ := setupUserTest(t)

	ctx := context.Background()
	userID := "user123"
	sessionID := "session123"

	t.Run("successful verification", func(t *testing.T) {
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockRepo.EXPECT().
			GetSessionByID(ctx, sessionID).
			Return(session, nil)

		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		result, err := service.VerifyUserSession(ctx, userID, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
	})

	t.Run("session not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetSessionByID(ctx, sessionID).
			Return(nil, errors.New("session not found"))

		mockLogger.EXPECT().
			WithField("user_id", userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("session_id", sessionID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "session not found").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to get session by ID")

		result, err := service.VerifyUserSession(ctx, userID, sessionID)

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("session expired", func(t *testing.T) {
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		mockRepo.EXPECT().
			GetSessionByID(ctx, sessionID).
			Return(session, nil)

		mockLogger.EXPECT().
			WithField("user_id", userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("session_id", sessionID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("expires_at", session.ExpiresAt).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Session expired")

		result, err := service.VerifyUserSession(ctx, userID, sessionID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, ErrSessionExpired, err)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	mockRepo, _, mockLogger, service, _ := setupUserTest(t)

	ctx := context.Background()
	userID := "user123"

	t.Run("successful retrieval", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		result, err := service.GetUserByID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("user not found"))

		mockLogger.EXPECT().
			WithField("user_id", userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "user not found").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to get user by ID")

		result, err := service.GetUserByID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	mockRepo, _, mockLogger, service, _ := setupUserTest(t)

	ctx := context.Background()
	email := "test@example.com"

	t.Run("successful retrieval", func(t *testing.T) {
		user := &domain.User{
			ID:    "user123",
			Email: email,
		}

		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(user, nil)

		result, err := service.GetUserByEmail(ctx, email)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, email, result.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errors.New("user not found"))

		mockLogger.EXPECT().
			WithField("email", email).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("error", "user not found").
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to get user by email")

		result, err := service.GetUserByEmail(ctx, email)

		require.Error(t, err)
		require.Nil(t, result)
	})
}
