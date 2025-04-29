package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// broadcastRepository implements domain.BroadcastRepository for PostgreSQL
type broadcastRepository struct {
	workspaceRepo domain.WorkspaceRepository
	logger        logger.Logger
}

// NewBroadcastRepository creates a new PostgreSQL broadcast repository
func NewBroadcastRepository(workspaceRepo domain.WorkspaceRepository, logger logger.Logger) domain.BroadcastRepository {
	return &broadcastRepository{
		workspaceRepo: workspaceRepo,
		logger:        logger,
	}
}

// WithTransaction executes a function within a transaction
func (r *broadcastRepository) WithTransaction(ctx context.Context, workspaceID string, fn func(*sql.Tx) error) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Begin a transaction
	tx, err := workspaceDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback - this will be a no-op if we successfully commit
	defer tx.Rollback()

	// Execute the provided function with the transaction
	if err := fn(tx); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
			tracking_enabled, 
			utm_parameters, 
			metadata, 
			total_sent, 
			total_delivered, 
			total_bounced, 
			total_complained, 
			total_failed, 
			total_opens,
			total_clicks,
			winning_variation, 
			test_sent_at, 
			winner_sent_at, 
			created_at, 
			updated_at, 
			started_at, 
			completed_at, 
			cancelled_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
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
		broadcast.TrackingEnabled,
		broadcast.UTMParameters,
		broadcast.Metadata,
		broadcast.TotalSent,
		broadcast.TotalDelivered,
		broadcast.TotalBounced,
		broadcast.TotalComplained,
		broadcast.TotalFailed,
		broadcast.TotalOpens,
		broadcast.TotalClicks,
		broadcast.WinningVariation,
		broadcast.TestSentAt,
		broadcast.WinnerSentAt,
		broadcast.CreatedAt,
		broadcast.UpdatedAt,
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
			tracking_enabled, 
			utm_parameters, 
			metadata, 
			total_sent, 
			total_delivered, 
			total_bounced, 
			total_complained, 
			total_failed, 
			total_opens,
			total_clicks,
			winning_variation, 
			test_sent_at, 
			winner_sent_at, 
			created_at, 
			updated_at, 
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
			tracking_enabled = $8,
			utm_parameters = $9,
			metadata = $10,
			total_sent = $11,
			total_delivered = $12,
			total_bounced = $13,
			total_complained = $14,
			total_failed = $15,
			total_opens = $16,
			total_clicks = $17,
			winning_variation = $18,
			test_sent_at = $19,
			winner_sent_at = $20,
			updated_at = $21,
			started_at = $22,
			completed_at = $23,
			cancelled_at = $24
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
		broadcast.TrackingEnabled,
		broadcast.UTMParameters,
		broadcast.Metadata,
		broadcast.TotalSent,
		broadcast.TotalDelivered,
		broadcast.TotalBounced,
		broadcast.TotalComplained,
		broadcast.TotalFailed,
		broadcast.TotalOpens,
		broadcast.TotalClicks,
		broadcast.WinningVariation,
		broadcast.TestSentAt,
		broadcast.WinnerSentAt,
		broadcast.UpdatedAt,
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
				tracking_enabled, 
				utm_parameters, 
				metadata, 
				total_sent, 
				total_delivered, 
				total_bounced, 
				total_complained, 
				total_failed, 
				total_opens,
				total_clicks,
				winning_variation, 
				test_sent_at, 
				winner_sent_at, 
				created_at, 
				updated_at, 
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
				tracking_enabled, 
				utm_parameters, 
				metadata, 
				total_sent, 
				total_delivered, 
				total_bounced, 
				total_complained, 
				total_failed, 
				total_opens,
				total_clicks,
				winning_variation, 
				test_sent_at, 
				winner_sent_at, 
				created_at, 
				updated_at, 
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

// DeleteBroadcast deletes a broadcast from the database
func (r *broadcastRepository) DeleteBroadcast(ctx context.Context, workspaceID, id string) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		DELETE FROM broadcasts
		WHERE id = $1 AND workspace_id = $2
	`

	result, err := workspaceDB.ExecContext(ctx, query, id, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.ErrBroadcastNotFound{ID: id}
	}

	return nil
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
		&broadcast.TrackingEnabled,
		&broadcast.UTMParameters,
		&broadcast.Metadata,
		&broadcast.TotalSent,
		&broadcast.TotalDelivered,
		&broadcast.TotalBounced,
		&broadcast.TotalComplained,
		&broadcast.TotalFailed,
		&broadcast.TotalOpens,
		&broadcast.TotalClicks,
		&broadcast.WinningVariation,
		&broadcast.TestSentAt,
		&broadcast.WinnerSentAt,
		&broadcast.CreatedAt,
		&broadcast.UpdatedAt,
		&broadcast.StartedAt,
		&broadcast.CompletedAt,
		&broadcast.CancelledAt,
	)

	if err != nil {
		return nil, err
	}

	return broadcast, nil
}
