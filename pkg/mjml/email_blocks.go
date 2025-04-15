package mjml

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/osteele/liquid"
)

// indentPad returns a string of n spaces
func indentPad(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// trackURL adds UTM parameters to a URL string.
func trackURL(urlString string, urlParams map[string]string) string {
	// Ignore if URL is empty, a placeholder, mailto:, tel:, or already tracked (basic check)
	if urlString == "" || strings.Contains(urlString, "{{") || strings.Contains(urlString, "{%") || strings.HasPrefix(urlString, "mailto:") || strings.HasPrefix(urlString, "tel:") || strings.Contains(urlString, "utm_source=") {
		return urlString
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Warning: Could not parse URL for tracking, returning original: %s, error: %v", urlString, err)
		return urlString // Return original URL if parsing fails (e.g., relative URL)
	}

	// Only proceed if scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return urlString
	}

	query := parsedURL.Query()
	if value, ok := urlParams["utm_source"]; ok && value != "" && !query.Has("utm_source") {
		query.Add("utm_source", value)
	}
	if value, ok := urlParams["utm_medium"]; ok && value != "" && !query.Has("utm_medium") {
		query.Add("utm_medium", value)
	}
	if value, ok := urlParams["utm_campaign"]; ok && value != "" && !query.Has("utm_campaign") {
		query.Add("utm_campaign", value)
	}
	if value, ok := urlParams["utm_content"]; ok && value != "" && !query.Has("utm_content") {
		query.Add("utm_content", value)
	}
	if value, ok := urlParams["utm_id"]; ok && value != "" && !query.Has("utm_id") {
		query.Add("utm_id", value)
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// TagConversion maps custom tags to MJML tags (if any were used - seems unused in the converted code)
var TagConversion = map[string]string{
	"mj-dev": "mj-raw",
}

// lineAttributes converts a map of attributes to a sorted string of key="value" pairs,
// excluding the "passport" attribute and empty/nil values.
func lineAttributes(attrs map[string]interface{}) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		// Exclude passport and filter out nil or empty string values
		if k != "passport" && attrs[k] != nil {
			if strVal, ok := attrs[k].(string); !ok || strVal != "" {
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		// Format based on type for better representation (e.g., bool as string)
		switch v := attrs[k].(type) {
		case bool:
			pairs = append(pairs, fmt.Sprintf(`%s="%t"`, k, v))
		case int, int64, float64: // Handle common numeric types
			pairs = append(pairs, fmt.Sprintf(`%s="%v"`, k, v))
		default:
			pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, fmt.Sprintf("%v", v))) // Default to %v, ensure strings are quoted
		}
	}
	return strings.Join(pairs, " ")
}

// --- New robust toKebabCase function ---
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// toKebabCase converts a string to kebab-case.
// Handles CamelCase, camelCase, and potential acronyms (e.g., URL -> url).
func toKebabCase(str string) string {
	kebab := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	kebab = matchAllCap.ReplaceAllString(kebab, "${1}-${2}")
	return strings.ToLower(kebab)
}

// --- End toKebabCase ---

// EmailBlock represents the structure received from the frontend or stored.
// It mirrors the BlockInterface in TypeScript.
type EmailBlock struct {
	ID       string       `json:"id"`
	Kind     string       `json:"kind"`
	Path     string       `json:"path"` // Path might not be directly used in MJML generation
	Children []EmailBlock `json:"children"`
	// Data holds the specific block data. Use type assertion based on Kind.
	// Often unmarshalled as map[string]interface{} if not explicitly typed beforehand.
	Data interface{} `json:"data"`
}

// GetBlockData returns the block data cast to the appropriate type based on the Kind
func (n *EmailBlock) GetBlockData() any {
	switch n.Kind {
	case "button":
		return n.Data.(ButtonBlockData)
	case "image":
		return n.Data.(ImageBlockData)
	case "column":
		return n.Data.(ColumnBlockData)
	case "divider":
		return n.Data.(DividerBlockData)
	case "section":
		return n.Data.(SectionBlockData)
	case "openTracking":
		return n.Data.(OpenTrackingBlockData)
	case "text":
		return n.Data.(TextBlockData)
	default:
		return n.Data
	}
}

// BaseStyles represents common style properties used across multiple blocks
type BaseStyles struct {
	PaddingTop    string `json:"paddingTop,omitempty"`
	PaddingRight  string `json:"paddingRight,omitempty"`
	PaddingBottom string `json:"paddingBottom,omitempty"`
	PaddingLeft   string `json:"paddingLeft,omitempty"`
	Padding       string `json:"padding,omitempty"`
	// Border fields for borderControl="all"
	BorderStyle string `json:"borderStyle,omitempty"`
	BorderWidth string `json:"borderWidth,omitempty"`
	BorderColor string `json:"borderColor,omitempty"`
	// Separate border fields for borderControl="separate"
	BorderTopStyle    string `json:"borderTopStyle,omitempty"`
	BorderTopWidth    string `json:"borderTopWidth,omitempty"`
	BorderTopColor    string `json:"borderTopColor,omitempty"`
	BorderRightStyle  string `json:"borderRightStyle,omitempty"`
	BorderRightWidth  string `json:"borderRightWidth,omitempty"`
	BorderRightColor  string `json:"borderRightColor,omitempty"`
	BorderBottomStyle string `json:"borderBottomStyle,omitempty"`
	BorderBottomWidth string `json:"borderBottomWidth,omitempty"`
	BorderBottomColor string `json:"borderBottomColor,omitempty"`
	BorderLeftStyle   string `json:"borderLeftStyle,omitempty"`
	BorderLeftWidth   string `json:"borderLeftWidth,omitempty"`
	BorderLeftColor   string `json:"borderLeftColor,omitempty"`
	// Shared border property
	BorderRadius string `json:"borderRadius,omitempty"`
}

// WrapperStyles represents common wrapper properties for blocks
type WrapperStyles struct {
	Align          string `json:"align"`
	PaddingControl string `json:"paddingControl"` // "all" or "separate"
	Padding        string `json:"padding,omitempty"`
	PaddingTop     string `json:"paddingTop,omitempty"`
	PaddingRight   string `json:"paddingRight,omitempty"`
	PaddingBottom  string `json:"paddingBottom,omitempty"`
	PaddingLeft    string `json:"paddingLeft,omitempty"`
	// Added Border fields
	BorderControl string `json:"borderControl,omitempty"` // "all" or "separate"
	// Border fields for borderControl="all"
	BorderStyle string `json:"borderStyle,omitempty"`
	BorderWidth string `json:"borderWidth,omitempty"`
	BorderColor string `json:"borderColor,omitempty"`
	// Separate border fields for borderControl="separate"
	BorderTopStyle    string `json:"borderTopStyle,omitempty"`
	BorderTopWidth    string `json:"borderTopWidth,omitempty"`
	BorderTopColor    string `json:"borderTopColor,omitempty"`
	BorderRightStyle  string `json:"borderRightStyle,omitempty"`
	BorderRightWidth  string `json:"borderRightWidth,omitempty"`
	BorderRightColor  string `json:"borderRightColor,omitempty"`
	BorderBottomStyle string `json:"borderBottomStyle,omitempty"`
	BorderBottomWidth string `json:"borderBottomWidth,omitempty"`
	BorderBottomColor string `json:"borderBottomColor,omitempty"`
	BorderLeftStyle   string `json:"borderLeftStyle,omitempty"`
	BorderLeftWidth   string `json:"borderLeftWidth,omitempty"`
	BorderLeftColor   string `json:"borderLeftColor,omitempty"`
	// Shared border property
	BorderRadius string `json:"borderRadius,omitempty"`
}

// ButtonBlockData represents the data structure for a button block
type ButtonBlockData struct {
	Button struct {
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
		BorderControl          string `json:"borderControl"` // "all" or "separate"
		// Added Border fields to match original logic applyBorders(..., btnData, ...)
		// Border fields for borderControl="all"
		BorderStyle string `json:"borderStyle,omitempty"`
		BorderWidth string `json:"borderWidth,omitempty"`
		BorderColor string `json:"borderColor,omitempty"`
		// Separate border fields for borderControl="separate"
		BorderTopStyle    string `json:"borderTopStyle,omitempty"`
		BorderTopWidth    string `json:"borderTopWidth,omitempty"`
		BorderTopColor    string `json:"borderTopColor,omitempty"`
		BorderRightStyle  string `json:"borderRightStyle,omitempty"`
		BorderRightWidth  string `json:"borderRightWidth,omitempty"`
		BorderRightColor  string `json:"borderRightColor,omitempty"`
		BorderBottomStyle string `json:"borderBottomStyle,omitempty"`
		BorderBottomWidth string `json:"borderBottomWidth,omitempty"`
		BorderBottomColor string `json:"borderBottomColor,omitempty"`
		BorderLeftStyle   string `json:"borderLeftStyle,omitempty"`
		BorderLeftWidth   string `json:"borderLeftWidth,omitempty"`
		BorderLeftColor   string `json:"borderLeftColor,omitempty"`
	} `json:"button"`
	Wrapper WrapperStyles `json:"wrapper"`
}

