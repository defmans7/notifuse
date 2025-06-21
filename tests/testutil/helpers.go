package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite provides a complete testing environment
type IntegrationTestSuite struct {
	DBManager     *DatabaseManager
	ServerManager *ServerManager
	APIClient     *APIClient
	DataFactory   *TestDataFactory
	Config        *config.Config
	T             *testing.T
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T, appFactory func(*config.Config) AppInterface) *IntegrationTestSuite {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	suite := &IntegrationTestSuite{T: t}

	// Setup database
	suite.DBManager = NewDatabaseManager()
	err := suite.DBManager.Setup()
	require.NoError(t, err, "Failed to setup test database")

	// Wait for database to be ready
	err = suite.DBManager.WaitForDatabase(30)
	require.NoError(t, err, "Database not ready")

	// Setup server
	suite.ServerManager = NewServerManager(appFactory, suite.DBManager)
	err = suite.ServerManager.Start()
	require.NoError(t, err, "Failed to start test server")

	// Setup API client
	suite.APIClient = NewAPIClient(suite.ServerManager.GetURL())

	// Setup data factory with repositories from the app
	app := suite.ServerManager.GetApp()
	suite.DataFactory = NewTestDataFactory(
		suite.DBManager.GetDB(),
		app.GetUserRepository(),
		app.GetWorkspaceRepository(),
		app.GetContactRepository(),
		app.GetListRepository(),
		app.GetTemplateRepository(),
		app.GetBroadcastRepository(),
		app.GetMessageHistoryRepository(),
		app.GetContactListRepository(),
	)

	// Seed initial test data
	err = suite.DBManager.SeedTestData()
	require.NoError(t, err, "Failed to seed test data")

	// Set workspace ID for API client
	suite.APIClient.SetWorkspaceID("test-workspace-id")

	suite.Config = suite.ServerManager.GetApp().GetConfig()

	return suite
}

// Cleanup cleans up all test resources
func (s *IntegrationTestSuite) Cleanup() {
	if s.ServerManager != nil {
		s.ServerManager.Stop()
	}
	if s.DBManager != nil {
		s.DBManager.Cleanup()
	}
}

// ResetData cleans and reseeds test data
func (s *IntegrationTestSuite) ResetData() {
	err := s.DBManager.CleanupTestData()
	require.NoError(s.T, err, "Failed to cleanup test data")

	err = s.DBManager.SeedTestData()
	require.NoError(s.T, err, "Failed to seed test data")
}

// WaitForCondition waits for a condition to be true within a timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to a deterministic approach if random fails
		for i := range b {
			b[i] = charset[i%len(charset)]
		}
	} else {
		for i := range b {
			b[i] = charset[b[i]%byte(len(charset))]
		}
	}
	return string(b)
}

// GenerateTestEmail generates a test email address
func GenerateTestEmail() string {
	return fmt.Sprintf("test-%s@example.com", GenerateRandomString(8))
}

// CreateTestLogger creates a logger for testing
func CreateTestLogger() logger.Logger {
	return logger.NewLogger()
}

// AssertEventuallyTrue asserts that a condition becomes true within a timeout
func AssertEventuallyTrue(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	require.Eventually(t, condition, timeout, 100*time.Millisecond, message)
}

// AssertNeverTrue asserts that a condition never becomes true within a duration
func AssertNeverTrue(t *testing.T, condition func() bool, duration time.Duration, message string) {
	require.Never(t, condition, duration, 100*time.Millisecond, message)
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// RequireEnvironmentVar requires an environment variable to be set
func RequireEnvironmentVar(t *testing.T, envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", envVar)
	}
	return value
}

// SetupTestEnvironment sets up environment variables for testing
func SetupTestEnvironment() {
	os.Setenv("TEST_DB_HOST", "localhost")
	os.Setenv("TEST_DB_PORT", "5433")
	os.Setenv("TEST_DB_USER", "notifuse_test")
	os.Setenv("TEST_DB_PASSWORD", "test_password")
	os.Setenv("ENVIRONMENT", "test")
}

// CleanupTestEnvironment cleans up test environment variables
func CleanupTestEnvironment() {
	os.Unsetenv("TEST_DB_HOST")
	os.Unsetenv("TEST_DB_PORT")
	os.Unsetenv("TEST_DB_USER")
	os.Unsetenv("TEST_DB_PASSWORD")
	os.Unsetenv("ENVIRONMENT")
}
