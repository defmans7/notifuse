package domain

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapOfAny_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    MapOfAny
		wantErr bool
	}{
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key": "value", "number": 123}`),
			want: MapOfAny{
				"key":    "value",
				"number": float64(123), // JSON unmarshals numbers as float64
			},
			wantErr: false,
		},
		{
			name:  "valid JSON string",
			input: `{"key": "value", "number": 123}`,
			want: MapOfAny{
				"key":    "value",
				"number": float64(123),
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid json`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m MapOfAny
			err := m.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestMapOfAny_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   MapOfAny
		want    driver.Value
		wantErr bool
	}{
		{
			name: "valid map",
			input: MapOfAny{
				"key":    "value",
				"number": 123,
			},
			want:    []byte(`{"key":"value","number":123}`),
			wantErr: false,
		},
		{
			name:    "nil map",
			input:   nil,
			want:    []byte("null"),
			wantErr: false,
		},
		{
			name: "complex map",
			input: MapOfAny{
				"string": "value",
				"number": 123,
				"bool":   true,
				"null":   nil,
				"array":  []interface{}{1, "two", 3},
				"object": map[string]interface{}{"nested": "value"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			} else {
				// For complex cases, verify the JSON is valid
				jsonBytes, ok := got.([]byte)
				assert.True(t, ok)

				var unmarshaled interface{}
				err := json.Unmarshal(jsonBytes, &unmarshaled)
				assert.NoError(t, err)
			}
		})
	}
}