// ImageBlockData represents the data structure for an image block
type ImageBlockData struct {
	Image struct {
		Src           string `json:"src"`
		Alt           string `json:"alt"`
		Href          string `json:"href"`
		Width         string `json:"width"`
		BorderControl string `json:"borderControl"` // "all" or "separate"
	} `json:"image"`
	Wrapper WrapperStyles `json:"wrapper"`
}

// ColumnBlockData represents the data structure for a column block
type ColumnBlockData struct {
	Styles struct {
		VerticalAlign   string `json:"verticalAlign"` // "top", "middle", "bottom"
		BackgroundColor string `json:"backgroundColor,omitempty"`
		MinHeight       string `json:"minHeight,omitempty"`
		BaseStyles
	} `json:"styles"`
	PaddingControl string `json:"paddingControl"` // "all" or "separate"
	BorderControl  string `json:"borderControl"`  // "all" or "separate"
}

// DividerBlockData represents the data structure for a divider block
type DividerBlockData struct {
	Align           string `json:"align"` // "left", "center", "right"
	BorderColor     string `json:"borderColor"`
	BorderStyle     string `json:"borderStyle"`
	BorderWidth     string `json:"borderWidth"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Width           string `json:"width"`
	PaddingControl  string `json:"paddingControl"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// SectionBlockData represents the data structure for a section block
type SectionBlockData struct {
	ColumnsOnMobile     bool   `json:"columnsOnMobile"`
	StackColumnsAtWidth int    `json:"stackColumnsAtWidth"`
	BackgroundType      string `json:"backgroundType"` // "color" or "image"
	PaddingControl      string `json:"paddingControl"` // "all" or "separate"
	BorderControl       string `json:"borderControl"`  // "all" or "separate"
	Styles              struct {
		TextAlign        string `json:"textAlign"`                  // "left", "center", "right", "justify"
		BackgroundRepeat string `json:"backgroundRepeat,omitempty"` // "repeat", "no-repeat", "repeat-x", "repeat-y"
		Padding          string `json:"padding,omitempty"`
		BorderWidth      string `json:"borderWidth,omitempty"`
		BorderStyle      string `json:"borderStyle,omitempty"`
		BorderColor      string `json:"borderColor,omitempty"`
		BackgroundColor  string `json:"backgroundColor,omitempty"`
		BackgroundImage  string `json:"backgroundImage,omitempty"`
		BackgroundSize   string `json:"backgroundSize,omitempty"` // "cover" or "contain"
		BaseStyles
	} `json:"styles"`
}

// OpenTrackingBlockData represents the data structure for an open tracking block
type OpenTrackingBlockData struct {
	// No specific data structure
}

// TextBlockData represents the data structure for a text block
type TextBlockData struct {
	Align           string `json:"align"` // "left", "center", "right"
	Width           string `json:"width"`
	HyperlinkStyles struct {
		Color          string `json:"color"`
		TextDecoration string `json:"textDecoration"`
		FontFamily     string `json:"fontFamily"`
		FontSize       string `json:"fontSize"`
		FontWeight     int    `json:"fontWeight"`
		FontStyle      string `json:"fontStyle"`
		TextTransform  string `json:"textTransform"`
	} `json:"hyperlinkStyles"`
	EditorData []struct {
		Type     string                   `json:"type"`
		Children []map[string]interface{} `json:"children"` // Changed to map to preserve all fields
	} `json:"editorData"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	PaddingControl  string `json:"paddingControl,omitempty"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// HeadingBlockData represents the data structure for a heading block
type HeadingBlockData struct {
	Type       string `json:"type"`  // "h1", "h2", "h3"
	Align      string `json:"align"` // "left", "center", "right"
	Width      string `json:"width"`
	EditorData []struct {
		Type     string                   `json:"type"`
		Children []map[string]interface{} `json:"children"` // Changed to map to preserve all fields
	} `json:"editorData"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	PaddingControl  string `json:"paddingControl,omitempty"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// LiquidBlockData represents the data structure for a liquid template block
type LiquidBlockData struct {
	LiquidCode string `json:"liquidCode"`
}

// ColumnLayoutBlockData represents the base data structure for column layouts
type ColumnLayoutBlockData struct {
	SectionBlockData
	Columns []int `json:"columns"`
}

// OneColumnBlockData represents the data structure for a single column layout
type OneColumnBlockData struct {
	SectionBlockData
	Columns [1]int `json:"columns"` // [24]
}

// Column variations
type Columns168BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [16, 8]
}

type Columns204BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [20, 4]
}

type Columns420BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [4, 20]
}

type Columns816BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [8, 16]
}

type Columns888BlockData struct {
	ColumnLayoutBlockData
	Columns [3]int `json:"columns"` // [8, 8, 8]
}

type Columns1212BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [12, 12]
}

type Columns6666BlockData struct {
	ColumnLayoutBlockData
	Columns [4]int `json:"columns"` // [6, 6, 6, 6]
}

