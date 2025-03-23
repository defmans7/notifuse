package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestGetContactByUUID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactUUID := uuid.New().String()

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"uuid", "external_id", "email", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			contactUUID, "ext123", "test@example.com", "Europe/Paris",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
		WithArgs(contactUUID).
		WillReturnRows(rows)

	contact, err := repo.GetContactByUUID(context.Background(), contactUUID)
	require.NoError(t, err)
	assert.Equal(t, contactUUID, contact.UUID)
	assert.Equal(t, "test@example.com", contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
		WithArgs("non-existent-uuid").
		WillReturnError(sql.ErrNoRows)

	contact, err = repo.GetContactByUUID(context.Background(), "non-existent-uuid")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
	assert.Nil(t, contact)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
		WithArgs("error-uuid").
		WillReturnError(errors.New("database error"))

	contact, err = repo.GetContactByUUID(context.Background(), "error-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get contact")
	assert.Nil(t, contact)
}

func TestGetContactByEmail(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactUUID := uuid.New().String()
	email := "test@example.com"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"uuid", "external_id", "email", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			contactUUID, "ext123", email, "Europe/Paris",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	contact, err := repo.GetContactByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.Equal(t, contactUUID, contact.UUID)
	assert.Equal(t, email, contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	contact, err = repo.GetContactByEmail(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
	assert.Nil(t, contact)
}

func TestGetContactByExternalID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactUUID := uuid.New().String()
	externalID := "ext123"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"uuid", "external_id", "email", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			contactUUID, externalID, "test@example.com", "Europe/Paris",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs(externalID).
		WillReturnRows(rows)

	contact, err := repo.GetContactByExternalID(context.Background(), externalID)
	require.NoError(t, err)
	assert.Equal(t, contactUUID, contact.UUID)
	assert.Equal(t, externalID, contact.ExternalID)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs("nonexistent-ext-id").
		WillReturnError(sql.ErrNoRows)

	contact, err = repo.GetContactByExternalID(context.Background(), "nonexistent-ext-id")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
	assert.Nil(t, contact)
}

func TestGetContacts(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Test case 1: Multiple contacts found
	rows := sqlmock.NewRows([]string{
		"uuid", "external_id", "email", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			uuid.New().String(), "ext123", "test1@example.com", "Europe/Paris",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			now, now,
		).
		AddRow(
			uuid.New().String(), "ext456", "test2@example.com", "America/New_York",
			"Jane", "Smith", "+9876543210", "456 Oak St", "Suite 7C",
			"USA", "54321", "NY", "Manager",
			200.75, 10, now,
			"Custom A", "Custom B", "Custom C", "Custom D", "Custom E",
			52.0, 53.0, 54.0, 55.0, 56.0,
			now, now, now, now, now,
			now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contacts ORDER BY created_at DESC`).
		WillReturnRows(rows)

	contacts, err := repo.GetContacts(context.Background())
	require.NoError(t, err)
	assert.Len(t, contacts, 2)
	assert.Equal(t, "test1@example.com", contacts[0].Email)
	assert.Equal(t, "test2@example.com", contacts[1].Email)

	// Test case 2: No contacts found (empty result)
	emptyRows := sqlmock.NewRows([]string{
		"uuid", "external_id", "email", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM contacts ORDER BY created_at DESC`).
		WillReturnRows(emptyRows)

	contacts, err = repo.GetContacts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, contacts)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM contacts ORDER BY created_at DESC`).
		WillReturnError(errors.New("database error"))

	contacts, err = repo.GetContacts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get contacts")
	assert.Nil(t, contacts)
}

func TestDeleteContact(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	contactUUID := uuid.New().String()

	// Test case 1: Successful deletion
	mock.ExpectExec(`DELETE FROM contacts WHERE uuid = \$1`).
		WithArgs(contactUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteContact(context.Background(), contactUUID)
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectExec(`DELETE FROM contacts WHERE uuid = \$1`).
		WithArgs("non-existent-uuid").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteContact(context.Background(), "non-existent-uuid")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)

	// Test case 3: Database error
	mock.ExpectExec(`DELETE FROM contacts WHERE uuid = \$1`).
		WithArgs("error-uuid").
		WillReturnError(errors.New("database error"))

	err = repo.DeleteContact(context.Background(), "error-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete contact")
}

func TestBatchImportContacts(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create test contacts
	contacts := []*domain.Contact{
		{
			UUID:       uuid.New().String(),
			ExternalID: "ext1",
			Email:      "contact1@example.com",
			Timezone:   "UTC",
			FirstName:  "John",
			LastName:   "Doe",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			UUID:       uuid.New().String(),
			ExternalID: "ext2",
			Email:      "contact2@example.com",
			Timezone:   "Europe/Paris",
			FirstName:  "Jane",
			LastName:   "Smith",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	// Test case 1: Successful batch import
	mock.ExpectBegin()

	// Prepare statement expectation
	mock.ExpectPrepare("INSERT INTO contacts")

	// Execute for each contact expectation
	for _, contact := range contacts {
		mock.ExpectExec("INSERT INTO contacts").
			WithArgs(
				contact.UUID, contact.ExternalID, contact.Email, contact.Timezone,
				contact.FirstName, contact.LastName, contact.Phone, contact.AddressLine1, contact.AddressLine2,
				contact.Country, contact.Postcode, contact.State, contact.JobTitle,
				contact.LifetimeValue, contact.OrdersCount, contact.LastOrderAt,
				contact.CustomString1, contact.CustomString2, contact.CustomString3, contact.CustomString4, contact.CustomString5,
				contact.CustomNumber1, contact.CustomNumber2, contact.CustomNumber3, contact.CustomNumber4, contact.CustomNumber5,
				contact.CustomDatetime1, contact.CustomDatetime2, contact.CustomDatetime3, contact.CustomDatetime4, contact.CustomDatetime5,
				contact.CreatedAt, contact.UpdatedAt,
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	err := repo.BatchImportContacts(context.Background(), contacts)
	require.NoError(t, err)

	// Test case 2: Transaction begin fails
	mock.ExpectBegin().WillReturnError(errors.New("transaction begin error"))

	err = repo.BatchImportContacts(context.Background(), contacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")

	// Test case 3: Statement preparation fails
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO contacts").WillReturnError(errors.New("prepare statement error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), contacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")

	// Test case 4: Execute fails for one of the contacts
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO contacts")
	mock.ExpectExec("INSERT INTO contacts").
		WithArgs(
			contacts[0].UUID, contacts[0].ExternalID, contacts[0].Email, contacts[0].Timezone,
			contacts[0].FirstName, contacts[0].LastName, contacts[0].Phone, contacts[0].AddressLine1, contacts[0].AddressLine2,
			contacts[0].Country, contacts[0].Postcode, contacts[0].State, contacts[0].JobTitle,
			contacts[0].LifetimeValue, contacts[0].OrdersCount, contacts[0].LastOrderAt,
			contacts[0].CustomString1, contacts[0].CustomString2, contacts[0].CustomString3, contacts[0].CustomString4, contacts[0].CustomString5,
			contacts[0].CustomNumber1, contacts[0].CustomNumber2, contacts[0].CustomNumber3, contacts[0].CustomNumber4, contacts[0].CustomNumber5,
			contacts[0].CustomDatetime1, contacts[0].CustomDatetime2, contacts[0].CustomDatetime3, contacts[0].CustomDatetime4, contacts[0].CustomDatetime5,
			contacts[0].CreatedAt, contacts[0].UpdatedAt,
		).WillReturnError(errors.New("execution error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), contacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute statement")

	// Test case 5: Commit fails
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO contacts")
	for _, contact := range contacts {
		mock.ExpectExec("INSERT INTO contacts").
			WithArgs(
				contact.UUID, contact.ExternalID, contact.Email, contact.Timezone,
				contact.FirstName, contact.LastName, contact.Phone, contact.AddressLine1, contact.AddressLine2,
				contact.Country, contact.Postcode, contact.State, contact.JobTitle,
				contact.LifetimeValue, contact.OrdersCount, contact.LastOrderAt,
				contact.CustomString1, contact.CustomString2, contact.CustomString3, contact.CustomString4, contact.CustomString5,
				contact.CustomNumber1, contact.CustomNumber2, contact.CustomNumber3, contact.CustomNumber4, contact.CustomNumber5,
				contact.CustomDatetime1, contact.CustomDatetime2, contact.CustomDatetime3, contact.CustomDatetime4, contact.CustomDatetime5,
				contact.CreatedAt, contact.UpdatedAt,
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err = repo.BatchImportContacts(context.Background(), contacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

func TestUpsertContact(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactUUID := uuid.New().String()

	t.Run("Create new contact", func(t *testing.T) {
		// First, contact doesn't exist
		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
			WithArgs(contactUUID).
			WillReturnError(sql.ErrNoRows)

		// Then the INSERT with ON CONFLICT clause
		mock.ExpectExec(`INSERT INTO contacts`).
			WithArgs(
				contactUUID, "ext123", "test@example.com", "Europe/Paris",
				"John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		contact := &domain.Contact{
			UUID:       contactUUID,
			ExternalID: "ext123",
			Email:      "test@example.com",
			Timezone:   "Europe/Paris",
			FirstName:  "John",
			LastName:   "Doe",
		}

		isNew, err := repo.UpsertContact(context.Background(), contact)
		require.NoError(t, err)
		assert.True(t, isNew, "Expected isNew to be true for a new contact")
		mock.ExpectationsWereMet()
	})

	t.Run("Update existing contact", func(t *testing.T) {
		// Reset expectations
		var err error
		db.Close()
		db, mock, err = sqlmock.New()
		require.NoError(t, err)
		cleanup = func() { db.Close() }
		repo = NewContactRepository(db)

		// Create a mock row with empty strings instead of NULLs to avoid Scan issues
		var emptyTime time.Time
		var zeroFloat float64
		var zeroInt int

		columns := []string{
			"uuid", "external_id", "email", "timezone",
			"first_name", "last_name", "phone", "address_line_1", "address_line_2",
			"country", "postcode", "state", "job_title",
			"lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
			"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
			"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"created_at", "updated_at",
		}

		mockRow := sqlmock.NewRows(columns).
			AddRow(
				contactUUID, "old-ext", "old@example.com", "UTC",
				"Old", "Name", "", "", "", // Empty strings instead of NULL
				"", "", "", "",
				zeroFloat, zeroInt, emptyTime,
				"", "", "", "", "",
				zeroFloat, zeroFloat, zeroFloat, zeroFloat, zeroFloat,
				emptyTime, emptyTime, emptyTime, emptyTime, emptyTime,
				now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
			WithArgs(contactUUID).
			WillReturnRows(mockRow)

		// Then the INSERT with ON CONFLICT clause (acts as an update)
		mock.ExpectExec(`INSERT INTO contacts`).
			WithArgs(
				contactUUID, "ext123", "test@example.com", "Europe/Paris",
				"John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		contact := &domain.Contact{
			UUID:       contactUUID,
			ExternalID: "ext123",
			Email:      "test@example.com",
			Timezone:   "Europe/Paris",
			FirstName:  "John",
			LastName:   "Doe",
		}

		isNew, err := repo.UpsertContact(context.Background(), contact)
		require.NoError(t, err)
		assert.False(t, isNew, "Expected isNew to be false for an existing contact")
		mock.ExpectationsWereMet()
	})

	t.Run("Error checking if contact exists", func(t *testing.T) {
		// Reset expectations
		var err error
		db.Close()
		db, mock, err = sqlmock.New()
		require.NoError(t, err)
		cleanup = func() { db.Close() }
		repo = NewContactRepository(db)

		// Error when checking if contact exists
		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
			WithArgs(contactUUID).
			WillReturnError(errors.New("database error"))

		contact := &domain.Contact{
			UUID:       contactUUID,
			ExternalID: "ext123",
			Email:      "test@example.com",
			Timezone:   "Europe/Paris",
		}

		_, err = repo.UpsertContact(context.Background(), contact)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if contact exists")
		mock.ExpectationsWereMet()
	})

	t.Run("Error during upsert", func(t *testing.T) {
		// Reset expectations
		var err error
		db.Close()
		db, mock, err = sqlmock.New()
		require.NoError(t, err)
		cleanup = func() { db.Close() }
		repo = NewContactRepository(db)

		// Contact doesn't exist
		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE uuid = \$1`).
			WithArgs(contactUUID).
			WillReturnError(sql.ErrNoRows)

		// Error during upsert
		mock.ExpectExec(`INSERT INTO contacts`).
			WithArgs(
				contactUUID, "ext123", "test@example.com", "Europe/Paris",
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		contact := &domain.Contact{
			UUID:       contactUUID,
			ExternalID: "ext123",
			Email:      "test@example.com",
			Timezone:   "Europe/Paris",
		}

		_, err = repo.UpsertContact(context.Background(), contact)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upsert contact")
		mock.ExpectationsWereMet()
	})
}
