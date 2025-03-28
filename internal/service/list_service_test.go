package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {
	args := m.Called(ctx, workspaceID, list)
	return args.Error(0)
}

func (m *MockListRepository) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
	args := m.Called(ctx, workspaceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.List), args.Error(1)
}

func (m *MockListRepository) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.List), args.Error(1)
}

func (m *MockListRepository) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	args := m.Called(ctx, workspaceID, list)
	return args.Error(0)
}

func (m *MockListRepository) DeleteList(ctx context.Context, workspaceID string, id string) error {
	args := m.Called(ctx, workspaceID, id)
	return args.Error(0)
}

func TestListService_CreateList(t *testing.T) {
	mockRepo := new(MockListRepository)
	mockAuthService := new(MockAuthService)
	mockLogger := new(MockLogger)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	list := &domain.List{
		ID:            "testlist123",
		Name:          "Test List",
		Type:          "private",
		IsDoubleOptin: true,
		IsPublic:      false,
		Description:   "Test Description",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("CreateList", ctx, workspaceID, list).Return(nil)

		err := service.CreateList(ctx, workspaceID, list)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.CreateList(ctx, workspaceID, list)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		invalidList := &domain.List{
			ID:   "testlist123",
			Name: "Test List",
			Type: "invalid",
		}
		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.CreateList(ctx, workspaceID, invalidList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("CreateList", ctx, workspaceID, list).Return(errors.New("repo error"))
		mockLogger.On("WithField", "list_id", list.ID).Return(mockLogger)
		mockLogger.On("Error", "Failed to create list: repo error").Return()

		err := service.CreateList(ctx, workspaceID, list)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListService_GetListByID(t *testing.T) {
	mockRepo := new(MockListRepository)
	mockAuthService := new(MockAuthService)
	mockLogger := new(MockLogger)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "testlist123"
	list := &domain.List{
		ID:            listID,
		Name:          "Test List",
		Type:          "private",
		IsDoubleOptin: true,
		IsPublic:      false,
		Description:   "Test Description",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetListByID", ctx, workspaceID, listID).Return(list, nil)

		result, err := service.GetListByID(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.Equal(t, list, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("list not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetListByID", ctx, workspaceID, listID).Return(nil, &domain.ErrListNotFound{})

		result, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetListByID", ctx, workspaceID, listID).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to get list: repo error").Return()

		result, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListService_GetLists(t *testing.T) {
	mockRepo := new(MockListRepository)
	mockAuthService := new(MockAuthService)
	mockLogger := new(MockLogger)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	lists := []*domain.List{
		{
			ID:            "testlist123",
			Name:          "Test List 1",
			Type:          "private",
			IsDoubleOptin: true,
			IsPublic:      false,
			Description:   "Test Description 1",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            "testlist456",
			Name:          "Test List 2",
			Type:          "private",
			IsDoubleOptin: false,
			IsPublic:      true,
			Description:   "Test Description 2",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetLists", ctx, workspaceID).Return(lists, nil)

		result, err := service.GetLists(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetLists(ctx, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetLists", ctx, workspaceID).Return(nil, errors.New("repo error"))
		mockLogger.On("Error", "Failed to get lists: repo error").Return()

		result, err := service.GetLists(ctx, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListService_UpdateList(t *testing.T) {
	mockRepo := new(MockListRepository)
	mockAuthService := new(MockAuthService)
	mockLogger := new(MockLogger)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	list := &domain.List{
		ID:            "testlist123",
		Name:          "Test List",
		Type:          "private",
		IsDoubleOptin: true,
		IsPublic:      false,
		Description:   "Test Description",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	t.Run("successful update", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("UpdateList", ctx, workspaceID, list).Return(nil)

		err := service.UpdateList(ctx, workspaceID, list)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.UpdateList(ctx, workspaceID, list)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		invalidList := &domain.List{
			ID:   "testlist123",
			Name: "Test List",
			Type: "invalid",
		}
		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.UpdateList(ctx, workspaceID, invalidList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("UpdateList", ctx, workspaceID, list).Return(errors.New("repo error"))
		mockLogger.On("WithField", "list_id", list.ID).Return(mockLogger)
		mockLogger.On("Error", "Failed to update list: repo error").Return()

		err := service.UpdateList(ctx, workspaceID, list)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListService_DeleteList(t *testing.T) {
	mockRepo := new(MockListRepository)
	mockAuthService := new(MockAuthService)
	mockLogger := new(MockLogger)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "testlist123"

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("DeleteList", ctx, workspaceID, listID).Return(nil)

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("DeleteList", ctx, workspaceID, listID).Return(errors.New("repo error"))
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to delete list: repo error").Return()

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
