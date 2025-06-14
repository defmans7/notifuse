package notifuse_mjml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertJSONToMJML(t *testing.T) {
	// Create a simple email
	email := CreateSimpleEmail()

	// Convert to MJML
	mjml := ConvertJSONToMJML(email)

	// Basic validation
	if !strings.Contains(mjml, "<mjml>") {
		t.Error("MJML should contain <mjml> tag")
	}

	if !strings.Contains(mjml, "<mj-head>") {
		t.Error("MJML should contain <mj-head> tag")
	}

	if !strings.Contains(mjml, "<mj-body") {
		t.Error("MJML should contain <mj-body> tag")
	}

	if !strings.Contains(mjml, "<mj-section") {
		t.Error("MJML should contain <mj-section> tag")
	}

	if !strings.Contains(mjml, "Welcome to our newsletter!") {
		t.Error("MJML should contain the text content")
	}

	if !strings.Contains(mjml, "Get Started") {
		t.Error("MJML should contain the button content")
	}

	t.Logf("Generated MJML:\n%s", mjml)
}

func TestCamelToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"backgroundColor", "background-color"},
		{"fontSize", "font-size"},
		{"paddingTop", "padding-top"},
		{"fullWidthBackgroundColor", "full-width-background-color"},
		{"innerBorderRadius", "inner-border-radius"},
		{"align", "align"},
		{"src", "src"},
	}

	for _, test := range tests {
		result := camelToKebab(test.input)
		if result != test.expected {
			t.Errorf("camelToKebab(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestEscapeAttributeValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello & world", "hello &amp; world"},
		{`he said "hello"`, "he said &quot;hello&quot;"},
		{"it's a test", "it&#39;s a test"},
		{"<script>", "&lt;script&gt;"},
	}

	for _, test := range tests {
		result := escapeAttributeValue(test.input)
		if result != test.expected {
			t.Errorf("escapeAttributeValue(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestFormatSingleAttribute(t *testing.T) {
	tests := []struct {
		key      string
		value    interface{}
		expected string
	}{
		{"fontSize", "14px", ` font-size="14px"`},
		{"backgroundColor", "#ff0000", ` background-color="#ff0000"`},
		{"padding", "10px", ` padding="10px"`},
		{"disabled", true, " disabled"},
		{"disabled", false, ""},
		{"width", "", ""},
	}

	for _, test := range tests {
		result := formatSingleAttribute(test.key, test.value)
		if result != test.expected {
			t.Errorf("formatSingleAttribute(%s, %v) = %q, expected %q", test.key, test.value, result, test.expected)
		}
	}
}

func TestConvertToMJMLString(t *testing.T) {
	// Test with valid email
	email := CreateSimpleEmail()
	mjml, err := ConvertToMJMLString(email)
	if err != nil {
		t.Errorf("ConvertToMJMLString failed: %v", err)
	}
	if mjml == "" {
		t.Error("ConvertToMJMLString returned empty string")
	}

	// Test with nil email
	_, err = ConvertToMJMLString(nil)
	if err == nil {
		t.Error("ConvertToMJMLString should return error for nil email")
	}
}

func TestConvertToMJMLWithOptions(t *testing.T) {
	email := CreateSimpleEmail()

	// Test with validation enabled
	options := MJMLConvertOptions{
		Validate:      true,
		PrettyPrint:   true,
		IncludeXMLTag: false,
	}

	mjml, err := ConvertToMJMLWithOptions(email, options)
	if err != nil {
		t.Errorf("ConvertToMJMLWithOptions failed: %v", err)
	}

	if !strings.Contains(mjml, "<mjml>") {
		t.Error("MJML should contain <mjml> tag")
	}

	// Test with XML tag
	options.IncludeXMLTag = true
	mjmlWithXML, err := ConvertToMJMLWithOptions(email, options)
	if err != nil {
		t.Errorf("ConvertToMJMLWithOptions with XML failed: %v", err)
	}

	if !strings.Contains(mjmlWithXML, "<?xml") {
		t.Error("MJML with XML option should contain XML declaration")
	}
}

func TestLiquidTemplatingInTextBlock(t *testing.T) {
	// Create a text block with Liquid templating
	textBlock := &MJTextBlock{
		BaseBlock: BaseBlock{
			ID:   "text-liquid",
			Type: MJMLComponentMjText,
		},
		Content: stringPtr("Hello {{ user.name }}, you have {{ notifications.count }} new notifications!"),
	}

	// Create a simple MJML structure
	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-1",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:   "body-1",
						Type: MJMLComponentMjBody,
						Children: []interface{}{
							&MJSectionBlock{
								BaseBlock: BaseBlock{
									ID:   "section-1",
									Type: MJMLComponentMjSection,
									Children: []interface{}{
										&MJColumnBlock{
											BaseBlock: BaseBlock{
												ID:       "column-1",
												Type:     MJMLComponentMjColumn,
												Children: []interface{}{textBlock},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Template data JSON
	templateData := `{
		"user": {
			"name": "John Doe"
		},
		"notifications": {
			"count": 5
		}
	}`

	// Convert with template data
	result, err := ConvertJSONToMJMLWithData(mjml, templateData)
	require.NoError(t, err)

	// Check if Liquid variables were replaced
	if !strings.Contains(result, "Hello John Doe") {
		t.Error("Expected 'Hello John Doe' in output, Liquid templating failed")
	}
	if !strings.Contains(result, "you have 5 new notifications") {
		t.Error("Expected '5 new notifications' in output, Liquid templating failed")
	}

	t.Logf("Generated MJML with Liquid:\n%s", result)
}

func TestLiquidTemplatingInButtonBlock(t *testing.T) {
	// Create a button block with Liquid templating
	buttonBlock := &MJButtonBlock{
		BaseBlock: BaseBlock{
			ID:   "button-liquid",
			Type: MJMLComponentMjButton,
		},
		Content: stringPtr("View {{ itemType | capitalize }} #{{ item.id }}"),
		Attributes: &MJButtonAttributes{
			LinkAttributes: LinkAttributes{
				Href: stringPtr("https://example.com/{{ itemType }}/{{ item.id }}"),
			},
		},
	}

	// Create a simple MJML structure
	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-1",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:   "body-1",
						Type: MJMLComponentMjBody,
						Children: []interface{}{
							&MJSectionBlock{
								BaseBlock: BaseBlock{
									ID:   "section-1",
									Type: MJMLComponentMjSection,
									Children: []interface{}{
										&MJColumnBlock{
											BaseBlock: BaseBlock{
												ID:       "column-1",
												Type:     MJMLComponentMjColumn,
												Children: []interface{}{buttonBlock},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Template data JSON
	templateData := `{
		"itemType": "order",
		"item": {
			"id": 12345
		}
	}`

	// Convert with template data
	result, err := ConvertJSONToMJMLWithData(mjml, templateData)
	require.NoError(t, err)

	// Check if Liquid variables were replaced
	if !strings.Contains(result, "View Order #12345") {
		t.Error("Expected 'View Order #12345' in output, Liquid templating failed")
	}

	t.Logf("Generated MJML with Liquid button:\n%s", result)
}

func TestLiquidTemplatingWithInvalidJSON(t *testing.T) {
	// Create a text block with Liquid templating
	textBlock := &MJTextBlock{
		BaseBlock: BaseBlock{
			ID:   "text-invalid",
			Type: MJMLComponentMjText,
		},
		Content: stringPtr("Hello {{ user.name }}!"),
	}

	// Create a simple MJML structure
	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-1",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:       "body-1",
						Type:     MJMLComponentMjBody,
						Children: []interface{}{textBlock},
					},
				},
			},
		},
	}

	// Invalid JSON template data
	invalidTemplateData := `{"user": "invalid json`

	// Convert with invalid template data - should return error
	_, err := ConvertJSONToMJMLWithData(mjml, invalidTemplateData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JSON in templateData")
}

func TestLiquidTemplatingWithEmptyData(t *testing.T) {
	// Create a text block with Liquid templating
	textBlock := &MJTextBlock{
		BaseBlock: BaseBlock{
			ID:   "text-empty",
			Type: MJMLComponentMjText,
		},
		Content: stringPtr("Hello {{ user.name | default: 'Guest' }}!"),
	}

	// Create a simple MJML structure
	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-1",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:       "body-1",
						Type:     MJMLComponentMjBody,
						Children: []interface{}{textBlock},
					},
				},
			},
		},
	}

	// Convert with empty template data
	result, err := ConvertJSONToMJMLWithData(mjml, "")
	if err != nil {
		t.Errorf("ConvertJSONToMJMLWithData failed: %v", err)
	}

	// Should use default value
	if !strings.Contains(result, "Hello Guest!") {
		t.Error("Expected 'Hello Guest!' when using default filter with empty data")
	}

	t.Logf("Generated MJML with empty data:\n%s", result)
}

func TestConvertToMJMLStringWithData(t *testing.T) {
	// Create email with Liquid content
	email := CreateSimpleEmail()

	// Modify the text content to include Liquid templating
	// We need to access the text block and modify its content
	// This is a bit complex due to the nested structure, but for testing we can create a new email

	templateData := `{"user": {"name": "Test User"}}`

	// Test the function
	mjml, err := ConvertToMJMLStringWithData(email, templateData)
	if err != nil {
		t.Errorf("ConvertToMJMLStringWithData failed: %v", err)
	}

	if mjml == "" {
		t.Error("ConvertToMJMLStringWithData returned empty string")
	}

	// Test with nil email
	_, err = ConvertToMJMLStringWithData(nil, templateData)
	if err == nil {
		t.Error("ConvertToMJMLStringWithData should return error for nil email")
	}
}

func TestProcessLiquidContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		templateData string
		expected     string
		shouldError  bool
	}{
		{
			name:         "No Liquid markup",
			content:      "Plain text content",
			templateData: `{"user": {"name": "John"}}`,
			expected:     "Plain text content",
			shouldError:  false,
		},
		{
			name:         "Simple variable replacement",
			content:      "Hello {{ user.name }}!",
			templateData: `{"user": {"name": "John"}}`,
			expected:     "Hello John!",
			shouldError:  false,
		},
		{
			name:         "Multiple variables",
			content:      "{{ greeting }} {{ user.name }}, you have {{ count }} items.",
			templateData: `{"greeting": "Hi", "user": {"name": "Jane"}, "count": 3}`,
			expected:     "Hi Jane, you have 3 items.",
			shouldError:  false,
		},
		{
			name:         "With filter",
			content:      "Hello {{ user.name | upcase }}!",
			templateData: `{"user": {"name": "john"}}`,
			expected:     "Hello JOHN!",
			shouldError:  false,
		},
		{
			name:         "Empty template data",
			content:      "Hello {{ user.name | default: 'Guest' }}!",
			templateData: "",
			expected:     "Hello Guest!",
			shouldError:  false,
		},
		{
			name:         "Invalid JSON",
			content:      "Hello {{ user.name }}!",
			templateData: `{"invalid": json}`,
			expected:     "Hello {{ user.name }}!",
			shouldError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := processLiquidContent(test.content, test.templateData, "test-block")

			if test.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !test.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestLiquidTemplatingInRawBlock(t *testing.T) {
	// Create an mj-raw block with Liquid templating that outputs MJML
	rawBlock := &MJRawBlock{
		BaseBlock: BaseBlock{
			ID:   "raw-liquid",
			Type: MJMLComponentMjRaw,
		},
		Content: stringPtr(`{% if user.showButton %}<mj-button href="{{ user.buttonUrl }}">{{ user.buttonText }}</mj-button>{% endif %}`),
	}

	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-raw-test",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:   "body-raw-test",
						Type: MJMLComponentMjBody,
						Children: []interface{}{
							&MJSectionBlock{
								BaseBlock: BaseBlock{
									ID:   "section-raw-test",
									Type: MJMLComponentMjSection,
									Children: []interface{}{
										&MJColumnBlock{
											BaseBlock: BaseBlock{
												ID:       "column-raw-test",
												Type:     MJMLComponentMjColumn,
												Children: []interface{}{rawBlock},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Template data that enables the button
	templateData := `{
		"user": {
			"showButton": true,
			"buttonUrl": "https://example.com/action",
			"buttonText": "Click Me"
		}
	}`

	// Convert with template data
	result, err := ConvertJSONToMJMLWithData(mjml, templateData)
	require.NoError(t, err)

	// Check that Liquid was processed and MJML button was generated
	expectedContains := []string{
		"<mj-button href=\"https://example.com/action\">Click Me</mj-button>",
		"<mj-raw>",
		"</mj-raw>",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain %q, but it didn't. Result: %s", expected, result)
		}
	}

	t.Logf("Generated MJML with Liquid raw block:\n%s", result)
}

func TestLiquidTemplatingInRawBlockWithDisabledCondition(t *testing.T) {
	// Create an mj-raw block with Liquid templating
	rawBlock := &MJRawBlock{
		BaseBlock: BaseBlock{
			ID:   "raw-liquid-disabled",
			Type: MJMLComponentMjRaw,
		},
		Content: stringPtr(`{% if user.showButton %}<mj-button href="{{ user.buttonUrl }}">{{ user.buttonText }}</mj-button>{% endif %}`),
	}

	mjml := &MJMLBlock{
		BaseBlock: BaseBlock{
			ID:   "mjml-raw-disabled",
			Type: MJMLComponentMjml,
			Children: []interface{}{
				&MJBodyBlock{
					BaseBlock: BaseBlock{
						ID:       "body-raw-disabled",
						Type:     MJMLComponentMjBody,
						Children: []interface{}{rawBlock},
					},
				},
			},
		},
	}

	// Template data that disables the button
	templateData := `{
		"user": {
			"showButton": false,
			"buttonUrl": "https://example.com/action",
			"buttonText": "Click Me"
		}
	}`

	// Convert with template data
	result, err := ConvertJSONToMJMLWithData(mjml, templateData)
	require.NoError(t, err)

	// Check that the button was not generated (condition was false)
	if strings.Contains(result, "<mj-button") {
		t.Errorf("Expected button not to be generated when condition is false, but found mj-button in result: %s", result)
	}

	// Should still have the mj-raw wrapper but empty content
	if !strings.Contains(result, "<mj-raw></mj-raw>") {
		t.Errorf("Expected empty mj-raw block, but got: %s", result)
	}

	t.Logf("Generated MJML with disabled condition:\n%s", result)
}
