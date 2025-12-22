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
	projects []database.Project
	users    map[string]database.User
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
