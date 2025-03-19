package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
	"notifuse/server/pkg/logger"
)

func TestEmailSender_SendMagicCode(t *testing.T) {
	// Redirect log output for testing
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // Restore default output

	// Create email sender
	sender := &emailSender{}

	// Test sending a magic code
	err := sender.SendMagicCode("test@example.com", "123456")

	// Verify no error
	assert.NoError(t, err)

	// Verify log output contains the expected message
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Sending magic code to test@example.com: 123456")
}

func TestConfigLoading(t *testing.T) {
	// Skip in CI environment to avoid env file requirements
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Test loading configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{
		EnvFile: ".env.test",
	})

	// If there's no env file, this will fail - that's expected in test environments
	if err != nil {
		assert.Contains(t, err.Error(), "PASETO_")
		return
	}
	assert.NotNil(t, cfg)
}

// Mock implementations for reference if needed for future tests
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUserRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockUserRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *MockUserRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}
