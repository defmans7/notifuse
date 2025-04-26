package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type broadcastRepository struct {
	workspaceRepo domain.WorkspaceRepository
}

// NewBroadcastRepository creates a new PostgreSQL broadcast repository
func NewBroadcastRepository(workspaceRepo domain.WorkspaceRepository) domain.BroadcastRepository {
	return &broadcastRepository{
		workspaceRepo: workspaceRepo,
	}
}

// CreateBroadcast persists a new broadcast
func (r *broadcastRepository) CreateBroadcast(ctx context.Context, broadcast *domain.Broadcast) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, broadcast.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Set created and updated timestamps
	now := time.Now().UTC()
	broadcast.CreatedAt = now
	broadcast.UpdatedAt = now

	// Insert the broadcast
	query := `
		INSERT INTO broadcasts (
			id, 
			workspace_id,
			name, 
			status, 
			audience, 
			schedule, 
			test_settings, 
			goal_id, 
			tracking_enabled, 
			utm_parameters, 
			metadata, 
			sent_count, 
			delivered_count, 
			failed_count, 
			winning_variation, 
			test_sent_at, 
			winner_sent_at, 
			created_at, 
			updated_at, 
			scheduled_at, 
			started_at, 
			completed_at, 
			cancelled_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 
			$16, $17, $18, $19, $20, $21, $22, $23
		)
	`

	_, err = workspaceDB.ExecContext(ctx, query,
		broadcast.ID,
		broadcast.WorkspaceID,
		broadcast.Name,
		broadcast.Status,
		broadcast.Audience,
		broadcast.Schedule,
		broadcast.TestSettings,
		broadcast.GoalID,
		broadcast.TrackingEnabled,
		broadcast.UTMParameters,
		broadcast.Metadata,
		broadcast.SentCount,
		broadcast.DeliveredCount,
		broadcast.FailedCount,
		broadcast.WinningVariation,
		broadcast.TestSentAt,
		broadcast.WinnerSentAt,
		broadcast.CreatedAt,
		broadcast.UpdatedAt,
		broadcast.ScheduledAt,
		broadcast.StartedAt,
		broadcast.CompletedAt,
		broadcast.CancelledAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create broadcast: %w", err)
	}

	return nil
}

// GetBroadcast retrieves a broadcast by ID
func (r *broadcastRepository) GetBroadcast(ctx context.Context, workspaceID, id string) (*domain.Broadcast, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT 
			id, 
			workspace_id,
			name, 
			status, 
			audience, 
			schedule, 
			test_settings, 
			goal_id, 
			tracking_enabled, 
			utm_parameters, 
			metadata, 
			sent_count, 
			delivered_count, 
			failed_count, 
			winning_variation, 
			test_sent_at, 
			winner_sent_at, 
			created_at, 
			updated_at, 
			scheduled_at, 
			started_at, 
			completed_at, 
			cancelled_at
		FROM broadcasts
		WHERE id = $1 AND workspace_id = $2
	`

	row := workspaceDB.QueryRowContext(ctx, query, id, workspaceID)

	broadcast, err := scanBroadcast(row)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrBroadcastNotFound{ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}

	return broadcast, nil
}

// UpdateBroadcast updates an existing broadcast
func (r *broadcastRepository) UpdateBroadcast(ctx context.Context, broadcast *domain.Broadcast) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, broadcast.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Update the timestamp
	broadcast.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE broadcasts SET
			name = $3,
			status = $4,
			audience = $5,
			schedule = $6,
			test_settings = $7,
			goal_id = $8,
			tracking_enabled = $9,
			utm_parameters = $10,
			metadata = $11,
			sent_count = $12,
			delivered_count = $13,
			failed_count = $14,
			winning_variation = $15,
			test_sent_at = $16,
			winner_sent_at = $17,
			updated_at = $18,
			scheduled_at = $19,
			started_at = $20,
			completed_at = $21,
			cancelled_at = $22
		WHERE id = $1 AND workspace_id = $2
	`

	result, err := workspaceDB.ExecContext(ctx, query,
		broadcast.ID,
		broadcast.WorkspaceID,
		broadcast.Name,
		broadcast.Status,
		broadcast.Audience,
		broadcast.Schedule,
		broadcast.TestSettings,
		broadcast.GoalID,
		broadcast.TrackingEnabled,
		broadcast.UTMParameters,
		broadcast.Metadata,
		broadcast.SentCount,
		broadcast.DeliveredCount,
		broadcast.FailedCount,
		broadcast.WinningVariation,
		broadcast.TestSentAt,
		broadcast.WinnerSentAt,
		broadcast.UpdatedAt,
		broadcast.ScheduledAt,
		broadcast.StartedAt,
		broadcast.CompletedAt,
		broadcast.CancelledAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update broadcast: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.ErrBroadcastNotFound{ID: broadcast.ID}
	}

	return nil
}

