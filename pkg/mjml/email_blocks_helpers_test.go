package mjml

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
)

// --- Tests for Helper Functions ---

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
		{"HTTPRequest", "http-request"}, // Consecutive capitals
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
		{"withFragment", "http://example.com/page#section1", "http://example.com/page?utm_campaign=test_campaign&utm_medium=test_medium&utm_source=test_source#section1"}, // Preserve fragment
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

func TestIndentPad(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{"zero spaces", 0, ""},
		{"one space", 1, " "},
		{"four spaces", 4, "    "},
		{"negative spaces", -1, ""}, // Based on strings.Repeat behavior
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indentPad(tt.n); got != tt.expected {
				t.Errorf("indentPad() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestApplyPaddingToStyleListFromStruct(t *testing.T) {
	tests := []struct {
		name           string
		paddingControl string
		padding        string
		paddingTop     string
		paddingRight   string
		paddingBottom  string
		paddingLeft    string
		initialStyles  []string
		expectedStyles []string // Expect styles sorted for comparison
	}{
		{
			name:           "all control with value",
			paddingControl: "all",
			padding:        "10px !important",
			initialStyles:  []string{"color: red"},
			expectedStyles: []string{"color: red", "padding: 10px !important"},
		},
		{
			name:           "separate control some values",
			paddingControl: "separate",
			paddingTop:     "5px !important",
			paddingBottom:  "15px !important",
			initialStyles:  []string{"font-weight: bold"},
			expectedStyles: []string{"font-weight: bold", "padding-bottom: 15px !important", "padding-top: 5px !important"},
		},
		{
			name:           "separate control zero value",
			paddingControl: "separate",
			paddingTop:     "0px !important", // Should be ignored
			paddingLeft:    "20px !important",
			initialStyles:  []string{},
			expectedStyles: []string{"padding-left: 20px !important"},
		},
		{
			name:           "no control uses shorthand",
			paddingControl: "",
			padding:        "5px 10px !important",
			initialStyles:  []string{},
			expectedStyles: []string{"padding: 5px 10px !important"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualStyles := make([]string, len(tt.initialStyles))
			copy(actualStyles, tt.initialStyles)

			// Note: The function takes a pointer to the slice
			applyPaddingToStyleListFromStruct(tt.paddingControl, tt.padding, tt.paddingTop, tt.paddingRight, tt.paddingBottom, tt.paddingLeft, &actualStyles, "") // Suffix added in test data

			// Sort slices for consistent comparison
			sort.Strings(actualStyles)
			sort.Strings(tt.expectedStyles)

			if !equalStringSlices(actualStyles, tt.expectedStyles) {
				t.Errorf("applyPaddingToStyleListFromStruct() got %v, want %v", actualStyles, tt.expectedStyles)
			}
		})
	}
}

func TestApplyMarginToStyleListFromStruct(t *testing.T) {
	tests := []struct {
		name           string
		marginControl  string
		margin         string
		marginTop      string
		marginRight    string
		marginBottom   string
		marginLeft     string
		initialStyles  []string
		expectedStyles []string // Expect styles sorted for comparison
	}{
		{
			name:           "all control with value",
			marginControl:  "all",
			margin:         "10px !important",
			initialStyles:  []string{"color: blue"},
			expectedStyles: []string{"color: blue", "margin: 10px !important"},
		},
		{
			name:           "separate control some values",
			marginControl:  "separate",
			marginTop:      "5px !important",
			marginBottom:   "15px !important",
			initialStyles:  []string{"padding: 0"},
			expectedStyles: []string{"margin-bottom: 15px !important", "margin-top: 5px !important", "padding: 0"},
		},
		{
			name:           "separate control zero value",
			marginControl:  "separate",
			marginTop:      "0px !important", // Should be ignored
			marginLeft:     "20px !important",
			initialStyles:  []string{},
			expectedStyles: []string{"margin-left: 20px !important"},
		},
		{
			name:           "no control uses shorthand",
			marginControl:  "",
			margin:         "5px 10px !important",
			initialStyles:  []string{},
			expectedStyles: []string{"margin: 5px 10px !important"},
		},
		{
			name:           "no control, no shorthand -> default 0px",
			marginControl:  "",
			margin:         "",
			initialStyles:  []string{"color: green"},
			expectedStyles: []string{"color: green", "margin: 0px !important"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualStyles := make([]string, len(tt.initialStyles))
			copy(actualStyles, tt.initialStyles)

			applyMarginToStyleListFromStruct(tt.marginControl, tt.margin, tt.marginTop, tt.marginRight, tt.marginBottom, tt.marginLeft, &actualStyles, "") // Suffix added in test data

			sort.Strings(actualStyles)
			sort.Strings(tt.expectedStyles)

			if !equalStringSlices(actualStyles, tt.expectedStyles) {
				t.Errorf("applyMarginToStyleListFromStruct() got %v, want %v", actualStyles, tt.expectedStyles)
			}
		})
	}
}

// Helper function to compare string slices (order doesn't matter after sorting)
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetMapBool(t *testing.T) {
	dataMap := map[string]interface{}{
		"isTrue":   true,
		"isFalse":  false,
		"strTrue":  "true",
		"strTRUE":  "TRUE",
		"strFalse": "false",
		"strOther": "not true",
		"numOne":   1,
		"numZero":  0,
		"nilVal":   nil,
		// "missing" key is absent
	}
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"actual true", "isTrue", true},
		{"actual false", "isFalse", false},
		{"string true lower", "strTrue", true},
		{"string true upper", "strTRUE", true},
		{"string false", "strFalse", false},
		{"string other", "strOther", false},
		{"number one", "numOne", false}, // Not considered true
		{"number zero", "numZero", false},
		{"nil value", "nilVal", false},
		{"missing key", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMapBool(dataMap, tt.key); got != tt.expected {
				t.Errorf("getMapBool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetMapString(t *testing.T) {
	dataMap := map[string]interface{}{
		"strVal":    "hello",
		"intVal":    123,
		"floatVal":  45.6,
		"floatInt":  78.0,
		"boolTrue":  true,
		"boolFalse": false,
		"nilVal":    nil,
		// "missing" key is absent
	}
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"string value", "strVal", "hello"},
		{"int value", "intVal", "123"},
		{"float value", "floatVal", "45.6"},
		{"float int value", "floatInt", "78"},
		{"bool true value", "boolTrue", "true"},
		{"bool false value", "boolFalse", "false"},
		{"nil value", "nilVal", ""},
		{"missing key", "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMapString(dataMap, tt.key); got != tt.expected {
				t.Errorf("getMapString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatAttrs(t *testing.T) {
	tests := []struct {
		name     string
		attrs    string
		expected string
	}{
		{"empty attrs", "", ""},
		{"non-empty attrs", `key="value" another="one"`, ` key="value" another="one"`}, // Note leading space
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatAttrs(tt.attrs); got != tt.expected {
				t.Errorf("formatAttrs() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatStyleAttr(t *testing.T) {
	tests := []struct {
		name     string
		styles   []string
		expected string
	}{
		{"empty slice", []string{}, ""},
		{"slice with one style", []string{"color: red"}, ` style="color: red"`},
		{"slice with multiple styles", []string{"color: blue", "font-weight: bold"}, ` style="color: blue; font-weight: bold"`},
		{"slice with empty string", []string{"padding: 5px", "", "margin: 0"}, ` style="padding: 5px; margin: 0"`},
		{"slice with only empty string", []string{""}, ""},
		{"nil slice", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatStyleAttr(tt.styles); got != tt.expected {
				t.Errorf("formatStyleAttr() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no escaping needed", "Hello World", "Hello World"},
		{"less than", "<tag>", "&lt;tag&gt;"},
		{"greater than", "script > value", "script &gt; value"},
		{"ampersand", "foo & bar", "foo &amp; bar"},
		{"double quote", `alt="text"`, `alt=&quot;text&quot;`},
		{"single quote", "style='color: red;'", "style=&#39;color: red;&#39;"},
		{"mixed chars", "<a href=\"url?a=1&b=2\">'Link' & Stuff</a>", "&lt;a href=&quot;url?a=1&amp;b=2&quot;&gt;&#39;Link&#39; &amp; Stuff&lt;/a&gt;"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeHTML(tt.input); got != tt.expected {
				t.Errorf("escapeHTML() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestApplyPaddingFromStruct(t *testing.T) {
	tests := []struct {
		name           string
		paddingControl string
		padding        string
		paddingTop     string
		paddingRight   string
		paddingBottom  string
		paddingLeft    string
		initialAttrs   map[string]interface{}
		expectedAttrs  map[string]interface{}
	}{
		{
			name:           "all control with value",
			paddingControl: "all",
			padding:        "10px",
			initialAttrs:   map[string]interface{}{"padding": "0"}, // Initial padding to override
			expectedAttrs:  map[string]interface{}{"padding": "10px"},
		},
		{
			name:           "all control empty value",
			paddingControl: "all",
			padding:        "", // Empty shorthand
			initialAttrs:   map[string]interface{}{"padding": "0"},
			expectedAttrs:  map[string]interface{}{"padding": "0"}, // Should keep initial/default
		},
		{
			name:           "all control zero value",
			paddingControl: "all",
			padding:        "0px",                                    // Zero shorthand
			initialAttrs:   map[string]interface{}{"padding": "5px"}, // Initial padding
			expectedAttrs:  map[string]interface{}{"padding": "5px"}, // Should not override with 0px
		},
		{
			name:           "separate control some values",
			paddingControl: "separate",
			paddingTop:     "5px",
			paddingBottom:  "15px",
			initialAttrs:   map[string]interface{}{"padding": "10px"}, // Initial padding to remove
			expectedAttrs:  map[string]interface{}{"padding-top": "5px", "padding-bottom": "15px"},
		},
		{
			name:           "separate control all values",
			paddingControl: "separate",
			paddingTop:     "5px",
			paddingRight:   "10px",
			paddingBottom:  "15px",
			paddingLeft:    "20px",
			initialAttrs:   map[string]interface{}{},
			expectedAttrs:  map[string]interface{}{"padding-top": "5px", "padding-right": "10px", "padding-bottom": "15px", "padding-left": "20px"},
		},
		{
			name:           "separate control zero values",
			paddingControl: "separate",
			paddingTop:     "0px", // Zero values should be ignored
			paddingRight:   "10px",
			paddingBottom:  "0px",
			paddingLeft:    "20px",
			initialAttrs:   map[string]interface{}{},
			expectedAttrs:  map[string]interface{}{"padding-right": "10px", "padding-left": "20px"},
		},
		{
			name:           "no control uses shorthand",
			paddingControl: "", // Missing control
			padding:        "5px 10px",
			initialAttrs:   map[string]interface{}{},
			expectedAttrs:  map[string]interface{}{"padding": "5px 10px"},
		},
		{
			name:           "no control empty shorthand",
			paddingControl: "", // Missing control
			padding:        "",
			initialAttrs:   map[string]interface{}{},
			expectedAttrs:  map[string]interface{}{}, // Expect no padding attribute
		},
		{
			name:           "invalid control uses shorthand",
			paddingControl: "invalid", // Invalid control
			padding:        "15px",
			initialAttrs:   map[string]interface{}{},
			expectedAttrs:  map[string]interface{}{"padding": "15px"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy initialAttrs to avoid modifying it across tests
			actualAttrs := make(map[string]interface{})
			for k, v := range tt.initialAttrs {
				actualAttrs[k] = v
			}

			applyPaddingFromStruct(tt.paddingControl, tt.padding, tt.paddingTop, tt.paddingRight, tt.paddingBottom, tt.paddingLeft, actualAttrs)

			// Deep comparison of maps
			if len(actualAttrs) != len(tt.expectedAttrs) {
				t.Fatalf("applyPaddingFromStruct() resulted in map length %d, want %d. Got: %v, Want: %v", len(actualAttrs), len(tt.expectedAttrs), actualAttrs, tt.expectedAttrs)
			}
			for k, expectedV := range tt.expectedAttrs {
				if actualV, ok := actualAttrs[k]; !ok || actualV != expectedV {
					t.Errorf("applyPaddingFromStruct() map key %q = %v, want %v. Full map: %v", k, actualV, expectedV, actualAttrs)
				}
			}
		})
	}
}

func TestApplyBordersFromStruct(t *testing.T) {
	tests := []struct {
		name          string
		borderControl string
		// "all" fields
		borderStyle string
		borderWidth string
		borderColor string
		// "separate" fields
		borderTopStyle    string
		borderTopWidth    string
		borderTopColor    string
		borderRightStyle  string
		borderRightWidth  string
		borderRightColor  string
		borderBottomStyle string
		borderBottomWidth string
		borderBottomColor string
		borderLeftStyle   string
		borderLeftWidth   string
		borderLeftColor   string
		// Input/Output
		initialAttrs  map[string]interface{}
		expectedAttrs map[string]interface{}
	}{
		{
			name:          "all control solid border",
			borderControl: "all",
			borderStyle:   "solid",
			borderWidth:   "2px",
			borderColor:   "#ff0000",
			initialAttrs:  map[string]interface{}{},
			expectedAttrs: map[string]interface{}{"border": "2px solid #ff0000"},
		},
		{
			name:          "all control no border style",
			borderControl: "all",
			borderStyle:   "none", // Style 'none' should prevent border attribute
			borderWidth:   "1px",
			borderColor:   "#000000",
			initialAttrs:  map[string]interface{}{},
			expectedAttrs: map[string]interface{}{}, // Expect no border attribute
		},
		{
			name:          "all control zero width",
			borderControl: "all",
			borderStyle:   "dotted",
			borderWidth:   "0px", // Zero width should prevent border attribute
			borderColor:   "#000000",
			initialAttrs:  map[string]interface{}{},
			expectedAttrs: map[string]interface{}{}, // Expect no border attribute
		},
		{
			name:              "separate control top and bottom",
			borderControl:     "separate",
			borderTopStyle:    "dashed",
			borderTopWidth:    "1px",
			borderTopColor:    "#00ff00",
			borderBottomStyle: "solid",
			borderBottomWidth: "3px",
			borderBottomColor: "#0000ff",
			initialAttrs:      map[string]interface{}{},
			expectedAttrs:     map[string]interface{}{"border-top": "1px dashed #00ff00", "border-bottom": "3px solid #0000ff"},
		},
		{
			name:             "separate control right only with zero width",
			borderControl:    "separate",
			borderRightStyle: "solid",
			borderRightWidth: "0px", // Zero width ignored
			borderRightColor: "#ff00ff",
			initialAttrs:     map[string]interface{}{},
			expectedAttrs:    map[string]interface{}{}, // No border attributes expected
		},
		{
			name:           "separate control all sides",
			borderControl:  "separate",
			borderTopStyle: "solid", borderTopWidth: "1px", borderTopColor: "red",
			borderRightStyle: "dashed", borderRightWidth: "2px", borderRightColor: "green",
			borderBottomStyle: "dotted", borderBottomWidth: "3px", borderBottomColor: "blue",
			borderLeftStyle: "double", borderLeftWidth: "4px", borderLeftColor: "yellow",
			initialAttrs: map[string]interface{}{},
			expectedAttrs: map[string]interface{}{
				"border-top":    "1px solid red",
				"border-right":  "2px dashed green",
				"border-bottom": "3px dotted blue",
				"border-left":   "4px double yellow",
			},
		},
		{
			name:          "no control",
			borderControl: "",      // No control
			borderStyle:   "solid", // Should be ignored
			borderWidth:   "1px",
			borderColor:   "#ccc",
			initialAttrs:  map[string]interface{}{},
			expectedAttrs: map[string]interface{}{}, // No border attributes expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualAttrs := make(map[string]interface{})
			for k, v := range tt.initialAttrs {
				actualAttrs[k] = v
			}

			applyBordersFromStruct(
				actualAttrs, tt.borderControl,
				tt.borderStyle, tt.borderWidth, tt.borderColor,
				tt.borderTopStyle, tt.borderTopWidth, tt.borderTopColor,
				tt.borderRightStyle, tt.borderRightWidth, tt.borderRightColor,
				tt.borderBottomStyle, tt.borderBottomWidth, tt.borderBottomColor,
				tt.borderLeftStyle, tt.borderLeftWidth, tt.borderLeftColor,
			)

			// Deep comparison of maps
			if len(actualAttrs) != len(tt.expectedAttrs) {
				t.Fatalf("applyBordersFromStruct() resulted in map length %d, want %d. Got: %v, Want: %v", len(actualAttrs), len(tt.expectedAttrs), actualAttrs, tt.expectedAttrs)
			}
			for k, expectedV := range tt.expectedAttrs {
				if actualV, ok := actualAttrs[k]; !ok || actualV != expectedV {
					t.Errorf("applyBordersFromStruct() map key %q = %v, want %v. Full map: %v", k, actualV, expectedV, actualAttrs)
				}
			}
		})
	}
}

func TestGetStyleString(t *testing.T) {
	styleMap := map[string]interface{}{
		"stringValue":       "10px",
		"intValue":          20,
		"floatValue":        30.5,
		"floatIntegerValue": 40.0,
		"boolValueTrue":     true,
		"boolValueFalse":    false,
		"nilValue":          nil,
		// "missingValue" is intentionally missing
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"string value", "stringValue", "10px"},
		{"int value", "intValue", "20"},
		{"float value", "floatValue", "30.5"},
		{"float integer value", "floatIntegerValue", "40"}, // Should format as int
		{"bool true value", "boolValueTrue", "true"},
		{"bool false value", "boolValueFalse", "false"},
		{"nil value", "nilValue", ""},       // Should return empty for nil
		{"missing key", "missingValue", ""}, // Should return empty for missing key
	}

	// Re-use the getStyleString function definition locally for testing
	// (Alternatively, make it public or use build tags, but this is simpler for now)
	getStyleString := func(styleMap map[string]interface{}, key string) string {
		if val, ok := styleMap[key]; ok && val != nil {
			switch v := val.(type) {
			case string:
				return v
			case float64:
				if v == float64(int64(v)) {
					return strconv.FormatInt(int64(v), 10)
				}
				return strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				return strconv.Itoa(v)
			case int64:
				return strconv.FormatInt(v, 10)
			case bool:
				return strconv.FormatBool(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStyleString(styleMap, tt.key)
			if result != tt.expected {
				t.Errorf("getStyleString(styleMap, %q) = %q; want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestLineAttributes(t *testing.T) {
	tests := []struct {
		name        string
		attrs       map[string]interface{}
		expectedStr string // Expect attributes sorted alphabetically by key
	}{
		{
			name: "basic types sorted",
			attrs: map[string]interface{}{
				"width":         "100px",
				"height":        50,  // int
				"border-radius": 5.5, // float
				"align":         "center",
				"active":        true, // bool
			},
			expectedStr: `active="true" align="center" border-radius="5.5" height="50" width="100px"`,
		},
		{
			name: "filter empty string and nil",
			attrs: map[string]interface{}{
				"a": "value",
				"b": "",  // Should be filtered
				"c": nil, // Should be filtered
				"d": "end",
			},
			expectedStr: `a="value" d="end"`,
		},
		{
			name: "filter passport",
			attrs: map[string]interface{}{
				"id":       "abc",
				"passport": "xyz", // Should be filtered
				"class":    "test",
			},
			expectedStr: `class="test" id="abc"`,
		},
		{
			name:        "empty map",
			attrs:       map[string]interface{}{},
			expectedStr: "",
		},
		{
			name:        "map with only filtered values",
			attrs:       map[string]interface{}{"a": "", "b": nil, "passport": "123"},
			expectedStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lineAttributes(tt.attrs)
			if result != tt.expectedStr {
				t.Errorf("lineAttributes() = %q; want %q", result, tt.expectedStr)
			}
		})
	}
}
