package mjml

import (
	"strings"
	"testing"
)

// --- Test Helper Functions ---

// Helper to create a basic text block for testing
func createTextBlock(id, textContent string) EmailBlock {
	return EmailBlock{
		ID:   id,
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"editorData": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"children": []interface{}{
						map[string]interface{}{"text": textContent},
					},
				},
			},
		},
	}
}

// Helper to create basic root styles for testing
func createRootStyles() map[string]interface{} {
	return map[string]interface{}{
		"body": map[string]interface{}{
			"width":           "600px",
			"backgroundColor": "#ffffff",
		},
		"paragraph": map[string]interface{}{
			"color":      "#000000",
			"fontSize":   "14px",
			"fontWeight": 400,
			"fontFamily": "Arial",
			"margin":     "0px", // Explicitly test margin reset
		},
		"h1": map[string]interface{}{ // Add styles for h1 if needed for tests
			"color":    "#111111",
			"fontSize": "24px",
		},
		// Add other styles (h2, h3, hyperlink) if required by test cases
	}
}

// --- Tests for Helper Functions ---

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"camelCase", "camel-case"},
		{"PascalCase", "pascal-case"},
		{"lowercase", "lowercase"},
		{"UPPERCASE", "uppercase"},
		{"with1Number", "with1-number"},
		{"withID", "with-id"},
		{"MyURLValue", "my-url-value"},
		{"backgroundColor", "background-color"},
		{"fontFamily", "font-family"},
		{"HTMLElement", "html-element"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toKebabCase(tt.input)
			if result != tt.expect {
				t.Errorf("toKebabCase(%q) = %q; want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestTrackURL(t *testing.T) {
	params := map[string]string{
		"utm_source":   "test_source",
		"utm_medium":   "test_medium",
		"utm_campaign": "test_campaign",
	}
	tests := []struct {
		name     string
		inputURL string
		expect   string
	}{
		{"basic", "http://example.com", "http://example.com?utm_campaign=test_campaign&utm_medium=test_medium&utm_source=test_source"},
		{"withExistingParam", "http://example.com?other=val", "http://example.com?other=val&utm_campaign=test_campaign&utm_medium=test_medium&utm_source=test_source"},
		{"withExistingUTM", "http://example.com?utm_source=original", "http://example.com?utm_source=original"},
		{"https", "https://secure.example.com/path?p=1", "https://secure.example.com/path?p=1&utm_campaign=test_campaign&utm_medium=test_medium&utm_source=test_source"},
		{"liquidPlaceholder", "{{ variable_url }}", "{{ variable_url }}"},
		{"emptyURL", "", ""},
		{"mailto", "mailto:test@example.com", "mailto:test@example.com"},
		{"tel", "tel:+1234567890", "tel:+1234567890"},
		{"invalidURL", "://invalid", "://invalid"}, // Test invalid URL handling
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trackURL(tt.inputURL, params)
			// Comparing URLs can be tricky due to query param order, basic compare for now
			if result != tt.expect {
				t.Errorf("trackURL(%q) = %q; want %q", tt.inputURL, result, tt.expect)
			}
		})
	}
}

// --- Tests for TreeToMjml ---

func TestTreeToMjml_SimpleText(t *testing.T) {
	rootStyles := createRootStyles()
	block := createTextBlock("txt1", "Hello World")
	urlParams := map[string]string{}

	mjml, err := TreeToMjml(rootStyles, block, "", urlParams, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly: %v", err)
	}

	// Basic checks - not exact formatting
	if !strings.Contains(mjml, "<mj-text") {
		t.Error("Expected output to contain <mj-text>")
	}
	if !strings.Contains(mjml, "Hello World") {
		t.Error("Expected output to contain 'Hello World'")
	}
	if !strings.Contains(mjml, "align=\"left\"") {
		t.Error("Expected output to contain align attribute")
	}
	if !strings.Contains(mjml, "padding=\"0\"") {
		t.Error("Expected output to contain padding='0' attribute")
	}
	// Check for paragraph tag with styles from rootStyles
	if !strings.Contains(mjml, `<p style="color: #000000 !important; font-family: Arial !important; font-size: 14px !important; font-weight: 400 !important; margin: 0px !important">Hello World</p>`) {
		t.Errorf("Expected paragraph with styles not found in output:\n%s", mjml)
	}

	// t.Logf("Simple Text MJML:\n%s", mjml) // Uncomment to view output
}

func TestTreeToMjml_Nested(t *testing.T) {
	rootStyles := createRootStyles()
	textBlock := createTextBlock("txt_nested", "Nested Content")
	columnBlock := EmailBlock{
		ID:   "col1",
		Kind: "column",
		Data: map[string]interface{}{ // Add basic column data if needed
			"styles": map[string]interface{}{"verticalAlign": "top"},
		},
		Children: []EmailBlock{textBlock},
	}
	sectionBlock := EmailBlock{
		ID:   "sec1",
		Kind: "oneColumn", // Using oneColumn which acts like mj-section
		Data: map[string]interface{}{ // Add basic section data
			"styles": map[string]interface{}{"textAlign": "center"},
		},
		Children: []EmailBlock{columnBlock},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles}, // Root data contains the styles
		Children: []EmailBlock{sectionBlock},
	}

	urlParams := map[string]string{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", urlParams, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for nested structure: %v", err)
	}

	// Check for expected tags and content
	if !strings.Contains(mjml, "<mjml>") {
		t.Error("Missing <mjml> tag")
	}
	if !strings.Contains(mjml, "<mj-body") {
		t.Error("Missing <mj-body> tag")
	}
	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Missing <mj-section> tag")
	}
	if !strings.Contains(mjml, "text-align=\"center\"") {
		t.Error("Missing text-align on section")
	}
	if !strings.Contains(mjml, "<mj-column") {
		t.Error("Missing <mj-column> tag")
	}
	if !strings.Contains(mjml, "vertical-align=\"top\"") {
		t.Error("Missing vertical-align on column")
	}
	if !strings.Contains(mjml, "<mj-text") {
		t.Error("Missing <mj-text> tag")
	}
	if !strings.Contains(mjml, "Nested Content") {
		t.Error("Missing nested content 'Nested Content'")
	}

	// t.Logf("Nested MJML:\n%s", mjml) // Uncomment to view output
}

