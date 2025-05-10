package mjml

func NewOpenTrackingBlock() EmailBlock {
	return EmailBlock{
		ID:       "openTracking",
		Kind:     "openTracking",
		Data:     map[string]interface{}{},
		Children: []EmailBlock{},
	}
}

// DeepCopyBlock creates a deep copy of an EmailBlock to prevent modifying original blocks by reference
func DeepCopyBlock(block EmailBlock) EmailBlock {
	// Create a new block with the same ID and Kind
	newBlock := EmailBlock{
		ID:       block.ID,
		Kind:     block.Kind,
		Children: make([]EmailBlock, 0, len(block.Children)),
	}

	// Deep copy the data
	if block.Data != nil {
		// For data, we need to create a new map
		dataMap, isMap := block.Data.(map[string]interface{})
		if isMap {
			newDataMap := make(map[string]interface{})
			for k, v := range dataMap {
				newDataMap[k] = v
			}
			newBlock.Data = newDataMap
		} else {
			// If it's not a map, just assign the value directly
			newBlock.Data = block.Data
		}
	}

	// Deep copy children recursively
	for _, child := range block.Children {
		newBlock.Children = append(newBlock.Children, DeepCopyBlock(child))
	}

	return newBlock
}

// DefaultTemplateStructure creates a basic email template structure with some standard blocks
func DefaultTemplateStructure() EmailBlock {
	blocks := DefaultBlocks()

	// Get the default styled email template
	emailTemplate := DefaultEmailStyles()

	// Access the content column
	contentColumn := emailTemplate.Children[0].Children[0]

	// Initialize children array with standard blocks
	newChildren := []EmailBlock{}

	// Add standard logo
	logoBlock := DeepCopyBlock(blocks["image"])
	logoBlock.ID = "logo"
	logoBlock.Data.(map[string]interface{})["image"].(map[string]interface{})["src"] = "https://placehold.co/100x50?text=Your+Logo"
	logoBlock.Data.(map[string]interface{})["image"].(map[string]interface{})["alt"] = "Logo"
	logoBlock.Data.(map[string]interface{})["image"].(map[string]interface{})["width"] = "100px"
	logoBlock.Data.(map[string]interface{})["wrapper"].(map[string]interface{})["align"] = "center"
	newChildren = append(newChildren, logoBlock)

	// Add standard footer
	dividerBlock := DeepCopyBlock(blocks["divider"])
	dividerBlock.ID = "divider"
	dividerBlock.Data.(map[string]interface{})["paddingControl"] = "separate"
	dividerBlock.Data.(map[string]interface{})["paddingTop"] = "40px"
	dividerBlock.Data.(map[string]interface{})["paddingBottom"] = "20px"
	dividerBlock.Data.(map[string]interface{})["paddingLeft"] = "200px"
	dividerBlock.Data.(map[string]interface{})["paddingRight"] = "200px"
	newChildren = append(newChildren, dividerBlock)

	// Add standard copyright footer
	footerBlock := DeepCopyBlock(blocks["text"])
	footerBlock.ID = "footer"
	footerBlock.Data.(map[string]interface{})["align"] = "center"
	footerBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{
					"text":     "© {{ current_year }} Your Company. All rights reserved.",
					"color":    "#666666",
					"fontSize": "12px",
				},
			},
		},
	}
	newChildren = append(newChildren, footerBlock)

	// Add tracking pixel
	openTrackingBlock := DeepCopyBlock(blocks["openTracking"])
	openTrackingBlock.ID = "open-tracking"
	newChildren = append(newChildren, openTrackingBlock)

	// Create a new column with the children properly set
	updatedColumn := DeepCopyBlock(contentColumn)
	updatedColumn.Children = newChildren

	// Create a new section with the updated column
	updatedSection := DeepCopyBlock(emailTemplate.Children[0])
	updatedSection.Children = []EmailBlock{updatedColumn}

	// Create a new root with the updated section
	updatedRoot := DeepCopyBlock(emailTemplate)
	updatedRoot.Children = []EmailBlock{updatedSection}

	return updatedRoot
}

