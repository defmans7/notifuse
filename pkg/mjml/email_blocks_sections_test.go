package mjml

import (
	"strings"
	"testing"
)

// --- Test Helper Functions ---

func TestTreeToMjml_SectionBackground(t *testing.T) {
	rootStyles := createRootStyles()
	textBlock := createTextBlock("txt_bg", "Content in section with background")

	columnBlock := EmailBlock{
		ID:       "col_bg",
		Kind:     "column",
		Children: []EmailBlock{textBlock},
	}

	sectionBlock := EmailBlock{
		ID:   "sec_bg",
		Kind: "oneColumn", // Use oneColumn, but set background properties
		Data: map[string]interface{}{ // SectionBlockData structure
			"backgroundType": "image", // Set background type to image
			"styles": map[string]interface{}{ // Styles relevant to background image
				"backgroundImage":  "http://example.com/background.jpg",
				"backgroundSize":   "cover",
				"backgroundRepeat": "no-repeat",
				"padding":          "20px", // Add some padding for visual context
			},
			"paddingControl": "all",
		},
		Children: []EmailBlock{columnBlock},
	}

	rootBlock := EmailBlock{
		ID:       "root_bg",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{sectionBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for section background test: %v", err)
	}

	// Check for section tag
	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Expected output to contain <mj-section>")
	}

	// Check background attributes
	if !strings.Contains(mjml, `background-url="http://example.com/background.jpg"`) {
		t.Error("Missing section background-url attribute")
	}
	if !strings.Contains(mjml, `background-size="cover"`) {
		t.Error("Missing section background-size attribute")
	}
	if !strings.Contains(mjml, `background-repeat="no-repeat"`) {
		t.Error("Missing section background-repeat attribute")
	}

	// Check other attributes (e.g., padding)
	if !strings.Contains(mjml, `padding="20px"`) {
		t.Error("Missing section padding attribute")
	}

	// Check for content
	if !strings.Contains(mjml, "Content in section with background") {
		t.Error("Missing section content")
	}

	// t.Logf("Section Background MJML:\n%s", mjml)
}

func TestTreeToMjml_SectionBackgroundColor(t *testing.T) {
	rootStyles := createRootStyles()
	textBlock := createTextBlock("txt_bg_color", "Content in section with bg color")
	columnBlock := EmailBlock{ID: "col_bg_color", Kind: "column", Children: []EmailBlock{textBlock}}

	sectionBlock := EmailBlock{
		ID:   "sec_bg_color",
		Kind: "oneColumn",
		Data: map[string]interface{}{ // SectionBlockData
			"backgroundType": "color", // Set background type to color
			"styles": map[string]interface{}{ // Styles
				"backgroundColor": "#abcdef",
			},
		},
		Children: []EmailBlock{columnBlock},
	}

	rootBlock := EmailBlock{ID: "root_bg_color", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for section background color test: %v", err)
	}

	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Missing <mj-section>")
	}
	if !strings.Contains(mjml, `background-color="#abcdef"`) {
		t.Error("Missing section background-color attribute")
	}
	if !strings.Contains(mjml, "Content in section with bg color") {
		t.Error("Missing section content")
	}
	// t.Logf("Section BG Color MJML:\n%s", mjml)
}

func TestTreeToMjml_SectionColumnsOnMobile(t *testing.T) {
	rootStyles := createRootStyles()
	textBlockLeft := createTextBlock("txt_mobile_left", "Left")
	textBlockRight := createTextBlock("txt_mobile_right", "Right")
	columnLeft := EmailBlock{ID: "col_mobile_left", Kind: "column", Children: []EmailBlock{textBlockLeft}}
	columnRight := EmailBlock{ID: "col_mobile_right", Kind: "column", Children: []EmailBlock{textBlockRight}}

	sectionBlock := EmailBlock{
		ID:   "sec_mobile",
		Kind: "columns1212", // Use a multi-column layout
		Data: map[string]interface{}{ // SectionBlockData
			"columnsOnMobile": true, // <<< Enable column stacking via mj-group
			"styles":          map[string]interface{}{},
		},
		Children: []EmailBlock{columnLeft, columnRight},
	}

	rootBlock := EmailBlock{
		ID:       "root_mobile",
		Kind:     "root",
		Data:     map[string]interface{}{"styles": rootStyles},
		Children: []EmailBlock{sectionBlock},
	}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for columnsOnMobile test: %v", err)
	}

	// Check for mj-section
	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Missing <mj-section>")
	}
	// Check for mj-group wrapping the columns
	if !strings.Contains(mjml, "<mj-group>") {
		t.Error("Missing <mj-group> tag for columnsOnMobile=true")
	}
	// Check that columns are inside mj-group (less brittle check)
	cleanedMjml := strings.ReplaceAll(strings.ReplaceAll(mjml, "\n", ""), " ", "")
	if !strings.Contains(cleanedMjml, "<mj-group><mj-column") {
		t.Errorf("Expected <mj-column> to be nested within <mj-group> after cleaning whitespace. Cleaned MJML: %s", cleanedMjml)
	}
	if strings.Count(mjml, "<mj-column") != 2 {
		t.Errorf("Expected 2 <mj-column> tags, found %d", strings.Count(mjml, "<mj-column"))
	}

	// t.Logf("ColumnsOnMobile MJML:\n%s", mjml)
}

func TestTreeToMjml_SectionBorders(t *testing.T) {
	rootStyles := createRootStyles()
	textBlock := createTextBlock("txt_border", "Content in section with border")
	columnBlock := EmailBlock{ID: "col_border", Kind: "column", Children: []EmailBlock{textBlock}}

	sectionBlock := EmailBlock{
		ID:   "sec_border",
		Kind: "oneColumn",
		Data: map[string]interface{}{ // SectionBlockData
			"borderControl": "all", // Use shorthand border
			"styles": map[string]interface{}{ // Styles
				"borderStyle":  "dashed",
				"borderWidth":  "3px",
				"borderColor":  "#008800",
				"borderRadius": "10px", // Test border radius
			},
		},
		Children: []EmailBlock{columnBlock},
	}

	rootBlock := EmailBlock{ID: "root_border", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []EmailBlock{sectionBlock}}

	trackingSettings := TrackingSettings{}
	mjml, err := TreeToMjml(rootStyles, rootBlock, "", trackingSettings, 0, nil)
	if err != nil {
		t.Fatalf("TreeToMjml failed unexpectedly for section border test: %v", err)
	}

	if !strings.Contains(mjml, "<mj-section") {
		t.Error("Missing <mj-section>")
	}
	// Check for shorthand border attribute
	if !strings.Contains(mjml, `border="3px dashed #008800"`) {
		t.Error("Missing section border attribute")
	}
	// Check for border-radius attribute
	if !strings.Contains(mjml, `border-radius="10px"`) {
		t.Error("Missing section border-radius attribute")
	}
	if !strings.Contains(mjml, "Content in section with border") {
		t.Error("Missing section content")
	}
	// t.Logf("Section Border MJML:\n%s", mjml)
}
