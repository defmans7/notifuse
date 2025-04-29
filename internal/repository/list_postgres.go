package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type listRepository struct {
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
		INSERT INTO lists (id, name, is_double_optin, is_public, description, 
		                   double_optin_template, welcome_template, unsubscribe_template,
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = workspaceDB.ExecContext(ctx, query,
		list.ID,
		list.Name,
		list.IsDoubleOptin,
		list.IsPublic,
		list.Description,
		list.DoubleOptInTemplate,
		list.WelcomeTemplate,
		list.UnsubscribeTemplate,
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
		SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
		total_unsubscribed, total_bounced, total_complained, double_optin_template, 
		welcome_template, unsubscribe_template, created_at, updated_at, deleted_at
		FROM lists
		WHERE id = $1 AND deleted_at IS NULL
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
		SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
		total_unsubscribed, total_bounced, total_complained, double_optin_template, 
		welcome_template, unsubscribe_template, created_at, updated_at, deleted_at
		FROM lists
		WHERE deleted_at IS NULL
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
		SET name = $1, is_double_optin = $2, is_public = $3, description = $4, updated_at = $5,
		    double_optin_template = $6, welcome_template = $7, unsubscribe_template = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	result, err := workspaceDB.ExecContext(ctx, query,
		list.Name,
		list.IsDoubleOptin,
		list.IsPublic,
		list.Description,
		list.UpdatedAt,
		list.DoubleOptInTemplate,
		list.WelcomeTemplate,
		list.UnsubscribeTemplate,
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
		return &domain.ErrListNotFound{Message: "list not found or already deleted"}
	}

	return nil
}

func (r *listRepository) DeleteList(ctx context.Context, workspaceID string, id string) error {

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Start a transaction to ensure both list and contact_list entries are deleted together
	tx, err := workspaceDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback - it will be a no-op if Commit() is called
	defer tx.Rollback()

	now := time.Now().UTC()

	// First soft delete the list
	listQuery := `UPDATE lists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := tx.ExecContext(ctx, listQuery, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrListNotFound{Message: "list not found or already deleted"}
	}

	// Then soft delete all related contact_list entries
	contactListQuery := `UPDATE contact_lists SET deleted_at = $1 WHERE list_id = $2 AND deleted_at IS NULL`
	_, err = tx.ExecContext(ctx, contactListQuery, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete contact list entries: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *listRepository) IncrementTotal(ctx context.Context, workspaceID string, listID string, totalType domain.ContactListTotalType) error {
	var columnName string
	switch totalType {
	case domain.TotalTypePending:
		columnName = "total_pending"
	case domain.TotalTypeUnsubscribed:
		columnName = "total_unsubscribed"
	case domain.TotalTypeBounced:
		columnName = "total_bounced"
	case domain.TotalTypeComplained:
		columnName = "total_complained"
	case domain.TotalTypeActive:
		columnName = "total_active"
	default:
		return fmt.Errorf("invalid total type: %s", totalType)
	}

	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := fmt.Sprintf("UPDATE lists SET %s = %s + 1 WHERE id = $1 AND deleted_at IS NULL", columnName, columnName)
	result, err := workspaceDB.ExecContext(ctx, query, listID)
	if err != nil {
		return fmt.Errorf("failed to increment total: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrListNotFound{Message: "list not found or already deleted"}
	}

	return nil
}

func (r *listRepository) DecrementTotal(ctx context.Context, workspaceID string, listID string, totalType domain.ContactListTotalType) error {
	var columnName string
	switch totalType {
	case domain.TotalTypePending:
		columnName = "total_pending"
	case domain.TotalTypeUnsubscribed:
		columnName = "total_unsubscribed"
	case domain.TotalTypeBounced:
		columnName = "total_bounced"
	case domain.TotalTypeComplained:
		columnName = "total_complained"
	case domain.TotalTypeActive:
		columnName = "total_active"
	default:
		return fmt.Errorf("invalid total type: %s", totalType)
	}

	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := fmt.Sprintf("UPDATE lists SET %s = GREATEST(%s - 1, 0) WHERE id = $1 AND deleted_at IS NULL", columnName, columnName)
	result, err := workspaceDB.ExecContext(ctx, query, listID)
	if err != nil {
		return fmt.Errorf("failed to decrement total: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrListNotFound{Message: "list not found or already deleted"}
	}

	return nil
}