// DefaultOptinConfirmationEmail returns an EmailBlock tree for an opt-in confirmation email
func DefaultOptinConfirmationEmail() EmailBlock {
	blocks := DefaultBlocks()

	// Get the template with standard structure
	emailTemplate := DefaultTemplateStructure()

	// Access the content column to add our specific blocks
	contentSection := emailTemplate.Children[0]
	contentColumn := contentSection.Children[0]

	// Store the standard blocks (logo, divider, footer, tracking)
	standardBlocks := contentColumn.Children

	// Initialize new children array
	newChildren := []EmailBlock{}

	// Add logo if standard blocks exist
	if len(standardBlocks) > 0 {
		newChildren = append(newChildren, standardBlocks[0]) // Add logo
	}

	// Add heading
	headingBlock := DeepCopyBlock(blocks["heading"])
	headingBlock.ID = "heading"
	headingBlock.Data.(map[string]interface{})["type"] = "h2"
	headingBlock.Data.(map[string]interface{})["align"] = "center"
	headingBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "h2",
			"children": []map[string]interface{}{
				{"text": "Please Confirm Your Subscription"},
			},
		},
	}
	newChildren = append(newChildren, headingBlock)

	// Add text content
	textBlock := DeepCopyBlock(blocks["text"])
	textBlock.ID = "text"
	textBlock.Data.(map[string]interface{})["align"] = "center"
	textBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "Thank you for subscribing to our newsletter. To complete your subscription, please click the button below to confirm your email address."},
			},
		},
	}
	newChildren = append(newChildren, textBlock)

	// Add confirm button
	buttonBlock := DeepCopyBlock(blocks["button"])
	buttonBlock.ID = "button"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["backgroundColor"] = "#4e6cff"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["href"] = "{{ confirmation_url }}"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["text"] = "CONFIRM SUBSCRIPTION"
	newChildren = append(newChildren, buttonBlock)

	// Add disclaimer
	disclaimerBlock := DeepCopyBlock(blocks["text"])
	disclaimerBlock.ID = "disclaimer"
	disclaimerBlock.Data.(map[string]interface{})["align"] = "center"
	disclaimerBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{
					"text":     "If you did not request this subscription, you can safely ignore this email.",
					"color":    "#666666",
					"fontSize": "14px",
				},
			},
		},
	}
	newChildren = append(newChildren, disclaimerBlock)

	// Add back the standard blocks (divider, footer, tracking) if they exist
	if len(standardBlocks) > 1 {
		newChildren = append(newChildren, standardBlocks[1:]...)
	}

	// Create a new column with the children properly set
	updatedColumn := DeepCopyBlock(contentColumn)
	updatedColumn.Children = newChildren

	// Create a new section with the updated column
	updatedSection := DeepCopyBlock(contentSection)
	updatedSection.Children = []EmailBlock{updatedColumn}

	// Create a new root with the updated section
	updatedRoot := DeepCopyBlock(emailTemplate)
	updatedRoot.Children = []EmailBlock{updatedSection}

	return updatedRoot
}

