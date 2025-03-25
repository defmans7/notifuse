package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type contactRepository struct {
	db *sql.DB
}

// NewContactRepository creates a new PostgreSQL contact repository
func NewContactRepository(db *sql.DB) domain.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) GetContactByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	query := `
		SELECT 
			email, external_id, timezone, language,
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			custom_json_1, custom_json_2, custom_json_3, custom_json_4, custom_json_5,
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
			email, external_id, timezone, language,
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			custom_json_1, custom_json_2, custom_json_3, custom_json_4, custom_json_5,
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

func (r *contactRepository) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	// Build the base query
	baseQuery := `
		SELECT 
			email, external_id, timezone, language,
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			custom_json_1, custom_json_2, custom_json_3, custom_json_4, custom_json_5,
			created_at, updated_at
		FROM contacts
		WHERE workspace_id = $1
	`

	// Build the WHERE clause for filters
	var conditions []string
	var args []interface{}
	argIndex := 1
	args = append(args, req.WorkspaceID)

	if req.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.Email+"%")
		argIndex++
	}

	if req.ExternalID != "" {
		conditions = append(conditions, fmt.Sprintf("external_id ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.ExternalID+"%")
		argIndex++
	}

	if req.FirstName != "" {
		conditions = append(conditions, fmt.Sprintf("first_name ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.FirstName+"%")
		argIndex++
	}

	if req.LastName != "" {
		conditions = append(conditions, fmt.Sprintf("last_name ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.LastName+"%")
		argIndex++
	}

	if req.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.Phone+"%")
		argIndex++
	}

	if req.Country != "" {
		conditions = append(conditions, fmt.Sprintf("country ILIKE $%d", argIndex+1))
		args = append(args, "%"+req.Country+"%")
		argIndex++
	}

	// Add cursor condition if provided
	if req.Cursor != "" {
		// Parse the cursor timestamp
		cursorTime, err := time.Parse(time.RFC3339, req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", argIndex+1))
		args = append(args, cursorTime)
		argIndex++
	}

	// Combine conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Always order by created_at DESC for consistent pagination
	baseQuery += " ORDER BY created_at DESC"

	// Add LIMIT clause (get one extra to determine if there are more results)
	baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex+1)
	args = append(args, req.Limit+1)

	// Execute the query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
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

	// Handle pagination
	var nextCursor string
	if len(contacts) > req.Limit {
		// Remove the extra item we fetched
		contacts = contacts[:req.Limit]
		// Set the next cursor to the created_at of the last item
		nextCursor = contacts[len(contacts)-1].CreatedAt.Format(time.RFC3339)
	}

	return &domain.GetContactsResponse{
		Contacts:   contacts,
		NextCursor: nextCursor,
	}, nil
}

func (r *contactRepository) DeleteContact(ctx context.Context, email string) error {
	query := `DELETE FROM contacts WHERE email = $1`

	result, err := r.db.ExecContext(ctx, query, email)
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
	// Prepare a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if there's a panic or error

	// Prepare a statement for contact insertion
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO contacts (
			email, external_id, timezone, language,
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			custom_json_1, custom_json_2, custom_json_3, custom_json_4, custom_json_5,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39
		)
		ON CONFLICT (email) DO UPDATE SET
			external_id = EXCLUDED.external_id,
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
			custom_json_1 = EXCLUDED.custom_json_1,
			custom_json_2 = EXCLUDED.custom_json_2,
			custom_json_3 = EXCLUDED.custom_json_3,
			custom_json_4 = EXCLUDED.custom_json_4,
			custom_json_5 = EXCLUDED.custom_json_5,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute in batches
	const batchSize = 100
	for i := 0; i < len(contacts); i += batchSize {
		end := i + batchSize
		if end > len(contacts) {
			end = len(contacts)
		}

		batch := contacts[i:end]
		for _, contact := range batch {
			// Convert domain nullable types to SQL nullable types
			var firstNameSQL, lastNameSQL, phoneSQL, addressLine1SQL, addressLine2SQL sql.NullString
			var countrySQL, postcodeSQL, stateSQL, jobTitleSQL sql.NullString
			var customString1SQL, customString2SQL, customString3SQL, customString4SQL, customString5SQL sql.NullString
			var lifetimeValueSQL, ordersCountSQL sql.NullFloat64
			var customNumber1SQL, customNumber2SQL, customNumber3SQL, customNumber4SQL, customNumber5SQL sql.NullFloat64
			var lastOrderAtSQL, customDatetime1SQL, customDatetime2SQL, customDatetime3SQL, customDatetime4SQL, customDatetime5SQL sql.NullTime
			var customJSON1SQL, customJSON2SQL, customJSON3SQL, customJSON4SQL, customJSON5SQL []byte

			// String fields
			if !contact.FirstName.IsNull {
				firstNameSQL = sql.NullString{String: contact.FirstName.String, Valid: true}
			}
			if !contact.LastName.IsNull {
				lastNameSQL = sql.NullString{String: contact.LastName.String, Valid: true}
			}
			if !contact.Phone.IsNull {
				phoneSQL = sql.NullString{String: contact.Phone.String, Valid: true}
			}
			if !contact.AddressLine1.IsNull {
				addressLine1SQL = sql.NullString{String: contact.AddressLine1.String, Valid: true}
			}
			if !contact.AddressLine2.IsNull {
				addressLine2SQL = sql.NullString{String: contact.AddressLine2.String, Valid: true}
			}
			if !contact.Country.IsNull {
				countrySQL = sql.NullString{String: contact.Country.String, Valid: true}
			}
			if !contact.Postcode.IsNull {
				postcodeSQL = sql.NullString{String: contact.Postcode.String, Valid: true}
			}
			if !contact.State.IsNull {
				stateSQL = sql.NullString{String: contact.State.String, Valid: true}
			}
			if !contact.JobTitle.IsNull {
				jobTitleSQL = sql.NullString{String: contact.JobTitle.String, Valid: true}
			}

			// Custom string fields
			if !contact.CustomString1.IsNull {
				customString1SQL = sql.NullString{String: contact.CustomString1.String, Valid: true}
			}
			if !contact.CustomString2.IsNull {
				customString2SQL = sql.NullString{String: contact.CustomString2.String, Valid: true}
			}
			if !contact.CustomString3.IsNull {
				customString3SQL = sql.NullString{String: contact.CustomString3.String, Valid: true}
			}
			if !contact.CustomString4.IsNull {
				customString4SQL = sql.NullString{String: contact.CustomString4.String, Valid: true}
			}
			if !contact.CustomString5.IsNull {
				customString5SQL = sql.NullString{String: contact.CustomString5.String, Valid: true}
			}

			// Number fields
			if !contact.LifetimeValue.IsNull {
				lifetimeValueSQL = sql.NullFloat64{Float64: contact.LifetimeValue.Float64, Valid: true}
			}
			if !contact.OrdersCount.IsNull {
				ordersCountSQL = sql.NullFloat64{Float64: contact.OrdersCount.Float64, Valid: true}
			}

			// Custom number fields
			if !contact.CustomNumber1.IsNull {
				customNumber1SQL = sql.NullFloat64{Float64: contact.CustomNumber1.Float64, Valid: true}
			}
			if !contact.CustomNumber2.IsNull {
				customNumber2SQL = sql.NullFloat64{Float64: contact.CustomNumber2.Float64, Valid: true}
			}
			if !contact.CustomNumber3.IsNull {
				customNumber3SQL = sql.NullFloat64{Float64: contact.CustomNumber3.Float64, Valid: true}
			}
			if !contact.CustomNumber4.IsNull {
				customNumber4SQL = sql.NullFloat64{Float64: contact.CustomNumber4.Float64, Valid: true}
			}
			if !contact.CustomNumber5.IsNull {
				customNumber5SQL = sql.NullFloat64{Float64: contact.CustomNumber5.Float64, Valid: true}
			}

			// Datetime fields
			if !contact.LastOrderAt.IsNull {
				lastOrderAtSQL = sql.NullTime{Time: contact.LastOrderAt.Time, Valid: true}
			}

			// Custom datetime fields
			if !contact.CustomDatetime1.IsNull {
				customDatetime1SQL = sql.NullTime{Time: contact.CustomDatetime1.Time, Valid: true}
			}
			if !contact.CustomDatetime2.IsNull {
				customDatetime2SQL = sql.NullTime{Time: contact.CustomDatetime2.Time, Valid: true}
			}
			if !contact.CustomDatetime3.IsNull {
				customDatetime3SQL = sql.NullTime{Time: contact.CustomDatetime3.Time, Valid: true}
			}
			if !contact.CustomDatetime4.IsNull {
				customDatetime4SQL = sql.NullTime{Time: contact.CustomDatetime4.Time, Valid: true}
			}
			if !contact.CustomDatetime5.IsNull {
				customDatetime5SQL = sql.NullTime{Time: contact.CustomDatetime5.Time, Valid: true}
			}

			// Custom JSON fields
			if contact.CustomJSON1.Valid {
				customJSON1SQL, err = json.Marshal(contact.CustomJSON1.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON1: %w", err)
				}
			}
			if contact.CustomJSON2.Valid {
				customJSON2SQL, err = json.Marshal(contact.CustomJSON2.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON2: %w", err)
				}
			}
			if contact.CustomJSON3.Valid {
				customJSON3SQL, err = json.Marshal(contact.CustomJSON3.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON3: %w", err)
				}
			}
			if contact.CustomJSON4.Valid {
				customJSON4SQL, err = json.Marshal(contact.CustomJSON4.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON4: %w", err)
				}
			}
			if contact.CustomJSON5.Valid {
				customJSON5SQL, err = json.Marshal(contact.CustomJSON5.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON5: %w", err)
				}
			}

			_, err := stmt.ExecContext(ctx,
				contact.Email,
				contact.ExternalID,
				contact.Timezone,
				contact.Language,
				firstNameSQL,
				lastNameSQL,
				phoneSQL,
				addressLine1SQL,
				addressLine2SQL,
				countrySQL,
				postcodeSQL,
				stateSQL,
				jobTitleSQL,
				lifetimeValueSQL,
				ordersCountSQL,
				lastOrderAtSQL,
				customString1SQL,
				customString2SQL,
				customString3SQL,
				customString4SQL,
				customString5SQL,
				customNumber1SQL,
				customNumber2SQL,
				customNumber3SQL,
				customNumber4SQL,
				customNumber5SQL,
				customDatetime1SQL,
				customDatetime2SQL,
				customDatetime3SQL,
				customDatetime4SQL,
				customDatetime5SQL,
				customJSON1SQL,
				customJSON2SQL,
				customJSON3SQL,
				customJSON4SQL,
				customJSON5SQL,
				contact.CreatedAt,
				contact.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to execute statement for contact %s: %w", contact.Email, err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *contactRepository) UpsertContact(ctx context.Context, contact *domain.Contact) (bool, error) {
	// Check if contact exists first
	_, err := r.GetContactByEmail(ctx, contact.Email)
	isNew := err != nil && err.Error() == (&domain.ErrContactNotFound{Message: "contact not found"}).Error()

	// If there was an error other than "not found", return it
	if err != nil && !isNew {
		return false, fmt.Errorf("failed to check if contact exists: %w", err)
	}

	query := `
		INSERT INTO contacts (
			email, external_id, timezone, language,
			first_name, last_name, phone, address_line_1, address_line_2,
			country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at,
			custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5,
			custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5,
			custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5,
			custom_json_1, custom_json_2, custom_json_3, custom_json_4, custom_json_5,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39
		)
		ON CONFLICT (email) DO UPDATE SET
			external_id = EXCLUDED.external_id,
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
			custom_json_1 = EXCLUDED.custom_json_1,
			custom_json_2 = EXCLUDED.custom_json_2,
			custom_json_3 = EXCLUDED.custom_json_3,
			custom_json_4 = EXCLUDED.custom_json_4,
			custom_json_5 = EXCLUDED.custom_json_5,
			updated_at = EXCLUDED.updated_at
	`

	// Convert domain nullable types to SQL nullable types
	var firstNameSQL, lastNameSQL, phoneSQL, addressLine1SQL, addressLine2SQL sql.NullString
	var countrySQL, postcodeSQL, stateSQL, jobTitleSQL sql.NullString
	var customString1SQL, customString2SQL, customString3SQL, customString4SQL, customString5SQL sql.NullString
	var lifetimeValueSQL, ordersCountSQL sql.NullFloat64
	var customNumber1SQL, customNumber2SQL, customNumber3SQL, customNumber4SQL, customNumber5SQL sql.NullFloat64
	var lastOrderAtSQL, customDatetime1SQL, customDatetime2SQL, customDatetime3SQL, customDatetime4SQL, customDatetime5SQL sql.NullTime
	var customJSON1SQL, customJSON2SQL, customJSON3SQL, customJSON4SQL, customJSON5SQL []byte

	// String fields
	if !contact.FirstName.IsNull {
		firstNameSQL = sql.NullString{String: contact.FirstName.String, Valid: true}
	}
	if !contact.LastName.IsNull {
		lastNameSQL = sql.NullString{String: contact.LastName.String, Valid: true}
	}
	if !contact.Phone.IsNull {
		phoneSQL = sql.NullString{String: contact.Phone.String, Valid: true}
	}
	if !contact.AddressLine1.IsNull {
		addressLine1SQL = sql.NullString{String: contact.AddressLine1.String, Valid: true}
	}
	if !contact.AddressLine2.IsNull {
		addressLine2SQL = sql.NullString{String: contact.AddressLine2.String, Valid: true}
	}
	if !contact.Country.IsNull {
		countrySQL = sql.NullString{String: contact.Country.String, Valid: true}
	}
	if !contact.Postcode.IsNull {
		postcodeSQL = sql.NullString{String: contact.Postcode.String, Valid: true}
	}
	if !contact.State.IsNull {
		stateSQL = sql.NullString{String: contact.State.String, Valid: true}
	}
	if !contact.JobTitle.IsNull {
		jobTitleSQL = sql.NullString{String: contact.JobTitle.String, Valid: true}
	}

	// Custom string fields
	if !contact.CustomString1.IsNull {
		customString1SQL = sql.NullString{String: contact.CustomString1.String, Valid: true}
	}
	if !contact.CustomString2.IsNull {
		customString2SQL = sql.NullString{String: contact.CustomString2.String, Valid: true}
	}
	if !contact.CustomString3.IsNull {
		customString3SQL = sql.NullString{String: contact.CustomString3.String, Valid: true}
	}
	if !contact.CustomString4.IsNull {
		customString4SQL = sql.NullString{String: contact.CustomString4.String, Valid: true}
	}
	if !contact.CustomString5.IsNull {
		customString5SQL = sql.NullString{String: contact.CustomString5.String, Valid: true}
	}

	// Number fields
	if !contact.LifetimeValue.IsNull {
		lifetimeValueSQL = sql.NullFloat64{Float64: contact.LifetimeValue.Float64, Valid: true}
	}
	if !contact.OrdersCount.IsNull {
		ordersCountSQL = sql.NullFloat64{Float64: contact.OrdersCount.Float64, Valid: true}
	}

	// Custom number fields
	if !contact.CustomNumber1.IsNull {
		customNumber1SQL = sql.NullFloat64{Float64: contact.CustomNumber1.Float64, Valid: true}
	}
	if !contact.CustomNumber2.IsNull {
		customNumber2SQL = sql.NullFloat64{Float64: contact.CustomNumber2.Float64, Valid: true}
	}
	if !contact.CustomNumber3.IsNull {
		customNumber3SQL = sql.NullFloat64{Float64: contact.CustomNumber3.Float64, Valid: true}
	}
	if !contact.CustomNumber4.IsNull {
		customNumber4SQL = sql.NullFloat64{Float64: contact.CustomNumber4.Float64, Valid: true}
	}
	if !contact.CustomNumber5.IsNull {
		customNumber5SQL = sql.NullFloat64{Float64: contact.CustomNumber5.Float64, Valid: true}
	}

	// Datetime fields
	if !contact.LastOrderAt.IsNull {
		lastOrderAtSQL = sql.NullTime{Time: contact.LastOrderAt.Time, Valid: true}
	}

	// Custom datetime fields
	if !contact.CustomDatetime1.IsNull {
		customDatetime1SQL = sql.NullTime{Time: contact.CustomDatetime1.Time, Valid: true}
	}
	if !contact.CustomDatetime2.IsNull {
		customDatetime2SQL = sql.NullTime{Time: contact.CustomDatetime2.Time, Valid: true}
	}
	if !contact.CustomDatetime3.IsNull {
		customDatetime3SQL = sql.NullTime{Time: contact.CustomDatetime3.Time, Valid: true}
	}
	if !contact.CustomDatetime4.IsNull {
		customDatetime4SQL = sql.NullTime{Time: contact.CustomDatetime4.Time, Valid: true}
	}
	if !contact.CustomDatetime5.IsNull {
		customDatetime5SQL = sql.NullTime{Time: contact.CustomDatetime5.Time, Valid: true}
	}

	// Custom JSON fields
	if contact.CustomJSON1.Valid {
		customJSON1SQL, err = json.Marshal(contact.CustomJSON1.Data)
		if err != nil {
			return false, fmt.Errorf("failed to marshal CustomJSON1: %w", err)
		}
	}
	if contact.CustomJSON2.Valid {
		customJSON2SQL, err = json.Marshal(contact.CustomJSON2.Data)
		if err != nil {
			return false, fmt.Errorf("failed to marshal CustomJSON2: %w", err)
		}
	}
	if contact.CustomJSON3.Valid {
		customJSON3SQL, err = json.Marshal(contact.CustomJSON3.Data)
		if err != nil {
			return false, fmt.Errorf("failed to marshal CustomJSON3: %w", err)
		}
	}
	if contact.CustomJSON4.Valid {
		customJSON4SQL, err = json.Marshal(contact.CustomJSON4.Data)
		if err != nil {
			return false, fmt.Errorf("failed to marshal CustomJSON4: %w", err)
		}
	}
	if contact.CustomJSON5.Valid {
		customJSON5SQL, err = json.Marshal(contact.CustomJSON5.Data)
		if err != nil {
			return false, fmt.Errorf("failed to marshal CustomJSON5: %w", err)
		}
	}

	// Convert domain nullable types to SQL nullable types
	var externalIDSQL, timezoneSQL sql.NullString
	if !contact.ExternalID.IsNull {
		externalIDSQL = sql.NullString{String: contact.ExternalID.String, Valid: true}
	}
	if !contact.Timezone.IsNull {
		timezoneSQL = sql.NullString{String: contact.Timezone.String, Valid: true}
	}

	_, err = r.db.ExecContext(ctx, query,
		contact.Email,
		externalIDSQL,
		timezoneSQL,
		contact.Language,
		firstNameSQL,
		lastNameSQL,
		phoneSQL,
		addressLine1SQL,
		addressLine2SQL,
		countrySQL,
		postcodeSQL,
		stateSQL,
		jobTitleSQL,
		lifetimeValueSQL,
		ordersCountSQL,
		lastOrderAtSQL,
		customString1SQL,
		customString2SQL,
		customString3SQL,
		customString4SQL,
		customString5SQL,
		customNumber1SQL,
		customNumber2SQL,
		customNumber3SQL,
		customNumber4SQL,
		customNumber5SQL,
		customDatetime1SQL,
		customDatetime2SQL,
		customDatetime3SQL,
		customDatetime4SQL,
		customDatetime5SQL,
		customJSON1SQL,
		customJSON2SQL,
		customJSON3SQL,
		customJSON4SQL,
		customJSON5SQL,
		contact.CreatedAt,
		contact.UpdatedAt,
	)

	if err != nil {
		return false, fmt.Errorf("failed to upsert contact: %w", err)
	}

	return isNew, nil
}
