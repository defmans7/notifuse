package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestListService_CreateList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	list := &domain.List{
		ID:   "list123",
		Name: "Test List",
	}

	t.Run("successful create", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).Return(nil)
		mockLogger.EXPECT().WithField("list_id", list.ID).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		err := service.CreateList(ctx, workspaceID, list)
		assert.NoError(t, err)
		assert.NotZero(t, list.CreatedAt)
		assert.NotZero(t, list.UpdatedAt)
	})

	t.Run("create with template fields", func(t *testing.T) {
		listWithTemplates := &domain.List{
			ID:            "list123",
			Name:          "List With Templates",
			IsDoubleOptin: true,
			DoubleOptInTemplate: &domain.TemplateReference{
				ID:      "template123",
				Version: 1,
			},
			WelcomeTemplate: &domain.TemplateReference{
				ID:      "welcome123",
				Version: 2,
			},
			UnsubscribeTemplate: &domain.TemplateReference{
				ID:      "unsub123",
				Version: 3,
			},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, wsID string, l *domain.List) error {
				assert.Equal(t, "list123", l.ID)
				assert.Equal(t, "template123", l.DoubleOptInTemplate.ID)
				assert.Equal(t, int64(1), l.DoubleOptInTemplate.Version)
				assert.Equal(t, "welcome123", l.WelcomeTemplate.ID)
				assert.Equal(t, int64(2), l.WelcomeTemplate.Version)
				assert.Equal(t, "unsub123", l.UnsubscribeTemplate.ID)
				assert.Equal(t, int64(3), l.UnsubscribeTemplate.Version)
				return nil
			})

		err := service.CreateList(ctx, workspaceID, listWithTemplates)
		assert.NoError(t, err)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.CreateList(ctx, workspaceID, list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("validation failure", func(t *testing.T) {
		invalidList := &domain.List{} // Missing required fields
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.CreateList(ctx, workspaceID, invalidList)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid list")
	})

	t.Run("repository failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).Return(errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", list.ID).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		err := service.CreateList(ctx, workspaceID, list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create list")
	})
}

func TestListService_GetListByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "list123"
	expectedList := &domain.List{
		ID:   listID,
		Name: "Test List",
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetListByID(ctx, workspaceID, listID).Return(expectedList, nil)
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		list, err := service.GetListByID(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.Equal(t, expectedList, list)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		list, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("list not found", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetListByID(ctx, workspaceID, listID).Return(nil, &domain.ErrListNotFound{})
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger).Times(0)

		list, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, list)
		var notFoundErr *domain.ErrListNotFound
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetListByID(ctx, workspaceID, listID).Return(nil, errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		list, err := service.GetListByID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to get list")
	})
}

func TestListService_GetLists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	expectedLists := []*domain.List{
		{ID: "list1", Name: "List 1"},
		{ID: "list2", Name: "List 2"},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetLists(ctx, workspaceID).Return(expectedLists, nil)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		lists, err := service.GetLists(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, expectedLists, lists)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		lists, err := service.GetLists(ctx, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, lists)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetLists(ctx, workspaceID).Return(nil, errors.New("db error"))
		mockLogger.EXPECT().Error(gomock.Any())

		lists, err := service.GetLists(ctx, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, lists)
		assert.Contains(t, err.Error(), "failed to get lists")
	})
}

func TestListService_UpdateList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	list := &domain.List{
		ID:   "list123",
		Name: "Updated List",
	}

	t.Run("successful update", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpdateList(ctx, workspaceID, gomock.Any()).Return(nil)
		mockLogger.EXPECT().WithField("list_id", list.ID).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		err := service.UpdateList(ctx, workspaceID, list)
		assert.NoError(t, err)
		assert.NotZero(t, list.UpdatedAt)
	})

	t.Run("update with template fields", func(t *testing.T) {
		listWithTemplates := &domain.List{
			ID:            "list123",
			Name:          "Updated List",
			IsDoubleOptin: true,
			DoubleOptInTemplate: &domain.TemplateReference{
				ID:      "template123",
				Version: 1,
			},
			WelcomeTemplate: &domain.TemplateReference{
				ID:      "welcome123",
				Version: 2,
			},
			UnsubscribeTemplate: &domain.TemplateReference{
				ID:      "unsub123",
				Version: 3,
			},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpdateList(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, wsID string, l *domain.List) error {
				assert.Equal(t, "list123", l.ID)
				assert.Equal(t, "template123", l.DoubleOptInTemplate.ID)
				assert.Equal(t, int64(1), l.DoubleOptInTemplate.Version)
				assert.Equal(t, "welcome123", l.WelcomeTemplate.ID)
				assert.Equal(t, int64(2), l.WelcomeTemplate.Version)
				assert.Equal(t, "unsub123", l.UnsubscribeTemplate.ID)
				assert.Equal(t, int64(3), l.UnsubscribeTemplate.Version)
				return nil
			})

		err := service.UpdateList(ctx, workspaceID, listWithTemplates)
		assert.NoError(t, err)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.UpdateList(ctx, workspaceID, list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("validation failure", func(t *testing.T) {
		invalidList := &domain.List{} // Missing required fields
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.UpdateList(ctx, workspaceID, invalidList)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid list")
	})

	t.Run("repository failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpdateList(ctx, workspaceID, gomock.Any()).Return(errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", list.ID).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		err := service.UpdateList(ctx, workspaceID, list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update list")
	})
}

func TestListService_DeleteList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "list123"

	t.Run("successful deletion", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().DeleteList(ctx, workspaceID, listID).Return(nil)
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.NoError(t, err)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("repository failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().DeleteList(ctx, workspaceID, listID).Return(errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		err := service.DeleteList(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete list")
	})
}

func TestListService_IncrementTotal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "list123"
	totalType := domain.TotalTypeActive

	t.Run("successful increment", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().IncrementTotal(ctx, workspaceID, listID, totalType).Return(nil)
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		err := service.IncrementTotal(ctx, workspaceID, listID, totalType)
		assert.NoError(t, err)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.IncrementTotal(ctx, workspaceID, listID, totalType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("invalid total type", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.IncrementTotal(ctx, workspaceID, listID, domain.ContactListTotalType("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid total type")
	})

	t.Run("repository failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().IncrementTotal(ctx, workspaceID, listID, totalType).Return(errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger)
		mockLogger.EXPECT().WithField("total_type", totalType).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		err := service.IncrementTotal(ctx, workspaceID, listID, totalType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to increment total")
	})
}

func TestListService_DecrementTotal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockListRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewListService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "list123"
	totalType := domain.TotalTypeActive

	t.Run("successful decrement", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().DecrementTotal(ctx, workspaceID, listID, totalType).Return(nil)
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).Times(0)
		mockLogger.EXPECT().Error(gomock.Any()).Times(0)

		err := service.DecrementTotal(ctx, workspaceID, listID, totalType)
		assert.NoError(t, err)
	})

	t.Run("authentication failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.DecrementTotal(ctx, workspaceID, listID, totalType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("invalid total type", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.DecrementTotal(ctx, workspaceID, listID, domain.ContactListTotalType("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid total type")
	})

	t.Run("repository failure", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().DecrementTotal(ctx, workspaceID, listID, totalType).Return(errors.New("db error"))
		mockLogger.EXPECT().WithField("list_id", listID).Return(mockLogger)
		mockLogger.EXPECT().WithField("total_type", totalType).Return(mockLogger)
		mockLogger.EXPECT().Error(gomock.Any())

		err := service.DecrementTotal(ctx, workspaceID, listID, totalType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrement total")
	})
}