// TreeToMjml converts an EmailBlock tree into an MJML string.
// rootStyles: Expected to be a map derived from RootBlockData.Styles, used for text/heading defaults.
// block: The current EmailBlock node to process.
// templateData: JSON string containing data for Liquid processing.
// urlParams: Map containing UTM parameters for URL tracking.
// indent: Current indentation level.
// parent: Pointer to the parent EmailBlock, used for context (e.g., column width calculation).
//
// Returns the MJML string and an error if one occurred during processing.
func TreeToMjml(rootStyles map[string]interface{}, block EmailBlock, templateData string, urlParams map[string]string, indent int, parent *EmailBlock) (string, error) {
	var sb strings.Builder
	space := indentPad(indent)
	tagName := ""
	attributes := make(map[string]interface{}) // Use kebab-case keys directly where known
	content := ""
	children := block.Children
	childrenMjml := "" // Stores pre-rendered children MJML (e.g., for mj-group)

	// Helper function to marshal/unmarshal block.Data into a specific struct
	unmarshalBlockData := func(target interface{}) error {
		if block.Data == nil {
			// If data is nil, nothing to unmarshal, return nil error
			// The target struct should be initialized with zero values.
			return nil
		}
		// Marshal the interface{} data back to JSON
		jsonData, err := json.Marshal(block.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal block data (ID: %s, Kind: %s): %w", block.ID, block.Kind, err)
		}
		// Unmarshal the JSON into the target struct
		err = json.Unmarshal(jsonData, target)
		if err != nil {
			return fmt.Errorf("failed to unmarshal block data into %T (ID: %s, Kind: %s): %w", target, block.ID, block.Kind, err)
		}
		return nil
	}

	// --- Helper functions for safe style extraction from rootStyles ---
	getStyleMap := func(key string) map[string]interface{} {
		if styleVal, ok := rootStyles[key]; ok {
			if styleMap, ok := styleVal.(map[string]interface{}); ok {
				return styleMap
			}
			log.Printf("Warning: rootStyle key '%s' is not a map[string]interface{}", key)
		}
		// log.Printf("Warning: rootStyle key '%s' not found", key)
		return make(map[string]interface{}) // Return empty map if not found or wrong type
	}
	// Safely get string, handling potential float64/int from JSON unmarshalling
	getStyleString := func(styleMap map[string]interface{}, key string) string {
		if val, ok := styleMap[key]; ok && val != nil {
			switch v := val.(type) {
			case string:
				return v
			case float64: // JSON numbers often unmarshal as float64
				// Format float simply, avoid scientific notation for reasonable numbers
				if v == float64(int64(v)) { // Check if it's effectively an integer
					return strconv.FormatInt(int64(v), 10)
				}
				return strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				return strconv.Itoa(v)
			case int64:
				return strconv.FormatInt(v, 10)
			case bool: // Handle boolean case if needed for some styles
				return strconv.FormatBool(v)
			default:
				// log.Printf("Warning: style key '%s' has unexpected type %T", key, v)
				return fmt.Sprintf("%v", v) // Fallback to generic format
			}
		}
		return ""
	}
	// --- End Helper functions ---

	switch block.Kind {
	case "root":
		tagName = "mjml"
		bodyStyles := getStyleMap("body") // Get styles for the body tag from the passed map

		bodyAttrs := make(map[string]interface{}) // Build body attributes map
		bodyAttrs["width"] = getStyleString(bodyStyles, "width")
		bodyAttrs["background-color"] = getStyleString(bodyStyles, "backgroundColor")
		// Note: 'margin' is explicitly deleted in TS, so we don't add it. Filter in lineAttributes handles empty values.

		bodyAttrsStr := lineAttributes(bodyAttrs)
		bodySpace := indentPad(indent + 2)

		var bodyChildrenSb strings.Builder
		for _, child := range children {
			childMjml, err := TreeToMjml(rootStyles, child, templateData, urlParams, indent+4, &block) // Pass current block as parent
			if err != nil {
				return "", fmt.Errorf("error processing child block (ID: %s, Kind: %s): %w", child.ID, child.Kind, err)
			}
			if strings.TrimSpace(childMjml) != "" {
				bodyChildrenSb.WriteString(childMjml)
				bodyChildrenSb.WriteString("\n") // Newline between children
			}
		}
		bodyChildren := strings.TrimSuffix(bodyChildrenSb.String(), "\n") // Remove trailing newline

		// Assemble root MJML structure
		fmt.Fprintf(&sb, "%s<mjml>\n%s<mj-body%s>\n%s\n%s</mj-body>\n%s</mjml>",
			space, bodySpace, formatAttrs(bodyAttrsStr), bodyChildren, bodySpace, space)
		return sb.String(), nil // Return immediately for root

	case "columns168", "columns204", "columns420", "columns816", "columns888", "columns1212", "columns6666", "oneColumn":
		var sectionData ColumnLayoutBlockData // Use the base struct for common fields
		if err := unmarshalBlockData(&sectionData); err != nil {
			return "", err
		}

		tagName = "mj-section"
		sectionAttrs := make(map[string]interface{}) // Use kebab-case keys directly
		sectionAttrs["text-align"] = sectionData.Styles.TextAlign

		// Background
		if sectionData.BackgroundType != "" {
			if sectionData.BackgroundType == "image" {
				sectionAttrs["background-url"] = sectionData.Styles.BackgroundImage
				if sectionData.Styles.BackgroundSize != "" {
					sectionAttrs["background-size"] = sectionData.Styles.BackgroundSize
				}
				if sectionData.Styles.BackgroundRepeat != "" {
					sectionAttrs["background-repeat"] = sectionData.Styles.BackgroundRepeat
				}
			} else if sectionData.BackgroundType == "color" {
				if sectionData.Styles.BackgroundColor != "" {
					sectionAttrs["background-color"] = sectionData.Styles.BackgroundColor
				}
			}
		}

		// Border
		applyBordersFromStruct(sectionAttrs, sectionData.BorderControl,
			sectionData.Styles.BorderStyle, sectionData.Styles.BorderWidth, sectionData.Styles.BorderColor,
			sectionData.Styles.BorderTopStyle, sectionData.Styles.BorderTopWidth, sectionData.Styles.BorderTopColor,
			sectionData.Styles.BorderRightStyle, sectionData.Styles.BorderRightWidth, sectionData.Styles.BorderRightColor,
			sectionData.Styles.BorderBottomStyle, sectionData.Styles.BorderBottomWidth, sectionData.Styles.BorderBottomColor,
			sectionData.Styles.BorderLeftStyle, sectionData.Styles.BorderLeftWidth, sectionData.Styles.BorderLeftColor)
		if sectionData.Styles.BorderRadius != "" && sectionData.Styles.BorderRadius != "0px" {
			sectionAttrs["border-radius"] = sectionData.Styles.BorderRadius
		}

		// Padding
		applyPaddingFromStruct(sectionData.PaddingControl, sectionData.Styles.Padding, sectionData.Styles.PaddingTop, sectionData.Styles.PaddingRight, sectionData.Styles.PaddingBottom, sectionData.Styles.PaddingLeft, sectionAttrs)

		attributes = sectionAttrs // Assign directly

		// Handle mj-group wrapping
		if sectionData.ColumnsOnMobile && len(children) > 0 {
			var groupChildrenSb strings.Builder
			for _, child := range children {
				childMjml, err := TreeToMjml(rootStyles, child, templateData, urlParams, indent+4, &block)
				if err != nil {
					return "", fmt.Errorf("error processing child block for mj-group (ID: %s, Kind: %s): %w", child.ID, child.Kind, err)
				}

				if strings.TrimSpace(childMjml) != "" {
					groupChildrenSb.WriteString(childMjml)
					groupChildrenSb.WriteString("\n")
				}
			}
			groupChildrenStr := strings.TrimSuffix(groupChildrenSb.String(), "\n")

			if strings.TrimSpace(groupChildrenStr) != "" {
				groupSpace := indentPad(indent + 2)
				childrenMjml = fmt.Sprintf("%s<mj-group>\n%s\n%s</mj-group>", groupSpace, groupChildrenStr, groupSpace)
			} else {
				childrenMjml = "" // No group needed if children render empty
			}
			children = nil // Prevent default children processing below switch
		}
		// Go to common assembly below the switch

	case "column":
		var columnData ColumnBlockData
		if err := unmarshalBlockData(&columnData); err != nil {
			return "", err
		}

		tagName = "mj-column"
		columnAttrs := make(map[string]interface{}) // Use kebab-case keys directly
		columnAttrs["vertical-align"] = columnData.Styles.VerticalAlign

		// Calculate width based on parent kind
		if parent != nil && len(parent.Children) > 0 {
			index := -1
			for i, c := range parent.Children {
				if c.ID == block.ID {
					index = i
					break
				}
			}

			if index != -1 {
				// Need to access parent's Columns data if available (requires unmarshalling parent.Data)
				// This highlights complexity - better if structure was more directly linked.
				// For now, keep the kind-based switch. A full refactor might change this.
				switch parent.Kind {
				case "columns168":
					columnAttrs["width"] = map[int]string{0: "66.66%", 1: "33.33%"}[index]
				case "columns204":
					columnAttrs["width"] = map[int]string{0: "83.33%", 1: "16.66%"}[index]
				case "columns420":
					columnAttrs["width"] = map[int]string{0: "16.66%", 1: "83.33%"}[index]
				case "columns816":
					columnAttrs["width"] = map[int]string{0: "33.33%", 1: "66.66%"}[index]
				case "columns888":
					columnAttrs["width"] = "33.33%" // Adjusted from 25% - likely 3 equal columns
				case "columns1212":
					columnAttrs["width"] = "50%"
				case "columns6666":
					columnAttrs["width"] = "25%"
					// case "oneColumn": width is implicit 100%
				}
			}
		}

		// Background
		if columnData.Styles.BackgroundColor != "" {
			columnAttrs["background-color"] = columnData.Styles.BackgroundColor
		}

		// Border
		applyBordersFromStruct(columnAttrs, columnData.BorderControl,
			columnData.Styles.BorderStyle, columnData.Styles.BorderWidth, columnData.Styles.BorderColor,
			columnData.Styles.BorderTopStyle, columnData.Styles.BorderTopWidth, columnData.Styles.BorderTopColor,
			columnData.Styles.BorderRightStyle, columnData.Styles.BorderRightWidth, columnData.Styles.BorderRightColor,
			columnData.Styles.BorderBottomStyle, columnData.Styles.BorderBottomWidth, columnData.Styles.BorderBottomColor,
			columnData.Styles.BorderLeftStyle, columnData.Styles.BorderLeftWidth, columnData.Styles.BorderLeftColor)
		if columnData.Styles.BorderRadius != "" && columnData.Styles.BorderRadius != "0px" {
			columnAttrs["border-radius"] = columnData.Styles.BorderRadius
		}

		// Padding
		applyPaddingFromStruct(columnData.PaddingControl, columnData.Styles.Padding, columnData.Styles.PaddingTop, columnData.Styles.PaddingRight, columnData.Styles.PaddingBottom, columnData.Styles.PaddingLeft, columnAttrs)

		attributes = columnAttrs // Assign directly
		// Go to common assembly below the switch

	case "text":
		var textData TextBlockData
		if err := unmarshalBlockData(&textData); err != nil {
			return "", err
		}

		tagName = "mj-text"
		textAttrs := make(map[string]interface{}) // Use kebab-case keys directly
		textAttrs["align"] = textData.Align
		textAttrs["padding"] = "0" // Override default

		if textData.BackgroundColor != "" {
			textAttrs["container-background-color"] = textData.BackgroundColor
		}

		// Padding for mj-text container using the block's padding fields
		applyPaddingFromStruct(textData.PaddingControl, textData.Padding, textData.PaddingTop, textData.PaddingRight, textData.PaddingBottom, textData.PaddingLeft, textAttrs)

		attributes = textAttrs

		// --- Content Generation ---
		var contentSb strings.Builder
		// editorData is still []struct{... map[string]interface{}}
		// Accessing hyperlinkStyles now uses textData struct.

		for _, lineRaw := range textData.EditorData { // Use textData.EditorData directly
			// lineRaw is the struct {Type string, Children []...}, not a map
			lineType := lineRaw.Type // e.g., "paragraph"
			childrenRaw := lineRaw.Children

			var lineContentSb strings.Builder
			for _, part := range childrenRaw { // part is the struct { Text string }, but might have other fields
				// We still need to treat part like a map for optional styles (bold, italic, etc.)
				// This requires reflection or another marshal/unmarshal if we want pure struct access here.
				// Sticking to map access for 'part' data for now, as EditorData structure wasn't changed.

				// Marshal the part struct back to JSON, then unmarshal to map
				var partMap map[string]interface{}
				partJson, err := json.Marshal(part)
				if err != nil {
					log.Printf("Warning: could not marshal text part: %v", err)
					continue
				}
				err = json.Unmarshal(partJson, &partMap)
				if err != nil {
					log.Printf("Warning: could not unmarshal text part to map: %v", err)
					continue
				}

				partText := getMapString(partMap, "text") // Get text from map

				// --- Liquid Processing ---
				if strings.Contains(partText, "{{") || strings.Contains(partText, "{%") {
					log.Printf("Liquid processing: %s", partText)
					engine := liquid.NewEngine() // Consider creating the engine once outside the loop if performance is critical
					var jsonData map[string]interface{}
					if templateData != "" { // Check if templateData is provided
						err := json.Unmarshal([]byte(templateData), &jsonData)
						if err != nil {
							// Return specific error for invalid JSON in templateData
							return "", fmt.Errorf("invalid JSON in templateData for text block (ID: %s): %w", block.ID, err)
						}
					} else {
						// Initialize empty map if templateData is empty, prevents nil map error in Render
						jsonData = make(map[string]interface{})
					}
					renderedContent, err := engine.ParseAndRenderString(partText, jsonData)
					if err != nil {
						// Return specific error for Liquid rendering issues
						return "", fmt.Errorf("liquid rendering error in text block (ID: %s): %w", block.ID, err)
					}
					partText = renderedContent // Update content with rendered result
				} else {
					// Only escape if it wasn't processed as Liquid
					partText = escapeHTML(partText)
				}

				// Extract styles/flags from the partMap
				isBold := getMapBool(partMap, "bold")
				isItalic := getMapBool(partMap, "italic")
				isUnderlined := getMapBool(partMap, "underlined")
				fontSize := getMapString(partMap, "fontSize")
				fontColor := getMapString(partMap, "fontColor")
				fontFamily := getMapString(partMap, "fontFamily")
				fontWeight := getMapString(partMap, "fontWeight")
				fontStyle := getMapString(partMap, "fontStyle")

				hyperlinkRaw, _ := partMap["hyperlink"].(map[string]interface{})
				isHyperlink := hyperlinkRaw != nil
				needsSpan := !isHyperlink && (isBold || isItalic || isUnderlined || fontSize != "" || fontColor != "" || fontFamily != "" || fontWeight != "" || fontStyle != "")

				if isHyperlink {
					linkURL := getMapString(hyperlinkRaw, "url")
					disableTracking := getMapBool(hyperlinkRaw, "disable_tracking")
					finalURL := linkURL
					if !disableTracking && linkURL != "" {
						finalURL = trackURL(linkURL, urlParams)
					}

					var linkStyleList []string
					// Style priority: Part styles > Block Hyperlink styles > Defaults

					// Color
					linkColor := fontColor // From part first
					if linkColor == "" {
						linkColor = textData.HyperlinkStyles.Color
					} // Use struct field
					if linkColor != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("color: %s !important", linkColor))
					}

					// Font Family
					linkFontFamily := fontFamily
					if linkFontFamily == "" {
						linkFontFamily = textData.HyperlinkStyles.FontFamily
					} // Use struct field
					if linkFontFamily != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-family: %s !important", linkFontFamily))
					}

					// Font Size
					linkFontSize := fontSize
					if linkFontSize == "" {
						linkFontSize = textData.HyperlinkStyles.FontSize
					} // Use struct field
					if linkFontSize != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-size: %s !important", linkFontSize))
					}

					// Font Style
					linkFontStyle := fontStyle // From part first
					if linkFontStyle == "" {
						linkFontStyle = textData.HyperlinkStyles.FontStyle
					} // Use struct field
					if isItalic {
						linkFontStyle = "italic"
					}
					if linkFontStyle != "normal" && linkFontStyle != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-style: %s !important", linkFontStyle))
					}

					// Font Weight
					linkFontWeight := fontWeight // From part first
					if linkFontWeight == "" {
						// Convert int fontWeight from struct to string
						if textData.HyperlinkStyles.FontWeight > 0 {
							linkFontWeight = strconv.Itoa(textData.HyperlinkStyles.FontWeight)
						} else {
							linkFontWeight = "normal" // Default if 0 or not set
						}
					}
					if isBold {
						linkFontWeight = "bold"
					}
					if linkFontWeight != "normal" && linkFontWeight != "" && linkFontWeight != "0" { // Add check for 0
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-weight: %s !important", linkFontWeight))
					}

					// Text Decoration (Underline)
					blockTextDecoration := textData.HyperlinkStyles.TextDecoration // Use struct field
					underline := true                                              // Default true for links
					if underlinedVal, ok := partMap["underlined"].(bool); ok {
						underline = underlinedVal
					} else if blockTextDecoration == "none" {
						underline = false
					}
					linkStyleList = append(linkStyleList, fmt.Sprintf("text-decoration: %s !important", map[bool]string{true: "underline", false: "none"}[underline]))

					// Text Transform
					linkTextTransform := textData.HyperlinkStyles.TextTransform // Use struct field
					if linkTextTransform != "none" && linkTextTransform != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("text-transform: %s !important", linkTextTransform))
					}

					linkStylesStr := strings.Join(linkStyleList, "; ")
					fmt.Fprintf(&lineContentSb, `<a style="%s" href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, linkStylesStr, finalURL, partText)

				} else if needsSpan {
					var spanStyleList []string
					if isBold {
						spanStyleList = append(spanStyleList, "font-weight: bold")
					} else if fontWeight != "" && fontWeight != "normal" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-weight: %s", fontWeight))
					}
					if isItalic {
						spanStyleList = append(spanStyleList, "font-style: italic")
					} else if fontStyle != "" && fontStyle != "normal" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-style: %s", fontStyle))
					}
					if isUnderlined {
						spanStyleList = append(spanStyleList, "text-decoration: underline")
					}
					if fontSize != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-size: %s !important", fontSize))
					}
					if fontColor != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("color: %s !important", fontColor))
					}
					if fontFamily != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-family: %s !important", fontFamily))
					}

					if len(spanStyleList) > 0 {
						fmt.Fprintf(&lineContentSb, `<span style="%s">%s</span>`, strings.Join(spanStyleList, "; "), partText)
					} else {
						lineContentSb.WriteString(partText)
					}
				} else {
					lineContentSb.WriteString(partText)
				}
			} // End parts loop
			lineContent := lineContentSb.String()

			// Apply styles for the line block (h1, p, etc.) using rootStyles
			var lineBlockStyleList []string
			lineTypeStyleMap := getStyleMap(lineType) // e.g., "paragraph", "h1"

			addStyleFromMapToList := func(cssProp, mapKey string, styleMap map[string]interface{}, styles *[]string, suffix string) {
				if valStr := getStyleString(styleMap, mapKey); valStr != "" && valStr != "0px" && valStr != "none" && valStr != "normal" && valStr != "0" {
					*styles = append(*styles, fmt.Sprintf("%s: %s%s", cssProp, valStr, suffix))
				}
			}

			addStyleFromMapToList("color", "color", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-family", "fontFamily", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-size", "fontSize", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-style", "fontStyle", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-weight", "fontWeight", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("line-height", "lineHeight", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("letter-spacing", "letterSpacing", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("text-decoration", "textDecoration", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("text-transform", "textTransform", lineTypeStyleMap, &lineBlockStyleList, " !important")

			// Padding & Margin using struct helpers with values from rootStyles map
			linePaddingControl := getStyleString(lineTypeStyleMap, "paddingControl")
			linePadding := getStyleString(lineTypeStyleMap, "padding")
			linePaddingTop := getStyleString(lineTypeStyleMap, "paddingTop")
			linePaddingRight := getStyleString(lineTypeStyleMap, "paddingRight")
			linePaddingBottom := getStyleString(lineTypeStyleMap, "paddingBottom")
			linePaddingLeft := getStyleString(lineTypeStyleMap, "paddingLeft")
			applyPaddingToStyleListFromStruct(linePaddingControl, linePadding, linePaddingTop, linePaddingRight, linePaddingBottom, linePaddingLeft, &lineBlockStyleList, " !important")

			lineMarginControl := getStyleString(lineTypeStyleMap, "marginControl")
			lineMargin := getStyleString(lineTypeStyleMap, "margin")
			lineMarginTop := getStyleString(lineTypeStyleMap, "marginTop")
			lineMarginRight := getStyleString(lineTypeStyleMap, "marginRight")
			lineMarginBottom := getStyleString(lineTypeStyleMap, "marginBottom")
			lineMarginLeft := getStyleString(lineTypeStyleMap, "marginLeft")
			applyMarginToStyleListFromStruct(lineMarginControl, lineMargin, lineMarginTop, lineMarginRight, lineMarginBottom, lineMarginLeft, &lineBlockStyleList, " !important")

			// Determine HTML tag (h1, h2, h3, p)
			htmlTag := lineType
			if htmlTag != "h1" && htmlTag != "h2" && htmlTag != "h3" {
				htmlTag = "p"
			} // Fallback just in case

			if strings.TrimSpace(lineContent) != "" {
				fmt.Fprintf(&contentSb, "<%s%s>%s</%s>\n",
					htmlTag, formatStyleAttr(lineBlockStyleList), lineContent, htmlTag)
			}
		} // End lines loop
		content = strings.TrimSpace(contentSb.String())
		children = nil

	case "heading": // Very similar to text, but uses HeadingBlockData
		var headingData HeadingBlockData
		if err := unmarshalBlockData(&headingData); err != nil {
			return "", err
		}

		tagName = "mj-text" // Headings are also mj-text
		textAttrs := make(map[string]interface{})
		textAttrs["align"] = headingData.Align
		textAttrs["padding"] = "0" // Override default

		if headingData.BackgroundColor != "" {
			textAttrs["container-background-color"] = headingData.BackgroundColor
		}

		// Padding for mj-text container using the block's padding fields
		applyPaddingFromStruct(headingData.PaddingControl, headingData.Padding, headingData.PaddingTop, headingData.PaddingRight, headingData.PaddingBottom, headingData.PaddingLeft, textAttrs)

		attributes = textAttrs

		// --- Content Generation --- (Mostly identical to text block, uses headingData.EditorData)
		var contentSb strings.Builder
		for _, lineRaw := range headingData.EditorData { // Use headingData
			lineType := lineRaw.Type                                     // Expected to be headingData.Type ("h1", "h2", "h3")? Or always "paragraph"? Assuming it matches headingData.Type
			if lineType != headingData.Type && lineType != "paragraph" { // Allow paragraph within heading block?
				log.Printf("Warning: Heading block (ID: %s, Type: %s) contains line of unexpected type: %s", block.ID, headingData.Type, lineType)
				// Decide how to handle - skip, default to headingData.Type, or use lineType? Let's use headingData.Type for consistency.
				lineType = headingData.Type
			}

			childrenRaw := lineRaw.Children

			var lineContentSb strings.Builder
			for _, part := range childrenRaw {
				// Marshal part back to map for style access (same as text block)
				var partMap map[string]interface{}
				partJson, err := json.Marshal(part)
				if err != nil {
					log.Printf("Warning: could not marshal heading part: %v", err)
					continue
				}
				err = json.Unmarshal(partJson, &partMap)
				if err != nil {
					log.Printf("Warning: could not unmarshal heading part to map: %v", err)
					continue
				}

				partText := getMapString(partMap, "text") // Get text from map

				// --- Liquid Processing --- (identical to text block)
				if strings.Contains(partText, "{{") || strings.Contains(partText, "{%") {
					engine := liquid.NewEngine()
					var jsonData map[string]interface{}
					if templateData != "" { // Check if templateData is provided
						err := json.Unmarshal([]byte(templateData), &jsonData)
						if err != nil {
							return "", fmt.Errorf("invalid JSON in templateData for heading block (ID: %s): %w", block.ID, err)
						}
					} else {
						jsonData = make(map[string]interface{})
					}
					renderedContent, err := engine.ParseAndRenderString(partText, jsonData) // Use partText here
					if err != nil {
						return "", fmt.Errorf("liquid rendering error in heading block (ID: %s): %w", block.ID, err)
					}
					log.Printf("Heading block partText: %s", renderedContent)
					partText = renderedContent
				} else {
					// Only escape if it wasn't processed as Liquid
					partText = escapeHTML(partText)
				}

				isBold := getMapBool(partMap, "bold")
				isItalic := getMapBool(partMap, "italic")
				isUnderlined := getMapBool(partMap, "underlined")
				fontSize := getMapString(partMap, "fontSize")
				fontColor := getMapString(partMap, "fontColor")
				fontFamily := getMapString(partMap, "fontFamily")
				fontWeight := getMapString(partMap, "fontWeight")
				fontStyle := getMapString(partMap, "fontStyle")

				hyperlinkRaw, _ := partMap["hyperlink"].(map[string]interface{})
				isHyperlink := hyperlinkRaw != nil
				needsSpan := !isHyperlink && (isBold || isItalic || isUnderlined || fontSize != "" || fontColor != "" || fontFamily != "" || fontWeight != "" || fontStyle != "")

				if isHyperlink { // Hyperlink logic identical to text block, but uses rootStyles["hyperlink"]
					linkURL := getMapString(hyperlinkRaw, "url")
					disableTracking := getMapBool(hyperlinkRaw, "disable_tracking")
					finalURL := linkURL
					if !disableTracking && linkURL != "" {
						finalURL = trackURL(linkURL, urlParams)
					}

					var linkStyleList []string
					hyperlinkRootStyles := getStyleMap("hyperlink") // Use rootStyles for links within headings too

					linkColor := fontColor
					if linkColor == "" {
						linkColor = getStyleString(hyperlinkRootStyles, "color")
					}
					if linkColor != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("color: %s !important", linkColor))
					}

					linkFontFamily := fontFamily
					if linkFontFamily == "" {
						linkFontFamily = getStyleString(hyperlinkRootStyles, "fontFamily")
					}
					if linkFontFamily != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-family: %s !important", linkFontFamily))
					}

					linkFontSize := fontSize
					if linkFontSize == "" {
						linkFontSize = getStyleString(hyperlinkRootStyles, "fontSize")
					}
					if linkFontSize != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-size: %s !important", linkFontSize))
					}

					linkFontStyle := fontStyle
					if linkFontStyle == "" {
						linkFontStyle = getStyleString(hyperlinkRootStyles, "fontStyle")
					}
					if isItalic {
						linkFontStyle = "italic"
					}
					if linkFontStyle != "normal" && linkFontStyle != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-style: %s !important", linkFontStyle))
					}

					linkFontWeight := fontWeight
					if linkFontWeight == "" {
						linkFontWeight = getStyleString(hyperlinkRootStyles, "fontWeight")
					}
					if isBold {
						linkFontWeight = "bold"
					}
					if linkFontWeight != "normal" && linkFontWeight != "" && linkFontWeight != "0" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("font-weight: %s !important", linkFontWeight))
					}

					blockTextDecoration := getStyleString(hyperlinkRootStyles, "textDecoration")
					underline := true
					if underlinedVal, ok := partMap["underlined"].(bool); ok {
						underline = underlinedVal
					} else if blockTextDecoration == "none" {
						underline = false
					}
					linkStyleList = append(linkStyleList, fmt.Sprintf("text-decoration: %s !important", map[bool]string{true: "underline", false: "none"}[underline]))

					linkTextTransform := getStyleString(hyperlinkRootStyles, "textTransform")
					if linkTextTransform != "none" && linkTextTransform != "" {
						linkStyleList = append(linkStyleList, fmt.Sprintf("text-transform: %s !important", linkTextTransform))
					}

					linkStylesStr := strings.Join(linkStyleList, "; ")
					fmt.Fprintf(&lineContentSb, `<a style="%s" href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, linkStylesStr, finalURL, partText)

				} else if needsSpan { // Span logic identical to text block
					var spanStyleList []string
					if isBold {
						spanStyleList = append(spanStyleList, "font-weight: bold")
					} else if fontWeight != "" && fontWeight != "normal" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-weight: %s", fontWeight))
					}
					if isItalic {
						spanStyleList = append(spanStyleList, "font-style: italic")
					} else if fontStyle != "" && fontStyle != "normal" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-style: %s", fontStyle))
					}
					if isUnderlined {
						spanStyleList = append(spanStyleList, "text-decoration: underline")
					}
					if fontSize != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-size: %s !important", fontSize))
					}
					if fontColor != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("color: %s !important", fontColor))
					}
					if fontFamily != "" {
						spanStyleList = append(spanStyleList, fmt.Sprintf("font-family: %s !important", fontFamily))
					}

					if len(spanStyleList) > 0 {
						fmt.Fprintf(&lineContentSb, `<span style="%s">%s</span>`, strings.Join(spanStyleList, "; "), partText)
					} else {
						lineContentSb.WriteString(partText)
					}
				} else {
					lineContentSb.WriteString(partText)
				}
			} // End parts loop
			lineContent := lineContentSb.String()

			// Apply styles for the line block (h1, h2, h3) using rootStyles
			var lineBlockStyleList []string
			// Define lineTypeStyleMap *before* using it
			lineTypeStyleMap := getStyleMap(lineType)

			// Use the same helper as text block
			addStyleFromMapToList := func(cssProp, mapKey string, styleMap map[string]interface{}, styles *[]string, suffix string) {
				if valStr := getStyleString(styleMap, mapKey); valStr != "" && valStr != "0px" && valStr != "none" && valStr != "normal" && valStr != "0" {
					*styles = append(*styles, fmt.Sprintf("%s: %s%s", cssProp, valStr, suffix))
				}
			}

			addStyleFromMapToList("color", "color", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-family", "fontFamily", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-size", "fontSize", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-style", "fontStyle", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("font-weight", "fontWeight", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("line-height", "lineHeight", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("letter-spacing", "letterSpacing", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("text-decoration", "textDecoration", lineTypeStyleMap, &lineBlockStyleList, " !important")
			addStyleFromMapToList("text-transform", "textTransform", lineTypeStyleMap, &lineBlockStyleList, " !important")

			// Padding & Margin using struct helpers with values from rootStyles map
			linePaddingControl := getStyleString(lineTypeStyleMap, "paddingControl")
			linePadding := getStyleString(lineTypeStyleMap, "padding")
			linePaddingTop := getStyleString(lineTypeStyleMap, "paddingTop")
			linePaddingRight := getStyleString(lineTypeStyleMap, "paddingRight")
			linePaddingBottom := getStyleString(lineTypeStyleMap, "paddingBottom")
			linePaddingLeft := getStyleString(lineTypeStyleMap, "paddingLeft")
			applyPaddingToStyleListFromStruct(linePaddingControl, linePadding, linePaddingTop, linePaddingRight, linePaddingBottom, linePaddingLeft, &lineBlockStyleList, " !important")

			lineMarginControl := getStyleString(lineTypeStyleMap, "marginControl")
			lineMargin := getStyleString(lineTypeStyleMap, "margin")
			lineMarginTop := getStyleString(lineTypeStyleMap, "marginTop")
			lineMarginRight := getStyleString(lineTypeStyleMap, "marginRight")
			lineMarginBottom := getStyleString(lineTypeStyleMap, "marginBottom")
			lineMarginLeft := getStyleString(lineTypeStyleMap, "marginLeft")
			applyMarginToStyleListFromStruct(lineMarginControl, lineMargin, lineMarginTop, lineMarginRight, lineMarginBottom, lineMarginLeft, &lineBlockStyleList, " !important")

			// Use the determined HTML tag (h1, h2, h3)
			htmlTag := lineType
			if htmlTag != "h1" && htmlTag != "h2" && htmlTag != "h3" {
				htmlTag = "p"
			} // Fallback just in case

			if strings.TrimSpace(lineContent) != "" {
				fmt.Fprintf(&contentSb, "<%s%s>%s</%s>\n",
					htmlTag, formatStyleAttr(lineBlockStyleList), lineContent, htmlTag)
			}
		} // End lines loop
		content = strings.TrimSpace(contentSb.String())

		children = nil

	case "image":
		var imgData ImageBlockData
		if err := unmarshalBlockData(&imgData); err != nil {
			return "", err
		}

		tagName = "mj-image"
		imageAttrs := make(map[string]interface{})
		imageAttrs["align"] = imgData.Wrapper.Align // Use wrapper align
		imageAttrs["src"] = imgData.Image.Src
		imageAttrs["alt"] = imgData.Image.Alt // Default to empty string if not present

		// Height/Width handling - Note: height is not in ImageBlockData, maybe intended?
		// Assume height comes from elsewhere or is auto if not in struct.
		// if height := imgData.Image.Height; height != "auto" { imageAttrs["height"] = height }
		if width := imgData.Image.Width; width != "" && width != "auto" { // Check if width is defined
			// MJML validator seems to require px unit, contradicting some docs.
			// Ensure px unit is present.
			cleanedWidth := strings.TrimSuffix(strings.TrimSpace(width), "px")
			// Check if it's just a number before appending px
			if _, err := strconv.Atoi(cleanedWidth); err == nil {
				imageAttrs["width"] = cleanedWidth + "px"
			} else {
				// If it's not a simple number (e.g., percentage?), log warning or handle differently?
				// For now, let's pass the original value, assuming it might be valid in some context
				// Or perhaps default to auto? Let's stick with original for now.
				imageAttrs["width"] = width // Keep original if not a simple number
				log.Printf("Warning: mj-image width '%s' is not a simple pixel value. Passing as is.", width)
			}
			// Old code: imageAttrs["width"] = strings.TrimSuffix(width, "px")
		}
		// Add fluid-on-mobile if needed (assuming it's a boolean field in ImageBlockData, which it isn't currently)
		// if imgData.Image.FullWidthOnMobile { imageAttrs["fluid-on-mobile"] = true }
		imageAttrs["padding"] = "0" // Reset default padding initially

		// Href
		if href := imgData.Image.Href; href != "" {
			// Assume disable_tracking is a field under Image struct if needed
			// disableTracking := imgData.Image.DisableTracking
			disableTracking := false // Default to false if field doesn't exist
			finalURL := href
			if !disableTracking {
				finalURL = trackURL(href, urlParams)
			}
			imageAttrs["href"] = finalURL
			imageAttrs["target"] = "_blank"
			imageAttrs["rel"] = "noopener noreferrer"
		}

		// Border Radius - Use wrapper border radius
		if borderRadius := imgData.Wrapper.BorderRadius; borderRadius != "" && borderRadius != "0px" {
			imageAttrs["border-radius"] = borderRadius
		}

		// Padding from wrapper - Use new struct helper
		applyPaddingFromStruct(imgData.Wrapper.PaddingControl, imgData.Wrapper.Padding, imgData.Wrapper.PaddingTop, imgData.Wrapper.PaddingRight, imgData.Wrapper.PaddingBottom, imgData.Wrapper.PaddingLeft, imageAttrs)

		// Border from wrapper - Use new struct helper
		applyBordersFromStruct(imageAttrs, imgData.Wrapper.BorderControl,
			imgData.Wrapper.BorderStyle, imgData.Wrapper.BorderWidth, imgData.Wrapper.BorderColor,
			imgData.Wrapper.BorderTopStyle, imgData.Wrapper.BorderTopWidth, imgData.Wrapper.BorderTopColor,
			imgData.Wrapper.BorderRightStyle, imgData.Wrapper.BorderRightWidth, imgData.Wrapper.BorderRightColor,
			imgData.Wrapper.BorderBottomStyle, imgData.Wrapper.BorderBottomWidth, imgData.Wrapper.BorderBottomColor,
			imgData.Wrapper.BorderLeftStyle, imgData.Wrapper.BorderLeftWidth, imgData.Wrapper.BorderLeftColor)

		// Container background from wrapper (assuming field exists)
		// if wrapBgColor := imgData.Wrapper.BackgroundColor; wrapBgColor != "" { imageAttrs["container-background-color"] = wrapBgColor }

		attributes = imageAttrs
		children = nil // mj-image is self-contained

	case "button":
		var buttonData ButtonBlockData
		if err := unmarshalBlockData(&buttonData); err != nil {
			return "", err
		}
		// Add validation: Check if essential fields were populated
		if buttonData.Button.Text == "" {
			// Log or return error if critical data is missing after unmarshal
			log.Printf("Warning: Button block (ID: %s) missing text after data unmarshal. Input data might be invalid.", block.ID)
			// Depending on desired strictness, we could return an error:
			// return "", fmt.Errorf("invalid data for button block (ID: %s): missing button text", block.ID)
		}

		tagName = "mj-button"
		buttonAttrs := make(map[string]interface{})
		buttonAttrs["align"] = buttonData.Wrapper.Align // Use wrapper align

		// Href
		if href := buttonData.Button.Href; href != "" {
			finalURL := href
			if !buttonData.Button.DisableTracking {
				finalURL = trackURL(href, urlParams)
			}
			buttonAttrs["href"] = finalURL
			buttonAttrs["target"] = "_blank"
			buttonAttrs["rel"] = "noopener noreferrer"
		}

		buttonAttrs["background-color"] = buttonData.Button.BackgroundColor
		buttonAttrs["font-family"] = buttonData.Button.FontFamily
		if fs := buttonData.Button.FontSize; fs != "" {
			// Ensure px unit is present for font-size
			cleanedFs := strings.TrimSuffix(strings.TrimSpace(fs), "px")
			if _, err := strconv.Atoi(cleanedFs); err == nil {
				buttonAttrs["font-size"] = cleanedFs + "px"
			} else {
				buttonAttrs["font-size"] = fs // Keep original if not a simple number
				log.Printf("Warning: mj-button font-size '%s' is not a simple pixel value. Passing as is.", fs)
			}
			// Old code: buttonAttrs["font-size"] = strings.TrimSuffix(fs, "px")
		}
		buttonAttrs["font-weight"] = buttonData.Button.FontWeight // Let lineAttributes handle formatting
		if fst := buttonData.Button.FontStyle; fst != "normal" {
			buttonAttrs["font-style"] = fst
		}
		buttonAttrs["color"] = buttonData.Button.Color
		buttonAttrs["padding"] = "0" // Reset default padding initially

		// Inner padding
		innerVPad := buttonData.Button.InnerVerticalPadding
		innerHPad := buttonData.Button.InnerHorizontalPadding
		// Provide defaults if empty
		if innerVPad == "" {
			innerVPad = "10"
		}
		if innerHPad == "" {
			innerHPad = "25"
		}
		buttonAttrs["inner-padding"] = fmt.Sprintf("%spx %spx", strings.TrimSuffix(innerVPad, "px"), strings.TrimSuffix(innerHPad, "px"))

		if tt := buttonData.Button.TextTransform; tt != "none" {
			buttonAttrs["text-transform"] = tt
		}
		if br := buttonData.Button.BorderRadius; br != "0px" {
			buttonAttrs["border-radius"] = br
		}
		if width := buttonData.Button.Width; width != "auto" {
			buttonAttrs["width"] = strings.TrimSuffix(width, "px")
		}
		// buttonAttrs["vertical-align"] = buttonData.Wrapper.VerticalAlign // Field does not exist on WrapperStyles

		// Padding from wrapper - Use new struct helper
		applyPaddingFromStruct(buttonData.Wrapper.PaddingControl, buttonData.Wrapper.Padding, buttonData.Wrapper.PaddingTop, buttonData.Wrapper.PaddingRight, buttonData.Wrapper.PaddingBottom, buttonData.Wrapper.PaddingLeft, buttonAttrs)

		// Border from button data itself - Use new struct helper
		applyBordersFromStruct(buttonAttrs, buttonData.Button.BorderControl,
			buttonData.Button.BorderStyle, buttonData.Button.BorderWidth, buttonData.Button.BorderColor,
			buttonData.Button.BorderTopStyle, buttonData.Button.BorderTopWidth, buttonData.Button.BorderTopColor,
			buttonData.Button.BorderRightStyle, buttonData.Button.BorderRightWidth, buttonData.Button.BorderRightColor,
			buttonData.Button.BorderBottomStyle, buttonData.Button.BorderBottomWidth, buttonData.Button.BorderBottomColor,
			buttonData.Button.BorderLeftStyle, buttonData.Button.BorderLeftWidth, buttonData.Button.BorderLeftColor)
		// Note: BorderRadius for button is handled separately below (if needed)

		// Apply button's specific border-radius (overrides any from wrapper if set on button)
		if br := buttonData.Button.BorderRadius; br != "" && br != "0px" {
			buttonAttrs["border-radius"] = br
		}

		// Container background from wrapper (assuming field exists)
		// if wrapBgColor := buttonData.Wrapper.BackgroundColor; wrapBgColor != "" { buttonAttrs["container-background-color"] = wrapBgColor }

		attributes = buttonAttrs
		content = escapeHTML(buttonData.Button.Text) // Use struct field
		children = nil

	case "divider":
		var dividerData DividerBlockData
		if err := unmarshalBlockData(&dividerData); err != nil {
			return "", err
		}

		tagName = "mj-divider"
		dividerAttrs := make(map[string]interface{})
		dividerAttrs["align"] = dividerData.Align
		dividerAttrs["border-color"] = dividerData.BorderColor
		dividerAttrs["border-style"] = dividerData.BorderStyle
		dividerAttrs["border-width"] = dividerData.BorderWidth
		if width := dividerData.Width; width != "" && width != "100%" { // Check if width is defined
			dividerAttrs["width"] = strings.TrimSuffix(width, "px")
		}
		dividerAttrs["padding"] = "0" // Reset default padding initially

		if bgColor := dividerData.BackgroundColor; bgColor != "" {
			dividerAttrs["container-background-color"] = bgColor
		}

		// Padding for divider container - Use new struct helper
		applyPaddingFromStruct(dividerData.PaddingControl, dividerData.Padding, dividerData.PaddingTop, dividerData.PaddingRight, dividerData.PaddingBottom, dividerData.PaddingLeft, dividerAttrs)

		attributes = dividerAttrs
		children = nil

	case "openTracking":
		// No data struct needed, logic remains the same
		tagName = "mj-raw"
		attributes = nil // Not needed for mj-raw content
		// Note: Using single quotes within the style attribute for the Go string literal.
		content = `<img src="{{ open_tracking_pixel_src }}" alt="" height="1" width="1" style="display:block; max-height:1px; max-width:1px; visibility:hidden; mso-hide:all; border:0; padding:0;" />`
		children = nil

	case "liquid":
		var liquidData LiquidBlockData
		if err := unmarshalBlockData(&liquidData); err != nil {
			return "", err
		}

		liquidCode := liquidData.LiquidCode // Use struct field
		if liquidCode == "" {
			log.Printf("Warning: Liquid block (ID: %s) has empty liquidCode", block.ID)
			return fmt.Sprintf("<!-- Liquid block ID %s has empty code -->", block.ID), nil
		}

		engine := liquid.NewEngine()
		var jsonData map[string]interface{}
		if templateData != "" {
			err := json.Unmarshal([]byte(templateData), &jsonData)
			if err != nil {
				return "", fmt.Errorf("invalid JSON in templateData for liquid block (ID: %s): %w", block.ID, err)
			}
		} else {
			jsonData = make(map[string]interface{})
		}

		renderedContent, err := engine.ParseAndRenderString(liquidCode, jsonData)
		if err != nil {
			return "", fmt.Errorf("liquid rendering error in liquid block (ID: %s): %w", block.ID, err)
		}
		return renderedContent, nil // Return raw rendered content

	case "spacer": // <<< ADD THIS CASE
		var spacerData map[string]interface{} // Use map for simple structure
		if block.Data != nil {
			var ok bool
			spacerData, ok = block.Data.(map[string]interface{})
			if !ok {
				// Attempt unmarshal if it's not already a map
				if err := unmarshalBlockData(&spacerData); err != nil {
					// If unmarshal also fails, return error or log warning
					log.Printf("Warning: spacer block data (ID: %s) is not map[string]interface{} and failed to unmarshal: %v", block.ID, err)
					spacerData = make(map[string]interface{}) // Use empty map
				}
			}
		} else {
			spacerData = make(map[string]interface{}) // Use empty map if Data is nil
		}

		tagName = "mj-spacer"
		spacerAttrs := make(map[string]interface{})
		if height, ok := spacerData["height"]; ok {
			spacerAttrs["height"] = fmt.Sprintf("%v", height)
		}
		if bgColor, ok := spacerData["backgroundColor"]; ok && fmt.Sprintf("%v", bgColor) != "" {
			spacerAttrs["container-background-color"] = fmt.Sprintf("%v", bgColor)
		}
		// Spacers generally don't have padding or borders in the same way
		attributes = spacerAttrs

	default:
		log.Printf("Warning: MJML conversion not implemented for block kind: %s (ID: %s)", block.Kind, block.ID)
		return fmt.Sprintf("%s<!-- MJML Not Implemented: %s -->", space, block.Kind), nil
	}

	// --- Common Processing for Children (if not handled above and childrenMjml not set) ---
	if len(children) > 0 && childrenMjml == "" {
		var childrenSb strings.Builder
		for _, child := range children {
			childMjml, err := TreeToMjml(rootStyles, child, templateData, urlParams, indent+2, &block)
			if err != nil {
				return "", fmt.Errorf("error processing child block (ID: %s, Kind: %s): %w", child.ID, child.Kind, err)
			}
			if strings.TrimSpace(childMjml) != "" {
				childrenSb.WriteString(childMjml)
				childrenSb.WriteString("\n") // Newline between children
			}
		}
		childrenMjml = strings.TrimSuffix(childrenSb.String(), "\n") // Remove trailing newline
	}

	// --- Assemble MJML String ---
	if tagName == "" {
		if childrenMjml != "" {
			return childrenMjml, nil
		}
		return "", nil
	}

	attrString := lineAttributes(attributes)
	openTag := fmt.Sprintf("%s<%s%s>", space, tagName, formatAttrs(attrString))
	closeTag := fmt.Sprintf("</%s>", tagName)

	trimmedContent := strings.TrimSpace(content)
	trimmedChildrenMjml := strings.TrimSpace(childrenMjml)

	if trimmedContent != "" {
		var indentedContent string
		if strings.Contains(content, "\n") {
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				lines[i] = indentPad(indent+2) + line
			}
			indentedContent = strings.Join(lines, "\n")
		} else {
			indentedContent = indentPad(indent+2) + content
		}
		fmt.Fprintf(&sb, "%s\n%s\n%s%s", openTag, indentedContent, space, closeTag)
	} else if trimmedChildrenMjml != "" {
		fmt.Fprintf(&sb, "%s\n%s\n%s%s", openTag, childrenMjml, space, closeTag)
	} else {
		fmt.Fprintf(&sb, "%s%s", openTag, closeTag)
	}

	return sb.String(), nil // Success
}

