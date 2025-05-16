package service

import (
	"context"
	"fmt"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/tracing"
)

// MessageHistoryService implements domain.MessageHistoryService interface
type MessageHistoryService struct {
	repo   domain.MessageHistoryRepository
	logger logger.Logger
}

// NewMessageHistoryService creates a new message history service
func NewMessageHistoryService(repo domain.MessageHistoryRepository, logger logger.Logger) *MessageHistoryService {
	return &MessageHistoryService{
		repo:   repo,
		logger: logger,
	}
}

// ListMessages retrieves messages for a workspace with cursor-based pagination and filters
func (s *MessageHistoryService) ListMessages(ctx context.Context, workspaceID string, params domain.MessageListParams) (*domain.MessageListResult, error) {
	// codecov:ignore:start
	ctx, span := tracing.StartServiceSpan(ctx, "MessageHistoryService", "ListMessages")
	defer tracing.EndSpan(span, nil)
	tracing.AddAttribute(ctx, "workspaceID", workspaceID)
	// codecov:ignore:end

	// Call repository method with pagination and filtering parameters
	messages, nextCursor, err := s.repo.ListMessages(ctx, workspaceID, params)

	// codecov:ignore:start
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list messages: %v", err))
		tracing.MarkSpanError(ctx, err)
		return nil, err
	}
	// codecov:ignore:end

	return &domain.MessageListResult{
		Messages:   messages,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	}, nil
}
