package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ListService struct {
	repo        domain.ListRepository
	authService domain.AuthService
	logger      logger.Logger
}

func NewListService(repo domain.ListRepository, authService domain.AuthService, logger logger.Logger) *ListService {
	return &ListService{
		repo:        repo,
		authService: authService,
		logger:      logger,
	}
}

func (s *ListService) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	if err := list.Validate(); err != nil {
		return fmt.Errorf("invalid list: %w", err)
	}

	if err := s.repo.CreateList(ctx, workspaceID, list); err != nil {
		s.logger.WithField("list_id", list.ID).Error(fmt.Sprintf("Failed to create list: %v", err))
		return fmt.Errorf("failed to create list: %w", err)
	}

	return nil
}

func (s *ListService) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	list, err := s.repo.GetListByID(ctx, workspaceID, id)
	if err != nil {
		if _, ok := err.(*domain.ErrListNotFound); ok {
			return nil, err
		}
		s.logger.WithField("list_id", id).Error(fmt.Sprintf("Failed to get list: %v", err))
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return list, nil
}

func (s *ListService) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	lists, err := s.repo.GetLists(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get lists: %v", err))
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	return lists, nil
}

func (s *ListService) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	list.UpdatedAt = time.Now().UTC()

	if err := list.Validate(); err != nil {
		return fmt.Errorf("invalid list: %w", err)
	}

	if err := s.repo.UpdateList(ctx, workspaceID, list); err != nil {
		s.logger.WithField("list_id", list.ID).Error(fmt.Sprintf("Failed to update list: %v", err))
		return fmt.Errorf("failed to update list: %w", err)
	}

	return nil
}

func (s *ListService) DeleteList(ctx context.Context, workspaceID string, id string) error {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	if err := s.repo.DeleteList(ctx, workspaceID, id); err != nil {
		s.logger.WithField("list_id", id).Error(fmt.Sprintf("Failed to delete list: %v", err))
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return nil
}

func (s *ListService) IncrementTotal(ctx context.Context, workspaceID string, listID string, totalType domain.ContactListTotalType) error {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	if err := totalType.Validate(); err != nil {
		return err
	}

	if err := s.repo.IncrementTotal(ctx, workspaceID, listID, totalType); err != nil {
		s.logger.WithField("list_id", listID).WithField("total_type", totalType).Error(fmt.Sprintf("Failed to increment total: %v", err))
		return fmt.Errorf("failed to increment total: %w", err)
	}

	return nil
}

func (s *ListService) DecrementTotal(ctx context.Context, workspaceID string, listID string, totalType domain.ContactListTotalType) error {
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	if err := totalType.Validate(); err != nil {
		return err
	}

	if err := s.repo.DecrementTotal(ctx, workspaceID, listID, totalType); err != nil {
		s.logger.WithField("list_id", listID).WithField("total_type", totalType).Error(fmt.Sprintf("Failed to decrement total: %v", err))
		return fmt.Errorf("failed to decrement total: %w", err)
	}

	return nil
}
