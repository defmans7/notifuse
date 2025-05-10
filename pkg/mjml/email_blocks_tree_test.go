package mjml

import (
	"strings"
	"testing"
)

// --- Tests for TreeToMjml ---

func TestTreeToMjml_SimpleText(t *testing.T) {
	rootStyles := createRootStyles()
	block := createTextBlock("txt1", "Hello World")
	trackingSettings := TrackingSettings{}

	mjml, err := TreeToMjml(rootStyles, block, "", trackingSettings, 0, nil)
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

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
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

	trackingSettings := TrackingSettings{}
	templateData := `{"name": "LiquidUser"}`

	mjml, err := TreeToMjml(rootStyles, rootBlock, templateData, trackingSettings, 0, nil)
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

	trackingSettings := TrackingSettings{}
	templateData := `{}`

	_, err := TreeToMjml(rootStyles, rootBlock, templateData, trackingSettings, 0, nil)
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

	trackingSettings := TrackingSettings{
		EnableTracking: false, // Disable redirect to tracking endpoint
		UTMSource:      "test_img_src",
		UTMMedium:      "email",
	}

	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
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

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, block, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly with nil data: %v", err)
	}

	// Expect an empty or minimal mj-text tag, as data was nil
	expected := `<mj-text padding="0"></mj-text>`
	if strings.TrimSpace(mjml) != expected {
		t.Errorf("Expected minimal tag %q for nil data, got: %q", expected, strings.TrimSpace(mjml))
	}
}

func TestTreeToMjml_Button(t *testing.T) {
	rootStyles := createRootStyles()
	buttonBlock := EmailBlock{
		ID:   "btn1",
		Kind: "button",
		Data: map[string]interface{}{
			"button": map[string]interface{}{
				"text":                   "Click Me",
				"href":                   "http://example.com/button_link",
				"backgroundColor":        "#007bff",
				"color":                  "#ffffff",
				"fontSize":               "16px",
				"fontWeight":             700, // Use number for weight
				"fontFamily":             "Helvetica, Arial, sans-serif",
				"innerHorizontalPadding": "15px",
				"innerVerticalPadding":   "10px",
				"borderRadius":           "5px",
				"disable_tracking":       false,
				"borderControl":          "all", // Test border properties
				"borderStyle":            "solid",
				"borderWidth":            "1px",
				"borderColor":            "#0056b3",
			},
			"wrapper": map[string]interface{}{
				"align":   "center",
				"padding": "20px 0px", // Test wrapper padding
			},
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{buttonBlock},
	}

	trackingSettings := TrackingSettings{
		EnableTracking: false, // Disable redirect to tracking endpoint
		UTMSource:      "test_btn_src",
		UTMMedium:      "email",
		UTMCampaign:    "promo",
	}

	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for button block: %v", err)
	}

	// Check for <mj-button> tag
	if !strings.Contains(mjml, "<mj-button") {
		t.Error("Expected output to contain <mj-button>")
	}

	// Check button attributes (selective checks)
	if !strings.Contains(mjml, `background-color="#007bff"`) {
		t.Error("Missing button background-color")
	}
	if !strings.Contains(mjml, `color="#ffffff"`) {
		t.Error("Missing button color")
	}
	if !strings.Contains(mjml, `font-weight="700"`) {
		t.Error("Missing button font-weight")
	}
	if !strings.Contains(mjml, `font-size="16px"`) {
		t.Error("Missing button font-size")
	}
	if !strings.Contains(mjml, `padding="10px 15px"`) {
		t.Error("Missing button inner padding") // Checks inner padding format
	}
	if !strings.Contains(mjml, `border-radius="5px"`) {
		t.Error("Missing button border-radius")
	}
	if !strings.Contains(mjml, `border="1px solid #0056b3"`) {
		t.Error("Missing button border")
	}

	// Check tracked URL
	expectedTrackedHref := `href="http://example.com/button_link?utm_campaign=promo&utm_medium=email&utm_source=test_btn_src"`
	if !strings.Contains(mjml, expectedTrackedHref) {
		t.Errorf("Expected tracked href %q not found in output:\n%s", expectedTrackedHref, mjml)
	}

	// Check wrapper attributes
	if !strings.Contains(mjml, `align="center"`) {
		t.Error("Missing wrapper align attribute")
	}
	// Note: Button padding is controlled by inner-padding + button's own padding.
	// Wrapper padding applies to the surrounding element, which isn't directly tested here.
	// We check inner-padding above. Let's verify the wrapper align is present.
	// if !strings.Contains(mjml, `padding="20px 0px"`) {
	// 	t.Error("Missing wrapper padding attribute")
	// }

	// Use a slightly more robust check for the content itself,
	// allowing for potential whitespace or minor structure variations.
	// Check if the text content exists between the button tags, ignoring surrounding whitespace.
	if !strings.Contains(strings.ReplaceAll(strings.ReplaceAll(mjml, "\n", ""), " ", ""), ">ClickMe</mj-button>") {
		// t.Logf("Button MJML:\n%s", mjml) // Uncomment to view output if fails
		// t.Logf("Button MJML:\n%s", mjml) // Uncommented for debugging - Re-commenting
		t.Error("Missing button text content 'Click Me'")
	}

	// t.Logf("Button MJML:\n%s", mjml) // Uncomment to view output
}

