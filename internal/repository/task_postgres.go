package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/google/uuid"
)

// TaskRepository implements the domain.TaskRepository interface using PostgreSQL
type TaskRepository struct {
	systemDB *sql.DB
}

// NewTaskRepository creates a new TaskRepository instance
func NewTaskRepository(db *sql.DB) domain.TaskRepository {
	return &TaskRepository{
		systemDB: db,
	}
}

// WithTransaction executes a function within a transaction
func (r *TaskRepository) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	// Begin a transaction
	tx, err := r.systemDB.BeginTx(ctx, nil)
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

// Create adds a new task
func (r *TaskRepository) Create(ctx context.Context, workspace string, task *domain.Task) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.CreateTx(ctx, tx, workspace, task)
	})
}

// CreateTx adds a new task within a transaction
func (r *TaskRepository) CreateTx(ctx context.Context, tx *sql.Tx, workspace string, task *domain.Task) error {
	// Generate ID if not provided
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Initialize timestamps
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Set default status if not set
	if task.Status == "" {
		task.Status = domain.TaskStatusPending
	}

	// Convert state to JSON
	stateJSON, err := json.Marshal(task.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Insert the task
	query := `
		INSERT INTO tasks (
			id, workspace_id, type, status, progress, state,
			error_message, created_at, updated_at, last_run_at,
			completed_at, next_run_after, timeout_after,
			max_runtime, max_retries, retry_count, retry_interval,
			broadcast_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		task.ID,
		workspace,
		task.Type,
		task.Status,
		task.Progress,
		stateJSON,
		task.ErrorMessage,
		task.CreatedAt,
		task.UpdatedAt,
		task.LastRunAt,
		task.CompletedAt,
		task.NextRunAfter,
		task.TimeoutAfter,
		task.MaxRuntime,
		task.MaxRetries,
		task.RetryCount,
		task.RetryInterval,
		task.BroadcastID,
	)

	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, workspace, id string) (*domain.Task, error) {
	var task *domain.Task
	var err error

	err = r.WithTransaction(ctx, func(tx *sql.Tx) error {
		task, err = r.GetTx(ctx, tx, workspace, id)
		return err
	})

	return task, err
}

// GetTx retrieves a task by ID within a transaction
func (r *TaskRepository) GetTx(ctx context.Context, tx *sql.Tx, workspace, id string) (*domain.Task, error) {
	query := `
		SELECT
			id, workspace_id, type, status, progress, state,
			error_message, created_at, updated_at, last_run_at,
			completed_at, next_run_after, timeout_after,
			max_runtime, max_retries, retry_count, retry_interval,
			broadcast_id
		FROM tasks
		WHERE id = $1 AND workspace_id = $2
	`

	var task domain.Task
	var stateJSON []byte
	var lastRunAt, completedAt, nextRunAfter, timeoutAfter sql.NullTime
	var broadcastID sql.NullString

	err := tx.QueryRowContext(ctx, query, id, workspace).Scan(
		&task.ID,
		&task.WorkspaceID,
		&task.Type,
		&task.Status,
		&task.Progress,
		&stateJSON,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
		&lastRunAt,
		&completedAt,
		&nextRunAfter,
		&timeoutAfter,
		&task.MaxRuntime,
		&task.MaxRetries,
		&task.RetryCount,
		&task.RetryInterval,
		&broadcastID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Handle nullable times
	if lastRunAt.Valid {
		task.LastRunAt = &lastRunAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if nextRunAfter.Valid {
		task.NextRunAfter = &nextRunAfter.Time
	}
	if timeoutAfter.Valid {
		task.TimeoutAfter = &timeoutAfter.Time
	}

	// Handle optional broadcast ID
	if broadcastID.Valid {
		task.BroadcastID = &broadcastID.String
	}

	// Unmarshal state
	if stateJSON != nil {
		task.State = &domain.TaskState{}
		if err := json.Unmarshal(stateJSON, task.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &task, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, workspace string, task *domain.Task) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.UpdateTx(ctx, tx, workspace, task)
	})
}

// UpdateTx updates an existing task within a transaction
func (r *TaskRepository) UpdateTx(ctx context.Context, tx *sql.Tx, workspace string, task *domain.Task) error {
	// Update timestamp
	task.UpdatedAt = time.Now().UTC()

	// Convert state to JSON
	stateJSON, err := json.Marshal(task.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Update the task
	query := `
		UPDATE tasks
		SET
			type = $3,
			status = $4,
			progress = $5,
			state = $6,
			error_message = $7,
			updated_at = $8,
			last_run_at = $9,
			completed_at = $10,
			next_run_after = $11,
			timeout_after = $12,
			max_runtime = $13,
			max_retries = $14,
			retry_count = $15,
			retry_interval = $16,
			broadcast_id = $17
		WHERE id = $1 AND workspace_id = $2
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		task.ID,
		workspace,
		task.Type,
		task.Status,
		task.Progress,
		stateJSON,
		task.ErrorMessage,
		task.UpdatedAt,
		task.LastRunAt,
		task.CompletedAt,
		task.NextRunAfter,
		task.TimeoutAfter,
		task.MaxRuntime,
		task.MaxRetries,
		task.RetryCount,
		task.RetryInterval,
		task.BroadcastID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// Delete removes a task
func (r *TaskRepository) Delete(ctx context.Context, workspace, id string) error {
	query := `DELETE FROM tasks WHERE id = $1 AND workspace_id = $2`
	result, err := r.systemDB.ExecContext(ctx, query, id, workspace)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// List retrieves tasks with optional filtering
func (r *TaskRepository) List(ctx context.Context, workspace string, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	// First, build a query to get the total count
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Base query conditions
	countQuery := psql.Select("COUNT(*)").
		From("tasks").
		Where(sq.Eq{"workspace_id": workspace})

	// Apply filters
	if len(filter.Status) > 0 {
		// Convert domain.TaskStatus to strings for SQL
		statusStrings := make([]string, len(filter.Status))
		for i, s := range filter.Status {
			statusStrings[i] = string(s)
		}
		countQuery = countQuery.Where(sq.Eq{"status": statusStrings})
	}

	if len(filter.Type) > 0 {
		countQuery = countQuery.Where(sq.Eq{"type": filter.Type})
	}

	if filter.CreatedAfter != nil {
		countQuery = countQuery.Where(sq.GtOrEq{"created_at": filter.CreatedAfter})
	}

	if filter.CreatedBefore != nil {
		countQuery = countQuery.Where(sq.LtOrEq{"created_at": filter.CreatedBefore})
	}

	// Execute count query
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var totalCount int
	err = r.systemDB.QueryRowContext(ctx, countSQL, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Build the data query
	dataQuery := psql.Select(
		"id", "workspace_id", "type", "status", "progress", "state",
		"error_message", "created_at", "updated_at", "last_run_at",
		"completed_at", "next_run_after", "timeout_after",
		"max_runtime", "max_retries", "retry_count", "retry_interval",
		"broadcast_id",
	).
		From("tasks").
		Where(sq.Eq{"workspace_id": workspace})

	// Apply the same filters
	if len(filter.Status) > 0 {
		statusStrings := make([]string, len(filter.Status))
		for i, s := range filter.Status {
			statusStrings[i] = string(s)
		}
		dataQuery = dataQuery.Where(sq.Eq{"status": statusStrings})
	}

	if len(filter.Type) > 0 {
		dataQuery = dataQuery.Where(sq.Eq{"type": filter.Type})
	}

	if filter.CreatedAfter != nil {
		dataQuery = dataQuery.Where(sq.GtOrEq{"created_at": filter.CreatedAfter})
	}

	if filter.CreatedBefore != nil {
		dataQuery = dataQuery.Where(sq.LtOrEq{"created_at": filter.CreatedBefore})
	}

	// Add order, limit and offset
	dataQuery = dataQuery.
		OrderBy("created_at DESC").
		Limit(uint64(filter.Limit))

	if filter.Offset > 0 {
		dataQuery = dataQuery.Offset(uint64(filter.Offset))
	}

	// Execute data query
	dataSql, dataArgs, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build tasks query: %w", err)
	}

	rows, err := r.systemDB.QueryContext(ctx, dataSql, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		var stateJSON []byte
		var lastRunAt, completedAt, nextRunAfter, timeoutAfter sql.NullTime
		var broadcastID sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.WorkspaceID,
			&task.Type,
			&task.Status,
			&task.Progress,
			&stateJSON,
			&task.ErrorMessage,
			&task.CreatedAt,
			&task.UpdatedAt,
			&lastRunAt,
			&completedAt,
			&nextRunAfter,
			&timeoutAfter,
			&task.MaxRuntime,
			&task.MaxRetries,
			&task.RetryCount,
			&task.RetryInterval,
			&broadcastID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task row: %w", err)
		}

		// Handle nullable times
		if lastRunAt.Valid {
			task.LastRunAt = &lastRunAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if nextRunAfter.Valid {
			task.NextRunAfter = &nextRunAfter.Time
		}
		if timeoutAfter.Valid {
			task.TimeoutAfter = &timeoutAfter.Time
		}

		// Handle optional broadcast ID
		if broadcastID.Valid {
			task.BroadcastID = &broadcastID.String
		}

		// Unmarshal state
		if stateJSON != nil {
			task.State = &domain.TaskState{}
			if err := json.Unmarshal(stateJSON, task.State); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal state: %w", err)
			}
		}

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, totalCount, nil
}

// GetNextBatch retrieves tasks that are ready to be processed
func (r *TaskRepository) GetNextBatch(ctx context.Context, limit int) ([]*domain.Task, error) {
	now := time.Now().UTC()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// We want tasks that are:
	// 1. Pending and ready to run (next_run_after is null or in the past)
	// 2. Paused but ready to resume (next_run_after in the past)
	// 3. Running but have timed out (timeout_after in the past)
	query := psql.Select(
		"id", "workspace_id", "type", "status", "progress", "state",
		"error_message", "created_at", "updated_at", "last_run_at",
		"completed_at", "next_run_after", "timeout_after",
		"max_runtime", "max_retries", "retry_count", "retry_interval",
		"broadcast_id",
	).
		From("tasks").
		Where(sq.Or{
			sq.And{
				sq.Eq{"status": string(domain.TaskStatusPending)},
				sq.Or{
					sq.Eq{"next_run_after": nil},
					sq.LtOrEq{"next_run_after": now},
				},
			},
			sq.And{
				sq.Eq{"status": string(domain.TaskStatusPaused)},
				sq.LtOrEq{"next_run_after": now},
			},
			sq.And{
				sq.Eq{"status": string(domain.TaskStatusRunning)},
				sq.LtOrEq{"timeout_after": now},
			},
		}).
		OrderBy("next_run_after NULLS FIRST, created_at").
		Limit(uint64(limit)).
		Suffix("FOR UPDATE SKIP LOCKED")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build next batch query: %w", err)
	}

	rows, err := r.systemDB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get next batch of tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		var stateJSON []byte
		var lastRunAt, completedAt, nextRunAfter, timeoutAfter sql.NullTime
		var broadcastID sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.WorkspaceID,
			&task.Type,
			&task.Status,
			&task.Progress,
			&stateJSON,
			&task.ErrorMessage,
			&task.CreatedAt,
			&task.UpdatedAt,
			&lastRunAt,
			&completedAt,
			&nextRunAfter,
			&timeoutAfter,
			&task.MaxRuntime,
			&task.MaxRetries,
			&task.RetryCount,
			&task.RetryInterval,
			&broadcastID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		// Handle nullable times
		if lastRunAt.Valid {
			task.LastRunAt = &lastRunAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if nextRunAfter.Valid {
			task.NextRunAfter = &nextRunAfter.Time
		}
		if timeoutAfter.Valid {
			task.TimeoutAfter = &timeoutAfter.Time
		}

		// Handle optional broadcast ID
		if broadcastID.Valid {
			task.BroadcastID = &broadcastID.String
		}

		// Unmarshal state
		if stateJSON != nil {
			task.State = &domain.TaskState{}
			if err := json.Unmarshal(stateJSON, task.State); err != nil {
				return nil, fmt.Errorf("failed to unmarshal state: %w", err)
			}
		}

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// SaveState saves the current state of a running task
func (r *TaskRepository) SaveState(ctx context.Context, workspace, id string, progress float64, state *domain.TaskState) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.SaveStateTx(ctx, tx, workspace, id, progress, state)
	})
}

// SaveStateTx saves the current state of a running task within a transaction
func (r *TaskRepository) SaveStateTx(ctx context.Context, tx *sql.Tx, workspace, id string, progress float64, state *domain.TaskState) error {
	// Convert state to JSON
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	now := time.Now().UTC()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update("tasks").
		Set("progress", progress).
		Set("state", stateJSON).
		Set("updated_at", now).
		Where(sq.And{
			sq.Eq{
				"id":           id,
				"workspace_id": workspace,
				"status":       domain.TaskStatusRunning,
			},
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to save task state: %w", err)
	}

	// Check if the task was found and is running
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found or not in running state")
	}

	return nil
}

// MarkAsRunning marks a task as running and sets timeout
func (r *TaskRepository) MarkAsRunning(ctx context.Context, workspace, id string, timeoutAfter time.Time) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.MarkAsRunningTx(ctx, tx, workspace, id, timeoutAfter)
	})
}

