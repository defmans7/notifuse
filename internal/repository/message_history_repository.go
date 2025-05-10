package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

// MessageHistoryRepository implements domain.MessageHistoryRepository
type MessageHistoryRepository struct {
	workspaceRepo domain.WorkspaceRepository
}

// NewMessageHistoryRepository creates a new message history repository
func NewMessageHistoryRepository(workspaceRepo domain.WorkspaceRepository) *MessageHistoryRepository {
	return &MessageHistoryRepository{
		workspaceRepo: workspaceRepo,
	}
}

// Create adds a new message history record
func (r *MessageHistoryRepository) Create(ctx context.Context, workspaceID string, message *domain.MessageHistory) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		INSERT INTO message_history (
			id, contact_id, broadcast_id, template_id, template_version, 
			channel, status, message_data, sent_at, delivered_at, 
			failed_at, opened_at, clicked_at, bounced_at, complained_at, 
			unsubscribed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, $10, 
			$11, $12, $13, $14, $15, 
			$16, $17, $18
		)
	`

	_, err = workspaceDB.ExecContext(
		ctx,
		query,
		message.ID,
		message.ContactID,
		message.BroadcastID,
		message.TemplateID,
		message.TemplateVersion,
		message.Channel,
		message.Status,
		message.MessageData,
		message.SentAt,
		message.DeliveredAt,
		message.FailedAt,
		message.OpenedAt,
		message.ClickedAt,
		message.BouncedAt,
		message.ComplainedAt,
		message.UnsubscribedAt,
		message.CreatedAt,
		message.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message history: %w", err)
	}

	return nil
}

// Update updates an existing message history record
func (r *MessageHistoryRepository) Update(ctx context.Context, workspaceID string, message *domain.MessageHistory) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		UPDATE message_history SET
			contact_id = $2,
			broadcast_id = $3,
			template_id = $4,
			template_version = $5,
			channel = $6,
			status = $7,
			message_data = $8,
			sent_at = $9,
			delivered_at = $10,
			failed_at = $11,
			opened_at = $12,
			clicked_at = $13,
			bounced_at = $14,
			complained_at = $15,
			unsubscribed_at = $16,
			updated_at = $17
		WHERE id = $1
	`

	_, err = workspaceDB.ExecContext(
		ctx,
		query,
		message.ID,
		message.ContactID,
		message.BroadcastID,
		message.TemplateID,
		message.TemplateVersion,
		message.Channel,
		message.Status,
		message.MessageData,
		message.SentAt,
		message.DeliveredAt,
		message.FailedAt,
		message.OpenedAt,
		message.ClickedAt,
		message.BouncedAt,
		message.ComplainedAt,
		message.UnsubscribedAt,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update message history: %w", err)
	}

	return nil
}

