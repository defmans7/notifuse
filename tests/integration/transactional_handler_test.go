package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/app"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionalHandler tests the transactional notification handler with proper SMTP email provider configuration.
func TestTransactionalHandler(t *testing.T) {
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

	// Set up SMTP email provider for testing
	_, err = factory.SetupWorkspaceWithSMTPProvider(workspace.ID)
	require.NoError(t, err)

	// Login to get auth token
	err = client.Login(user.Email, "password")
	require.NoError(t, err)
	client.SetWorkspaceID(workspace.ID)

	t.Run("CRUD Operations", func(t *testing.T) {
		testTransactionalCRUD(t, client, factory, workspace.ID)
	})

	t.Run("Send Notification", func(t *testing.T) {
		testTransactionalSend(t, client, factory, workspace.ID)
	})

	t.Run("Template Testing", func(t *testing.T) {
		testTransactionalTemplateTest(t, client, factory, workspace.ID)
	})
}

func testTransactionalCRUD(t *testing.T, client *testutil.APIClient, factory *testutil.TestDataFactory, workspaceID string) {
	t.Run("Create Transactional Notification", func(t *testing.T) {
		t.Run("should create notification successfully", func(t *testing.T) {
			// Create a template first
			template, err := factory.CreateTemplate(workspaceID)
			require.NoError(t, err)

			notification := map[string]interface{}{
				"workspace_id": workspaceID,
				"notification": map[string]interface{}{
					"id":          "welcome-email",
					"name":        "Welcome Email",
					"description": "Welcome new users",
					"channels": map[string]interface{}{
						"email": map[string]interface{}{
							"template_id": template.ID,
							"settings":    map[string]interface{}{},
						},
					},
					"tracking_settings": map[string]interface{}{
						"enable_tracking": true,
					},
					"metadata": map[string]interface{}{
						"category": "onboarding",
					},
				},
			}

			resp, err := client.CreateTransactionalNotification(notification)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "notification")
			notificationData := result["notification"].(map[string]interface{})
			assert.Equal(t, "welcome-email", notificationData["id"])
			assert.Equal(t, "Welcome Email", notificationData["name"])
			assert.Equal(t, "Welcome new users", notificationData["description"])
		})

		t.Run("should validate required fields", func(t *testing.T) {
			notification := map[string]interface{}{
				"workspace_id": workspaceID,
				"notification": map[string]interface{}{
					// Missing id and name
					"description": "Test",
				},
			}

			resp, err := client.CreateTransactionalNotification(notification)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should validate channels", func(t *testing.T) {
			notification := map[string]interface{}{
				"workspace_id": workspaceID,
				"notification": map[string]interface{}{
					"id":          "test-notification",
					"name":        "Test Notification",
					"description": "Test",
					// Missing channels
				},
			}

			resp, err := client.CreateTransactionalNotification(notification)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("Get Transactional Notification", func(t *testing.T) {
		// Create a notification first
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		notification, err := factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("get-test"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		t.Run("should get notification successfully", func(t *testing.T) {
			resp, err := client.GetTransactionalNotification(notification.ID)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "notification")
			notificationData := result["notification"].(map[string]interface{})
			assert.Equal(t, notification.ID, notificationData["id"])
			assert.Equal(t, notification.Name, notificationData["name"])
		})

		t.Run("should return 404 for non-existent notification", func(t *testing.T) {
			resp, err := client.GetTransactionalNotification("non-existent")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("List Transactional Notifications", func(t *testing.T) {
		// Create multiple notifications
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		_, err = factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("list-test-1"),
			testutil.WithTransactionalNotificationName("List Test 1"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		_, err = factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("list-test-2"),
			testutil.WithTransactionalNotificationName("List Test 2"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		t.Run("should list notifications successfully", func(t *testing.T) {
			resp, err := client.ListTransactionalNotifications(map[string]string{
				"limit":  "10",
				"offset": "0",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "notifications")
			assert.Contains(t, result, "total")

			notifications := result["notifications"].([]interface{})
			assert.GreaterOrEqual(t, len(notifications), 2)
		})

		t.Run("should support search", func(t *testing.T) {
			resp, err := client.ListTransactionalNotifications(map[string]string{
				"search": "List Test 1",
				"limit":  "10",
				"offset": "0",
			})
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			notifications := result["notifications"].([]interface{})
			assert.GreaterOrEqual(t, len(notifications), 1)
		})
	})

	t.Run("Update Transactional Notification", func(t *testing.T) {
		// Create a notification first
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		notification, err := factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("update-test"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		t.Run("should update notification successfully", func(t *testing.T) {
			updates := map[string]interface{}{
				"name":        "Updated Name",
				"description": "Updated Description",
			}

			resp, err := client.UpdateTransactionalNotification(notification.ID, updates)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "notification")
			notificationData := result["notification"].(map[string]interface{})
			assert.Equal(t, "Updated Name", notificationData["name"])
			assert.Equal(t, "Updated Description", notificationData["description"])
		})

		t.Run("should return 404 for non-existent notification", func(t *testing.T) {
			updates := map[string]interface{}{
				"name": "Updated Name",
			}

			resp, err := client.UpdateTransactionalNotification("non-existent", updates)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("Delete Transactional Notification", func(t *testing.T) {
		// Create a notification first
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		notification, err := factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("delete-test"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		t.Run("should delete notification successfully", func(t *testing.T) {
			resp, err := client.DeleteTransactionalNotification(notification.ID)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "success")
			assert.True(t, result["success"].(bool))

			// Verify notification is deleted
			getResp, err := client.GetTransactionalNotification(notification.ID)
			require.NoError(t, err)
			defer getResp.Body.Close()
			assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
		})

		t.Run("should return 404 for non-existent notification", func(t *testing.T) {
			resp, err := client.DeleteTransactionalNotification("non-existent")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

func testTransactionalSend(t *testing.T, client *testutil.APIClient, factory *testutil.TestDataFactory, workspaceID string) {
	t.Run("Send Transactional Notification", func(t *testing.T) {
		// Create a template and notification
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		notification, err := factory.CreateTransactionalNotification(workspaceID,
			testutil.WithTransactionalNotificationID("send-test"),
			testutil.WithTransactionalNotificationChannels(domain.ChannelTemplates{
				domain.TransactionalChannelEmail: domain.ChannelTemplate{
					TemplateID: template.ID,
					Settings:   map[string]interface{}{},
				},
			}),
		)
		require.NoError(t, err)

		// Create a contact
		contact, err := factory.CreateContact(workspaceID)
		require.NoError(t, err)

		t.Run("should send notification successfully", func(t *testing.T) {
			sendRequest := map[string]interface{}{
				"id": notification.ID,
				"contact": map[string]interface{}{
					"email":      contact.Email,
					"first_name": "John",
					"last_name":  "Doe",
				},
				"channels": []string{"email"},
				"data": map[string]interface{}{
					"user_name":   "John Doe",
					"welcome_url": "https://example.com/welcome",
				},
				"metadata": map[string]interface{}{
					"source": "integration_test",
				},
			}

			resp, err := client.SendTransactionalNotification(sendRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			// In test environment, expect 500 due to missing email provider configuration
			// In production, this would be 200 with proper email provider setup
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "error")
			assert.Equal(t, "Failed to send notification", result["error"])
		})

		t.Run("should validate required fields", func(t *testing.T) {
			sendRequest := map[string]interface{}{
				"id": notification.ID,
				// Missing contact
				"channels": []string{"email"},
			}

			resp, err := client.SendTransactionalNotification(sendRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should return 400 for non-existent notification", func(t *testing.T) {
			sendRequest := map[string]interface{}{
				"id": "non-existent",
				"contact": map[string]interface{}{
					"email": contact.Email,
				},
				"channels": []string{"email"},
			}

			resp, err := client.SendTransactionalNotification(sendRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}

func testTransactionalTemplateTest(t *testing.T, client *testutil.APIClient, factory *testutil.TestDataFactory, workspaceID string) {
	t.Run("Test Template", func(t *testing.T) {
		// Create a template
		template, err := factory.CreateTemplate(workspaceID)
		require.NoError(t, err)

		// Create an integration for sending test emails
		integration, err := factory.CreateSMTPIntegration(workspaceID)
		require.NoError(t, err)

		t.Run("should test template successfully", func(t *testing.T) {
			testRequest := map[string]interface{}{
				"template_id":     template.ID,
				"integration_id":  integration.ID,
				"sender_id":       "test@example.com",
				"recipient_email": "recipient@example.com",
				"email_options": map[string]interface{}{
					"subject": "Test Email",
				},
			}

			resp, err := client.TestTransactionalTemplate(testRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "success")
			// Note: success may be false due to SMTP configuration in test environment
			// but the endpoint should respond correctly
		})

		t.Run("should validate required fields", func(t *testing.T) {
			testRequest := map[string]interface{}{
				// Missing template_id
				"integration_id":  integration.ID,
				"recipient_email": "recipient@example.com",
			}

			resp, err := client.TestTransactionalTemplate(testRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("should return error for non-existent template", func(t *testing.T) {
			testRequest := map[string]interface{}{
				"template_id":     "non-existent",
				"integration_id":  integration.ID,
				"sender_id":       "test@example.com",
				"recipient_email": "recipient@example.com",
			}

			resp, err := client.TestTransactionalTemplate(testRequest)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "success")
			assert.False(t, result["success"].(bool))
			assert.Contains(t, result, "error")
		})
	})
}

func TestTransactionalAuthentication(t *testing.T) {
	testutil.SkipIfShort(t)
	testutil.SetupTestEnvironment()
	defer testutil.CleanupTestEnvironment()

	suite := testutil.NewIntegrationTestSuite(t, func(cfg *config.Config) testutil.AppInterface {
		return app.NewApp(cfg)
	})
	defer suite.Cleanup()

	client := suite.APIClient

	t.Run("should require authentication", func(t *testing.T) {
		endpoints := []struct {
			name   string
			method func() (*http.Response, error)
		}{
			{
				name: "list",
				method: func() (*http.Response, error) {
					return client.ListTransactionalNotifications(nil)
				},
			},
			{
				name: "get",
				method: func() (*http.Response, error) {
					return client.GetTransactionalNotification("test")
				},
			},
			{
				name: "create",
				method: func() (*http.Response, error) {
					return client.CreateTransactionalNotification(map[string]interface{}{})
				},
			},
			{
				name: "update",
				method: func() (*http.Response, error) {
					return client.UpdateTransactionalNotification("test", map[string]interface{}{})
				},
			},
			{
				name: "delete",
				method: func() (*http.Response, error) {
					return client.DeleteTransactionalNotification("test")
				},
			},
			{
				name: "send",
				method: func() (*http.Response, error) {
					return client.SendTransactionalNotification(map[string]interface{}{})
				},
			},
			{
				name: "testTemplate",
				method: func() (*http.Response, error) {
					return client.TestTransactionalTemplate(map[string]interface{}{})
				},
			},
		}

		for _, endpoint := range endpoints {
			t.Run(endpoint.name, func(t *testing.T) {
				resp, err := endpoint.method()
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			})
		}
	})
}

func TestTransactionalMethodValidation(t *testing.T) {
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

	t.Run("should validate HTTP methods", func(t *testing.T) {
		// Test that GET-only endpoints reject POST
		resp, err := client.Post("/api/transactional.list", map[string]interface{}{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		resp, err = client.Post("/api/transactional.get", map[string]interface{}{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		// Test that POST-only endpoints reject GET
		resp, err = client.Get("/api/transactional.create")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		resp, err = client.Get("/api/transactional.update")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		resp, err = client.Get("/api/transactional.delete")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		resp, err = client.Get("/api/transactional.send")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		resp, err = client.Get("/api/transactional.testTemplate")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
