package notifuse_mjml

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/osteele/liquid"
)

// ConvertJSONToMJML converts an EmailBlock JSON tree to MJML string
func ConvertJSONToMJML(tree EmailBlock) string {
	return convertBlockToMJML(tree, 0, "")
}

// ConvertJSONToMJMLWithData converts an EmailBlock JSON tree to MJML string with template data
func ConvertJSONToMJMLWithData(tree EmailBlock, templateData string) (string, error) {
	return convertBlockToMJMLWithError(tree, 0, templateData)
}

// convertBlockToMJMLWithError recursively converts a single EmailBlock to MJML string with error handling
func convertBlockToMJMLWithError(block EmailBlock, indentLevel int, templateData string) (string, error) {
	indent := strings.Repeat("  ", indentLevel)
	tagName := string(block.GetType())
	children := block.GetChildren()

	// Handle self-closing tags that don't have children but may have content
	if len(children) == 0 {
		// Check if the block has content (for mj-text, mj-button, etc.)
		content := getBlockContent(block)

		if content != "" {
			// Process Liquid templating for mj-text, mj-button, mj-title, mj-preview, and mj-raw blocks
			blockType := block.GetType()
			if blockType == MJMLComponentMjText || blockType == MJMLComponentMjButton || blockType == MJMLComponentMjTitle || blockType == MJMLComponentMjPreview || blockType == MJMLComponentMjRaw {
				processedContent, err := processLiquidContent(content, templateData, block.GetID())
				if err != nil {
					// Return error instead of just logging
					return "", fmt.Errorf("liquid processing failed for block %s: %v", block.GetID(), err)
				} else {
					content = processedContent
				}
			}

			// Block with content - don't escape for mj-raw, mj-text, and mj-button (they can contain HTML)
			attributeString := formatAttributes(block.GetAttributes())
			if blockType == MJMLComponentMjRaw || blockType == MJMLComponentMjText || blockType == MJMLComponentMjButton {
				return fmt.Sprintf("%s<%s%s>%s</%s>", indent, tagName, attributeString, content, tagName), nil
			} else {
				return fmt.Sprintf("%s<%s%s>%s</%s>", indent, tagName, attributeString, escapeContent(content), tagName), nil
			}
		} else {
			// Self-closing block or empty block
			attributeString := formatAttributes(block.GetAttributes())
			if attributeString != "" {
				return fmt.Sprintf("%s<%s%s />", indent, tagName, attributeString), nil
			} else {
				return fmt.Sprintf("%s<%s />", indent, tagName), nil
			}
		}
	}

	// Block with children
	attributeString := formatAttributes(block.GetAttributes())
	openTag := fmt.Sprintf("%s<%s%s>", indent, tagName, attributeString)
	closeTag := fmt.Sprintf("%s</%s>", indent, tagName)

	// Process children
	var childrenMJML []string
	for _, child := range children {
		if child != nil {
			childMJML, err := convertBlockToMJMLWithError(child, indentLevel+1, templateData)
			if err != nil {
				return "", err
			}
			childrenMJML = append(childrenMJML, childMJML)
		}
	}

	return fmt.Sprintf("%s\n%s\n%s", openTag, strings.Join(childrenMJML, "\n"), closeTag), nil
}

