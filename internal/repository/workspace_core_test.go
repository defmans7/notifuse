package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
)

func TestWorkspaceRepository_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWorkspaceRepository(ctrl)

	t.Run("successful retrieval", func(t *testing.T) {
		workspaceID := "testworkspace"
		workspaceName := "Test Workspace"
		createdAt := time.Now()
		updatedAt := time.Now()

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: workspaceName,
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			Integrations: []domain.Integration{
				{
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind: domain.EmailProviderKindMailgun,
					},
				},
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		repo.EXPECT().GetByID(context.Background(), workspaceID).Return(expectedWorkspace, nil)

		workspace, err := repo.GetByID(context.Background(), workspaceID)
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, workspaceName, workspace.Name)
		assert.Equal(t, "UTC", workspace.Settings.Timezone)
		assert.Equal(t, createdAt.Unix(), workspace.CreatedAt.Unix())
		assert.Equal(t, updatedAt.Unix(), workspace.UpdatedAt.Unix())
		assert.NotNil(t, workspace.Integrations)
		assert.Len(t, workspace.Integrations, 1)
	})

	t.Run("workspace not found", func(t *testing.T) {
		repo.EXPECT().GetByID(context.Background(), "nonexistent").Return(nil, fmt.Errorf("workspace not found"))

		workspace, err := repo.GetByID(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, workspace)
	})

	t.Run("database connection error", func(t *testing.T) {
		repo.EXPECT().GetByID(context.Background(), "testworkspace").Return(nil, fmt.Errorf("connection refused"))

		workspace, err := repo.GetByID(context.Background(), "testworkspace")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, workspace)
	})

	t.Run("empty workspace ID", func(t *testing.T) {
		repo.EXPECT().GetByID(context.Background(), "").Return(nil, fmt.Errorf("workspace not found"))

		workspace, err := repo.GetByID(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, workspace)
	})

	t.Run("workspace with minimal settings", func(t *testing.T) {
		workspaceID := "minimal-workspace"
		workspaceName := "Minimal Workspace"
		createdAt := time.Now()
		updatedAt := time.Now()

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: workspaceName,
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
			Integrations: []domain.Integration{},
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		repo.EXPECT().GetByID(context.Background(), workspaceID).Return(expectedWorkspace, nil)

		workspace, err := repo.GetByID(context.Background(), workspaceID)
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, workspaceName, workspace.Name)
		assert.Equal(t, "UTC", workspace.Settings.Timezone)
		assert.Empty(t, workspace.Integrations)
	})

	t.Run("workspace with null integrations", func(t *testing.T) {
		workspaceID := "null-integrations-workspace"
		workspaceName := "Null Integrations Workspace"
		createdAt := time.Now()
		updatedAt := time.Now()

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: workspaceName,
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			Integrations: []domain.Integration{},
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		repo.EXPECT().GetByID(context.Background(), workspaceID).Return(expectedWorkspace, nil)

		workspace, err := repo.GetByID(context.Background(), workspaceID)
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, workspaceName, workspace.Name)
		assert.Equal(t, "UTC", workspace.Settings.Timezone)
		assert.Empty(t, workspace.Integrations)
	})
}

