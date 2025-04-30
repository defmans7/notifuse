package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *TaskRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewTaskRepository(db).(*TaskRepository)
	return db, mock, repo
}

func TestTaskRepository_WithTransaction(t *testing.T) {
	db, mock, repo := setupTaskMock(t)
	defer db.Close()

	// Setup mock expectations
	mock.ExpectBegin()
	mock.ExpectCommit()

	// Test successful transaction
	err := repo.WithTransaction(context.Background(), func(tx *sql.Tx) error {
		// Just return nil to simulate successful operation
		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test transaction with error
	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := fmt.Errorf("test error")
	err = repo.WithTransaction(context.Background(), func(tx *sql.Tx) error {
		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_CreateWithTransaction(t *testing.T) {
	db, mock, repo := setupTaskMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "test-workspace"
	task := &domain.Task{
		ID:          "task-123",
		WorkspaceID: workspace,
		Type:        "test-task",
		Status:      domain.TaskStatusPending,
		Progress:    0,
		State: &domain.TaskState{
			SendBroadcast: &domain.SendBroadcastState{
				BroadcastID: "test-broadcast",
				BatchSize:   100,
			},
		},
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		MaxRuntime:    60,
		MaxRetries:    3,
		RetryCount:    0,
		RetryInterval: 60,
		BroadcastID:   nil,
	}

	// Setup mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO tasks").
		WithArgs(
			task.ID, workspace, task.Type, task.Status, task.Progress,
			sqlmock.AnyArg(), // State JSON
			task.ErrorMessage,
			sqlmock.AnyArg(), // CreatedAt (use AnyArg to avoid timestamp precision issues)
			sqlmock.AnyArg(), // UpdatedAt
			task.LastRunAt,
			task.CompletedAt, task.NextRunAfter, task.TimeoutAfter,
			task.MaxRuntime, task.MaxRetries, task.RetryCount, task.RetryInterval,
			task.BroadcastID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Test create task with transaction
	err := repo.Create(ctx, workspace, task)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_GetWithTransaction(t *testing.T) {
	db, mock, repo := setupTaskMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "test-workspace"
	taskID := uuid.New().String()
	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"id", "workspace_id", "type", "status", "progress", "state",
		"error_message", "created_at", "updated_at", "last_run_at",
		"completed_at", "next_run_after", "timeout_after",
		"max_runtime", "max_retries", "retry_count", "retry_interval",
		"broadcast_id",
	}).AddRow(
		taskID, workspace, "test-task", domain.TaskStatusPending, 0, "{}",
		"", now, now, nil,
		nil, nil, nil,
		60, 3, 0, 60,
		nil, // broadcast_id
	)

	// Setup mock expectations
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM tasks WHERE id = .* AND workspace_id = .*").
		WithArgs(taskID, workspace).
		WillReturnRows(rows)
	mock.ExpectCommit()

	// Test get task with transaction
	task, err := repo.Get(ctx, workspace, taskID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, taskID, task.ID)
	assert.Equal(t, workspace, task.WorkspaceID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_UpdateWithTransaction(t *testing.T) {
	db, mock, repo := setupTaskMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "test-workspace"
	taskID := uuid.New().String()
	now := time.Now().UTC()

	task := &domain.Task{
		ID:          taskID,
		WorkspaceID: workspace,
		Type:        "test-task",
		Status:      domain.TaskStatusRunning,
		Progress:    50,
		State: &domain.TaskState{
			SendBroadcast: &domain.SendBroadcastState{
				BroadcastID: "test-broadcast",
				BatchSize:   100,
			},
		},
		CreatedAt:     now.Add(-1 * time.Hour),
		UpdatedAt:     now,
		MaxRuntime:    60,
		MaxRetries:    3,
		RetryCount:    0,
		RetryInterval: 60,
		BroadcastID:   nil,
	}

	// Setup mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE tasks").
		WithArgs(
			taskID, workspace, task.Type, task.Status, task.Progress,
			sqlmock.AnyArg(), // State JSON
			task.ErrorMessage, sqlmock.AnyArg(), task.LastRunAt,
			task.CompletedAt, task.NextRunAfter, task.TimeoutAfter,
			task.MaxRuntime, task.MaxRetries, task.RetryCount, task.RetryInterval,
			task.BroadcastID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Test update task with transaction
	err := repo.Update(ctx, workspace, task)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
