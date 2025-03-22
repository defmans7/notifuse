package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type listRepository struct {
	db *sql.DB
}

// NewListRepository creates a new PostgreSQL list repository
func NewListRepository(db *sql.DB) domain.ListRepository {
	return &listRepository{db: db}
}

func (r *listRepository) CreateList(ctx context.Context, list *domain.List) error {
	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	query := `
		INSERT INTO lists (id, name, type, is_double_optin, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		list.ID,
		list.Name,
		list.Type,
		list.IsDoubleOptin,
		list.Description,
		list.CreatedAt,
		list.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}
	return nil
}

func (r *listRepository) GetListByID(ctx context.Context, id string) (*domain.List, error) {
	query := `
		SELECT id, name, type, is_double_optin, description, created_at, updated_at
		FROM lists
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	list, err := domain.ScanList(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrListNotFound{Message: "list not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return list, nil
}

func (r *listRepository) GetLists(ctx context.Context) ([]*domain.List, error) {
	query := `
		SELECT id, name, type, is_double_optin, description, created_at, updated_at
		FROM lists
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
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

func (r *listRepository) UpdateList(ctx context.Context, list *domain.List) error {
	list.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE lists
		SET name = $1, type = $2, is_double_optin = $3, description = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		list.Name,
		list.Type,
		list.IsDoubleOptin,
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

func (r *listRepository) DeleteList(ctx context.Context, id string) error {
	query := `DELETE FROM lists WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
