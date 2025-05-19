package domain

import (
	"encoding/json"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/pkg/mjml"
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
				SenderID:         "test123",
				Subject:          "Test Subject",
				CompiledPreview:  "<html>Test content</html>",
				VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
			name: "invalid template with version 0",
			template: func() *Template {
				t := createValidTemplate()
				t.Version = 0
				return t
			}(),
			wantErr: true,
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
			name: "valid reference with version 0",
			ref: &TemplateReference{
				ID:      "test123",
				Version: 0,
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
			name: "invalid reference - negative version",
			ref: &TemplateReference{
				ID:      "test123",
				Version: -1,
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
		testData MapOfAny
		wantErr  bool
	}{
		{
			name: "valid template",
			template: &EmailTemplate{
				SenderID:         "test123",
				Subject:          "Test Subject",
				CompiledPreview:  "<html>Test content</html>",
				VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
			},
			testData: nil,
			wantErr:  false,
		},
		{
			name: "invalid email template - missing sender_id",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
				}
				e.SenderID = ""
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
		{
			name: "invalid email template - missing subject",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
				}
				e.Subject = ""
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
		{
			name: "invalid email template - missing compiled_preview but valid tree",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
				}
				return e
			}(),
			testData: nil,
			wantErr:  false,
		},
		{
			name: "invalid email template - missing compiled_preview and missing root data",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: nil},
				}
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
		{
			name: "invalid email template - missing compiled_preview and invalid root data type",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: "not a map"},
				}
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
		{
			name: "invalid email template - missing compiled_preview and missing styles in root data",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"other": "stuff"}},
				}
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
		{
			name: "invalid email template - invalid visual_editor_tree kind",
			template: func() *EmailTemplate {
				e := &EmailTemplate{
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "not-root"},
				}
				return e
			}(),
			testData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate(tt.testData)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.name == "invalid email template - missing compiled_preview but valid tree" {
					assert.NotEmpty(t, tt.template.CompiledPreview, "CompiledPreview should be populated after validation")
				}
			}
		})
	}
}

