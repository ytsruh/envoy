package utils

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	database "ytsruh.com/envoy/server/database/generated"
)

type mockQuerier struct {
	projectOwnerCounts     map[string]int64                // key: "projectID:userID"
	editorPermissionCounts map[string]int64                // key: "projectID:userID"
	accessibleProjects     map[string]database.Project     // key: "projectID:userID"
	projectMemberships     map[string]database.ProjectUser // key: "projectID:userID"
}

func (m *mockQuerier) IsProjectOwner(ctx context.Context, arg database.IsProjectOwnerParams) (int64, error) {
	key := fmt.Sprintf("%d:%s", arg.ID, arg.OwnerID)
	if count, exists := m.projectOwnerCounts[key]; exists {
		return count, nil
	}
	return 0, nil
}

// Add other required methods with no-op implementations
func (m *mockQuerier) AddUserToProject(ctx context.Context, arg database.AddUserToProjectParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, nil
}

func (m *mockQuerier) CanUserModifyEnvironment(ctx context.Context, arg database.CanUserModifyEnvironmentParams) (int64, error) {
	return 0, nil
}

func (m *mockQuerier) CanUserModifyEnvironmentVariable(ctx context.Context, arg database.CanUserModifyEnvironmentVariableParams) (int64, error) {
	return 0, nil
}

func (m *mockQuerier) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	// Check if user is owner (OwnerID and UserID are the same)
	if arg.OwnerID == arg.UserID {
		key := fmt.Sprintf("%d:%s", arg.ID, arg.OwnerID)
		if ownerCount, exists := m.projectOwnerCounts[key]; exists {
			return ownerCount, nil
		}
	}

	// Check if user has editor permissions (OwnerID and UserID are different in tests)
	key := fmt.Sprintf("%d:%s:%s", arg.ID, arg.OwnerID, arg.UserID)
	if count, exists := m.editorPermissionCounts[key]; exists {
		return count, nil
	}
	return 0, nil
}

func (m *mockQuerier) CreateEnvironment(ctx context.Context, arg database.CreateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *mockQuerier) CreateProject(ctx context.Context, arg database.CreateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m *mockQuerier) DeleteEnvironment(ctx context.Context, arg database.DeleteEnvironmentParams) error {
	return nil
}

func (m *mockQuerier) DeleteProject(ctx context.Context, arg database.DeleteProjectParams) error {
	return nil
}

func (m *mockQuerier) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m *mockQuerier) GetAccessibleEnvironment(ctx context.Context, arg database.GetAccessibleEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *mockQuerier) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	// Check if user is owner (OwnerID and UserID are the same)
	if arg.OwnerID == arg.UserID {
		key := fmt.Sprintf("%d:%s", arg.ID, arg.OwnerID)
		if ownerCount, exists := m.projectOwnerCounts[key]; exists && ownerCount > 0 {
			// Return a dummy project for owners
			return database.Project{ID: arg.ID, Name: "Test Project"}, nil
		}
	}

	// Check if user has viewer permissions
	key := fmt.Sprintf("%d:%s:%s", arg.ID, arg.OwnerID, arg.UserID)
	if project, exists := m.accessibleProjects[key]; exists {
		return project, nil
	}
	return database.Project{}, sql.ErrNoRows
}

func (m *mockQuerier) GetEnvironment(ctx context.Context, id int64) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *mockQuerier) GetProject(ctx context.Context, id int64) (database.Project, error) {
	return database.Project{}, nil
}

func (m *mockQuerier) GetProjectByGitRepo(ctx context.Context, arg database.GetProjectByGitRepoParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m *mockQuerier) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	return "", nil
}

func (m *mockQuerier) GetProjectMembership(ctx context.Context, arg database.GetProjectMembershipParams) (database.ProjectUser, error) {
	key := fmt.Sprintf("%d:%s", arg.ProjectID, arg.UserID)
	if membership, exists := m.projectMemberships[key]; exists {
		return membership, nil
	}
	return database.ProjectUser{}, sql.ErrNoRows
}

