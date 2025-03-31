package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type listRepository struct {
	db            *sql.DB
	workspaceRepo domain.WorkspaceRepository
}

// NewListRepository creates a new PostgreSQL list repository
func NewListRepository(workspaceRepo domain.WorkspaceRepository) domain.ListRepository {
	return &listRepository{
		workspaceRepo: workspaceRepo,
	}
}

func (r *listRepository) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	query := `
		INSERT INTO lists (id, name, is_double_optin, is_public, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = workspaceDB.ExecContext(ctx, query,
		list.ID,
		list.Name,
		list.IsDoubleOptin,
		list.IsPublic,
		list.Description,
		list.CreatedAt,
		list.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}
	return nil
}

func (r *listRepository) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT id, name, is_double_optin, is_public, description, created_at, updated_at
		FROM lists
		WHERE id = $1
	`

	row := workspaceDB.QueryRowContext(ctx, query, id)
	list, err := domain.ScanList(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrListNotFound{Message: "list not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return list, nil
}

func (r *listRepository) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT id, name, is_double_optin, is_public, description, created_at, updated_at
		FROM lists
		ORDER BY created_at DESC
	`

	rows, err := workspaceDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}
	defer rows.Close()

	var lists []*domain.List
	for rows.Next() {
		list, err := domain.ScanList(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan list: %w", err)
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating list rows: %w", err)
	}

	return lists, nil
}

func (r *listRepository) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	list.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE lists
		SET name = $1, is_double_optin = $2, is_public = $3, description = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := workspaceDB.ExecContext(ctx, query,
		list.Name,
		list.IsDoubleOptin,
		list.IsPublic,
		list.Description,
		list.UpdatedAt,
		list.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrListNotFound{Message: "list not found"}
	}

	return nil
}

func (r *listRepository) DeleteList(ctx context.Context, workspaceID string, id string) error {

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `DELETE FROM lists WHERE id = $1`

	result, err := workspaceDB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrListNotFound{Message: "list not found"}
	}

	return nil
}
