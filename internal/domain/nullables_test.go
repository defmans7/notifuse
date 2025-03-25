package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullableString(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableString
			expected interface{}
			wantErr  bool
		}{
			{
				name:     "non-null string",
				input:    NullableString{String: "test", IsNull: false},
				expected: "test",
				wantErr:  false,
			},
			{
				name:     "null string",
				input:    NullableString{String: "", IsNull: true},
				expected: nil,
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				value, err := tt.input.Value()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, value)
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected NullableString
			wantErr  bool
		}{
			{
				name:     "scan string",
				input:    "test",
				expected: NullableString{String: "test", IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan []byte",
				input:    []byte("test"),
				expected: NullableString{String: "test", IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan nil",
				input:    nil,
				expected: NullableString{String: "", IsNull: true},
				wantErr:  false,
			},
			{
				name:     "scan invalid type",
				input:    123,
				expected: NullableString{},
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var ns NullableString
				err := ns.Scan(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, ns)
				}
			})
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableString
			expected string
			wantErr  bool
		}{
			{
				name:     "marshal non-null string",
				input:    NullableString{String: "test", IsNull: false},
				expected: `"test"`,
				wantErr:  false,
			},
			{
				name:     "marshal null string",
				input:    NullableString{String: "", IsNull: true},
				expected: "null",
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data, err := json.Marshal(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, string(data))
				}
			})
		}
	})
}

func TestNullableFloat64(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableFloat64
			expected interface{}
			wantErr  bool
		}{
			{
				name:     "non-null float",
				input:    NullableFloat64{Float64: 123.45, IsNull: false},
				expected: 123.45,
				wantErr:  false,
			},
			{
				name:     "null float",
				input:    NullableFloat64{Float64: 0, IsNull: true},
				expected: nil,
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				value, err := tt.input.Value()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, value)
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected NullableFloat64
			wantErr  bool
		}{
			{
				name:     "scan float64",
				input:    123.45,
				expected: NullableFloat64{Float64: 123.45, IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan int64",
				input:    int64(123),
				expected: NullableFloat64{Float64: 123, IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan []byte",
				input:    []byte("123.45"),
				expected: NullableFloat64{Float64: 123.45, IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan nil",
				input:    nil,
				expected: NullableFloat64{Float64: 0, IsNull: true},
				wantErr:  false,
			},
			{
				name:     "scan invalid type",
				input:    "invalid",
				expected: NullableFloat64{},
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var nf NullableFloat64
				err := nf.Scan(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, nf)
				}
			})
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableFloat64
			expected string
			wantErr  bool
		}{
			{
				name:     "marshal non-null float",
				input:    NullableFloat64{Float64: 123.45, IsNull: false},
				expected: "123.450000",
				wantErr:  false,
			},
			{
				name:     "marshal null float",
				input:    NullableFloat64{Float64: 0, IsNull: true},
				expected: "null",
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data, err := json.Marshal(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, string(data))
				}
			})
		}
	})
}

func TestNullableTime(t *testing.T) {
	fixedTime := time.Date(2024, 3, 25, 12, 0, 0, 0, time.UTC)

	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableTime
			expected interface{}
			wantErr  bool
		}{
			{
				name:     "non-null time",
				input:    NullableTime{Time: fixedTime, IsNull: false},
				expected: fixedTime,
				wantErr:  false,
			},
			{
				name:     "null time",
				input:    NullableTime{Time: time.Time{}, IsNull: true},
				expected: nil,
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				value, err := tt.input.Value()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, value)
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected NullableTime
			wantErr  bool
		}{
			{
				name:     "scan time",
				input:    fixedTime,
				expected: NullableTime{Time: fixedTime, IsNull: false},
				wantErr:  false,
			},
			{
				name:     "scan nil",
				input:    nil,
				expected: NullableTime{Time: time.Time{}, IsNull: true},
				wantErr:  false,
			},
			{
				name:     "scan invalid type",
				input:    "invalid",
				expected: NullableTime{},
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var nt NullableTime
				err := nt.Scan(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, nt)
				}
			})
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		tests := []struct {
			name     string
			input    NullableTime
			expected string
			wantErr  bool
		}{
			{
				name:     "marshal non-null time",
				input:    NullableTime{Time: fixedTime, IsNull: false},
				expected: `"2024-03-25T12:00:00Z"`,
				wantErr:  false,
			},
			{
				name:     "marshal null time",
				input:    NullableTime{Time: time.Time{}, IsNull: true},
				expected: "null",
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data, err := json.Marshal(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, string(data))
				}
			})
		}
	})
}
