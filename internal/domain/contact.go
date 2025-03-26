package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/tidwall/gjson"
)

// Contact represents a contact in the system
type Contact struct {
	// Required fields
	Email string `json:"email" valid:"required,email"`

	// Optional fields
	ExternalID   NullableString `json:"external_id,omitempty" valid:"optional"`
	Timezone     NullableString `json:"timezone,omitempty" valid:"optional,timezone"`
	Language     NullableString `json:"language,omitempty" valid:"optional"`
	FirstName    NullableString `json:"first_name,omitempty" valid:"optional"`
	LastName     NullableString `json:"last_name,omitempty" valid:"optional"`
	Phone        NullableString `json:"phone,omitempty" valid:"optional"`
	AddressLine1 NullableString `json:"address_line_1,omitempty" valid:"optional"`
	AddressLine2 NullableString `json:"address_line_2,omitempty" valid:"optional"`
	Country      NullableString `json:"country,omitempty" valid:"optional"`
	Postcode     NullableString `json:"postcode,omitempty" valid:"optional"`
	State        NullableString `json:"state,omitempty" valid:"optional"`
	JobTitle     NullableString `json:"job_title,omitempty" valid:"optional"`

	// Commerce related fields
	LifetimeValue NullableFloat64 `json:"lifetime_value,omitempty" valid:"optional"`
	OrdersCount   NullableFloat64 `json:"orders_count,omitempty" valid:"optional"`
	LastOrderAt   NullableTime    `json:"last_order_at,omitempty" valid:"optional"`

	// Custom fields
	CustomString1 NullableString `json:"custom_string_1,omitempty" valid:"optional"`
	CustomString2 NullableString `json:"custom_string_2,omitempty" valid:"optional"`
	CustomString3 NullableString `json:"custom_string_3,omitempty" valid:"optional"`
	CustomString4 NullableString `json:"custom_string_4,omitempty" valid:"optional"`
	CustomString5 NullableString `json:"custom_string_5,omitempty" valid:"optional"`

	CustomNumber1 NullableFloat64 `json:"custom_number_1,omitempty" valid:"optional"`
	CustomNumber2 NullableFloat64 `json:"custom_number_2,omitempty" valid:"optional"`
	CustomNumber3 NullableFloat64 `json:"custom_number_3,omitempty" valid:"optional"`
	CustomNumber4 NullableFloat64 `json:"custom_number_4,omitempty" valid:"optional"`
	CustomNumber5 NullableFloat64 `json:"custom_number_5,omitempty" valid:"optional"`

	CustomDatetime1 NullableTime `json:"custom_datetime_1,omitempty" valid:"optional"`
	CustomDatetime2 NullableTime `json:"custom_datetime_2,omitempty" valid:"optional"`
	CustomDatetime3 NullableTime `json:"custom_datetime_3,omitempty" valid:"optional"`
	CustomDatetime4 NullableTime `json:"custom_datetime_4,omitempty" valid:"optional"`
	CustomDatetime5 NullableTime `json:"custom_datetime_5,omitempty" valid:"optional"`

	CustomJSON1 NullableJSON `json:"custom_json_1,omitempty" valid:"optional"`
	CustomJSON2 NullableJSON `json:"custom_json_2,omitempty" valid:"optional"`
	CustomJSON3 NullableJSON `json:"custom_json_3,omitempty" valid:"optional"`
	CustomJSON4 NullableJSON `json:"custom_json_4,omitempty" valid:"optional"`
	CustomJSON5 NullableJSON `json:"custom_json_5,omitempty" valid:"optional"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// isValidEmail checks if the email is valid
func isValidEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}
	return true
}

// Validate ensures that the contact has all required fields
func (c *Contact) Validate() error {
	// Email is required
	if c.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Email must be valid
	if !isValidEmail(c.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// For database scanning
type dbContact struct {
	Email      string
	ExternalID sql.NullString
	Timezone   sql.NullString
	Language   sql.NullString

	FirstName    sql.NullString
	LastName     sql.NullString
	Phone        sql.NullString
	AddressLine1 sql.NullString
	AddressLine2 sql.NullString
	Country      sql.NullString
	Postcode     sql.NullString
	State        sql.NullString
	JobTitle     sql.NullString

	LifetimeValue sql.NullFloat64
	OrdersCount   sql.NullFloat64
	LastOrderAt   sql.NullTime

	CustomString1 sql.NullString
	CustomString2 sql.NullString
	CustomString3 sql.NullString
	CustomString4 sql.NullString
	CustomString5 sql.NullString

	CustomNumber1 sql.NullFloat64
	CustomNumber2 sql.NullFloat64
	CustomNumber3 sql.NullFloat64
	CustomNumber4 sql.NullFloat64
	CustomNumber5 sql.NullFloat64

	CustomDatetime1 sql.NullTime
	CustomDatetime2 sql.NullTime
	CustomDatetime3 sql.NullTime
	CustomDatetime4 sql.NullTime
	CustomDatetime5 sql.NullTime

	CustomJSON1 []byte
	CustomJSON2 []byte
	CustomJSON3 []byte
	CustomJSON4 []byte
	CustomJSON5 []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ScanContact scans a contact from the database
func ScanContact(scanner interface {
	Scan(dest ...interface{}) error
}) (*Contact, error) {
	var dbc dbContact
	if err := scanner.Scan(
		&dbc.Email,
		&dbc.ExternalID,
		&dbc.Timezone,
		&dbc.Language,
		&dbc.FirstName,
		&dbc.LastName,
		&dbc.Phone,
		&dbc.AddressLine1,
		&dbc.AddressLine2,
		&dbc.Country,
		&dbc.Postcode,
		&dbc.State,
		&dbc.JobTitle,
		&dbc.LifetimeValue,
		&dbc.OrdersCount,
		&dbc.LastOrderAt,
		&dbc.CustomString1,
		&dbc.CustomString2,
		&dbc.CustomString3,
		&dbc.CustomString4,
		&dbc.CustomString5,
		&dbc.CustomNumber1,
		&dbc.CustomNumber2,
		&dbc.CustomNumber3,
		&dbc.CustomNumber4,
		&dbc.CustomNumber5,
		&dbc.CustomDatetime1,
		&dbc.CustomDatetime2,
		&dbc.CustomDatetime3,
		&dbc.CustomDatetime4,
		&dbc.CustomDatetime5,
		&dbc.CustomJSON1,
		&dbc.CustomJSON2,
		&dbc.CustomJSON3,
		&dbc.CustomJSON4,
		&dbc.CustomJSON5,
		&dbc.CreatedAt,
		&dbc.UpdatedAt,
	); err != nil {
		return nil, err
	}

	c := &Contact{
		Email:      dbc.Email,
		ExternalID: NullableString{String: dbc.ExternalID.String, IsNull: !dbc.ExternalID.Valid},
		Timezone:   NullableString{String: dbc.Timezone.String, IsNull: !dbc.Timezone.Valid},
		Language:   NullableString{String: dbc.Language.String, IsNull: !dbc.Language.Valid},

		FirstName: NullableString{
			String: dbc.FirstName.String,
			IsNull: !dbc.FirstName.Valid,
		},
		LastName: NullableString{
			String: dbc.LastName.String,
			IsNull: !dbc.LastName.Valid,
		},
		Phone: NullableString{
			String: dbc.Phone.String,
			IsNull: !dbc.Phone.Valid,
		},
		AddressLine1: NullableString{
			String: dbc.AddressLine1.String,
			IsNull: !dbc.AddressLine1.Valid,
		},
		AddressLine2: NullableString{
			String: dbc.AddressLine2.String,
			IsNull: !dbc.AddressLine2.Valid,
		},
		Country: NullableString{
			String: dbc.Country.String,
			IsNull: !dbc.Country.Valid,
		},
		Postcode: NullableString{
			String: dbc.Postcode.String,
			IsNull: !dbc.Postcode.Valid,
		},
		State: NullableString{
			String: dbc.State.String,
			IsNull: !dbc.State.Valid,
		},
		JobTitle: NullableString{
			String: dbc.JobTitle.String,
			IsNull: !dbc.JobTitle.Valid,
		},

		LifetimeValue: NullableFloat64{
			Float64: dbc.LifetimeValue.Float64,
			IsNull:  !dbc.LifetimeValue.Valid,
		},
		OrdersCount: NullableFloat64{
			Float64: dbc.OrdersCount.Float64,
			IsNull:  !dbc.OrdersCount.Valid,
		},
		LastOrderAt: NullableTime{
			Time:   dbc.LastOrderAt.Time,
			IsNull: !dbc.LastOrderAt.Valid,
		},

		CustomString1: NullableString{
			String: dbc.CustomString1.String,
			IsNull: !dbc.CustomString1.Valid,
		},
		CustomString2: NullableString{
			String: dbc.CustomString2.String,
			IsNull: !dbc.CustomString2.Valid,
		},
		CustomString3: NullableString{
			String: dbc.CustomString3.String,
			IsNull: !dbc.CustomString3.Valid,
		},
		CustomString4: NullableString{
			String: dbc.CustomString4.String,
			IsNull: !dbc.CustomString4.Valid,
		},
		CustomString5: NullableString{
			String: dbc.CustomString5.String,
			IsNull: !dbc.CustomString5.Valid,
		},

		CustomNumber1: NullableFloat64{
			Float64: dbc.CustomNumber1.Float64,
			IsNull:  !dbc.CustomNumber1.Valid,
		},
		CustomNumber2: NullableFloat64{
			Float64: dbc.CustomNumber2.Float64,
			IsNull:  !dbc.CustomNumber2.Valid,
		},
		CustomNumber3: NullableFloat64{
			Float64: dbc.CustomNumber3.Float64,
			IsNull:  !dbc.CustomNumber3.Valid,
		},
		CustomNumber4: NullableFloat64{
			Float64: dbc.CustomNumber4.Float64,
			IsNull:  !dbc.CustomNumber4.Valid,
		},
		CustomNumber5: NullableFloat64{
			Float64: dbc.CustomNumber5.Float64,
			IsNull:  !dbc.CustomNumber5.Valid,
		},

		CustomDatetime1: NullableTime{
			Time:   dbc.CustomDatetime1.Time,
			IsNull: !dbc.CustomDatetime1.Valid,
		},
		CustomDatetime2: NullableTime{
			Time:   dbc.CustomDatetime2.Time,
			IsNull: !dbc.CustomDatetime2.Valid,
		},
		CustomDatetime3: NullableTime{
			Time:   dbc.CustomDatetime3.Time,
			IsNull: !dbc.CustomDatetime3.Valid,
		},
		CustomDatetime4: NullableTime{
			Time:   dbc.CustomDatetime4.Time,
			IsNull: !dbc.CustomDatetime4.Valid,
		},
		CustomDatetime5: NullableTime{
			Time:   dbc.CustomDatetime5.Time,
			IsNull: !dbc.CustomDatetime5.Valid,
		},

		CreatedAt: dbc.CreatedAt,
		UpdatedAt: dbc.UpdatedAt,
	}

	// Handle custom JSON fields
	if len(dbc.CustomJSON1) > 0 {
		var data interface{}
		if err := json.Unmarshal(dbc.CustomJSON1, &data); err == nil {
			c.CustomJSON1 = NullableJSON{Data: data, Valid: true}
		}
	}
	if len(dbc.CustomJSON2) > 0 {
		var data interface{}
		if err := json.Unmarshal(dbc.CustomJSON2, &data); err == nil {
			c.CustomJSON2 = NullableJSON{Data: data, Valid: true}
		}
	}
	if len(dbc.CustomJSON3) > 0 {
		var data interface{}
		if err := json.Unmarshal(dbc.CustomJSON3, &data); err == nil {
			c.CustomJSON3 = NullableJSON{Data: data, Valid: true}
		}
	}
	if len(dbc.CustomJSON4) > 0 {
		var data interface{}
		if err := json.Unmarshal(dbc.CustomJSON4, &data); err == nil {
			c.CustomJSON4 = NullableJSON{Data: data, Valid: true}
		}
	}
	if len(dbc.CustomJSON5) > 0 {
		var data interface{}
		if err := json.Unmarshal(dbc.CustomJSON5, &data); err == nil {
			c.CustomJSON5 = NullableJSON{Data: data, Valid: true}
		}
	}

	return c, nil
}

// GetContactsRequest represents a request to get contacts with filters and pagination
type GetContactsRequest struct {
	// Required fields
	WorkspaceID string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`

	// Optional filters
	Email      string `json:"email,omitempty" valid:"optional,email"`
	ExternalID string `json:"external_id,omitempty" valid:"optional"`
	FirstName  string `json:"first_name,omitempty" valid:"optional"`
	LastName   string `json:"last_name,omitempty" valid:"optional"`
	Phone      string `json:"phone,omitempty" valid:"optional"`
	Country    string `json:"country,omitempty" valid:"optional"`

	// Pagination
	Limit  int    `json:"limit,omitempty" valid:"optional,range(1|100)"`
	Cursor string `json:"cursor,omitempty" valid:"optional"`
}

