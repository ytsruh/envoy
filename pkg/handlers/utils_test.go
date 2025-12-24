package handlers

import (
	"context"

	database "ytsruh.com/envoy/pkg/database/generated"
)

type mockQuerier struct{}

func (m mockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m mockQuerier) GetUser(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m mockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return []database.User{}, nil
}

func (m mockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) CreateProject(ctx context.Context, arg database.CreateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m mockQuerier) GetProject(ctx context.Context, id int64) (database.Project, error) {
	return database.Project{}, nil
}

func (m mockQuerier) ListProjectsByOwner(ctx context.Context, ownerID string) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m mockQuerier) UpdateProject(ctx context.Context, arg database.UpdateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m mockQuerier) DeleteProject(ctx context.Context, arg database.DeleteProjectParams) error {
	return nil
}

// Project sharing methods
func (m mockQuerier) AddUserToProject(ctx context.Context, arg database.AddUserToProjectParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, nil
}

func (m mockQuerier) RemoveUserFromProject(ctx context.Context, arg database.RemoveUserFromProjectParams) error {
	return nil
}

func (m mockQuerier) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	return nil
}

func (m mockQuerier) GetProjectUsers(ctx context.Context, projectID int64) ([]database.ProjectUser, error) {
	return []database.ProjectUser{}, nil
}

func (m mockQuerier) GetUserProjects(ctx context.Context, arg database.GetUserProjectsParams) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m mockQuerier) GetProjectMembership(ctx context.Context, arg database.GetProjectMembershipParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, nil
}

func (m mockQuerier) IsProjectOwner(ctx context.Context, arg database.IsProjectOwnerParams) (int64, error) {
	return 0, nil
}

func (m mockQuerier) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m mockQuerier) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	return 0, nil
}

func (m mockQuerier) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	return "", nil
}

// Environment methods
func (m mockQuerier) CreateEnvironment(ctx context.Context, arg database.CreateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m mockQuerier) GetEnvironment(ctx context.Context, id int64) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m mockQuerier) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]database.Environment, error) {
	return []database.Environment{}, nil
}

func (m mockQuerier) UpdateEnvironment(ctx context.Context, arg database.UpdateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m mockQuerier) DeleteEnvironment(ctx context.Context, arg database.DeleteEnvironmentParams) error {
	return nil
}

func (m mockQuerier) GetAccessibleEnvironment(ctx context.Context, arg database.GetAccessibleEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m mockQuerier) CanUserModifyEnvironment(ctx context.Context, arg database.CanUserModifyEnvironmentParams) (int64, error) {
	return 0, nil
}

// Environment variable methods
func (m mockQuerier) CreateEnvironmentVariable(ctx context.Context, arg database.CreateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m mockQuerier) DeleteEnvironmentVariable(ctx context.Context, id int64) error {
	return nil
}

func (m mockQuerier) GetAccessibleEnvironmentVariable(ctx context.Context, arg database.GetAccessibleEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m mockQuerier) GetEnvironmentVariable(ctx context.Context, id int64) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m mockQuerier) ListEnvironmentVariablesByEnvironment(ctx context.Context, environmentID int64) ([]database.EnvironmentVariable, error) {
	return []database.EnvironmentVariable{}, nil
}

func (m mockQuerier) UpdateEnvironmentVariable(ctx context.Context, arg database.UpdateEnvironmentVariableParams) (database.EnvironmentVariable, error) {
	return database.EnvironmentVariable{}, nil
}

func (m mockQuerier) CanUserModifyEnvironmentVariable(ctx context.Context, arg database.CanUserModifyEnvironmentVariableParams) (int64, error) {
	return 0, nil
}
