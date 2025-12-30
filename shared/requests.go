package shared

// CreateProjectRequest is used to create a new project.
type CreateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateProjectRequest is used to update an existing project.
type UpdateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CreateEnvironmentRequest is used to create a new environment within a project.
type CreateEnvironmentRequest struct {
	ProjectID ProjectID `json:"project_id" validate:"required"`
	Name      string    `json:"name" validate:"required,min=1,max=100"`
}

// UpdateEnvironmentRequest is used to update an existing environment.
type UpdateEnvironmentRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CreateEnvironmentVariableRequest is used to create a new environment variable.
type CreateEnvironmentVariableRequest struct {
	EnvironmentID EnvironmentID `json:"environment_id" validate:"required"`
	Name          string        `json:"name" validate:"required,min=1,max=255"`
	Value         string        `json:"value" validate:"required"`
	Description   string        `json:"description" validate:"omitempty,max=1000"`
}

// UpdateEnvironmentVariableRequest is used to update an existing environment variable.
type UpdateEnvironmentVariableRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

// LoginRequest is used to authenticate a user.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest is used to register a new user account.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// ShareProjectRequest is used to share a project with another user.
type ShareProjectRequest struct {
	UserID UserID `json:"user_id" validate:"required"`
	Role   Role   `json:"role" validate:"required,oneof=editor viewer"`
}

// UpdateRoleRequest is used to update a user's role in a project.
type UpdateRoleRequest struct {
	Role Role `json:"role" validate:"required,oneof=editor viewer"`
}