// MarkAsRunningTx marks a task as running and sets timeout within a transaction
func (r *TaskRepository) MarkAsRunningTx(ctx context.Context, tx *sql.Tx, workspace, id string, timeoutAfter time.Time) error {
	now := time.Now().UTC()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update("tasks").
		Set("status", domain.TaskStatusRunning).
		Set("updated_at", now).
		Set("last_run_at", now).
		Set("timeout_after", timeoutAfter).
		Where(sq.Eq{
			"id":           id,
			"workspace_id": workspace,
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// MarkAsCompleted marks a task as completed
func (r *TaskRepository) MarkAsCompleted(ctx context.Context, workspace, id string) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.MarkAsCompletedTx(ctx, tx, workspace, id)
	})
}

// MarkAsCompletedTx marks a task as completed within a transaction
func (r *TaskRepository) MarkAsCompletedTx(ctx context.Context, tx *sql.Tx, workspace, id string) error {
	now := time.Now().UTC()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update("tasks").
		Set("status", domain.TaskStatusCompleted).
		Set("progress", 100).
		Set("updated_at", now).
		Set("completed_at", now).
		Set("timeout_after", nil).
		Where(sq.Eq{
			"id":           id,
			"workspace_id": workspace,
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// MarkAsFailed marks a task as failed
func (r *TaskRepository) MarkAsFailed(ctx context.Context, workspace, id string, errorMsg string) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.MarkAsFailedTx(ctx, tx, workspace, id, errorMsg)
	})
}

