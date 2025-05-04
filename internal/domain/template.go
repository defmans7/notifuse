package domain

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/pkg/mjml" // Import the mjml package
	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_template_service.go -package mocks github.com/Notifuse/notifuse/internal/domain TemplateService
//go:generate mockgen -destination mocks/mock_template_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain TemplateRepository

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
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}

func (t *Template) Validate() error {
	// First validate the template itself
	if t.ID == "" {
		return fmt.Errorf("invalid template: id is required")
	}
	if len(t.ID) > 32 {
		return fmt.Errorf("invalid template: id length must be between 1 and 32")
	}

	if t.Name == "" {
		return fmt.Errorf("invalid template: name is required")
	}
	if len(t.Name) > 32 {
		return fmt.Errorf("invalid template: name length must be between 1 and 32")
	}

	if t.Version <= 0 {
		return fmt.Errorf("invalid template: version must be positive")
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

	if t.TestData == nil {
		t.TestData = MapOfAny{}
	}

	if err := t.Email.Validate(t.TestData); err != nil {
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
	if len(t.ID) > 32 {
		return fmt.Errorf("invalid template reference: id length must be between 1 and 32")
	}

	if t.Version < 0 {
		return fmt.Errorf("invalid template reference: version must be zero or positive")
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
	FromAddress      string          `json:"from_address"`
	FromName         string          `json:"from_name"`
	ReplyTo          *string         `json:"reply_to,omitempty"`
	Subject          string          `json:"subject"`
	SubjectPreview   *string         `json:"subject_preview,omitempty"`
	CompiledPreview  string          `json:"compiled_preview"` // compiled html
	VisualEditorTree mjml.EmailBlock `json:"visual_editor_tree"`
	Text             *string         `json:"text,omitempty"`
}

func (e *EmailTemplate) Validate(testData MapOfAny) error {
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
	if len(e.FromName) > 32 {
		return fmt.Errorf("invalid email template: from_name length must be between 1 and 32")
	}
	if e.Subject == "" {
		return fmt.Errorf("invalid email template: subject is required")
	}
	if len(e.Subject) > 32 {
		return fmt.Errorf("invalid email template: subject length must be between 1 and 32")
	}
	if e.VisualEditorTree.Kind != "root" {
		return fmt.Errorf("invalid email template: visual_editor_tree must have kind 'root'")
	}
	if e.VisualEditorTree.Data == nil {
		return fmt.Errorf("invalid email template: visual_editor_tree root block must have data (styles)")
	}
	if e.CompiledPreview == "" {
		// Extract root styles from the tree data
		rootDataMap, ok := e.VisualEditorTree.Data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid email template: root block data is not a map")
		}
		rootStyles, _ := rootDataMap["styles"].(map[string]interface{})
		if rootStyles == nil {
			return fmt.Errorf("invalid email template: root block styles are missing")
		}

		// Prepare template data JSON string
		var templateDataStr string
		if testData != nil && len(testData) > 0 {
			jsonDataBytes, err := json.Marshal(testData)
			if err != nil {
				return fmt.Errorf("failed to marshal test_data: %w", err)
			}
			templateDataStr = string(jsonDataBytes)
		}

		// Compile tree to MJML using our pkg/mjml function
		mjmlResult, err := mjml.TreeToMjml(rootStyles, e.VisualEditorTree, templateDataStr, map[string]string{}, 0, nil)
		if err != nil {
			return fmt.Errorf("failed to generate MJML from tree: %w", err)
		}
		e.CompiledPreview = mjmlResult
	}

	// Validate optional fields
	if e.ReplyTo != nil && !govalidator.IsEmail(*e.ReplyTo) {
		return fmt.Errorf("invalid email template: reply_to is not a valid email")
	}
	if e.SubjectPreview != nil && len(*e.SubjectPreview) > 32 {
		return fmt.Errorf("invalid email template: subject_preview length must be between 1 and 32")
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

//go:generate mockgen -destination mocks/mock_template_service.go -package mocks github.com/Notifuse/notifuse/internal/domain TemplateService
//go:generate mockgen -destination mocks/mock_template_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain TemplateRepository

// Request/Response types
type CreateTemplateRequest struct {
	WorkspaceID     string         `json:"workspace_id"`
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Channel         string         `json:"channel"`
	Email           *EmailTemplate `json:"email"`
	Category        string         `json:"category"`
	TemplateMacroID *string        `json:"template_macro_id,omitempty"`
	UTMSource       *string        `json:"utm_source,omitempty"`
	UTMMedium       *string        `json:"utm_medium,omitempty"`
	UTMCampaign     *string        `json:"utm_campaign,omitempty"`
	TestData        MapOfAny       `json:"test_data,omitempty"`
	Settings        MapOfAny       `json:"settings,omitempty"`
}

func (r *CreateTemplateRequest) Validate() (template *Template, workspaceID string, err error) {
	if r.WorkspaceID == "" {
		return nil, "", fmt.Errorf("invalid create template request: workspace_id is required")
	}
	if r.ID == "" {
		return nil, "", fmt.Errorf("invalid create template request: id is required")
	}
	if len(r.ID) > 32 {
		return nil, "", fmt.Errorf("invalid create template request: id length must be between 1 and 32")
	}

	if r.Name == "" {
		return nil, "", fmt.Errorf("invalid create template request: name is required")
	}
	if len(r.Name) > 32 {
		return nil, "", fmt.Errorf("invalid create template request: name length must be between 1 and 32")
	}

	if r.Channel == "" {
		return nil, "", fmt.Errorf("invalid create template request: channel is required")
	}
	if len(r.Channel) > 20 {
		return nil, "", fmt.Errorf("invalid create template request: channel length must be between 1 and 20")
	}

	if r.Category == "" {
		return nil, "", fmt.Errorf("invalid create template request: category is required")
	}
	if len(r.Category) > 20 {
		return nil, "", fmt.Errorf("invalid create template request: category length must be between 1 and 20")
	}

	if r.Email == nil {
		return nil, "", fmt.Errorf("invalid create template request: email is required")
	}

	if err := r.Email.Validate(r.TestData); err != nil {
		return nil, "", fmt.Errorf("invalid create template request: %w", err)
	}

	return &Template{
		ID:              r.ID,
		Name:            r.Name,
		Version:         1, // Start with version 1 for new templates
		Channel:         r.Channel,
		Email:           r.Email,
		Category:        r.Category,
		TemplateMacroID: r.TemplateMacroID,
		UTMSource:       r.UTMSource,
		UTMMedium:       r.UTMMedium,
		UTMCampaign:     r.UTMCampaign,
		TestData:        r.TestData,
		Settings:        r.Settings,
	}, r.WorkspaceID, nil
}

type GetTemplatesRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Category    string `json:"category,omitempty"`
}

func (r *GetTemplatesRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.Category = queryParams.Get("category")

	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid get templates request: workspace_id is required")
	}
	if len(r.WorkspaceID) > 20 {
		return fmt.Errorf("invalid get templates request: workspace_id length must be between 1 and 20")
	}

	if r.Category != "" {
		if len(r.Category) > 20 {
			return fmt.Errorf("invalid get templates request: category length must be between 1 and 20")
		}
	}

	return nil
}

type GetTemplateRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
	Version     int64  `json:"version,omitempty"`
}

func (r *GetTemplateRequest) FromURLParams(queryParams url.Values) (err error) {
	r.WorkspaceID = queryParams.Get("workspace_id")
	r.ID = queryParams.Get("id")
	versionStr := queryParams.Get("version")

	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid get template request: workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("invalid get template request: id is required")
	}
	if len(r.ID) > 32 {
		return fmt.Errorf("invalid get template request: id length must be between 1 and 32")
	}

	if versionStr != "" {
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid get template request: version must be a valid integer")
		}
		r.Version = version
	}

	return nil
}