// --- NEW Helper functions for Struct-based data ---

// applyPaddingFromStruct adds padding attributes based on struct fields.
func applyPaddingFromStruct(paddingControl, padding, paddingTop, paddingRight, paddingBottom, paddingLeft string, attrs map[string]interface{}) {
	if paddingControl == "all" && padding != "" && padding != "0px" {
		attrs["padding"] = padding // Overrides previous "padding: 0" if set
	} else if paddingControl == "separate" {
		// Remove global padding if specific ones are set
		delete(attrs, "padding")
		if paddingTop != "" && paddingTop != "0px" {
			attrs["padding-top"] = paddingTop
		}
		if paddingRight != "" && paddingRight != "0px" {
			attrs["padding-right"] = paddingRight
		}
		if paddingBottom != "" && paddingBottom != "0px" {
			attrs["padding-bottom"] = paddingBottom
		}
		if paddingLeft != "" && paddingLeft != "0px" {
			attrs["padding-left"] = paddingLeft
		}
	} else { // Added else block for default/invalid control
		// Default behavior: use shorthand padding if provided and not zero
		if padding != "" && padding != "0px" {
			attrs["padding"] = padding
		} // Otherwise, leave attrs unchanged (don't default to padding:0 here)
	}
}

// applyBordersFromStruct adds border attributes based on struct fields.
// Supports both "all" and "separate" borderControl.
func applyBordersFromStruct(
	attrs map[string]interface{}, // Attributes map to modify
	borderControl string, // "all" or "separate"
	// Fields for borderControl="all"
	borderStyle string,
	borderWidth string,
	borderColor string,
	// Fields for borderControl="separate"
	borderTopStyle string, borderTopWidth string, borderTopColor string,
	borderRightStyle string, borderRightWidth string, borderRightColor string,
	borderBottomStyle string, borderBottomWidth string, borderBottomColor string,
	borderLeftStyle string, borderLeftWidth string, borderLeftColor string,
) {
	if borderControl == "all" && borderStyle != "" && borderStyle != "none" && borderWidth != "" && borderWidth != "0px" && borderColor != "" {
		attrs["border"] = fmt.Sprintf("%s %s %s", borderWidth, borderStyle, borderColor)
	} else if borderControl == "separate" {
		// Add individual border sides if defined
		if borderTopStyle != "" && borderTopStyle != "none" && borderTopWidth != "" && borderTopWidth != "0px" && borderTopColor != "" {
			attrs["border-top"] = fmt.Sprintf("%s %s %s", borderTopWidth, borderTopStyle, borderTopColor)
		}
		if borderRightStyle != "" && borderRightStyle != "none" && borderRightWidth != "" && borderRightWidth != "0px" && borderRightColor != "" {
			attrs["border-right"] = fmt.Sprintf("%s %s %s", borderRightWidth, borderRightStyle, borderRightColor)
		}
		if borderBottomStyle != "" && borderBottomStyle != "none" && borderBottomWidth != "" && borderBottomWidth != "0px" && borderBottomColor != "" {
			attrs["border-bottom"] = fmt.Sprintf("%s %s %s", borderBottomWidth, borderBottomStyle, borderBottomColor)
		}
		if borderLeftStyle != "" && borderLeftStyle != "none" && borderLeftWidth != "" && borderLeftWidth != "0px" && borderLeftColor != "" {
			attrs["border-left"] = fmt.Sprintf("%s %s %s", borderLeftWidth, borderLeftStyle, borderLeftColor)
		}
	}
	// Note: BorderRadius is handled separately where needed as it applies regardless of control type.
}

