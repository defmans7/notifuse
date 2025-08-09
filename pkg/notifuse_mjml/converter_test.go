package notifuse_mjml

import (
	"testing"
)

func TestProcessLiquidTemplate(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		templateData map[string]interface{}
		context      string
		expected     string
		expectError  bool
	}{
		{
			name:         "no liquid tags",
			content:      "Hello World",
			templateData: map[string]interface{}{"name": "John"},
			context:      "test",
			expected:     "Hello World",
			expectError:  false,
		},
		{
			name:         "simple variable interpolation",
			content:      "Hello {{name}}!",
			templateData: map[string]interface{}{"name": "John"},
			context:      "test",
			expected:     "Hello John!",
			expectError:  false,
		},
		{
			name:         "multiple variables",
			content:      "Hello {{name}}, welcome to {{company}}!",
			templateData: map[string]interface{}{"name": "John", "company": "ACME Corp"},
			context:      "test",
			expected:     "Hello John, welcome to ACME Corp!",
			expectError:  false,
		},
		{
			name:         "conditional content",
			content:      "{% if isPremium %}Premium Member{% else %}Standard Member{% endif %}",
			templateData: map[string]interface{}{"isPremium": true},
			context:      "test",
			expected:     "Premium Member",
			expectError:  false,
		},
		{
			name:         "conditional content false",
			content:      "{% if isPremium %}Premium Member{% else %}Standard Member{% endif %}",
			templateData: map[string]interface{}{"isPremium": false},
			context:      "test",
			expected:     "Standard Member",
			expectError:  false,
		},
		{
			name:         "liquid filters",
			content:      "Hello {{name | upcase}}!",
			templateData: map[string]interface{}{"name": "john"},
			context:      "test",
			expected:     "Hello JOHN!",
			expectError:  false,
		},
		{
			name:         "empty template data",
			content:      "Hello {{name | default: 'Guest'}}!",
			templateData: nil,
			context:      "test",
			expected:     "Hello Guest!",
			expectError:  false,
		},
		{
			name:         "undefined variable with default",
			content:      "Hello {{unknown | default: 'Guest'}}!",
			templateData: map[string]interface{}{"name": "John"},
			context:      "test",
			expected:     "Hello Guest!",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessLiquidTemplate(tt.content, tt.templateData, tt.context)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcessLiquidTemplateEmailSubjects(t *testing.T) {
	// Specific tests for email subject line scenarios
	tests := []struct {
		name         string
		subject      string
		templateData map[string]interface{}
		expected     string
	}{
		{
			name:         "personalized subject",
			subject:      "Welcome {{firstName}}!",
			templateData: map[string]interface{}{"firstName": "John", "lastName": "Doe"},
			expected:     "Welcome John!",
		},
		{
			name:         "company and user subject",
			subject:      "{{firstName}}, your {{company}} order is ready",
			templateData: map[string]interface{}{"firstName": "Jane", "company": "ACME Corp"},
			expected:     "Jane, your ACME Corp order is ready",
		},
		{
			name:         "conditional urgency",
			subject:      "{% if urgent %}URGENT: {% endif %}Your order update",
			templateData: map[string]interface{}{"urgent": true},
			expected:     "URGENT: Your order update",
		},
		{
			name:         "non-urgent conditional",
			subject:      "{% if urgent %}URGENT: {% endif %}Your order update",
			templateData: map[string]interface{}{"urgent": false},
			expected:     "Your order update",
		},
		{
			name:         "order count simple",
			subject:      "You have {{orderCount}} order(s)",
			templateData: map[string]interface{}{"orderCount": 1},
			expected:     "You have 1 order(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessLiquidTemplate(tt.subject, tt.templateData, "email_subject")

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
