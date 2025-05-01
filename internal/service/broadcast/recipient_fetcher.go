package broadcast

import (
	"context"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

//go:generate mockgen -destination=./mocks/mock_recipient_fetcher.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast RecipientFetcher

// RecipientFetcher is the interface for fetching recipients for broadcasts
type RecipientFetcher interface {
	// GetTotalRecipientCount gets the total number of recipients for a broadcast
	GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error)

	// FetchBatch retrieves a batch of recipients for a broadcast
	FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error)
}

// recipientFetcher implements the RecipientFetcher interface
type recipientFetcher struct {
	broadcastService domain.BroadcastSender
	contactRepo      domain.ContactRepository
	logger           logger.Logger
	config           *Config
}

// NewRecipientFetcher creates a new recipient fetcher
func NewRecipientFetcher(broadcastService domain.BroadcastSender, contactRepo domain.ContactRepository,
	logger logger.Logger, config *Config) RecipientFetcher {
	if config == nil {
		config = DefaultConfig()
	}
	return &recipientFetcher{
		broadcastService: broadcastService,
		contactRepo:      contactRepo,
		logger:           logger,
		config:           config,
	}
}

// GetTotalRecipientCount gets the total number of recipients for a broadcast
func (f *recipientFetcher) GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	startTime := time.Now()
	defer func() {
		f.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Recipient count completed")
	}()

	// Get the broadcast to access audience settings
	broadcast, err := f.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		f.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient count")
		return 0, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Use the contact repository to count recipients
	count, err := f.contactRepo.CountContactsForBroadcast(ctx, workspaceID, broadcast.Audience)
	if err != nil {
		f.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to count recipients for broadcast")
		return 0, NewBroadcastError(ErrCodeRecipientFetch, "failed to count recipients", true, err)
	}

	f.logger.WithFields(map[string]interface{}{
		"broadcast_id":      broadcastID,
		"workspace_id":      workspaceID,
		"recipient_count":   count,
		"audience_lists":    len(broadcast.Audience.Lists),
		"audience_segments": len(broadcast.Audience.Segments),
	}).Info("Got recipient count for broadcast")

	return count, nil
}

// FetchBatch retrieves a batch of recipients for a broadcast
func (f *recipientFetcher) FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error) {
	startTime := time.Now()
	defer func() {
		f.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
		}).Debug("Recipient batch fetch completed")
	}()

	// Get the broadcast to access audience settings
	broadcast, err := f.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		f.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient fetch")
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Apply the actual batch limit from config if not specified
	if limit <= 0 {
		limit = f.config.FetchBatchSize
	}

	// Fetch contacts based on broadcast audience
	contactsWithList, err := f.contactRepo.GetContactsForBroadcast(ctx, workspaceID, broadcast.Audience, limit, offset)
	if err != nil {
		f.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
			"error":        err.Error(),
		}).Error("Failed to fetch recipients for broadcast")
		return nil, NewBroadcastError(ErrCodeRecipientFetch, "failed to fetch recipients", true, err)
	}

	f.logger.WithFields(map[string]interface{}{
		"broadcast_id":     broadcastID,
		"workspace_id":     workspaceID,
		"offset":           offset,
		"limit":            limit,
		"contacts_fetched": len(contactsWithList),
		"with_list_info":   true,
	}).Info("Fetched recipient batch with list info")

	return contactsWithList, nil
}