// applyPaddingToStyleListFromStruct adds padding styles based on struct fields.
func applyPaddingToStyleListFromStruct(paddingControl, padding, paddingTop, paddingRight, paddingBottom, paddingLeft string, styleList *[]string, suffix string) {
	if paddingControl == "all" {
		// Check value *without* suffix first
		if padding != "" && padding != "0px" {
			*styleList = append(*styleList, fmt.Sprintf("padding: %s%s", padding, suffix))
		}
	} else if paddingControl == "separate" {
		// Check value without suffix or "!important" suffix
		if paddingTop != "" && paddingTop != "0px" && paddingTop != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("padding-top: %s%s", paddingTop, suffix))
		}
		if paddingRight != "" && paddingRight != "0px" && paddingRight != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("padding-right: %s%s", paddingRight, suffix))
		}
		if paddingBottom != "" && paddingBottom != "0px" && paddingBottom != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("padding-bottom: %s%s", paddingBottom, suffix))
		}
		if paddingLeft != "" && paddingLeft != "0px" && paddingLeft != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("padding-left: %s%s", paddingLeft, suffix))
		}
	} else {
		// Default behavior: use shorthand padding if provided and not zero (check *without* suffix)
		if padding != "" && padding != "0px" {
			*styleList = append(*styleList, fmt.Sprintf("padding: %s%s", padding, suffix))
		}
		// No default padding:0 is added here, as it might conflict with other styles
	}
}

