package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	generated "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type mockEnvironmentVariableQueries struct {
	envVars                []generated.EnvironmentVariable
	editorPermissionCounts map[string]int64
}

func (m *mockEnvironmentVariableQueries) CreateEnvironmentVariable(ctx context.Context, arg generated.CreateEnvironmentVariableParams) (generated.EnvironmentVariable, error) {
	envVar := generated.EnvironmentVariable{
		ID:            int64(len(m.envVars) + 1),
		EnvironmentID: arg.EnvironmentID,
		Key:           arg.Key,
		Value:         arg.Value,
		Description:   arg.Description,
		CreatedAt:     arg.CreatedAt,
		UpdatedAt:     arg.UpdatedAt,
	}
	m.envVars = append(m.envVars, envVar)
	return envVar, nil
}

func (m *mockEnvironmentVariableQueries) GetEnvironmentVariable(ctx context.Context, id int64) (generated.EnvironmentVariable, error) {
	for _, envVar := range m.envVars {
		if envVar.ID == id {
			return envVar, nil
		}
	}
	return generated.EnvironmentVariable{}, sql.ErrNoRows
}

func (m *mockEnvironmentVariableQueries) UpdateEnvironmentVariable(ctx context.Context, arg generated.UpdateEnvironmentVariableParams) (generated.EnvironmentVariable, error) {
	for i, envVar := range m.envVars {
		if envVar.ID == arg.ID {
			m.envVars[i].Key = arg.Key
			m.envVars[i].Value = arg.Value
			if arg.Description.Valid {
				m.envVars[i].Description = arg.Description
			}
			m.envVars[i].UpdatedAt = arg.UpdatedAt
			return m.envVars[i], nil
		}
	}
	return generated.EnvironmentVariable{}, sql.ErrNoRows
}

func (m *mockEnvironmentVariableQueries) DeleteEnvironmentVariable(ctx context.Context, id int64) error {
	for i, envVar := range m.envVars {
		if envVar.ID == id {
			m.envVars = append(m.envVars[:i], m.envVars[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockEnvironmentVariableQueries) ListEnvironmentVariablesByEnvironment(ctx context.Context, environmentID int64) ([]generated.EnvironmentVariable, error) {
	var result []generated.EnvironmentVariable
	for _, envVar := range m.envVars {
		if envVar.EnvironmentID == environmentID {
			result = append(result, envVar)
		}
	}
	return result, nil
}

func (m *mockEnvironmentVariableQueries) GetAccessibleEnvironmentVariable(ctx context.Context, arg generated.GetAccessibleEnvironmentVariableParams) (generated.EnvironmentVariable, error) {
	for _, envVar := range m.envVars {
		if envVar.ID == arg.ID {
			return envVar, nil
		}
	}
	return generated.EnvironmentVariable{}, sql.ErrNoRows
}

func (m *mockEnvironmentVariableQueries) CanUserModifyEnvironmentVariable(ctx context.Context, arg generated.CanUserModifyEnvironmentVariableParams) (int64, error) {
	return 1, nil
}

// Implement all other required methods with no-op implementations
func (m *mockEnvironmentVariableQueries) AddUserToProject(ctx context.Context, arg generated.AddUserToProjectParams) (generated.ProjectUser, error) {
	return generated.ProjectUser{}, nil
}

func (m *mockEnvironmentVariableQueries) CanUserModifyEnvironment(ctx context.Context, arg generated.CanUserModifyEnvironmentParams) (int64, error) {
	return 1, nil
}

func (m *mockEnvironmentVariableQueries) CanUserModifyProject(ctx context.Context, arg generated.CanUserModifyProjectParams) (int64, error) {
	return 1, nil
}

func (m *mockEnvironmentVariableQueries) CreateEnvironment(ctx context.Context, arg generated.CreateEnvironmentParams) (generated.Environment, error) {
	return generated.Environment{}, nil
}

func (m *mockEnvironmentVariableQueries) CreateProject(ctx context.Context, arg generated.CreateProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) CreateUser(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentVariableQueries) DeleteEnvironment(ctx context.Context, arg generated.DeleteEnvironmentParams) error {
	return nil
}

func (m *mockEnvironmentVariableQueries) DeleteProject(ctx context.Context, arg generated.DeleteProjectParams) error {
	return nil
}

func (m *mockEnvironmentVariableQueries) DeleteUser(ctx context.Context, arg generated.DeleteUserParams) error {
	return nil
}

func (m *mockEnvironmentVariableQueries) GetAccessibleEnvironment(ctx context.Context, arg generated.GetAccessibleEnvironmentParams) (generated.Environment, error) {
	return generated.Environment{}, nil
}

func (m *mockEnvironmentVariableQueries) GetAccessibleProject(ctx context.Context, arg generated.GetAccessibleProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) GetEnvironment(ctx context.Context, id int64) (generated.Environment, error) {
	return generated.Environment{}, nil
}

func (m *mockEnvironmentVariableQueries) GetProject(ctx context.Context, id int64) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) GetProjectMemberRole(ctx context.Context, arg generated.GetProjectMemberRoleParams) (string, error) {
	return "", nil
}

func (m *mockEnvironmentVariableQueries) GetProjectMembership(ctx context.Context, arg generated.GetProjectMembershipParams) (generated.ProjectUser, error) {
	return generated.ProjectUser{}, nil
}

func (m *mockEnvironmentVariableQueries) GetProjectUsers(ctx context.Context, projectID int64) ([]generated.ProjectUser, error) {
	return []generated.ProjectUser{}, nil
}

func (m *mockEnvironmentVariableQueries) GetUser(ctx context.Context, id string) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentVariableQueries) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentVariableQueries) GetUserProjects(ctx context.Context, arg generated.GetUserProjectsParams) ([]generated.Project, error) {
	return []generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *mockEnvironmentVariableQueries) IsProjectOwner(ctx context.Context, arg generated.IsProjectOwnerParams) (int64, error) {
	return 1, nil
}

func (m *mockEnvironmentVariableQueries) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]generated.Environment, error) {
	return []generated.Environment{}, nil
}

