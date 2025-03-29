package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockUserService(t *testing.T) {
	mockService := &MockUserService{}
	ctx := context.Background()

	t.Run("SignIn", func(t *testing.T) {
		// Test success case
		input := domain.SignInInput{
			Email: "test@example.com",
		}
		expectedCode := "123456"
		mockService.On("SignIn", ctx, input).Return(expectedCode, nil).Once()

		code, err := mockService.SignIn(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedCode, code)

		// Test error case
		mockService.On("SignIn", ctx, input).Return("", fmt.Errorf("invalid input")).Once()

		code, err = mockService.SignIn(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "", code)
	})

	t.Run("VerifyCode", func(t *testing.T) {
		// Test success case
		input := domain.VerifyCodeInput{
			Email: "test@example.com",
			Code:  "123456",
		}
		expectedResponse := &domain.AuthResponse{
			Token: "test-token",
			User: domain.User{
				Email: "test@example.com",
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		mockService.On("VerifyCode", ctx, input).Return(expectedResponse, nil).Once()

		response, err := mockService.VerifyCode(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		// Test error case
		mockService.On("VerifyCode", ctx, input).Return(nil, fmt.Errorf("invalid code")).Once()

		response, err = mockService.VerifyCode(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("VerifyUserSession", func(t *testing.T) {
		// Test success case
		userID := "user-123"
		sessionID := "session-123"
		expectedUser := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}
		mockService.On("VerifyUserSession", ctx, userID, sessionID).Return(expectedUser, nil).Once()

		user, err := mockService.VerifyUserSession(ctx, userID, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)

		// Test error case
		mockService.On("VerifyUserSession", ctx, userID, sessionID).Return(nil, &domain.ErrSessionNotFound{Message: "session not found"}).Once()

		user, err = mockService.VerifyUserSession(ctx, userID, sessionID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Test success case
		userID := "user-123"
		expectedUser := &domain.User{
			ID:    userID,
			Email: "test@example.com",
		}
		mockService.On("GetUserByID", ctx, userID).Return(expectedUser, nil).Once()

		user, err := mockService.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)

		// Test error case
		mockService.On("GetUserByID", ctx, userID).Return(nil, &domain.ErrUserNotFound{Message: "user not found"}).Once()

		user, err = mockService.GetUserByID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		// Test success case
		email := "test@example.com"
		expectedUser := &domain.User{
			Email: email,
		}
		mockService.On("GetUserByEmail", ctx, email).Return(expectedUser, nil).Once()

		user, err := mockService.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)

		// Test error case
		mockService.On("GetUserByEmail", ctx, email).Return(nil, &domain.ErrUserNotFound{Message: "user not found"}).Once()

		user, err = mockService.GetUserByEmail(ctx, email)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	// Verify that all expected mock calls were made
	mockService.AssertExpectations(t)
}
