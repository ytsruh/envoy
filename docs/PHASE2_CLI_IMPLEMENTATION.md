# Phase 2: CLI Client Implementation

## Overview
Create a Cobra-based CLI client for the Envoy server with authentication, git-aware project management, and secure token storage.

---

## Step 1: Add Cobra Dependency

```bash
go get github.com/spf13/cobra@latest
go mod tidy
```

---

## Step 2: Create Config Package

**New File: `pkg/config/config.go`**

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName   = "config.json"
	projectsFileName = "projects.json"
)

type Config struct {
	AuthToken string `json:"auth_token"`
	ServerURL string `json:"server_url,omitempty"`
}

type ProjectTracking struct {
	ProjectID int64  `json:"project_id"`
	GitRepo   string `json:"git_repo"`
	CreatedAt string `json:"created_at"`
	RemoteURL string `json:"remote_url"`
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".envoy")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return configDir, nil
}

func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

func GetProjectsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, projectsFileName), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func LoadProjects() (map[string]ProjectTracking, error) {
	projectsPath, err := GetProjectsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(projectsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]ProjectTracking), nil
		}
		return nil, err
	}
	var projects map[string]ProjectTracking
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func SaveProjects(projects map[string]ProjectTracking) error {
	projectsPath, err := GetProjectsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(projectsPath, data, 0600)
}

func SetAuthToken(token string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg.AuthToken = token
	return SaveConfig(cfg)
}

func GetAuthToken() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return cfg.AuthToken, nil
}

func GetServerURL() string {
	if url := os.Getenv("ENVOY_SERVER_URL"); url != "" {
		return url
	}
	cfg, _ := LoadConfig()
	if cfg.ServerURL != "" {
		return cfg.ServerURL
	}
	return "http://localhost:8080"
}
```

---

## Step 3: Create CLI Client

**New File: `pkg/cli/client.go`**

```go
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ytsruh.com/envoy/pkg/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:    config.GetServerURL(),
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, authToken string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	return c.httpClient.Do(req)
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	} `json:"user"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GitRepo     string `json:"git_repo"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GitRepo     string `json:"git_repo,omitempty"`
}

type ProjectResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description *string `json:"description"`
	GitRepo     string `json:"git_repo"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (c *Client) Register(name, email, password string) (*AuthResponse, error) {
	req := RegisterRequest{Name: name, Email: email, Password: password}
	resp, err := c.doRequest("POST", "/auth/register", req, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}
	return &authResp, nil
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	req := LoginRequest{Email: email, Password: password}
	resp, err := c.doRequest("POST", "/auth/login", req, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}
	return &authResp, nil
}

func (c *Client) CreateProject(token, name, description, gitRepo string) (*ProjectResponse, error) {
	req := CreateProjectRequest{Name: name, Description: description, GitRepo: gitRepo}
	resp, err := c.doRequest("POST", "/projects", req, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var projResp ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projResp); err != nil {
		return nil, err
	}
	return &projResp, nil
}

func (c *Client) GetProject(token string, id int64) (*ProjectResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/projects/%d", id), nil, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var projResp ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projResp); err != nil {
		return nil, err
	}
	return &projResp, nil
}

func (c *Client) ListProjects(token string) ([]ProjectResponse, error) {
	resp, err := c.doRequest("GET", "/projects", nil, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var projects []ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) UpdateProject(token string, id int64, name, description string) (*ProjectResponse, error) {
	req := UpdateProjectRequest{Name: name, Description: description}
	resp, err := c.doRequest("PUT", fmt.Sprintf("/projects/%d", id), req, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf(errResp.Error)
	}

	var projResp ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projResp); err != nil {
		return nil, err
	}
	return &projResp, nil
}

func (c *Client) DeleteProject(token string, id int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/projects/%d", id), nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf(errResp.Error)
	}
	return nil
}
```

---

## Step 4: Create Git Validation

**New File: `pkg/cli/git.go`**

