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

func TestCreateContact(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	contactUUID := uuid.New().String()

	// Test case 1: Successful contact creation
	contact := &domain.Contact{
		UUID:       contactUUID,
		ExternalID: "ext123",
		Email:      "test@example.com",
		Timezone:   "Europe/Paris",
		FirstName:  "John",
		LastName:   "Doe",
	}

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			contact.UUID, contact.ExternalID, contact.Email, contact.Timezone,
			contact.FirstName, contact.LastName, contact.Phone, contact.AddressLine1, contact.AddressLine2,
			contact.Country, contact.Postcode, contact.State, contact.JobTitle,
			contact.LifetimeValue, contact.OrdersCount, contact.LastOrderAt,
			contact.CustomString1, contact.CustomString2, contact.CustomString3, contact.CustomString4, contact.CustomString5,
			contact.CustomNumber1, contact.CustomNumber2, contact.CustomNumber3, contact.CustomNumber4, contact.CustomNumber5,
			contact.CustomDatetime1, contact.CustomDatetime2, contact.CustomDatetime3, contact.CustomDatetime4, contact.CustomDatetime5,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateContact(context.Background(), contact)
	require.NoError(t, err)

	// Test case 2: Error during contact creation
	contactWithError := &domain.Contact{
		UUID:       uuid.New().String(),
		ExternalID: "ext456",
		Email:      "error@example.com",
		Timezone:   "Europe/Paris",
	}

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			contactWithError.UUID, contactWithError.ExternalID, contactWithError.Email, contactWithError.Timezone,
			contactWithError.FirstName, contactWithError.LastName, contactWithError.Phone, contactWithError.AddressLine1, contactWithError.AddressLine2,
			contactWithError.Country, contactWithError.Postcode, contactWithError.State, contactWithError.JobTitle,
			contactWithError.LifetimeValue, contactWithError.OrdersCount, contactWithError.LastOrderAt,
			contactWithError.CustomString1, contactWithError.CustomString2, contactWithError.CustomString3, contactWithError.CustomString4, contactWithError.CustomString5,
			contactWithError.CustomNumber1, contactWithError.CustomNumber2, contactWithError.CustomNumber3, contactWithError.CustomNumber4, contactWithError.CustomNumber5,
			contactWithError.CustomDatetime1, contactWithError.CustomDatetime2, contactWithError.CustomDatetime3, contactWithError.CustomDatetime4, contactWithError.CustomDatetime5,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(errors.New("database error"))

	err = repo.CreateContact(context.Background(), contactWithError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create contact")
}

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

func TestUpdateContact(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	contactUUID := uuid.New().String()

	// Test case 1: Successful update
	contact := &domain.Contact{
		UUID:       contactUUID,
		ExternalID: "ext123",
		Email:      "test@example.com",
		Timezone:   "Europe/Paris",
		FirstName:  "John",
		LastName:   "Doe",
	}

	mock.ExpectExec(`UPDATE contacts SET (.+) WHERE uuid = \$32`).
		WithArgs(
			contact.ExternalID, contact.Email, contact.Timezone,
			contact.FirstName, contact.LastName, contact.Phone, contact.AddressLine1, contact.AddressLine2,
			contact.Country, contact.Postcode, contact.State, contact.JobTitle,
			contact.LifetimeValue, contact.OrdersCount, contact.LastOrderAt,
			contact.CustomString1, contact.CustomString2, contact.CustomString3, contact.CustomString4, contact.CustomString5,
			contact.CustomNumber1, contact.CustomNumber2, contact.CustomNumber3, contact.CustomNumber4, contact.CustomNumber5,
			contact.CustomDatetime1, contact.CustomDatetime2, contact.CustomDatetime3, contact.CustomDatetime4, contact.CustomDatetime5,
			sqlmock.AnyArg(), contact.UUID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateContact(context.Background(), contact)
	require.NoError(t, err)

	// Test case 2: Contact not found
	nonExistentContact := &domain.Contact{
		UUID:       "non-existent-uuid",
		ExternalID: "ext456",
		Email:      "nonexistent@example.com",
		Timezone:   "Europe/Paris",
	}

	mock.ExpectExec(`UPDATE contacts SET (.+) WHERE uuid = \$32`).
		WithArgs(
			nonExistentContact.ExternalID, nonExistentContact.Email, nonExistentContact.Timezone,
			nonExistentContact.FirstName, nonExistentContact.LastName, nonExistentContact.Phone, nonExistentContact.AddressLine1, nonExistentContact.AddressLine2,
			nonExistentContact.Country, nonExistentContact.Postcode, nonExistentContact.State, nonExistentContact.JobTitle,
			nonExistentContact.LifetimeValue, nonExistentContact.OrdersCount, nonExistentContact.LastOrderAt,
			nonExistentContact.CustomString1, nonExistentContact.CustomString2, nonExistentContact.CustomString3, nonExistentContact.CustomString4, nonExistentContact.CustomString5,
			nonExistentContact.CustomNumber1, nonExistentContact.CustomNumber2, nonExistentContact.CustomNumber3, nonExistentContact.CustomNumber4, nonExistentContact.CustomNumber5,
			nonExistentContact.CustomDatetime1, nonExistentContact.CustomDatetime2, nonExistentContact.CustomDatetime3, nonExistentContact.CustomDatetime4, nonExistentContact.CustomDatetime5,
			sqlmock.AnyArg(), nonExistentContact.UUID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateContact(context.Background(), nonExistentContact)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
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
