package mjml

// DefaultOptinConfirmationEmail returns an EmailBlock tree for an opt-in confirmation email
func DefaultOptinConfirmationEmail() EmailBlock {
	logoBlock := EmailBlock{
		ID:   "logo",
		Kind: "image",
		Data: map[string]interface{}{
			"image": map[string]interface{}{
				"src":   "https://placehold.co/200x50?text=Your+Logo",
				"alt":   "Logo",
				"width": "200px",
			},
			"wrapper": map[string]interface{}{
				"align":   "center",
				"padding": "20px 0",
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading",
		Kind: "heading",
		Data: map[string]interface{}{
			"type":  "h2",
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "Please Confirm Your Subscription"},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	textBlock := EmailBlock{
		ID:   "text",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "Thank you for subscribing to our newsletter. To complete your subscription, please click the button below to confirm your email address."},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	buttonBlock := EmailBlock{
		ID:   "button",
		Kind: "button",
		Data: map[string]interface{}{
			"button": map[string]interface{}{
				"text":            "CONFIRM SUBSCRIPTION",
				"href":            "{{ confirmation_url }}",
				"backgroundColor": "#2e58ff",
				"color":           "#ffffff",
				"fontWeight":      700,
				"borderRadius":    "4px",
				"width":           "300px",
			},
			"wrapper": map[string]interface{}{
				"align": "center",
			},
		},
	}

	disclaimerBlock := EmailBlock{
		ID:   "disclaimer",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
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
			},
			"paddingTop": "20px",
		},
	}

	dividerBlock := EmailBlock{
		ID:   "divider",
		Kind: "divider",
		Data: map[string]interface{}{
			"borderColor": "#e0e0e0",
			"borderWidth": "1px",
			"borderStyle": "solid",
			"padding":     "20px 0",
		},
	}

	footerBlock := EmailBlock{
		ID:   "footer",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
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
			},
		},
	}

	contentColumn := EmailBlock{
		ID:   "content-column",
		Kind: "column",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"verticalAlign": "top",
			},
		},
		Children: []EmailBlock{logoBlock, headingBlock, textBlock, buttonBlock, disclaimerBlock, dividerBlock, footerBlock},
	}

	contentSection := EmailBlock{
		ID:   "content-section",
		Kind: "oneColumn",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#ffffff",
			},
		},
		Children: []EmailBlock{contentColumn},
	}

	rootBlock := EmailBlock{
		ID:   "root",
		Kind: "root",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#f8f8f8",
				"fontFamily":      "Helvetica, Arial, sans-serif",
				"fontSize":        "16px",
				"color":           "#000000",
				"lineHeight":      "24px",
				"body": map[string]interface{}{
					"width":           "460px",
					"backgroundColor": "#f8f8f8",
				},
			},
		},
		Children: []EmailBlock{contentSection},
	}

	return rootBlock
}

// DefaultUnsubscribeConfirmationEmail returns an EmailBlock tree for an unsubscribe confirmation email
func DefaultUnsubscribeConfirmationEmail() EmailBlock {
	logoBlock := EmailBlock{
		ID:   "logo",
		Kind: "image",
		Data: map[string]interface{}{
			"image": map[string]interface{}{
				"src":   "https://placehold.co/200x50?text=Your+Logo",
				"alt":   "Logo",
				"width": "200px",
			},
			"wrapper": map[string]interface{}{
				"align":   "center",
				"padding": "20px 0",
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading",
		Kind: "heading",
		Data: map[string]interface{}{
			"type":  "h2",
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "You've Been Unsubscribed"},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	textBlock := EmailBlock{
		ID:   "text",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "We're sorry to see you go. You have been successfully unsubscribed from our mailing list and will no longer receive emails from us."},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	noteBlock := EmailBlock{
		ID:   "note",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{
							"text":     "If you unsubscribed by accident or would like to resubscribe in the future, you can do so by visiting our website.",
							"color":    "#666666",
							"fontSize": "14px",
						},
					},
				},
			},
			"paddingTop": "10px",
		},
	}

	dividerBlock := EmailBlock{
		ID:   "divider",
		Kind: "divider",
		Data: map[string]interface{}{
			"borderColor": "#e0e0e0",
			"borderWidth": "1px",
			"borderStyle": "solid",
			"padding":     "20px 0",
		},
	}

	footerBlock := EmailBlock{
		ID:   "footer",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
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
			},
		},
	}

	contentColumn := EmailBlock{
		ID:   "content-column",
		Kind: "column",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"verticalAlign": "top",
			},
		},
		Children: []EmailBlock{logoBlock, headingBlock, textBlock, noteBlock, dividerBlock, footerBlock},
	}

	contentSection := EmailBlock{
		ID:   "content-section",
		Kind: "oneColumn",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#ffffff",
			},
		},
		Children: []EmailBlock{contentColumn},
	}

	rootBlock := EmailBlock{
		ID:   "root",
		Kind: "root",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#f8f8f8",
				"fontFamily":      "Helvetica, Arial, sans-serif",
				"fontSize":        "16px",
				"color":           "#000000",
				"lineHeight":      "24px",
				"body": map[string]interface{}{
					"width":           "460px",
					"backgroundColor": "#f8f8f8",
				},
			},
		},
		Children: []EmailBlock{contentSection},
	}

	return rootBlock
}

