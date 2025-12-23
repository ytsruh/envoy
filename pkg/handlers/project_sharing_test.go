package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

// MockSharingQueries implements database.Querier for testing sharing functionality
type MockSharingQueries struct {
	projects      []database.Project
	projectUsers  []database.ProjectUser
	users         map[string]database.User
	nextProjectID int64
}

func NewMockSharingQueries() *MockSharingQueries {
	return &MockSharingQueries{
		projects:      []database.Project{},
		projectUsers:  []database.ProjectUser{},
		users:         make(map[string]database.User),
		nextProjectID: 1,
	}
}

func (m *MockSharingQueries) CreateProject(ctx context.Context, arg database.CreateProjectParams) (database.Project, error) {
	project := database.Project{
		ID:          m.nextProjectID,
		Name:        arg.Name,
		Description: arg.Description,
		OwnerID:     arg.OwnerID,
		CreatedAt:   arg.CreatedAt,
		UpdatedAt:   arg.UpdatedAt,
		DeletedAt:   sql.NullTime{Valid: false},
	}
	m.projects = append(m.projects, project)
	m.nextProjectID++
	return project, nil
}

func (m *MockSharingQueries) GetProject(ctx context.Context, id int64) (database.Project, error) {
	for _, p := range m.projects {
		if p.ID == id && !p.DeletedAt.Valid {
			return p, nil
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockSharingQueries) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			// Check if owner or shared user
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return p, nil
			}
			// Check project users
			for _, pu := range m.projectUsers {
				if pu.ProjectID == arg.ID && pu.UserID == arg.UserID {
					return p, nil
				}
			}
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockSharingQueries) ListProjectsByOwner(ctx context.Context, ownerID string) ([]database.Project, error) {
	var result []database.Project
	for _, p := range m.projects {
		if p.OwnerID == ownerID && !p.DeletedAt.Valid {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockSharingQueries) GetUserProjects(ctx context.Context, arg database.GetUserProjectsParams) ([]database.Project, error) {
	var result []database.Project
	for _, p := range m.projects {
		if !p.DeletedAt.Valid {
			// Owner access
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				result = append(result, p)
				continue
			}
			// Shared access
			for _, pu := range m.projectUsers {
				if pu.ProjectID == p.ID && pu.UserID == arg.UserID {
					result = append(result, p)
					break
				}
			}
		}
	}
	return result, nil
}

func (m *MockSharingQueries) UpdateProject(ctx context.Context, arg database.UpdateProjectParams) (database.Project, error) {
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

func (m *MockSharingQueries) DeleteProject(ctx context.Context, arg database.DeleteProjectParams) error {
	for i, p := range m.projects {
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID && !p.DeletedAt.Valid {
			m.projects[i].DeletedAt = arg.DeletedAt
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *MockSharingQueries) AddUserToProject(ctx context.Context, arg database.AddUserToProjectParams) (database.ProjectUser, error) {
	// Check if already exists
	for _, pu := range m.projectUsers {
		if pu.ProjectID == arg.ProjectID && pu.UserID == arg.UserID {
			return database.ProjectUser{}, errors.New("user already exists")
		}
	}

	projectUser := database.ProjectUser{
		ID:        int64(len(m.projectUsers) + 1),
		ProjectID: arg.ProjectID,
		UserID:    arg.UserID,
		Role:      arg.Role,
		CreatedAt: arg.CreatedAt,
		UpdatedAt: arg.UpdatedAt,
	}
	m.projectUsers = append(m.projectUsers, projectUser)
	return projectUser, nil
}

func (m *MockSharingQueries) RemoveUserFromProject(ctx context.Context, arg database.RemoveUserFromProjectParams) error {
	for i, pu := range m.projectUsers {
		if pu.ProjectID == arg.ProjectID && pu.UserID == arg.UserID {
			m.projectUsers = append(m.projectUsers[:i], m.projectUsers[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *MockSharingQueries) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	for i, pu := range m.projectUsers {
		if pu.ProjectID == arg.ProjectID && pu.UserID == arg.UserID {
			m.projectUsers[i].Role = arg.Role
			m.projectUsers[i].UpdatedAt = arg.UpdatedAt
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *MockSharingQueries) GetProjectUsers(ctx context.Context, projectID int64) ([]database.ProjectUser, error) {
	var result []database.ProjectUser
	for _, pu := range m.projectUsers {
		if pu.ProjectID == projectID {
			result = append(result, pu)
		}
	}
	return result, nil
}

func (m *MockSharingQueries) GetProjectMembership(ctx context.Context, arg database.GetProjectMembershipParams) (database.ProjectUser, error) {
	for _, pu := range m.projectUsers {
		if pu.ProjectID == arg.ProjectID && pu.UserID == arg.UserID {
			return pu, nil
		}
	}
	return database.ProjectUser{}, sql.ErrNoRows
}

func (m *MockSharingQueries) IsProjectOwner(ctx context.Context, arg database.IsProjectOwnerParams) (int64, error) {
	for _, p := range m.projects {
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID && !p.DeletedAt.Valid {
			return 1, nil
		}
	}
	return 0, nil
}

func (m *MockSharingQueries) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			// Owner can modify
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return 1, nil
			}
			// Check if editor
			for _, pu := range m.projectUsers {
				if pu.ProjectID == arg.ID && pu.UserID == arg.UserID && pu.Role == "editor" {
					return 1, nil
				}
			}
			break
		}
	}
	return 0, nil
}

func (m *MockSharingQueries) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	for _, pu := range m.projectUsers {
		if pu.ProjectID == arg.ProjectID && pu.UserID == arg.UserID {
			return pu.Role, nil
		}
	}
	return "", sql.ErrNoRows
}

// User methods
func (m *MockSharingQueries) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	user := database.User{
		ID:        arg.ID,
		Name:      arg.Name,
		Email:     arg.Email,
		Password:  arg.Password,
		CreatedAt: arg.CreatedAt,
		UpdatedAt: arg.UpdatedAt,
		DeletedAt: sql.NullTime{Valid: false},
	}
	m.users[arg.ID] = user
	return user, nil
}

func (m *MockSharingQueries) GetUser(ctx context.Context, id string) (database.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return database.User{}, sql.ErrNoRows
}

func (m *MockSharingQueries) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return database.User{}, sql.ErrNoRows
}

func (m *MockSharingQueries) ListUsers(ctx context.Context) ([]database.User, error) {
	var users []database.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockSharingQueries) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m *MockSharingQueries) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m *MockSharingQueries) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

