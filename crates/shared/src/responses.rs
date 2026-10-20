use serde::{Deserialize, Serialize};

use crate::types::Timestamp;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorBody {
    pub error: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageBody {
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisteredUser {
    pub user_id: String,
    pub name: String,
    pub email: String,
    pub created_at: Timestamp,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthResponse {
    pub token: String,
    pub user: RegisteredUser,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProfileResponse {
    pub user_id: String,
    pub email: String,
    pub issued_at: i64,
    pub expires_at: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserSearchResponse {
    pub user_id: String,
    pub name: String,
    pub email: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InviteResponse {
    pub invite_id: String,
    pub email: String,
    pub token: String,
    pub expires_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InviteListItem {
    pub invite_id: String,
    pub email: String,
    pub name: Option<String>,
    pub expires_at: String,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectResponse {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub git_repo: Option<String>,
    pub owner_id: String,
    pub created_at: Timestamp,
    pub updated_at: Timestamp,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnvironmentResponse {
    pub id: String,
    pub project_id: String,
    pub name: String,
    pub description: Option<String>,
    pub created_at: Timestamp,
    pub updated_at: Timestamp,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VariableResponse {
    pub id: String,
    pub environment_id: String,
    pub key: String,
    pub value: String,
    pub description: Option<String>,
    #[serde(default)]
    pub file_path: String,
    pub created_at: Timestamp,
    pub updated_at: Timestamp,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectMemberResponse {
    pub user_id: String,
    pub role: String,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub health: String,
}