// DefaultWelcomeEmail returns an EmailBlock tree for a welcome email after subscription
func DefaultWelcomeEmail() EmailBlock {
	logoBlock := EmailBlock{
		ID:   "logo",
		Kind: "image",
		Data: map[string]interface{}{
			"image": map[string]interface{}{
				"src":   "https://placehold.co/200x50?text=Your+Logo",
				"alt":   "Logo",
				"width": "200px",
			},
			"wrapper": map[string]interface{}{
				"align":   "center",
				"padding": "20px 0",
			},
		},
	}

	headingBlock := EmailBlock{
		ID:   "heading",
		Kind: "heading",
		Data: map[string]interface{}{
			"type":  "h2",
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "Welcome to Our Community!"},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	introBlock := EmailBlock{
		ID:   "intro",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "Thank you for subscribing to our newsletter. We're excited to have you join our community! You'll be receiving updates, news, and exclusive content directly to your inbox."},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	expectationBlock := EmailBlock{
		ID:   "expectation",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "Here's what you can expect from us:"},
					},
				},
			},
			"paddingBottom": "20px",
		},
	}

	bulletPoint1Block := EmailBlock{
		ID:   "bullet1",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "• Regular updates on our latest products and services"},
					},
				},
			},
			"paddingLeft":   "40px",
			"paddingBottom": "5px",
		},
	}

	bulletPoint2Block := EmailBlock{
		ID:   "bullet2",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "• Exclusive offers and discounts for subscribers"},
					},
				},
			},
			"paddingLeft":   "40px",
			"paddingBottom": "5px",
		},
	}

	bulletPoint3Block := EmailBlock{
		ID:   "bullet3",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{"text": "• Tips, guides, and valuable industry insights"},
					},
				},
			},
			"paddingLeft":   "40px",
			"paddingBottom": "20px",
		},
	}

	dividerBlock := EmailBlock{
		ID:   "divider",
		Kind: "divider",
		Data: map[string]interface{}{
			"borderColor": "#e0e0e0",
			"borderWidth": "1px",
			"borderStyle": "solid",
			"padding":     "20px 0",
		},
	}

	footerBlock := EmailBlock{
		ID:   "footer",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
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
			},
		},
	}

	unsubscribeBlock := EmailBlock{
		ID:   "unsubscribe",
		Kind: "text",
		Data: map[string]interface{}{
			"align": "center",
			"editorData": []map[string]interface{}{
				{
					"type": "paragraph",
					"children": []map[string]interface{}{
						{
							"text":     "You can ",
							"color":    "#666666",
							"fontSize": "12px",
						},
						{
							"text": "unsubscribe",
							"hyperlink": map[string]interface{}{
								"url": "{{ unsubscribe_url }}",
							},
							"color":    "#666666",
							"fontSize": "12px",
						},
						{
							"text":     " at any time.",
							"color":    "#666666",
							"fontSize": "12px",
						},
					},
				},
			},
			"paddingTop": "10px",
		},
	}

	contentColumn := EmailBlock{
		ID:   "content-column",
		Kind: "column",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"verticalAlign": "top",
			},
		},
		Children: []EmailBlock{logoBlock, headingBlock, introBlock, expectationBlock, bulletPoint1Block, bulletPoint2Block, bulletPoint3Block, dividerBlock, footerBlock, unsubscribeBlock},
	}

	contentSection := EmailBlock{
		ID:   "content-section",
		Kind: "oneColumn",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#ffffff",
			},
		},
		Children: []EmailBlock{contentColumn},
	}

	rootBlock := EmailBlock{
		ID:   "root",
		Kind: "root",
		Data: map[string]interface{}{
			"styles": map[string]interface{}{
				"backgroundColor": "#f8f8f8",
				"fontFamily":      "Helvetica, Arial, sans-serif",
				"fontSize":        "16px",
				"color":           "#000000",
				"lineHeight":      "24px",
				"body": map[string]interface{}{
					"width":           "460px",
					"backgroundColor": "#f8f8f8",
				},
			},
		},
		Children: []EmailBlock{contentSection},
	}

	return rootBlock
}
