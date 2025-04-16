package mjml

import (
	"strings"
	"testing"
)

func TestDefaultOptinConfirmationEmail(t *testing.T) {
	email := DefaultOptinConfirmationEmail()

	// Check basic structure
	if email.Kind != "root" {
		t.Errorf("Expected root block, got %s", email.Kind)
	}
	if len(email.Children) != 1 {
		t.Errorf("Expected 1 child in root, got %d", len(email.Children))
	}

	// Check section
	section := email.Children[0]
	if section.Kind != "oneColumn" {
		t.Errorf("Expected oneColumn section, got %s", section.Kind)
	}
	if len(section.Children) != 1 {
		t.Errorf("Expected 1 child in section, got %d", len(section.Children))
	}

	// Check column
	column := section.Children[0]
	if column.Kind != "column" {
		t.Errorf("Expected column block, got %s", column.Kind)
	}

	// Check for required elements
	var hasLogo, hasHeading, hasText, hasButton bool
	for _, block := range column.Children {
		switch block.ID {
		case "logo":
			hasLogo = true
			if block.Kind != "image" {
				t.Errorf("Logo should be an image block, got %s", block.Kind)
			}
		case "heading":
			hasHeading = true
			if block.Kind != "heading" {
				t.Errorf("Heading should be a heading block, got %s", block.Kind)
			}
		case "text":
			hasText = true
			if block.Kind != "text" {
				t.Errorf("Text should be a text block, got %s", block.Kind)
			}
		case "button":
			hasButton = true
			if block.Kind != "button" {
				t.Errorf("Button should be a button block, got %s", block.Kind)
			}
			// Check button properties
			buttonData, ok := block.Data.(map[string]interface{})
			if !ok {
				t.Error("Button data should be a map")
			} else {
				buttonInfo, ok := buttonData["button"].(map[string]interface{})
				if !ok {
					t.Error("Button info should be a map")
				} else {
					// Check confirmation URL
					href, ok := buttonInfo["href"].(string)
					if !ok || !strings.Contains(href, "confirmation_url") {
						t.Errorf("Button should link to confirmation URL, got %v", href)
					}
				}
			}
		}
	}

	if !hasLogo {
		t.Error("Email should contain a logo block")
	}
	if !hasHeading {
		t.Error("Email should contain a heading block")
	}
	if !hasText {
		t.Error("Email should contain a text block")
	}
	if !hasButton {
		t.Error("Email should contain a button block")
	}

	// Test MJML generation
	mjml, err := TreeToMjml(map[string]interface{}{}, email, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("Error generating MJML: %v", err)
	}
	if !strings.Contains(mjml, "<mjml>") {
		t.Error("Generated MJML should contain <mjml> tag")
	}
}

func TestDefaultUnsubscribeConfirmationEmail(t *testing.T) {
	email := DefaultUnsubscribeConfirmationEmail()

	// Check basic structure
	if email.Kind != "root" {
		t.Errorf("Expected root block, got %s", email.Kind)
	}

	// Check for required elements
	var hasLogo, hasHeading, hasText bool
	for _, section := range email.Children {
		for _, column := range section.Children {
			for _, block := range column.Children {
				switch block.ID {
				case "logo":
					hasLogo = true
				case "heading":
					hasHeading = true
				case "text":
					hasText = true
				}
			}
		}
	}

	if !hasLogo {
		t.Error("Unsubscribe email should contain a logo block")
	}
	if !hasHeading {
		t.Error("Unsubscribe email should contain a heading block")
	}
	if !hasText {
		t.Error("Unsubscribe email should contain a text block")
	}

	// Test MJML generation
	mjml, err := TreeToMjml(map[string]interface{}{}, email, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("Error generating MJML: %v", err)
	}
	if !strings.Contains(mjml, "You") && !strings.Contains(mjml, "Unsubscribed") {
		t.Error("Generated MJML should contain unsubscribe heading text")
	}
}

func TestDefaultWelcomeEmail(t *testing.T) {
	email := DefaultWelcomeEmail()

	// Check basic structure
	if email.Kind != "root" {
		t.Errorf("Expected root block, got %s", email.Kind)
	}

	// Check for required elements
	var hasLogo, hasHeading, hasText, hasBulletPoints bool
	bulletPointCount := 0

	for _, section := range email.Children {
		for _, column := range section.Children {
			for _, block := range column.Children {
				switch block.ID {
				case "logo":
					hasLogo = true
				case "heading":
					hasHeading = true
				case "intro", "expectation":
					hasText = true
				case "bullet1", "bullet2", "bullet3":
					bulletPointCount++
				}
			}
		}
	}

	hasBulletPoints = bulletPointCount >= 3

	if !hasLogo {
		t.Error("Welcome email should contain a logo block")
	}
	if !hasHeading {
		t.Error("Welcome email should contain a heading block")
	}
	if !hasText {
		t.Error("Welcome email should contain text blocks")
	}
	if !hasBulletPoints {
		t.Error("Welcome email should contain bullet points")
	}

	// Test MJML generation
	mjml, err := TreeToMjml(map[string]interface{}{}, email, "", map[string]string{}, 0, nil)
	if err != nil {
		t.Fatalf("Error generating MJML: %v", err)
	}
	if !strings.Contains(mjml, "Welcome to Our Community") {
		t.Error("Generated MJML should contain welcome heading")
	}
	if !strings.Contains(mjml, "unsubscribe") {
		t.Error("Generated MJML should contain unsubscribe link")
	}
}
