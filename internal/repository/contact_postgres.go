package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type contactRepository struct {
	workspaceRepo domain.WorkspaceRepository
}

// NewContactRepository creates a new PostgreSQL contact repository
func NewContactRepository(workspaceRepo domain.WorkspaceRepository) domain.ContactRepository {
	return &contactRepository{
		workspaceRepo: workspaceRepo,
	}
}

func (r *contactRepository) GetContactByEmail(ctx context.Context, workspaceID, email string) (*domain.Contact, error) {
	db, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `SELECT c.* FROM contacts c WHERE c.email = $1`

	row := db.QueryRowContext(ctx, query, email)

	contact, err := domain.ScanContact(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (r *contactRepository) GetContactByExternalID(ctx context.Context, externalID, workspaceID string) (*domain.Contact, error) {
	db, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `SELECT c.* FROM contacts c WHERE c.external_id = $1`
	row := db.QueryRowContext(ctx, query, externalID)

	contact, err := domain.ScanContact(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (r *contactRepository) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	db, err := r.workspaceRepo.GetConnection(ctx, req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Build the query
	query := strings.Builder{}
	query.WriteString("SELECT c.* FROM contacts c ")

	// Add filters
	var filters []string
	var args []interface{}
	argCount := 1

	if req.Email != "" {
		filters = append(filters, fmt.Sprintf("c.email ILIKE $%d", argCount))
		args = append(args, "%"+req.Email+"%")
		argCount++
	}

	if req.ExternalID != "" {
		filters = append(filters, fmt.Sprintf("c.external_id ILIKE $%d", argCount))
		args = append(args, "%"+req.ExternalID+"%")
		argCount++
	}

	if req.FirstName != "" {
		filters = append(filters, fmt.Sprintf("c.first_name ILIKE $%d", argCount))
		args = append(args, "%"+req.FirstName+"%")
		argCount++
	}

	if req.LastName != "" {
		filters = append(filters, fmt.Sprintf("c.last_name ILIKE $%d", argCount))
		args = append(args, "%"+req.LastName+"%")
		argCount++
	}

	if req.Phone != "" {
		filters = append(filters, fmt.Sprintf("c.phone ILIKE $%d", argCount))
		args = append(args, "%"+req.Phone+"%")
		argCount++
	}

	if req.Country != "" {
		filters = append(filters, fmt.Sprintf("c.country ILIKE $%d", argCount))
		args = append(args, "%"+req.Country+"%")
		argCount++
	}

	if req.Language != "" {
		filters = append(filters, fmt.Sprintf("c.language ILIKE $%d", argCount))
		args = append(args, "%"+req.Language+"%")
		argCount++
	}

	if req.Cursor != "" {
		// Parse cursor as timestamp
		cursorTime, err := time.Parse(time.RFC3339, req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}
		filters = append(filters, fmt.Sprintf("c.created_at < $%d", argCount))
		args = append(args, cursorTime)
		argCount++
	}

	if len(filters) > 0 {
		query.WriteString("WHERE " + strings.Join(filters, " AND "))
	}

	// Add order by and limit
	query.WriteString(" ORDER BY c.created_at DESC")
	query.WriteString(fmt.Sprintf(" LIMIT $%d", argCount))
	args = append(args, req.Limit+1) // Get one extra to determine if there are more results

	// Execute query
	rows, err := db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Process results
	var contacts []*domain.Contact
	var nextCursor string

	for rows.Next() {
		contact, err := domain.ScanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	// Handle pagination
	if len(contacts) > req.Limit {
		// Remove the extra contact we fetched
		lastContact := contacts[req.Limit-1]
		contacts = contacts[:req.Limit]
		nextCursor = lastContact.CreatedAt.Format(time.RFC3339)
	}

	// If WithContactLists is true, fetch contact lists in a separate query
	if req.WithContactLists && len(contacts) > 0 {
		// Build list of contact emails
		emails := make([]string, len(contacts))
		for i, contact := range contacts {
			emails[i] = contact.Email
		}

		// Build the IN clause with placeholders
		placeholders := make([]string, len(emails))
		for i := range emails {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		// Query for contact lists
		listQuery := fmt.Sprintf(`
			SELECT email, list_id, status, created_at, updated_at
			FROM contact_lists
			WHERE email IN (%s)
		`, strings.Join(placeholders, ","))

		// Convert emails to interface slice for query args
		emailArgs := make([]interface{}, len(emails))
		for i, email := range emails {
			emailArgs[i] = email
		}

		listRows, err := db.QueryContext(ctx, listQuery, emailArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to query contact lists: %w", err)
		}
		defer listRows.Close()

		// Create a map of contacts by email for quick lookup
		contactMap := make(map[string]*domain.Contact)
		for _, contact := range contacts {
			contact.ContactLists = []*domain.ContactList{}
			contactMap[contact.Email] = contact
		}

		// Process contact list results
		for listRows.Next() {
			var email string
			var list domain.ContactList
			err := listRows.Scan(&email, &list.ListID, &list.Status, &list.CreatedAt, &list.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan contact list: %w", err)
			}

			if contact, ok := contactMap[email]; ok {
				contact.ContactLists = append(contact.ContactLists, &list)
			}
		}

		if err = listRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over contact list rows: %w", err)
		}
	}

	return &domain.GetContactsResponse{
		Contacts:   contacts,
		NextCursor: nextCursor,
	}, nil
}

func (r *contactRepository) DeleteContact(ctx context.Context, email string, workspaceID string) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	query := `DELETE FROM contacts WHERE email = $1`

	result, err := workspaceDB.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("contact not found")
	}

	return nil
}

func (r *contactRepository) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) (isNew bool, err error) {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Start a transaction
	tx, err := workspaceDB.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if there's a panic or error

	// Check if contact exists with FOR UPDATE lock
	var existingContact *domain.Contact
	query := `SELECT c.* FROM contacts c WHERE c.email = $1 FOR UPDATE`
	row := tx.QueryRowContext(ctx, query, contact.Email)
	existingContact, err = domain.ScanContact(row)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("failed to check existing contact: %w", err)
		}

		// Contact doesn't exist, do an INSERT
		insertQuery := `
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
				$33, $34, $35, $36, $37, COALESCE($38, NOW()), NOW()
			)
		`

		// Convert domain nullable types to SQL nullable types
		var firstNameSQL, lastNameSQL, phoneSQL, addressLine1SQL, addressLine2SQL sql.NullString
		var countrySQL, postcodeSQL, stateSQL, jobTitleSQL sql.NullString
		var lifetimeValueSQL, ordersCountSQL sql.NullFloat64
		var lastOrderAtSQL sql.NullTime
		var customString1SQL, customString2SQL, customString3SQL, customString4SQL, customString5SQL sql.NullString
		var customNumber1SQL, customNumber2SQL, customNumber3SQL, customNumber4SQL, customNumber5SQL sql.NullFloat64
		var customDatetime1SQL, customDatetime2SQL, customDatetime3SQL, customDatetime4SQL, customDatetime5SQL sql.NullTime
		var customJSON1SQL, customJSON2SQL, customJSON3SQL, customJSON4SQL, customJSON5SQL sql.NullString

		// String fields
		if contact.FirstName != nil {
			if !contact.FirstName.IsNull {
				firstNameSQL = sql.NullString{String: contact.FirstName.String, Valid: true}
			} else {
				firstNameSQL = sql.NullString{Valid: false}
			}
		}
		if contact.LastName != nil {
			if !contact.LastName.IsNull {
				lastNameSQL = sql.NullString{String: contact.LastName.String, Valid: true}
			} else {
				lastNameSQL = sql.NullString{Valid: false}
			}
		}
		if contact.Phone != nil {
			if !contact.Phone.IsNull {
				phoneSQL = sql.NullString{String: contact.Phone.String, Valid: true}
			} else {
				phoneSQL = sql.NullString{Valid: false}
			}
		}
		if contact.AddressLine1 != nil {
			if !contact.AddressLine1.IsNull {
				addressLine1SQL = sql.NullString{String: contact.AddressLine1.String, Valid: true}
			} else {
				addressLine1SQL = sql.NullString{Valid: false}
			}
		}
		if contact.AddressLine2 != nil {
			if !contact.AddressLine2.IsNull {
				addressLine2SQL = sql.NullString{String: contact.AddressLine2.String, Valid: true}
			} else {
				addressLine2SQL = sql.NullString{Valid: false}
			}
		}
		if contact.Country != nil {
			if !contact.Country.IsNull {
				countrySQL = sql.NullString{String: contact.Country.String, Valid: true}
			} else {
				countrySQL = sql.NullString{Valid: false}
			}
		}
		if contact.Postcode != nil {
			if !contact.Postcode.IsNull {
				postcodeSQL = sql.NullString{String: contact.Postcode.String, Valid: true}
			} else {
				postcodeSQL = sql.NullString{Valid: false}
			}
		}
		if contact.State != nil {
			if !contact.State.IsNull {
				stateSQL = sql.NullString{String: contact.State.String, Valid: true}
			} else {
				stateSQL = sql.NullString{Valid: false}
			}
		}
		if contact.JobTitle != nil {
			if !contact.JobTitle.IsNull {
				jobTitleSQL = sql.NullString{String: contact.JobTitle.String, Valid: true}
			} else {
				jobTitleSQL = sql.NullString{Valid: false}
			}
		}

		// Custom string fields
		if contact.CustomString1 != nil {
			if !contact.CustomString1.IsNull {
				customString1SQL = sql.NullString{String: contact.CustomString1.String, Valid: true}
			} else {
				customString1SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomString2 != nil {
			if !contact.CustomString2.IsNull {
				customString2SQL = sql.NullString{String: contact.CustomString2.String, Valid: true}
			} else {
				customString2SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomString3 != nil {
			if !contact.CustomString3.IsNull {
				customString3SQL = sql.NullString{String: contact.CustomString3.String, Valid: true}
			} else {
				customString3SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomString4 != nil {
			if !contact.CustomString4.IsNull {
				customString4SQL = sql.NullString{String: contact.CustomString4.String, Valid: true}
			} else {
				customString4SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomString5 != nil {
			if !contact.CustomString5.IsNull {
				customString5SQL = sql.NullString{String: contact.CustomString5.String, Valid: true}
			} else {
				customString5SQL = sql.NullString{Valid: false}
			}
		}

		// Number fields
		if contact.LifetimeValue != nil {
			if !contact.LifetimeValue.IsNull {
				lifetimeValueSQL = sql.NullFloat64{Float64: contact.LifetimeValue.Float64, Valid: true}
			} else {
				lifetimeValueSQL = sql.NullFloat64{Valid: false}
			}
		}
		if contact.OrdersCount != nil {
			if !contact.OrdersCount.IsNull {
				ordersCountSQL = sql.NullFloat64{Float64: contact.OrdersCount.Float64, Valid: true}
			} else {
				ordersCountSQL = sql.NullFloat64{Valid: false}
			}
		}

		// Custom number fields
		if contact.CustomNumber1 != nil {
			if !contact.CustomNumber1.IsNull {
				customNumber1SQL = sql.NullFloat64{Float64: contact.CustomNumber1.Float64, Valid: true}
			} else {
				customNumber1SQL = sql.NullFloat64{Valid: false}
			}
		}
		if contact.CustomNumber2 != nil {
			if !contact.CustomNumber2.IsNull {
				customNumber2SQL = sql.NullFloat64{Float64: contact.CustomNumber2.Float64, Valid: true}
			} else {
				customNumber2SQL = sql.NullFloat64{Valid: false}
			}
		}
		if contact.CustomNumber3 != nil {
			if !contact.CustomNumber3.IsNull {
				customNumber3SQL = sql.NullFloat64{Float64: contact.CustomNumber3.Float64, Valid: true}
			} else {
				customNumber3SQL = sql.NullFloat64{Valid: false}
			}
		}
		if contact.CustomNumber4 != nil {
			if !contact.CustomNumber4.IsNull {
				customNumber4SQL = sql.NullFloat64{Float64: contact.CustomNumber4.Float64, Valid: true}
			} else {
				customNumber4SQL = sql.NullFloat64{Valid: false}
			}
		}
		if contact.CustomNumber5 != nil {
			if !contact.CustomNumber5.IsNull {
				customNumber5SQL = sql.NullFloat64{Float64: contact.CustomNumber5.Float64, Valid: true}
			} else {
				customNumber5SQL = sql.NullFloat64{Valid: false}
			}
		}

		// Datetime fields
		if contact.LastOrderAt != nil {
			if !contact.LastOrderAt.IsNull {
				lastOrderAtSQL = sql.NullTime{Time: contact.LastOrderAt.Time, Valid: true}
			} else {
				lastOrderAtSQL = sql.NullTime{Valid: false}
			}
		}

		// Custom datetime fields
		if contact.CustomDatetime1 != nil {
			if !contact.CustomDatetime1.IsNull {
				customDatetime1SQL = sql.NullTime{Time: contact.CustomDatetime1.Time, Valid: true}
			} else {
				customDatetime1SQL = sql.NullTime{Valid: false}
			}
		}
		if contact.CustomDatetime2 != nil {
			if !contact.CustomDatetime2.IsNull {
				customDatetime2SQL = sql.NullTime{Time: contact.CustomDatetime2.Time, Valid: true}
			} else {
				customDatetime2SQL = sql.NullTime{Valid: false}
			}
		}
		if contact.CustomDatetime3 != nil {
			if !contact.CustomDatetime3.IsNull {
				customDatetime3SQL = sql.NullTime{Time: contact.CustomDatetime3.Time, Valid: true}
			} else {
				customDatetime3SQL = sql.NullTime{Valid: false}
			}
		}
		if contact.CustomDatetime4 != nil {
			if !contact.CustomDatetime4.IsNull {
				customDatetime4SQL = sql.NullTime{Time: contact.CustomDatetime4.Time, Valid: true}
			} else {
				customDatetime4SQL = sql.NullTime{Valid: false}
			}
		}
		if contact.CustomDatetime5 != nil {
			if !contact.CustomDatetime5.IsNull {
				customDatetime5SQL = sql.NullTime{Time: contact.CustomDatetime5.Time, Valid: true}
			} else {
				customDatetime5SQL = sql.NullTime{Valid: false}
			}
		}

		// Custom JSON fields
		if contact.CustomJSON1 != nil {
			if !contact.CustomJSON1.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON1.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_1: %w", err)
				}
				customJSON1SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			} else {
				customJSON1SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomJSON2 != nil {
			if !contact.CustomJSON2.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON2.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_2: %w", err)
				}
				customJSON2SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			} else {
				customJSON2SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomJSON3 != nil {
			if !contact.CustomJSON3.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON3.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_3: %w", err)
				}
				customJSON3SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			} else {
				customJSON3SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomJSON4 != nil {
			if !contact.CustomJSON4.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON4.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_4: %w", err)
				}
				customJSON4SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			} else {
				customJSON4SQL = sql.NullString{Valid: false}
			}
		}
		if contact.CustomJSON5 != nil {
			if !contact.CustomJSON5.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON5.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_5: %w", err)
				}
				customJSON5SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			} else {
				customJSON5SQL = sql.NullString{Valid: false}
			}
		}

		// Execute the insert query within the transaction
		_, err = tx.ExecContext(ctx, insertQuery,
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
		)
		if err != nil {
			return false, fmt.Errorf("failed to insert contact: %w", err)
		}

		// For new contacts, return isNew = true
		isNew = true
	} else {
		// Contact exists, merge with the existing contact
		existingContact.Merge(contact)

		// Build a complete update query with all fields
		updateQuery := `
			UPDATE contacts SET
				external_id = $1,
				timezone = $2,
				language = $3,
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
				custom_json_1 = $31,
				custom_json_2 = $32,
				custom_json_3 = $33,
				custom_json_4 = $34,
				custom_json_5 = $35,
				updated_at = NOW()
			WHERE email = $36
		`

		// Convert domain nullable types to SQL nullable types for the update
		// String fields
		var externalIDSQL sql.NullString
		var timezoneSQL sql.NullString
		var languageSQL sql.NullString
		var firstNameSQL sql.NullString
		var lastNameSQL sql.NullString
		var phoneSQL sql.NullString
		var addressLine1SQL sql.NullString
		var addressLine2SQL sql.NullString
		var countrySQL sql.NullString
		var postcodeSQL sql.NullString
		var stateSQL sql.NullString
		var jobTitleSQL sql.NullString
		var customString1SQL sql.NullString
		var customString2SQL sql.NullString
		var customString3SQL sql.NullString
		var customString4SQL sql.NullString
		var customString5SQL sql.NullString

		// Number fields
		var lifetimeValueSQL sql.NullFloat64
		var ordersCountSQL sql.NullFloat64
		var customNumber1SQL sql.NullFloat64
		var customNumber2SQL sql.NullFloat64
		var customNumber3SQL sql.NullFloat64
		var customNumber4SQL sql.NullFloat64
		var customNumber5SQL sql.NullFloat64

		// DateTime fields
		var lastOrderAtSQL sql.NullTime
		var customDatetime1SQL sql.NullTime
		var customDatetime2SQL sql.NullTime
		var customDatetime3SQL sql.NullTime
		var customDatetime4SQL sql.NullTime
		var customDatetime5SQL sql.NullTime

		// JSON fields
		var customJSON1SQL sql.NullString
		var customJSON2SQL sql.NullString
		var customJSON3SQL sql.NullString
		var customJSON4SQL sql.NullString
		var customJSON5SQL sql.NullString

		// Convert external ID, timezone, language
		if existingContact.ExternalID != nil {
			if !existingContact.ExternalID.IsNull {
				externalIDSQL = sql.NullString{String: existingContact.ExternalID.String, Valid: true}
			}
		}
		if existingContact.Timezone != nil {
			if !existingContact.Timezone.IsNull {
				timezoneSQL = sql.NullString{String: existingContact.Timezone.String, Valid: true}
			}
		}
		if existingContact.Language != nil {
			if !existingContact.Language.IsNull {
				languageSQL = sql.NullString{String: existingContact.Language.String, Valid: true}
			}
		}

		// Convert string fields
		if existingContact.FirstName != nil {
			if !existingContact.FirstName.IsNull {
				firstNameSQL = sql.NullString{String: existingContact.FirstName.String, Valid: true}
			}
		}
		if existingContact.LastName != nil {
			if !existingContact.LastName.IsNull {
				lastNameSQL = sql.NullString{String: existingContact.LastName.String, Valid: true}
			}
		}
		if existingContact.Phone != nil {
			if !existingContact.Phone.IsNull {
				phoneSQL = sql.NullString{String: existingContact.Phone.String, Valid: true}
			}
		}
		if existingContact.AddressLine1 != nil {
			if !existingContact.AddressLine1.IsNull {
				addressLine1SQL = sql.NullString{String: existingContact.AddressLine1.String, Valid: true}
			}
		}
		if existingContact.AddressLine2 != nil {
			if !existingContact.AddressLine2.IsNull {
				addressLine2SQL = sql.NullString{String: existingContact.AddressLine2.String, Valid: true}
			}
		}
		if existingContact.Country != nil {
			if !existingContact.Country.IsNull {
				countrySQL = sql.NullString{String: existingContact.Country.String, Valid: true}
			}
		}
		if existingContact.Postcode != nil {
			if !existingContact.Postcode.IsNull {
				postcodeSQL = sql.NullString{String: existingContact.Postcode.String, Valid: true}
			}
		}
		if existingContact.State != nil {
			if !existingContact.State.IsNull {
				stateSQL = sql.NullString{String: existingContact.State.String, Valid: true}
			}
		}
		if existingContact.JobTitle != nil {
			if !existingContact.JobTitle.IsNull {
				jobTitleSQL = sql.NullString{String: existingContact.JobTitle.String, Valid: true}
			}
		}

		// Convert custom string fields
		if existingContact.CustomString1 != nil {
			if !existingContact.CustomString1.IsNull {
				customString1SQL = sql.NullString{String: existingContact.CustomString1.String, Valid: true}
			}
		}
		if existingContact.CustomString2 != nil {
			if !existingContact.CustomString2.IsNull {
				customString2SQL = sql.NullString{String: existingContact.CustomString2.String, Valid: true}
			}
		}
		if existingContact.CustomString3 != nil {
			if !existingContact.CustomString3.IsNull {
				customString3SQL = sql.NullString{String: existingContact.CustomString3.String, Valid: true}
			}
		}
		if existingContact.CustomString4 != nil {
			if !existingContact.CustomString4.IsNull {
				customString4SQL = sql.NullString{String: existingContact.CustomString4.String, Valid: true}
			}
		}
		if existingContact.CustomString5 != nil {
			if !existingContact.CustomString5.IsNull {
				customString5SQL = sql.NullString{String: existingContact.CustomString5.String, Valid: true}
			}
		}

		// Convert number fields
		if existingContact.LifetimeValue != nil {
			if !existingContact.LifetimeValue.IsNull {
				lifetimeValueSQL = sql.NullFloat64{Float64: existingContact.LifetimeValue.Float64, Valid: true}
			}
		}
		if existingContact.OrdersCount != nil {
			if !existingContact.OrdersCount.IsNull {
				ordersCountSQL = sql.NullFloat64{Float64: existingContact.OrdersCount.Float64, Valid: true}
			}
		}

		// Convert custom number fields
		if existingContact.CustomNumber1 != nil {
			if !existingContact.CustomNumber1.IsNull {
				customNumber1SQL = sql.NullFloat64{Float64: existingContact.CustomNumber1.Float64, Valid: true}
			}
		}
		if existingContact.CustomNumber2 != nil {
			if !existingContact.CustomNumber2.IsNull {
				customNumber2SQL = sql.NullFloat64{Float64: existingContact.CustomNumber2.Float64, Valid: true}
			}
		}
		if existingContact.CustomNumber3 != nil {
			if !existingContact.CustomNumber3.IsNull {
				customNumber3SQL = sql.NullFloat64{Float64: existingContact.CustomNumber3.Float64, Valid: true}
			}
		}
		if existingContact.CustomNumber4 != nil {
			if !existingContact.CustomNumber4.IsNull {
				customNumber4SQL = sql.NullFloat64{Float64: existingContact.CustomNumber4.Float64, Valid: true}
			}
		}
		if existingContact.CustomNumber5 != nil {
			if !existingContact.CustomNumber5.IsNull {
				customNumber5SQL = sql.NullFloat64{Float64: existingContact.CustomNumber5.Float64, Valid: true}
			}
		}

		// Convert datetime fields
		if existingContact.LastOrderAt != nil {
			if !existingContact.LastOrderAt.IsNull {
				lastOrderAtSQL = sql.NullTime{Time: existingContact.LastOrderAt.Time, Valid: true}
			}
		}

		// Convert custom datetime fields
		if existingContact.CustomDatetime1 != nil {
			if !existingContact.CustomDatetime1.IsNull {
				customDatetime1SQL = sql.NullTime{Time: existingContact.CustomDatetime1.Time, Valid: true}
			}
		}
		if existingContact.CustomDatetime2 != nil {
			if !existingContact.CustomDatetime2.IsNull {
				customDatetime2SQL = sql.NullTime{Time: existingContact.CustomDatetime2.Time, Valid: true}
			}
		}
		if existingContact.CustomDatetime3 != nil {
			if !existingContact.CustomDatetime3.IsNull {
				customDatetime3SQL = sql.NullTime{Time: existingContact.CustomDatetime3.Time, Valid: true}
			}
		}
		if existingContact.CustomDatetime4 != nil {
			if !existingContact.CustomDatetime4.IsNull {
				customDatetime4SQL = sql.NullTime{Time: existingContact.CustomDatetime4.Time, Valid: true}
			}
		}
		if existingContact.CustomDatetime5 != nil {
			if !existingContact.CustomDatetime5.IsNull {
				customDatetime5SQL = sql.NullTime{Time: existingContact.CustomDatetime5.Time, Valid: true}
			}
		}

		// Convert JSON fields
		if existingContact.CustomJSON1 != nil {
			if !existingContact.CustomJSON1.IsNull {
				jsonBytes, err := json.Marshal(existingContact.CustomJSON1.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_1: %w", err)
				}
				customJSON1SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
		}
		if existingContact.CustomJSON2 != nil {
			if !existingContact.CustomJSON2.IsNull {
				jsonBytes, err := json.Marshal(existingContact.CustomJSON2.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_2: %w", err)
				}
				customJSON2SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
		}
		if existingContact.CustomJSON3 != nil {
			if !existingContact.CustomJSON3.IsNull {
				jsonBytes, err := json.Marshal(existingContact.CustomJSON3.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_3: %w", err)
				}
				customJSON3SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
		}
		if existingContact.CustomJSON4 != nil {
			if !existingContact.CustomJSON4.IsNull {
				jsonBytes, err := json.Marshal(existingContact.CustomJSON4.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_4: %w", err)
				}
				customJSON4SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
		}
		if existingContact.CustomJSON5 != nil {
			if !existingContact.CustomJSON5.IsNull {
				jsonBytes, err := json.Marshal(existingContact.CustomJSON5.Data)
				if err != nil {
					return false, fmt.Errorf("failed to marshal custom_json_5: %w", err)
				}
				customJSON5SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
		}

		// Execute the update query with all fields
		_, err = tx.ExecContext(ctx, updateQuery,
			externalIDSQL,
			timezoneSQL,
			languageSQL,
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
			existingContact.Email,
		)
		if err != nil {
			return false, fmt.Errorf("failed to update contact: %w", err)
		}

		// For updates, return isNew = false
		isNew = false
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return isNew, nil
}
