package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupAuthTest(t *testing.T) (
	*mocks.MockAuthRepository,
	*mocks.MockWorkspaceRepository,
	*mocks.MockLogger,
	*AuthService,
) {
	ctrl := gomock.NewController(t)
	mockAuthRepo := mocks.NewMockAuthRepository(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	// Generate test keys
	privateKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := privateKey.Public()

	service, err := NewAuthService(AuthServiceConfig{
		Repository:          mockAuthRepo,
		WorkspaceRepository: mockWorkspaceRepo,
		PrivateKey:          privateKey.ExportBytes(),
		PublicKey:           publicKey.ExportBytes(),
		Logger:              mockLogger,
	})
	require.NoError(t, err)

	return mockAuthRepo, mockWorkspaceRepo, mockLogger, service
}

func TestAuthService_AuthenticateUserFromContext(t *testing.T) {
	mockAuthRepo, _, _, service := setupAuthTest(t)

	userID := "user123"
	sessionID := "session123"

	t.Run("successful authentication", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		ctx := context.WithValue(context.WithValue(context.Background(), domain.UserIDKey, userID), domain.SessionIDKey, sessionID)

		mockAuthRepo.EXPECT().
			GetSessionByID(ctx, sessionID, userID).
			Return(&expiresAt, nil)

		mockAuthRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		result, err := service.AuthenticateUserFromContext(ctx)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
	})

	t.Run("missing user_id in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.SessionIDKey, sessionID)

		result, err := service.AuthenticateUserFromContext(ctx)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, result)
	})

	t.Run("missing session_id in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.UserIDKey, userID)

		result, err := service.AuthenticateUserFromContext(ctx)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, result)
	})
}

func TestAuthService_AuthenticateUserForWorkspace(t *testing.T) {
	mockAuthRepo, mockWorkspaceRepo, _, service := setupAuthTest(t)

	userID := "user123"
	sessionID := "session123"
	workspaceID := "workspace123"

	t.Run("successful authentication", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		ctx := context.WithValue(context.WithValue(context.Background(), domain.UserIDKey, userID), domain.SessionIDKey, sessionID)

		mockAuthRepo.EXPECT().
			GetSessionByID(ctx, sessionID, userID).
			Return(&expiresAt, nil)

		mockAuthRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		mockWorkspaceRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      userID,
				WorkspaceID: workspaceID,
				Role:        "member",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil)

		newCtx, result, err := service.AuthenticateUserForWorkspace(ctx, workspaceID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)

		// Verify that the user is stored in the context
		storedUser, ok := newCtx.Value(domain.WorkspaceUserKey(workspaceID)).(*domain.User)
		require.True(t, ok)
		require.Equal(t, userID, storedUser.ID)
	})

	t.Run("user already in context", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		// Create a context with the user already stored for this workspace
		ctx := context.WithValue(context.Background(), domain.WorkspaceUserKey(workspaceID), user)

		// No mock expectations should be called since the user is already in context

		newCtx, result, err := service.AuthenticateUserForWorkspace(ctx, workspaceID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
		require.Equal(t, ctx, newCtx) // Context should be unchanged
	})

	t.Run("user not in workspace", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		ctx := context.WithValue(context.WithValue(context.Background(), domain.UserIDKey, userID), domain.SessionIDKey, sessionID)

		mockAuthRepo.EXPECT().
			GetSessionByID(ctx, sessionID, userID).
			Return(&expiresAt, nil)

		mockAuthRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		mockWorkspaceRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(nil, errors.New("not found"))

		newCtx, result, err := service.AuthenticateUserForWorkspace(ctx, workspaceID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, ctx, newCtx) // Context should be unchanged on error
	})
}