func TestWorkspaceRepository_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWorkspaceRepository(ctrl)

	t.Run("successful retrieval with multiple workspaces", func(t *testing.T) {
		workspace1CreatedAt := time.Now().Add(-2 * time.Hour)
		workspace1UpdatedAt := time.Now().Add(-2 * time.Hour)
		workspace2CreatedAt := time.Now().Add(-1 * time.Hour)
		workspace2UpdatedAt := time.Now().Add(-1 * time.Hour)

		expectedWorkspaces := []*domain.Workspace{
			{
				ID:   "workspace2",
				Name: "Workspace 2",
				Settings: domain.WorkspaceSettings{
					Timezone:  "Europe/London",
					SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				},
				Integrations: []domain.Integration{},
				CreatedAt:    workspace2CreatedAt,
				UpdatedAt:    workspace2UpdatedAt,
			},
			{
				ID:   "workspace1",
				Name: "Workspace 1",
				Settings: domain.WorkspaceSettings{
					Timezone:  "UTC",
					SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				},
				Integrations: []domain.Integration{
					{
						Type: domain.IntegrationTypeEmail,
						EmailProvider: domain.EmailProvider{
							Kind: domain.EmailProviderKindMailgun,
						},
					},
				},
				CreatedAt: workspace1CreatedAt,
				UpdatedAt: workspace1UpdatedAt,
			},
		}

		repo.EXPECT().List(context.Background()).Return(expectedWorkspaces, nil)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, len(workspaces))

		// Verify order (newest first)
		assert.Equal(t, "workspace2", workspaces[0].ID)
		assert.Equal(t, "Workspace 2", workspaces[0].Name)
		assert.Equal(t, "Europe/London", workspaces[0].Settings.Timezone)

		assert.Equal(t, "workspace1", workspaces[1].ID)
		assert.Equal(t, "Workspace 1", workspaces[1].Name)
		assert.Equal(t, "UTC", workspaces[1].Settings.Timezone)
		assert.Len(t, workspaces[1].Integrations, 1)
	})

	t.Run("empty result set", func(t *testing.T) {
		repo.EXPECT().List(context.Background()).Return([]*domain.Workspace{}, nil)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, workspaces)
	})

	t.Run("single workspace", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := time.Now()

		expectedWorkspaces := []*domain.Workspace{
			{
				ID:   "single-workspace",
				Name: "Single Workspace",
				Settings: domain.WorkspaceSettings{
					Timezone:  "America/New_York",
					SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				},
				Integrations: []domain.Integration{
					{
						Type: domain.IntegrationTypeEmail,
						EmailProvider: domain.EmailProvider{
							Kind: domain.EmailProviderKindSMTP,
						},
					},
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		}

		repo.EXPECT().List(context.Background()).Return(expectedWorkspaces, nil)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, len(workspaces))
		assert.Equal(t, "single-workspace", workspaces[0].ID)
		assert.Equal(t, "Single Workspace", workspaces[0].Name)
		assert.Equal(t, "America/New_York", workspaces[0].Settings.Timezone)
		assert.Len(t, workspaces[0].Integrations, 1)
	})

	t.Run("database connection error", func(t *testing.T) {
		repo.EXPECT().List(context.Background()).Return(nil, fmt.Errorf("connection timeout"))

		workspaces, err := repo.List(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection timeout")
		assert.Nil(t, workspaces)
	})

	t.Run("workspaces with various configurations", func(t *testing.T) {
		workspace1CreatedAt := time.Now().Add(-2 * time.Hour)
		workspace1UpdatedAt := time.Now().Add(-2 * time.Hour)
		workspace2CreatedAt := time.Now().Add(-1 * time.Hour)
		workspace2UpdatedAt := time.Now().Add(-1 * time.Hour)

		expectedWorkspaces := []*domain.Workspace{
			{
				ID:   "full-workspace",
				Name: "Full Workspace",
				Settings: domain.WorkspaceSettings{
					Timezone:  "Asia/Tokyo",
					SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				},
				Integrations: []domain.Integration{
					{
						Type: domain.IntegrationTypeEmail,
						EmailProvider: domain.EmailProvider{
							Kind: domain.EmailProviderKindMailgun,
						},
					},
					{
						Type: domain.IntegrationTypeEmail,
						EmailProvider: domain.EmailProvider{
							Kind: domain.EmailProviderKindSMTP,
						},
					},
				},
				CreatedAt: workspace2CreatedAt,
				UpdatedAt: workspace2UpdatedAt,
			},
			{
				ID:   "minimal-workspace",
				Name: "Minimal Workspace",
				Settings: domain.WorkspaceSettings{
					Timezone: "UTC",
				},
				Integrations: []domain.Integration{},
				CreatedAt:    workspace1CreatedAt,
				UpdatedAt:    workspace1UpdatedAt,
			},
		}

		repo.EXPECT().List(context.Background()).Return(expectedWorkspaces, nil)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, len(workspaces))

		// Verify full workspace
		assert.Equal(t, "full-workspace", workspaces[0].ID)
		assert.Equal(t, "Asia/Tokyo", workspaces[0].Settings.Timezone)
		assert.Len(t, workspaces[0].Integrations, 2)

		// Verify minimal workspace
		assert.Equal(t, "minimal-workspace", workspaces[1].ID)
		assert.Equal(t, "UTC", workspaces[1].Settings.Timezone)
		assert.Empty(t, workspaces[1].Integrations)
	})
}

func TestWorkspaceRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWorkspaceRepository(ctrl)

	t.Run("successful creation", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		repo.EXPECT().Create(context.Background(), workspace).Return(nil)

		err := repo.Create(context.Background(), workspace)
		require.NoError(t, err)
	})

	t.Run("workspace ID already exists", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "existing-workspace",
			Name: "Existing Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		repo.EXPECT().Create(context.Background(), workspace).Return(fmt.Errorf("workspace with ID existing-workspace already exists"))

		err := repo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("database error during creation", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		repo.EXPECT().Create(context.Background(), workspace).Return(fmt.Errorf("failed to create workspace: database error"))

		err := repo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create workspace")
	})
}

func TestWorkspaceRepository_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWorkspaceRepository(ctrl)

	t.Run("successful update", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "America/New_York",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			Integrations: []domain.Integration{
				{
					ID:   "integration-1",
					Name: "SendGrid Integration",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind: domain.EmailProviderKindSMTP,
					},
				},
			},
		}

		repo.EXPECT().Update(context.Background(), workspace).Return(nil)

		err := repo.Update(context.Background(), workspace)
		require.NoError(t, err)
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "nonexistent-workspace",
			Name: "Nonexistent Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		repo.EXPECT().Update(context.Background(), workspace).Return(fmt.Errorf("workspace with ID nonexistent-workspace not found"))

		err := repo.Update(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("database connection error", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "Europe/Paris",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		repo.EXPECT().Update(context.Background(), workspace).Return(fmt.Errorf("failed to update workspace: connection lost"))

		err := repo.Update(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection lost")
	})
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWorkspaceRepository(ctrl)

	t.Run("successful deletion", func(t *testing.T) {
		workspaceID := "test-workspace"

		repo.EXPECT().Delete(context.Background(), workspaceID).Return(nil)

		err := repo.Delete(context.Background(), workspaceID)
		require.NoError(t, err)
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspaceID := "nonexistent-workspace"

		repo.EXPECT().Delete(context.Background(), workspaceID).Return(fmt.Errorf("workspace not found"))

		err := repo.Delete(context.Background(), workspaceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace not found")
	})

	t.Run("database error during deletion", func(t *testing.T) {
		workspaceID := "error-workspace"

		repo.EXPECT().Delete(context.Background(), workspaceID).Return(fmt.Errorf("database error"))

		err := repo.Delete(context.Background(), workspaceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("permission denied during database deletion", func(t *testing.T) {
		workspaceID := "permission-denied-workspace"

		repo.EXPECT().Delete(context.Background(), workspaceID).Return(fmt.Errorf("permission denied"))

		err := repo.Delete(context.Background(), workspaceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}
