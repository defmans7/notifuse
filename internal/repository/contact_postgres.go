package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Notifuse/notifuse/internal/domain"
)

type contactRepository struct {
	db *sql.DB
}

// NewContactRepository creates a new PostgreSQL contact repository
func NewContactRepository(db *sql.DB) domain.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) CreateContact(ctx context.Context, contact *domain.Contact) error {
	if contact.UUID == "" {
		contact.UUID = uuid.New().String()
	}
	now := time.Now().UTC()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	query := `
		INSERT INTO contacts (
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		contact.UUID,
		contact.ExternalID,
		contact.Email,
		contact.Timezone,
		contact.FirstName,
		contact.LastName,
		contact.Phone,
		contact.AddressLine1,
		contact.AddressLine2,
		contact.Country,
		contact.Postcode,
		contact.State,
		contact.JobTitle,
		contact.LifetimeValue,
		contact.OrdersCount,
		contact.LastOrderAt,
		contact.CustomString1,
		contact.CustomString2,
		contact.CustomString3,
		contact.CustomString4,
		contact.CustomString5,
		contact.CustomNumber1,
		contact.CustomNumber2,
		contact.CustomNumber3,
		contact.CustomNumber4,
		contact.CustomNumber5,
		contact.CustomDatetime1,
		contact.CustomDatetime2,
		contact.CustomDatetime3,
		contact.CustomDatetime4,
		contact.CustomDatetime5,
		contact.CreatedAt,
		contact.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	return nil
}

func (r *contactRepository) GetContactByUUID(ctx context.Context, uuid string) (*domain.Contact, error) {
	query := `
		SELECT 
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		FROM contacts
		WHERE uuid = $1
	`

	row := r.db.QueryRowContext(ctx, query, uuid)
	contact, err := domain.ScanContact(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrContactNotFound{Message: "contact not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (r *contactRepository) GetContactByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	query := `
		SELECT 
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		FROM contacts
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)
	contact, err := domain.ScanContact(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrContactNotFound{Message: "contact not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (r *contactRepository) GetContactByExternalID(ctx context.Context, externalID string) (*domain.Contact, error) {
	query := `
		SELECT 
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		FROM contacts
		WHERE external_id = $1
	`

	row := r.db.QueryRowContext(ctx, query, externalID)
	contact, err := domain.ScanContact(row)

	if err == sql.ErrNoRows {
		return nil, &domain.ErrContactNotFound{Message: "contact not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (r *contactRepository) GetContacts(ctx context.Context) ([]*domain.Contact, error) {
	query := `
		SELECT 
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		FROM contacts
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		contact, err := domain.ScanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contacts rows: %w", err)
	}

	return contacts, nil
}

func (r *contactRepository) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	contact.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE contacts
		SET 
			external_id = $1,
			email = $2,
			timezone = $3,
			first_name = $4,
			last_name = $5,
			phone = $6,
			address_line_1 = $7,
			address_line_2 = $8,
			country = $9,
			postcode = $10,
			state = $11,
			job_title = $12,
			lifetime_value = $13,
			orders_count = $14,
			last_order_at = $15,
			custom_string_1 = $16,
			custom_string_2 = $17,
			custom_string_3 = $18,
			custom_string_4 = $19,
			custom_string_5 = $20,
			custom_number_1 = $21,
			custom_number_2 = $22,
			custom_number_3 = $23,
			custom_number_4 = $24,
			custom_number_5 = $25,
			custom_datetime_1 = $26,
			custom_datetime_2 = $27,
			custom_datetime_3 = $28,
			custom_datetime_4 = $29,
			custom_datetime_5 = $30,
			updated_at = $31
		WHERE uuid = $32
	`

	result, err := r.db.ExecContext(ctx, query,
		contact.ExternalID,
		contact.Email,
		contact.Timezone,
		contact.FirstName,
		contact.LastName,
		contact.Phone,
		contact.AddressLine1,
		contact.AddressLine2,
		contact.Country,
		contact.Postcode,
		contact.State,
		contact.JobTitle,
		contact.LifetimeValue,
		contact.OrdersCount,
		contact.LastOrderAt,
		contact.CustomString1,
		contact.CustomString2,
		contact.CustomString3,
		contact.CustomString4,
		contact.CustomString5,
		contact.CustomNumber1,
		contact.CustomNumber2,
		contact.CustomNumber3,
		contact.CustomNumber4,
		contact.CustomNumber5,
		contact.CustomDatetime1,
		contact.CustomDatetime2,
		contact.CustomDatetime3,
		contact.CustomDatetime4,
		contact.CustomDatetime5,
		contact.UpdatedAt,
		contact.UUID,
	)

	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrContactNotFound{Message: "contact not found"}
	}

	return nil
}

func (r *contactRepository) DeleteContact(ctx context.Context, uuid string) error {
	query := `DELETE FROM contacts WHERE uuid = $1`

	result, err := r.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return &domain.ErrContactNotFound{Message: "contact not found"}
	}

	return nil
}

func (r *contactRepository) BatchImportContacts(ctx context.Context, contacts []*domain.Contact) error {
	// Begin a transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the insert/update query (upsert operation)
	query := `
		INSERT INTO contacts (
			uuid, external_id, email, timezone, 
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33
		)
		ON CONFLICT (uuid) DO UPDATE
		SET 
			external_id = EXCLUDED.external_id,
			email = EXCLUDED.email,
			timezone = EXCLUDED.timezone,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			phone = EXCLUDED.phone,
			address_line_1 = EXCLUDED.address_line_1,
			address_line_2 = EXCLUDED.address_line_2,
			country = EXCLUDED.country,
			postcode = EXCLUDED.postcode,
			state = EXCLUDED.state,
			job_title = EXCLUDED.job_title,
			lifetime_value = EXCLUDED.lifetime_value,
			orders_count = EXCLUDED.orders_count,
			last_order_at = EXCLUDED.last_order_at,
			custom_string_1 = EXCLUDED.custom_string_1,
			custom_string_2 = EXCLUDED.custom_string_2,
			custom_string_3 = EXCLUDED.custom_string_3,
			custom_string_4 = EXCLUDED.custom_string_4,
			custom_string_5 = EXCLUDED.custom_string_5,
			custom_number_1 = EXCLUDED.custom_number_1,
			custom_number_2 = EXCLUDED.custom_number_2,
			custom_number_3 = EXCLUDED.custom_number_3,
			custom_number_4 = EXCLUDED.custom_number_4,
			custom_number_5 = EXCLUDED.custom_number_5,
			custom_datetime_1 = EXCLUDED.custom_datetime_1,
			custom_datetime_2 = EXCLUDED.custom_datetime_2,
			custom_datetime_3 = EXCLUDED.custom_datetime_3,
			custom_datetime_4 = EXCLUDED.custom_datetime_4,
			custom_datetime_5 = EXCLUDED.custom_datetime_5,
			updated_at = EXCLUDED.updated_at
	`

	// Prepare the statement for better performance with multiple executions
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute for each contact
	for _, contact := range contacts {
		_, err = stmt.ExecContext(ctx,
			contact.UUID,
			contact.ExternalID,
			contact.Email,
			contact.Timezone,
			contact.FirstName,
			contact.LastName,
			contact.Phone,
			contact.AddressLine1,
			contact.AddressLine2,
			contact.Country,
			contact.Postcode,
			contact.State,
			contact.JobTitle,
			contact.LifetimeValue,
			contact.OrdersCount,
			contact.LastOrderAt,
			contact.CustomString1,
			contact.CustomString2,
			contact.CustomString3,
			contact.CustomString4,
			contact.CustomString5,
			contact.CustomNumber1,
			contact.CustomNumber2,
			contact.CustomNumber3,
			contact.CustomNumber4,
			contact.CustomNumber5,
			contact.CustomDatetime1,
			contact.CustomDatetime2,
			contact.CustomDatetime3,
			contact.CustomDatetime4,
			contact.CustomDatetime5,
			contact.CreatedAt,
			contact.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert contact: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