// DefaultUnsubscribeConfirmationEmail returns an EmailBlock tree for an unsubscribe confirmation email
func DefaultUnsubscribeConfirmationEmail() EmailBlock {
	blocks := DefaultBlocks()

	// Get the template with standard structure
	emailTemplate := DefaultTemplateStructure()

	// Access the content column to add our specific blocks
	contentSection := emailTemplate.Children[0]
	contentColumn := contentSection.Children[0]

	// Store the standard blocks (logo, divider, footer, tracking)
	standardBlocks := contentColumn.Children

	// Initialize new children array
	newChildren := []EmailBlock{}

	// Add logo if standard blocks exist
	if len(standardBlocks) > 0 {
		newChildren = append(newChildren, standardBlocks[0]) // Add logo
	}

	// Add heading
	headingBlock := DeepCopyBlock(blocks["heading"])
	headingBlock.ID = "heading"
	headingBlock.Data.(map[string]interface{})["type"] = "h2"
	headingBlock.Data.(map[string]interface{})["align"] = "center"
	headingBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "h2",
			"children": []map[string]interface{}{
				{"text": "You Have Been Unsubscribed"},
			},
		},
	}
	newChildren = append(newChildren, headingBlock)

	// Add text content
	textBlock := DeepCopyBlock(blocks["text"])
	textBlock.ID = "text"
	textBlock.Data.(map[string]interface{})["align"] = "center"
	textBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "We're sorry to see you go! You have been successfully unsubscribed from our newsletter."},
			},
		},
	}
	newChildren = append(newChildren, textBlock)

	// Add resubscribe section
	resubTextBlock := DeepCopyBlock(blocks["text"])
	resubTextBlock.ID = "resub-text"
	resubTextBlock.Data.(map[string]interface{})["align"] = "center"
	resubTextBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "Changed your mind? You can resubscribe at any time by clicking the button below."},
			},
		},
	}
	newChildren = append(newChildren, resubTextBlock)

	// Add resubscribe button
	buttonBlock := DeepCopyBlock(blocks["button"])
	buttonBlock.ID = "button"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["backgroundColor"] = "#4e6cff"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["href"] = "{{ resubscribe_url }}"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["text"] = "RESUBSCRIBE"
	newChildren = append(newChildren, buttonBlock)

	// Add back the standard blocks (divider, footer, tracking) if they exist
	if len(standardBlocks) > 1 {
		newChildren = append(newChildren, standardBlocks[1:]...)
	}

	// Create a new column with the children properly set
	updatedColumn := DeepCopyBlock(contentColumn)
	updatedColumn.Children = newChildren

	// Create a new section with the updated column
	updatedSection := DeepCopyBlock(contentSection)
	updatedSection.Children = []EmailBlock{updatedColumn}

	// Create a new root with the updated section
	updatedRoot := DeepCopyBlock(emailTemplate)
	updatedRoot.Children = []EmailBlock{updatedSection}

	return updatedRoot
}