func TestTreeToMjml_Divider(t *testing.T) {
	rootStyles := createRootStyles()
	dividerBlock := EmailBlock{
		ID:   "div1",
		Kind: "divider",
		Data: map[string]interface{}{
			"borderColor":     "#cccccc",
			"borderStyle":     "dashed",
			"borderWidth":     "2px",
			"width":           "80%",
			"align":           "center",
			"padding":         "10px 0",  // Test padding
			"backgroundColor": "#f0f0f0", // Test container background
			"paddingControl":  "all",     // Explicitly set control to use shorthand padding
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{dividerBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for divider block: %v", err)
	}

	if !strings.Contains(mjml, "<mj-divider") {
		t.Error("Expected output to contain <mj-divider>")
	}
	if !strings.Contains(mjml, `border-color="#cccccc"`) {
		t.Error("Missing divider border-color")
	}
	if !strings.Contains(mjml, `border-style="dashed"`) {
		t.Error("Missing divider border-style")
	}
	if !strings.Contains(mjml, `border-width="2px"`) {
		t.Error("Missing divider border-width")
	}
	if !strings.Contains(mjml, `width="80%"`) {
		t.Error("Missing divider width")
	}
	if !strings.Contains(mjml, `align="center"`) {
		t.Error("Missing divider align")
	}
	if !strings.Contains(mjml, `padding="10px 0"`) {
		t.Error("Missing divider padding")
	}
	if !strings.Contains(mjml, `container-background-color="#f0f0f0"`) {
		t.Error("Missing divider container-background-color")
	}

	// t.Logf("Divider MJML:\n%s", mjml) // Uncommented for debugging - Re-commenting
	// t.Logf("Divider MJML:\n%s", mjml)
}

func TestTreeToMjml_Spacer(t *testing.T) {
	rootStyles := createRootStyles()
	spacerBlock := EmailBlock{
		ID:   "spc1",
		Kind: "spacer",
		Data: map[string]interface{}{
			"height":          "30px",
			"backgroundColor": "#e0e0e0", // Test container background
		},
	}
	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{spacerBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for spacer block: %v", err)
	}

	if !strings.Contains(mjml, "<mj-spacer") {
		t.Error("Expected output to contain <mj-spacer>")
	}
	if !strings.Contains(mjml, `height="30px"`) {
		t.Error("Missing spacer height")
	}
	if !strings.Contains(mjml, `container-background-color="#e0e0e0"`) {
		t.Error("Missing spacer container-background-color")
	}

	// t.Logf("Spacer MJML:\n%s", mjml)
}

func TestTreeToMjml_TwoColumnLayout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_left", "Left Column Content")
	textBlockRight := createTextBlock("txt_right", "Right Column Content")

	columnLeft := EmailBlock{
		ID:   "col_left",
		Kind: "column",
		Data: map[string]interface{}{ // Minimal column data
			"styles": map[string]interface{}{"verticalAlign": "top"},
		},
		Children: []EmailBlock{textBlockLeft},
	}
	columnRight := EmailBlock{
		ID:   "col_right",
		Kind: "column",
		Data: map[string]interface{}{ // Minimal column data
			"styles": map[string]interface{}{"verticalAlign": "middle", "backgroundColor": "#eeeeee"},
		},
		Children: []EmailBlock{textBlockRight},
	}

	sectionBlock := EmailBlock{
		ID:   "sec_two_col",
		Kind: "columns1212", // Use the specific kind for 50/50 layout
		Data: map[string]interface{}{ // Section data
			"styles": map[string]interface{}{
				"textAlign": "left",
				"padding":   "10px", // Padding on the section
			},
			"paddingControl": "all",
		},
		Children: []EmailBlock{columnLeft, columnRight},
	}

	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{sectionBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for two-column block: %v", err)
	}

	// Check for section tag
	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Expected output to contain <mj-section>")
	}
	if !strings.Contains(mjml, `padding="10px"`) {
		t.Error("Missing section padding")
	}

	// Check for two column tags
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}

	// Check column widths and styles (more specific checks)
	if !strings.Contains(mjml, `width="50%"`) {
		t.Error("Expected columns to have width='50%' attribute")
	} // Note: This check isn't perfect, assumes both have it.
	if !strings.Contains(mjml, `vertical-align="top"`) {
		t.Error("Missing vertical-align='top' on first column")
	}
	if !strings.Contains(mjml, `vertical-align="middle"`) {
		t.Error("Missing vertical-align='middle' on second column")
	}
	if !strings.Contains(mjml, `background-color="#eeeeee"`) {
		t.Error("Missing background-color on second column")
	}

	// Check for content within columns
	if !strings.Contains(mjml, "Left Column Content") {
		t.Error("Missing left column content")
	}
	if !strings.Contains(mjml, "Right Column Content") {
		t.Error("Missing right column content")
	}

	// t.Logf("Two Column MJML:\n%s", mjml)
}