func (m *mockQuerier) GetProjectUsers(ctx context.Context, projectID int64) ([]database.ProjectUser, error) {
	return []database.ProjectUser{}, nil
}

func (m *mockQuerier) GetUser(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return database.User{}, nil
}

func (m *mockQuerier) GetUserProjects(ctx context.Context, arg database.GetUserProjectsParams) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m *mockQuerier) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *mockQuerier) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]database.Environment, error) {
	return []database.Environment{}, nil
}

func (m *mockQuerier) ListProjectsByOwner(ctx context.Context, ownerID string) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m *mockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return []database.User{}, nil
}

func (m *mockQuerier) RemoveUserFromProject(ctx context.Context, arg database.RemoveUserFromProjectParams) error {
	return nil
}

func (m *mockQuerier) UpdateEnvironment(ctx context.Context, arg database.UpdateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *mockQuerier) UpdateProject(ctx context.Context, arg database.UpdateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m *mockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m *mockQuerier) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	return nil
}

func (m *mockQuerier) CreateEnvironmentVariable(ctx context.Context, arg database.CreateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *mockQuerier) DeleteEnvironmentVariable(ctx context.Context, id int64) error {
	return nil
}

func (m *mockQuerier) GetAccessibleEnvironmentVariable(ctx context.Context, arg database.GetAccessibleEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *mockQuerier) GetEnvironmentVariable(ctx context.Context, id int64) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m *mockQuerier) ListEnvironmentVariablesByEnvironment(ctx context.Context, environmentID int64) ([]database.EnvironmentVariable, error) {
	return []database.EnvironmentVariable{}, nil
}

func (m *mockQuerier) UpdateEnvironmentVariable(ctx context.Context, arg database.UpdateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func TestAccessControlService_RequireOwner(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		projectID     int64
		setupCounts   map[string]int64
		expectedError string
	}{
		{
			name:      "owner can access project",
			userID:    "user1",
			projectID: 1,
			setupCounts: map[string]int64{
				"1:user1": 1,
			},
			expectedError: "",
		},
		{
			name:      "non-owner cannot access project",
			userID:    "user2",
			projectID: 1,
			setupCounts: map[string]int64{
				"1:user1": 1,
			},
			expectedError: "access denied",
		},
		{
			name:          "no owner data",
			userID:        "user1",
			projectID:     1,
			setupCounts:   map[string]int64{},
			expectedError: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockQuerier{
				projectOwnerCounts: tt.setupCounts,
			}
			service := NewAccessControlService(mock)

			err := service.RequireOwner(context.Background(), tt.projectID, tt.userID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if len(err.Error()) == 0 || !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%s'", err.Error())
				}
			}
		})
	}
}

func TestAccessControlService_RequireEditor(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		projectID         int64
		setupOwnerCounts  map[string]int64
		setupEditorCounts map[string]int64
		expectedError     string
	}{
		{
			name:      "owner can edit project",
			userID:    "user1",
			projectID: 1,
			setupOwnerCounts: map[string]int64{
				"1:user1": 1,
			},
			expectedError: "",
		},
		{
			name:             "editor can edit project",
			userID:           "user2",
			projectID:        1,
			setupOwnerCounts: map[string]int64{},
			setupEditorCounts: map[string]int64{
				"1:user2:user2": 1,
			},
			expectedError: "",
		},
		{
			name:      "viewer cannot edit project",
			userID:    "user3",
			projectID: 1,
			setupEditorCounts: map[string]int64{
				"1:user3:user3": 0, // user3 is not an editor
			},
			expectedError: "access denied",
		},
		{
			name:          "no editor data",
			userID:        "user1",
			projectID:     1,
			expectedError: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockQuerier{
				projectOwnerCounts:     tt.setupOwnerCounts,
				editorPermissionCounts: tt.setupEditorCounts,
			}
			service := NewAccessControlService(mock)

			err := service.RequireEditor(context.Background(), tt.projectID, tt.userID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if len(err.Error()) == 0 || !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%s'", err.Error())
				}
			}
		})
	}
}