func (m *mockEnvironmentVariableQueries) ListProjectsByOwner(ctx context.Context, ownerID string) ([]generated.Project, error) {
	return []generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) ListUsers(ctx context.Context) ([]generated.User, error) {
	return []generated.User{}, nil
}

func (m *mockEnvironmentVariableQueries) RemoveUserFromProject(ctx context.Context, arg generated.RemoveUserFromProjectParams) error {
	return nil
}

func (m *mockEnvironmentVariableQueries) UpdateEnvironment(ctx context.Context, arg generated.UpdateEnvironmentParams) (generated.Environment, error) {
	return generated.Environment{}, nil
}

func (m *mockEnvironmentVariableQueries) UpdateProject(ctx context.Context, arg generated.UpdateProjectParams) (generated.Project, error) {
	return generated.Project{}, nil
}

func (m *mockEnvironmentVariableQueries) UpdateUser(ctx context.Context, arg generated.UpdateUserParams) (generated.User, error) {
	return generated.User{}, nil
}

func (m *mockEnvironmentVariableQueries) UpdateUserRole(ctx context.Context, arg generated.UpdateUserRoleParams) error {
	return nil
}

func createEnvVarTestRequest(method, path, body string, userID string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	claims := &utils.JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
	}
	ctx := context.WithValue(req.Context(), "user", claims)
	return req.WithContext(ctx)
}