// DefaultSubscriptionConfirmationEmail returns an EmailBlock tree for a subscription confirmation email
func DefaultSubscriptionConfirmationEmail() EmailBlock {
	blocks := DefaultBlocks()

	// Get the template with standard structure
	emailTemplate := DefaultTemplateStructure()

	// Access the content column to add our specific blocks
	contentSection := emailTemplate.Children[0]
	contentColumn := contentSection.Children[0]

	// Store the standard blocks (logo, divider, footer, tracking)
	standardBlocks := contentColumn.Children

	// Initialize new children array
	newChildren := []EmailBlock{}

	// Add logo if standard blocks exist
	if len(standardBlocks) > 0 {
		newChildren = append(newChildren, standardBlocks[0]) // Add logo
	}

	// Add heading
	headingBlock := DeepCopyBlock(blocks["heading"])
	headingBlock.ID = "heading"
	headingBlock.Data.(map[string]interface{})["type"] = "h2"
	headingBlock.Data.(map[string]interface{})["align"] = "center"
	headingBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "h2",
			"children": []map[string]interface{}{
				{"text": "Subscription Confirmed!"},
			},
		},
	}
	newChildren = append(newChildren, headingBlock)

	// Add text content
	textBlock := DeepCopyBlock(blocks["text"])
	textBlock.ID = "text"
	textBlock.Data.(map[string]interface{})["align"] = "center"
	textBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "Thank you for subscribing to our newsletter! We're excited to have you as part of our community."},
			},
		},
	}
	newChildren = append(newChildren, textBlock)

	// Add welcome message
	welcomeTextBlock := DeepCopyBlock(blocks["text"])
	welcomeTextBlock.ID = "welcome-text"
	welcomeTextBlock.Data.(map[string]interface{})["align"] = "center"
	welcomeTextBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "You'll now receive updates, news, and special offers from us. We promise not to spam your inbox!"},
			},
		},
	}
	newChildren = append(newChildren, welcomeTextBlock)

	// Add unsubscribe text
	unsubTextBlock := DeepCopyBlock(blocks["text"])
	unsubTextBlock.ID = "unsub-text"
	unsubTextBlock.Data.(map[string]interface{})["align"] = "center"
	unsubTextBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{
					"text":     "If you ever wish to unsubscribe, simply click the link below:",
					"color":    "#666666",
					"fontSize": "14px",
				},
			},
		},
	}
	newChildren = append(newChildren, unsubTextBlock)

	// Add unsubscribe link
	unsubLinkBlock := DeepCopyBlock(blocks["text"])
	unsubLinkBlock.ID = "unsub-link"
	unsubLinkBlock.Data.(map[string]interface{})["align"] = "center"
	unsubLinkBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{
					"text":       "Unsubscribe",
					"color":      "#4e6cff",
					"fontSize":   "14px",
					"linkTarget": "_blank",
					"linkHref":   "{{ unsubscribe_url }}",
				},
			},
		},
	}
	newChildren = append(newChildren, unsubLinkBlock)

	// Add back the standard blocks (divider, footer, tracking) if they exist
	if len(standardBlocks) > 1 {
		newChildren = append(newChildren, standardBlocks[1:]...)
	}

	// Create a new column with the children properly set
	updatedColumn := DeepCopyBlock(contentColumn)
	updatedColumn.Children = newChildren

	// Create a new section with the updated column
	updatedSection := DeepCopyBlock(contentSection)
	updatedSection.Children = []EmailBlock{updatedColumn}

	// Create a new root with the updated section
	updatedRoot := DeepCopyBlock(emailTemplate)
	updatedRoot.Children = []EmailBlock{updatedSection}

	return updatedRoot
}

// DefaultTransactionalEmail returns an EmailBlock tree for a transactional notification email
func DefaultTransactionalEmail() EmailBlock {
	blocks := DefaultBlocks()

	// Get the template with standard structure
	emailTemplate := DefaultTemplateStructure()

	// Access the content column to add our specific blocks
	contentSection := emailTemplate.Children[0]
	contentColumn := contentSection.Children[0]

	// Store the standard blocks (logo, divider, footer, tracking)
	standardBlocks := contentColumn.Children

	// Initialize new children array
	newChildren := []EmailBlock{}

	// Add logo if standard blocks exist
	if len(standardBlocks) > 0 {
		newChildren = append(newChildren, standardBlocks[0]) // Add logo
	}

	// Add heading
	headingBlock := DeepCopyBlock(blocks["heading"])
	headingBlock.ID = "heading"
	headingBlock.Data.(map[string]interface{})["type"] = "h2"
	headingBlock.Data.(map[string]interface{})["align"] = "center"
	headingBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "h2",
			"children": []map[string]interface{}{
				{"text": "{{ heading }}"},
			},
		},
	}
	newChildren = append(newChildren, headingBlock)

	// Add main content
	textBlock := DeepCopyBlock(blocks["text"])
	textBlock.ID = "main-content"
	textBlock.Data.(map[string]interface{})["align"] = "center"
	textBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{"text": "{{ message }}"},
			},
		},
	}
	newChildren = append(newChildren, textBlock)

	// Add CTA button if needed
	buttonBlock := DeepCopyBlock(blocks["button"])
	buttonBlock.ID = "cta-button"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["backgroundColor"] = "#4e6cff"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["href"] = "{{ cta_url }}"
	buttonBlock.Data.(map[string]interface{})["button"].(map[string]interface{})["text"] = "{{ cta_text }}"
	newChildren = append(newChildren, buttonBlock)

	// Add additional information text
	additionalBlock := DeepCopyBlock(blocks["text"])
	additionalBlock.ID = "additional-info"
	additionalBlock.Data.(map[string]interface{})["align"] = "center"
	additionalBlock.Data.(map[string]interface{})["editorData"] = []map[string]interface{}{
		{
			"type": "paragraph",
			"children": []map[string]interface{}{
				{
					"text":     "{{ additional_info }}",
					"fontSize": "14px",
				},
			},
		},
	}
	newChildren = append(newChildren, additionalBlock)

	// Add back the standard blocks (divider, footer, tracking) if they exist
	if len(standardBlocks) > 1 {
		newChildren = append(newChildren, standardBlocks[1:]...)
	}

	// Create a new column with the children properly set
	updatedColumn := DeepCopyBlock(contentColumn)
	updatedColumn.Children = newChildren

	// Create a new section with the updated column
	updatedSection := DeepCopyBlock(contentSection)
	updatedSection.Children = []EmailBlock{updatedColumn}

	// Create a new root with the updated section
	updatedRoot := DeepCopyBlock(emailTemplate)
	updatedRoot.Children = []EmailBlock{updatedSection}

	return updatedRoot
}

