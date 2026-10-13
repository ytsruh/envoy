use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use envoy_shared::{
    requests::{AddMemberRequest, UpdateRoleRequest},
    responses::ProjectMemberResponse,
    types::Role,
};
use uuid::Uuid;

use crate::{error::{ApiError, message}, extract::AuthUser, rows::get_string, state::AppState};

pub async fn add_member(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(project_id): Path<String>,
    Json(req): Json<AddMemberRequest>,
) -> Result<(StatusCode, Json<ProjectMemberResponse>), ApiError> {
    if req.user_id.is_empty() {
        return Err(ApiError::bad("user_id is required"));
    }
    let role = Role::parse(&req.role)
        .filter(|r| r.is_assignable())
        .ok_or_else(|| ApiError::bad("role must be one of: viewer, editor"))?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_owner(&conn, &project_id, &user.user_id)
        .await
        .map_err(|_| ApiError::forbidden("only project owners can add users"))?;
    let mut target = conn
        .query(
            "SELECT id FROM users WHERE id = ?1 AND deleted_at IS NULL",
            libsql::params![req.user_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    if target.next().await.map_err(|e| ApiError::Internal(e.into()))?.is_none() {
        return Err(ApiError::not_found("user not found"));
    }
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    conn.execute(
        "INSERT INTO project_users (id, project_id, user_id, role, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?5) ON CONFLICT(project_id, user_id) DO UPDATE SET role = excluded.role, updated_at = excluded.updated_at",
        libsql::params![id, project_id.clone(), req.user_id.clone(), role.as_str().to_string(), now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok((
        StatusCode::CREATED,
        Json(ProjectMemberResponse {
            user_id: req.user_id,
            role: role.as_str().to_string(),
            created_at: now,
        }),
    ))
}

pub async fn list_members(
    State(state): State<AppState>,
    _user: AuthUser,
    Path(project_id): Path<String>,
) -> Result<Json<Vec<ProjectMemberResponse>>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    let mut rows = conn
        .query(
            "SELECT user_id, role, created_at FROM project_users WHERE project_id = ?1",
            libsql::params![project_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows.next().await.map_err(|e| ApiError::Internal(e.into()))? {
        let created: Option<String> = row.get(2).map_err(|e| ApiError::Internal(e.into()))?;
        out.push(ProjectMemberResponse {
            user_id: get_string(&row, 0).map_err(|e| ApiError::Internal(e.into()))?,
            role: get_string(&row, 1).map_err(|e| ApiError::Internal(e.into()))?,
            created_at: created.unwrap_or_default(),
        });
    }
    Ok(Json(out))
}

pub async fn update_member_role(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, member_id)): Path<(String, String)>,
    Json(req): Json<UpdateRoleRequest>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    if member_id.is_empty() {
        return Err(ApiError::bad("invalid user ID"));
    }
    let role = Role::parse(&req.role)
        .filter(|r| r.is_assignable())
        .ok_or_else(|| ApiError::bad("role must be one of: viewer, editor"))?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_owner(&conn, &project_id, &user.user_id)
        .await
        .map_err(|_| ApiError::forbidden("only project owners can update user roles"))?;
    conn.execute(
        "UPDATE project_users SET role = ?1, updated_at = CURRENT_TIMESTAMP WHERE project_id = ?2 AND user_id = ?3",
        libsql::params![role.as_str().to_string(), project_id.clone(), member_id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("User role updated successfully"))
}

pub async fn remove_member(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, member_id)): Path<(String, String)>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    if member_id.is_empty() {
        return Err(ApiError::bad("invalid user ID"));
    }
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_owner(&conn, &project_id, &user.user_id)
        .await
        .map_err(|_| ApiError::forbidden("only project owners can remove users"))?;
    conn.execute(
        "DELETE FROM project_users WHERE project_id = ?1 AND user_id = ?2",
        libsql::params![project_id.clone(), member_id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("User removed successfully"))
}
