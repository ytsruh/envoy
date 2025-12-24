package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type MockQueries struct {
	projects            []database.Project
	users               map[string]database.User
	canUserModifyResult int64
}

func (m *MockQueries) CreateProject(ctx context.Context, arg database.CreateProjectParams) (database.Project, error) {
	project := database.Project{
		ID:          int64(len(m.projects) + 1),
		Name:        arg.Name,
		Description: arg.Description,
		OwnerID:     arg.OwnerID,
		CreatedAt:   arg.CreatedAt,
		UpdatedAt:   arg.UpdatedAt,
		DeletedAt:   sql.NullTime{Valid: false},
	}
	m.projects = append(m.projects, project)
	return project, nil
}

func (m *MockQueries) GetProject(ctx context.Context, id int64) (database.Project, error) {
	for _, p := range m.projects {
		if p.ID == id && !p.DeletedAt.Valid {
			return p, nil
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockQueries) ListProjectsByOwner(ctx context.Context, ownerID string) ([]database.Project, error) {
	var result []database.Project
	for _, p := range m.projects {
		if p.OwnerID == ownerID && !p.DeletedAt.Valid {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockQueries) UpdateProject(ctx context.Context, arg database.UpdateProjectParams) (database.Project, error) {
	for i, p := range m.projects {
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID && !p.DeletedAt.Valid {
			m.projects[i].Name = arg.Name
			m.projects[i].Description = arg.Description
			m.projects[i].UpdatedAt = arg.UpdatedAt
			return m.projects[i], nil
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockQueries) DeleteProject(ctx context.Context, arg database.DeleteProjectParams) error {
	for i, p := range m.projects {
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID && !p.DeletedAt.Valid {
			m.projects[i].DeletedAt = arg.DeletedAt
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *MockQueries) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	user := database.User{
		ID:        arg.ID,
		Name:      arg.Name,
		Email:     arg.Email,
		Password:  arg.Password,
		CreatedAt: arg.CreatedAt,
		UpdatedAt: arg.UpdatedAt,
		DeletedAt: arg.DeletedAt,
	}
	m.users[arg.ID] = user
	return user, nil
}

func (m *MockQueries) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m *MockQueries) GetUser(ctx context.Context, id string) (database.User, error) {
	user, ok := m.users[id]
	if !ok {
		return database.User{}, sql.ErrNoRows
	}
	return user, nil
}

func (m *MockQueries) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return database.User{}, sql.ErrNoRows
}

func (m *MockQueries) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *MockQueries) ListUsers(ctx context.Context) ([]database.User, error) {
	var users []database.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockQueries) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

// Project sharing methods
func (m *MockQueries) AddUserToProject(ctx context.Context, arg database.AddUserToProjectParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, nil
}

func (m *MockQueries) RemoveUserFromProject(ctx context.Context, arg database.RemoveUserFromProjectParams) error {
	return nil
}

func (m *MockQueries) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	return nil
}

func (m *MockQueries) GetProjectUsers(ctx context.Context, projectID int64) ([]database.ProjectUser, error) {
	return []database.ProjectUser{}, nil
}

func (m *MockQueries) GetUserProjects(ctx context.Context, arg database.GetUserProjectsParams) ([]database.Project, error) {
	var result []database.Project
	for _, p := range m.projects {
		if !p.DeletedAt.Valid {
			// Owner access
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				result = append(result, p)
			}
		}
	}
	return result, nil
}

func (m *MockQueries) GetProjectMembership(ctx context.Context, arg database.GetProjectMembershipParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, sql.ErrNoRows
}

func (m *MockQueries) IsProjectOwner(ctx context.Context, arg database.IsProjectOwnerParams) (int64, error) {
	return 0, nil
}

func (m *MockQueries) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			// Check if owner or shared user
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return p, nil
			}
			// For tests, we don't have project users in this mock, so just check ownership
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockQueries) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	// If we've set a custom result for testing, use it
	if m.canUserModifyResult != 0 {
		return m.canUserModifyResult, nil
	}

	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			// Owner can modify
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return 1, nil
			}
			// For tests, we don't have project users in this mock, so just check ownership
			break
		}
	}
	return 0, nil
}

func (m *MockQueries) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	return "", sql.ErrNoRows
}

// Environment methods
func (m *MockQueries) CreateEnvironment(ctx context.Context, arg database.CreateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQueries) GetEnvironment(ctx context.Context, id int64) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQueries) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]database.Environment, error) {
	return []database.Environment{}, nil
}

func (m *MockQueries) UpdateEnvironment(ctx context.Context, arg database.UpdateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQueries) DeleteEnvironment(ctx context.Context, arg database.DeleteEnvironmentParams) error {
	return nil
}

func (m *MockQueries) GetAccessibleEnvironment(ctx context.Context, arg database.GetAccessibleEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQueries) CanUserModifyEnvironment(ctx context.Context, arg database.CanUserModifyEnvironmentParams) (int64, error) {
	return 0, nil
}

// Environment variable methods
func (m *MockQueries) CreateEnvironmentVariable(ctx context.Context, arg database.CreateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *MockQueries) DeleteEnvironmentVariable(ctx context.Context, id int64) error {
	return nil
}

