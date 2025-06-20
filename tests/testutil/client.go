package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient provides HTTP client functionality for integration tests
type APIClient struct {
	baseURL     string
	client      *http.Client
	token       string
	workspaceID string
}

// NewAPIClient creates a new API client for testing
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (c *APIClient) SetToken(token string) {
	c.token = token
}

// SetWorkspaceID sets the default workspace ID for requests
func (c *APIClient) SetWorkspaceID(workspaceID string) {
	c.workspaceID = workspaceID
}

// GetWorkspaceID returns the current workspace ID
func (c *APIClient) GetWorkspaceID() string {
	return c.workspaceID
}

// Login authenticates and sets the token
func (c *APIClient) Login(email, password string) error {
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.Post("/api/auth.login", loginReq)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = loginResp.Token
	return nil
}

// Get makes a GET request
func (c *APIClient) Get(endpoint string, params ...map[string]string) (*http.Response, error) {
	return c.request(http.MethodGet, endpoint, nil, params...)
}

// Post makes a POST request
func (c *APIClient) Post(endpoint string, body interface{}, params ...map[string]string) (*http.Response, error) {
	return c.request(http.MethodPost, endpoint, body, params...)
}

// Put makes a PUT request
func (c *APIClient) Put(endpoint string, body interface{}, params ...map[string]string) (*http.Response, error) {
	return c.request(http.MethodPut, endpoint, body, params...)
}

// Delete makes a DELETE request
func (c *APIClient) Delete(endpoint string, params ...map[string]string) (*http.Response, error) {
	return c.request(http.MethodDelete, endpoint, nil, params...)
}

// request makes an HTTP request
func (c *APIClient) request(method, endpoint string, body interface{}, params ...map[string]string) (*http.Response, error) {
	// Build URL with query parameters
	reqURL := c.baseURL + endpoint
	if len(params) > 0 && params[0] != nil {
		urlParams := url.Values{}
		for key, value := range params[0] {
			urlParams.Add(key, value)
		}
		if len(urlParams) > 0 {
			reqURL += "?" + urlParams.Encode()
		}
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication token if available
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Add workspace ID if available and not already in params
	if c.workspaceID != "" && !strings.Contains(reqURL, "workspace_id=") {
		q := req.URL.Query()
		q.Add("workspace_id", c.workspaceID)
		req.URL.RawQuery = q.Encode()
	}

	// Make request
	return c.client.Do(req)
}

// GetJSON makes a GET request and decodes JSON response
func (c *APIClient) GetJSON(endpoint string, result interface{}, params ...map[string]string) error {
	resp, err := c.Get(endpoint, params...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// PostJSON makes a POST request and decodes JSON response
func (c *APIClient) PostJSON(endpoint string, reqBody interface{}, result interface{}, params ...map[string]string) error {
	resp, err := c.Post(endpoint, reqBody, params...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

// ExpectStatus checks if response has expected status code
func (c *APIClient) ExpectStatus(resp *http.Response, expectedStatus int) error {
	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status %d, got %d: %s", expectedStatus, resp.StatusCode, string(body))
	}
	return nil
}

// ReadBody reads and returns response body as string
func (c *APIClient) ReadBody(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DecodeJSON decodes response body as JSON
func (c *APIClient) DecodeJSON(resp *http.Response, result interface{}) error {
	return json.NewDecoder(resp.Body).Decode(result)
}

// Broadcast API helpers
func (c *APIClient) CreateBroadcast(broadcast map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/broadcasts.create", broadcast)
}

func (c *APIClient) GetBroadcast(broadcastID string) (*http.Response, error) {
	params := map[string]string{
		"id": broadcastID,
	}
	return c.Get("/api/broadcasts.get", params)
}

func (c *APIClient) ListBroadcasts(params map[string]string) (*http.Response, error) {
	return c.Get("/api/broadcasts.list", params)
}

// Contact API helpers
func (c *APIClient) CreateContact(contact map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/contacts.upsert", contact)
}

func (c *APIClient) GetContactByEmail(email string) (*http.Response, error) {
	params := map[string]string{
		"email": email,
	}
	return c.Get("/api/contacts.getByEmail", params)
}

func (c *APIClient) ListContacts(params map[string]string) (*http.Response, error) {
	return c.Get("/api/contacts.list", params)
}

// Template API helpers
func (c *APIClient) CreateTemplate(template map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/templates.create", template)
}

func (c *APIClient) GetTemplate(templateID string) (*http.Response, error) {
	params := map[string]string{
		"id": templateID,
	}
	return c.Get("/api/templates.get", params)
}

func (c *APIClient) UpdateTemplate(template map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/templates.update", template)
}

func (c *APIClient) DeleteTemplate(workspaceID, templateID string) (*http.Response, error) {
	deleteReq := map[string]interface{}{
		"workspace_id": workspaceID,
		"id":           templateID,
	}
	return c.Post("/api/templates.delete", deleteReq)
}

func (c *APIClient) CompileTemplate(compileReq map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/templates.compile", compileReq)
}

func (c *APIClient) ListTemplates(params map[string]string) (*http.Response, error) {
	return c.Get("/api/templates.list", params)
}

// Workspace API helpers
func (c *APIClient) CreateWorkspace(workspace map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/workspaces.create", workspace)
}

func (c *APIClient) GetWorkspace(workspaceID string) (*http.Response, error) {
	params := map[string]string{
		"id": workspaceID,
	}
	return c.Get("/api/workspaces.get", params)
}

// List API helpers
func (c *APIClient) CreateList(list map[string]interface{}) (*http.Response, error) {
	return c.Post("/api/lists.create", list)
}

func (c *APIClient) GetList(listID string) (*http.Response, error) {
	params := map[string]string{
		"id": listID,
	}
	return c.Get("/api/lists.get", params)
}

func (c *APIClient) ListLists(params map[string]string) (*http.Response, error) {
	return c.Get("/api/lists.list", params)
}
