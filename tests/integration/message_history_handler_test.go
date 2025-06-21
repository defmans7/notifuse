package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/app"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageHistoryHandler(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	client := suite.APIClient
	factory := suite.DataFactory

	// Create test user and workspace
	user, err := factory.CreateUser()
	require.NoError(t, err)
	workspace, err := factory.CreateWorkspace()
	require.NoError(t, err)

	// Add user to workspace as owner
	err = factory.AddUserToWorkspace(user.ID, workspace.ID, "owner")
	require.NoError(t, err)

	// Login to get auth token
	err = client.Login(user.Email, "password")
	require.NoError(t, err)
	client.SetWorkspaceID(workspace.ID)

	t.Run("Messages List Endpoint", func(t *testing.T) {
		testMessagesList(t, client, factory, workspace.ID)
	})

	t.Run("Broadcast Stats Endpoint", func(t *testing.T) {
		testBroadcastStats(t, client, factory, workspace.ID)
	})
}

func testMessagesList(t *testing.T, client *testutil.APIClient, factory *testutil.TestDataFactory, workspaceID string) {
	t.Run("GET /api/messages.list", func(t *testing.T) {
		t.Run("should return empty list when no messages exist", func(t *testing.T) {
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result domain.MessageListResult
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Empty(t, result.Messages)
			assert.Empty(t, result.NextCursor)
			assert.False(t, result.HasMore)
		})

		t.Run("should return 400 when workspace_id is missing", func(t *testing.T) {
			// Clear workspace_id from client to test missing workspace_id scenario
			originalWorkspaceID := client.GetWorkspaceID()
			client.SetWorkspaceID("")
			defer client.SetWorkspaceID(originalWorkspaceID) // Restore for other tests

			resp, err := client.Get("/api/messages.list")
			require.NoError(t, err)
			defer resp.Body.Close()

			// Debug: print response status and body
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response status: %d, body: %s", resp.StatusCode, string(body))

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should return 405 for non-GET methods", func(t *testing.T) {
			resp, err := client.Post("/api/messages.list", nil, map[string]string{
				"workspace_id": workspaceID,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})

		t.Run("should handle pagination parameters", func(t *testing.T) {
			// Create some test messages
			contact, err := factory.CreateContact(workspaceID)
			require.NoError(t, err)
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)
			broadcast, err := factory.CreateBroadcast(workspaceID)
			require.NoError(t, err)

			// Create multiple messages
			for i := 0; i < 5; i++ {
				_, err := factory.CreateMessageHistory(workspaceID, testutil.WithMessageContact(contact.Email),
					testutil.WithMessageTemplate(template.ID), testutil.WithMessageBroadcast(broadcast.ID))
				require.NoError(t, err)
			}

			// Test with limit
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"limit":        "3",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result domain.MessageListResult
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Len(t, result.Messages, 3)
			assert.True(t, result.HasMore)
			assert.NotEmpty(t, result.NextCursor)
		})

		t.Run("should handle filter parameters", func(t *testing.T) {
			// Create contact and template
			contact, err := factory.CreateContact(workspaceID)
			require.NoError(t, err)
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)

			// Create messages with different channels
			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageChannel("email"))
			require.NoError(t, err)

			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageChannel("sms"))
			require.NoError(t, err)

			// Filter by channel
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"channel":      "email",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result domain.MessageListResult
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			// Should only return email messages
			for _, msg := range result.Messages {
				assert.Equal(t, "email", msg.Channel)
			}
		})

		t.Run("should validate filter parameters", func(t *testing.T) {
			// Test invalid channel
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"channel":      "invalid_channel",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should handle boolean filters", func(t *testing.T) {
			// Create contact and template
			contact, err := factory.CreateContact(workspaceID)
			require.NoError(t, err)
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)

			// Create messages with different statuses
			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageDelivered(true))
			require.NoError(t, err)

			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageDelivered(false))
			require.NoError(t, err)

			// Filter by delivered status
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"is_delivered": "true",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result domain.MessageListResult
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			// Should only return delivered messages
			for _, msg := range result.Messages {
				assert.NotNil(t, msg.DeliveredAt)
			}
		})

		t.Run("should handle time range filters", func(t *testing.T) {
			// Create contact and template
			contact, err := factory.CreateContact(workspaceID)
			require.NoError(t, err)
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)

			now := time.Now().UTC()

			// Create messages with different sent times
			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageSentAt(now.Add(-2*time.Hour)))
			require.NoError(t, err)

			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageSentAt(now.Add(-1*time.Hour)))
			require.NoError(t, err)

			// Filter by time range
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"sent_after":   now.Add(-90 * time.Minute).Format(time.RFC3339),
				"sent_before":  now.Format(time.RFC3339),
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result domain.MessageListResult
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			// Should only return messages within the time range
			for _, msg := range result.Messages {
				assert.True(t, msg.SentAt.After(now.Add(-90*time.Minute)))
				assert.True(t, msg.SentAt.Before(now))
			}
		})

		t.Run("should handle invalid time format", func(t *testing.T) {
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspaceID,
				"sent_after":   "invalid-time-format",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}

func testBroadcastStats(t *testing.T, client *testutil.APIClient, factory *testutil.TestDataFactory, workspaceID string) {
	t.Run("GET /api/messages.broadcastStats", func(t *testing.T) {
		t.Run("should return stats for existing broadcast", func(t *testing.T) {
			// Create test data
			contact, err := factory.CreateContact(workspaceID)
			require.NoError(t, err)
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)
			broadcast, err := factory.CreateBroadcast(workspaceID)
			require.NoError(t, err)

			// Create messages with different statuses
			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageBroadcast(broadcast.ID),
				testutil.WithMessageDelivered(true))
			require.NoError(t, err)

			_, err = factory.CreateMessageHistory(workspaceID,
				testutil.WithMessageContact(contact.Email),
				testutil.WithMessageTemplate(template.ID),
				testutil.WithMessageBroadcast(broadcast.ID),
				testutil.WithMessageOpened(true))
			require.NoError(t, err)

			resp, err := client.Get("/api/messages.broadcastStats", map[string]string{
				"workspace_id": workspaceID,
				"broadcast_id": broadcast.ID,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, broadcast.ID, result["broadcast_id"])
			assert.Contains(t, result, "stats")

			stats, ok := result["stats"].(map[string]interface{})
			require.True(t, ok)
			assert.Contains(t, stats, "total_sent")
			assert.Contains(t, stats, "total_delivered")
			assert.Contains(t, stats, "total_opened")
		})

		t.Run("should return 400 when broadcast_id is missing", func(t *testing.T) {
			resp, err := client.Get("/api/messages.broadcastStats", map[string]string{
				"workspace_id": workspaceID,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should return 400 when workspace_id is missing", func(t *testing.T) {
			// Clear workspace_id from client to test missing workspace_id scenario
			originalWorkspaceID := client.GetWorkspaceID()
			client.SetWorkspaceID("")
			defer client.SetWorkspaceID(originalWorkspaceID) // Restore for other tests

			resp, err := client.Get("/api/messages.broadcastStats", map[string]string{
				"broadcast_id": "test-broadcast-id",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			// Debug: print response status and body
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response status: %d, body: %s", resp.StatusCode, string(body))

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should return 405 for non-GET methods", func(t *testing.T) {
			resp, err := client.Post("/api/messages.broadcastStats", nil, map[string]string{
				"workspace_id": workspaceID,
				"broadcast_id": "test-broadcast-id",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})

		t.Run("should handle non-existent broadcast", func(t *testing.T) {
			resp, err := client.Get("/api/messages.broadcastStats", map[string]string{
				"workspace_id": workspaceID,
				"broadcast_id": "non-existent-broadcast-id",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should return OK with empty stats
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, "non-existent-broadcast-id", result["broadcast_id"])
			assert.Contains(t, result, "stats")
		})
	})
}

func TestMessageHistoryAuthentication(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	client := suite.APIClient
	factory := suite.DataFactory

	// Create test user and workspace
	user, err := factory.CreateUser()
	require.NoError(t, err)
	workspace, err := factory.CreateWorkspace()
	require.NoError(t, err)

	// Add user to workspace as owner
	err = factory.AddUserToWorkspace(user.ID, workspace.ID, "owner")
	require.NoError(t, err)

	t.Run("should require authentication", func(t *testing.T) {
		// Don't login, make requests without auth
		client.SetToken("")

		t.Run("messages.list", func(t *testing.T) {
			resp, err := client.Get("/api/messages.list", map[string]string{
				"workspace_id": workspace.ID,
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("messages.broadcastStats", func(t *testing.T) {
			resp, err := client.Get("/api/messages.broadcastStats", map[string]string{
				"workspace_id": workspace.ID,
				"broadcast_id": "test-broadcast-id",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}

func TestMessageHistoryDataFactory(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	factory := suite.DataFactory
	workspace, err := factory.CreateWorkspace()
	require.NoError(t, err)

	t.Run("CreateMessageHistory", func(t *testing.T) {
		contact, err := factory.CreateContact(workspace.ID)
		require.NoError(t, err)
		template, err := factory.CreateTemplate(workspace.ID)
		require.NoError(t, err)

		message, err := factory.CreateMessageHistory(workspace.ID,
			testutil.WithMessageContact(contact.Email),
			testutil.WithMessageTemplate(template.ID))
		require.NoError(t, err)
		require.NotNil(t, message)

		assert.NotEmpty(t, message.ID)
		assert.Equal(t, contact.Email, message.ContactEmail)
		assert.Equal(t, template.ID, message.TemplateID)
		assert.Equal(t, "email", message.Channel) // default channel
		assert.NotZero(t, message.SentAt)
		assert.NotZero(t, message.CreatedAt)
		assert.NotZero(t, message.UpdatedAt)
	})

	t.Run("CreateMessageHistory with options", func(t *testing.T) {
		contact, err := factory.CreateContact(workspace.ID)
		require.NoError(t, err)
		template, err := factory.CreateTemplate(workspace.ID)
		require.NoError(t, err)
		broadcast, err := factory.CreateBroadcast(workspace.ID)
		require.NoError(t, err)

		now := time.Now().UTC()
		message, err := factory.CreateMessageHistory(workspace.ID,
			testutil.WithMessageContact(contact.Email),
			testutil.WithMessageTemplate(template.ID),
			testutil.WithMessageBroadcast(broadcast.ID),
			testutil.WithMessageChannel("sms"),
			testutil.WithMessageSentAt(now),
			testutil.WithMessageDelivered(true),
			testutil.WithMessageOpened(true))
		require.NoError(t, err)

		assert.Equal(t, contact.Email, message.ContactEmail)
		assert.Equal(t, template.ID, message.TemplateID)
		assert.Equal(t, broadcast.ID, *message.BroadcastID)
		assert.Equal(t, "sms", message.Channel)
		assert.Equal(t, now.Format(time.RFC3339), message.SentAt.Format(time.RFC3339))
		assert.NotNil(t, message.DeliveredAt)
		assert.NotNil(t, message.OpenedAt)
	})

	t.Run("CreateMessageHistory persisted to database", func(t *testing.T) {
		contact, err := factory.CreateContact(workspace.ID)
		require.NoError(t, err)
		template, err := factory.CreateTemplate(workspace.ID)
		require.NoError(t, err)

		message, err := factory.CreateMessageHistory(workspace.ID,
			testutil.WithMessageContact(contact.Email),
			testutil.WithMessageTemplate(template.ID))
		require.NoError(t, err)

		// Verify message exists in database using repository
		app := suite.ServerManager.GetApp()
		messageHistoryRepo := app.GetMessageHistoryRepository()

		retrievedMessage, err := messageHistoryRepo.Get(context.Background(), workspace.ID, message.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedMessage)
		assert.Equal(t, contact.Email, retrievedMessage.ContactEmail)
		assert.Equal(t, template.ID, retrievedMessage.TemplateID)
	})
}
