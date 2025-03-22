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

// Test IDs that follow alphanum validation
const (
	testListID1 = "test123"
	testListID2 = "list456"
)

type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) CreateList(ctx context.Context, list *domain.List) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockListRepository) GetListByID(ctx context.Context, id string) (*domain.List, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.List), args.Error(1)
}

func (m *MockListRepository) GetLists(ctx context.Context) ([]*domain.List, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.List), args.Error(1)
}

func (m *MockListRepository) UpdateList(ctx context.Context, list *domain.List) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockListRepository) DeleteList(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestListService_CreateList(t *testing.T) {
	t.Run("should create list successfully", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID:            testListID1,
			Name:          "Test List",
			Description:   "This is a test list",
			Type:          "public",
			IsDoubleOptin: false,
		}

		mockRepo.On("CreateList", ctx, mock.MatchedBy(func(l *domain.List) bool {
			// Verify timestamps are set
			return l.ID == list.ID && l.Name == list.Name &&
				!l.CreatedAt.IsZero() && !l.UpdatedAt.IsZero()
		})).Return(nil).Once()

		// Act
		err := service.CreateList(ctx, list)

		// Assert
		assert.NoError(t, err)
		assert.False(t, list.CreatedAt.IsZero())
		assert.False(t, list.UpdatedAt.IsZero())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should generate ID if not provided", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()

		// ID must be provided since there's no ID generation in the service
		list := &domain.List{
			ID:            testListID1, // Use a valid pre-defined ID
			Name:          "Test List",
			Description:   "This is a test list",
			Type:          "public",
			IsDoubleOptin: false,
		}

		mockRepo.On("CreateList", ctx, mock.MatchedBy(func(l *domain.List) bool {
			return l.ID == testListID1 && l.Name == list.Name
		})).Return(nil).Once()

		// Act
		err := service.CreateList(ctx, list)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, testListID1, list.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if list is invalid", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID: "test-id", // Invalid ID with hyphen
			// Name missing, which is required
		}

		// Act
		err := service.CreateList(ctx, list)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "CreateList")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID:            testListID1,
			Name:          "Test List",
			Description:   "This is a test list",
			Type:          "public",
			IsDoubleOptin: false,
		}

		repoErr := errors.New("repository error")
		mockRepo.On("CreateList", ctx, mock.MatchedBy(func(l *domain.List) bool {
			return l.ID == list.ID && l.Name == list.Name
		})).Return(repoErr).Once()

		// Act
		err := service.CreateList(ctx, list)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create list")
		mockRepo.AssertExpectations(t)
	})
}

func TestListService_GetListByID(t *testing.T) {
	t.Run("should get list by ID successfully", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		id := testListID1
		expectedList := &domain.List{
			ID:            id,
			Name:          "Test List",
			Description:   "This is a test list",
			Type:          "private",
			IsDoubleOptin: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mockRepo.On("GetListByID", ctx, id).Return(expectedList, nil).Once()

		// Act
		list, err := service.GetListByID(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedList, list)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		id := "nonexistent"
		notFoundErr := &domain.ErrListNotFound{Message: "list not found"}

		mockRepo.On("GetListByID", ctx, id).Return(nil, notFoundErr).Once()

		// Act
		list, err := service.GetListByID(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.IsType(t, &domain.ErrListNotFound{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		id := testListID1
		repoErr := errors.New("repository error")

		mockRepo.On("GetListByID", ctx, id).Return(nil, repoErr).Once()

		// Act
		list, err := service.GetListByID(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to get list")
		mockRepo.AssertExpectations(t)
	})
}

func TestListService_GetLists(t *testing.T) {
	t.Run("should get all lists successfully", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		expectedLists := []*domain.List{
			{
				ID:            testListID1,
				Name:          "List 1",
				Description:   "First list",
				Type:          "public",
				IsDoubleOptin: false,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			{
				ID:            testListID2,
				Name:          "List 2",
				Description:   "Second list",
				Type:          "private",
				IsDoubleOptin: true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		mockRepo.On("GetLists", ctx).Return(expectedLists, nil).Once()

		// Act
		lists, err := service.GetLists(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedLists, lists)
		assert.Len(t, lists, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return empty slice when no lists", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		expectedLists := []*domain.List{}

		mockRepo.On("GetLists", ctx).Return(expectedLists, nil).Once()

		// Act
		lists, err := service.GetLists(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, lists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		repoErr := errors.New("repository error")

		mockRepo.On("GetLists", ctx).Return(nil, repoErr).Once()

		// Act
		lists, err := service.GetLists(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, lists)
		assert.Contains(t, err.Error(), "failed to get lists")
		mockRepo.AssertExpectations(t)
	})
}

func TestListService_UpdateList(t *testing.T) {
	t.Run("should update list successfully", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID:            testListID1,
			Name:          "Updated List",
			Description:   "This list was updated",
			Type:          "private",
			IsDoubleOptin: true,
			CreatedAt:     time.Now().Add(-24 * time.Hour), // Created a day ago
		}

		mockRepo.On("UpdateList", ctx, mock.MatchedBy(func(l *domain.List) bool {
			// Verify UpdatedAt is set
			return l.ID == list.ID && l.Name == list.Name &&
				!l.UpdatedAt.IsZero() && l.UpdatedAt.After(l.CreatedAt)
		})).Return(nil).Once()

		// Act
		err := service.UpdateList(ctx, list)

		// Assert
		assert.NoError(t, err)
		assert.False(t, list.UpdatedAt.IsZero())
		assert.True(t, list.UpdatedAt.After(list.CreatedAt))
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if list is invalid", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID: "test-id", // Invalid ID with hyphen
			// Name missing, which is required
		}

		// Act
		err := service.UpdateList(ctx, list)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "UpdateList")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		list := &domain.List{
			ID:            testListID1,
			Name:          "Test List",
			Description:   "This is a test list",
			Type:          "public",
			IsDoubleOptin: false,
		}

		repoErr := errors.New("repository error")
		mockRepo.On("UpdateList", ctx, mock.MatchedBy(func(l *domain.List) bool {
			return l.ID == list.ID && l.Name == list.Name
		})).Return(repoErr).Once()

		// Act
		err := service.UpdateList(ctx, list)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update list")
		mockRepo.AssertExpectations(t)
	})
}

func TestListService_DeleteList(t *testing.T) {
	t.Run("should delete list successfully", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		id := testListID1

		mockRepo.On("DeleteList", ctx, id).Return(nil).Once()

		// Act
		err := service.DeleteList(ctx, id)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewListService(mockRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		id := testListID1
		repoErr := errors.New("repository error")

		mockRepo.On("DeleteList", ctx, id).Return(repoErr).Once()

		// Act
		err := service.DeleteList(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete list")
		mockRepo.AssertExpectations(t)
	})
}
