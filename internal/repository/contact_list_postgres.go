package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type contactListRepository struct {
	db *sql.DB
}

// NewContactListRepository creates a new PostgreSQL contact list repository
func NewContactListRepository(db *sql.DB) domain.ContactListRepository {
	return &contactListRepository{db: db}
}

func (r *contactListRepository) AddContactToList(ctx context.Context, contactList *domain.ContactList) error {
	now := time.Now().UTC()
	contactList.CreatedAt = now
	contactList.UpdatedAt = now

	query := `
		INSERT INTO contact_lists (email, list_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email, list_id) DO UPDATE
		SET status = $3, updated_at = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		contactList.Email,
		contactList.ListID,
		contactList.Status,
		contactList.CreatedAt,
		contactList.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add contact to list: %w", err)
	}
	return nil
}

func (r *contactListRepository) GetContactListByIDs(ctx context.Context, email, listID string) (*domain.ContactList, error) {
	query := `
		SELECT email, list_id, status, created_at, updated_at
		FROM contact_lists
		WHERE email = $1 AND list_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, email, listID)
	contactList, err := domain.ScanContactList(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrContactListNotFound{Message: "contact list not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contact list: %w", err)
	}

	return contactList, nil
}

func (r *contactListRepository) GetContactsByListID(ctx context.Context, listID string) ([]*domain.ContactList, error) {
	query := `
		SELECT email, list_id, status, created_at, updated_at
		FROM contact_lists
		WHERE list_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts for list: %w", err)
	}
	defer rows.Close()

	var contactLists []*domain.ContactList
	for rows.Next() {
		contactList, err := domain.ScanContactList(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact list: %w", err)
		}
		contactLists = append(contactLists, contactList)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contact list rows: %w", err)
	}

	return contactLists, nil
}

func (r *contactListRepository) GetListsByEmail(ctx context.Context, email string) ([]*domain.ContactList, error) {
	query := `
		SELECT email, list_id, status, created_at, updated_at
		FROM contact_lists
		WHERE email = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get lists for contact: %w", err)
	}
	defer rows.Close()

	var contactLists []*domain.ContactList
	for rows.Next() {
		contactList, err := domain.ScanContactList(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact list: %w", err)
		}
		contactLists = append(contactLists, contactList)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contact list rows: %w", err)
	}

	return contactLists, nil
}

func (r *contactListRepository) UpdateContactListStatus(ctx context.Context, email, listID string, status domain.ContactListStatus) error {
	now := time.Now().UTC()

	query := `
		UPDATE contact_lists
		SET status = $1, updated_at = $2
		WHERE email = $3 AND list_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		status,
		now,
		email,
		listID,
	)

	if err != nil {
		return fmt.Errorf("failed to update contact list status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrContactListNotFound{Message: "contact list not found"}
	}

	return nil
}

func (r *contactListRepository) RemoveContactFromList(ctx context.Context, email, listID string) error {
	query := `DELETE FROM contact_lists WHERE email = $1 AND list_id = $2`

	result, err := r.db.ExecContext(ctx, query, email, listID)
	if err != nil {
		return fmt.Errorf("failed to remove contact from list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrContactListNotFound{Message: "contact list not found"}
	}

	return nil
}