type UpdateTemplateRequest struct {
	WorkspaceID     string         `json:"workspace_id"`
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Channel         string         `json:"channel"`
	Email           *EmailTemplate `json:"email"`
	Category        string         `json:"category"`
	TemplateMacroID *string        `json:"template_macro_id,omitempty"`
	UTMSource       *string        `json:"utm_source,omitempty"`
	UTMMedium       *string        `json:"utm_medium,omitempty"`
	UTMCampaign     *string        `json:"utm_campaign,omitempty"`
	TestData        MapOfAny       `json:"test_data,omitempty"`
	Settings        MapOfAny       `json:"settings,omitempty"`
}

func (r *UpdateTemplateRequest) Validate() (template *Template, workspaceID string, err error) {
	if r.WorkspaceID == "" {
		return nil, "", fmt.Errorf("invalid update template request: workspace_id is required")
	}
	if r.ID == "" {
		return nil, "", fmt.Errorf("invalid update template request: id is required")
	}
	if len(r.ID) > 32 {
		return nil, "", fmt.Errorf("invalid update template request: id length must be between 1 and 32")
	}

	if r.Name == "" {
		return nil, "", fmt.Errorf("invalid update template request: name is required")
	}
	if len(r.Name) > 32 {
		return nil, "", fmt.Errorf("invalid update template request: name length must be between 1 and 32")
	}

	if r.Channel == "" {
		return nil, "", fmt.Errorf("invalid update template request: channel is required")
	}
	if len(r.Channel) > 20 {
		return nil, "", fmt.Errorf("invalid update template request: channel length must be between 1 and 20")
	}

	if r.Category == "" {
		return nil, "", fmt.Errorf("invalid update template request: category is required")
	}
	if len(r.Category) > 20 {
		return nil, "", fmt.Errorf("invalid update template request: category length must be between 1 and 20")
	}

	if r.Email == nil {
		return nil, "", fmt.Errorf("invalid update template request: email is required")
	}

	if err := r.Email.Validate(r.TestData); err != nil {
		return nil, "", fmt.Errorf("invalid update template request: %w", err)
	}

	return &Template{
		ID:              r.ID,
		Name:            r.Name,
		Channel:         r.Channel,
		Email:           r.Email,
		Category:        r.Category,
		TemplateMacroID: r.TemplateMacroID,
		UTMSource:       r.UTMSource,
		UTMMedium:       r.UTMMedium,
		UTMCampaign:     r.UTMCampaign,
		TestData:        r.TestData,
		Settings:        r.Settings,
	}, r.WorkspaceID, nil
}

type DeleteTemplateRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

func (r *DeleteTemplateRequest) Validate() (workspaceID string, id string, err error) {
	if r.WorkspaceID == "" {
		return "", "", fmt.Errorf("invalid delete template request: workspace_id is required")
	}

	if r.ID == "" {
		return "", "", fmt.Errorf("invalid delete template request: id is required")
	}
	if len(r.ID) > 32 {
		return "", "", fmt.Errorf("invalid delete template request: id length must be between 1 and 32")
	}

	return r.WorkspaceID, r.ID, nil
}

// --- Compile Request/Response ---

type CompileTemplateRequest struct {
	WorkspaceID      string          `json:"workspace_id"`
	VisualEditorTree mjml.EmailBlock `json:"visual_editor_tree"` // Use the struct from pkg/mjml
	TestData         MapOfAny        `json:"test_data,omitempty"`
}

func (r *CompileTemplateRequest) Validate() (workspaceID string, tree mjml.EmailBlock, testData MapOfAny, err error) {
	if r.WorkspaceID == "" {
		return "", mjml.EmailBlock{}, nil, fmt.Errorf("invalid compile template request: workspace_id is required")
	}
	// Basic validation for the tree root kind
	if r.VisualEditorTree.Kind != "root" {
		return "", mjml.EmailBlock{}, nil, fmt.Errorf("invalid compile template request: visual_editor_tree must have kind 'root'")
	}
	if r.VisualEditorTree.Data == nil {
		// Add default root styles if missing, or return error? Let's return error for now.
		// Alternatively, could initialize with default rootStyles here.
		return "", mjml.EmailBlock{}, nil, fmt.Errorf("invalid compile template request: visual_editor_tree root block must have data (styles)")
	}

	return r.WorkspaceID, r.VisualEditorTree, r.TestData, nil
}

type CompileTemplateResponse struct {
	Success bool          `json:"success"`
	MJML    *string       `json:"mjml,omitempty"`  // Pointer, omit if nil
	HTML    *string       `json:"html,omitempty"`  // Pointer, omit if nil
	Error   *mjmlgo.Error `json:"error,omitempty"` // Pointer, omit if nil
}

// TestTemplateRequest represents a request to test a template
type TestTemplateRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	TemplateID     string `json:"template_id"`
	IntegrationID  string `json:"integration_id"`
	RecipientEmail string `json:"recipient_email"`
}

func (r *TestTemplateRequest) Validate() (string, string, string, string, error) {
	if r.WorkspaceID == "" {
		return "", "", "", "", fmt.Errorf("workspace_id is required")
	}
	if r.TemplateID == "" {
		return "", "", "", "", fmt.Errorf("template_id is required")
	}
	if r.IntegrationID == "" {
		return "", "", "", "", fmt.Errorf("integration_id is required")
	}
	if r.RecipientEmail == "" {
		return "", "", "", "", fmt.Errorf("recipient_email is required")
	}
	if !govalidator.IsEmail(r.RecipientEmail) {
		return "", "", "", "", fmt.Errorf("invalid recipient_email format")
	}

	return r.WorkspaceID, r.TemplateID, r.IntegrationID, r.RecipientEmail, nil
}

// TestTemplateResponse represents the response from testing a template
type TestTemplateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// TemplateService provides operations for managing templates
type TemplateService interface {
	// CreateTemplate creates a new template
	CreateTemplate(ctx context.Context, workspaceID string, template *Template) error

	// GetTemplateByID retrieves a template by ID and optional version
	GetTemplateByID(ctx context.Context, workspaceID string, id string, version int64) (*Template, error)

	// GetTemplates retrieves all templates
	GetTemplates(ctx context.Context, workspaceID string, category string) ([]*Template, error)

	// UpdateTemplate updates an existing template
	UpdateTemplate(ctx context.Context, workspaceID string, template *Template) error

	// DeleteTemplate deletes a template by ID
	DeleteTemplate(ctx context.Context, workspaceID string, id string) error

	// CompileTemplate compiles a visual editor tree to MJML and HTML
	CompileTemplate(ctx context.Context, workspaceID string, tree mjml.EmailBlock, testData MapOfAny) (*CompileTemplateResponse, error) // Use mjml.EmailBlock
}

