package domain

import (
	"encoding/json"
	"net/url"
	"strconv"
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
		{
			name: "invalid template - missing channel",
			template: func() *Template {
				t := createValidTemplate()
				t.Channel = ""
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - channel too long",
			template: func() *Template {
				t := createValidTemplate()
				t.Channel = "this_channel_name_is_too_long_for_validation"
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - missing category",
			template: func() *Template {
				t := createValidTemplate()
				t.Category = ""
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - category too long",
			template: func() *Template {
				t := createValidTemplate()
				t.Category = "this_category_name_is_too_long_for_validation"
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - zero version",
			template: func() *Template {
				t := createValidTemplate()
				t.Version = 0
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - negative version",
			template: func() *Template {
				t := createValidTemplate()
				t.Version = -1
				return t
			}(),
			wantErr: true,
		},
		{
			name: "invalid template - missing email",
			template: func() *Template {
				t := createValidTemplate()
				t.Email = nil
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
	tests := []struct {
		name     string
		template *EmailTemplate
		wantErr  bool
	}{
		{
			name: "valid template",
			template: &EmailTemplate{
				FromAddress:      "test@example.com",
				FromName:         "Test Sender",
				Subject:          "Test Subject",
				Content:          "<html>Test content</html>",
				VisualEditorTree: MapOfAny{"type": "root"},
			},
			wantErr: false,
		},
		{
			name: "invalid email template - missing from address",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				}
				e.FromAddress = ""
				return e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - invalid email format",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				}
				e.FromAddress = "invalid-email"
				return e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing from name",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				}
				e.FromName = ""
				return e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing subject",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				}
				e.Subject = ""
				return e
			}(),
			wantErr: true,
		},
		{
			name: "invalid email template - missing content",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				}
				e.Content = ""
				return e
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

func TestCreateTemplateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *CreateTemplateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: &CreateTemplateRequest{
				WorkspaceID: "",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "ID too long",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "this_id_is_way_too_long_for_the_validation_to_pass_properly",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing name",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing channel",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing category",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: "",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email:       nil,
				Category:    string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "invalid email template",
			request: &CreateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "invalid-email",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, workspaceID, err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, template)
				assert.Equal(t, tt.request.WorkspaceID, workspaceID)
				assert.Equal(t, tt.request.ID, template.ID)
				assert.Equal(t, tt.request.Name, template.Name)
				assert.Equal(t, int64(1), template.Version) // Should always be 1 for new templates
				assert.Equal(t, tt.request.Channel, template.Channel)
				assert.Equal(t, tt.request.Email, template.Email)
				assert.Equal(t, tt.request.Category, template.Category)
			}
		})
	}
}

func TestGetTemplatesRequest_FromURLParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams url.Values
		wantErr     bool
	}{
		{
			name: "valid request",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			wantErr: false,
		},
		{
			name:        "missing workspace_id",
			queryParams: url.Values{},
			wantErr:     true,
		},
		{
			name: "workspace_id too long",
			queryParams: url.Values{
				"workspace_id": []string{"workspace_id_that_is_way_too_long_for_validation"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &GetTemplatesRequest{}
			err := req.FromURLParams(tt.queryParams)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.queryParams.Get("workspace_id"), req.WorkspaceID)
			}
		})
	}
}

func TestGetTemplateRequest_FromURLParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams url.Values
		wantErr     bool
	}{
		{
			name: "valid request with ID only",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"template123"},
			},
			wantErr: false,
		},
		{
			name: "valid request with ID and version",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"template123"},
				"version":      []string{"2"},
			},
			wantErr: false,
		},
		{
			name: "missing workspace_id",
			queryParams: url.Values{
				"id": []string{"template123"},
			},
			wantErr: true,
		},
		{
			name: "missing id",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			wantErr: true,
		},
		{
			name: "id too long",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"template_id_that_is_way_too_long_for_validation_to_pass_properly"},
			},
			wantErr: true,
		},
		{
			name: "invalid version format",
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"template123"},
				"version":      []string{"not-a-number"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &GetTemplateRequest{}
			err := req.FromURLParams(tt.queryParams)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.queryParams.Get("workspace_id"), req.WorkspaceID)
				assert.Equal(t, tt.queryParams.Get("id"), req.ID)
				if versionStr := tt.queryParams.Get("version"); versionStr != "" {
					version, _ := strconv.ParseInt(versionStr, 10, 64)
					assert.Equal(t, version, req.Version)
				}
			}
		})
	}
}

func TestUpdateTemplateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *UpdateTemplateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: &UpdateTemplateRequest{
				WorkspaceID: "",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "ID too long",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "this_id_is_way_too_long_for_the_validation_to_pass_properly",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing name",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing channel",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "missing category",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "test@example.com",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: "",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email:       nil,
				Category:    string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
		{
			name: "invalid email template",
			request: &UpdateTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
				Name:        "Test Template",
				Channel:     "email",
				Email: &EmailTemplate{
					FromAddress:      "invalid-email",
					FromName:         "Test Sender",
					Subject:          "Test Subject",
					Content:          "<html>Test content</html>",
					VisualEditorTree: MapOfAny{"type": "root"},
				},
				Category: string(TemplateCategoryMarketing),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, workspaceID, err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, template)
				assert.Equal(t, tt.request.WorkspaceID, workspaceID)
				assert.Equal(t, tt.request.ID, template.ID)
				assert.Equal(t, tt.request.Name, template.Name)
				assert.Equal(t, tt.request.Channel, template.Channel)
				assert.Equal(t, tt.request.Email, template.Email)
				assert.Equal(t, tt.request.Category, template.Category)
			}
		})
	}
}

func TestDeleteTemplateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *DeleteTemplateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &DeleteTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "template123",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: &DeleteTemplateRequest{
				WorkspaceID: "",
				ID:          "template123",
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			request: &DeleteTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "",
			},
			wantErr: true,
		},
		{
			name: "ID too long",
			request: &DeleteTemplateRequest{
				WorkspaceID: "workspace123",
				ID:          "this_id_is_way_too_long_for_the_validation_to_pass_properly",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaceID, id, err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.request.WorkspaceID, workspaceID)
				assert.Equal(t, tt.request.ID, id)
			}
		})
	}
}

func TestErrTemplateNotFound_Error(t *testing.T) {
	err := &ErrTemplateNotFound{Message: "template not found"}
	assert.Equal(t, "template not found", err.Error())
}
