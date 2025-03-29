package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockLogger(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MockLogger)
		validate func(*testing.T, *MockLogger)
	}{
		{
			name: "Debug method",
			setup: func(m *MockLogger) {
				m.On("Debug", "test message")
			},
			validate: func(t *testing.T, m *MockLogger) {
				m.Debug("test message")
				m.AssertExpectations(t)
			},
		},
		{
			name: "Info method",
			setup: func(m *MockLogger) {
				m.On("Info", "test message")
			},
			validate: func(t *testing.T, m *MockLogger) {
				m.Info("test message")
				m.AssertExpectations(t)
			},
		},
		{
			name: "Warn method",
			setup: func(m *MockLogger) {
				m.On("Warn", "test message")
			},
			validate: func(t *testing.T, m *MockLogger) {
				m.Warn("test message")
				m.AssertExpectations(t)
			},
		},
		{
			name: "Error method",
			setup: func(m *MockLogger) {
				m.On("Error", "test message")
			},
			validate: func(t *testing.T, m *MockLogger) {
				m.Error("test message")
				m.AssertExpectations(t)
			},
		},
		{
			name: "Fatal method",
			setup: func(m *MockLogger) {
				m.On("Fatal", "test message")
			},
			validate: func(t *testing.T, m *MockLogger) {
				m.Fatal("test message")
				m.AssertExpectations(t)
			},
		},
		{
			name: "WithField method",
			setup: func(m *MockLogger) {
				m.On("WithField", "key", "value").Return(m)
			},
			validate: func(t *testing.T, m *MockLogger) {
				result := m.WithField("key", "value")
				assert.Equal(t, m, result)
				m.AssertExpectations(t)
			},
		},
		{
			name: "WithFields method",
			setup: func(m *MockLogger) {
				fields := map[string]interface{}{"key": "value"}
				m.On("WithFields", fields).Return(m)
			},
			validate: func(t *testing.T, m *MockLogger) {
				fields := map[string]interface{}{"key": "value"}
				result := m.WithFields(fields)
				assert.Equal(t, m, result)
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			if tt.setup != nil {
				tt.setup(mockLogger)
			}
			if tt.validate != nil {
				tt.validate(t, mockLogger)
			}
		})
	}
}