// MarkAsFailedTx marks a task as failed within a transaction
func (r *TaskRepository) MarkAsFailedTx(ctx context.Context, tx *sql.Tx, workspace, id string, errorMsg string) error {
	// Get current task to check retry counts
	task, err := r.GetTx(ctx, tx, workspace, id)
	if err != nil {
		return fmt.Errorf("failed to get task for retry check: %w", err)
	}

	now := time.Now().UTC()
	newStatus := domain.TaskStatusFailed
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Handle retries if applicable
	var nextRunAfter *time.Time
	if task.RetryCount < task.MaxRetries {
		// Calculate next retry time
		retryTime := now.Add(time.Duration(task.RetryInterval) * time.Second)
		nextRunAfter = &retryTime
		newStatus = domain.TaskStatusPending
	}

	query := psql.Update("tasks").
		Set("status", newStatus).
		Set("error_message", errorMsg).
		Set("updated_at", now).
		Set("retry_count", sq.Expr("retry_count + 1")).
		Set("next_run_after", nextRunAfter).
		Set("timeout_after", nil).
		Where(sq.Eq{
			"id":           id,
			"workspace_id": workspace,
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// MarkAsPaused marks a task as paused and sets the next run time
func (r *TaskRepository) MarkAsPaused(ctx context.Context, workspace, id string, nextRunAfter time.Time) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.MarkAsPausedTx(ctx, tx, workspace, id, nextRunAfter)
	})
}

