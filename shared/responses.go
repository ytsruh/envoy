package shared

// LoginResponse contains the authentication token and user information after successful login.
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    UserID    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt Timestamp `json:"created_at"`
}

// ProjectResponse represents a project in the system.
type ProjectResponse struct {
	ID        ProjectID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`
}

// EnvironmentResponse represents an environment within a project.
type EnvironmentResponse struct {
	ID        ProjectID `json:"id"`
	ProjectID ProjectID `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`
}

// EnvironmentVariableResponse represents an environment variable.
type EnvironmentVariableResponse struct {
	ID            EnvironmentVariableID `json:"id"`
	EnvironmentID EnvironmentID         `json:"environment_id"`
	Name          string                `json:"name"`
	Value         string                `json:"value"`
	Description   *string               `json:"description"`
	CreatedAt     Timestamp             `json:"created_at"`
	UpdatedAt     Timestamp             `json:"updated_at"`
}

// ProjectShareResponse represents a user's access to a shared project.
type ProjectShareResponse struct {
	ID        int64     `json:"id"`
	ProjectID ProjectID `json:"project_id"`
	UserID    UserID    `json:"user_id"`
	Role      Role      `json:"role"`
	CreatedAt Timestamp `json:"created_at"`
}

// UserProjectResponse represents a project that a user has access to.
type UserProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	GitRepo     *string   `json:"git_repo"`
	OwnerID     UserID    `json:"owner_id"`
	CreatedAt   Timestamp `json:"created_at"`
	UpdatedAt   Timestamp `json:"updated_at"`
}

// RegisterResponse contains the user information after successful registration.
type RegisterResponse struct {
	UserID    UserID    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt Timestamp `json:"created_at"`
}

// ProjectsResponse is a wrapper for a list of projects.
type ProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

// EnvironmentsResponse is a wrapper for a list of environments.
type EnvironmentsResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
}

// EnvironmentVariablesResponse is a wrapper for a list of environment variables.
type EnvironmentVariablesResponse struct {
	Variables []EnvironmentVariableResponse `json:"variables"`
}