// Get retrieves a message history by ID
func (r *MessageHistoryRepository) Get(ctx context.Context, workspaceID, id string) (*domain.MessageHistory, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT 
			id, contact_id, broadcast_id, template_id, template_version, 
			channel, status, message_data, sent_at, delivered_at, 
			failed_at, opened_at, clicked_at, bounced_at, complained_at, 
			unsubscribed_at, created_at, updated_at
		FROM message_history
		WHERE id = $1
	`

	var message domain.MessageHistory
	err = workspaceDB.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ContactID,
		&message.BroadcastID,
		&message.TemplateID,
		&message.TemplateVersion,
		&message.Channel,
		&message.Status,
		&message.MessageData,
		&message.SentAt,
		&message.DeliveredAt,
		&message.FailedAt,
		&message.OpenedAt,
		&message.ClickedAt,
		&message.BouncedAt,
		&message.ComplainedAt,
		&message.UnsubscribedAt,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message history with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	return &message, nil
}

// GetByContact retrieves message history for a specific contact
func (r *MessageHistoryRepository) GetByContact(ctx context.Context, workspaceID, contactID string, limit, offset int) ([]*domain.MessageHistory, int, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First get total count
	countQuery := `SELECT COUNT(*) FROM message_history WHERE contact_id = $1`
	var totalCount int
	err = workspaceDB.QueryRowContext(ctx, countQuery, contactID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count message history: %w", err)
	}

	// Set default limit and offset if not provided
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			id, contact_id, broadcast_id, template_id, template_version, 
			channel, status, message_data, sent_at, delivered_at, 
			failed_at, opened_at, clicked_at, bounced_at, complained_at, 
			unsubscribed_at, created_at, updated_at
		FROM message_history
		WHERE contact_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := workspaceDB.QueryContext(ctx, query, contactID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query message history: %w", err)
	}
	defer rows.Close()

	var messages []*domain.MessageHistory
	for rows.Next() {
		var message domain.MessageHistory
		err := rows.Scan(
			&message.ID,
			&message.ContactID,
			&message.BroadcastID,
			&message.TemplateID,
			&message.TemplateVersion,
			&message.Channel,
			&message.Status,
			&message.MessageData,
			&message.SentAt,
			&message.DeliveredAt,
			&message.FailedAt,
			&message.OpenedAt,
			&message.ClickedAt,
			&message.BouncedAt,
			&message.ComplainedAt,
			&message.UnsubscribedAt,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message history: %w", err)
		}
		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating message history rows: %w", err)
	}

	return messages, totalCount, nil
}

// GetByBroadcast retrieves message history for a specific broadcast
func (r *MessageHistoryRepository) GetByBroadcast(ctx context.Context, workspaceID, broadcastID string, limit, offset int) ([]*domain.MessageHistory, int, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First get total count
	countQuery := `SELECT COUNT(*) FROM message_history WHERE broadcast_id = $1`
	var totalCount int
	err = workspaceDB.QueryRowContext(ctx, countQuery, broadcastID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count message history: %w", err)
	}

	// Set default limit and offset if not provided
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			id, contact_id, broadcast_id, template_id, template_version, 
			channel, status, message_data, sent_at, delivered_at, 
			failed_at, opened_at, clicked_at, bounced_at, complained_at, 
			unsubscribed_at, created_at, updated_at
		FROM message_history
		WHERE broadcast_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := workspaceDB.QueryContext(ctx, query, broadcastID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query message history: %w", err)
	}
	defer rows.Close()

	var messages []*domain.MessageHistory
	for rows.Next() {
		var message domain.MessageHistory
		err := rows.Scan(
			&message.ID,
			&message.ContactID,
			&message.BroadcastID,
			&message.TemplateID,
			&message.TemplateVersion,
			&message.Channel,
			&message.Status,
			&message.MessageData,
			&message.SentAt,
			&message.DeliveredAt,
			&message.FailedAt,
			&message.OpenedAt,
			&message.ClickedAt,
			&message.BouncedAt,
			&message.ComplainedAt,
			&message.UnsubscribedAt,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message history: %w", err)
		}
		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating message history rows: %w", err)
	}

	return messages, totalCount, nil
}

// UpdateStatus updates the status of a message and sets the corresponding timestamp
func (r *MessageHistoryRepository) UpdateStatus(ctx context.Context, workspaceID, id string, status domain.MessageStatus, timestamp time.Time) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	var field string

	switch status {
	case domain.MessageStatusDelivered:
		field = "delivered_at"
	case domain.MessageStatusFailed:
		field = "failed_at"
	case domain.MessageStatusOpened:
		field = "opened_at"
	case domain.MessageStatusClicked:
		field = "clicked_at"
	case domain.MessageStatusBounced:
		field = "bounced_at"
	case domain.MessageStatusComplained:
		field = "complained_at"
	case domain.MessageStatusUnsubscribed:
		field = "unsubscribed_at"
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	query := fmt.Sprintf(`
		UPDATE message_history 
		SET status = $1, %s = $2, updated_at = $3
		WHERE id = $4
	`, field)

	_, err = workspaceDB.ExecContext(ctx, query, status, timestamp, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

func (r *MessageHistoryRepository) SetClicked(ctx context.Context, workspaceID, id string, timestamp time.Time) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First query: Update clicked_at if it's null
	clickQuery := `
		UPDATE message_history 
		SET 
			clicked_at = $1,
			status = 'clicked',
			updated_at = NOW()
		WHERE id = $2 AND clicked_at IS NULL
	`

	_, err = workspaceDB.ExecContext(ctx, clickQuery, timestamp, id)
	if err != nil {
		return fmt.Errorf("failed to set clicked: %w", err)
	}

	// Second query: Update opened_at if it's null as a click means the message was opened
	openQuery := `
		UPDATE message_history 
		SET 
			opened_at = $1,
			updated_at = NOW()
		WHERE id = $2 AND opened_at IS NULL
	`

	_, err = workspaceDB.ExecContext(ctx, openQuery, timestamp, id)
	if err != nil {
		return fmt.Errorf("failed to set opened: %w", err)
	}

	return nil
}

func (r *MessageHistoryRepository) SetOpened(ctx context.Context, workspaceID, id string, timestamp time.Time) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First query: Update opened_at if it's null
	query := `
		UPDATE message_history 
		SET 
			opened_at = $1,
			updated_at = NOW()
		WHERE id = $2 AND opened_at IS NULL
	`

	_, err = workspaceDB.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		return fmt.Errorf("failed to set opened: %w", err)
	}

	return nil
}