```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"ytsruh.com/envoy/pkg/config"
	"time"
)

type GitInfo struct {
	IsGitRepo bool
	GitDir    string
	RemoteURL string
	GitRepo   string
}

func DetectGitRepo() (*GitInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	gitDir := filepath.Join(cwd, ".git")
	info := &GitInfo{}

	if _, err := os.Stat(gitDir); err == nil {
		info.IsGitRepo = true
		info.GitDir = gitDir
	}

	if info.IsGitRepo {
		configPath := filepath.Join(gitDir, "config")
		if file, err := os.Open(configPath); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "url") {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						info.RemoteURL = strings.TrimSpace(parts[1])
						info.GitRepo = extractGitRepo(info.RemoteURL)
						break
					}
				}
			}
		}
	}

	return info, nil
}

func extractGitRepo(remoteURL string) string {
	repo := remoteURL

	if strings.HasPrefix(repo, "https://") {
		repo = strings.TrimPrefix(repo, "https://")
	} else if strings.HasPrefix(repo, "git@") {
		repo = strings.TrimPrefix(repo, "git@")
		repo = strings.Replace(repo, ":", "/", 1)
	}

	if strings.Contains(repo, ".git") {
		repo = strings.TrimSuffix(repo, ".git")
	}

	parts := strings.Split(repo, "/")
	if len(parts) >= 2 {
		owner := parts[len(parts)-2]
		name := parts[len(parts)-1]
		return fmt.Sprintf("%s/%s", owner, name)
	}

	return ""
}

func IsGitRepoTracked(gitRepo string) (bool, error) {
	if gitRepo == "" {
		return false, nil
	}

	projects, err := config.LoadProjects()
	if err != nil {
		return false, err
	}

	_, exists := projects[gitRepo]
	return exists, nil
}

func TrackGitRepo(gitRepo string, projectID int64, remoteURL string) error {
	if gitRepo == "" {
		return nil
	}

	projects, err := config.LoadProjects()
	if err != nil {
		return err
	}

	projects[gitRepo] = config.ProjectTracking{
		ProjectID: projectID,
		GitRepo:   gitRepo,
		CreatedAt: time.Now().Format(time.RFC3339),
		RemoteURL: remoteURL,
	}

	return config.SaveProjects(projects)
}

func UntrackGitRepo(gitRepo string) error {
	if gitRepo == "" {
		return nil
	}

	projects, err := config.LoadProjects()
	if err != nil {
		return err
	}

	delete(projects, gitRepo)
	return config.SaveProjects(projects)
}
```

---

## Step 5: Create Prompts

**New File: `pkg/cli/prompts.go`**

```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"golang.org/x/term"
)

func PromptString(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func PromptPassword(prompt string) string {
	fmt.Print(prompt)
	password, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password)
}

func Confirm(prompt string) bool {
	input := PromptString(fmt.Sprintf("%s (y/n): ", prompt))
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}
```

---

## Step 6: Create CLI Commands

**New File: `pkg/cli/root.go`**

```go
package cli

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envoy",
	Short: "Envoy CLI client",
	Long:  "CLI client for the Envoy project management server",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(projectsCmd)
}
```

**New File: `pkg/cli/auth.go`**

```go
package cli

import (
	"fmt"
	"ytsruh.com/envoy/pkg/config"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Run: func(cmd *cobra.Command, args []string) {
		name := PromptString("Enter your name: ")
		email := PromptString("Enter your email: ")
		password := PromptPassword("Enter your password: ")

		client := NewClient()
		resp, err := client.Register(name, email, password)
		if err != nil {
			fmt.Printf("Registration failed: %v\n", err)
			return
		}

		if err := config.SetAuthToken(resp.Token); err != nil {
			fmt.Printf("Warning: Could not save auth token: %v\n", err)
		}

		fmt.Printf("Registration successful!\n")
		fmt.Printf("Welcome, %s (%s)\n", resp.User.Name, resp.User.Email)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account",
	Run: func(cmd *cobra.Command, args []string) {
		email := PromptString("Enter your email: ")
		password := PromptPassword("Enter your password: ")

		client := NewClient()
		resp, err := client.Login(email, password)
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		if err := config.SetAuthToken(resp.Token); err != nil {
			fmt.Printf("Warning: Could not save auth token: %v\n", err)
		}

		fmt.Printf("Login successful!\n")
		fmt.Printf("Welcome back, %s (%s)\n", resp.User.Name, resp.User.Email)
	},
}
```

**New File: `pkg/cli/projects.go`**