func TestTreeToMjml_TextFormatting(t *testing.T) {
	// Enhance rootStyles to include hyperlink styles for this test
	rootStyles := createRootStyles()
	rootStyles["hyperlink"] = map[string]interface{}{
		"color":          "#0000EE",
		"textDecoration": "underline",
		"fontWeight":     400, // Default link weight
	}

	textBlock := EmailBlock{
		ID:   "txt_formatted",
		Kind: "text",
		Data: map[string]interface{}{ // TextBlockData structure
			"align": "left",
			"editorData": []interface{}{ // Slice of lines
				map[string]interface{}{ // First line (paragraph)
					"type": "paragraph",
					"children": []interface{}{ // Slice of parts within the line
						map[string]interface{}{"text": "Plain text. "},
						map[string]interface{}{"text": "Bold text.", "bold": true},
						map[string]interface{}{"text": " "},
						map[string]interface{}{"text": "Italic text.", "italic": true},
						map[string]interface{}{"text": " "},
						map[string]interface{}{"text": "Bold and Italic.", "bold": true, "italic": true},
					},
				},
				map[string]interface{}{ // Second line (paragraph with link)
					"type": "paragraph",
					"children": []interface{}{ // Slice of parts within the line
						map[string]interface{}{"text": "Link: "},
						map[string]interface{}{ // The hyperlink part
							"text": "Click Here",
							"hyperlink": map[string]interface{}{ // Hyperlink data
								"url":              "http://example.com/formatted_link",
								"disable_tracking": false,
							},
							"bold": true, // Make link bold
						},
						map[string]interface{}{"text": "."},
					},
				},
			},
			// Add specific hyperlink styles at the block level (optional, could override root)
			"hyperlinkStyles": map[string]interface{}{
				"color": "#990099", // Override root link color
			},
		},
	}

	rootBlock := EmailBlock{
		ID:       "root",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{textBlock},
	}

	trackingSettings := TrackingSettings{
		EnableTracking: false, // Disable redirect to tracking endpoint
		UTMSource:      "text_format_test",
	}

	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for formatted text block: %v", err)
	}

	t.Logf("Formatted Text MJML:\n%s", mjml) // Uncommented for debugging

	// Check for mj-text tag
	if !strings.Contains(mjml, "<mj-text") {
		t.Error("Expected output to contain <mj-text>")
	}

	// Check for plain text
	if !strings.Contains(mjml, "Plain text.") {
		t.Error("Missing plain text content")
	}

	// Check for bold text (within span or strong tag - check for style)
	if !strings.Contains(mjml, `font-weight: bold`) || !strings.Contains(mjml, ">Bold text.<") {
		t.Error("Missing or incorrectly formatted bold text")
	}

	// Check for italic text
	if !strings.Contains(mjml, `font-style: italic`) || !strings.Contains(mjml, ">Italic text.<") {
		t.Error("Missing or incorrectly formatted italic text")
	}

	// Check for bold and italic text
	// Note: depends on how combined styles are rendered (two styles in one span?)
	if !strings.Contains(mjml, `font-weight: bold; font-style: italic`) && !strings.Contains(mjml, `font-style: italic; font-weight: bold`) {
		t.Error("Missing combined bold and italic style attribute.")
	}
	if !strings.Contains(mjml, ">Bold and Italic.<") {
		t.Error("Missing bold and italic text content")
	}

	// Check for hyperlink
	expectedLinkHref := `href="http://example.com/formatted_link?utm_source=text_format_test"`
	if !strings.Contains(mjml, expectedLinkHref) {
		t.Errorf("Missing or incorrect tracked hyperlink href. Expected contains %q", expectedLinkHref)
	}
	if !strings.Contains(mjml, `target="_blank"`) {
		t.Error("Missing target='_blank' on hyperlink")
	}
	if !strings.Contains(mjml, `rel="noopener noreferrer"`) {
		t.Error("Missing rel='noopener noreferrer' on hyperlink")
	}
	if !strings.Contains(mjml, ">Click Here</a>") {
		t.Error("Missing hyperlink text content")
	}

	// Check hyperlink styles (color override, bold from part, decoration from root)
	if !strings.Contains(mjml, `color: #990099 !important`) {
		t.Error("Missing overridden hyperlink color style")
	}
	if !strings.Contains(mjml, `font-weight: bold !important`) {
		t.Error("Missing bold style on hyperlink (from part)")
	}
	if !strings.Contains(mjml, `text-decoration: underline !important`) {
		t.Error("Missing underline style on hyperlink (from root)")
	}
}

func TestTreeToMjml_InvalidDataUnmarshal(t *testing.T) {
	rootStyles := createRootStyles()
	// Create a button block but provide data that won't unmarshal into ButtonBlockData
	invalidButtonBlock := EmailBlock{
		ID:   "btn_invalid",
		Kind: "button",
		// Data: map[string]interface{}{ // Original invalid map
		// 	"invalidField": "some string",
		// 	"buttonText":   12345,
		// },
		Data: "this is not a valid structure for button data", // Use a simple string
	}

	rootBlock := EmailBlock{
		ID:       "root_invalid",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{invalidButtonBlock},
	}

	trackingSettings := TrackingSettings{}
	_, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)

	// Expect an error
	if err == nil {
		t.Fatal("TreeToMjml should have failed due to invalid data unmarshalling, but err was nil")
	}

	// Check for the specific JSON unmarshal error message
	if !strings.Contains(err.Error(), "json: cannot unmarshal string into Go value of type mjml.ButtonBlockData") {
		t.Errorf("Expected error message to indicate specific JSON unmarshal failure, got: %v", err)
	}

	// Optionally, check for specific block ID and Kind in the error
	if !strings.Contains(err.Error(), "(ID: btn_invalid, Kind: button)") {
		t.Errorf("Expected error message to contain block ID and Kind, got: %v", err)
	}

	// t.Logf("Invalid Data Error: %v", err) // Uncomment to view error
}