func TestEmailTemplate_Scan_Value(t *testing.T) {
	email := &EmailTemplate{
		SenderID:         "test123",
		Subject:          "Test Subject",
		CompiledPreview:  "<html>Test content</html>",
		VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
	assert.Equal(t, email.SenderID, newEmail.SenderID)
	assert.Equal(t, email.Subject, newEmail.Subject)
	assert.Equal(t, email.CompiledPreview, newEmail.CompiledPreview)

	// Test Scan() method with string
	err = newEmail.Scan(string(bytes))
	assert.NoError(t, err)
	assert.Equal(t, email.SenderID, newEmail.SenderID)
	assert.Equal(t, email.Subject, newEmail.Subject)
	assert.Equal(t, email.CompiledPreview, newEmail.CompiledPreview)

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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
				assert.Equal(t, int64(1), template.Version)
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
					SenderID:         "test123",
					Subject:          "Test Subject",
					CompiledPreview:  "<html>Test content</html>",
					VisualEditorTree: mjml.EmailBlock{Kind: "root", Data: map[string]interface{}{"styles": map[string]interface{}{}}},
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
				Email:       &EmailTemplate{},
				Category:    string(TemplateCategoryMarketing),
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

func TestBuildTemplateData(t *testing.T) {
	t.Run("with complete data", func(t *testing.T) {
		// Setup test data
		workspaceID := "ws-123"
		apiEndpoint := "https://api.example.com"
		messageID := "msg-456"

		firstName := &NullableString{String: "John", IsNull: false}
		lastName := &NullableString{String: "Doe", IsNull: false}

		contact := &Contact{
			Email:     "test@example.com",
			FirstName: firstName,
			LastName:  lastName,
			// Don't use Properties field as it doesn't exist in Contact struct
		}

		contactWithList := ContactWithList{
			Contact:  contact,
			ListID:   "list-789",
			ListName: "Newsletter",
		}

		broadcast := &Broadcast{
			ID:   "broadcast-001",
			Name: "Test Broadcast",
		}

		trackingSettings := mjml.TrackingSettings{
			Endpoint:    apiEndpoint,
			UTMSource:   "newsletter",
			UTMMedium:   "email",
			UTMCampaign: "welcome",
			UTMTerm:     "new-users",
			UTMContent:  "button-1",
		}

		// Call the function
		data, err := BuildTemplateData(workspaceID, "", contactWithList, messageID, trackingSettings, broadcast)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check contact data
		contactData, ok := data["contact"].(MapOfAny)
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", contactData["email"])
		assert.Equal(t, "John", contactData["first_name"])
		assert.Equal(t, "Doe", contactData["last_name"])

		// Check broadcast data
		broadcastData, ok := data["broadcast"].(MapOfAny)
		assert.True(t, ok)
		assert.Equal(t, "broadcast-001", broadcastData["id"])
		assert.Equal(t, "Test Broadcast", broadcastData["name"])

		// Check UTM parameters
		assert.Equal(t, "newsletter", data["utm_source"])
		assert.Equal(t, "email", data["utm_medium"])
		assert.Equal(t, "welcome", data["utm_campaign"])
		assert.Equal(t, "new-users", data["utm_term"])
		assert.Equal(t, "button-1", data["utm_content"])

		// Check list data
		listData, ok := data["list"].(MapOfAny)
		assert.True(t, ok)
		assert.Equal(t, "list-789", listData["id"])
		assert.Equal(t, "Newsletter", listData["name"])

		// Check unsubscribe URL
		unsubscribeURL, ok := data["unsubscribe_url"].(string)
		assert.True(t, ok)
		assert.Contains(t, unsubscribeURL, "https://api.example.com/unsubscribe")
		assert.Contains(t, unsubscribeURL, "email=test%40example.com")
		assert.Contains(t, unsubscribeURL, "lid=list-789")
		assert.Contains(t, unsubscribeURL, "lname=Newsletter")
		assert.Contains(t, unsubscribeURL, "wid=ws-123")
		assert.Contains(t, unsubscribeURL, "mid=msg-456")

		// Check tracking data
		assert.Equal(t, messageID, data["message_id"])

		// Check tracking pixel URL
		trackingPixelURL, ok := data["tracking_opens_url"].(string)
		assert.True(t, ok)
		assert.Contains(t, trackingPixelURL, "https://api.example.com/opens")
		assert.Contains(t, trackingPixelURL, "mid=msg-456")
		assert.Contains(t, trackingPixelURL, "wid=ws-123")
	})

	t.Run("with minimal data", func(t *testing.T) {
		// Setup minimal test data
		workspaceID := "ws-123"
		messageID := "msg-456"

		contactWithList := ContactWithList{
			Contact: nil,
		}
		trackingSettings := mjml.TrackingSettings{
			Endpoint:    "https://api.example.com",
			UTMSource:   "newsletter",
			UTMMedium:   "email",
			UTMCampaign: "welcome",
			UTMTerm:     "new-users",
			UTMContent:  "button-1",
		}
		// Call the function
		data, err := BuildTemplateData(workspaceID, "", contactWithList, messageID, trackingSettings, nil)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check contact data should be empty
		contactData, ok := data["contact"].(MapOfAny)
		assert.True(t, ok)
		assert.Empty(t, contactData)

		// Check message ID still exists
		assert.Equal(t, messageID, data["message_id"])

		// Check tracking opens URL still exists even without API endpoint
		trackingPixelURL, ok := data["tracking_opens_url"].(string)
		assert.True(t, ok)
		assert.Contains(t, trackingPixelURL, "/opens")
		assert.Contains(t, trackingPixelURL, "mid=msg-456")
		assert.Contains(t, trackingPixelURL, "wid=ws-123")

		// No unsubscribe URL should be present
		_, exists := data["unsubscribe_url"]
		assert.False(t, exists)
	})

	// We'll skip the third test case since it would require mocking
}

// TestGenerateEmailRedirectionEndpoint tests the generation of the URL for tracking email redirections
func TestGenerateEmailRedirectionEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		messageID   string
		apiEndpoint string
		expected    string
	}{
		{
			name:        "with all parameters",
			workspaceID: "ws-123",
			messageID:   "msg-456",
			apiEndpoint: "https://api.example.com",
			expected:    "https://api.example.com/visit?mid=msg-456&wid=ws-123",
		},
		{
			name:        "with empty api endpoint",
			workspaceID: "ws-123",
			messageID:   "msg-456",
			apiEndpoint: "",
			expected:    "/visit?mid=msg-456&wid=ws-123",
		},
		{
			name:        "with special characters that need encoding",
			workspaceID: "ws/123&test=1",
			messageID:   "msg=456?test=1",
			apiEndpoint: "https://api.example.com",
			expected:    "https://api.example.com/visit?mid=msg%3D456%3Ftest%3D1&wid=ws%2F123%26test%3D1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := GenerateEmailRedirectionEndpoint(tt.workspaceID, tt.messageID, tt.apiEndpoint)
			assert.Equal(t, tt.expected, url)
		})
	}
}
