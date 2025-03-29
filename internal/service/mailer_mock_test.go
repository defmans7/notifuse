package service

import (
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockMailer(t *testing.T) {
	mockMailer := &MockMailer{}

	t.Run("SendInvitationEmail", func(t *testing.T) {
		invitation := &domain.WorkspaceInvitation{
			ID:          "test-invitation",
			WorkspaceID: "test-workspace",
			InviterID:   "test-inviter",
			Email:       "test@example.com",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Test success case
		mockMailer.On("SendInvitationEmail", invitation).Return(nil).Once()
		err := mockMailer.SendInvitationEmail(invitation)
		assert.NoError(t, err)
		mockMailer.AssertExpectations(t)

		// Test error case
		testError := errors.New("failed to send invitation")
		mockMailer.On("SendInvitationEmail", invitation).Return(testError).Once()
		err = mockMailer.SendInvitationEmail(invitation)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
		mockMailer.AssertExpectations(t)
	})

	t.Run("SendMagicCode", func(t *testing.T) {
		email := "test@example.com"
		code := "123456"

		// Test success case
		mockMailer.On("SendMagicCode", email, code).Return(nil).Once()
		err := mockMailer.SendMagicCode(email, code)
		assert.NoError(t, err)
		mockMailer.AssertExpectations(t)

		// Test error case
		testError := errors.New("failed to send magic code")
		mockMailer.On("SendMagicCode", email, code).Return(testError).Once()
		err = mockMailer.SendMagicCode(email, code)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
		mockMailer.AssertExpectations(t)
	})

	t.Run("SendWorkspaceInvitation", func(t *testing.T) {
		email := "test@example.com"
		workspaceName := "Test Workspace"
		inviterName := "John Doe"
		token := "test-token"

		// Test success case
		mockMailer.On("SendWorkspaceInvitation", email, workspaceName, inviterName, token).Return(nil).Once()
		err := mockMailer.SendWorkspaceInvitation(email, workspaceName, inviterName, token)
		assert.NoError(t, err)
		mockMailer.AssertExpectations(t)

		// Test error case
		testError := errors.New("failed to send workspace invitation")
		mockMailer.On("SendWorkspaceInvitation", email, workspaceName, inviterName, token).Return(testError).Once()
		err = mockMailer.SendWorkspaceInvitation(email, workspaceName, inviterName, token)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
		mockMailer.AssertExpectations(t)
	})
}