// ListBroadcasts retrieves a list of broadcasts
func (r *broadcastRepository) ListBroadcasts(ctx context.Context, params domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, params.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First count total records that match the criteria
	var countQuery string
	var countArgs []interface{}

	if params.Status != "" {
		countQuery = `
			SELECT COUNT(*)
			FROM broadcasts
			WHERE workspace_id = $1 AND status = $2
		`
		countArgs = []interface{}{params.WorkspaceID, params.Status}
	} else {
		countQuery = `
			SELECT COUNT(*)
			FROM broadcasts
			WHERE workspace_id = $1
		`
		countArgs = []interface{}{params.WorkspaceID}
	}

	var totalCount int
	err = workspaceDB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count broadcasts: %w", err)
	}

	// Then query paginated data
	var dataQuery string
	var dataArgs []interface{}

	if params.Status != "" {
		dataQuery = `
			SELECT 
				id, 
				workspace_id,
				name, 
				status, 
				audience, 
				schedule, 
				test_settings, 
				goal_id, 
				tracking_enabled, 
				utm_parameters, 
				metadata, 
				sent_count, 
				delivered_count, 
				failed_count, 
				winning_variation, 
				test_sent_at, 
				winner_sent_at, 
				created_at, 
				updated_at, 
				scheduled_at, 
				started_at, 
				completed_at, 
				cancelled_at
			FROM broadcasts
			WHERE workspace_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		dataArgs = []interface{}{params.WorkspaceID, params.Status, params.Limit, params.Offset}
	} else {
		dataQuery = `
			SELECT 
				id, 
				workspace_id,
				name, 
				status, 
				audience, 
				schedule, 
				test_settings, 
				goal_id, 
				tracking_enabled, 
				utm_parameters, 
				metadata, 
				sent_count, 
				delivered_count, 
				failed_count, 
				winning_variation, 
				test_sent_at, 
				winner_sent_at, 
				created_at, 
				updated_at, 
				scheduled_at, 
				started_at, 
				completed_at, 
				cancelled_at
			FROM broadcasts
			WHERE workspace_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		dataArgs = []interface{}{params.WorkspaceID, params.Limit, params.Offset}
	}

	rows, err := workspaceDB.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}
	defer rows.Close()

	var broadcasts []*domain.Broadcast
	for rows.Next() {
		broadcast, err := scanBroadcast(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan broadcast: %w", err)
		}
		broadcasts = append(broadcasts, broadcast)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broadcast rows: %w", err)
	}

	return &domain.BroadcastListResponse{
		Broadcasts: broadcasts,
		TotalCount: totalCount,
	}, nil
}

// scanBroadcast scans a row into a Broadcast struct
func scanBroadcast(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Broadcast, error) {
	broadcast := &domain.Broadcast{}

	err := scanner.Scan(
		&broadcast.ID,
		&broadcast.WorkspaceID,
		&broadcast.Name,
		&broadcast.Status,
		&broadcast.Audience,
		&broadcast.Schedule,
		&broadcast.TestSettings,
		&broadcast.GoalID,
		&broadcast.TrackingEnabled,
		&broadcast.UTMParameters,
		&broadcast.Metadata,
		&broadcast.SentCount,
		&broadcast.DeliveredCount,
		&broadcast.FailedCount,
		&broadcast.WinningVariation,
		&broadcast.TestSentAt,
		&broadcast.WinnerSentAt,
		&broadcast.CreatedAt,
		&broadcast.UpdatedAt,
		&broadcast.ScheduledAt,
		&broadcast.StartedAt,
		&broadcast.CompletedAt,
		&broadcast.CancelledAt,
	)

	if err != nil {
		return nil, err
	}

	return broadcast, nil
}