// Environment methods
func (m *MockSharingQueries) CreateEnvironment(ctx context.Context, arg database.CreateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockSharingQueries) GetEnvironment(ctx context.Context, id int64) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockSharingQueries) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]database.Environment, error) {
	return []database.Environment{}, nil
}

func (m *MockSharingQueries) UpdateEnvironment(ctx context.Context, arg database.UpdateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockSharingQueries) DeleteEnvironment(ctx context.Context, arg database.DeleteEnvironmentParams) error {
	return nil
}

func (m *MockSharingQueries) GetAccessibleEnvironment(ctx context.Context, arg database.GetAccessibleEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockSharingQueries) CanUserModifyEnvironment(ctx context.Context, arg database.CanUserModifyEnvironmentParams) (int64, error) {
	return 0, nil
}

// Helper function for tests
func createSharingTestUser(id string) *utils.JWTClaims {
	return &utils.JWTClaims{
		UserID: id,
		Email:  "test@example.com",
		Iat:    time.Now().Unix(),
		Exp:    time.Now().Add(time.Hour).Unix(),
	}
}

func createTestProject(t *testing.T, mock *MockSharingQueries, ownerID string) database.Project {
	project, err := mock.CreateProject(context.Background(), database.CreateProjectParams{
		Name:        "Test Project",
		Description: sql.NullString{String: "A test project", Valid: true},
		OwnerID:     ownerID,
		CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	return project
}

func createTestUserInMock(t *testing.T, mock *MockSharingQueries, userID string) {
	_, err := mock.CreateUser(context.Background(), database.CreateUserParams{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashed-password",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

// Test cases
func TestAddUserToProject(t *testing.T) {
	mock := NewMockSharingQueries()
	handler := NewProjectSharingHandler(mock)

	ownerID := uuid.New().String()
	userID := uuid.New().String()
	claims := createSharingTestUser(ownerID)

	// Create test project
	_ = createTestProject(t, mock, ownerID)
	// Create test user
	createTestUserInMock(t, mock, userID)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful add user as viewer",
			requestBody: AddUserRequest{
				UserID: userID,
				Role:   "viewer",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response ProjectUserResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.UserID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, response.UserID)
				}
				if response.Role != "viewer" {
					t.Errorf("Expected role viewer, got %s", response.Role)
				}
			},
		},
		{
			name: "successful add user as editor",
			requestBody: AddUserRequest{
				UserID: userID,
				Role:   "editor",
			},
			expectedStatus: http.StatusConflict, // User already added from previous test
			checkResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error != "User is already a member of this project" {
					t.Errorf("Expected conflict error, got: %s", response.Error)
				}
			},
		},
		{
			name: "invalid role",
			requestBody: AddUserRequest{
				UserID: uuid.New().String(),
				Role:   "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-owner tries to add user",
			requestBody: AddUserRequest{
				UserID: uuid.New().String(),
				Role:   "viewer",
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For non-owner test, use different claims
			testClaims := claims
			if tt.name == "non-owner tries to add user" {
				testClaims = createSharingTestUser(uuid.New().String())
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/projects/1/members", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), "user", testClaims))
			w := httptest.NewRecorder()

			err := handler.AddUserToProject(w, req)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestGetProjectUsers(t *testing.T) {
	mock := NewMockSharingQueries()
	handler := NewProjectSharingHandler(mock)

	ownerID := uuid.New().String()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()
	claims := createSharingTestUser(ownerID)

	// Create test project
	project := createTestProject(t, mock, ownerID)
	// Create test users
	createTestUserInMock(t, mock, userID1)
	createTestUserInMock(t, mock, userID2)

	// Add users to project
	_, err := mock.AddUserToProject(context.Background(), database.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    userID1,
		Role:      "viewer",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to add user1 to project: %v", err)
	}

	_, err = mock.AddUserToProject(context.Background(), database.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    userID2,
		Role:      "editor",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to add user2 to project: %v", err)
	}

	req := httptest.NewRequest("GET", "/projects/1/members", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err = handler.GetProjectUsers(w, req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []ProjectUserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 users, got %d", len(response))
	}
}

func TestListUserProjects(t *testing.T) {
	mock := NewMockSharingQueries()
	handler := NewProjectSharingHandler(mock)

	userID := uuid.New().String()
	claims := createSharingTestUser(userID)

	// Create owned project
	_ = createTestProject(t, mock, userID)

	// Create project owned by someone else
	otherOwnerID := uuid.New().String()
	project2 := createTestProject(t, mock, otherOwnerID)

	// Share second project with user
	createTestUserInMock(t, mock, userID)
	_, err := mock.AddUserToProject(context.Background(), database.AddUserToProjectParams{
		ProjectID: project2.ID,
		UserID:    userID,
		Role:      "viewer",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to add user to project: %v", err)
	}

	req := httptest.NewRequest("GET", "/user/projects", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err = handler.ListUserProjects(w, req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []UserProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 projects (1 owned + 1 shared), got %d", len(response))
	}
}

func TestUpdateUserRole(t *testing.T) {
	mock := NewMockSharingQueries()
	handler := NewProjectSharingHandler(mock)

	ownerID := uuid.New().String()
	userID := uuid.New().String()
	claims := createSharingTestUser(ownerID)

	// Create test project
	project := createTestProject(t, mock, ownerID)
	// Create test user
	createTestUserInMock(t, mock, userID)

	// Add user as viewer
	_, err := mock.AddUserToProject(context.Background(), database.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "viewer",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to add user to project: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "successful role update",
			requestBody: UpdateRoleRequest{
				Role: "editor",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "same role update",
			requestBody: UpdateRoleRequest{
				Role: "viewer", // Need to re-add as viewer first
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "same role update" {
				// Reset back to viewer for this test
				mock.UpdateUserRole(context.Background(), database.UpdateUserRoleParams{
					Role:      "viewer",
					UpdatedAt: time.Now(),
					ProjectID: project.ID,
					UserID:    userID,
				})
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/projects/1/members/"+userID, bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), "user", claims))
			w := httptest.NewRecorder()

			err := handler.UpdateUserRole(w, req)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRemoveUserFromProject(t *testing.T) {
	mock := NewMockSharingQueries()
	handler := NewProjectSharingHandler(mock)

	ownerID := uuid.New().String()
	userID := uuid.New().String()
	claims := createSharingTestUser(ownerID)

	// Create test project
	project := createTestProject(t, mock, ownerID)
	// Create test user
	createTestUserInMock(t, mock, userID)

	// Add user to project
	_, err := mock.AddUserToProject(context.Background(), database.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "viewer",
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to add user to project: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/projects/1/members/"+userID, nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", claims))
	w := httptest.NewRecorder()

	err = handler.RemoveUserFromProject(w, req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify user is removed
	users, err := mock.GetProjectUsers(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("Failed to get project users: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users after removal, got %d", len(users))
	}
}
