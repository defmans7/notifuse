package mjml

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	mjmlgo "github.com/Boostport/mjml-go"
)

// MapOfAny represents a map of string to any value, used for template data
type MapOfAny map[string]any

// CompileTemplateRequest represents the request for compiling a template
type CompileTemplateRequest struct {
	WorkspaceID      string     `json:"workspace_id"`
	MessageID        string     `json:"message_id"`
	VisualEditorTree EmailBlock `json:"visual_editor_tree"`
	TemplateData     MapOfAny   `json:"test_data,omitempty"`
	TrackingEnabled  bool       `json:"tracking_enabled,omitempty"`
	UTMSource        *string    `json:"utm_source,omitempty"`
	UTMMedium        *string    `json:"utm_medium,omitempty"`
	UTMCampaign      *string    `json:"utm_campaign,omitempty"`
	UTMContent       *string    `json:"utm_content,omitempty"`
	UTMTerm          *string    `json:"utm_term,omitempty"`
}

// Validate ensures that the compile template request has all required fields
func (r *CompileTemplateRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid compile template request: workspace_id is required")
	}
	if r.MessageID == "" {
		return fmt.Errorf("invalid compile template request: message_id is required")
	}
	// Basic validation for the tree root kind
	if r.VisualEditorTree.Kind != "root" {
		return fmt.Errorf("invalid compile template request: visual_editor_tree must have kind 'root'")
	}
	if r.VisualEditorTree.Data == nil {
		return fmt.Errorf("invalid compile template request: visual_editor_tree root block must have data (styles)")
	}

	return nil
}

// CompileTemplateResponse represents the response from compiling a template
type CompileTemplateResponse struct {
	Success bool          `json:"success"`
	MJML    *string       `json:"mjml,omitempty"`  // Pointer, omit if nil
	HTML    *string       `json:"html,omitempty"`  // Pointer, omit if nil
	Error   *mjmlgo.Error `json:"error,omitempty"` // Pointer, omit if nil
}

// GenerateEmailRedirectionEndpoint generates the email redirection endpoint URL
func GenerateEmailRedirectionEndpoint(workspaceID string, messageID string, apiEndpoint string) string {
	// URL encode the parameters to handle special characters
	encodedMID := url.QueryEscape(messageID)
	encodedWID := url.QueryEscape(workspaceID)
	return fmt.Sprintf("%s/visit?mid=%s&wid=%s",
		apiEndpoint, encodedMID, encodedWID)
}

// CompileTemplate compiles a visual editor tree to MJML and HTML
func CompileTemplate(apiEndpoint string, payload CompileTemplateRequest) (*CompileTemplateResponse, error) {
	// Extract root styles from the tree data
	rootDataMap, ok := payload.VisualEditorTree.Data.(map[string]interface{})
	if !ok {
		// Return standard Go error for non-compilation issues
		return nil, fmt.Errorf("invalid root block data format")
	}
	rootStyles, _ := rootDataMap["styles"].(map[string]interface{})
	if rootStyles == nil {
		// Return standard Go error for non-compilation issues
		return nil, fmt.Errorf("root block styles are required for compilation")
	}

	// Prepare template data JSON string
	var templateDataStr string
	if payload.TemplateData != nil && len(payload.TemplateData) > 0 {
		jsonDataBytes, err := json.Marshal(payload.TemplateData)
		if err != nil {
			// Return standard Go error for non-compilation issues
			return nil, fmt.Errorf("failed to marshal test_data: %w", err)
		}
		templateDataStr = string(jsonDataBytes)
	}

	trackingSettings := TrackingSettings{
		EnableTracking: payload.TrackingEnabled,
		Endpoint:       GenerateEmailRedirectionEndpoint(payload.WorkspaceID, payload.MessageID, apiEndpoint),
	}

	if payload.UTMSource != nil {
		trackingSettings.UTMSource = *payload.UTMSource
	}
	if payload.UTMMedium != nil {
		trackingSettings.UTMMedium = *payload.UTMMedium
	}
	if payload.UTMCampaign != nil {
		trackingSettings.UTMCampaign = *payload.UTMCampaign
	}
	if payload.UTMContent != nil {
		trackingSettings.UTMContent = *payload.UTMContent
	}
	if payload.UTMTerm != nil {
		trackingSettings.UTMTerm = *payload.UTMTerm
	}

	// Compile tree to MJML using our pkg/mjml function
	mjmlResult, err := TreeToMjml(rootStyles, payload.VisualEditorTree, templateDataStr, trackingSettings, 0, nil)
	if err != nil {
		return &CompileTemplateResponse{
			Success: false,
			MJML:    nil,
			HTML:    nil,
			Error: &mjmlgo.Error{
				Message: err.Error(),
			},
		}, nil
	}

	ctx := context.Background()

	// Compile MJML to HTML using mjml-go library
	htmlResult, err := mjmlgo.ToHTML(ctx, mjmlResult)
	if err != nil {
		// Return the response struct with Success=false and the Error details
		return &CompileTemplateResponse{
			Success: false,
			MJML:    &mjmlResult, // Include original MJML for context if desired
			HTML:    nil,
			Error: &mjmlgo.Error{
				Message: err.Error(),
			},
		}, nil
	}

	// Return successful response
	return &CompileTemplateResponse{
		Success: true,
		MJML:    &mjmlResult,
		HTML:    &htmlResult,
		Error:   nil,
	}, nil
}
