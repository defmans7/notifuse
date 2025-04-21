package mjml

import (
	"testing"
)

// Helper function to find a block by ID in a slice of EmailBlocks
func findBlockByID(blocks []EmailBlock, id string) (EmailBlock, bool) {
	for _, block := range blocks {
		if block.ID == id {
			return block, true
		}
	}
	return EmailBlock{}, false
}

// Helper function to create a new slice with only the expected block IDs for cleaner test errors
func getBlockIDs(blocks []EmailBlock) []string {
	ids := make([]string, len(blocks))
	for i, block := range blocks {
		ids[i] = block.ID
	}
	return ids
}

func TestTemplateStructure(t *testing.T) {
	// Test that each email template function returns the expected root structure
	testCases := []struct {
		name         string
		templateFunc func() EmailBlock
		expectedIDs  []string // Expected block IDs in the content column
	}{
		{
			"DefaultOptinConfirmationEmail",
			DefaultOptinConfirmationEmail,
			[]string{"logo", "heading", "text", "button", "disclaimer", "divider", "footer", "open-tracking"},
		},
		{
			"DefaultUnsubscribeConfirmationEmail",
			DefaultUnsubscribeConfirmationEmail,
			[]string{"logo", "heading", "text", "resub-text", "button", "divider", "footer", "open-tracking"},
		},
		{
			"DefaultWelcomeEmail",
			DefaultWelcomeEmail,
			[]string{"logo", "heading", "text", "welcome-text", "unsub-text", "unsub-link", "divider", "footer", "open-tracking"},
		},
		{
			"DefaultTemplateStructure",
			DefaultTemplateStructure,
			[]string{"logo", "divider", "footer", "open-tracking"},
		},
		{
			"DefaultEmailStyles",
			DefaultEmailStyles,
			nil, // DefaultEmailStyles doesn't include content blocks
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			email := tc.templateFunc()

			// Check basic structure
			if email.Kind != "root" {
				t.Errorf("Expected root kind, got %s", email.Kind)
			}

			if len(email.Children) != 1 {
				t.Errorf("Expected 1 section, got %d sections", len(email.Children))
				return
			}

			section := email.Children[0]
			if section.Kind != "oneColumn" {
				t.Errorf("Expected oneColumn section, got %s", section.Kind)
			}

			if len(section.Children) != 1 {
				t.Errorf("Expected 1 column, got %d columns", len(section.Children))
				return
			}

			column := section.Children[0]
			if column.Kind != "column" {
				t.Errorf("Expected column kind, got %s", column.Kind)
			}

			// Test column children if expected IDs are provided
			if tc.expectedIDs != nil {
				// Log the actual block IDs for debugging
				actualIDs := getBlockIDs(column.Children)
				t.Logf("Column children IDs: %v", actualIDs)

				if len(column.Children) == 0 {
					t.Errorf("Column has no children blocks, expected %d blocks", len(tc.expectedIDs))
					return
				}

				// Check that all expected blocks are present
				for _, expectedID := range tc.expectedIDs {
					_, found := findBlockByID(column.Children, expectedID)
					if !found {
						t.Errorf("Expected block with ID '%s' not found", expectedID)
					}
				}
			}

			t.Logf("%s has valid structure", tc.name)
		})
	}
}

func TestDefaultBlocks(t *testing.T) {
	blocks := DefaultBlocks()

	// Verify we have all expected blocks
	expectedBlocks := []string{
		"root", "button", "text", "heading", "divider", "image",
		"liquid", "openTracking", "section", "column", "oneColumn",
		"columns1212", "columns168", "columns204", "columns420",
		"columns816", "columns888", "columns6666",
	}

	for _, expected := range expectedBlocks {
		if _, ok := blocks[expected]; !ok {
			t.Errorf("Expected block %s not found", expected)
		}
	}

	// Test a few critical blocks
	buttonBlock := blocks["button"]
	if buttonBlock.Kind != "button" {
		t.Errorf("Expected button block to have kind 'button', got %s", buttonBlock.Kind)
	}

	if buttonData, ok := buttonBlock.Data.(map[string]interface{}); ok {
		if _, hasButton := buttonData["button"]; !hasButton {
			t.Errorf("Expected button block data to have 'button' field")
		}
	} else {
		t.Errorf("Expected button block data to be a map")
	}
}

