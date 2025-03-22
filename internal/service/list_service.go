package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ListService struct {
	repo   domain.ListRepository
	logger logger.Logger
}

func NewListService(repo domain.ListRepository, logger logger.Logger) *ListService {
	return &ListService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ListService) CreateList(ctx context.Context, list *domain.List) error {
	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	if err := list.Validate(); err != nil {
		return fmt.Errorf("invalid list: %w", err)
	}

	if err := s.repo.CreateList(ctx, list); err != nil {
		s.logger.WithField("list_id", list.ID).Error(fmt.Sprintf("Failed to create list: %v", err))
		return fmt.Errorf("failed to create list: %w", err)
	}

	return nil
}

func (s *ListService) GetListByID(ctx context.Context, id string) (*domain.List, error) {
	list, err := s.repo.GetListByID(ctx, id)
	if err != nil {
		if _, ok := err.(*domain.ErrListNotFound); ok {
			return nil, err
		}
		s.logger.WithField("list_id", id).Error(fmt.Sprintf("Failed to get list: %v", err))
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return list, nil
}

func (s *ListService) GetLists(ctx context.Context) ([]*domain.List, error) {
	lists, err := s.repo.GetLists(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get lists: %v", err))
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	return lists, nil
}

func (s *ListService) UpdateList(ctx context.Context, list *domain.List) error {
	list.UpdatedAt = time.Now().UTC()

	if err := list.Validate(); err != nil {
		return fmt.Errorf("invalid list: %w", err)
	}

	if err := s.repo.UpdateList(ctx, list); err != nil {
		s.logger.WithField("list_id", list.ID).Error(fmt.Sprintf("Failed to update list: %v", err))
		return fmt.Errorf("failed to update list: %w", err)
	}

	return nil
}

func (s *ListService) DeleteList(ctx context.Context, id string) error {
	if err := s.repo.DeleteList(ctx, id); err != nil {
		s.logger.WithField("list_id", id).Error(fmt.Sprintf("Failed to delete list: %v", err))
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return nil
}
