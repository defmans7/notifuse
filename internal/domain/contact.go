package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
)

// Contact represents a contact in the system
type Contact struct {
	// Required fields
	UUID       string `json:"uuid" valid:"required,uuid"`
	ExternalID string `json:"external_id" valid:"required"`
	Email      string `json:"email" valid:"required,email"`
	Timezone   string `json:"timezone" valid:"required,timezone"`

	// Optional fields
	FirstName    string `json:"first_name,omitempty" valid:"optional"`
	LastName     string `json:"last_name,omitempty" valid:"optional"`
	Phone        string `json:"phone,omitempty" valid:"optional"`
	AddressLine1 string `json:"address_line_1,omitempty" valid:"optional"`
	AddressLine2 string `json:"address_line_2,omitempty" valid:"optional"`
	Country      string `json:"country,omitempty" valid:"optional"`
	Postcode     string `json:"postcode,omitempty" valid:"optional"`
	State        string `json:"state,omitempty" valid:"optional"`
	JobTitle     string `json:"job_title,omitempty" valid:"optional"`

	// Commerce related fields
	LifetimeValue float64   `json:"lifetime_value,omitempty" valid:"optional"`
	OrdersCount   int       `json:"orders_count,omitempty" valid:"optional"`
	LastOrderAt   time.Time `json:"last_order_at,omitempty" valid:"optional"`

	// Custom fields
	CustomString1 string `json:"custom_string_1,omitempty" valid:"optional"`
	CustomString2 string `json:"custom_string_2,omitempty" valid:"optional"`
	CustomString3 string `json:"custom_string_3,omitempty" valid:"optional"`
	CustomString4 string `json:"custom_string_4,omitempty" valid:"optional"`
	CustomString5 string `json:"custom_string_5,omitempty" valid:"optional"`

	CustomNumber1 float64 `json:"custom_number_1,omitempty" valid:"optional"`
	CustomNumber2 float64 `json:"custom_number_2,omitempty" valid:"optional"`
	CustomNumber3 float64 `json:"custom_number_3,omitempty" valid:"optional"`
	CustomNumber4 float64 `json:"custom_number_4,omitempty" valid:"optional"`
	CustomNumber5 float64 `json:"custom_number_5,omitempty" valid:"optional"`

	CustomDatetime1 time.Time `json:"custom_datetime_1,omitempty" valid:"optional"`
	CustomDatetime2 time.Time `json:"custom_datetime_2,omitempty" valid:"optional"`
	CustomDatetime3 time.Time `json:"custom_datetime_3,omitempty" valid:"optional"`
	CustomDatetime4 time.Time `json:"custom_datetime_4,omitempty" valid:"optional"`
	CustomDatetime5 time.Time `json:"custom_datetime_5,omitempty" valid:"optional"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate performs validation on the contact fields
func (c *Contact) Validate() error {
	// Register custom validators
	govalidator.TagMap["timezone"] = govalidator.Validator(func(str string) bool {
		return IsValidTimezone(str)
	})

	if _, err := govalidator.ValidateStruct(c); err != nil {
		return fmt.Errorf("invalid contact: %w", err)
	}
	return nil
}

// For database scanning
type dbContact struct {
	UUID       string
	ExternalID string
	Email      string
	Timezone   string

	FirstName    string
	LastName     string
	Phone        string
	AddressLine1 string
	AddressLine2 string
	Country      string
	Postcode     string
	State        string
	JobTitle     string

	LifetimeValue float64
	OrdersCount   int
	LastOrderAt   time.Time

	CustomString1 string
	CustomString2 string
	CustomString3 string
	CustomString4 string
	CustomString5 string

	CustomNumber1 float64
	CustomNumber2 float64
	CustomNumber3 float64
	CustomNumber4 float64
	CustomNumber5 float64

	CustomDatetime1 time.Time
	CustomDatetime2 time.Time
	CustomDatetime3 time.Time
	CustomDatetime4 time.Time
	CustomDatetime5 time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ScanContact scans a contact from the database
func ScanContact(scanner interface {
	Scan(dest ...interface{}) error
}) (*Contact, error) {
	var dbc dbContact
	if err := scanner.Scan(
		&dbc.UUID,
		&dbc.ExternalID,
		&dbc.Email,
		&dbc.Timezone,
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
		&dbc.CreatedAt,
		&dbc.UpdatedAt,
	); err != nil {
		return nil, err
	}

	c := &Contact{
		UUID:            dbc.UUID,
		ExternalID:      dbc.ExternalID,
		Email:           dbc.Email,
		Timezone:        dbc.Timezone,
		FirstName:       dbc.FirstName,
		LastName:        dbc.LastName,
		Phone:           dbc.Phone,
		AddressLine1:    dbc.AddressLine1,
		AddressLine2:    dbc.AddressLine2,
		Country:         dbc.Country,
		Postcode:        dbc.Postcode,
		State:           dbc.State,
		JobTitle:        dbc.JobTitle,
		LifetimeValue:   dbc.LifetimeValue,
		OrdersCount:     dbc.OrdersCount,
		LastOrderAt:     dbc.LastOrderAt,
		CustomString1:   dbc.CustomString1,
		CustomString2:   dbc.CustomString2,
		CustomString3:   dbc.CustomString3,
		CustomString4:   dbc.CustomString4,
		CustomString5:   dbc.CustomString5,
		CustomNumber1:   dbc.CustomNumber1,
		CustomNumber2:   dbc.CustomNumber2,
		CustomNumber3:   dbc.CustomNumber3,
		CustomNumber4:   dbc.CustomNumber4,
		CustomNumber5:   dbc.CustomNumber5,
		CustomDatetime1: dbc.CustomDatetime1,
		CustomDatetime2: dbc.CustomDatetime2,
		CustomDatetime3: dbc.CustomDatetime3,
		CustomDatetime4: dbc.CustomDatetime4,
		CustomDatetime5: dbc.CustomDatetime5,
		CreatedAt:       dbc.CreatedAt,
		UpdatedAt:       dbc.UpdatedAt,
	}

	return c, nil
}

type ContactRepository interface {
	// CreateContact creates a new contact in the database
	CreateContact(ctx context.Context, contact *Contact) error

	// GetContactByUUID retrieves a contact by its UUID
	GetContactByUUID(ctx context.Context, uuid string) (*Contact, error)

	// GetContactByEmail retrieves a contact by its email
	GetContactByEmail(ctx context.Context, email string) (*Contact, error)

	// GetContactByExternalID retrieves a contact by its external ID
	GetContactByExternalID(ctx context.Context, externalID string) (*Contact, error)

	// GetContacts retrieves all contacts
	GetContacts(ctx context.Context) ([]*Contact, error)

	// UpdateContact updates an existing contact
	UpdateContact(ctx context.Context, contact *Contact) error

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, uuid string) error

	// BatchImportContacts inserts or updates multiple contacts in a batch operation
	BatchImportContacts(ctx context.Context, contacts []*Contact) error
}

// ErrContactNotFound is returned when a contact is not found
type ErrContactNotFound struct {
	Message string
}

func (e *ErrContactNotFound) Error() string {
	return e.Message
}