// MarkAsPausedTx marks a task as paused and sets the next run time within a transaction
func (r *TaskRepository) MarkAsPausedTx(ctx context.Context, tx *sql.Tx, workspace, id string, nextRunAfter time.Time) error {
	now := time.Now().UTC()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update("tasks").
		Set("status", domain.TaskStatusPaused).
		Set("updated_at", now).
		Set("next_run_after", nextRunAfter).
		Set("timeout_after", nil).
		Where(sq.Eq{
			"id":           id,
			"workspace_id": workspace,
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to mark task as paused: %w", err)
	}

	// Check if the task was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// CreateSubtasks creates multiple subtasks for a parent task
func (r *TaskRepository) CreateSubtasks(ctx context.Context, workspace string, taskID string, count int) ([]*domain.Subtask, error) {
	var subtasks []*domain.Subtask
	var err error

	err = r.WithTransaction(ctx, func(tx *sql.Tx) error {
		subtasks, err = r.CreateSubtasksTx(ctx, tx, workspace, taskID, count)
		return err
	})

	return subtasks, err
}

// CreateSubtasksTx creates multiple subtasks for a parent task within a transaction
func (r *TaskRepository) CreateSubtasksTx(ctx context.Context, tx *sql.Tx, workspace string, taskID string, count int) ([]*domain.Subtask, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	// Verify the parent task exists
	parentTask, err := r.GetTx(ctx, tx, workspace, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent task: %w", err)
	}

	// Create subtasks
	subtasks := make([]*domain.Subtask, count)
	now := time.Now().UTC()

	for i := 0; i < count; i++ {
		subtaskID := uuid.New().String()
		subtasks[i] = &domain.Subtask{
			ID:           subtaskID,
			ParentTaskID: taskID,
			Status:       domain.SubtaskStatusPending,
			Progress:     0,
			CreatedAt:    now,
			UpdatedAt:    now,
			State:        domain.TaskState{}, // Initialize empty state
			Index:        i,
			Total:        count,
			BroadcastID:  parentTask.BroadcastID, // Inherit broadcast ID from parent task
		}

		// Insert the subtask
		query := `
			INSERT INTO task_subtasks (
				id, parent_task_id, status, progress, state,
				error_message, created_at, updated_at, started_at, completed_at,
				index, total, broadcast_id
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
			)
		`

		// Marshal the state to JSON
		stateJSON, err := json.Marshal(subtasks[i].State)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			query,
			subtasks[i].ID,
			subtasks[i].ParentTaskID,
			subtasks[i].Status,
			subtasks[i].Progress,
			stateJSON,
			subtasks[i].ErrorMessage,
			subtasks[i].CreatedAt,
			subtasks[i].UpdatedAt,
			nil, // started_at (null)
			nil, // completed_at (null)
			subtasks[i].Index,
			subtasks[i].Total,
			subtasks[i].BroadcastID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to insert subtask: %w", err)
		}
	}

	return subtasks, nil
}

// GetSubtask retrieves a subtask by ID
func (r *TaskRepository) GetSubtask(ctx context.Context, subtaskID string) (*domain.Subtask, error) {
	var subtask *domain.Subtask
	var err error

	err = r.WithTransaction(ctx, func(tx *sql.Tx) error {
		subtask, err = r.GetSubtaskTx(ctx, tx, subtaskID)
		return err
	})

	return subtask, err
}

// GetSubtaskTx retrieves a subtask by ID within a transaction
func (r *TaskRepository) GetSubtaskTx(ctx context.Context, tx *sql.Tx, subtaskID string) (*domain.Subtask, error) {
	query := `
		SELECT
			id, parent_task_id, status, progress, state,
			error_message, created_at, updated_at, started_at, completed_at,
			index, total, broadcast_id
		FROM task_subtasks
		WHERE id = $1
		FOR UPDATE
	`

	var subtask domain.Subtask
	var stateJSON []byte
	var startedAt, completedAt sql.NullTime
	var index, total sql.NullInt32
	var broadcastID sql.NullString

	err := tx.QueryRowContext(ctx, query, subtaskID).Scan(
		&subtask.ID,
		&subtask.ParentTaskID,
		&subtask.Status,
		&subtask.Progress,
		&stateJSON,
		&subtask.ErrorMessage,
		&subtask.CreatedAt,
		&subtask.UpdatedAt,
		&startedAt,
		&completedAt,
		&index,
		&total,
		&broadcastID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("subtask not found")
		}
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}

	// Handle nullable times
	if startedAt.Valid {
		subtask.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		subtask.CompletedAt = &completedAt.Time
	}

	// Handle nullable integers
	if index.Valid {
		subtask.Index = int(index.Int32)
	}
	if total.Valid {
		subtask.Total = int(total.Int32)
	}

	// Handle optional broadcast ID
	if broadcastID.Valid {
		subtask.BroadcastID = &broadcastID.String
	}

	// Unmarshal state
	if stateJSON != nil {
		if err := json.Unmarshal(stateJSON, &subtask.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &subtask, nil
}

// GetSubtasks retrieves all subtasks for a task
func (r *TaskRepository) GetSubtasks(ctx context.Context, taskID string) ([]*domain.Subtask, error) {
	var subtasks []*domain.Subtask
	var err error

	err = r.WithTransaction(ctx, func(tx *sql.Tx) error {
		subtasks, err = r.GetSubtasksTx(ctx, tx, taskID)
		return err
	})

	return subtasks, err
}

// GetSubtasksTx retrieves all subtasks for a task within a transaction
func (r *TaskRepository) GetSubtasksTx(ctx context.Context, tx *sql.Tx, taskID string) ([]*domain.Subtask, error) {
	query := `
		SELECT
			id, parent_task_id, status, progress, state,
			error_message, created_at, updated_at, started_at, completed_at,
			index, total, broadcast_id
		FROM task_subtasks
		WHERE parent_task_id = $1
		ORDER BY created_at
	`

	rows, err := tx.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subtasks: %w", err)
	}
	defer rows.Close()

	var subtasks []*domain.Subtask

	for rows.Next() {
		var subtask domain.Subtask
		var stateJSON []byte
		var startedAt, completedAt sql.NullTime
		var index, total sql.NullInt32
		var broadcastID sql.NullString

		if err := rows.Scan(
			&subtask.ID,
			&subtask.ParentTaskID,
			&subtask.Status,
			&subtask.Progress,
			&stateJSON,
			&subtask.ErrorMessage,
			&subtask.CreatedAt,
			&subtask.UpdatedAt,
			&startedAt,
			&completedAt,
			&index,
			&total,
			&broadcastID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subtask: %w", err)
		}

		// Handle nullable times
		if startedAt.Valid {
			subtask.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			subtask.CompletedAt = &completedAt.Time
		}

		// Handle nullable integers
		if index.Valid {
			subtask.Index = int(index.Int32)
		}
		if total.Valid {
			subtask.Total = int(total.Int32)
		}

		// Handle optional broadcast ID
		if broadcastID.Valid {
			subtask.BroadcastID = &broadcastID.String
		}

		// Unmarshal state
		if stateJSON != nil {
			if err := json.Unmarshal(stateJSON, &subtask.State); err != nil {
				return nil, fmt.Errorf("failed to unmarshal state: %w", err)
			}
		}

		subtasks = append(subtasks, &subtask)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subtask rows: %w", err)
	}

	return subtasks, nil
}

// UpdateSubtaskProgress updates the progress and state of a subtask
func (r *TaskRepository) UpdateSubtaskProgress(ctx context.Context, subtaskID string, progress float64, state domain.TaskState) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.UpdateSubtaskProgressTx(ctx, tx, subtaskID, progress, state)
	})
}