// DefaultEmailStyles returns a base email template with standard styling
func DefaultEmailStyles() EmailBlock {
	blocks := DefaultBlocks()

	// Create root block with standard styling
	rootBlock := DeepCopyBlock(blocks["root"])

	// Create section with a single column
	contentSection := DeepCopyBlock(blocks["oneColumn"])
	contentSection.ID = "content-section"

	// Create the content column
	contentColumn := DeepCopyBlock(blocks["column"])
	contentColumn.ID = "content-column"

	// Structure the email template
	contentSection.Children = []EmailBlock{contentColumn}
	rootBlock.Children = []EmailBlock{contentSection}

	return rootBlock
}

// DefaultBlocks returns a map of default email blocks with common settings
func DefaultBlocks() map[string]EmailBlock {
	blocks := make(map[string]EmailBlock)

	// Root block
	blocks["root"] = EmailBlock{
		ID:   "root",
		Kind: "root",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"body": map[string]interface{}{
					"width":           "600px",
					"margin":          "0 auto",
					"backgroundColor": "#FFFFFF",
				},
				"h1": map[string]interface{}{
					"color":          "#000000",
					"fontSize":       "34px",
					"fontStyle":      "normal",
					"fontWeight":     400,
					"paddingControl": "separate",
					"padding":        "0px",
					"paddingTop":     "60px",
					"paddingRight":   "0px",
					"paddingBottom":  "60px",
					"paddingLeft":    "0px",
					"margin":         0,
					"fontFamily":     "Helvetica, sans-serif",
				},
				"h2": map[string]interface{}{
					"color":          "#000000",
					"fontSize":       "28px",
					"fontStyle":      "normal",
					"fontWeight":     400,
					"paddingControl": "separate",
					"padding":        "0px",
					"paddingTop":     "40px",
					"paddingRight":   "0px",
					"paddingBottom":  "40px",
					"paddingLeft":    "0px",
					"margin":         0,
					"fontFamily":     "Helvetica, sans-serif",
				},
				"h3": map[string]interface{}{
					"color":          "#000000",
					"fontSize":       "22px",
					"fontStyle":      "normal",
					"fontWeight":     400,
					"paddingControl": "separate",
					"padding":        "0px",
					"paddingTop":     "20px",
					"paddingRight":   "0px",
					"paddingBottom":  "20px",
					"paddingLeft":    "0px",
					"margin":         0,
					"fontFamily":     "Helvetica, sans-serif",
				},
				"paragraph": map[string]interface{}{
					"color":          "#000000",
					"fontSize":       "16px",
					"fontStyle":      "normal",
					"fontWeight":     400,
					"paddingControl": "separate",
					"padding":        "0px",
					"paddingTop":     "0px",
					"paddingRight":   "0px",
					"paddingBottom":  "20px",
					"paddingLeft":    "0px",
					"margin":         0,
					"fontFamily":     "Helvetica, sans-serif",
				},
				"hyperlink": map[string]interface{}{
					"color":          "#4e6cff",
					"textDecoration": "none",
					"fontFamily":     "Helvetica, sans-serif",
					"fontSize":       "16px",
					"fontWeight":     400,
					"fontStyle":      "normal",
					"textTransform":  "none",
				},
			},
		},
		Children: []EmailBlock{},
	}

	// Button block
	blocks["button"] = EmailBlock{
		ID:   "button",
		Kind: "button",
		Data: map[string]interface{}{
			"wrapper": map[string]interface{}{
				"align":          "center",
				"paddingControl": "all",
				"padding":        "20px",
			},
			"button": map[string]interface{}{
				"backgroundColor":        "#00BCD4",
				"href":                   "https://example.com",
				"text":                   "Click me!",
				"innerVerticalPadding":   "10px",
				"innerHorizontalPadding": "25px",
				"borderControl":          "all",
				"borderColor":            "#000000",
				"borderWidth":            "2px",
				"borderStyle":            "none",
				"borderRadius":           "8px",
				"width":                  "auto",
				"color":                  "#FFFFFF",
				"fontFamily":             "Helvetica, sans-serif",
				"fontWeight":             600,
				"fontSize":               "15px",
				"fontStyle":              "normal",
				"textTransform":          "uppercase",
			},
		},
		Children: []EmailBlock{},
	}

	// Text block
	blocks["text"] = EmailBlock{
		ID:   "text",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"width": "100%",
			"hyperlinkStyles": map[string]interface{}{
				"color":          "#4e6cff",
				"textDecoration": "none",
				"fontFamily":     "Helvetica, sans-serif",
				"fontSize":       "16px",
				"fontWeight":     400,
				"fontStyle":      "normal",
				"textTransform":  "none",
			},
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "A line of text in a paragraph."},
					},
				},
			},
		},
		Children: []EmailBlock{},
	}

	// Heading block
	blocks["heading"] = EmailBlock{
		ID:   "heading",
		Kind: "heading",
		Data: map[string]interface{}{
			"type":  "h1",
			"align": "left",
			"width": "100%",
			"editorData": []map[string]interface{}{
				{
					"type": "h1",
					"children": []map[string]interface{}{
						{"text": "Heading"},
					},
				},
			},
		},
		Children: []EmailBlock{},
	}

	// Divider block
	blocks["divider"] = EmailBlock{
		ID:   "divider",
		Kind: "divider",
		Data: map[string]interface{}{
			"align":          "center",
			"paddingControl": "all",
			"padding":        "20px",
			"borderColor":    "#B0BEC5",
			"borderWidth":    "1px",
			"borderStyle":    "solid",
			"width":          "100%",
		},
		Children: []EmailBlock{},
	}

	// Image block
	blocks["image"] = EmailBlock{
		ID:   "image",
		Kind: "image",
		Data: map[string]interface{}{
			"wrapper": map[string]interface{}{
				"align":          "center",
				"paddingControl": "all",
				"padding":        "20px",
			},
			"image": map[string]interface{}{
				"src":               "https://placehold.co/600x400",
				"alt":               "Image description",
				"width":             "300px",
				"height":            "auto",
				"href":              "",
				"borderControl":     "all",
				"borderWidth":       "2px",
				"borderStyle":       "none",
				"borderColor":       "#000000",
				"fullWidthOnMobile": false,
			},
		},
		Children: []EmailBlock{},
	}

	// Liquid block
	blocks["liquid"] = EmailBlock{
		ID:   "liquid",
		Kind: "liquid",
		Data: map[string]interface{}{
			"liquidCode": `{% if contact %}
<mj-text font-size="20px" color="#333333" font-family="helvetica">
  Hello {{ contact.first_name }}!
</mj-text>
<mj-text font-size="16px" color="#666666" font-family="helvetica">
  Email: {{ contact.email }}<br/>
  Phone: {{ contact.phone }}<br/>
  Country: {{ contact.country }}
</mj-text>
<mj-button background-color="#4CAF50" href="mailto:{{ contact.email }}">
  Contact Now
</mj-button>
{% else %}
<mj-text font-size="20px" color="#333333" font-family="helvetica">
  No Contact Provided
</mj-text>
<mj-text font-size="16px" color="#666666" font-family="helvetica">
  Please provide a contact to view their details.
</mj-text>
{% endif %}`,
		},
		Children: []EmailBlock{},
	}

	// OpenTracking block
	blocks["openTracking"] = NewOpenTrackingBlock()

	// Section block
	blocks["section"] = EmailBlock{
		ID:   "section",
		Kind: "section",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{},
	}

	// Column block
	blocks["column"] = EmailBlock{
		ID:   "column",
		Kind: "column",
		Data: map[string]interface{}{
			"paddingControl": "all",
			"borderControl":  "all",
			"styles": map[string]interface{}{
				"verticalAlign": "top",
				"minHeight":     "30px",
			},
		},
		Children: []EmailBlock{},
	}

	// One Column section
	blocks["oneColumn"] = EmailBlock{
		ID:   "oneColumn",
		Kind: "oneColumn",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			blocks["column"],
		},
	}

	// Two Columns (1:1)
	blocks["columns1212"] = EmailBlock{
		ID:   "columns1212",
		Kind: "columns1212",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{12, 12},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingRight":  "15px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingLeft":   "15px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Two Columns (2:1)
	blocks["columns168"] = EmailBlock{
		ID:   "columns168",
		Kind: "columns168",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{16, 8},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingRight":  "15px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingLeft":   "15px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Two Columns (1:2)
	blocks["columns816"] = EmailBlock{
		ID:   "columns816",
		Kind: "columns816",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{8, 16},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingRight":  "15px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingLeft":   "15px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Two Columns (5:1)
	blocks["columns204"] = EmailBlock{
		ID:   "columns204",
		Kind: "columns204",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{20, 4},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingRight":  "15px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingLeft":   "15px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Two Columns (1:5)
	blocks["columns420"] = EmailBlock{
		ID:   "columns420",
		Kind: "columns420",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{4, 20},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingRight":  "15px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "separate",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
						"paddingLeft":   "15px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Three Columns (1:1:1)
	blocks["columns888"] = EmailBlock{
		ID:   "columns888",
		Kind: "columns888",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{8, 8, 8},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column3",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	// Four Columns (1:1:1:1)
	blocks["columns6666"] = EmailBlock{
		ID:   "columns6666",
		Kind: "columns6666",
		Data: map[string]interface{}{
			"columnsOnMobile":     false,
			"stackColumnsAtWidth": 480,
			"backgroundType":      "color",
			"paddingControl":      "all",
			"borderControl":       "all",
			"columns":             []int{6, 6, 6, 6},
			"styles": map[string]interface{}{
				"textAlign":        "center",
				"backgroundRepeat": "repeat",
				"padding":          "30px",
				"borderWidth":      "0px",
				"borderStyle":      "none",
				"borderColor":      "#000000",
			},
		},
		Children: []EmailBlock{
			{
				ID:   "column1",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column2",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column3",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
			{
				ID:   "column4",
				Kind: "column",
				Data: map[string]interface{}{
					"paddingControl": "all",
					"borderControl":  "all",
					"styles": map[string]interface{}{
						"verticalAlign": "top",
						"minHeight":     "30px",
					},
				},
				Children: []EmailBlock{},
			},
		},
	}

	return blocks
}
