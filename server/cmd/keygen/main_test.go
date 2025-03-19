package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateEnvContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		privateKey  string
		publicKey   string
		expected    string
		description string
	}{
		{
			name:        "empty_content",
			content:     "",
			privateKey:  "private-key-value",
			publicKey:   "public-key-value",
			expected:    "PASETO_PRIVATE_KEY=private-key-value\nPASETO_PUBLIC_KEY=public-key-value\n",
			description: "Should add both keys when content is empty",
		},
		{
			name:        "replace_existing_keys",
			content:     "PASETO_PRIVATE_KEY=old-private\nPASETO_PUBLIC_KEY=old-public\n",
			privateKey:  "new-private",
			publicKey:   "new-public",
			expected:    "PASETO_PRIVATE_KEY=new-private\nPASETO_PUBLIC_KEY=new-public\n",
			description: "Should replace existing keys",
		},
		{
			name:        "add_missing_keys",
			content:     "OTHER_KEY=some-value\n",
			privateKey:  "private-key-value",
			publicKey:   "public-key-value",
			expected:    "OTHER_KEY=some-value\nPASETO_PRIVATE_KEY=private-key-value\nPASETO_PUBLIC_KEY=public-key-value\n",
			description: "Should add missing keys while preserving other content",
		},
		{
			name: "preserve_comments",
			content: "# This is a comment\nOTHER_KEY=value\n" +
				"PASETO_PRIVATE_KEY=old-private\n# Another comment\nPASETO_PUBLIC_KEY=old-public\n",
			privateKey: "new-private",
			publicKey:  "new-public",
			expected: "# This is a comment\nOTHER_KEY=value\n" +
				"PASETO_PRIVATE_KEY=new-private\n# Another comment\nPASETO_PUBLIC_KEY=new-public\n",
			description: "Should preserve comments and other content",
		},
		{
			name:        "handle_empty_lines",
			content:     "PASETO_PRIVATE_KEY=old-private\n\nPASETO_PUBLIC_KEY=old-public\n",
			privateKey:  "new-private",
			publicKey:   "new-public",
			expected:    "PASETO_PRIVATE_KEY=new-private\n\nPASETO_PUBLIC_KEY=new-public\n",
			description: "Should preserve empty lines",
		},
		{
			name:        "partial_keys_present",
			content:     "PASETO_PRIVATE_KEY=old-private\n",
			privateKey:  "new-private",
			publicKey:   "public-key-value",
			expected:    "PASETO_PRIVATE_KEY=new-private\nPASETO_PUBLIC_KEY=public-key-value\n",
			description: "Should update existing key and add missing one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateEnvContent(tt.content, tt.privateKey, tt.publicKey)

			// Normalize line endings for comparison
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")
			result = strings.ReplaceAll(result, "\r\n", "\n")

			assert.Equal(t, expected, result, tt.description)
		})
	}
}