// GetContactsResponse represents the response from getting contacts
type GetContactsResponse struct {
	Contacts   []*Contact `json:"contacts"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// Validate ensures that the request has all required fields and valid values
func (r *GetContactsRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	// Set default limit if not provided
	if r.Limit == 0 {
		r.Limit = 20
	}

	// Enforce maximum limit
	if r.Limit > 100 {
		r.Limit = 100
	}

	return nil
}

// ContactService provides operations for managing contacts
type ContactService interface {
	// GetContactByEmail retrieves a contact by email
	GetContactByEmail(ctx context.Context, email string) (*Contact, error)

	// GetContactByExternalID retrieves a contact by external ID
	GetContactByExternalID(ctx context.Context, externalID string) (*Contact, error)

	// GetContacts retrieves contacts with filters and pagination
	GetContacts(ctx context.Context, req *GetContactsRequest) (*GetContactsResponse, error)

	// DeleteContact deletes a contact by email
	DeleteContact(ctx context.Context, email string) error

	// BatchImportContacts imports a batch of contacts (create or update)
	BatchImportContacts(ctx context.Context, contacts []*Contact) error

	// UpsertContact creates a new contact or updates an existing one
	// Returns a boolean indicating whether a new contact was created (true) or an existing one was updated (false)
	UpsertContact(ctx context.Context, contact *Contact) (bool, error)
}

type ContactRepository interface {
	// GetContactByEmail retrieves a contact by its email
	GetContactByEmail(ctx context.Context, email string) (*Contact, error)

	// GetContactByExternalID retrieves a contact by its external ID
	GetContactByExternalID(ctx context.Context, externalID string) (*Contact, error)

	// GetContacts retrieves contacts with filters and pagination
	GetContacts(ctx context.Context, req *GetContactsRequest) (*GetContactsResponse, error)

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, email string) error

	// BatchImportContacts inserts or updates multiple contacts in a batch operation
	BatchImportContacts(ctx context.Context, contacts []*Contact) error

	// UpsertContact creates a new contact or updates an existing one
	UpsertContact(ctx context.Context, contact *Contact) (bool, error)
}

// ErrContactNotFound is returned when a contact is not found
type ErrContactNotFound struct {
	Message string
}

func (e *ErrContactNotFound) Error() string {
	return e.Message
}

// FromJSON parses JSON data into a Contact struct
// The JSON data can be provided as a []byte or as a gjson.Result
func FromJSON(data interface{}) (*Contact, error) {
	var jsonResult gjson.Result

	switch v := data.(type) {
	case []byte:
		jsonResult = gjson.ParseBytes(v)
	case gjson.Result:
		jsonResult = v
	case string:
		jsonResult = gjson.Parse(v)
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	// Extract required fields
	email := jsonResult.Get("email").String()
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Validate email format
	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Create the contact with required fields
	contact := &Contact{
		Email: email,
	}
	// Parse nullable string fields
	parseNullableString(jsonResult, "external_id", &contact.ExternalID)
	parseNullableString(jsonResult, "timezone", &contact.Timezone)
	parseNullableString(jsonResult, "language", &contact.Language)
	parseNullableString(jsonResult, "first_name", &contact.FirstName)
	parseNullableString(jsonResult, "last_name", &contact.LastName)
	parseNullableString(jsonResult, "phone", &contact.Phone)
	parseNullableString(jsonResult, "address_line_1", &contact.AddressLine1)
	parseNullableString(jsonResult, "address_line_2", &contact.AddressLine2)
	parseNullableString(jsonResult, "country", &contact.Country)
	parseNullableString(jsonResult, "postcode", &contact.Postcode)
	parseNullableString(jsonResult, "state", &contact.State)
	parseNullableString(jsonResult, "job_title", &contact.JobTitle)

	// Parse custom string fields
	parseNullableString(jsonResult, "custom_string_1", &contact.CustomString1)
	parseNullableString(jsonResult, "custom_string_2", &contact.CustomString2)
	parseNullableString(jsonResult, "custom_string_3", &contact.CustomString3)
	parseNullableString(jsonResult, "custom_string_4", &contact.CustomString4)
	parseNullableString(jsonResult, "custom_string_5", &contact.CustomString5)

	// Parse nullable number fields
	parseNullableFloat(jsonResult, "lifetime_value", &contact.LifetimeValue)
	parseNullableFloat(jsonResult, "orders_count", &contact.OrdersCount)

	// Parse custom number fields
	parseNullableFloat(jsonResult, "custom_number_1", &contact.CustomNumber1)
	parseNullableFloat(jsonResult, "custom_number_2", &contact.CustomNumber2)
	parseNullableFloat(jsonResult, "custom_number_3", &contact.CustomNumber3)
	parseNullableFloat(jsonResult, "custom_number_4", &contact.CustomNumber4)
	parseNullableFloat(jsonResult, "custom_number_5", &contact.CustomNumber5)

	// Parse date fields
	parseNullableTime(jsonResult, "last_order_at", &contact.LastOrderAt)

	// Parse custom datetime fields
	parseNullableTime(jsonResult, "custom_datetime_1", &contact.CustomDatetime1)
	parseNullableTime(jsonResult, "custom_datetime_2", &contact.CustomDatetime2)
	parseNullableTime(jsonResult, "custom_datetime_3", &contact.CustomDatetime3)
	parseNullableTime(jsonResult, "custom_datetime_4", &contact.CustomDatetime4)
	parseNullableTime(jsonResult, "custom_datetime_5", &contact.CustomDatetime5)

	return contact, nil
}

// Helper functions for parsing nullable fields from JSON
func parseNullableString(result gjson.Result, field string, target *NullableString) {
	if value := result.Get(field); value.Exists() {
		if value.Type == gjson.Null {
			*target = NullableString{IsNull: true}
		} else {
			*target = NullableString{String: value.String(), IsNull: false}
		}
	}
}

func parseNullableFloat(result gjson.Result, field string, target *NullableFloat64) {
	if value := result.Get(field); value.Exists() {
		if value.Type == gjson.Null {
			*target = NullableFloat64{IsNull: true}
		} else {
			*target = NullableFloat64{Float64: value.Float(), IsNull: false}
		}
	}
}

func parseNullableTime(result gjson.Result, field string, target *NullableTime) {
	if value := result.Get(field); value.Exists() {
		if value.Type == gjson.Null {
			*target = NullableTime{IsNull: true}
		} else {
			t, err := time.Parse(time.RFC3339, value.String())
			if err == nil {
				*target = NullableTime{Time: t, IsNull: false}
			} else {
				*target = NullableTime{IsNull: true}
			}
		}
	}
}
