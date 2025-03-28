package service

import (
	"context"
	"database/sql"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockContactRepository is a mock implementation of domain.ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) GetContactByEmail(ctx context.Context, email string, workspaceID string) (*domain.Contact, error) {
	args := m.Called(ctx, email, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {
	args := m.Called(ctx, workspaceID, contacts)
	return args.Error(0)
}

func (m *MockContactRepository) DeleteContact(ctx context.Context, email string, workspaceID string) error {
	args := m.Called(ctx, email, workspaceID)
	return args.Error(0)
}

func (m *MockContactRepository) GetContactByExternalID(ctx context.Context, externalID string, workspaceID string) (*domain.Contact, error) {
	args := m.Called(ctx, externalID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GetContactsResponse), args.Error(1)
}

func (m *MockContactRepository) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
	args := m.Called(ctx, workspaceID, contact)
	return args.Bool(0), args.Error(1)
}

// MockWorkspaceRepository is a mock implementation of domain.WorkspaceRepository
type MockWorkspaceRepository struct {
	mock.Mock
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.DB), args.Error(1)
}

func (m *MockWorkspaceRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) DeleteDatabase(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	args := m.Called(ctx, userWorkspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	args := m.Called(ctx, userID, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspace), args.Error(1)
}

func (m *MockWorkspaceRepository) GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspaceWithEmail), args.Error(1)
}

func (m *MockWorkspaceRepository) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	args := m.Called(ctx, userID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWorkspace), args.Error(1)
}

func (m *MockWorkspaceRepository) CreateInvitation(ctx context.Context, invitation *domain.WorkspaceInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetInvitationByID(ctx context.Context, id string) (*domain.WorkspaceInvitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.Error(1)
}

func (m *MockWorkspaceRepository) GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.Error(1)
}

func (m *MockWorkspaceRepository) IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	args := m.Called(ctx, userID, workspaceID)
	return args.Bool(0), args.Error(1)
}

// MockAuthService is a mock implementation of domain.AuthService
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

// MockContactListRepository is a mock implementation of domain.ContactListRepository
type MockContactListRepository struct {
	mock.Mock
}

func (m *MockContactListRepository) AddContactToList(ctx context.Context, workspaceID string, contactList *domain.ContactList) error {
	args := m.Called(ctx, workspaceID, contactList)
	return args.Error(0)
}

func (m *MockContactListRepository) GetContactListByIDs(ctx context.Context, workspaceID string, email, listID string) (*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, email, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetContactsByListID(ctx context.Context, workspaceID string, listID string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetListsByEmail(ctx context.Context, workspaceID string, email string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) UpdateContactListStatus(ctx context.Context, workspaceID string, email, listID string, status domain.ContactListStatus) error {
	args := m.Called(ctx, workspaceID, email, listID, status)
	return args.Error(0)
}

func (m *MockContactListRepository) RemoveContactFromList(ctx context.Context, workspaceID string, email, listID string) error {
	args := m.Called(ctx, workspaceID, email, listID)
	return args.Error(0)
}