// TemplateRepository provides database operations for templates
type TemplateRepository interface {
	// CreateTemplate creates a new template in the database
	CreateTemplate(ctx context.Context, workspaceID string, template *Template) error

	// GetTemplateByID retrieves a template by its ID and optional version
	GetTemplateByID(ctx context.Context, workspaceID string, id string, version int64) (*Template, error)

	// GetTemplateLatestVersion retrieves the latest version of a template
	GetTemplateLatestVersion(ctx context.Context, workspaceID string, id string) (int64, error)

	// GetTemplates retrieves all templates
	GetTemplates(ctx context.Context, workspaceID string, category string) ([]*Template, error)

	// UpdateTemplate updates an existing template, creating a new version
	UpdateTemplate(ctx context.Context, workspaceID string, template *Template) error

	// DeleteTemplate deletes a template
	DeleteTemplate(ctx context.Context, workspaceID string, id string) error
}

// ErrTemplateNotFound is returned when a template is not found
type ErrTemplateNotFound struct {
	Message string
}

func (e *ErrTemplateNotFound) Error() string {
	return e.Message
}

// BuildTemplateData creates a template data map with flexible options
func BuildTemplateData(workspaceID string, contactWithList ContactWithList, messageID string, apiEndpoint string, broadcast *Broadcast) (MapOfAny, error) {
	templateData := MapOfAny{}

	if contactWithList.Contact != nil {

		// Use all contact data
		contactData, err := contactWithList.Contact.ToMapOfAny()
		if err != nil {
			return nil, fmt.Errorf("failed to convert contact to template data: %w", err)
		}
		templateData["contact"] = contactData

	} else {
		// Create empty contact object if none provided
		templateData["contact"] = MapOfAny{}
	}

	// Add broadcast data if available
	if broadcast != nil {

		templateData["broadcast"] = MapOfAny{
			"id":   broadcast.ID,
			"name": broadcast.Name,
		}

		// Add UTM parameters if tracking is enabled
		if broadcast.TrackingEnabled && broadcast.UTMParameters != nil {
			templateData["utm_source"] = broadcast.UTMParameters.Source
			templateData["utm_medium"] = broadcast.UTMParameters.Medium
			templateData["utm_campaign"] = broadcast.UTMParameters.Campaign
			templateData["utm_term"] = broadcast.UTMParameters.Term
			templateData["utm_content"] = broadcast.UTMParameters.Content
		}
	}

	// Add list data and unsubscribe link if available
	if contactWithList.ListID != "" && apiEndpoint != "" && workspaceID != "" {

		templateData["list"] = MapOfAny{
			"id":   contactWithList.ListID,
			"name": contactWithList.ListName,
		}

		// Create unsubscribe link
		email := url.QueryEscape(contactWithList.Contact.Email)
		listID := url.QueryEscape(contactWithList.ListID)
		listName := url.QueryEscape(contactWithList.ListName)
		workspaceID := url.QueryEscape(workspaceID)

		// Note: In a real implementation, you would add a signature token for security
		unsubscribeURL := fmt.Sprintf("%s/public/unsubscribe?email=%s&list=%s&list_name=%s&workspace=%s&message=%s",
			apiEndpoint, email, listID, listName, workspaceID, messageID)

		templateData["unsubscribe_url"] = unsubscribeURL
	}

	// Add tracking data
	templateData["message_id"] = messageID

	// Add tracking pixel if API endpoint is provided

	// Format: {apiEndpoint}/api/pixel?id={messageID}&t=o&w={workspaceID}
	messageID = url.QueryEscape(messageID)
	workspaceID = url.QueryEscape(workspaceID)

	// Tracking pixel for opens
	trackingPixelURL := fmt.Sprintf("%s/public/opens?id=%s&t=o&w=%s",
		apiEndpoint, messageID, workspaceID)

	templateData["tracking_opens_url"] = trackingPixelURL

	// Base URL for click tracking
	// Usage in templates: {{tracking_base}}&url={{encoded_destination_url}}
	trackingBaseURL := fmt.Sprintf("%s/public/click?id=%s&w=%s",
		apiEndpoint, messageID, workspaceID)

	templateData["tracking_click_url"] = trackingBaseURL

	return templateData, nil
}
