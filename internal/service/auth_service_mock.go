package service

import (
	"context"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthenticateUserFromContext(ctx context.Context) (*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) AuthenticateUserForWorkspace(ctx context.Context, workspaceID string) (*domain.User, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GenerateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
	args := m.Called(user, sessionID, expiresAt)
	return args.String(0)
}

func (m *MockAuthService) GenerateInvitationToken(invitation *domain.WorkspaceInvitation) string {
	args := m.Called(invitation)
	return args.String(0)
}

func (m *MockAuthService) GetPrivateKey() paseto.V4AsymmetricSecretKey {
	args := m.Called()
	return args.Get(0).(paseto.V4AsymmetricSecretKey)
}
