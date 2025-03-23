package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestGetContactByEmail(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			email, "ext123", "Europe/Paris",
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
	assert.Equal(t, email, contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByEmail(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
}

func TestGetContactByExternalID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	externalID := "ext123"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			"test@example.com", externalID, "Europe/Paris",
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

	_, err := repo.GetContactByExternalID(context.Background(), externalID)
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs("nonexistent-ext-id").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByExternalID(context.Background(), "nonexistent-ext-id")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)

	// Test: get contact by external ID successful case
	t.Run("successful case", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT email, external_id, timezone, first_name, last_name, phone, address_line_1, address_line_2, country, postcode, state, job_title, lifetime_value, orders_count, last_order_at, custom_string_1, custom_string_2, custom_string_3, custom_string_4, custom_string_5, custom_number_1, custom_number_2, custom_number_3, custom_number_4, custom_number_5, custom_datetime_1, custom_datetime_2, custom_datetime_3, custom_datetime_4, custom_datetime_5, created_at, updated_at FROM contacts WHERE external_id = $1`)).
			WithArgs("e-123").
			WillReturnRows(
				sqlmock.NewRows([]string{"email", "external_id", "timezone", "first_name", "last_name", "phone", "address_line_1", "address_line_2", "country", "postcode", "state", "job_title", "lifetime_value", "orders_count", "last_order_at", "custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5", "created_at", "updated_at"}).
					AddRow("test@example.com", "e-123", "Europe/Paris", "John", "Doe", "", "", "", "", "", "", "", 0, 0, time.Time{}, "", "", "", "", "", 0, 0, 0, 0, 0, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Now(), time.Now()),
			)

		_, err := repo.GetContactByExternalID(context.Background(), "e-123")
		assert.NoError(t, err)
	})
}

func TestGetContacts(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Test case 1: Multiple contacts found
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			"test1@example.com", "ext123", "Europe/Paris",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			now, now,
		).
		AddRow(
			"test2@example.com", "ext456", "America/New_York",
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
		"email", "external_id", "timezone",
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
	email := "test@example.com"

	// Test case 1: Contact successfully deleted
	mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteContact(context.Background(), email)
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteContact(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
}

func TestBatchImportContacts(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create some test contacts
	contact1 := &domain.Contact{
		Email:      "contact1@example.com",
		ExternalID: "ext1",
		Timezone:   "Europe/Paris",
		FirstName:  domain.NullableString{String: "John", IsNull: false},
		LastName:   domain.NullableString{String: "Doe", IsNull: false},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	contact2 := &domain.Contact{
		Email:      "contact2@example.com",
		ExternalID: "ext2",
		Timezone:   "America/New_York",
		FirstName:  domain.NullableString{String: "Jane", IsNull: false},
		LastName:   domain.NullableString{String: "Smith", IsNull: false},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	testContacts := []*domain.Contact{contact1, contact2}

	// Test case 1: Successful batch import
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`)

	for range testContacts {
		mock.ExpectExec(`INSERT INTO contacts`).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err := repo.BatchImportContacts(context.Background(), testContacts)
	require.NoError(t, err)

	// Test case 2: Transaction begin error
	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	err = repo.BatchImportContacts(context.Background(), testContacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")

	// Test case 3: Statement preparation error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`).WillReturnError(errors.New("prepare error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), testContacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")

	// Test case 4: Execution error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`)
	mock.ExpectExec(`INSERT INTO contacts`).WillReturnError(errors.New("execution error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), []*domain.Contact{contact1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute statement")

	// Test case 5: Commit error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`)
	mock.ExpectExec(`INSERT INTO contacts`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err = repo.BatchImportContacts(context.Background(), []*domain.Contact{contact1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

func TestUpsertContact(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"

	testContact := &domain.Contact{
		Email:      email,
		ExternalID: "ext123",
		Timezone:   "Europe/Paris",
		FirstName:  domain.NullableString{String: "John", IsNull: false},
		LastName:   domain.NullableString{String: "Doe", IsNull: false},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Test case 1: Insert new contact
	// First, check if contact exists
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	// Then, insert the contact
	mock.ExpectExec(`INSERT INTO contacts`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err := repo.UpsertContact(context.Background(), testContact)
	require.NoError(t, err)
	assert.True(t, created)

	// Test case 2: Update existing contact
	// First, check if contact exists
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"created_at", "updated_at",
	}).
		AddRow(
			email, "old-ext-id", "Europe/Paris",
			"Old", "Name", "", "", "",
			"", "", "", "",
			0, 0, time.Time{},
			"", "", "", "", "",
			0, 0, 0, 0, 0,
			time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{},
			now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	// Then, update the contact
	mock.ExpectExec(`INSERT INTO contacts`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err = repo.UpsertContact(context.Background(), testContact)
	require.NoError(t, err)
	assert.False(t, created)

	// Test case 3: Error checking if contact exists
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(errors.New("check error"))

	created, err = repo.UpsertContact(context.Background(), testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if contact exists")

	// Test case 4: Error upserting contact
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WillReturnError(errors.New("upsert error"))

	created, err = repo.UpsertContact(context.Background(), testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert contact")
}