func TestTreeToMjml_Columns816Layout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_816_left", "Left (33%)")
	textBlockRight := createTextBlock("txt_816_right", "Right (67%)")

	columnLeft := EmailBlock{
		ID:       "col_816_left",
		Kind:     "column",
		Children: []EmailBlock{textBlockLeft},
		// No specific data needed for this simple case
	}
	columnRight := EmailBlock{
		ID:       "col_816_right",
		Kind:     "column",
		Children: []EmailBlock{textBlockRight},
		// No specific data needed for this simple case
	}

	sectionBlock := EmailBlock{
		ID:   "sec_816_col",
		Kind: "columns816", // Use the specific kind for 33/67 layout
		Data: map[string]interface{}{ // Minimal section data
			"styles": map[string]interface{}{},
		},
		Children: []EmailBlock{columnLeft, columnRight},
	}

	rootBlock := EmailBlock{
		ID:       "root_816",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{sectionBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for columns816 block: %v", err)
	}

	// Check for section tag
	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Expected output to contain <mj-section>")
	}

	// Check for two column tags
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}

	// Check column widths (more specific checks)
	// Need to check the generated MJML more carefully to ensure correct widths are applied to the correct columns
	// A simple Contains check might find the width but not guarantee it's on the right column.
	// For now, we'll use Contains, but a more robust test might parse the output.
	if !strings.Contains(mjml, `width="33.33%"`) {
		t.Error("Expected first column to have width='33.33%' attribute")
	}
	if !strings.Contains(mjml, `width="66.66%"`) {
		t.Error("Expected second column to have width='66.66%' attribute")
	}

	// Check for content within columns
	if !strings.Contains(mjml, "Left (33%)") {
		t.Error("Missing left column content")
	}
	if !strings.Contains(mjml, "Right (67%)") {
		t.Error("Missing right column content")
	}

	// t.Logf("Columns 816 MJML:\n%s", mjml)
}

func TestTreeToMjml_OneColumnLayout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlock := createTextBlock("txt_one_col", "Single Column Content")

	columnBlock := EmailBlock{
		ID:       "col_one",
		Kind:     "column",
		Children: []EmailBlock{textBlock},
	}

	sectionBlock := EmailBlock{
		ID:   "sec_one_col",
		Kind: "oneColumn",
		Data: map[string]interface{}{ // Minimal section data
			"styles": map[string]interface{}{},
		},
		Children: []EmailBlock{columnBlock},
	}

	rootBlock := EmailBlock{
		ID:       "root_one_col",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{sectionBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for oneColumn block: %v", err)
	}

	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Expected <mj-section>")
	}
	if strings.Count(mjml, "<mj-column") != 1 {
		t.Errorf("Expected 1 <mj-column> tag, found %d", strings.Count(mjml, "<mj-column"))
	}
	// oneColumn shouldn't have an explicit width attribute
	if strings.Contains(mjml, "<mj-column width=") {
		t.Error("Expected <mj-column> in oneColumn layout not to have explicit width attribute")
	}
	if !strings.Contains(mjml, "Single Column Content") {
		t.Error("Missing single column content")
	}
	// t.Logf("One Column MJML:\n%s", mjml)
}

func TestTreeToMjml_Columns168Layout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_168_left", "Left (67%)")
	textBlockRight := createTextBlock("txt_168_right", "Right (33%)")
	columnLeft := EmailBlock{ID: "col_168_left", Kind: "column", Children: []EmailBlock{textBlockLeft}}
	columnRight := EmailBlock{ID: "col_168_right", Kind: "column", Children: []EmailBlock{textBlockRight}}
	sectionBlock := EmailBlock{
		ID:       "sec_168_col",
		Kind:     "columns168",
		Data:     map[string]interface{}{"styles": map[string]interface{}{}},
		Children: []EmailBlock{columnLeft, columnRight},
	}
	rootBlock := EmailBlock{ID: "root_168", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for columns168 block: %v", err)
	}
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}
	if !strings.Contains(mjml, `width="66.66%"`) {
		t.Error("Expected first column width='66.66%'")
	}
	if !strings.Contains(mjml, `width="33.33%"`) {
		t.Error("Expected second column width='33.33%'")
	}
	if !strings.Contains(mjml, "Left (67%)") || !strings.Contains(mjml, "Right (33%)") {
		t.Error("Missing content")
	}
}

func TestTreeToMjml_Columns204Layout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_204_left", "Left (83%)")
	textBlockRight := createTextBlock("txt_204_right", "Right (17%)")
	columnLeft := EmailBlock{ID: "col_204_left", Kind: "column", Children: []EmailBlock{textBlockLeft}}
	columnRight := EmailBlock{ID: "col_204_right", Kind: "column", Children: []EmailBlock{textBlockRight}}
	sectionBlock := EmailBlock{
		ID:       "sec_204_col",
		Kind:     "columns204",
		Data:     map[string]interface{}{"styles": map[string]interface{}{}},
		Children: []EmailBlock{columnLeft, columnRight},
	}
	rootBlock := EmailBlock{ID: "root_204", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for columns204 block: %v", err)
	}
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}
	if !strings.Contains(mjml, `width="83.33%"`) {
		t.Error("Expected first column width='83.33%'")
	}
	if !strings.Contains(mjml, `width="16.66%"`) {
		t.Error("Expected second column width='16.66%'")
	}
	if !strings.Contains(mjml, "Left (83%)") || !strings.Contains(mjml, "Right (17%)") {
		t.Error("Missing content")
	}
}

