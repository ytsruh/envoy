package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	generated "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type mockEnvironmentQueries struct {
	environments           []generated.Environment
	project                generated.Project
	editorPermissionCounts map[string]int64
}

func (m *mockEnvironmentQueries) CreateEnvironment(ctx context.Context, arg generated.CreateEnvironmentParams) (generated.Environment, error) {
	env := generated.Environment{
		ID:          int64(len(m.environments) + 1),
		ProjectID:   arg.ProjectID,
		Name:        arg.Name,
		Description: arg.Description,
		CreatedAt:   arg.CreatedAt,
		UpdatedAt:   arg.UpdatedAt,
	}
	m.environments = append(m.environments, env)
	return env, nil
}

func (m *mockEnvironmentQueries) GetEnvironment(ctx context.Context, id int64) (generated.Environment, error) {
	for _, env := range m.environments {
		if env.ID == id {
			return env, nil
		}
	}
	return generated.Environment{}, sql.ErrNoRows
}

func (m *mockEnvironmentQueries) UpdateEnvironment(ctx context.Context, arg generated.UpdateEnvironmentParams) (generated.Environment, error) {
	for i, env := range m.environments {
		if env.ID == arg.ID {
			m.environments[i].Name = arg.Name
			if arg.Description.Valid {
				m.environments[i].Description = arg.Description
			}
			m.environments[i].UpdatedAt = arg.UpdatedAt
			return m.environments[i], nil
		}
	}
	return generated.Environment{}, sql.ErrNoRows
}

func (m *mockEnvironmentQueries) DeleteEnvironment(ctx context.Context, arg generated.DeleteEnvironmentParams) error {
	for i, env := range m.environments {
		if env.ID == arg.ID {
			m.environments = append(m.environments[:i], m.environments[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockEnvironmentQueries) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]generated.Environment, error) {
	var result []generated.Environment
	for _, env := range m.environments {
		if env.ProjectID == projectID {
			result = append(result, env)
		}
	}
	return result, nil
}

func (m *mockEnvironmentQueries) GetProject(ctx context.Context, id int64) (generated.Project, error) {
	if m.project.ID == id {
		return m.project, nil
	}
	return generated.Project{}, sql.ErrNoRows
}

// Implement all other required methods with no-op implementations
func (m *mockEnvironmentQueries) AddUserToProject(ctx context.Context, arg generated.AddUserToProjectParams) (generated.ProjectUser, error) {
	return generated.ProjectUser{}, nil
}

func (m *mockEnvironmentQueries) CanUserModifyEnvironment(ctx context.Context, arg generated.CanUserModifyEnvironmentParams) (int64, error) {
	return 1, nil // Allow modification
}

func (m *mockEnvironmentQueries) CanUserModifyProject(ctx context.Context, arg generated.CanUserModifyProjectParams) (int64, error) {
	return 1, nil // Allow modification
}

func (m *mockEnvironmentQueries) CreateProject(ctx context.Context, arg generated.CreateProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentQueries) CreateUser(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentQueries) DeleteProject(ctx context.Context, arg generated.DeleteProjectParams) error {
	return nil
}

func (m *mockEnvironmentQueries) DeleteUser(ctx context.Context, arg generated.DeleteUserParams) error {
	return nil
}

func (m *mockEnvironmentQueries) GetAccessibleEnvironment(ctx context.Context, arg generated.GetAccessibleEnvironmentParams) (generated.Environment, error) {
	for _, env := range m.environments {
		if env.ID == arg.ID {
			return env, nil
		}
	}
	return generated.Environment{}, sql.ErrNoRows
}

func (m *mockEnvironmentQueries) GetAccessibleProject(ctx context.Context, arg generated.GetAccessibleProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentQueries) GetProjectMemberRole(ctx context.Context, arg generated.GetProjectMemberRoleParams) (string, error) {
	return "", nil
}

func (m *mockEnvironmentQueries) GetProjectMembership(ctx context.Context, arg generated.GetProjectMembershipParams) (generated.ProjectUser, error) {
	return generated.ProjectUser{}, nil
}

func (m *mockEnvironmentQueries) GetProjectUsers(ctx context.Context, projectID int64) ([]generated.ProjectUser, error) {
	return []generated.ProjectUser{}, nil
}

func (m *mockEnvironmentQueries) GetUser(ctx context.Context, id string) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentQueries) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentQueries) GetUserProjects(ctx context.Context, arg generated.GetUserProjectsParams) ([]generated.Project, error) {
	return []generated.Project{}, nil
}

func (m *mockEnvironmentQueries) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *mockEnvironmentQueries) IsProjectOwner(ctx context.Context, arg generated.IsProjectOwnerParams) (int64, error) {
	return 1, nil // Assume ownership
}

func (m *mockEnvironmentQueries) ListProjectsByOwner(ctx context.Context, ownerID string) ([]generated.Project, error) {
	return []generated.Project{}, nil
}

func (m *mockEnvironmentQueries) ListUsers(ctx context.Context) ([]generated.User, error) {
	return []generated.User{}, nil
}

func (m *mockEnvironmentQueries) RemoveUserFromProject(ctx context.Context, arg generated.RemoveUserFromProjectParams) error {
	return nil
}

func (m *mockEnvironmentQueries) UpdateProject(ctx context.Context, arg generated.UpdateProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentQueries) UpdateUser(ctx context.Context, arg generated.UpdateUserParams) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentQueries) UpdateUserRole(ctx context.Context, arg generated.UpdateUserRoleParams) error {
	return nil
}

