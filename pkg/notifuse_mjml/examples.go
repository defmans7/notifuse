package notifuse_mjml

import (
	"fmt"
	"time"
)

// NewBaseBlock creates a new base block with default values
func NewBaseBlock(id string, componentType MJMLComponentType) BaseBlock {
	return BaseBlock{
		ID:         id,
		Type:       componentType,
		Children:   make([]interface{}, 0),
		Attributes: GetDefaultAttributes(componentType),
	}
}

// CreateSimpleEmail creates a basic MJML email structure
func CreateSimpleEmail() *MJMLBlock {
	// Create head section
	head := &MJHeadBlock{
		BaseBlock: NewBaseBlock("head-1", MJMLComponentMjHead),
	}

	// Add title to head
	title := &MJTitleBlock{
		BaseBlock: NewBaseBlock("title-1", MJMLComponentMjTitle),
		Content:   stringPtr("Welcome Email"),
	}
	head.BaseBlock.Children = []interface{}{title}

	// Create body section
	body := &MJBodyBlock{
		BaseBlock: NewBaseBlock("body-1", MJMLComponentMjBody),
		Attributes: &MJBodyAttributes{
			BackgroundAttributes: BackgroundAttributes{
				BackgroundColor: stringPtr("#f4f4f4"),
			},
			Width: stringPtr("600px"),
		},
	}

	// Create section
	section := &MJSectionBlock{
		BaseBlock: NewBaseBlock("section-1", MJMLComponentMjSection),
		Attributes: &MJSectionAttributes{
			BackgroundAttributes: BackgroundAttributes{
				BackgroundColor: stringPtr("#ffffff"),
			},
			PaddingAttributes: PaddingAttributes{
				PaddingTop:    stringPtr("20px"),
				PaddingBottom: stringPtr("20px"),
			},
		},
	}

	// Create column
	column := &MJColumnBlock{
		BaseBlock: NewBaseBlock("column-1", MJMLComponentMjColumn),
		Attributes: &MJColumnAttributes{
			PaddingAttributes: PaddingAttributes{
				PaddingLeft:  stringPtr("20px"),
				PaddingRight: stringPtr("20px"),
			},
		},
	}

	// Create text block
	textBlock := &MJTextBlock{
		BaseBlock: NewBaseBlock("text-1", MJMLComponentMjText),
		Content:   stringPtr("Welcome to our newsletter!"),
		Attributes: &MJTextAttributes{
			TextAttributes: TextAttributes{
				FontSize:   stringPtr("16px"),
				LineHeight: stringPtr("1.5"),
				Color:      stringPtr("#333333"),
				Align:      stringPtr("center"),
			},
		},
	}

	// Create button
	button := &MJButtonBlock{
		BaseBlock: NewBaseBlock("button-1", MJMLComponentMjButton),
		Content:   stringPtr("Get Started"),
		Attributes: &MJButtonAttributes{
			BackgroundAttributes: BackgroundAttributes{
				BackgroundColor: stringPtr("#007bff"),
			},
			TextAttributes: TextAttributes{
				Color:      stringPtr("#ffffff"),
				FontWeight: stringPtr("bold"),
			},
			BorderAttributes: BorderAttributes{
				BorderRadius: stringPtr("5px"),
			},
			LinkAttributes: LinkAttributes{
				Href: stringPtr("https://example.com"),
			},
			PaddingAttributes: PaddingAttributes{
				PaddingTop:    stringPtr("10px"),
				PaddingBottom: stringPtr("10px"),
			},
		},
	}

	// Assemble the structure
	column.BaseBlock.Children = []interface{}{textBlock, button}
	section.BaseBlock.Children = []interface{}{column}
	body.BaseBlock.Children = []interface{}{section}

	// Create root MJML block
	mjml := &MJMLBlock{
		BaseBlock: NewBaseBlock("mjml-1", MJMLComponentMjml),
	}
	mjml.BaseBlock.Children = []interface{}{head, body}

	return mjml
}