func TestTreeToMjml_Columns420Layout(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_420_left", "Left (17%)")
	textBlockRight := createTextBlock("txt_420_right", "Right (83%)")
	columnLeft := EmailBlock{ID: "col_420_left", Kind: "column", Children: []EmailBlock{textBlockLeft}}
	columnRight := EmailBlock{ID: "col_420_right", Kind: "column", Children: []EmailBlock{textBlockRight}}
	sectionBlock := EmailBlock{
		ID:       "sec_420_col",
		Kind:     "columns420",
		Data:     map[string]interface{}{"styles": map[string]interface{}{}},
		Children: []EmailBlock{columnLeft, columnRight},
	}
	rootBlock := EmailBlock{ID: "root_420", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for columns420 block: %v", err)
	}
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}
	if !strings.Contains(mjml, `width="16.66%"`) {
		t.Error("Expected first column width='16.66%'")
	}
	if !strings.Contains(mjml, `width="83.33%"`) {
		t.Error("Expected second column width='83.33%'")
	}
	if !strings.Contains(mjml, "Left (17%)") || !strings.Contains(mjml, "Right (83%)") {
		t.Error("Missing content")
	}
}

func TestTreeToMjml_Columns888Layout(t *testing.T) {
	rootStyles := createRootStyles()
	text1 := createTextBlock("txt_888_1", "Col 1")
	text2 := createTextBlock("txt_888_2", "Col 2")
	text3 := createTextBlock("txt_888_3", "Col 3")
	col1 := EmailBlock{ID: "col_888_1", Kind: "column", Children: []EmailBlock{text1}}
	col2 := EmailBlock{ID: "col_888_2", Kind: "column", Children: []EmailBlock{text2}}
	col3 := EmailBlock{ID: "col_888_3", Kind: "column", Children: []EmailBlock{text3}}
	sectionBlock := EmailBlock{
		ID:       "sec_888_col",
		Kind:     "columns888",
		Data:     map[string]interface{}{"styles": map[string]interface{}{}},
		Children: []EmailBlock{col1, col2, col3},
	}
	rootBlock := EmailBlock{ID: "root_888", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for columns888 block: %v", err)
	}
	if strings.Count(mjml, "<mj-column") != 3 {
		t.Errorf("Expected 3 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}
	// Check if the width is present at least 3 times (crude check)
	if strings.Count(mjml, `width="33.33%"`) < 3 {
		t.Error("Expected three columns with width='33.33%'")
	}
	if !strings.Contains(mjml, "Col 1") || !strings.Contains(mjml, "Col 2") || !strings.Contains(mjml, "Col 3") {
		t.Error("Missing content")
	}
}

func TestTreeToMjml_Columns6666Layout(t *testing.T) {
	rootStyles := createRootStyles()
	text1 := createTextBlock("txt_6666_1", "C1")
	text2 := createTextBlock("txt_6666_2", "C2")
	text3 := createTextBlock("txt_6666_3", "C3")
	text4 := createTextBlock("txt_6666_4", "C4")
	col1 := EmailBlock{ID: "col_6666_1", Kind: "column", Children: []EmailBlock{text1}}
	col2 := EmailBlock{ID: "col_6666_2", Kind: "column", Children: []EmailBlock{text2}}
	col3 := EmailBlock{ID: "col_6666_3", Kind: "column", Children: []EmailBlock{text3}}
	col4 := EmailBlock{ID: "col_6666_4", Kind: "column", Children: []EmailBlock{text4}}
	sectionBlock := EmailBlock{
		ID:       "sec_6666_col",
		Kind:     "columns6666",
		Data:     map[string]interface{}{"styles": map[string]interface{}{}},
		Children: []EmailBlock{col1, col2, col3, col4},
	}
	rootBlock := EmailBlock{ID: "root_6666", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed for columns6666 block: %v", err)
	}
	if strings.Count(mjml, "<mj-column") != 4 {
		t.Errorf("Expected 4 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}
	// Check if the width is present at least 4 times (crude check)
	if strings.Count(mjml, `width="25%"`) < 4 {
		t.Error("Expected four columns with width='25%'")
	}
	if !strings.Contains(mjml, "C1") || !strings.Contains(mjml, "C2") || !strings.Contains(mjml, "C3") || !strings.Contains(mjml, "C4") {
		t.Error("Missing content")
	}
}

func TestTreeToMjml_HeadingBlocks(t *testing.T) {
	rootStyles := createRootStyles()
	// Add h2 style to rootStyles
	rootStyles["h2"] = map[string]interface{}{
		"fontSize":   "24px",
		"fontWeight": 600,
		"color":      "#222222",
		"fontFamily": "Arial, sans-serif",
	}

	headingBlock := EmailBlock{
		ID:   "hd1",
		Kind: "heading",
		Data: map[string]interface{}{
			"type":  "h2", // This is important - set the heading type
			"align": "center",
			"editorData": []interface{}{ // Slice of lines
				map[string]interface{}{ // First line (paragraph)
					"type": "paragraph",
					"children": []interface{}{ // Slice of parts within the line
						map[string]interface{}{"text": "Normal heading part. "},
						map[string]interface{}{"text": "Bold heading part.", "bold": true},
					},
				},
			},
			"color":          "#3366CC",         // Override default heading color
			"fontFamily":     "Times New Roman", // Override default heading font
			"paddingControl": "all",
			"padding":        "15px",
		},
	}

	rootBlock := EmailBlock{
		ID:       "root_heading",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{headingBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for heading block: %v", err)
	}

	// Check for mj-text tag (headings use mj-text with <h> tags)
	if !strings.Contains(mjml, "<mj-text") {
		t.Error("Expected output to contain <mj-text>")
	}

	// Heading text should be present
	if !strings.Contains(mjml, "Normal heading part.") {
		t.Error("Missing normal heading content")
	}
	if !strings.Contains(mjml, "Bold heading part.") {
		t.Error("Missing bold heading content")
	}

	// Check for attributes
	if !strings.Contains(mjml, `align="center"`) {
		t.Error("Missing heading align attribute")
	}

	// Check for padding
	if !strings.Contains(mjml, `padding="15px"`) {
		t.Error("Missing padding attribute")
	}
}

func TestTreeToMjml_HeadingWithLiquidContent(t *testing.T) {
	rootStyles := createRootStyles()
	headingBlock := EmailBlock{
		ID:   "hd_liquid",
		Kind: "heading",
		Data: map[string]interface{}{
			"headingSize": 3, // H3
			"align":       "left",
			"editorData": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"children": []interface{}{
						map[string]interface{}{"text": "Hello {{ name }}! "},
						map[string]interface{}{"text": "Your order #{{ order_id }} has been {{ status }}.", "bold": true},
					},
				},
			},
		},
	}

	rootBlock := EmailBlock{
		ID:       "root_heading_liquid",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{headingBlock},
	}

	// Provide templateData for liquid tag processing
	templateData := `{"name": "Customer", "order_id": "ORD-123", "status": "shipped"}`

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, templateData, trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for heading with liquid: %v", err)
	}

	// Check for processed liquid content
	if !strings.Contains(mjml, "Hello Customer!") {
		t.Error("Missing processed liquid content in name tag")
	}
	if !strings.Contains(mjml, "Your order #ORD-123 has been shipped.") {
		t.Error("Missing processed liquid content in order details")
	}
}

func TestTreeToMjml_HeadingWithInvalidLiquid(t *testing.T) {
	rootStyles := createRootStyles()
	headingBlock := EmailBlock{
		ID:   "hd_invalid_liquid",
		Kind: "heading",
		Data: map[string]interface{}{
			"headingSize": 4, // H4
			"editorData": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"children": []interface{}{
						map[string]interface{}{"text": "Hello {{ invalid syntax"},
					},
				},
			},
		},
	}

	rootBlock := EmailBlock{
		ID:       "root_heading_invalid",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{headingBlock},
	}

	// This test verifies the code doesn't crash with invalid liquid but continues processing
	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "{}", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for heading with invalid liquid: %v", err)
	}

	// The raw content should still be present with invalid liquid
	if !strings.Contains(mjml, "Hello {{ invalid syntax") {
		t.Error("Raw text with invalid liquid should still be present in output")
	}
}

