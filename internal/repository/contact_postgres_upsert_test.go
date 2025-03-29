package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestUpsertContact(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"
	workspaceID := "workspace123"

	testContact := &domain.Contact{
		Email:           email,
		ExternalID:      &domain.NullableString{String: "ext123", IsNull: false},
		Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
		Language:        &domain.NullableString{String: "en-US", IsNull: false},
		FirstName:       &domain.NullableString{String: "John", IsNull: false},
		LastName:        &domain.NullableString{String: "Doe", IsNull: false},
		Phone:           &domain.NullableString{String: "", IsNull: true},
		AddressLine1:    &domain.NullableString{String: "", IsNull: true},
		AddressLine2:    &domain.NullableString{String: "", IsNull: true},
		Country:         &domain.NullableString{String: "", IsNull: true},
		Postcode:        &domain.NullableString{String: "", IsNull: true},
		State:           &domain.NullableString{String: "", IsNull: true},
		JobTitle:        &domain.NullableString{String: "", IsNull: true},
		LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
		LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomString1:   &domain.NullableString{String: "", IsNull: true},
		CustomString2:   &domain.NullableString{String: "", IsNull: true},
		CustomString3:   &domain.NullableString{String: "", IsNull: true},
		CustomString4:   &domain.NullableString{String: "", IsNull: true},
		CustomString5:   &domain.NullableString{String: "", IsNull: true},
		CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Test case 1: Insert new contact
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			testContact.Email,
			testContact.ExternalID.String,
			testContact.Timezone.String,
			testContact.Language.String,
			testContact.FirstName.String,
			testContact.LastName.String,
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			[]byte(nil),
			[]byte(nil),
			[]byte(nil),
			[]byte(nil),
			[]byte(nil),
			testContact.CreatedAt,
			testContact.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err := repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.NoError(t, err)
	assert.True(t, created)

	// Test case 2: Update existing contact with all fields set
	contactWithAllFields := &domain.Contact{
		Email:           "allfields@example.com",
		ExternalID:      &domain.NullableString{String: "ext456", IsNull: false},
		Timezone:        &domain.NullableString{String: "America/New_York", IsNull: false},
		Language:        &domain.NullableString{String: "fr-FR", IsNull: false},
		FirstName:       &domain.NullableString{String: "Jane", IsNull: false},
		LastName:        &domain.NullableString{String: "Smith", IsNull: false},
		Phone:           &domain.NullableString{String: "+1234567890", IsNull: false},
		AddressLine1:    &domain.NullableString{String: "123 Main St", IsNull: false},
		AddressLine2:    &domain.NullableString{String: "Apt 4B", IsNull: false},
		Country:         &domain.NullableString{String: "USA", IsNull: false},
		Postcode:        &domain.NullableString{String: "10001", IsNull: false},
		State:           &domain.NullableString{String: "NY", IsNull: false},
		JobTitle:        &domain.NullableString{String: "Engineer", IsNull: false},
		LifetimeValue:   &domain.NullableFloat64{Float64: 1000.50, IsNull: false},
		OrdersCount:     &domain.NullableFloat64{Float64: 5, IsNull: false},
		LastOrderAt:     &domain.NullableTime{Time: now, IsNull: false},
		CustomString1:   &domain.NullableString{String: "custom1", IsNull: false},
		CustomString2:   &domain.NullableString{String: "custom2", IsNull: false},
		CustomString3:   &domain.NullableString{String: "custom3", IsNull: false},
		CustomString4:   &domain.NullableString{String: "custom4", IsNull: false},
		CustomString5:   &domain.NullableString{String: "custom5", IsNull: false},
		CustomNumber1:   &domain.NullableFloat64{Float64: 1.1, IsNull: false},
		CustomNumber2:   &domain.NullableFloat64{Float64: 2.2, IsNull: false},
		CustomNumber3:   &domain.NullableFloat64{Float64: 3.3, IsNull: false},
		CustomNumber4:   &domain.NullableFloat64{Float64: 4.4, IsNull: false},
		CustomNumber5:   &domain.NullableFloat64{Float64: 5.5, IsNull: false},
		CustomDatetime1: &domain.NullableTime{Time: now, IsNull: false},
		CustomDatetime2: &domain.NullableTime{Time: now, IsNull: false},
		CustomDatetime3: &domain.NullableTime{Time: now, IsNull: false},
		CustomDatetime4: &domain.NullableTime{Time: now, IsNull: false},
		CustomDatetime5: &domain.NullableTime{Time: now, IsNull: false},
		CustomJSON1:     &domain.NullableJSON{Data: map[string]interface{}{"key1": "value1"}, IsNull: false},
		CustomJSON2:     &domain.NullableJSON{Data: map[string]interface{}{"key2": "value2"}, IsNull: false},
		CustomJSON3:     &domain.NullableJSON{Data: map[string]interface{}{"key3": "value3"}, IsNull: false},
		CustomJSON4:     &domain.NullableJSON{Data: map[string]interface{}{"key4": "value4"}, IsNull: false},
		CustomJSON5:     &domain.NullableJSON{Data: map[string]interface{}{"key5": "value5"}, IsNull: false},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone", "language",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
		"created_at", "updated_at",
	}).AddRow(
		"allfields@example.com", "old-ext-id", "UTC", "en-US",
		"Old", "Name", "", "", "",
		"", "", "", "",
		0.0, 0.0, time.Time{},
		"", "", "", "", "",
		0.0, 0.0, 0.0, 0.0, 0.0,
		time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{},
		[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("allfields@example.com").
		WillReturnRows(rows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			contactWithAllFields.Email,
			contactWithAllFields.ExternalID.String,
			contactWithAllFields.Timezone.String,
			contactWithAllFields.Language.String,
			contactWithAllFields.FirstName.String,
			contactWithAllFields.LastName.String,
			sql.NullString{String: contactWithAllFields.Phone.String, Valid: true},
			sql.NullString{String: contactWithAllFields.AddressLine1.String, Valid: true},
			sql.NullString{String: contactWithAllFields.AddressLine2.String, Valid: true},
			sql.NullString{String: contactWithAllFields.Country.String, Valid: true},
			sql.NullString{String: contactWithAllFields.Postcode.String, Valid: true},
			sql.NullString{String: contactWithAllFields.State.String, Valid: true},
			sql.NullString{String: contactWithAllFields.JobTitle.String, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.LifetimeValue.Float64, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.OrdersCount.Float64, Valid: true},
			sql.NullTime{Time: contactWithAllFields.LastOrderAt.Time, Valid: true},
			sql.NullString{String: contactWithAllFields.CustomString1.String, Valid: true},
			sql.NullString{String: contactWithAllFields.CustomString2.String, Valid: true},
			sql.NullString{String: contactWithAllFields.CustomString3.String, Valid: true},
			sql.NullString{String: contactWithAllFields.CustomString4.String, Valid: true},
			sql.NullString{String: contactWithAllFields.CustomString5.String, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.CustomNumber1.Float64, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.CustomNumber2.Float64, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.CustomNumber3.Float64, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.CustomNumber4.Float64, Valid: true},
			sql.NullFloat64{Float64: contactWithAllFields.CustomNumber5.Float64, Valid: true},
			sql.NullTime{Time: contactWithAllFields.CustomDatetime1.Time, Valid: true},
			sql.NullTime{Time: contactWithAllFields.CustomDatetime2.Time, Valid: true},
			sql.NullTime{Time: contactWithAllFields.CustomDatetime3.Time, Valid: true},
			sql.NullTime{Time: contactWithAllFields.CustomDatetime4.Time, Valid: true},
			sql.NullTime{Time: contactWithAllFields.CustomDatetime5.Time, Valid: true},
			[]byte(`{"key1":"value1"}`),
			[]byte(`{"key2":"value2"}`),
			[]byte(`{"key3":"value3"}`),
			[]byte(`{"key4":"value4"}`),
			[]byte(`{"key5":"value5"}`),
			contactWithAllFields.CreatedAt,
			contactWithAllFields.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err = repo.UpsertContact(context.Background(), workspaceID, contactWithAllFields)
	require.NoError(t, err)
	assert.False(t, created)

	// Test case 3: Error checking if contact exists
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(errors.New("check error"))

	created, err = repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if contact exists")

	// Test case 4: Error upserting contact
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WillReturnError(errors.New("upsert error"))

	created, err = repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert contact")

	// Test case 5: Contact with invalid JSON data
	contactWithInvalidJSON := &domain.Contact{
		Email:           "invalidjson@example.com",
		ExternalID:      &domain.NullableString{String: "ext789", IsNull: false},
		Timezone:        &domain.NullableString{String: "UTC", IsNull: false},
		Language:        &domain.NullableString{String: "en-US", IsNull: false},
		FirstName:       &domain.NullableString{String: "John", IsNull: false},
		LastName:        &domain.NullableString{String: "Doe", IsNull: false},
		Phone:           &domain.NullableString{String: "", IsNull: true},
		AddressLine1:    &domain.NullableString{String: "", IsNull: true},
		AddressLine2:    &domain.NullableString{String: "", IsNull: true},
		Country:         &domain.NullableString{String: "", IsNull: true},
		Postcode:        &domain.NullableString{String: "", IsNull: true},
		State:           &domain.NullableString{String: "", IsNull: true},
		JobTitle:        &domain.NullableString{String: "", IsNull: true},
		LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
		LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomString1:   &domain.NullableString{String: "", IsNull: true},
		CustomString2:   &domain.NullableString{String: "", IsNull: true},
		CustomString3:   &domain.NullableString{String: "", IsNull: true},
		CustomString4:   &domain.NullableString{String: "", IsNull: true},
		CustomString5:   &domain.NullableString{String: "", IsNull: true},
		CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomJSON1:     &domain.NullableJSON{Data: make(chan int), IsNull: false}, // Invalid JSON data
		CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("invalidjson@example.com").
		WillReturnError(sql.ErrNoRows)

	created, err = repo.UpsertContact(context.Background(), workspaceID, contactWithInvalidJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal CustomJSON1")

	assert.NoError(t, mock.ExpectationsWereMet())
}