```go
package cli

import (
	"fmt"
	"ytsruh.com/envoy/pkg/config"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var createProjectCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Not logged in. Please login first.")
			return
		}

		gitInfo, err := DetectGitRepo()
		if err != nil {
			fmt.Printf("Error detecting git repo: %v\n", err)
			return
		}

		var gitRepo string
		if gitInfo.IsGitRepo {
			gitRepo = gitInfo.GitRepo
			fmt.Printf("Detected git repository: %s\n", gitRepo)

			tracked, _ := IsGitRepoTracked(gitRepo)
			if tracked {
				fmt.Printf("Error: This git repository is already tracked as a project.\n")
				return
			}
		} else {
			fmt.Println("Warning: Current directory is not a git repository.")
			if !Confirm("Continue anyway?") {
				return
			}
		}

		name := PromptString("Enter project name: ")
		description := PromptString("Enter project description (optional, press Enter to skip): ")

		client := NewClient()
		resp, err := client.CreateProject(token, name, description, gitRepo)
		if err != nil {
			fmt.Printf("Failed to create project: %v\n", err)
			return
		}

		if gitRepo != "" {
			if err := TrackGitRepo(gitRepo, resp.ID, gitInfo.RemoteURL); err != nil {
				fmt.Printf("Warning: Could not track git repo: %v\n", err)
			}
		}

		fmt.Printf("Project created successfully!\n")
		fmt.Printf("ID: %d\n", resp.ID)
		fmt.Printf("Name: %s\n", resp.Name)
		if resp.Description != nil {
			fmt.Printf("Description: %s\n", *resp.Description)
		}
		if resp.GitRepo != "" {
			fmt.Printf("Git Repo: %s\n", resp.GitRepo)
		}
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Not logged in. Please login first.")
			return
		}

		client := NewClient()
		projects, err := client.ListProjects(token)
		if err != nil {
			fmt.Printf("Failed to list projects: %v\n", err)
			return
		}

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return
		}

		for _, p := range projects {
			fmt.Printf("ID: %d\n", p.ID)
			fmt.Printf("  Name: %s\n", p.Name)
			if p.Description != nil {
				fmt.Printf("  Description: %s\n", *p.Description)
			}
			if p.GitRepo != "" {
				fmt.Printf("  Git Repo: %s\n", p.GitRepo)
			}
			fmt.Printf("  Created: %s\n", p.CreatedAt)
			fmt.Println()
		}
	},
}

var getProjectCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Not logged in. Please login first.")
			return
		}

		var id int64
		_, err = fmt.Sscanf(args[0], "%d", &id)
		if err != nil {
			fmt.Println("Invalid project ID")
			return
		}

		client := NewClient()
		project, err := client.GetProject(token, id)
		if err != nil {
			fmt.Printf("Failed to get project: %v\n", err)
			return
		}

		fmt.Printf("ID: %d\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		if project.Description != nil {
			fmt.Printf("Description: %s\n", *project.Description)
		}
		if project.GitRepo != "" {
			fmt.Printf("Git Repo: %s\n", project.GitRepo)
		}
		fmt.Printf("Owner ID: %s\n", project.OwnerID)
		fmt.Printf("Created: %s\n", project.CreatedAt)
		fmt.Printf("Updated: %s\n", project.UpdatedAt)
	},
}

var updateProjectCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Not logged in. Please login first.")
			return
		}

		var id int64
		_, err = fmt.Sscanf(args[0], "%d", &id)
		if err != nil {
			fmt.Println("Invalid project ID")
			return
		}

		fmt.Println("Leave fields empty to keep current values")
		name := PromptString("Enter new name: ")
		description := PromptString("Enter new description (optional): ")

		if name == "" && description == "" {
			fmt.Println("No changes specified.")
			return
		}

		client := NewClient()
		project, err := client.UpdateProject(token, id, name, description)
		if err != nil {
			fmt.Printf("Failed to update project: %v\n", err)
			return
		}

		fmt.Printf("Project updated successfully!\n")
		fmt.Printf("ID: %d\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		if project.Description != nil {
			fmt.Printf("Description: %s\n", *project.Description)
		}
	},
}

var deleteProjectCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Not logged in. Please login first.")
			return
		}

		var id int64
		_, err = fmt.Sscanf(args[0], "%d", &id)
		if err != nil {
			fmt.Println("Invalid project ID")
			return
		}

		client := NewClient()

		project, err := client.GetProject(token, id)
		if err != nil {
			fmt.Printf("Failed to get project: %v\n", err)
			return
		}

		fmt.Printf("Are you sure you want to delete project '%s' (ID: %d)? ", project.Name, id)
		if !Confirm("") {
			fmt.Println("Delete cancelled.")
			return
		}

		if err := client.DeleteProject(token, id); err != nil {
			fmt.Printf("Failed to delete project: %v\n", err)
			return
		}

		if project.GitRepo != "" {
			if err := UntrackGitRepo(project.GitRepo); err != nil {
				fmt.Printf("Warning: Could not untrack git repo: %v\n", err)
			}
		}

		fmt.Printf("Project deleted successfully!\n")
	},
}

func init() {
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(getProjectCmd)
	projectsCmd.AddCommand(updateProjectCmd)
	projectsCmd.AddCommand(deleteProjectCmd)
}
```

