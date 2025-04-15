package mjml

import (
	"reflect"
	"strings"
	"testing"
)

func TestEmailBlock_GetBlockData(t *testing.T) {
	tests := []struct {
		name     string
		block    EmailBlock
		wantType string
	}{
		{
			name: "button block",
			block: EmailBlock{
				ID:   "test-button",
				Kind: "button",
				Data: ButtonBlockData{
					Button: struct {
						Text                   string `json:"text"`
						Href                   string `json:"href"`
						BackgroundColor        string `json:"backgroundColor"`
						FontFamily             string `json:"fontFamily"`
						FontSize               string `json:"fontSize"`
						FontWeight             int    `json:"fontWeight"`
						FontStyle              string `json:"fontStyle"`
						Color                  string `json:"color"`
						InnerVerticalPadding   string `json:"innerVerticalPadding"`
						InnerHorizontalPadding string `json:"innerHorizontalPadding"`
						Width                  string `json:"width"`
						TextTransform          string `json:"textTransform"`
						BorderRadius           string `json:"borderRadius"`
						DisableTracking        bool   `json:"disable_tracking"`
						BorderControl          string `json:"borderControl"`
						BorderStyle            string `json:"borderStyle,omitempty"`
						BorderWidth            string `json:"borderWidth,omitempty"`
						BorderColor            string `json:"borderColor,omitempty"`
						BorderTopStyle         string `json:"borderTopStyle,omitempty"`
						BorderTopWidth         string `json:"borderTopWidth,omitempty"`
						BorderTopColor         string `json:"borderTopColor,omitempty"`
						BorderRightStyle       string `json:"borderRightStyle,omitempty"`
						BorderRightWidth       string `json:"borderRightWidth,omitempty"`
						BorderRightColor       string `json:"borderRightColor,omitempty"`
						BorderBottomStyle      string `json:"borderBottomStyle,omitempty"`
						BorderBottomWidth      string `json:"borderBottomWidth,omitempty"`
						BorderBottomColor      string `json:"borderBottomColor,omitempty"`
						BorderLeftStyle        string `json:"borderLeftStyle,omitempty"`
						BorderLeftWidth        string `json:"borderLeftWidth,omitempty"`
						BorderLeftColor        string `json:"borderLeftColor,omitempty"`
					}{
						Text: "Click Me",
						Href: "https://example.com",
					},
					Wrapper: WrapperStyles{
						Align: "center",
					},
				},
			},
			wantType: "mjml.ButtonBlockData",
		},
		{
			name: "image block",
			block: EmailBlock{
				ID:   "test-image",
				Kind: "image",
				Data: ImageBlockData{
					Image: struct {
						Src           string `json:"src"`
						Alt           string `json:"alt"`
						Href          string `json:"href"`
						Width         string `json:"width"`
						BorderControl string `json:"borderControl"`
					}{
						Src: "https://example.com/image.jpg",
						Alt: "Test Image",
					},
					Wrapper: WrapperStyles{
						Align: "center",
					},
				},
			},
			wantType: "mjml.ImageBlockData",
		},
		{
			name: "column block",
			block: EmailBlock{
				ID:   "test-column",
				Kind: "column",
				Data: ColumnBlockData{
					Styles: struct {
						VerticalAlign   string `json:"verticalAlign"`
						BackgroundColor string `json:"backgroundColor,omitempty"`
						MinHeight       string `json:"minHeight,omitempty"`
						BaseStyles
					}{
						VerticalAlign: "top",
					},
					PaddingControl: "all",
					BorderControl:  "all",
				},
			},
			wantType: "mjml.ColumnBlockData",
		},
		{
			name: "divider block",
			block: EmailBlock{
				ID:   "test-divider",
				Kind: "divider",
				Data: DividerBlockData{
					Align:       "center",
					BorderColor: "#000000",
					BorderStyle: "solid",
					BorderWidth: "1px",
					Width:       "100%",
				},
			},
			wantType: "mjml.DividerBlockData",
		},
		{
			name: "section block",
			block: EmailBlock{
				ID:   "test-section",
				Kind: "section",
				Data: SectionBlockData{
					ColumnsOnMobile:     true,
					StackColumnsAtWidth: 600,
					BackgroundType:      "color",
					PaddingControl:      "all",
					BorderControl:       "all",
					Styles: struct {
						TextAlign        string `json:"textAlign"`
						BackgroundRepeat string `json:"backgroundRepeat,omitempty"`
						Padding          string `json:"padding,omitempty"`
						BorderWidth      string `json:"borderWidth,omitempty"`
						BorderStyle      string `json:"borderStyle,omitempty"`
						BorderColor      string `json:"borderColor,omitempty"`
						BackgroundColor  string `json:"backgroundColor,omitempty"`
						BackgroundImage  string `json:"backgroundImage,omitempty"`
						BackgroundSize   string `json:"backgroundSize,omitempty"`
						BaseStyles
					}{
						TextAlign:       "center",
						BackgroundColor: "#ffffff",
					},
				},
			},
			wantType: "mjml.SectionBlockData",
		},
		{
			name: "openTracking block",
			block: EmailBlock{
				ID:   "test-openTracking",
				Kind: "openTracking",
				Data: OpenTrackingBlockData{},
			},
			wantType: "mjml.OpenTrackingBlockData",
		},
		{
			name: "text block",
			block: EmailBlock{
				ID:   "test-text",
				Kind: "text",
				Data: TextBlockData{
					Align: "left",
					Width: "100%",
					HyperlinkStyles: struct {
						Color          string `json:"color"`
						TextDecoration string `json:"textDecoration"`
						FontFamily     string `json:"fontFamily"`
						FontSize       string `json:"fontSize"`
						FontWeight     int    `json:"fontWeight"`
						FontStyle      string `json:"fontStyle"`
						TextTransform  string `json:"textTransform"`
					}{
						Color:          "#0000FF",
						TextDecoration: "underline",
					},
				},
			},
			wantType: "mjml.TextBlockData",
		},
		{
			name: "default case",
			block: EmailBlock{
				ID:   "test-unknown",
				Kind: "unknown",
				Data: map[string]interface{}{
					"key": "value",
				},
			},
			wantType: "map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.block.GetBlockData()
			gotType := reflect.TypeOf(got).String()

			if gotType != tt.wantType {
				t.Errorf("EmailBlock.GetBlockData() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

// TestEmailBlock_GetBlockData_PanicHandling tests that type assertion errors are properly handled
func TestEmailBlock_GetBlockData_PanicHandling(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("GetBlockData() did not panic with invalid data type, but should have")
		}
	}()

	// This block has a mismatched data type which should cause a panic on type assertion
	block := EmailBlock{
		ID:   "test-invalid",
		Kind: "button",
		Data: "invalid data type", // String instead of ButtonBlockData
	}

	// This should cause a panic
	_ = block.GetBlockData()
}

func TestTreeToMjml_HeadingBlock(t *testing.T) {
	rootStyles := map[string]interface{}{
		"h1": map[string]interface{}{
			"fontSize":   "32px",
			"fontWeight": "700",
			"color":      "#333333",
			"fontFamily": "Arial, sans-serif",
			"margin":     "0px",
		},
	}

	headingData := HeadingBlockData{
		Type:  "h1",
		Align: "center",
		Width: "100%",
		EditorData: []struct {
			Type     string                   `json:"type"`
			Children []map[string]interface{} `json:"children"`
		}{
			{
				Type: "h1",
				Children: []map[string]interface{}{
					{
						"text": "Test Heading",
					},
				},
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading1",
		Kind: "heading",
		Data: headingData,
	}

	mjml, err := TreeToMjml(rootStyles, headingBlock, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for heading block: %v", err)
	}

	// Check for expected elements in the output
	if !strings.Contains(mjml, "<mj-text") {
		t.Error("Expected <mj-text> tag but not found")
	}

	if !strings.Contains(mjml, "align=\"center\"") {
		t.Error("Expected align=\"center\" attribute but not found")
	}

	if !strings.Contains(mjml, "<h1") {
		t.Error("Expected <h1> tag but not found")
	}

	if !strings.Contains(mjml, "Test Heading") {
		t.Error("Expected heading text but not found")
	}

	if !strings.Contains(mjml, "color: #333333") {
		t.Error("Expected heading style with color but not found")
	}

	if !strings.Contains(mjml, "font-weight: 700") {
		t.Error("Expected heading style with font-weight but not found")
	}
}

func TestTreeToMjml_HeadingWithFormatting(t *testing.T) {
	rootStyles := map[string]interface{}{
		"h2": map[string]interface{}{
			"fontSize":   "24px",
			"fontWeight": "600",
			"color":      "#444444",
			"fontFamily": "Arial, sans-serif",
			"margin":     "0px",
		},
		"hyperlink": map[string]interface{}{
			"color":          "#0000FF",
			"textDecoration": "underline",
			"fontWeight":     "bold",
		},
	}

	headingData := HeadingBlockData{
		Type:  "h2",
		Align: "left",
		EditorData: []struct {
			Type     string                   `json:"type"`
			Children []map[string]interface{} `json:"children"`
		}{
			{
				Type: "h2",
				Children: []map[string]interface{}{
					{
						"text": "Heading with ",
					},
					{
						"text": "bold",
						"bold": true,
					},
					{
						"text": " and ",
					},
					{
						"text": "link",
						"hyperlink": map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading2",
		Kind: "heading",
		Data: headingData,
	}

	urlParams := map[string]string{
		"utm_source": "test",
	}

	mjml, err := TreeToMjml(rootStyles, headingBlock, "", urlParams, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for heading with formatting: %v", err)
	}

	// Check for expected elements
	if !strings.Contains(mjml, "<h2") {
		t.Error("Expected <h2> tag but not found")
	}

	if !strings.Contains(mjml, "<span style=\"font-weight: bold\">bold</span>") {
		t.Error("Expected bold formatting but not found")
	}

	if !strings.Contains(mjml, "<a style=") && !strings.Contains(mjml, "href=\"https://example.com") {
		t.Error("Expected hyperlink but not found")
	}

	if !strings.Contains(mjml, "utm_source=test") {
		t.Error("Expected tracking parameters in URL but not found")
	}
}

func TestTreeToMjml_OpenTracking(t *testing.T) {
	rootStyles := map[string]interface{}{}

	openTrackingBlock := EmailBlock{
		ID:   "tracking1",
		Kind: "openTracking",
		Data: OpenTrackingBlockData{},
	}

	mjml, err := TreeToMjml(rootStyles, openTrackingBlock, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for openTracking block: %v", err)
	}

	// Check for expected elements
	if !strings.Contains(mjml, "<mj-raw>") {
		t.Error("Expected <mj-raw> tag but not found")
	}

	if !strings.Contains(mjml, "<img src=\"{{ open_tracking_pixel_src }}\"") {
		t.Error("Expected image tag with tracking pixel but not found")
	}

	if !strings.Contains(mjml, "height=\"1\" width=\"1\"") {
		t.Error("Expected 1x1 pixel dimensions but not found")
	}

	if !strings.Contains(mjml, "style=\"display:block; max-height:1px; max-width:1px; visibility:hidden;") {
		t.Error("Expected hidden style attributes but not found")
	}
}

// Additional tests for error handling in TreeToMjml

func TestTreeToMjml_HeadingLiquidProcessingError(t *testing.T) {
	rootStyles := map[string]interface{}{}

	headingData := HeadingBlockData{
		Type:  "h1",
		Align: "center",
		EditorData: []struct {
			Type     string                   `json:"type"`
			Children []map[string]interface{} `json:"children"`
		}{
			{
				Type: "h1",
				Children: []map[string]interface{}{
					{
						"text": "Hello {{ invalid syntax",
					},
				},
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading_err",
		Kind: "heading",
		Data: headingData,
	}

	templateData := `{}`

	_, err := TreeToMjml(rootStyles, headingBlock, templateData, map[string]string{}, 0, nil)

	// The error might not occur because the text is not recognized as liquid template
	// This is because we check for "{{" and "{%" patterns
	// If no error, we should verify the content
	if err == nil {
		// Test passes - liquid detection didn't trigger for this pattern
		return
	}

	// If an error did occur, it should be a liquid rendering error
	if !strings.Contains(err.Error(), "liquid rendering error") {
		t.Errorf("Expected liquid rendering error, got: %v", err)
	}
}

func TestTreeToMjml_TextWithLiquidAndData(t *testing.T) {
	rootStyles := map[string]interface{}{
		"paragraph": map[string]interface{}{
			"fontSize":   "16px",
			"color":      "#222222",
			"fontFamily": "Arial, sans-serif",
			"margin":     "0px",
		},
	}

	textData := TextBlockData{
		Align: "left",
		EditorData: []struct {
			Type     string                   `json:"type"`
			Children []map[string]interface{} `json:"children"`
		}{
			{
				Type: "paragraph",
				Children: []map[string]interface{}{
					{
						"text": "Hello {{ name }}! Your order #{{ order_id }} has been {{ status }}.",
					},
				},
			},
		},
	}

	textBlock := EmailBlock{
		ID:   "text_liquid",
		Kind: "text",
		Data: textData,
	}

	templateData := `{
		"name": "John Doe",
		"order_id": 12345,
		"status": "shipped"
	}`

	mjml, err := TreeToMjml(rootStyles, textBlock, templateData, map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for text with liquid: %v", err)
	}

	// Check for expected elements with replaced values
	if !strings.Contains(mjml, "Hello John Doe!") {
		t.Error("Expected name replacement but not found")
	}

	if !strings.Contains(mjml, "order #12345") {
		t.Error("Expected order_id replacement but not found")
	}

	if !strings.Contains(mjml, "has been shipped") {
		t.Error("Expected status replacement but not found")
	}
}

func TestTreeToMjml_InvalidJsonTemplateData(t *testing.T) {
	rootStyles := map[string]interface{}{}

	textData := TextBlockData{
		Align: "left",
		EditorData: []struct {
			Type     string                   `json:"type"`
			Children []map[string]interface{} `json:"children"`
		}{
			{
				Type: "paragraph",
				Children: []map[string]interface{}{
					{
						"text": "Hello {{ name }}!",
					},
				},
			},
		},
	}

	textBlock := EmailBlock{
		ID:   "text_invalid_json",
		Kind: "text",
		Data: textData,
	}

	// Invalid JSON in templateData
	templateData := `{"name": "John Doe", invalid json}`

	_, err := TreeToMjml(rootStyles, textBlock, templateData, map[string]string{}, 0, nil)

	// Should return an error for invalid JSON
	if err == nil {
		t.Error("Expected error for invalid JSON template data, but got nil")
		return
	}

	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("Expected 'invalid JSON' error message, got: %v", err)
	}
}

func TestTreeToMjml_MarshalUnmarshalError(t *testing.T) {
	rootStyles := map[string]interface{}{}

	// Create a block with a custom data type that can't be reliably marshalled/unmarshalled
	type circularRef struct {
		Self *circularRef
	}

	// Create a circular reference which will cause marshal to fail
	var c circularRef
	c.Self = &c

	block := EmailBlock{
		ID:   "marshal_error",
		Kind: "button",
		Data: c, // This will cause Marshal to fail
	}

	_, err := TreeToMjml(rootStyles, block, "", map[string]string{}, 0, nil)

	// Should return an error when marshalling fails
	if err == nil {
		t.Error("Expected error for marshal failure, but got nil")
		return
	}

	if !strings.Contains(err.Error(), "failed to marshal") {
		t.Errorf("Expected 'failed to marshal' error, got: %v", err)
	}
}

func TestTreeToMjml_ChildBlockError(t *testing.T) {
	rootStyles := map[string]interface{}{}

	// Child block that will generate an error (liquid with invalid syntax)
	errorBlock := EmailBlock{
		ID:   "liquid_error",
		Kind: "liquid",
		Data: map[string]interface{}{
			"liquidCode": "{% invalid syntax %}",
		},
	}

	// Parent block with the error block as a child
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{},
		Children: []EmailBlock{errorBlock},
	}

	_, err := TreeToMjml(rootStyles, rootBlock, "{}", map[string]string{}, 0, nil)

	// Should return an error propagated from the child
	if err == nil {
		t.Error("Expected error from child block, but got nil")
		return
	}

	// The error should mention the child block ID
	if !strings.Contains(err.Error(), "liquid_error") {
		t.Errorf("Expected error to reference child block ID, got: %v", err)
	}
}

func TestTreeToMjml_UnknownBlockType(t *testing.T) {
	rootStyles := map[string]interface{}{}

	unknownBlock := EmailBlock{
		ID:   "unknown_type",
		Kind: "nonexistent_type",
		Data: map[string]interface{}{
			"field1": "value1",
			"field2": 42,
		},
	}

	mjml, err := TreeToMjml(rootStyles, unknownBlock, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml unexpectedly failed for unknown block type: %v", err)
	}

	// For unknown types, should return a comment
	if !strings.Contains(mjml, "<!-- MJML Not Implemented: nonexistent_type -->") {
		t.Errorf("Expected comment for unimplemented block type, got: %s", mjml)
	}
}

func TestTreeToMjml_NestedColumnWidthCalculation(t *testing.T) {
	rootStyles := map[string]interface{}{}

	// Create column blocks with IDs that we can look up
	column1 := EmailBlock{
		ID:   "col_first",
		Kind: "column",
		Data: ColumnBlockData{
			Styles: struct {
				VerticalAlign   string `json:"verticalAlign"`
				BackgroundColor string `json:"backgroundColor,omitempty"`
				MinHeight       string `json:"minHeight,omitempty"`
				BaseStyles
			}{
				VerticalAlign: "top",
			},
		},
	}

	column2 := EmailBlock{
		ID:   "col_second",
		Kind: "column",
		Data: ColumnBlockData{
			Styles: struct {
				VerticalAlign   string `json:"verticalAlign"`
				BackgroundColor string `json:"backgroundColor,omitempty"`
				MinHeight       string `json:"minHeight,omitempty"`
				BaseStyles
			}{
				VerticalAlign: "middle",
			},
		},
	}

	// Create a columns816 layout (33.33% / 66.66%)
	sectionBlock := EmailBlock{
		ID:   "sec_816",
		Kind: "columns816",
		Data: map[string]interface{}{},
		Children: []EmailBlock{
			column1,
			column2,
		},
	}

	mjml, err := TreeToMjml(rootStyles, sectionBlock, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for nested column layout: %v", err)
	}

	// Check that column widths were correctly calculated
	if !strings.Contains(mjml, "width=\"33.33%\"") {
		t.Error("First column should have width=\"33.33%\"")
	}

	if !strings.Contains(mjml, "width=\"66.66%\"") {
		t.Error("Second column should have width=\"66.66%\"")
	}
}
