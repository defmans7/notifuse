package http

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/internal/service"
)

// AuthServiceMiddlewareAdapter adapts AuthService to implement middleware.AuthServiceInterface
type AuthServiceMiddlewareAdapter struct {
	AuthService *service.AuthService
}

// VerifyUserSession delegates the call to the underlying AuthService
func (a *AuthServiceMiddlewareAdapter) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	return a.AuthService.VerifyUserSession(ctx, userID, sessionID)
}

// GetUserByID delegates the call to the underlying AuthService
func (a *AuthServiceMiddlewareAdapter) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return a.AuthService.GetUserByID(ctx, userID)
}

// NewAuthServiceMiddlewareAdapter creates a new adapter for AuthService
func NewAuthServiceMiddlewareAdapter(authService *service.AuthService) *AuthServiceMiddlewareAdapter {
	return &AuthServiceMiddlewareAdapter{
		AuthService: authService,
	}
}

// Verify that AuthServiceMiddlewareAdapter implements middleware.AuthServiceInterface
var _ middleware.AuthServiceInterface = (*AuthServiceMiddlewareAdapter)(nil)

// UserServiceAdapter adapts AuthService to implement UserServiceInterface
type UserServiceAdapter struct {
	AuthService *service.AuthService
	UserService *service.UserService
}

// SignIn delegates the call to the underlying UserService
func (a *UserServiceAdapter) SignIn(ctx context.Context, input service.SignInInput) (string, error) {
	return a.UserService.SignIn(ctx, input)
}

// VerifyCode delegates the call to the underlying UserService
func (a *UserServiceAdapter) VerifyCode(ctx context.Context, input service.VerifyCodeInput) (*service.AuthResponse, error) {
	return a.UserService.VerifyCode(ctx, input)
}

// VerifyUserSession delegates the call to the underlying AuthService
func (a *UserServiceAdapter) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	return a.AuthService.VerifyUserSession(ctx, userID, sessionID)
}

// GetUserByID delegates the call to the underlying AuthService
func (a *UserServiceAdapter) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return a.AuthService.GetUserByID(ctx, userID)
}

// NewUserServiceAdapter creates a new adapter for UserService
func NewUserServiceAdapter(authService *service.AuthService, userService *service.UserService) *UserServiceAdapter {
	return &UserServiceAdapter{
		AuthService: authService,
		UserService: userService,
	}
}

// Verify that UserServiceAdapter implements UserServiceInterface
var _ UserServiceInterface = (*UserServiceAdapter)(nil)
