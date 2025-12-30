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
	GetProject(projectID string) (*ProjectResponse, error)
	UpdateProject(projectID string, name, description, gitRepo string) (*ProjectResponse, error)
	DeleteProject(projectID string) error

	CreateEnvironment(projectID string, name, description string) (*EnvironmentResponse, error)
	ListEnvironments(projectID string) ([]EnvironmentResponse, error)
	GetEnvironment(projectID string, environmentID string) (*EnvironmentResponse, error)
	UpdateEnvironment(projectID string, environmentID string, name, description string) (*EnvironmentResponse, error)
	DeleteEnvironment(projectID string, environmentID string) error

	CreateEnvironmentVariable(projectID string, environmentID string, key, value string) (*EnvironmentVariableResponse, error)
	ListEnvironmentVariables(projectID string, environmentID string) ([]EnvironmentVariableResponse, error)
	GetEnvironmentVariable(projectID string, environmentID string, variableID string) (*EnvironmentVariableResponse, error)
	UpdateEnvironmentVariable(projectID string, environmentID string, variableID string, key, value string) (*EnvironmentVariableResponse, error)
	DeleteEnvironmentVariable(projectID string, environmentID string, variableID string) error
}