// applyMarginToStyleListFromStruct adds margin styles based on struct fields.
func applyMarginToStyleListFromStruct(marginControl, margin, marginTop, marginRight, marginBottom, marginLeft string, styleList *[]string, suffix string) {
	hasSpecificMargin := false
	if marginControl == "all" {
		// Check value *without* suffix first
		if margin != "" && margin != "0px" {
			*styleList = append(*styleList, fmt.Sprintf("margin: %s%s", margin, suffix))
			hasSpecificMargin = true
		}
	} else if marginControl == "separate" {
		// Check value without suffix or "!important" suffix
		if marginTop != "" && marginTop != "0px" && marginTop != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("margin-top: %s%s", marginTop, suffix))
			hasSpecificMargin = true
		}
		if marginRight != "" && marginRight != "0px" && marginRight != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("margin-right: %s%s", marginRight, suffix))
			hasSpecificMargin = true
		}
		if marginBottom != "" && marginBottom != "0px" && marginBottom != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("margin-bottom: %s%s", marginBottom, suffix))
			hasSpecificMargin = true
		}
		if marginLeft != "" && marginLeft != "0px" && marginLeft != "0px !important" {
			*styleList = append(*styleList, fmt.Sprintf("margin-left: %s%s", marginLeft, suffix))
			hasSpecificMargin = true
		}
	} else {
		// Default behavior: use shorthand margin if provided and not zero (check *without* suffix)
		if margin != "" && margin != "0px" {
			*styleList = append(*styleList, fmt.Sprintf("margin: %s%s", margin, suffix))
			hasSpecificMargin = true
		}
	}

	// Ensure margin:0 is applied if no specific margin is set
	if !hasSpecificMargin {
		// Use explicit !important suffix for the default margin
		if suffix == "" {
			*styleList = append(*styleList, "margin: 0px !important")
		} else {
			*styleList = append(*styleList, fmt.Sprintf("margin: 0px%s", suffix))
		}
	}
}