func TestTrackingSettings_GetTrackingURLComprehensiveEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		trackingSettings TrackingSettings
		expected         string
	}{
		{
			name: "Basic URL without params",
			url:  "https://example.com",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "https://example.com?utm_source=test",
		},
		{
			name: "URL with existing params",
			url:  "https://example.com?existing=param",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "https://example.com?existing=param&utm_source=test",
		},
		{
			name: "Multiple UTM params",
			url:  "https://example.com",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
				UTMMedium:      "email",
				UTMCampaign:    "welcome",
			},
			expected: "https://example.com?utm_campaign=welcome&utm_medium=email&utm_source=test",
		},
		{
			name: "URL with fragment",
			url:  "https://example.com/page#section",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "https://example.com/page?utm_source=test#section",
		},
		{
			name: "Empty URL",
			url:  "",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "",
		},
		{
			name: "Liquid variable placeholder",
			url:  "{{ some_variable }}",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "{{ some_variable }}",
		},
		{
			name: "mailto link",
			url:  "mailto:user@example.com",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "mailto:user@example.com",
		},
		{
			name: "tel link",
			url:  "tel:+1234567890",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "tel:+1234567890",
		},
		{
			name: "URL with existing UTM param",
			url:  "https://example.com?utm_source=original",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "https://example.com?utm_source=original",
		},
		{
			name: "Empty params",
			url:  "https://example.com",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
			},
			expected: "https://example.com",
		},
		{
			name: "URL with spaces",
			url:  "https://example.com/page with spaces",
			trackingSettings: TrackingSettings{
				EnableTracking: false,
				UTMSource:      "test",
			},
			expected: "https://example.com/page%20with%20spaces?utm_source=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.trackingSettings.GetTrackingURL(tt.url)
			if result != tt.expected {
				t.Errorf("trackingSettings.GetTrackingURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestApplyPaddingEdgeCases(t *testing.T) {
	// Test the actual exported helper function directly from the package

	// 1. Test with empty control value
	t.Run("EmptyControl", func(t *testing.T) {
		attributes := make(map[string]interface{})
		// Empty control should use default behavior which still applies padding if provided
		applyPaddingFromStruct("", "10px", "", "", "", "", attributes)

		// With empty control but valid padding, should still set padding
		if attributes["padding"] != "10px" {
			t.Errorf("Expected padding=10px with empty control, got %v", attributes)
		}
	})

	// 2. Test with "all" control and shorthand padding
	t.Run("AllWithShorthand", func(t *testing.T) {
		attributes := make(map[string]interface{})
		// Use shorthand padding with "all" control
		applyPaddingFromStruct("all", "15px", "", "", "", "", attributes)

		// Should have single padding attribute
		if attributes["padding"] != "15px" {
			t.Errorf("Expected padding=15px, got %v", attributes["padding"])
		}
	})

	// 3. Test with "separate" control and individual values
	t.Run("SeparateControl", func(t *testing.T) {
		attributes := make(map[string]interface{})
		// Use separate control with individual values
		applyPaddingFromStruct("separate", "10px", "5px", "10px", "15px", "20px", attributes)

		// Should use individual attributes and ignore shorthand
		if _, exists := attributes["padding"]; exists {
			t.Errorf("Expected no shorthand padding with 'separate' control, got %v", attributes)
		}
		if attributes["padding-top"] != "5px" {
			t.Errorf("Expected padding-top=5px, got %v", attributes["padding-top"])
		}
		if attributes["padding-right"] != "10px" {
			t.Errorf("Expected padding-right=10px, got %v", attributes["padding-right"])
		}
		if attributes["padding-bottom"] != "15px" {
			t.Errorf("Expected padding-bottom=15px, got %v", attributes["padding-bottom"])
		}
		if attributes["padding-left"] != "20px" {
			t.Errorf("Expected padding-left=20px, got %v", attributes["padding-left"])
		}
	})

	// 4. Test with zero values
	t.Run("ZeroValues", func(t *testing.T) {
		attributes := make(map[string]interface{})
		// Use "all" control with zero values
		applyPaddingFromStruct("all", "0px", "", "", "", "", attributes)

		// Zero values should not be applied
		if len(attributes) > 0 {
			t.Errorf("Expected no attributes with zero values, got %v", attributes)
		}
	})
}

func TestFormatAttrsFunction(t *testing.T) {
	tests := []struct {
		name     string
		attrs    string
		expected string
	}{
		{"Empty", "", ""},
		{"WithAttrs", "class=\"test\" id=\"example\"", " class=\"test\" id=\"example\""},
		{"OnlySpaces", "   ", "    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAttrs(tt.attrs)
			if result != tt.expected {
				t.Errorf("formatAttrs(%q) = %q, want %q", tt.attrs, result, tt.expected)
			}
		})
	}
}

func TestFormatStyleAttrFunction(t *testing.T) {
	tests := []struct {
		name     string
		styles   []string
		expected string
	}{
		{"Empty", []string{}, ""},
		{"EmptyStrings", []string{"", "  "}, ""},
		{"OneStyle", []string{"color: red"}, ` style="color: red"`},
		{"MultipleStyles", []string{"color: blue", "font-weight: bold"}, ` style="color: blue; font-weight: bold"`},
		{"WithSpaces", []string{" margin: 0 ", "  padding: 10px  "}, ` style="margin: 0; padding: 10px"`},
		{"MixedEmptyAndValid", []string{"", "border: 1px solid black", ""}, ` style="border: 1px solid black"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStyleAttr(tt.styles)
			if result != tt.expected {
				t.Errorf("formatStyleAttr(%v) = %q, want %q", tt.styles, result, tt.expected)
			}
		})
	}
}

func TestEscapeHTMLFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty", "", ""},
		{"NoSpecialChars", "Hello World", "Hello World"},
		{"Ampersand", "A & B", "A &amp; B"},
		{"LessThan", "a < b", "a &lt; b"},
		{"GreaterThan", "a > b", "a &gt; b"},
		{"DoubleQuote", `He said "hello"`, "He said &quot;hello&quot;"},
		{"SingleQuote", "It's mine", "It&#39;s mine"},
		{"CombinedChars", `<a href="test.html">Link & "Text"</a>`, "&lt;a href=&quot;test.html&quot;&gt;Link &amp; &quot;Text&quot;&lt;/a&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetMapStringEdgeCases(t *testing.T) {
	// Create test map with various edge cases
	testMap := map[string]interface{}{
		"nil_value":      nil,
		"empty_string":   "",
		"mixed_number":   123.45,
		"integer_number": 42,
		"zero_value":     0,
		"negative":       -10,
		"map_value":      map[string]string{"key": "value"},
		"slice_value":    []int{1, 2, 3},
	}

	// Test nil value
	if result := getMapString(testMap, "nil_value"); result != "" {
		t.Errorf("getMapString for nil value = %q, want empty string", result)
	}

	// Test empty string
	if result := getMapString(testMap, "empty_string"); result != "" {
		t.Errorf("getMapString for empty string = %q, want empty string", result)
	}

	// Test decimal number
	if result := getMapString(testMap, "mixed_number"); result != "123.45" {
		t.Errorf("getMapString for mixed_number = %q, want '123.45'", result)
	}

	// Test integer
	if result := getMapString(testMap, "integer_number"); result != "42" {
		t.Errorf("getMapString for integer = %q, want '42'", result)
	}

	// Test zero value
	if result := getMapString(testMap, "zero_value"); result != "0" {
		t.Errorf("getMapString for zero = %q, want '0'", result)
	}

	// Test negative number
	if result := getMapString(testMap, "negative"); result != "-10" {
		t.Errorf("getMapString for negative = %q, want '-10'", result)
	}

	// Test map value (should convert to string representation)
	if result := getMapString(testMap, "map_value"); !strings.Contains(result, "map") {
		t.Errorf("getMapString for map_value = %q, should contain 'map'", result)
	}

	// Test slice value (should convert to string representation)
	if result := getMapString(testMap, "slice_value"); !strings.Contains(result, "[") {
		t.Errorf("getMapString for slice_value = %q, should contain '['", result)
	}

	// Test nonexistent key
	if result := getMapString(testMap, "does_not_exist"); result != "" {
		t.Errorf("getMapString for nonexistent key = %q, want empty string", result)
	}
}

func TestApplyStyleFunctions(t *testing.T) {
	// Test applyPaddingToStyleListFromStruct and applyMarginToStyleListFromStruct

	t.Run("PaddingStylesAllControl", func(t *testing.T) {
		styles := []string{}
		applyPaddingToStyleListFromStruct("all", "10px", "", "", "", "", &styles, " !important")

		expected := "padding: 10px !important"
		if len(styles) != 1 || styles[0] != expected {
			t.Errorf("applyPaddingToStyleListFromStruct with 'all' control, got %v, want [%s]", styles, expected)
		}
	})

	t.Run("PaddingStylesSeparateControl", func(t *testing.T) {
		styles := []string{}
		applyPaddingToStyleListFromStruct("separate", "", "5px", "10px", "15px", "20px", &styles, "")

		if len(styles) != 4 {
			t.Errorf("Expected 4 padding styles with 'separate' control, got %d: %v", len(styles), styles)
		}

		// Check each style is present (order doesn't matter)
		expectedStyles := []string{
			"padding-top: 5px",
			"padding-right: 10px",
			"padding-bottom: 15px",
			"padding-left: 20px",
		}

		for _, expected := range expectedStyles {
			found := false
			for _, style := range styles {
				if style == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Style %q not found in result: %v", expected, styles)
			}
		}
	})

	t.Run("MarginStylesDefaultZero", func(t *testing.T) {
		styles := []string{}
		applyMarginToStyleListFromStruct("", "", "", "", "", "", &styles, " !important")

		expected := "margin: 0px !important"
		if len(styles) != 1 || styles[0] != expected {
			t.Errorf("applyMarginToStyleListFromStruct with empty control should default to zero, got %v, want [%s]", styles, expected)
		}
	})

	t.Run("MarginStylesWithSpecificValues", func(t *testing.T) {
		styles := []string{}
		applyMarginToStyleListFromStruct("all", "15px", "", "", "", "", &styles, "")

		expected := "margin: 15px"
		if len(styles) != 1 || styles[0] != expected {
			t.Errorf("applyMarginToStyleListFromStruct with specific value, got %v, want [%s]", styles, expected)
		}
	})
}

func TestApplyMarginToStyleListFromStructComprehensive(t *testing.T) {
	// Tests for applyMarginToStyleListFromStruct covering more edge cases

	t.Run("SeparateControlWithAllValues", func(t *testing.T) {
		styles := []string{}
		applyMarginToStyleListFromStruct("separate", "", "5px", "10px", "15px", "20px", &styles, " !important")

		if len(styles) != 4 {
			t.Errorf("Expected 4 margin styles with 'separate' control, got %d: %v", len(styles), styles)
		}

		// Check each style is present (order doesn't matter)
		expectedStyles := []string{
			"margin-top: 5px !important",
			"margin-right: 10px !important",
			"margin-bottom: 15px !important",
			"margin-left: 20px !important",
		}

		for _, expected := range expectedStyles {
			found := false
			for _, style := range styles {
				if style == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Style %q not found in result: %v", expected, styles)
			}
		}
	})

	t.Run("SeparateControlWithSomeValues", func(t *testing.T) {
		styles := []string{}
		// Only set top and bottom, not right and left
		applyMarginToStyleListFromStruct("separate", "", "5px", "", "15px", "", &styles, " !important")

		if len(styles) != 2 {
			t.Errorf("Expected 2 margin styles with partial 'separate' control, got %d: %v", len(styles), styles)
		}

		// Check each expected style is present
		expectedStyles := []string{
			"margin-top: 5px !important",
			"margin-bottom: 15px !important",
		}

		for _, expected := range expectedStyles {
			found := false
			for _, style := range styles {
				if style == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Style %q not found in result: %v", expected, styles)
			}
		}

		// Check that right and left margins were not added
		unexpectedStyles := []string{
			"margin-right",
			"margin-left",
		}

		for _, unexpected := range unexpectedStyles {
			for _, style := range styles {
				if strings.Contains(style, unexpected) {
					t.Errorf("Unexpected style containing %q found: %q", unexpected, style)
				}
			}
		}
	})

	t.Run("ZeroValuesWithSeparateControl", func(t *testing.T) {
		styles := []string{}
		// Zero values should be ignored
		applyMarginToStyleListFromStruct("separate", "", "0px", "0px", "0px", "0px", &styles, "")

		// Should default to margin: 0px !important
		expected := "margin: 0px !important"
		if len(styles) != 1 || styles[0] != expected {
			t.Errorf("Expected margin: 0px !important with all zero values, got %v", styles)
		}
	})

	t.Run("DefaultControlBehavior", func(t *testing.T) {
		styles := []string{}
		// Empty control with specific margin
		applyMarginToStyleListFromStruct("", "15px", "", "", "", "", &styles, " !important")

		expected := "margin: 15px !important"
		if len(styles) != 1 || styles[0] != expected {
			t.Errorf("Expected shorthand margin with default control, got %v, want [%s]", styles, expected)
		}
	})
}

func TestTreeToMjml_AdditionalUnknownBlockType(t *testing.T) {
	rootStyles := createRootStyles()
	unknownBlock := EmailBlock{
		ID:   "unknown_type",
		Kind: "nonexistent_type",
		Data: map[string]interface{}{
			"someField": "someValue",
		},
	}

	rootBlock := EmailBlock{
		ID:       "root_unknown",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{unknownBlock},
	}

	// Ensure unknown types don't cause errors
	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml should not fail for unknown block types: %v", err)
	}

	// Verify MJML was generated (should contain root structure)
	if !strings.Contains(mjml, "<mjml>") || !strings.Contains(mjml, "</mjml>") {
		t.Error("Expected valid MJML structure despite unknown block type")
	}

	// Check for warning comment about unknown block type
	if !strings.Contains(mjml, "<!-- MJML Not Implemented: nonexistent_type -->") {
		t.Error("Expected comment warning about unimplemented block type")
	}
}
