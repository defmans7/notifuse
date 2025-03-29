package service

import (
	"context"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockAuthService(t *testing.T) {
	mockService := new(MockAuthService)
	ctx := context.Background()

	// Test user for mock responses
	testUser := &domain.User{
		ID:    "test-user-id",
		Email: "test@example.com",
	}

	t.Run("AuthenticateUserFromContext", func(t *testing.T) {
		mockService.On("AuthenticateUserFromContext", ctx).Return(testUser, nil)

		user, err := mockService.AuthenticateUserFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		mockService.AssertExpectations(t)
	})

	t.Run("AuthenticateUserForWorkspace", func(t *testing.T) {
		workspaceID := "test-workspace"
		mockService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(testUser, nil)

		user, err := mockService.AuthenticateUserForWorkspace(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		mockService.AssertExpectations(t)
	})

	t.Run("VerifyUserSession", func(t *testing.T) {
		userID := "test-user-id"
		sessionID := "test-session"
		mockService.On("VerifyUserSession", ctx, userID, sessionID).Return(testUser, nil)

		user, err := mockService.VerifyUserSession(ctx, userID, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		mockService.AssertExpectations(t)
	})

	t.Run("GenerateAuthToken", func(t *testing.T) {
		sessionID := "test-session"
		expiresAt := time.Now().Add(24 * time.Hour)
		expectedToken := "test-token"

		mockService.On("GenerateAuthToken", testUser, sessionID, expiresAt).Return(expectedToken)

		token := mockService.GenerateAuthToken(testUser, sessionID, expiresAt)
		assert.Equal(t, expectedToken, token)
		mockService.AssertExpectations(t)
	})

	t.Run("GenerateInvitationToken", func(t *testing.T) {
		invitation := &domain.WorkspaceInvitation{
			ID: "test-invitation",
		}
		expectedToken := "test-invitation-token"

		mockService.On("GenerateInvitationToken", invitation).Return(expectedToken)

		token := mockService.GenerateInvitationToken(invitation)
		assert.Equal(t, expectedToken, token)
		mockService.AssertExpectations(t)
	})

	t.Run("GetPrivateKey", func(t *testing.T) {
		expectedKey := paseto.V4AsymmetricSecretKey{}
		mockService.On("GetPrivateKey").Return(expectedKey)

		key := mockService.GetPrivateKey()
		assert.Equal(t, expectedKey, key)
		mockService.AssertExpectations(t)
	})
}