func TestTreeToMjml_Liquid(t *testing.T) {
	rootStyles := createRootStyles()
	liquidBlock := EmailBlock{
		ID:   "liq1",
		Kind: "liquid",
		Data: map[string]interface{}{
			"liquidCode": "<p>Hello {{ name }}!</p>",
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{liquidBlock},
	}

	urlParams := map[string]string{}
	templateData := `{"name": "LiquidUser"}`

	mjml, err := TreeToMjml(rootStyles, rootBlock, templateData, urlParams, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for liquid block: %v", err)
	}

	// Check if liquid was processed - liquid block content is inserted raw
	if !strings.Contains(mjml, "<p>Hello LiquidUser!</p>") {
		t.Errorf("Expected processed liquid content not found in MJML:\n%s", mjml)
	}

	// t.Logf("Liquid MJML:\n%s", mjml) // Uncomment to view output
}

func TestTreeToMjml_LiquidError(t *testing.T) {
	rootStyles := createRootStyles()
	liquidBlock := EmailBlock{
		ID:   "liq_err",
		Kind: "liquid",
		Data: map[string]interface{}{
			"liquidCode": "{% invalid tag %}",
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{liquidBlock},
	}

	urlParams := map[string]string{}
	templateData := `{}`

	_, err := TreeToMjml(rootStyles, rootBlock, templateData, urlParams, 0, nil)
	if err == nil {
		t.Fatal("TreeToMjml should have failed for invalid liquid tag, but err was nil")
	}

	// Check if the error message indicates a liquid parsing/rendering error
	if !strings.Contains(err.Error(), "liquid rendering error") {
		t.Errorf("Expected error message to contain 'liquid rendering error', got: %v", err)
	}

	// t.Logf("Liquid Error: %v", err) // Uncomment to view error
}

func TestTreeToMjml_ImageWithTracking(t *testing.T) {
	rootStyles := createRootStyles()
	imageBlock := EmailBlock{
		ID:   "img1",
		Kind: "image",
		Data: map[string]interface{}{ // Data structure mirrors ImageBlockData roughly
			"image": map[string]interface{}{ // Nested 'image' field
				"src":              "http://example.com/image.png",
				"alt":              "Test Image",
				"href":             "http://example.com/link",
				"disable_tracking": false, // Explicitly enable tracking
			},
			"wrapper": map[string]interface{}{ // Nested 'wrapper' field
				"align": "center",
			},
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{imageBlock},
	}

	urlParams := map[string]string{
		"utm_source": "test_img_src",
		"utm_medium": "email",
	}

	mjml, err := TreeToMjml(rootStyles, rootBlock, "", urlParams, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for image block: %v", err)
	}

	if !strings.Contains(mjml, "<mj-image") {
		t.Error("Expected output to contain <mj-image>")
	}
	if !strings.Contains(mjml, `src="http://example.com/image.png"`) {
		t.Error("Missing correct src attribute")
	}
	if !strings.Contains(mjml, `alt="Test Image"`) {
		t.Error("Missing correct alt attribute")
	}
	// Check for tracked URL
	expectedTrackedHref := `href="http://example.com/link?utm_medium=email&utm_source=test_img_src"`
	if !strings.Contains(mjml, expectedTrackedHref) {
		t.Errorf("Expected tracked href %q not found in output:\n%s", expectedTrackedHref, mjml)
	}
	if !strings.Contains(mjml, `target="_blank"`) || !strings.Contains(mjml, `rel="noopener noreferrer"`) {
		t.Error("Missing target or rel attribute on tracked link")
	}
	if !strings.Contains(mjml, `align="center"`) {
		t.Error("Missing align attribute from wrapper")
	}

	// t.Logf("Image MJML:\n%s", mjml) // Uncomment to view output
}

func TestTreeToMjml_NilData(t *testing.T) {
	rootStyles := createRootStyles()
	block := EmailBlock{
		ID:       "nil_data_block",
		Kind:     "text", // Use a kind that expects data
		Data:     nil,    // Explicitly set data to nil
		Children: nil,
	}

	mjml, err := TreeToMjml(rootStyles, block, "", nil, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly with nil data: %v", err)
	}

	// Expect an empty or minimal mj-text tag, as data was nil
	expected := `<mj-text padding="0"></mj-text>`
	if strings.TrimSpace(mjml) != expected {
		t.Errorf("Expected minimal tag %q for nil data, got: %q", expected, strings.TrimSpace(mjml))
	}
}
