package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
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
		if p.ID == id {
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
		if p.ID == arg.ID && !p.DeletedAt.Valid {
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
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID {
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
	for _, p := range m.projects {
		if p.ID == arg.ID && p.OwnerID == arg.OwnerID && !p.DeletedAt.Valid {
			return 1, nil
		}
	}
	return 0, nil
}

func (m *MockQueries) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return p, nil
			}
		}
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *MockQueries) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	if m.canUserModifyResult != 0 {
		return m.canUserModifyResult, nil
	}

	for _, p := range m.projects {
		if p.ID == arg.ID && !p.DeletedAt.Valid {
			if p.OwnerID == arg.OwnerID || p.OwnerID == arg.UserID {
				return 1, nil
			}
			break
		}
	}
	return 0, nil
}

func (m *MockQueries) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	return "", sql.ErrNoRows
}

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

func TestCreateProject(t *testing.T) {
	mock := &MockQueries{
		projects: []database.Project{},
		users:    make(map[string]database.User),
	}
	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)
	claims := CreateTestUser()

	reqBody := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", claims)

	err := CreateProject(c, ctx)
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
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