// UpdateSubtaskProgressTx updates the progress and state of a subtask within a transaction
func (r *TaskRepository) UpdateSubtaskProgressTx(ctx context.Context, tx *sql.Tx, subtaskID string, progress float64, state domain.TaskState) error {
	// Convert state to JSON
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	now := time.Now().UTC()
	query := `
		UPDATE task_subtasks
		SET
			progress = $2,
			state = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, subtaskID, progress, stateJSON, now)
	if err != nil {
		return fmt.Errorf("failed to update subtask progress: %w", err)
	}

	// Check if the subtask was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found")
	}

	return nil
}

// CompleteSubtask marks a subtask as completed
func (r *TaskRepository) CompleteSubtask(ctx context.Context, subtaskID string) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.CompleteSubtaskTx(ctx, tx, subtaskID)
	})
}

// CompleteSubtaskTx marks a subtask as completed within a transaction
func (r *TaskRepository) CompleteSubtaskTx(ctx context.Context, tx *sql.Tx, subtaskID string) error {
	now := time.Now().UTC()
	query := `
		UPDATE task_subtasks
		SET
			status = $2,
			progress = 100,
			updated_at = $3,
			completed_at = $4
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, subtaskID, domain.SubtaskStatusCompleted, now, now)
	if err != nil {
		return fmt.Errorf("failed to complete subtask: %w", err)
	}

	// Check if the subtask was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found")
	}

	return nil
}

