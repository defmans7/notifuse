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

func (r *contactRepository) GetContactByEmail(ctx context.Context, email, workspaceID string) (*domain.Contact, error) {
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
	query.WriteString("SELECT c.* ")
	if req.WithContactLists {
		query.WriteString(", cl.list_id, cl.status, cl.created_at, cl.updated_at ")
	}
	query.WriteString("FROM contacts c ")
	if req.WithContactLists {
		query.WriteString("LEFT JOIN contact_lists cl ON c.email = cl.email ")
	}

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
	contactMap := make(map[string]*domain.Contact)
	var nextCursor string

	for rows.Next() {
		contact, err := domain.ScanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}

		// If we're fetching with contact lists, we need to handle multiple rows for the same contact
		if req.WithContactLists {
			if existingContact, ok := contactMap[contact.Email]; ok {
				// If the contact already exists and has a contact list, merge it
				if len(contact.ContactLists) > 0 {
					existingContact.MergeContactLists(contact.ContactLists[0])
				}
			} else {
				// This is a new contact
				contactMap[contact.Email] = contact
				contacts = append(contacts, contact)
			}
		} else {
			contacts = append(contacts, contact)
		}
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

func (r *contactRepository) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Prepare a transaction
	tx, err := workspaceDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if there's a panic or error

	// Execute in batches
	const batchSize = 100
	for i := 0; i < len(contacts); i += batchSize {
		end := i + batchSize
		if end > len(contacts) {
			end = len(contacts)
		}

		batch := contacts[i:end]
		for _, contact := range batch {
			// Build dynamic update query for this contact
			updateClause := buildUpdateQuery(contact)

			// Construct full query with dynamic update clause
			query := fmt.Sprintf(`
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
					$33, $34, $35, $36, $37, $38
				)
				ON CONFLICT (email) DO UPDATE SET %s
			`, updateClause)

			// Convert domain nullable types to SQL nullable types
			var firstNameSQL, lastNameSQL, phoneSQL, addressLine1SQL, addressLine2SQL sql.NullString
			var countrySQL, postcodeSQL, stateSQL, jobTitleSQL sql.NullString
			var customString1SQL, customString2SQL, customString3SQL, customString4SQL, customString5SQL sql.NullString
			var lifetimeValueSQL, ordersCountSQL sql.NullFloat64
			var customNumber1SQL, customNumber2SQL, customNumber3SQL, customNumber4SQL, customNumber5SQL sql.NullFloat64
			var lastOrderAtSQL, customDatetime1SQL, customDatetime2SQL, customDatetime3SQL, customDatetime4SQL, customDatetime5SQL sql.NullTime
			var customJSON1SQL, customJSON2SQL, customJSON3SQL, customJSON4SQL, customJSON5SQL sql.NullString

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
			if !contact.CustomJSON1.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON1.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON1: %w", err)
				}
				customJSON1SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
			if !contact.CustomJSON2.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON2.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON2: %w", err)
				}
				customJSON2SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
			if !contact.CustomJSON3.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON3.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON3: %w", err)
				}
				customJSON3SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
			if !contact.CustomJSON4.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON4.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON4: %w", err)
				}
				customJSON4SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}
			if !contact.CustomJSON5.IsNull {
				jsonBytes, err := json.Marshal(contact.CustomJSON5.Data)
				if err != nil {
					return fmt.Errorf("failed to marshal CustomJSON5: %w", err)
				}
				customJSON5SQL = sql.NullString{String: string(jsonBytes), Valid: true}
			}

			_, err = tx.ExecContext(ctx, query,
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

// buildUpdateQuery builds the UPDATE part of the upsert query based on which fields are present in the contact
func buildUpdateQuery(contact *domain.Contact) string {
	var updates []string

	// Helper function to add a field to updates if it's not nil
	addField := func(fieldName string, field interface{}) {
		if field != nil {
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", fieldName, fieldName))
		}
	}

	// Add fields that are present in the contact
	addField("external_id", contact.ExternalID)
	addField("timezone", contact.Timezone)
	addField("language", contact.Language)
	addField("first_name", contact.FirstName)
	addField("last_name", contact.LastName)
	addField("phone", contact.Phone)
	addField("address_line_1", contact.AddressLine1)
	addField("address_line_2", contact.AddressLine2)
	addField("country", contact.Country)
	addField("postcode", contact.Postcode)
	addField("state", contact.State)
	addField("job_title", contact.JobTitle)
	addField("lifetime_value", contact.LifetimeValue)
	addField("orders_count", contact.OrdersCount)
	addField("last_order_at", contact.LastOrderAt)
	addField("custom_string_1", contact.CustomString1)
	addField("custom_string_2", contact.CustomString2)
	addField("custom_string_3", contact.CustomString3)
	addField("custom_string_4", contact.CustomString4)
	addField("custom_string_5", contact.CustomString5)
	addField("custom_number_1", contact.CustomNumber1)
	addField("custom_number_2", contact.CustomNumber2)
	addField("custom_number_3", contact.CustomNumber3)
	addField("custom_number_4", contact.CustomNumber4)
	addField("custom_number_5", contact.CustomNumber5)
	addField("custom_datetime_1", contact.CustomDatetime1)
	addField("custom_datetime_2", contact.CustomDatetime2)
	addField("custom_datetime_3", contact.CustomDatetime3)
	addField("custom_datetime_4", contact.CustomDatetime4)
	addField("custom_datetime_5", contact.CustomDatetime5)
	addField("custom_json_1", contact.CustomJSON1)
	addField("custom_json_2", contact.CustomJSON2)
	addField("custom_json_3", contact.CustomJSON3)
	addField("custom_json_4", contact.CustomJSON4)
	addField("custom_json_5", contact.CustomJSON5)

	// Always update the updated_at timestamp
	updates = append(updates, "updated_at = EXCLUDED.updated_at")

	return strings.Join(updates, ", ")
}

func (r *contactRepository) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) error {
	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// Check if contact exists
	existsQuery := `SELECT EXISTS(SELECT 1 FROM contacts WHERE email = $1)`
	var exists bool
	err = workspaceDB.QueryRowContext(ctx, existsQuery, contact.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if contact exists: %w", err)
	}

	// Build the base query with dynamic update part
	query := fmt.Sprintf(`
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
			$33, $34, $35, $36, $37, $38
		)
		ON CONFLICT (email) DO UPDATE SET %s
	`, buildUpdateQuery(contact))

	// Convert domain nullable types to SQL nullable types
	var firstNameSQL, lastNameSQL, phoneSQL, addressLine1SQL, addressLine2SQL sql.NullString
	var countrySQL, postcodeSQL, stateSQL, jobTitleSQL sql.NullString
	var customString1SQL, customString2SQL, customString3SQL, customString4SQL, customString5SQL sql.NullString
	var lifetimeValueSQL, ordersCountSQL sql.NullFloat64
	var customNumber1SQL, customNumber2SQL, customNumber3SQL, customNumber4SQL, customNumber5SQL sql.NullFloat64
	var lastOrderAtSQL, customDatetime1SQL, customDatetime2SQL, customDatetime3SQL, customDatetime4SQL, customDatetime5SQL sql.NullTime
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
				return fmt.Errorf("failed to marshal CustomJSON1: %w", err)
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
				return fmt.Errorf("failed to marshal CustomJSON2: %w", err)
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
				return fmt.Errorf("failed to marshal CustomJSON3: %w", err)
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
				return fmt.Errorf("failed to marshal CustomJSON4: %w", err)
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
				return fmt.Errorf("failed to marshal CustomJSON5: %w", err)
			}
			customJSON5SQL = sql.NullString{String: string(jsonBytes), Valid: true}
		} else {
			customJSON5SQL = sql.NullString{Valid: false}
		}
	}

	// Convert domain nullable types to SQL nullable types
	var externalIDSQL, timezoneSQL sql.NullString
	if contact.ExternalID != nil {
		if !contact.ExternalID.IsNull {
			externalIDSQL = sql.NullString{String: contact.ExternalID.String, Valid: true}
		} else {
			externalIDSQL = sql.NullString{Valid: false}
		}
	}
	if contact.Timezone != nil {
		if !contact.Timezone.IsNull {
			timezoneSQL = sql.NullString{String: contact.Timezone.String, Valid: true}
		} else {
			timezoneSQL = sql.NullString{Valid: false}
		}
	}

	_, err = workspaceDB.ExecContext(ctx, query,
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
		return fmt.Errorf("failed to upsert contact: %w, customJSON1: %v, customJSON2: %v, customJSON3: %v, customJSON4: %v, customJSON5: %v", err, contact.CustomJSON1, contact.CustomJSON2, contact.CustomJSON3, contact.CustomJSON4, contact.CustomJSON5)
	}

	return nil
}
