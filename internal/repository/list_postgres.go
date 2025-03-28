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
func NewListRepository(db *sql.DB, workspaceRepo domain.WorkspaceRepository) domain.ListRepository {
	return &listRepository{
		db:            db,
		workspaceRepo: workspaceRepo,
	}
}

// getWorkspaceIDFromContext tries to extract the workspace ID from the context
// In a real implementation, you'd have middleware that adds this to the context
func getWorkspaceIDFromContext(ctx context.Context) (string, bool) {
	// This is a placeholder - you need to implement how the workspace ID is stored in the context
	// For example, if you have a context key for workspace ID:
	// if workspaceID, ok := ctx.Value(WorkspaceIDKey).(string); ok && workspaceID != "" {
	//     return workspaceID, true
	// }

	// For now, we'll use a hardcoded value for development
	return "default", true
}

func (r *listRepository) CreateList(ctx context.Context, list *domain.List) error {
	// Extract workspace ID from context or use a default value
	workspaceID, ok := getWorkspaceIDFromContext(ctx)
	if !ok {
		// Fall back to using the main database if no workspace ID found
		return r.createListInMainDB(ctx, list)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	query := `
		INSERT INTO lists (id, name, type, is_double_optin, is_public, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = workspaceDB.ExecContext(ctx, query,
		list.ID,
		list.Name,
		list.Type,
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

// Fallback method for main database operations
func (r *listRepository) createListInMainDB(ctx context.Context, list *domain.List) error {
	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	query := `
		INSERT INTO lists (id, name, type, is_double_optin, is_public, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		list.ID,
		list.Name,
		list.Type,
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

func (r *listRepository) GetListByID(ctx context.Context, id string) (*domain.List, error) {
	// Extract workspace ID from context or use a default value
	workspaceID, ok := getWorkspaceIDFromContext(ctx)
	if !ok {
		// Fall back to using the main database if no workspace ID found
		return r.getListByIDFromMainDB(ctx, id)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
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

// Fallback method for main database operations
func (r *listRepository) getListByIDFromMainDB(ctx context.Context, id string) (*domain.List, error) {
	query := `
		SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
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
	// Extract workspace ID from context or use a default value
	workspaceID, ok := getWorkspaceIDFromContext(ctx)
	if !ok {
		// Fall back to using the main database if no workspace ID found
		return r.getListsFromMainDB(ctx)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `
		SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
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

// Fallback method for main database operations
func (r *listRepository) getListsFromMainDB(ctx context.Context) ([]*domain.List, error) {
	query := `
		SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
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
	// Extract workspace ID from context or use a default value
	workspaceID, ok := getWorkspaceIDFromContext(ctx)
	if !ok {
		// Fall back to using the main database if no workspace ID found
		return r.updateListInMainDB(ctx, list)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	list.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE lists
		SET name = $1, type = $2, is_double_optin = $3, is_public = $4, description = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := workspaceDB.ExecContext(ctx, query,
		list.Name,
		list.Type,
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

// Fallback method for main database operations
func (r *listRepository) updateListInMainDB(ctx context.Context, list *domain.List) error {
	list.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE lists
		SET name = $1, type = $2, is_double_optin = $3, is_public = $4, description = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		list.Name,
		list.Type,
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

func (r *listRepository) DeleteList(ctx context.Context, id string) error {
	// Extract workspace ID from context or use a default value
	workspaceID, ok := getWorkspaceIDFromContext(ctx)
	if !ok {
		// Fall back to using the main database if no workspace ID found
		return r.deleteListFromMainDB(ctx, id)
	}

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

// Fallback method for main database operations
func (r *listRepository) deleteListFromMainDB(ctx context.Context, id string) error {
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