func TestAuthService_VerifyUserSession(t *testing.T) {
	mockAuthRepo, _, mockLogger, service := setupAuthTest(t)

	userID := "user123"
	sessionID := "session123"

	t.Run("successful verification", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		mockAuthRepo.EXPECT().
			GetSessionByID(context.Background(), sessionID, userID).
			Return(&expiresAt, nil)

		mockAuthRepo.EXPECT().
			GetUserByID(context.Background(), userID).
			Return(user, nil)

		result, err := service.VerifyUserSession(context.Background(), userID, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
	})

	t.Run("session not found", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetSessionByID(context.Background(), sessionID, userID).
			Return(nil, sql.ErrNoRows)

		mockLogger.EXPECT().
			WithField(string(domain.UserIDKey), userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField(string(domain.SessionIDKey), sessionID).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Session not found")

		result, err := service.VerifyUserSession(context.Background(), userID, sessionID)

		require.Error(t, err)
		require.Equal(t, ErrSessionExpired, err)
		require.Nil(t, result)
	})

	t.Run("session expired", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)

		mockAuthRepo.EXPECT().
			GetSessionByID(context.Background(), sessionID, userID).
			Return(&expiresAt, nil)

		mockLogger.EXPECT().
			WithField(string(domain.UserIDKey), userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField(string(domain.SessionIDKey), sessionID).
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField("expires_at", &expiresAt).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Session expired")

		result, err := service.VerifyUserSession(context.Background(), userID, sessionID)

		require.Error(t, err)
		require.Equal(t, ErrSessionExpired, err)
		require.Nil(t, result)
	})

	t.Run("user not found", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)

		mockAuthRepo.EXPECT().
			GetSessionByID(context.Background(), sessionID, userID).
			Return(&expiresAt, nil)

		mockAuthRepo.EXPECT().
			GetUserByID(context.Background(), userID).
			Return(nil, sql.ErrNoRows)

		mockLogger.EXPECT().
			WithField(string(domain.UserIDKey), userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("User not found")

		result, err := service.VerifyUserSession(context.Background(), userID, sessionID)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, result)
	})
}

func TestAuthService_GenerateAuthToken(t *testing.T) {
	mockAuthRepo, _, _, service := setupAuthTest(t)

	userID := "user123"
	sessionID := "session123"
	expiresAt := time.Now().Add(1 * time.Hour)

	t.Run("successful token generation", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		token := service.GenerateAuthToken(user, sessionID, expiresAt)

		require.NotEmpty(t, token)
		require.NotNil(t, token)
	})

	t.Run("failed token generation", func(t *testing.T) {
		// Create a service with invalid key length
		_, err := NewAuthService(AuthServiceConfig{
			Repository:          mockAuthRepo,
			WorkspaceRepository: nil,
			PrivateKey:          []byte("invalid"),
			PublicKey:           []byte("invalid"),
			Logger:              nil,
		})
		require.Error(t, err)
	})
}

func TestAuthService_GenerateInvitationToken(t *testing.T) {
	mockAuthRepo, _, _, service := setupAuthTest(t)

	invitationID := "invitation123"
	workspaceID := "workspace123"
	inviterID := "inviter123"
	email := "test@example.com"

	t.Run("successful token generation", func(t *testing.T) {
		invitation := &domain.WorkspaceInvitation{
			ID:          invitationID,
			WorkspaceID: workspaceID,
			InviterID:   inviterID,
			Email:       email,
			ExpiresAt:   time.Now().Add(15 * 24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		token := service.GenerateInvitationToken(invitation)

		require.NotEmpty(t, token)
		require.NotNil(t, token)
	})

	t.Run("failed token generation", func(t *testing.T) {
		// Create a service with invalid key length
		_, err := NewAuthService(AuthServiceConfig{
			Repository:          mockAuthRepo,
			WorkspaceRepository: nil,
			PrivateKey:          []byte("invalid"),
			PublicKey:           []byte("invalid"),
			Logger:              nil,
		})
		require.Error(t, err)
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	mockAuthRepo, _, mockLogger, service := setupAuthTest(t)

	userID := "user123"

	t.Run("successful user retrieval", func(t *testing.T) {
		user := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockAuthRepo.EXPECT().
			GetUserByID(context.Background(), userID).
			Return(user, nil)

		result, err := service.GetUserByID(context.Background(), userID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.ID)
	})

	t.Run("user not found", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetUserByID(context.Background(), userID).
			Return(nil, sql.ErrNoRows)

		result, err := service.GetUserByID(context.Background(), userID)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, result)
	})

	t.Run("error retrieving user", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetUserByID(context.Background(), userID).
			Return(nil, errors.New("database error"))

		mockLogger.EXPECT().
			WithField("error", "database error").
			Return(mockLogger)

		mockLogger.EXPECT().
			WithField(string(domain.UserIDKey), userID).
			Return(mockLogger)

		mockLogger.EXPECT().
			Error("Failed to get user by ID")

		result, err := service.GetUserByID(context.Background(), userID)

		require.Error(t, err)
		require.Nil(t, result)
	})
}