func TestEnvironmentVariableHandler_CreateEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid environment variable creation",
			requestBody:    `{"key": "API_KEY", "value": "secret123"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "uppercase conversion",
			requestBody:    `{"key": "api_key", "value": "secret123"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing key field",
			requestBody:    `{"value": "secret123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key is required",
		},
		{
			name:           "missing value field",
			requestBody:    `{"key": "API_KEY"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Value is required",
		},
		{
			name:           "value too long",
			requestBody:    `{"key": "API_KEY", "value": "` + strings.Repeat("a", 256) + `"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "must be at most 255 characters",
		},
		{
			name:           "description too long",
			requestBody:    `{"key": "API_KEY", "value": "secret", "description": "` + strings.Repeat("a", 501) + `"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at most 500 characters",
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnvironmentVariableQueries{}
			ac := &mockAccessControlService{}
			handler := NewEnvironmentVariableHandler(mock, ac)

			req := createEnvVarTestRequest(http.MethodPost, "/projects/1/environments/1/variables", tt.requestBody, "user1")
			w := httptest.NewRecorder()

			err := handler.CreateEnvironmentVariable(w, req)

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

func TestEnvironmentVariableHandler_GetEnvironmentVariable(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentVariableQueries{
		envVars: []generated.EnvironmentVariable{
			{
				ID:            1,
				EnvironmentID: 1,
				Key:           "API_KEY",
				Value:         "secret123",
				CreatedAt:     sql.NullTime{Time: now, Valid: true},
				UpdatedAt:     sql.NullTime{Time: now, Valid: true},
			},
		},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	req := createEnvVarTestRequest(http.MethodGet, "/projects/1/environments/1/variables/1", "", "user1")
	w := httptest.NewRecorder()

	err := handler.GetEnvironmentVariable(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if response["id"].(float64) != 1 || response["key"].(string) != "API_KEY" {
		t.Errorf("expected ID 1 and key 'API_KEY', got ID %v and key '%s'", response["id"], response["key"])
	}
}

func TestEnvironmentVariableHandler_UpdateEnvironmentVariable(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentVariableQueries{
		envVars: []generated.EnvironmentVariable{
			{
				ID:            1,
				EnvironmentID: 1,
				Key:           "API_KEY",
				Value:         "secret123",
				CreatedAt:     sql.NullTime{Time: now, Valid: true},
				UpdatedAt:     sql.NullTime{Time: now, Valid: true},
			},
		},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	requestBody := `{"key": "NEW_KEY", "value": "newvalue"}`
	req := createEnvVarTestRequest(http.MethodPut, "/projects/1/environments/1/variables/1", requestBody, "user1")
	w := httptest.NewRecorder()

	err := handler.UpdateEnvironmentVariable(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if response["key"].(string) != "NEW_KEY" || response["value"].(string) != "newvalue" {
		t.Errorf("expected key 'NEW_KEY' and value 'newvalue', got key '%s' and value '%s'", response["key"], response["value"])
	}
}

func TestEnvironmentVariableHandler_DeleteEnvironmentVariable(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentVariableQueries{
		envVars: []generated.EnvironmentVariable{
			{
				ID:            1,
				EnvironmentID: 1,
				Key:           "API_KEY",
				Value:         "secret123",
				CreatedAt:     sql.NullTime{Time: now, Valid: true},
				UpdatedAt:     sql.NullTime{Time: now, Valid: true},
			},
		},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	req := createEnvVarTestRequest(http.MethodDelete, "/projects/1/environments/1/variables/1", "", "user1")
	w := httptest.NewRecorder()

	err := handler.DeleteEnvironmentVariable(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if len(mock.envVars) != 0 {
		t.Errorf("expected 0 environment variables after deletion, got %d", len(mock.envVars))
	}
}

func TestEnvironmentVariableHandler_ListEnvironmentVariables(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentVariableQueries{
		envVars: []generated.EnvironmentVariable{
			{ID: 1, EnvironmentID: 1, Key: "API_KEY", Value: "secret123", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
			{ID: 2, EnvironmentID: 1, Key: "DB_URL", Value: "localhost", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
			{ID: 3, EnvironmentID: 2, Key: "OTHER_KEY", Value: "value", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now, Valid: true}},
		},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	req := createEnvVarTestRequest(http.MethodGet, "/projects/1/environments/1/variables", "", "user1")
	w := httptest.NewRecorder()

	err := handler.ListEnvironmentVariables(w, req)
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
		t.Errorf("expected 2 environment variables for environment 1, got %d", len(response))
	}
}

func TestNewEnvironmentVariableHandler(t *testing.T) {
	mock := &mockEnvironmentVariableQueries{}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	if handler == nil {
		t.Fatal("NewEnvironmentVariableHandler() returned nil")
	}
}

func TestGetEnvironmentVariableIDFromPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedID    int64
		expectedError bool
	}{
		{
			name:          "valid environment variable ID",
			path:          "/projects/1/environments/1/variables/123",
			expectedID:    123,
			expectedError: false,
		},
		{
			name:          "missing environment variable ID",
			path:          "/projects/1/environments/1/variables",
			expectedID:    0,
			expectedError: true,
		},
		{
			name:          "invalid environment variable ID",
			path:          "/projects/1/environments/1/variables/abc",
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			id, err := getEnvironmentVariableIDFromPath(req)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if id != tt.expectedID {
					t.Errorf("expected ID %d, got %d", tt.expectedID, id)
				}
			}
		})
	}
}

func TestUppercaseKeyConversion(t *testing.T) {
	mock := &mockEnvironmentVariableQueries{}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	requestBody := `{"key": "api_key", "value": "secret123"}`
	req := createEnvVarTestRequest(http.MethodPost, "/projects/1/environments/1/variables", requestBody, "user1")
	w := httptest.NewRecorder()

	err := handler.CreateEnvironmentVariable(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	} else if response["key"].(string) != "API_KEY" {
		t.Errorf("expected key 'API_KEY' (uppercase), got '%s'", response["key"])
	}
}

func TestDuplicateKeysAllowed(t *testing.T) {
	now := time.Now()
	mock := &mockEnvironmentVariableQueries{
		envVars: []generated.EnvironmentVariable{
			{
				ID:            1,
				EnvironmentID: 1,
				Key:           "API_KEY",
				Value:         "secret123",
				CreatedAt:     sql.NullTime{Time: now, Valid: true},
				UpdatedAt:     sql.NullTime{Time: now, Valid: true},
			},
		},
	}
	ac := &mockAccessControlService{}
	handler := NewEnvironmentVariableHandler(mock, ac)

	requestBody := `{"key": "API_KEY", "value": "another_secret"}`
	req := createEnvVarTestRequest(http.MethodPost, "/projects/1/environments/1/variables", requestBody, "user1")
	w := httptest.NewRecorder()

	err := handler.CreateEnvironmentVariable(w, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d (duplicate keys should be allowed), got %d", http.StatusCreated, w.Code)
	}

	if len(mock.envVars) != 2 {
		t.Errorf("expected 2 environment variables (duplicates allowed), got %d", len(mock.envVars))
	}
}