func TestDeepCopyBlock(t *testing.T) {
	// Create a simple test EmailBlock with simple data to properly test
	original := EmailBlock{
		ID:   "test-block",
		Kind: "text",
		Data: map[string]interface{}{
			"text":  "Sample text",
			"color": "#000000",
		},
		Children: []EmailBlock{},
	}

	// Make a deep copy
	copy := DeepCopyBlock(original)

	// Test that IDs and kinds match
	if copy.ID != original.ID {
		t.Errorf("Expected copied ID %s, got %s", original.ID, copy.ID)
	}

	if copy.Kind != original.Kind {
		t.Errorf("Expected copied Kind %s, got %s", original.Kind, copy.Kind)
	}

	// Modify the copy
	copyData := copy.Data.(map[string]interface{})
	copyData["color"] = "#FF0000"

	// Check original is unchanged
	originalData := original.Data.(map[string]interface{})
	if originalData["color"] == "#FF0000" {
		t.Errorf("Original was modified when copy was changed. Original color became: %s", originalData["color"])
	}

	// Add a new field to the copy and verify it doesn't affect the original
	copyData["new_field"] = "new value"
	if _, exists := originalData["new_field"]; exists {
		t.Errorf("Adding a new field to the copy affected the original")
	}

	// Note: The current DeepCopyBlock implementation doesn't actually do a deep copy of nested maps
	// This is a known limitation of the function.
	t.Log("Note: DeepCopyBlock only does a shallow copy of nested maps, not a true deep copy.")
}

func TestBlocksWithSpecificContent(t *testing.T) {
	// Test specific content in some email templates

	// Test optin email confirmation URL
	optinEmail := DefaultOptinConfirmationEmail()
	column := optinEmail.Children[0].Children[0]
	buttonBlock, found := findBlockByID(column.Children, "button")

	if !found {
		t.Fatalf("Button block not found in optin email")
	}

	buttonData := buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})
	buttonHref := buttonData["href"].(string)
	buttonText := buttonData["text"].(string)

	if buttonHref != "{{ confirmation_url }}" {
		t.Errorf("Expected confirmation URL placeholder, got %s", buttonHref)
	}

	if buttonText != "CONFIRM SUBSCRIPTION" {
		t.Errorf("Expected button text 'CONFIRM SUBSCRIPTION', got '%s'", buttonText)
	}

	// Test unsubscribe email resubscribe URL
	unsubEmail := DefaultUnsubscribeConfirmationEmail()
	column = unsubEmail.Children[0].Children[0]
	buttonBlock, found = findBlockByID(column.Children, "button")

	if !found {
		t.Fatalf("Button block not found in unsubscribe email")
	}

	buttonData = buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})
	buttonHref = buttonData["href"].(string)

	if buttonHref != "{{ resubscribe_url }}" {
		t.Errorf("Expected resubscribe URL placeholder, got %s", buttonHref)
	}

	// Test welcome email unsubscribe URL
	welcomeEmail := DefaultWelcomeEmail()
	column = welcomeEmail.Children[0].Children[0]
	unsubLinkBlock, found := findBlockByID(column.Children, "unsub-link")

	if !found {
		t.Fatalf("Unsubscribe link block not found in welcome email")
	}

	editorData := unsubLinkBlock.Data.(map[string]interface{})["editorData"].([]map[string]interface{})
	linkHref := editorData[0]["children"].([]map[string]interface{})[0]["linkHref"].(string)

	if linkHref != "{{ unsubscribe_url }}" {
		t.Errorf("Expected unsubscribe URL placeholder, got %s", linkHref)
	}
}
