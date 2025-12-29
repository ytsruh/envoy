package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ytsruh.com/envoy/cli/config"
)

type Client struct {
	serverURL string
	token     string
	client    *http.Client
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrExpiredToken = fmt.Errorf("token has expired")
	ErrNoToken      = fmt.Errorf("not logged in")
)

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

func RequireToken() (*Client, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	if client.token == "" {
		return nil, ErrNoToken
	}

	return client, nil
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

			if resp.StatusCode == http.StatusUnauthorized && errResp.Error == "Token has expired" {
				if err := config.ClearToken(); err == nil {
					c.token = ""
				}
				return resp, ErrExpiredToken
			}

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
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d", projectID), nil, true)
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

	resp, err := c.doRequest("PUT", fmt.Sprintf("/projects/%d", projectID), reqBody, true)
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

func (c *Client) CreateEnvironment(projectID int64, name, description string) (*EnvironmentResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/projects/%d/environments", projectID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := c.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (c *Client) ListEnvironments(projectID int64) ([]EnvironmentResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d/environments", projectID), nil, true)
	if err != nil {
		return nil, err
	}

	var environments []EnvironmentResponse
	if err := c.decodeResponse(resp, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func (c *Client) GetEnvironment(projectID, environmentID int64) (*EnvironmentResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), nil, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := c.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (c *Client) UpdateEnvironment(projectID, environmentID int64, name, description string) (*EnvironmentResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := c.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (c *Client) DeleteEnvironment(projectID, environmentID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), nil, true)
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

func (c *Client) CreateEnvironmentVariable(projectID, environmentID int64, key, value string) (*EnvironmentVariableResponse, error) {
	reqBody := map[string]any{
		"key":   key,
		"value": value,
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/projects/%d/environments/%d/variables", projectID, environmentID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := c.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (c *Client) ListEnvironmentVariables(projectID, environmentID int64) ([]EnvironmentVariableResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d/variables", projectID, environmentID), nil, true)
	if err != nil {
		return nil, err
	}

	var variables []EnvironmentVariableResponse
	if err := c.decodeResponse(resp, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

func (c *Client) GetEnvironmentVariable(projectID, environmentID, variableID int64) (*EnvironmentVariableResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), nil, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := c.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (c *Client) UpdateEnvironmentVariable(projectID, environmentID, variableID int64, key, value string) (*EnvironmentVariableResponse, error) {
	reqBody := map[string]any{
		"key":   key,
		"value": value,
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := c.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (c *Client) DeleteEnvironmentVariable(projectID, environmentID, variableID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), nil, true)
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

type EnvironmentResponse struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type EnvironmentVariableResponse struct {
	ID            int64  `json:"id"`
	EnvironmentID int64  `json:"environment_id"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
