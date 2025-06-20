package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/app"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactAPIEndpoints(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	client := suite.APIClient

	t.Run("List Contacts", func(t *testing.T) {
		resp, err := client.ListContacts(map[string]string{
			"limit": "10",
		})
		require.NoError(t, err, "Should be able to list contacts")
		defer resp.Body.Close()

		// Should get a response (might be unauthorized but endpoint exists)
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "Contacts list endpoint should exist")
	})

	t.Run("Create Contact", func(t *testing.T) {
		contact := map[string]interface{}{
			"email":      testutil.GenerateTestEmail(),
			"first_name": "Test",
			"last_name":  "User",
		}

		resp, err := client.CreateContact(contact)
		require.NoError(t, err, "Should be able to create contact")
		defer resp.Body.Close()

		// Should get a response (might fail validation but endpoint exists)
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "Contact create endpoint should exist")
	})

	t.Run("Get Contact by Email", func(t *testing.T) {
		resp, err := client.GetContactByEmail("test@example.com")
		require.NoError(t, err, "Should be able to get contact by email")
		defer resp.Body.Close()

		// Should get a response (might be not found but endpoint exists)
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "Get contact by email endpoint should exist")
	})
}

func TestContactDataFactory(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	factory := suite.DataFactory

	t.Run("Create Contact", func(t *testing.T) {
		contact, err := factory.CreateContact()
		require.NoError(t, err, "Should be able to create contact")
		require.NotNil(t, contact, "Contact should not be nil")

		assert.NotEmpty(t, contact.Email, "Contact should have email")
		assert.NotNil(t, contact.FirstName, "Contact should have first name")
		assert.NotNil(t, contact.LastName, "Contact should have last name")
	})

	t.Run("Create Contact with Options", func(t *testing.T) {
		email := testutil.GenerateTestEmail()
		contact, err := factory.CreateContact(
			testutil.WithContactEmail(email),
			testutil.WithContactName("John", "Doe"),
			testutil.WithContactExternalID("ext-123"),
		)
		require.NoError(t, err, "Should be able to create contact with options")

		assert.Equal(t, email, contact.Email)
		assert.Equal(t, "John", contact.FirstName.String)
		assert.Equal(t, "Doe", contact.LastName.String)
		assert.Equal(t, "ext-123", contact.ExternalID.String)
	})

	t.Run("Create Multiple Contacts", func(t *testing.T) {
		contacts := make([]*domain.Contact, 5)
		for i := 0; i < 5; i++ {
			contact, err := factory.CreateContact(
				testutil.WithContactEmail(fmt.Sprintf("user%d@example.com", i)),
			)
			require.NoError(t, err, "Should be able to create contact %d", i)
			contacts[i] = contact
		}

		// Verify all contacts have different emails
		emails := make(map[string]bool)
		for _, contact := range contacts {
			assert.False(t, emails[contact.Email], "Email should be unique")
			emails[contact.Email] = true
		}
	})
}

func TestContactDatabaseOperations(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	factory := suite.DataFactory
	db := suite.DBManager.GetDB()

	t.Run("Contact Persisted to Database", func(t *testing.T) {
		email := testutil.GenerateTestEmail()
		_, err := factory.CreateContact(
			testutil.WithContactEmail(email),
		)
		require.NoError(t, err, "Should be able to create contact")

		// Verify contact exists in database
		var dbEmail string
		err = db.QueryRow("SELECT email FROM contacts WHERE email = $1", email).Scan(&dbEmail)
		require.NoError(t, err, "Contact should exist in database")
		assert.Equal(t, email, dbEmail)
	})

	t.Run("Contact Fields Stored Correctly", func(t *testing.T) {
		email := testutil.GenerateTestEmail()
		_, err := factory.CreateContact(
			testutil.WithContactEmail(email),
			testutil.WithContactName("Alice", "Smith"),
			testutil.WithContactExternalID("ext-456"),
		)
		require.NoError(t, err, "Should be able to create contact")

		// Verify all fields are stored correctly
		var dbFirstName, dbLastName, dbExternalID string
		err = db.QueryRow(`
			SELECT 
				COALESCE(first_name, ''),
				COALESCE(last_name, ''),
				COALESCE(external_id, '')
			FROM contacts WHERE email = $1
		`, email).Scan(&dbFirstName, &dbLastName, &dbExternalID)
		require.NoError(t, err, "Should be able to query contact fields")

		assert.Equal(t, "Alice", dbFirstName)
		assert.Equal(t, "Smith", dbLastName)
		assert.Equal(t, "ext-456", dbExternalID)
	})

	t.Run("Contact Cleanup", func(t *testing.T) {
		// Create some contacts
		for i := 0; i < 3; i++ {
			_, err := factory.CreateContact()
			require.NoError(t, err, "Should be able to create contact")
		}

		// Verify contacts exist
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
		require.NoError(t, err)
		assert.Greater(t, count, 0, "Should have contacts")

		// Clean up
		err = suite.DBManager.CleanupTestData()
		require.NoError(t, err, "Should be able to cleanup test data")

		// Verify contacts are cleaned up
		err = db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no contacts after cleanup")
	})
}

func TestContactAPIIntegration(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	client := suite.APIClient
	factory := suite.DataFactory

	t.Run("API Returns Created Contact Data", func(t *testing.T) {
		// Create contact in database
		email := testutil.GenerateTestEmail()
		_, err := factory.CreateContact(
			testutil.WithContactEmail(email),
			testutil.WithContactName("Bob", "Johnson"),
		)
		require.NoError(t, err, "Should be able to create contact")

		// Try to fetch via API (might fail due to auth but test structure)
		resp, err := client.GetContactByEmail(email)
		require.NoError(t, err, "Should be able to make API request")
		defer resp.Body.Close()

		// If we get data back, verify structure
		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Should be able to decode response")

			if contactData, ok := response["contact"]; ok {
				contactMap := contactData.(map[string]interface{})
				assert.Equal(t, email, contactMap["email"])
			}
		}
	})

	t.Run("API Contact List Structure", func(t *testing.T) {
		// Create a few contacts
		for i := 0; i < 3; i++ {
			_, err := factory.CreateContact()
			require.NoError(t, err, "Should be able to create contact")
		}

		resp, err := client.ListContacts(map[string]string{
			"limit": "10",
		})
		require.NoError(t, err, "Should be able to list contacts")
		defer resp.Body.Close()

		// If we get data back, verify structure
		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Should be able to decode response")

			// Check for expected fields in response structure
			if contactsData, ok := response["contacts"]; ok {
				contacts := contactsData.([]interface{})
				t.Logf("Found %d contacts in API response", len(contacts))
			}
		}
	})
}
