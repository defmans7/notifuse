package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockWorkspaceService(t *testing.T) {
	mockService := &MockWorkspaceService{}
	ctx := context.Background()

	t.Run("CreateWorkspace", func(t *testing.T) {
		// Test success case
		id := "workspace-123"
		name := "Test Workspace"
		websiteURL := "https://example.com"
		logoURL := "https://example.com/logo.png"
		coverURL := "https://example.com/cover.png"
		timezone := "UTC"

		expectedWorkspace := &domain.Workspace{
			ID:   id,
			Name: name,
			Settings: domain.WorkspaceSettings{
				WebsiteURL: websiteURL,
				LogoURL:    logoURL,
				CoverURL:   coverURL,
				Timezone:   timezone,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateWorkspace", ctx, id, name, websiteURL, logoURL, coverURL, timezone).
			Return(expectedWorkspace, nil).Once()

		workspace, err := mockService.CreateWorkspace(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
		assert.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)

		// Test error case
		mockService.On("CreateWorkspace", ctx, id, name, websiteURL, logoURL, coverURL, timezone).
			Return(nil, fmt.Errorf("failed to create workspace")).Once()

		workspace, err = mockService.CreateWorkspace(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
		assert.Error(t, err)
		assert.Nil(t, workspace)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		// Test success case
		id := "workspace-123"
		expectedWorkspace := &domain.Workspace{
			ID:   id,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetWorkspace", ctx, id).Return(expectedWorkspace, nil).Once()

		workspace, err := mockService.GetWorkspace(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)

		// Test error case
		mockService.On("GetWorkspace", ctx, id).Return(nil, fmt.Errorf("workspace not found")).Once()

		workspace, err = mockService.GetWorkspace(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, workspace)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		// Test success case
		expectedWorkspaces := []*domain.Workspace{
			{
				ID:   "workspace-1",
				Name: "Workspace 1",
				Settings: domain.WorkspaceSettings{
					Timezone: "UTC",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:   "workspace-2",
				Name: "Workspace 2",
				Settings: domain.WorkspaceSettings{
					Timezone: "UTC",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockService.On("ListWorkspaces", ctx).Return(expectedWorkspaces, nil).Once()

		workspaces, err := mockService.ListWorkspaces(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedWorkspaces, workspaces)

		// Test error case
		mockService.On("ListWorkspaces", ctx).Return(nil, fmt.Errorf("failed to list workspaces")).Once()

		workspaces, err = mockService.ListWorkspaces(ctx)
		assert.Error(t, err)
		assert.Nil(t, workspaces)
	})

	t.Run("UpdateWorkspace", func(t *testing.T) {
		// Test success case
		id := "workspace-123"
		name := "Updated Workspace"
		websiteURL := "https://example.com"
		logoURL := "https://example.com/logo.png"
		coverURL := "https://example.com/cover.png"
		timezone := "UTC"

		expectedWorkspace := &domain.Workspace{
			ID:   id,
			Name: name,
			Settings: domain.WorkspaceSettings{
				WebsiteURL: websiteURL,
				LogoURL:    logoURL,
				CoverURL:   coverURL,
				Timezone:   timezone,
			},
			UpdatedAt: time.Now(),
		}

		mockService.On("UpdateWorkspace", ctx, id, name, websiteURL, logoURL, coverURL, timezone).
			Return(expectedWorkspace, nil).Once()

		workspace, err := mockService.UpdateWorkspace(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
		assert.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)

		// Test error case
		mockService.On("UpdateWorkspace", ctx, id, name, websiteURL, logoURL, coverURL, timezone).
			Return(nil, fmt.Errorf("failed to update workspace")).Once()

		workspace, err = mockService.UpdateWorkspace(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
		assert.Error(t, err)
		assert.Nil(t, workspace)
	})

	t.Run("DeleteWorkspace", func(t *testing.T) {
		// Test success case
		id := "workspace-123"
		mockService.On("DeleteWorkspace", ctx, id).Return(nil).Once()

		err := mockService.DeleteWorkspace(ctx, id)
		assert.NoError(t, err)

		// Test error case
		mockService.On("DeleteWorkspace", ctx, id).Return(fmt.Errorf("failed to delete workspace")).Once()

		err = mockService.DeleteWorkspace(ctx, id)
		assert.Error(t, err)
	})

	t.Run("GetWorkspaceMembersWithEmail", func(t *testing.T) {
		// Test success case
		id := "workspace-123"
		expectedMembers := []*domain.UserWorkspaceWithEmail{
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user-1",
					WorkspaceID: id,
					Role:        "owner",
				},
				Email: "user1@example.com",
			},
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user-2",
					WorkspaceID: id,
					Role:        "member",
				},
				Email: "user2@example.com",
			},
		}

		mockService.On("GetWorkspaceMembersWithEmail", ctx, id).Return(expectedMembers, nil).Once()

		members, err := mockService.GetWorkspaceMembersWithEmail(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedMembers, members)

		// Test error case
		mockService.On("GetWorkspaceMembersWithEmail", ctx, id).Return(nil, fmt.Errorf("failed to get members")).Once()

		members, err = mockService.GetWorkspaceMembersWithEmail(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, members)
	})

	t.Run("InviteMember", func(t *testing.T) {
		// Test success case
		workspaceID := "workspace-123"
		email := "new@example.com"
		token := "invitation-token"
		expectedInvitation := &domain.WorkspaceInvitation{
			ID:          "invitation-123",
			WorkspaceID: workspaceID,
			Email:       email,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}

		mockService.On("InviteMember", ctx, workspaceID, email).Return(expectedInvitation, token, nil).Once()

		invitation, invitationToken, err := mockService.InviteMember(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.Equal(t, expectedInvitation, invitation)
		assert.Equal(t, token, invitationToken)

		// Test error case
		mockService.On("InviteMember", ctx, workspaceID, email).Return(nil, "", fmt.Errorf("failed to invite member")).Once()

		invitation, invitationToken, err = mockService.InviteMember(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, invitationToken)
	})

	// Verify that all expected mock calls were made
	mockService.AssertExpectations(t)
}

func TestMockWorkspaceService_CreateWorkspace(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()

	websiteURL := "https://example.com"
	logoURL := "https://example.com/logo.png"
	coverURL := "https://example.com/cover.png"
	timezone := "UTC"

	workspace := &domain.Workspace{
		ID:   "test123",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			CoverURL:   coverURL,
			Timezone:   timezone,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mockService.On("CreateWorkspace", ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone).
			Return(workspace, nil).Once()

		result, err := mockService.CreateWorkspace(ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone)

		assert.NoError(t, err)
		assert.Equal(t, workspace, result)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("creation failed")
		mockService.On("CreateWorkspace", ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone).
			Return(nil, expectedErr).Once()

		result, err := mockService.CreateWorkspace(ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}

func TestMockWorkspaceService_GetWorkspace(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()
	workspaceID := "test123"

	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	t.Run("success", func(t *testing.T) {
		mockService.On("GetWorkspace", ctx, workspaceID).Return(workspace, nil).Once()

		result, err := mockService.GetWorkspace(ctx, workspaceID)

		assert.NoError(t, err)
		assert.Equal(t, workspace, result)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("workspace not found")
		mockService.On("GetWorkspace", ctx, workspaceID).Return(nil, expectedErr).Once()

		result, err := mockService.GetWorkspace(ctx, workspaceID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}

func TestMockWorkspaceService_ListWorkspaces(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()

	workspaces := []*domain.Workspace{
		{
			ID:   "test1",
			Name: "Test Workspace 1",
		},
		{
			ID:   "test2",
			Name: "Test Workspace 2",
		},
	}

	t.Run("success", func(t *testing.T) {
		mockService.On("ListWorkspaces", ctx).Return(workspaces, nil).Once()

		result, err := mockService.ListWorkspaces(ctx)

		assert.NoError(t, err)
		assert.Equal(t, workspaces, result)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("failed to list workspaces")
		mockService.On("ListWorkspaces", ctx).Return(nil, expectedErr).Once()

		result, err := mockService.ListWorkspaces(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}

func TestMockWorkspaceService_UpdateWorkspace(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()

	websiteURL := "https://updated.com"
	logoURL := "https://updated.com/logo.png"
	coverURL := "https://updated.com/cover.png"
	timezone := "UTC"

	workspace := &domain.Workspace{
		ID:   "test123",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			CoverURL:   coverURL,
			Timezone:   timezone,
		},
	}

	t.Run("success", func(t *testing.T) {
		mockService.On("UpdateWorkspace", ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone).
			Return(workspace, nil).Once()

		result, err := mockService.UpdateWorkspace(ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone)

		assert.NoError(t, err)
		assert.Equal(t, workspace, result)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mockService.On("UpdateWorkspace", ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone).
			Return(nil, expectedErr).Once()

		result, err := mockService.UpdateWorkspace(ctx, workspace.ID, workspace.Name,
			websiteURL, logoURL, coverURL, timezone)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}

func TestMockWorkspaceService_DeleteWorkspace(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()
	workspaceID := "test123"

	t.Run("success", func(t *testing.T) {
		mockService.On("DeleteWorkspace", ctx, workspaceID).Return(nil).Once()

		err := mockService.DeleteWorkspace(ctx, workspaceID)

		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("deletion failed")
		mockService.On("DeleteWorkspace", ctx, workspaceID).Return(expectedErr).Once()

		err := mockService.DeleteWorkspace(ctx, workspaceID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}

func TestMockWorkspaceService_GetWorkspaceMembersWithEmail(t *testing.T) {
	mockService := new(MockWorkspaceService)
	ctx := context.Background()
	workspaceID := "test123"

	members := []*domain.UserWorkspaceWithEmail{
		{
			UserWorkspace: domain.UserWorkspace{
				UserID:      "user1",
				WorkspaceID: workspaceID,
				Role:        "owner",
			},
			Email: "user1@example.com",
		},
		{
			UserWorkspace: domain.UserWorkspace{
				UserID:      "user2",
				WorkspaceID: workspaceID,
				Role:        "member",
			},
			Email: "user2@example.com",
		},
	}

	t.Run("success", func(t *testing.T) {
		mockService.On("GetWorkspaceMembersWithEmail", ctx, workspaceID).Return(members, nil).Once()

		result, err := mockService.GetWorkspaceMembersWithEmail(ctx, workspaceID)

		assert.NoError(t, err)
		assert.Equal(t, members, result)
		mockService.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("failed to get members")
		mockService.On("GetWorkspaceMembersWithEmail", ctx, workspaceID).Return(nil, expectedErr).Once()

		result, err := mockService.GetWorkspaceMembersWithEmail(ctx, workspaceID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}