// convertBlockToMJML recursively converts a single EmailBlock to MJML string
func convertBlockToMJML(block EmailBlock, indentLevel int, templateData string) string {
	indent := strings.Repeat("  ", indentLevel)
	tagName := string(block.GetType())
	children := block.GetChildren()

	// Handle self-closing tags that don't have children but may have content
	if len(children) == 0 {
		// Check if the block has content (for mj-text, mj-button, etc.)
		content := getBlockContent(block)

		if content != "" {
			// Process Liquid templating for mj-text, mj-button, mj-title, mj-preview, and mj-raw blocks
			blockType := block.GetType()
			if blockType == MJMLComponentMjText || blockType == MJMLComponentMjButton || blockType == MJMLComponentMjTitle || blockType == MJMLComponentMjPreview || blockType == MJMLComponentMjRaw {
				processedContent, err := processLiquidContent(content, templateData, block.GetID())
				if err != nil {
					// Log error but continue with original content
					fmt.Printf("Warning: Liquid processing failed for block %s: %v\n", block.GetID(), err)
				} else {
					content = processedContent
				}
			}

			// Block with content - don't escape for mj-raw, mj-text, and mj-button (they can contain HTML)
			attributeString := formatAttributes(block.GetAttributes())
			if blockType == MJMLComponentMjRaw || blockType == MJMLComponentMjText || blockType == MJMLComponentMjButton {
				return fmt.Sprintf("%s<%s%s>%s</%s>", indent, tagName, attributeString, content, tagName)
			} else {
				return fmt.Sprintf("%s<%s%s>%s</%s>", indent, tagName, attributeString, escapeContent(content), tagName)
			}
		} else {
			// Self-closing block or empty block
			attributeString := formatAttributes(block.GetAttributes())
			if attributeString != "" {
				return fmt.Sprintf("%s<%s%s />", indent, tagName, attributeString)
			} else {
				return fmt.Sprintf("%s<%s />", indent, tagName)
			}
		}
	}

	// Block with children
	attributeString := formatAttributes(block.GetAttributes())
	openTag := fmt.Sprintf("%s<%s%s>", indent, tagName, attributeString)
	closeTag := fmt.Sprintf("%s</%s>", indent, tagName)

	// Process children
	var childrenMJML []string
	for _, child := range children {
		if child != nil {
			childrenMJML = append(childrenMJML, convertBlockToMJML(child, indentLevel+1, templateData))
		}
	}

	return fmt.Sprintf("%s\n%s\n%s", openTag, strings.Join(childrenMJML, "\n"), closeTag)
}

// processLiquidContent processes Liquid templating in content
func processLiquidContent(content, templateData, blockID string) (string, error) {
	// Check if content contains Liquid templating markup
	if !strings.Contains(content, "{{") && !strings.Contains(content, "{%") {
		return content, nil // No Liquid markup found, return original content
	}

	// Create Liquid engine
	engine := liquid.NewEngine()

	// Parse template data JSON
	var jsonData map[string]interface{}
	if templateData != "" {
		err := json.Unmarshal([]byte(templateData), &jsonData)
		if err != nil {
			return content, fmt.Errorf("invalid JSON in templateData for block (ID: %s): %w", blockID, err)
		}
	} else {
		// Initialize empty map if templateData is empty
		jsonData = make(map[string]interface{})
	}

	// Render the content with Liquid
	renderedContent, err := engine.ParseAndRenderString(content, jsonData)
	if err != nil {
		return content, fmt.Errorf("liquid rendering error in block (ID: %s): %w", blockID, err)
	}

	return renderedContent, nil
}

// getBlockContent extracts content from a block using type assertion
func getBlockContent(block EmailBlock) string {
	switch v := block.(type) {
	case *MJTextBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJButtonBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJRawBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJPreviewBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJStyleBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJTitleBlock:
		if v.Content != nil {
			return *v.Content
		}
	case *MJSocialElementBlock:
		if v.Content != nil {
			return *v.Content
		}
	}
	return ""
}

// formatAttributes formats attributes object into MJML attribute string
func formatAttributes(attributes map[string]interface{}) string {
	if len(attributes) == 0 {
		return ""
	}

	var attrPairs []string
	for key, value := range attributes {
		if shouldIncludeAttribute(value) {
			if attr := formatSingleAttribute(key, value); attr != "" {
				attrPairs = append(attrPairs, attr)
			}
		}
	}

	return strings.Join(attrPairs, "")
}

// shouldIncludeAttribute determines if an attribute value should be included in the output
func shouldIncludeAttribute(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return v != ""
	case *string:
		return v != nil && *v != ""
	case bool:
		return true // Include boolean attributes regardless of value
	case *bool:
		return v != nil
	case int, int32, int64, float32, float64:
		return true // Include numeric values
	default:
		return fmt.Sprintf("%v", value) != ""
	}
}