// CreateEmailWithImage creates an email with an image component
func CreateEmailWithImage() *MJMLBlock {
	mjml := CreateSimpleEmail()

	// Find the body and add an image section
	if len(mjml.BaseBlock.Children) > 1 {
		if body, ok := mjml.BaseBlock.Children[1].(*MJBodyBlock); ok {
			// Create new section with image
			imageSection := &MJSectionBlock{
				BaseBlock: NewBaseBlock("image-section-1", MJMLComponentMjSection),
				Attributes: &MJSectionAttributes{
					BackgroundAttributes: BackgroundAttributes{
						BackgroundColor: stringPtr("#ffffff"),
					},
				},
			}

			imageColumn := &MJColumnBlock{
				BaseBlock: NewBaseBlock("image-column-1", MJMLComponentMjColumn),
			}

			image := &MJImageBlock{
				BaseBlock: NewBaseBlock("image-1", MJMLComponentMjImage),
				Attributes: &MJImageAttributes{
					Src:           stringPtr("https://via.placeholder.com/600x300"),
					Alt:           stringPtr("Placeholder Image"),
					FluidOnMobile: stringPtr("true"),
					LayoutAttributes: LayoutAttributes{
						Width: stringPtr("600px"),
					},
				},
			}

			imageColumn.BaseBlock.Children = []interface{}{image}
			imageSection.BaseBlock.Children = []interface{}{imageColumn}

			// Insert image section before the existing section
			body.BaseBlock.Children = append([]interface{}{imageSection}, body.BaseBlock.Children...)
		}
	}

	return mjml
}

// CreateSocialEmail creates an email with social media links
func CreateSocialEmail() *MJMLBlock {
	mjml := CreateSimpleEmail()

	// Find the body and add a social section
	if len(mjml.BaseBlock.Children) > 1 {
		if body, ok := mjml.BaseBlock.Children[1].(*MJBodyBlock); ok {
			// Create social section
			socialSection := &MJSectionBlock{
				BaseBlock: NewBaseBlock("social-section-1", MJMLComponentMjSection),
				Attributes: &MJSectionAttributes{
					BackgroundAttributes: BackgroundAttributes{
						BackgroundColor: stringPtr("#f8f9fa"),
					},
					PaddingAttributes: PaddingAttributes{
						PaddingTop:    stringPtr("30px"),
						PaddingBottom: stringPtr("30px"),
					},
				},
			}

			socialColumn := &MJColumnBlock{
				BaseBlock: NewBaseBlock("social-column-1", MJMLComponentMjColumn),
			}

			socialBlock := &MJSocialBlock{
				BaseBlock: NewBaseBlock("social-1", MJMLComponentMjSocial),
				Attributes: &MJSocialAttributes{
					Align:        stringPtr("center"),
					IconSize:     stringPtr("40px"),
					Mode:         stringPtr("horizontal"),
					InnerPadding: stringPtr("4px"),
				},
			}

			// Add social elements
			facebookElement := &MJSocialElementBlock{
				BaseBlock: NewBaseBlock("facebook-1", MJMLComponentMjSocialElement),
				Attributes: &MJSocialElementAttributes{
					Name:            stringPtr("facebook"),
					Href:            stringPtr("https://facebook.com"),
					BackgroundColor: stringPtr("#1877f2"),
				},
			}

			twitterElement := &MJSocialElementBlock{
				BaseBlock: NewBaseBlock("twitter-1", MJMLComponentMjSocialElement),
				Attributes: &MJSocialElementAttributes{
					Name:            stringPtr("twitter"),
					Href:            stringPtr("https://twitter.com"),
					BackgroundColor: stringPtr("#1da1f2"),
				},
			}

			socialBlock.BaseBlock.Children = []interface{}{facebookElement, twitterElement}
			socialColumn.BaseBlock.Children = []interface{}{socialBlock}
			socialSection.BaseBlock.Children = []interface{}{socialColumn}

			// Add social section to the end
			body.BaseBlock.Children = append(body.BaseBlock.Children, socialSection)
		}
	}

	return mjml
}

// ConvertToEmailBuilderState converts an MJML structure to EmailBuilderState
func ConvertToEmailBuilderState(mjml EmailBlock) *EmailBuilderState {
	return &EmailBuilderState{
		SelectedBlockID: nil,
		History:         []EmailBlock{mjml},
		HistoryIndex:    0,
		ViewportMode:    stringPtr("desktop"),
	}
}

