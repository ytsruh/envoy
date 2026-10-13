use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

/// Matches Go handler-local structs (handlers define their own shapes;
/// shared/requests.go in Go is dead code for projects/envs/vars).
#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct CreateProjectRequest {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub git_repo: String,
}

pub type UpdateProjectRequest = CreateProjectRequest;

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct CreateEnvironmentRequest {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub description: String,
}

pub type UpdateEnvironmentRequest = CreateEnvironmentRequest;

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct CreateVariableRequest {
    #[serde(default)]
    pub key: String,
    #[serde(default)]
    pub value: String,
    #[serde(default)]
    pub description: String,
}

pub type UpdateVariableRequest = CreateVariableRequest;

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct RegisterRequest {
    pub name: String,
    pub email: String,
    pub password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct AddMemberRequest {
    pub user_id: String,
    pub role: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, ToSchema)]
pub struct UpdateRoleRequest {
    pub role: String,
}
