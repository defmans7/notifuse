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

// Create adds a new task
func (r *TaskRepository) Create(ctx context.Context, workspace string, task *domain.Task) error {
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
			max_runtime, max_retries, retry_count, retry_interval
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	_, err = r.systemDB.ExecContext(
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
	)

	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, workspace, id string) (*domain.Task, error) {
	query := `
		SELECT
			id, workspace_id, type, status, progress, state,
			error_message, created_at, updated_at, last_run_at,
			completed_at, next_run_after, timeout_after,
			max_runtime, max_retries, retry_count, retry_interval
		FROM tasks
		WHERE id = $1 AND workspace_id = $2
	`

	var task domain.Task
	var stateJSON []byte
	var lastRunAt, completedAt, nextRunAfter, timeoutAfter sql.NullTime

	err := r.systemDB.QueryRowContext(ctx, query, id, workspace).Scan(
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
			retry_interval = $16
		WHERE id = $1 AND workspace_id = $2
	`

	result, err := r.systemDB.ExecContext(
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

	result, err := r.systemDB.ExecContext(ctx, sqlQuery, args...)
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

	result, err := r.systemDB.ExecContext(ctx, sqlQuery, args...)
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

	result, err := r.systemDB.ExecContext(ctx, sqlQuery, args...)
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
	// Get current task to check retry counts
	task, err := r.Get(ctx, workspace, id)
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

	result, err := r.systemDB.ExecContext(ctx, sqlQuery, args...)
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

// MarkAsPaused marks a task as paused (e.g., due to timeout)
func (r *TaskRepository) MarkAsPaused(ctx context.Context, workspace, id string, nextRunAfter time.Time) error {
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

	result, err := r.systemDB.ExecContext(ctx, sqlQuery, args...)
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

// CreateSubtasks creates the specified number of subtasks for a parent task
func (r *TaskRepository) CreateSubtasks(ctx context.Context, workspace string, taskID string, count int) ([]*domain.Subtask, error) {
	// First, check that the parent task exists and update its subtask count
	task, err := r.Get(ctx, workspace, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent task: %w", err)
	}

	// Update the task to indicate it has parallel subtasks
	task.ParallelSubtasks = true
	task.SubtaskCount = count
	task.CompletedSubtasks = 0
	task.FailedSubtasks = 0

	// Update the parent task
	if err := r.Update(ctx, workspace, task); err != nil {
		return nil, fmt.Errorf("failed to update parent task: %w", err)
	}

	// Create subtasks
	now := time.Now().UTC()
	subtasks := make([]*domain.Subtask, count)

	// Begin transaction to ensure all subtasks are created
	tx, err := r.systemDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert query
	query := `
		INSERT INTO task_subtasks (
			id, parent_task_id, status, progress, state,
			error_message, created_at, updated_at, started_at,
			completed_at, timeout_after, index, total
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	// Create each subtask
	for i := 0; i < count; i++ {
		subtaskID := uuid.New().String()

		// Initialize subtask
		subtask := &domain.Subtask{
			ID:           subtaskID,
			ParentTaskID: taskID,
			Status:       domain.SubtaskStatusPending,
			Progress:     0,
			State:        domain.TaskState{},
			ErrorMessage: "",
			CreatedAt:    now,
			UpdatedAt:    now,
			Index:        i,
			Total:        count,
		}

		// Convert state to JSON
		stateJSON, err := json.Marshal(subtask.State)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state: %w", err)
		}

		// Execute insert
		_, err = tx.ExecContext(
			ctx,
			query,
			subtask.ID,
			subtask.ParentTaskID,
			subtask.Status,
			subtask.Progress,
			stateJSON,
			subtask.ErrorMessage,
			subtask.CreatedAt,
			subtask.UpdatedAt,
			subtask.StartedAt,
			subtask.CompletedAt,
			subtask.TimeoutAfter,
			subtask.Index,
			subtask.Total,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to insert subtask: %w", err)
		}

		subtasks[i] = subtask
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return subtasks, nil
}

// GetSubtask retrieves a specific subtask by ID
func (r *TaskRepository) GetSubtask(ctx context.Context, subtaskID string) (*domain.Subtask, error) {
	query := `
		SELECT
			id, parent_task_id, status, progress, state,
			error_message, created_at, updated_at, started_at,
			completed_at, timeout_after, index, total
		FROM task_subtasks
		WHERE id = $1
	`

	var subtask domain.Subtask
	var stateJSON []byte
	var startedAt, completedAt, timeoutAfter sql.NullTime

	err := r.systemDB.QueryRowContext(ctx, query, subtaskID).Scan(
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
		&timeoutAfter,
		&subtask.Index,
		&subtask.Total,
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
	if timeoutAfter.Valid {
		subtask.TimeoutAfter = &timeoutAfter.Time
	}

	// Unmarshal state
	if stateJSON != nil {
		if err := json.Unmarshal(stateJSON, &subtask.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &subtask, nil
}

// GetSubtasks retrieves all subtasks for a parent task
func (r *TaskRepository) GetSubtasks(ctx context.Context, taskID string) ([]*domain.Subtask, error) {
	query := `
		SELECT
			id, parent_task_id, status, progress, state,
			error_message, created_at, updated_at, started_at,
			completed_at, timeout_after, index, total
		FROM task_subtasks
		WHERE parent_task_id = $1
		ORDER BY index ASC
	`

	rows, err := r.systemDB.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subtasks: %w", err)
	}
	defer rows.Close()

	var subtasks []*domain.Subtask
	for rows.Next() {
		var subtask domain.Subtask
		var stateJSON []byte
		var startedAt, completedAt, timeoutAfter sql.NullTime

		err := rows.Scan(
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
			&timeoutAfter,
			&subtask.Index,
			&subtask.Total,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subtask row: %w", err)
		}

		// Handle nullable times
		if startedAt.Valid {
			subtask.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			subtask.CompletedAt = &completedAt.Time
		}
		if timeoutAfter.Valid {
			subtask.TimeoutAfter = &timeoutAfter.Time
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
		WHERE id = $1 AND status = $5
	`

	result, err := r.systemDB.ExecContext(
		ctx,
		query,
		subtaskID,
		progress,
		stateJSON,
		now,
		domain.SubtaskStatusRunning,
	)

	if err != nil {
		return fmt.Errorf("failed to update subtask progress: %w", err)
	}

	// Check if the subtask was found and is running
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found or not in running state")
	}

	return nil
}

// CompleteSubtask marks a subtask as completed
func (r *TaskRepository) CompleteSubtask(ctx context.Context, subtaskID string) error {
	// Start transaction
	tx, err := r.systemDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get subtask to find parent task
	var parentTaskID string
	err = tx.QueryRowContext(ctx, "SELECT parent_task_id FROM task_subtasks WHERE id = $1", subtaskID).Scan(&parentTaskID)
	if err != nil {
		return fmt.Errorf("failed to get parent task ID: %w", err)
	}

	// Mark subtask as completed
	now := time.Now().UTC()
	query := `
		UPDATE task_subtasks
		SET
			status = $2,
			progress = 100,
			updated_at = $3,
			completed_at = $3
		WHERE id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		subtaskID,
		domain.SubtaskStatusCompleted,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to mark subtask as completed: %w", err)
	}

	// Check if the subtask was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found")
	}

	// Update parent task completed subtasks count
	_, err = tx.ExecContext(
		ctx,
		"UPDATE tasks SET completed_subtasks = completed_subtasks + 1, updated_at = $2 WHERE id = $1",
		parentTaskID,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update parent task: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FailSubtask marks a subtask as failed
func (r *TaskRepository) FailSubtask(ctx context.Context, subtaskID string, errorMessage string) error {
	// Start transaction
	tx, err := r.systemDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get subtask to find parent task
	var parentTaskID string
	err = tx.QueryRowContext(ctx, "SELECT parent_task_id FROM task_subtasks WHERE id = $1", subtaskID).Scan(&parentTaskID)
	if err != nil {
		return fmt.Errorf("failed to get parent task ID: %w", err)
	}

	// Mark subtask as failed
	now := time.Now().UTC()
	query := `
		UPDATE task_subtasks
		SET
			status = $2,
			error_message = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		subtaskID,
		domain.SubtaskStatusFailed,
		errorMessage,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to mark subtask as failed: %w", err)
	}

	// Check if the subtask was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subtask not found")
	}

	// Update parent task failed subtasks count
	_, err = tx.ExecContext(
		ctx,
		"UPDATE tasks SET failed_subtasks = failed_subtasks + 1, updated_at = $2 WHERE id = $1",
		parentTaskID,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update parent task: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateTaskProgressFromSubtasks updates a task's progress based on its subtasks
func (r *TaskRepository) UpdateTaskProgressFromSubtasks(ctx context.Context, workspace, taskID string) error {
	// Start transaction
	tx, err := r.systemDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get the task
	task, err := r.Get(ctx, workspace, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get subtasks
	subtasks, err := r.GetSubtasks(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	// If no subtasks, nothing to do
	if len(subtasks) == 0 {
		return nil
	}

	// Calculate progress
	var totalProgress float64
	for _, subtask := range subtasks {
		totalProgress += subtask.Progress
	}

	// Calculate average progress
	averageProgress := totalProgress / float64(len(subtasks))

	// Check if all subtasks are completed or failed
	completedCount := task.CompletedSubtasks
	failedCount := task.FailedSubtasks
	totalCount := task.SubtaskCount

	// Update task progress
	now := time.Now().UTC()
	query := `
		UPDATE tasks
		SET
			progress = $2,
			updated_at = $3
		WHERE id = $1 AND workspace_id = $4
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		taskID,
		averageProgress,
		now,
		workspace,
	)

	if err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// If all subtasks are completed or failed, update task status accordingly
	if completedCount+failedCount >= totalCount {
		var status domain.TaskStatus
		if failedCount > 0 {
			// If any subtasks failed, mark the task as failed with an error message
			status = domain.TaskStatusFailed
			_, err = tx.ExecContext(
				ctx,
				`UPDATE tasks SET status = $2, error_message = $3, updated_at = $4, completed_at = $4 
				 WHERE id = $1 AND workspace_id = $5`,
				taskID,
				status,
				fmt.Sprintf("%d of %d subtasks failed", failedCount, totalCount),
				now,
				workspace,
			)
		} else {
			// All subtasks completed successfully
			status = domain.TaskStatusCompleted
			_, err = tx.ExecContext(
				ctx,
				`UPDATE tasks SET status = $2, updated_at = $3, completed_at = $3, progress = 100
				 WHERE id = $1 AND workspace_id = $4`,
				taskID,
				status,
				now,
				workspace,
			)
		}

		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
