package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ytsruh.com/envoy/pkg/config"
)

type Client struct {
	serverURL string
	token     string
	client    *http.Client
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewClient() (*Client, error) {
	serverURL, err := config.GetServerURL()
	if err != nil {
		return nil, err
	}

	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}

	return &Client{
		serverURL: serverURL,
		token:     token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) buildURL(path string) string {
	return c.serverURL + path
}

func (c *Client) doRequest(method, path string, body interface{}, authRequired bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.buildURL(path), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if authRequired && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			defer resp.Body.Close()
			return resp, fmt.Errorf("server error: %s", errResp.Error)
		}
	}

	return resp, nil
}

func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) Register(name, email, password string) (*AuthResponse, error) {
	reqBody := map[string]any{
		"name":     name,
		"email":    email,
		"password": password,
	}

	resp, err := c.doRequest("POST", "/auth/register", reqBody, false)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := c.decodeResponse(resp, &authResp); err != nil {
		return nil, err
	}

	if err := config.SetToken(authResp.Token); err != nil {
		return nil, err
	}

	c.SetToken(authResp.Token)

	return &authResp, nil
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	reqBody := map[string]any{
		"email":    email,
		"password": password,
	}

	resp, err := c.doRequest("POST", "/auth/login", reqBody, false)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := c.decodeResponse(resp, &authResp); err != nil {
		return nil, err
	}

	if err := config.SetToken(authResp.Token); err != nil {
		return nil, err
	}

	c.SetToken(authResp.Token)

	return &authResp, nil
}

func (c *Client) GetProfile() (*ProfileResponse, error) {
	resp, err := c.doRequest("GET", "/auth/profile", nil, true)
	if err != nil {
		return nil, err
	}

	var profileResp ProfileResponse
	if err := c.decodeResponse(resp, &profileResp); err != nil {
		return nil, err
	}

	return &profileResp, nil
}

func (c *Client) CreateProject(name, description, gitRepo string) (*ProjectResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	}
	if gitRepo != "" {
		reqBody["git_repo"] = gitRepo
	}

	resp, err := c.doRequest("POST", "/projects", reqBody, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := c.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (c *Client) ListProjects() ([]ProjectResponse, error) {
	resp, err := c.doRequest("GET", "/projects", nil, true)
	if err != nil {
		return nil, err
	}

	var projects []ProjectResponse
	if err := c.decodeResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) GetProject(projectID int64) (*ProjectResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/projects/%d", projectID), nil, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := c.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (c *Client) UpdateProject(projectID int64, name, description, gitRepo string) (*ProjectResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	}
	if gitRepo != "" {
		reqBody["git_repo"] = gitRepo
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/projects/%d", projectID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := c.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (c *Client) DeleteProject(projectID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/projects/%d", projectID), nil, true)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  UserData `json:"user"`
}

type UserData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Iat    int64  `json:"issued_at"`
	Exp    int64  `json:"expires_at"`
}

type ProjectResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	GitRepo     string `json:"git_repo,omitempty"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