func TestAccessControlService_RequireViewer(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		projectID        int64
		setupOwnerCounts map[string]int64
		setupProjects    map[string]database.Project
		expectedError    string
	}{
		{
			name:      "owner can view project",
			userID:    "user1",
			projectID: 1,
			setupOwnerCounts: map[string]int64{
				"1:user1": 1,
			},
			expectedError: "",
		},
		{
			name:      "viewer can view project",
			userID:    "user2",
			projectID: 1,
			setupProjects: map[string]database.Project{
				"1:user2:user2": {ID: 1, Name: "Test Project"},
			},
			expectedError: "",
		},
		{
			name:      "non-member cannot view project",
			userID:    "user3",
			projectID: 1,
			setupProjects: map[string]database.Project{
				"1:user2:user2": {ID: 1, Name: "Test Project"},
			},
			expectedError: "access denied",
		},
		{
			name:          "no accessible project data",
			userID:        "user1",
			projectID:     1,
			expectedError: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockQuerier{
				projectOwnerCounts: tt.setupOwnerCounts,
				accessibleProjects: tt.setupProjects,
			}
			service := NewAccessControlService(mock)

			err := service.RequireViewer(context.Background(), tt.projectID, tt.userID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if len(err.Error()) == 0 || !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%s'", err.Error())
				}
			}
		})
	}
}

func TestAccessControlService_GetRole(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		projectID        int64
		setupOwnerCounts map[string]int64
		setupMemberships map[string]database.ProjectUser
		expectedRole     string
		expectedError    string
	}{
		{
			name:      "user is owner",
			userID:    "user1",
			projectID: 1,
			setupOwnerCounts: map[string]int64{
				"1:user1": 1,
			},
			expectedRole:  "owner",
			expectedError: "",
		},
		{
			name:             "user is editor",
			userID:           "user2",
			projectID:        1,
			setupOwnerCounts: map[string]int64{},
			setupMemberships: map[string]database.ProjectUser{
				"1:user2": {ProjectID: 1, UserID: "user2", Role: "editor"},
			},
			expectedRole:  "editor",
			expectedError: "",
		},
		{
			name:             "user is viewer",
			userID:           "user3",
			projectID:        1,
			setupOwnerCounts: map[string]int64{},
			setupMemberships: map[string]database.ProjectUser{
				"1:user3": {ProjectID: 1, UserID: "user3", Role: "viewer"},
			},
			expectedRole:  "viewer",
			expectedError: "",
		},
		{
			name:             "user is not member",
			userID:           "user4",
			projectID:        1,
			setupOwnerCounts: map[string]int64{},
			setupMemberships: map[string]database.ProjectUser{},
			expectedRole:     "",
			expectedError:    "user is not a project member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockQuerier{
				projectOwnerCounts: tt.setupOwnerCounts,
				projectMemberships: tt.setupMemberships,
			}
			service := NewAccessControlService(mock)

			role, err := service.GetRole(context.Background(), tt.projectID, tt.userID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if len(err.Error()) == 0 || !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%s'", err.Error())
				}
				if role != tt.expectedRole {
					t.Errorf("expected role '%s', got '%s'", tt.expectedRole, role)
				}
			}
		})
	}
}

func TestNewAccessControlService(t *testing.T) {
	mock := &mockQuerier{}
	service := NewAccessControlService(mock)

	if service == nil {
		t.Fatal("NewAccessControlService() returned nil")
	}

	// Test that the service implements the interface by using it
	var _ AccessControlService = service

	// Check that the service has the expected methods by calling them
	err := service.RequireOwner(context.Background(), 1, "test")
	if err != nil {
		// Expected to fail since mock is empty, but method should exist
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