// FailSubtask marks a subtask as failed
func (r *TaskRepository) FailSubtask(ctx context.Context, subtaskID string, errorMessage string) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.FailSubtaskTx(ctx, tx, subtaskID, errorMessage)
	})
}

// FailSubtaskTx marks a subtask as failed within a transaction
func (r *TaskRepository) FailSubtaskTx(ctx context.Context, tx *sql.Tx, subtaskID string, errorMessage string) error {
	now := time.Now().UTC()
	query := `
		UPDATE task_subtasks
		SET
			status = $2,
			error_message = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, subtaskID, domain.SubtaskStatusFailed, errorMessage, now)
	if err != nil {
		return fmt.Errorf("failed to fail subtask: %w", err)
	}

	// Check if the subtask was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found")
	}

	return nil
}

// UpdateTaskProgressFromSubtasks recalculates and updates the parent task's progress based on subtask progress
func (r *TaskRepository) UpdateTaskProgressFromSubtasks(ctx context.Context, workspace, taskID string) error {
	return r.WithTransaction(ctx, func(tx *sql.Tx) error {
		return r.UpdateTaskProgressFromSubtasksTx(ctx, tx, workspace, taskID)
	})
}

// UpdateTaskProgressFromSubtasksTx recalculates and updates the parent task's progress based on subtask progress within a transaction
func (r *TaskRepository) UpdateTaskProgressFromSubtasksTx(ctx context.Context, tx *sql.Tx, workspace, taskID string) error {
	// Get all subtasks for the task
	subtasks, err := r.GetSubtasksTx(ctx, tx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		return nil // No subtasks, nothing to update
	}

	// Calculate average progress
	var totalProgress float64
	var completedCount int
	var failedCount int

	for _, subtask := range subtasks {
		totalProgress += subtask.Progress
		if subtask.Status == domain.SubtaskStatusCompleted {
			completedCount++
		} else if subtask.Status == domain.SubtaskStatusFailed {
			failedCount++
		}
	}

	averageProgress := totalProgress / float64(len(subtasks))

	// Update the parent task
	now := time.Now().UTC()
	var status domain.TaskStatus

	// Determine task status based on subtask completion
	if completedCount == len(subtasks) {
		// All subtasks completed, mark task as completed
		status = domain.TaskStatusCompleted

		query := `
			UPDATE tasks
			SET
				status = $3,
				progress = 100,
				updated_at = $4,
				completed_at = $4
			WHERE id = $1 AND workspace_id = $2
		`

		result, err := tx.ExecContext(ctx, query, taskID, workspace, status, now)
		if err != nil {
			return fmt.Errorf("failed to update task progress: %w", err)
		}

		// Check if the task was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("task not found")
		}
	} else if failedCount > 0 && failedCount+completedCount == len(subtasks) {
		// All subtasks are either completed or failed, and at least one failed
		// Mark task as failed
		status = domain.TaskStatusFailed

		query := `
			UPDATE tasks
			SET
				status = $3,
				progress = $4,
				updated_at = $5,
				error_message = $6
			WHERE id = $1 AND workspace_id = $2
		`

		result, err := tx.ExecContext(
			ctx,
			query,
			taskID,
			workspace,
			status,
			averageProgress,
			now,
			fmt.Sprintf("%d of %d subtasks failed", failedCount, len(subtasks)),
		)
		if err != nil {
			return fmt.Errorf("failed to update task progress: %w", err)
		}

		// Check if the task was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("task not found")
		}
	} else {
		// Some subtasks still in progress
		query := `
			UPDATE tasks
			SET
				progress = $3,
				updated_at = $4
			WHERE id = $1 AND workspace_id = $2
		`

		result, err := tx.ExecContext(ctx, query, taskID, workspace, averageProgress, now)
		if err != nil {
			return fmt.Errorf("failed to update task progress: %w", err)
		}

		// Check if the task was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("task not found")
		}
	}

	return nil
}

// GetTaskByBroadcastID retrieves a task associated with a specific broadcast ID
func (r *TaskRepository) GetTaskByBroadcastID(ctx context.Context, workspace, broadcastID string) (*domain.Task, error) {
	var task *domain.Task
	var err error

	err = r.WithTransaction(ctx, func(tx *sql.Tx) error {
		task, err = r.GetTaskByBroadcastIDTx(ctx, tx, workspace, broadcastID)
		return err
	})

	return task, err
}

// GetTaskByBroadcastIDTx retrieves a task by broadcast ID within a transaction
func (r *TaskRepository) GetTaskByBroadcastIDTx(ctx context.Context, tx *sql.Tx, workspace, broadcastID string) (*domain.Task, error) {
	query := `
		SELECT
			id, workspace_id, type, status, progress, state,
			error_message, created_at, updated_at, last_run_at,
			completed_at, next_run_after, timeout_after,
			max_runtime, max_retries, retry_count, retry_interval,
			broadcast_id
		FROM tasks
		WHERE workspace_id = $1 AND broadcast_id = $2
		AND type = 'send_broadcast'
		LIMIT 1
	`

	var task domain.Task
	var stateJSON []byte
	var lastRunAt, completedAt, nextRunAfter, timeoutAfter sql.NullTime
	var dbBroadcastID sql.NullString

	err := tx.QueryRowContext(ctx, query, workspace, broadcastID).Scan(
		&task.ID,
		&task.WorkspaceID,
		&task.Type,
		&task.Status,
		&task.Progress,
		&stateJSON,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
		&lastRunAt,
		&completedAt,
		&nextRunAfter,
		&timeoutAfter,
		&task.MaxRuntime,
		&task.MaxRetries,
		&task.RetryCount,
		&task.RetryInterval,
		&dbBroadcastID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found for broadcast ID %s", broadcastID)
		}
		return nil, fmt.Errorf("failed to get task by broadcast ID: %w", err)
	}

	// Handle nullable times
	if lastRunAt.Valid {
		task.LastRunAt = &lastRunAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if nextRunAfter.Valid {
		task.NextRunAfter = &nextRunAfter.Time
	}
	if timeoutAfter.Valid {
		task.TimeoutAfter = &timeoutAfter.Time
	}

	// Handle optional broadcast ID
	if dbBroadcastID.Valid {
		task.BroadcastID = &dbBroadcastID.String
	}

	// Unmarshal state
	if stateJSON != nil {
		task.State = &domain.TaskState{}
		if err := json.Unmarshal(stateJSON, task.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &task, nil
}