func (m *MockQueries) GetAccessibleEnvironmentVariable(ctx context.Context, arg database.GetAccessibleEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *MockQueries) GetEnvironmentVariable(ctx context.Context, id int64) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *MockQueries) ListEnvironmentVariablesByEnvironment(ctx context.Context, environmentID int64) ([]database.EnvironmentVariable, error) {
	return []database.EnvironmentVariable{}, nil
}

func (m *MockQueries) UpdateEnvironmentVariable(ctx context.Context, arg database.UpdateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *MockQueries) CanUserModifyEnvironmentVariable(ctx context.Context, arg database.CanUserModifyEnvironmentVariableParams) (int64, error) {
	return 0, nil
}

func createTestUser() *utils.JWTClaims {
	return &utils.JWTClaims{
		UserID: uuid.New().String(),
		Email:  "test@example.com",
		Iat:    time.Now().Unix(),
		Exp:    time.Now().Add(time.Hour).Unix(),
	}
}

func TestCreateProject(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	handler := NewProjectHandler(mock)
	claims := createTestUser()

	reqBody := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err := handler.CreateProject(w, req)
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response ProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", response.Name)
	}

	if response.OwnerID != claims.UserID {
		t.Errorf("Expected owner ID '%s', got '%s'", claims.UserID, response.OwnerID)
	}
}

func TestGetProject(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	handler := NewProjectHandler(mock)
	claims := createTestUser()

	projectID := int64(1)
	mock.projects = append(mock.projects, database.Project{
		ID:        projectID,
		Name:      "Test Project",
		OwnerID:   claims.UserID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{Valid: false},
	})

	req := httptest.NewRequest("GET", "/projects/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err := handler.GetProject(w, req)
	if err != nil {
		t.Fatalf("GetProject returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response ProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != projectID {
		t.Errorf("Expected ID %d, got %d", projectID, response.ID)
	}
}

func TestListProjects(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	handler := NewProjectHandler(mock)
	claims := createTestUser()

	mock.projects = append(mock.projects, database.Project{
		ID:        1,
		Name:      "Project 1",
		OwnerID:   claims.UserID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{Valid: false},
	})

	mock.projects = append(mock.projects, database.Project{
		ID:        2,
		Name:      "Project 2",
		OwnerID:   claims.UserID,
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{Valid: false},
	})

	req := httptest.NewRequest("GET", "/projects", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err := handler.ListProjects(w, req)
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []ProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(response))
	}
}

func TestUpdateProject(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	handler := NewProjectHandler(mock)

	ownerID := "owner123"
	editorID := "editor123"

	// Create owner and editor users
	mock.users[ownerID] = database.User{ID: ownerID, Name: "Owner", Email: "owner@example.com"}
	mock.users[editorID] = database.User{ID: editorID, Name: "Editor", Email: "editor@example.com"}

	// Create project
	project := database.Project{
		ID:          1,
		Name:        "Test Project",
		Description: sql.NullString{String: "Original description", Valid: true},
		OwnerID:     ownerID,
		CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:   time.Now(),
		DeletedAt:   sql.NullTime{Valid: false},
	}
	mock.projects = append(mock.projects, project)

	tests := []struct {
		name           string
		userID         string
		isOwner        bool
		expectedStatus int
	}{
		{
			name:           "owner can update project",
			userID:         ownerID,
			isOwner:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "editor can update project",
			userID:         editorID,
			isOwner:        false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-member cannot update project",
			userID:         "stranger123",
			isOwner:        false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup editor access for editor test
			if !tt.isOwner && tt.userID == editorID {
				// Update mock to allow editor to modify
				mock.canUserModifyResult = 1
			} else if tt.isOwner {
				mock.canUserModifyResult = 1
			} else {
				mock.canUserModifyResult = 0
			}

			reqBody := UpdateProjectRequest{
				Name:        "Updated Project Name",
				Description: "Updated description",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("PUT", "/projects/1", bytes.NewReader(body))
			claims := &utils.JWTClaims{
				UserID: tt.userID,
				Email:  "test@example.com",
				Iat:    time.Now().Unix(),
				Exp:    time.Now().Add(time.Hour).Unix(),
			}
			req = req.WithContext(context.WithValue(req.Context(), "user", claims))
			w := httptest.NewRecorder()

			err := handler.UpdateProject(w, req)
			if err != nil {
				t.Fatalf("UpdateProject returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestCreateProjectValidation(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	handler := NewProjectHandler(mock)
	claims := createTestUser()

	tests := []struct {
		name           string
		requestBody    CreateProjectRequest
		expectedError  string
		expectedStatus int
	}{
		{
			name: "empty name",
			requestBody: CreateProjectRequest{
				Name:        "",
				Description: "A test project",
			},
			expectedError:  "Name is required",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			requestBody: CreateProjectRequest{
				Name:        string(make([]byte, 101)), // 101 characters
				Description: "A test project",
			},
			expectedError:  "Name must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid characters in name",
			requestBody: CreateProjectRequest{
				Name:        "Test@Project!",
				Description: "A test project",
			},
			expectedError:  "Name must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "description too long",
			requestBody: CreateProjectRequest{
				Name:        "Valid Project Name",
				Description: string(make([]byte, 501)), // 501 characters
			},
			expectedError:  "Description must be at most 500 characters long",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), "user", claims))
			w := httptest.NewRecorder()

			err := handler.CreateProject(w, req)
			if err != nil {
				t.Fatalf("CreateProject returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response ErrorResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, response.Error)
			}
		})
	}
}
