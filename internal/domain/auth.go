package domain

import (
	"context"
	"time"

	"aidanwoods.dev/go-paseto"
)

// AuthRepository defines the interface for auth-related database operations
type AuthRepository interface {
	GetSessionByID(ctx context.Context, sessionID string, userID string) (*time.Time, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
}

type AuthService interface {
	AuthenticateUserFromContext(ctx context.Context) (*User, error)
	AuthenticateUserForWorkspace(ctx context.Context, workspaceID string) (*User, error)
	VerifyUserSession(ctx context.Context, userID, sessionID string) (*User, error)
	GenerateAuthToken(user *User, sessionID string, expiresAt time.Time) string
	GetPrivateKey() paseto.V4AsymmetricSecretKey
	GenerateInvitationToken(invitation *WorkspaceInvitation) string
}