// --- End NEW Helper functions ---

// --- Remaining required helpers ---

// getMapBool safely gets a boolean value from a map[string]interface{}.
// Used only for editorData parts which are handled as maps.
func getMapBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
		// Handle string "true" ?
		if strVal, ok := val.(string); ok {
			return strings.ToLower(strVal) == "true"
		}
	}
	return false
}

// getMapString safely gets a string value from a map[string]interface{}.
// Handles potential number types from JSON unmarshalling.
// Kept only for accessing editorData parts which remain maps temporarily.
func getMapString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok && val != nil {
		switch v := val.(type) {
		case string:
			return v
		case float64: // JSON numbers often unmarshal as float64
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

// formatAttrs adds a leading space if attrs string is not empty.
func formatAttrs(attrs string) string {
	if attrs != "" {
		return " " + attrs
	}
	return ""
}

// formatStyleAttr formats a list of styles into a style="..." attribute string.
func formatStyleAttr(styles []string) string {
	if len(styles) > 0 {
		// Filter out empty strings just in case
		validStyles := make([]string, 0, len(styles))
		for _, s := range styles {
			if strings.TrimSpace(s) != "" {
				validStyles = append(validStyles, strings.TrimSpace(s))
			}
		}
		if len(validStyles) > 0 {
			// Join with semicolon and space for readability, ensure no trailing semicolon
			return fmt.Sprintf(` style="%s"`, strings.TrimSuffix(strings.Join(validStyles, "; "), "; "))
		}
	}
	return ""
}

// escapeHTML performs basic HTML escaping for text content.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;") // Must be first
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;") // Use backticks for raw string
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