---

## Step 7: Create CLI Entry Point

**New File: `cmd/cli.go`**

```go
//go:generate templ generate
package main

import (
	"ytsruh.com/envoy/pkg/cli"
)

func main() {
	cli.Execute()
}
```

---

## Step 8: Update Makefile

Update `Makefile` to support both server and CLI binaries:

```makefile
# Binary names
SERVER_BINARY := envoy-server
CLI_BINARY := envoy
BINARY := $(SERVER_BINARY)

# Go tooling
GORUN  := go run
GOTEST := go test
GOBUILD := go build

# Directories
PKGS := ./...
CMD_DIR := ./cmd

# Air (live reload dev tool)
AIR := air

.PHONY: run dev cli test test-ci build start help

# Run the server
run:
	$(GORUN) ./cmd/server.go

# Run the CLI
cli:
	$(GORUN) ./cmd/cli.go

# Development run with air (live reload)
dev:
	@command -v $(AIR) >/dev/null 2>&1 || (echo "air not found;"; exit 1)
	$(AIR)

# Run the full test suite
test:
	$(GOTEST) -v $(PKGS)

# CI-friendly test with race detector and coverage
test-ci:
	$(GOTEST) -race -coverprofile=coverage.out $(PKGS)

# Build server binary
build-server:
	$(GOBUILD) -o $(SERVER_BINARY) ./cmd/server.go

# Build CLI binary
build-cli:
	$(GOBUILD) -o $(CLI_BINARY) ./cmd/cli.go

# Build all binaries
build: build-server build-cli

# Start the compiled binary (production)
start: build
	@echo "Starting $(SERVER_BINARY)..."
	./$(SERVER_BINARY)

# Generate SQL Queries
generate:
	@command sqlc generate

# Help
help:
	@echo "Makefile commands:"
	@echo "  make run          - run the server"
	@echo "  make cli          - run the CLI"
	@echo "  make dev          - run with air (live reload)"
	@echo "  make test         - run all tests (go test ./...)"
	@echo "  make test-ci      - run all tests with race detector and coverage"
	@echo "  make build        - build all binaries"
	@echo "  make build-server - build server binary"
	@echo "  make build-cli    - build CLI binary"
	@echo "  make start        - build and run the server binary"
	@echo "  make generate     - generate SQL queries"
```

---

## Step 9: Write Tests

Create unit tests:

**`pkg/config/config_test.go`** - Test config loading/saving
**`pkg/cli/git_test.go`** - Test git detection and extraction
**`pkg/cli/client_test.go`** - Test HTTP client (with mocked server)

---

## Step 10: Update Documentation

**File: `AGENTS.md`**

Add new section:

```markdown
## CLI Development
- CLI binary built with Cobra framework
- Config stored in `~/.envoy/config.json`
- Git tracking stored in `~/.envoy/projects.json`
- Server URL controlled by `ENVOY_SERVER_URL` env var
- Run CLI with: `make cli` or `./envoy`

### CLI Commands
- `envoy register` - Register a new user
- `envoy login` - Login to your account
- `envoy projects create` - Create a new project
- `envoy projects list` - List all projects
- `envoy projects get <id>` - Get project details
- `envoy projects update <id>` - Update a project
- `envoy projects delete <id>` - Delete a project
```

**File: `README.md`**

Add CLI usage section with examples.

---

## Step 11: Integration Test

1. Build CLI: `make build-cli`
2. Test registration: `./envoy register`
3. Test login: `./envoy login`
4. Create project in git repo: `./envoy projects create`
5. List projects: `./envoy projects list`
6. Verify git repo tracking in `~/.envoy/projects.json`
7. Test duplicate git repo protection
8. Test project update/delete