type mockAccessControlService struct {
	accessibleProjects map[string]generated.Project
}

func (m *mockAccessControlService) RequireOwner(ctx context.Context, projectID int64, userID string) error {
	return nil // Allow all access for testing
}

func (m *mockAccessControlService) RequireEditor(ctx context.Context, projectID int64, userID string) error {
	return nil // Allow all modification for testing
}

func (m *mockAccessControlService) RequireViewer(ctx context.Context, projectID int64, userID string) error {
	return nil // Allow all viewing for testing
}

func (m *mockAccessControlService) GetRole(ctx context.Context, projectID int64, userID string) (string, error) {
	return "owner", nil // Assume owner role for testing
}

// Helper function to extract environment ID from URL path
func getEnvironmentIDFromPathForTest(r *http.Request) (int64, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		return 0, fmt.Errorf("environment ID is required")
	}
	return strconv.ParseInt(pathParts[3], 10, 64)
}

func createTestRequest(method, path, body string, userID string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add user claims to context
	claims := &utils.JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
	}
	ctx := context.WithValue(req.Context(), "user", claims)
	return req.WithContext(ctx)
}

func TestEnvironmentsHandler_CreateEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid environment creation",
			requestBody:    `{"name": "Production"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid environment name - too short",
			requestBody:    `{"name": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name is required",
		},
		{
			name:           "invalid environment name - too long",
			requestBody:    `{"name": "` + strings.Repeat("a", 101) + `"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores",
		},
		{
			name:           "invalid environment name - invalid characters",
			requestBody:    `{"name": "Test@Environment"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores",
		},
		{
			name:           "missing name field",
			requestBody:    `{"description": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Name is required",
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnvironmentQueries{
				project: generated.Project{ID: 1, Name: "Test Project"},
			}
			ac := &mockAccessControlService{}
			handler := NewEnvironmentHandler(mock, ac)

			req := createTestRequest(http.MethodPost, "/projects/1/environments", tt.requestBody, "user1")
			w := httptest.NewRecorder()

			err := handler.CreateEnvironment(w, req)

			// Echo-style handlers return error, but we check the response
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				} else if errorMsg, ok := response["error"].(string); !ok || !contains(errorMsg, tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, errorMsg)
				}
			}
		})
	}
}

func TestEnvironmentsHandler_GetEnvironment(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentQueries{
		environments: []generated.Environment{
			{
				ID:        1,
				ProjectID: 1,
				Name:      "Production",
				CreatedAt: sql.NullTime{Time: now, Valid: true},
				UpdatedAt: sql.NullTime{Time: now, Valid: true},
			},
		},
		project: generated.Project{ID: 1, Name: "Test Project"},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentHandler(mock, ac)

	req := createTestRequest(http.MethodGet, "/projects/1/environments/1", "", "user1")
	w := httptest.NewRecorder()

	err := handler.GetEnvironment(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if response["id"].(float64) != 1 || response["name"].(string) != "Production" {
		t.Errorf("expected environment ID 1 and name 'Production', got ID %v and name '%s'", response["id"], response["name"])
	}
}

func TestEnvironmentsHandler_UpdateEnvironment(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentQueries{
		environments: []generated.Environment{
			{
				ID:        1,
				ProjectID: 1,
				Name:      "Production",
				CreatedAt: sql.NullTime{Time: now, Valid: true},
				UpdatedAt: sql.NullTime{Time: now, Valid: true},
			},
		},
		project: generated.Project{ID: 1, Name: "Test Project"},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentHandler(mock, ac)

	requestBody := `{"name": "Staging"}`
	req := createTestRequest(http.MethodPut, "/projects/1/environments/1", requestBody, "user1")
	w := httptest.NewRecorder()

	err := handler.UpdateEnvironment(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if response["name"].(string) != "Staging" {
		t.Errorf("expected name 'Staging', got '%s'", response["name"])
	}
}

func TestEnvironmentsHandler_DeleteEnvironment(t *testing.T) {
	mock := &mockEnvironmentQueries{
		environments: []generated.Environment{
			{
				ID:        1,
				ProjectID: 1,
				Name:      "Production",
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			},
		},
		project: generated.Project{ID: 1, Name: "Test Project"},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentHandler(mock, ac)

	req := createTestRequest(http.MethodDelete, "/projects/1/environments/1", "", "user1")
	w := httptest.NewRecorder()

	err := handler.DeleteEnvironment(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if len(mock.environments) != 0 {
		t.Errorf("expected 0 environments after deletion, got %d", len(mock.environments))
	}
}

func TestEnvironmentsHandler_ListEnvironments(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentQueries{
		environments: []generated.Environment{
			{ID: 1, ProjectID: 1, Name: "Production", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
			{ID: 2, ProjectID: 1, Name: "Staging", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
			{ID: 3, ProjectID: 2, Name: "Development", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
		},
		project: generated.Project{ID: 1, Name: "Test Project"},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentHandler(mock, ac)

	req := createTestRequest(http.MethodGet, "/projects/1/environments", "", "user1")
	w := httptest.NewRecorder()

	err := handler.ListEnvironments(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if len(response) != 2 {
		t.Errorf("expected 2 environments for project 1, got %d", len(response))
	}
}

func TestNewEnvironmentHandler(t *testing.T) {
	mock := &mockEnvironmentQueries{}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentHandler(mock, ac)

	if handler == nil {
		t.Fatal("NewEnvironmentHandler() returned nil")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
