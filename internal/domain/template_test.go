package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTemplateCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category TemplateCategory
		wantErr  bool
	}{
		{
			name:     "valid marketing category",
			category: TemplateCategoryMarketing,
			wantErr:  false,
		},
		{
			name:     "valid transactional category",
			category: TemplateCategoryTransactional,
			wantErr:  false,
		},
		{
			name:     "valid welcome category",
			category: TemplateCategoryWelcome,
			wantErr:  false,
		},
		{
			name:     "valid opt_in category",
			category: TemplateCategoryOptIn,
			wantErr:  false,
		},
		{
			name:     "valid unsubscribe category",
			category: TemplateCategoryUnsubscribe,
			wantErr:  false,
		},
		{
			name:     "valid bounce category",
			category: TemplateCategoryBounce,
			wantErr:  false,
		},
		{
			name:     "valid blocklist category",
			category: TemplateCategoryBlocklist,
			wantErr:  false,
		},
		{
			name:     "valid other category",
			category: TemplateCategoryOther,
			wantErr:  false,
		},
		{
			name:     "invalid category",
			category: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	now := time.Now()

	createValidTemplate := func() *Template {
		return &Template{
			ID:      "test123",
			Name:    "Test Template",
			Version: 1,
			Channel: "email",
			Email: &EmailTemplate{
				FromAddress:      "test@example.com",
				FromName:         "Test Sender",
				Subject:          "Test Subject",
				Content:          "<html>Test content</html>",
				VisualEditorTree: MapOfAny{"type": "root"},
			},
			Category:  string(TemplateCategoryMarketing),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	tests := []struct {
		name     string
		template *Template
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: createValidTemplate(),
			wantErr:  false,
		},
		{
			name: "invalid template - missing ID",
			template: func() *Template {
				t := createValidTemplate()
				t.ID = ""
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - missing name",
			template: func() *Template {
				t := createValidTemplate()
				t.Name = ""
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - invalid email",
			template: func() *Template {
				t := createValidTemplate()
				t.Email.FromAddress = "invalid-email"
				return t
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateReference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ref     *TemplateReference
		wantErr bool
	}{
		{
			name: "valid reference",
			ref: &TemplateReference{
				ID:      "test123",
				Version: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid reference - missing ID",
			ref: &TemplateReference{
				ID:      "",
				Version: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid reference - missing version",
			ref: &TemplateReference{
				ID:      "test123",
				Version: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateReference_Scan_Value(t *testing.T) {
	ref := &TemplateReference{
		ID:      "test123",
		Version: 1,
	}

	// Test Value() method
	value, err := ref.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test Scan() method with []byte
	bytes, err := json.Marshal(ref)
	assert.NoError(t, err)

	newRef := &TemplateReference{}
	err = newRef.Scan(bytes)
	assert.NoError(t, err)
	assert.Equal(t, ref.ID, newRef.ID)
	assert.Equal(t, ref.Version, newRef.Version)

	// Test Scan() method with string
	err = newRef.Scan(string(bytes))
	assert.NoError(t, err)
	assert.Equal(t, ref.ID, newRef.ID)
	assert.Equal(t, ref.Version, newRef.Version)

	// Test Scan() method with nil
	err = newRef.Scan(nil)
	assert.NoError(t, err)
}

func TestEmailTemplate_Validate(t *testing.T) {
	validEmail := &EmailTemplate{
		FromAddress:      "test@example.com",
		FromName:         "Test Sender",
		Subject:          "Test Subject",
		Content:          "<html>Test content</html>",
		VisualEditorTree: MapOfAny{"type": "root"},
	}

	tests := []struct {
		name    string
		email   *EmailTemplate
		wantErr bool
	}{
		{
			name:    "valid email template",
			email:   validEmail,
			wantErr: false,
		},
		{
			name: "invalid email template - missing from address",
			email: func() *EmailTemplate {
				e := *validEmail
				e.FromAddress = ""
				return &e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - invalid email format",
			email: func() *EmailTemplate {
				e := *validEmail
				e.FromAddress = "invalid-email"
				return &e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing from name",
			email: func() *EmailTemplate {
				e := *validEmail
				e.FromName = ""
				return &e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing subject",
			email: func() *EmailTemplate {
				e := *validEmail
				e.Subject = ""
				return &e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing content",
			email: func() *EmailTemplate {
				e := *validEmail
				e.Content = ""
				return &e
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailTemplate_Scan_Value(t *testing.T) {
	email := &EmailTemplate{
		FromAddress:      "test@example.com",
		FromName:         "Test Sender",
		Subject:          "Test Subject",
		Content:          "<html>Test content</html>",
		VisualEditorTree: MapOfAny{"type": "root"},
	}

	// Test Value() method
	value, err := email.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test Scan() method with []byte
	bytes, err := json.Marshal(email)
	assert.NoError(t, err)

	newEmail := &EmailTemplate{}
	err = newEmail.Scan(bytes)
	assert.NoError(t, err)
	assert.Equal(t, email.FromAddress, newEmail.FromAddress)
	assert.Equal(t, email.FromName, newEmail.FromName)
	assert.Equal(t, email.Subject, newEmail.Subject)
	assert.Equal(t, email.Content, newEmail.Content)

	// Test Scan() method with string
	err = newEmail.Scan(string(bytes))
	assert.NoError(t, err)
	assert.Equal(t, email.FromAddress, newEmail.FromAddress)
	assert.Equal(t, email.FromName, newEmail.FromName)
	assert.Equal(t, email.Subject, newEmail.Subject)
	assert.Equal(t, email.Content, newEmail.Content)

	// Test Scan() method with nil
	err = newEmail.Scan(nil)
	assert.NoError(t, err)
}