// CreateSavedBlock creates a saved block for storage
func CreateSavedBlock(id, name string, block EmailBlock) *SavedBlock {
	now := time.Now()
	return &SavedBlock{
		ID:      id,
		Name:    name,
		Block:   block,
		Created: &now,
		Updated: &now,
	}
}

// PrintEmailStructure prints the structure of an email for debugging
func PrintEmailStructure(block EmailBlock, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	fmt.Printf("%s%s (ID: %s)\n", indent, GetComponentDisplayName(block.GetType()), block.GetID())

	for _, child := range block.GetChildren() {
		if child != nil {
			PrintEmailStructure(child, depth+1)
		}
	}
}

// ValidateAndPrintEmail validates an email structure and prints any errors
func ValidateAndPrintEmail(email EmailBlock) {
	fmt.Printf("Email Structure:\n")
	PrintEmailStructure(email, 0)

	fmt.Printf("\nValidation:\n")
	if err := ValidateEmailStructure(email); err != nil {
		fmt.Printf("❌ Validation failed: %s\n", err)
	} else {
		fmt.Printf("✅ Email structure is valid\n")
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

// DemoConverter demonstrates the MJML conversion functionality
func DemoConverter() {
	fmt.Println("=== MJML Converter Demo ===")

	// Create a simple email
	email := CreateSimpleEmail()

	// Convert to MJML string
	mjmlString, err := ConvertToMJMLString(email)
	if err != nil {
		fmt.Printf("❌ Conversion error: %s\n", err)
		return
	}

	fmt.Printf("✅ Successfully converted to MJML:\n\n%s\n", mjmlString)

	// Demonstrate conversion with options
	fmt.Println("\n=== Conversion with Options ===")
	options := MJMLConvertOptions{
		Validate:      true,
		PrettyPrint:   true,
		IncludeXMLTag: true,
	}

	mjmlWithOptions, err := ConvertToMJMLWithOptions(email, options)
	if err != nil {
		fmt.Printf("❌ Conversion with options error: %s\n", err)
		return
	}

	fmt.Printf("✅ MJML with XML declaration:\n\n%s\n", mjmlWithOptions)
}

// ConvertEmailToMJMLDemo creates an email and shows the MJML output
func ConvertEmailToMJMLDemo() {
	// Create a more complex email with social elements
	email := CreateSocialEmail()

	fmt.Println("=== Email Structure ===")
	PrintEmailStructure(email, 0)

	fmt.Println("\n=== Generated MJML ===")
	mjml := ConvertJSONToMJML(email)
	fmt.Println(mjml)

	fmt.Println("\n=== Validation ===")
	if err := ValidateEmailStructure(email); err != nil {
		fmt.Printf("❌ Validation failed: %s\n", err)
	} else {
		fmt.Printf("✅ Email structure is valid\n")
	}
}

// TestConverterFunctions demonstrates individual converter functions
func TestConverterFunctions() {
	fmt.Println("=== Testing Individual Converter Functions ===")

	// Test camelToKebab conversion
	testCases := []string{
		"backgroundColor",
		"fontSize",
		"paddingTop",
		"fullWidthBackgroundColor",
		"innerBorderRadius",
	}

	fmt.Println("CamelCase to kebab-case conversion:")
	for _, test := range testCases {
		kebab := camelToKebab(test)
		fmt.Printf("  %s -> %s\n", test, kebab)
	}

	// Test attribute escaping
	fmt.Println("\nAttribute value escaping:")
	testValues := []string{
		"Hello & Goodbye",
		"<script>alert('test')</script>",
		`He said "Hello"`,
		"It's a test",
	}

	for _, test := range testValues {
		escaped := escapeAttributeValue(test, "title")
		fmt.Printf("  %s -> %s\n", test, escaped)
	}

	// Test content escaping
	fmt.Println("\nContent escaping:")
	testContent := []string{
		"<b>Bold text</b>",
		"A & B > C",
		"<script>alert('xss')</script>",
	}

	for _, test := range testContent {
		escaped := escapeContent(test)
		fmt.Printf("  %s -> %s\n", test, escaped)
	}
}
