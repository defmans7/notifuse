package domain

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
)

type TemplateCategory string

const (
	TemplateCategoryMarketing     TemplateCategory = "marketing"
	TemplateCategoryTransactional TemplateCategory = "transactional"
	TemplateCategoryWelcome       TemplateCategory = "welcome"
	TemplateCategoryOptIn         TemplateCategory = "opt_in"
	TemplateCategoryUnsubscribe   TemplateCategory = "unsubscribe"
	TemplateCategoryBounce        TemplateCategory = "bounce"
	TemplateCategoryBlocklist     TemplateCategory = "blocklist"
	TemplateCategoryOther         TemplateCategory = "other"
)

func (t TemplateCategory) Validate() error {
	switch t {
	case TemplateCategoryMarketing, TemplateCategoryTransactional, TemplateCategoryWelcome, TemplateCategoryOptIn, TemplateCategoryUnsubscribe, TemplateCategoryBounce, TemplateCategoryBlocklist, TemplateCategoryOther:
		return nil
	}
	return fmt.Errorf("invalid template category: %s", t)
}

type Template struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Version         int64          `json:"version"`
	Channel         string         `json:"channel"` // email for now
	Email           *EmailTemplate `json:"email"`
	Category        string         `json:"category"`
	TemplateMacroID *string        `json:"template_macro_id,omitempty"`
	UTMSource       *string        `json:"utm_source,omitempty"`
	UTMMedium       *string        `json:"utm_medium,omitempty"`
	UTMCampaign     *string        `json:"utm_campaign,omitempty"`
	TestData        MapOfAny       `json:"test_data,omitempty"`
	Settings        MapOfAny       `json:"settings,omitempty"` // Channels specific 3rd-party settings
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (t *Template) Validate() error {
	// First validate the template itself
	if t.ID == "" {
		return fmt.Errorf("invalid template: id is required")
	}
	if !govalidator.IsAlphanumeric(t.ID) {
		return fmt.Errorf("invalid template: id must be alphanumeric")
	}
	if len(t.ID) > 20 {
		return fmt.Errorf("invalid template: id length must be between 1 and 20")
	}

	if t.Name == "" {
		return fmt.Errorf("invalid template: name is required")
	}
	if len(t.Name) > 255 {
		return fmt.Errorf("invalid template: name length must be between 1 and 255")
	}

	if t.Version <= 0 {
		return fmt.Errorf("invalid template: version is required and must be positive")
	}

	if t.Channel == "" {
		return fmt.Errorf("invalid template: channel is required")
	}
	if len(t.Channel) > 20 {
		return fmt.Errorf("invalid template: channel length must be between 1 and 20")
	}

	if t.Category == "" {
		return fmt.Errorf("invalid template: category is required")
	}
	if len(t.Category) > 20 {
		return fmt.Errorf("invalid template: category length must be between 1 and 20")
	}

	// Then validate the email template
	if t.Email == nil {
		return fmt.Errorf("invalid template: email is required")
	}

	if err := t.Email.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	return nil
}

type TemplateReference struct {
	ID      string `json:"id"`
	Version int64  `json:"version"`
}

func (t *TemplateReference) Validate() error {
	// Validate the template reference
	if t.ID == "" {
		return fmt.Errorf("invalid template reference: id is required")
	}
	if !govalidator.IsAlphanumeric(t.ID) {
		return fmt.Errorf("invalid template reference: id must be alphanumeric")
	}
	if len(t.ID) > 20 {
		return fmt.Errorf("invalid template reference: id length must be between 1 and 20")
	}

	if t.Version <= 0 {
		return fmt.Errorf("invalid template reference: version is required and must be positive")
	}

	return nil
}

// scan implements the sql.Scanner interface
func (t *TemplateReference) Scan(val interface{}) error {
	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, t)
}

// value implements the driver.Valuer interface
func (t TemplateReference) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type EmailTemplate struct {
	FromAddress      string   `json:"from_address"`
	FromName         string   `json:"from_name"`
	ReplyTo          *string  `json:"reply_to,omitempty"`
	Subject          string   `json:"subject"`
	SubjectPreview   *string  `json:"subject_preview,omitempty"`
	Content          string   `json:"content"` // html
	VisualEditorTree MapOfAny `json:"visual_editor_tree"`
	Text             *string  `json:"text,omitempty"`
}

func (e *EmailTemplate) Validate() error {
	// Validate required fields
	if e.FromAddress == "" {
		return fmt.Errorf("invalid email template: from_address is required")
	}
	if !govalidator.IsEmail(e.FromAddress) {
		return fmt.Errorf("invalid email template: from_address is not a valid email")
	}
	if e.FromName == "" {
		return fmt.Errorf("invalid email template: from_name is required")
	}
	if len(e.FromName) > 255 {
		return fmt.Errorf("invalid email template: from_name length must be between 1 and 255")
	}
	if e.Subject == "" {
		return fmt.Errorf("invalid email template: subject is required")
	}
	if len(e.Subject) > 255 {
		return fmt.Errorf("invalid email template: subject length must be between 1 and 255")
	}
	if e.Content == "" {
		return fmt.Errorf("invalid email template: content is required")
	}
	if e.VisualEditorTree == nil {
		return fmt.Errorf("invalid email template: visual_editor_tree is required")
	}

	// Validate optional fields
	if e.ReplyTo != nil && !govalidator.IsEmail(*e.ReplyTo) {
		return fmt.Errorf("invalid email template: reply_to is not a valid email")
	}
	if e.SubjectPreview != nil && len(*e.SubjectPreview) > 255 {
		return fmt.Errorf("invalid email template: subject_preview length must be between 1 and 255")
	}

	return nil
}

func (x *EmailTemplate) Scan(val interface{}) error {
	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x EmailTemplate) Value() (driver.Value, error) {
	return json.Marshal(x)
}