// formatSingleAttribute formats a single attribute key-value pair
func formatSingleAttribute(key string, value interface{}) string {
	// Convert camelCase to kebab-case for MJML attributes
	kebabKey := camelToKebab(key)

	// Handle different value types
	switch v := value.(type) {
	case bool:
		if v {
			return fmt.Sprintf(" %s", kebabKey)
		}
		return ""
	case *bool:
		if v != nil && *v {
			return fmt.Sprintf(" %s", kebabKey)
		}
		return ""
	case string:
		if v == "" {
			return ""
		}
		escapedValue := escapeAttributeValue(v, kebabKey)
		return fmt.Sprintf(` %s="%s"`, kebabKey, escapedValue)
	case *string:
		if v == nil || *v == "" {
			return ""
		}
		escapedValue := escapeAttributeValue(*v, kebabKey)
		return fmt.Sprintf(` %s="%s"`, kebabKey, escapedValue)
	default:
		// Handle other types (int, float, etc.) by converting to string
		strValue := fmt.Sprintf("%v", value)
		if strValue == "" {
			return ""
		}
		escapedValue := escapeAttributeValue(strValue, kebabKey)
		return fmt.Sprintf(` %s="%s"`, kebabKey, escapedValue)
	}
}

// camelToKebab converts camelCase to kebab-case
func camelToKebab(str string) string {
	// Use regex to find capital letters and replace them with hyphen + lowercase
	re := regexp.MustCompile("([A-Z])")
	return re.ReplaceAllStringFunc(str, func(match string) string {
		return "-" + strings.ToLower(match)
	})
}

// escapeAttributeValue escapes attribute values for safe HTML output
// For URL attributes (src, href, action), we don't escape & to preserve URL query parameters
func escapeAttributeValue(value string, attributeName string) string {
	// Check if this is a URL attribute and the value looks like a URL
	isURLAttribute := attributeName == "src" || attributeName == "href" || attributeName == "action"
	looksLikeURL := strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "//")

	// Only skip escaping ampersands if it's a URL attribute AND the value looks like a URL
	if !(isURLAttribute && looksLikeURL) {
		value = strings.ReplaceAll(value, "&", "&amp;")
	}
	value = strings.ReplaceAll(value, "\"", "&quot;")
	value = strings.ReplaceAll(value, "'", "&#39;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	return value
}

// escapeContent escapes content for safe HTML output
func escapeContent(content string) string {
	content = strings.ReplaceAll(content, "&", "&amp;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	return content
}

// ConvertToMJMLString is a convenience function that converts an EmailBlock to MJML
// and wraps it in a complete MJML document structure if needed
func ConvertToMJMLString(block EmailBlock) (string, error) {
	return ConvertToMJMLStringWithData(block, "")
}

// ConvertToMJMLStringWithData converts an EmailBlock to MJML with template data
func ConvertToMJMLStringWithData(block EmailBlock, templateData string) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block cannot be nil")
	}

	// If the root block is not MJML, we need to validate the structure
	if block.GetType() != MJMLComponentMjml {
		return "", fmt.Errorf("root block must be of type 'mjml', got '%s'", block.GetType())
	}

	// Validate the email structure before converting
	if err := ValidateEmailStructure(block); err != nil {
		return "", fmt.Errorf("invalid email structure: %w", err)
	}

	return ConvertJSONToMJMLWithData(block, templateData)
}

// ConvertToMJMLWithOptions provides additional options for MJML conversion
type MJMLConvertOptions struct {
	Validate      bool   // Whether to validate the structure before converting
	PrettyPrint   bool   // Whether to format with proper indentation (always true for now)
	IncludeXMLTag bool   // Whether to include XML declaration at the beginning
	TemplateData  string // JSON string containing template data for Liquid processing
}

// ConvertToMJMLWithOptions converts an EmailBlock to MJML string with additional options
func ConvertToMJMLWithOptions(block EmailBlock, options MJMLConvertOptions) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block cannot be nil")
	}

	// Validate if requested
	if options.Validate {
		if err := ValidateEmailStructure(block); err != nil {
			return "", fmt.Errorf("validation failed: %w", err)
		}
	}

	// Convert to MJML with template data
	mjml, err := ConvertJSONToMJMLWithData(block, options.TemplateData)
	if err != nil {
		return "", fmt.Errorf("mjml conversion failed: %w", err)
	}

	// Add XML declaration if requested
	if options.IncludeXMLTag {
		mjml = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + mjml
	}

	return mjml, nil
}
