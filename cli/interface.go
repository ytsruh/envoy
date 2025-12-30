package cli

import "ytsruh.com/envoy/cli/controllers"

type AuthResponse = controllers.AuthResponse
type ProfileResponse = controllers.ProfileResponse
type ProjectResponse = controllers.ProjectResponse
type EnvironmentResponse = controllers.EnvironmentResponse
type EnvironmentVariableResponse = controllers.EnvironmentVariableResponse

type APIClient interface {
	Register(name, email, password string) (*AuthResponse, error)
	Login(email, password string) (*AuthResponse, error)
	GetProfile() (*ProfileResponse, error)

	CreateProject(name, description, gitRepo string) (*ProjectResponse, error)
	ListProjects() ([]ProjectResponse, error)
	GetProject(projectID int64) (*ProjectResponse, error)
	UpdateProject(projectID int64, name, description, gitRepo string) (*ProjectResponse, error)
	DeleteProject(projectID int64) error

	CreateEnvironment(projectID int64, name, description string) (*EnvironmentResponse, error)
	ListEnvironments(projectID int64) ([]EnvironmentResponse, error)
	GetEnvironment(projectID, environmentID int64) (*EnvironmentResponse, error)
	UpdateEnvironment(projectID, environmentID int64, name, description string) (*EnvironmentResponse, error)
	DeleteEnvironment(projectID, environmentID int64) error

	CreateEnvironmentVariable(projectID, environmentID int64, key, value string) (*EnvironmentVariableResponse, error)
	ListEnvironmentVariables(projectID, environmentID int64) ([]EnvironmentVariableResponse, error)
	GetEnvironmentVariable(projectID, environmentID, variableID int64) (*EnvironmentVariableResponse, error)
	UpdateEnvironmentVariable(projectID, environmentID, variableID int64, key, value string) (*EnvironmentVariableResponse, error)
	DeleteEnvironmentVariable(projectID, environmentID, variableID int64) error
}
