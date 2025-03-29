package service

import (
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockMailer mocks the mailer.Mailer interface
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendInvitationEmail(invitation *domain.WorkspaceInvitation) error {
	args := m.Called(invitation)
	return args.Error(0)
}

func (m *MockMailer) SendMagicCode(email, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
}

func (m *MockMailer) SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error {
	args := m.Called(email, workspaceName, inviterName, token)
	return args.Error(0)
}
