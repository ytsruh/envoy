use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use envoy_shared::{
    requests::{CreateProjectRequest, UpdateProjectRequest},
    responses::ProjectResponse,
    validation,
};
use uuid::Uuid;

use crate::{
    error::{ApiError, message},
    extract::AuthUser,
    rows::{get_opt_string, get_string, ts_from_row_string},
    state::AppState,
};

fn row_to_project(row: &libsql::Row) -> anyhow::Result<ProjectResponse> {
    Ok(ProjectResponse {
        id: get_string(row, 0)?,
        name: get_string(row, 1)?,
        description: get_opt_string(row, 2)?,
        git_repo: get_opt_string(row, 3)?,
        owner_id: get_string(row, 4)?,
        created_at: ts_from_row_string(row.get(5)?),
        updated_at: ts_from_row_string(row.get(6)?),
    })
}

const SELECT: &str = "SELECT id, name, description, git_repo, owner_id, created_at, updated_at FROM projects WHERE deleted_at IS NULL";

pub async fn create_project(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Json(req): Json<CreateProjectRequest>,
) -> Result<(StatusCode, Json<ProjectResponse>), ApiError> {
    validation::validate_name(&req.name).map_err(ApiError::bad)?;
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    validation::validate_git_repo(&req.git_repo).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    if !req.git_repo.is_empty() {
        let mut rows = conn
            .query(
                "SELECT id FROM projects WHERE owner_id = ?1 AND git_repo = ?2 AND deleted_at IS NULL",
                libsql::params![user.user_id.clone(), req.git_repo.clone()],
            )
            .await
            .map_err(|e| ApiError::Internal(e.into()))?;
        if rows.next().await.map_err(|e| ApiError::Internal(e.into()))?.is_some() {
            return Err(ApiError::conflict(
                "a project with this git repository already exists",
            ));
        }
    }
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let repo: Option<String> = if req.git_repo.is_empty() {
        None
    } else {
        Some(req.git_repo.clone())
    };
    conn.execute(
        "INSERT INTO projects (id, name, description, git_repo, owner_id, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?6)",
        libsql::params![id.clone(), req.name.clone(), desc.clone(), repo.clone(), user.user_id.clone(), now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    let ts = envoy_shared::types::parse_sqlite_time(&now).into();
    Ok((
        StatusCode::CREATED,
        Json(ProjectResponse {
            id,
            name: req.name,
            description: desc,
            git_repo: repo,
            owner_id: user.user_id,
            created_at: ts,
            updated_at: ts,
        }),
    ))
}

pub async fn list_projects(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
) -> Result<Json<Vec<ProjectResponse>>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    let mut rows = conn
        .query(
            &format!("{SELECT} AND (owner_id = ?1 OR id IN (SELECT project_id FROM project_users WHERE user_id = ?1)) ORDER BY created_at DESC"),
            libsql::params![user.user_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows.next().await.map_err(|e| ApiError::Internal(e.into()))? {
        out.push(row_to_project(&row).map_err(|e| ApiError::Internal(e.into()))?);
    }
    Ok(Json(out))
}

pub async fn list_user_projects(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
) -> Result<Json<Vec<ProjectResponse>>, ApiError> {
    list_projects(State(state), AuthUser(user)).await
}

pub async fn get_project(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(id): Path<String>,
) -> Result<Json<ProjectResponse>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    let mut rows = conn
        .query(
            &format!("{SELECT} AND id = ?1 AND (owner_id = ?2 OR id IN (SELECT project_id FROM project_users WHERE user_id = ?2))"),
            libsql::params![id.clone(), user.user_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("project not found or access denied"))?;
    Ok(Json(row_to_project(&row).map_err(|e| ApiError::Internal(e.into()))?))
}

pub async fn update_project(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(id): Path<String>,
    Json(req): Json<UpdateProjectRequest>,
) -> Result<Json<ProjectResponse>, ApiError> {
    validation::validate_name(&req.name).map_err(ApiError::bad)?;
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    validation::validate_git_repo(&req.git_repo).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_editor(&conn, &id, &user.user_id).await?;
    let mut rows = conn
        .query(
            &format!("{SELECT} AND id = ?1"),
            libsql::params![id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    rows.next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("project not found or access denied"))?;
    if !req.git_repo.is_empty() {
        let mut clash = conn
            .query(
                "SELECT id FROM projects WHERE owner_id = ?1 AND git_repo = ?2 AND id != ?3 AND deleted_at IS NULL",
                libsql::params![user.user_id.clone(), req.git_repo.clone(), id.clone()],
            )
            .await
            .map_err(|e| ApiError::Internal(e.into()))?;
        if clash.next().await.map_err(|e| ApiError::Internal(e.into()))?.is_some() {
            return Err(ApiError::conflict(
                "a project with this git repository already exists",
            ));
        }
    }
    let now = envoy_shared::types::sqlite_now();
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let repo: Option<String> = if req.git_repo.is_empty() {
        None
    } else {
        Some(req.git_repo.clone())
    };
    conn.execute(
        "UPDATE projects SET name = ?1, description = ?2, git_repo = ?3, updated_at = ?4 WHERE id = ?5",
        libsql::params![req.name.clone(), desc.clone(), repo.clone(), now.clone(), id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    get_project(State(state), AuthUser(user), Path(id)).await
}

pub async fn delete_project(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(id): Path<String>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_owner(&conn, &id, &user.user_id).await.map_err(|_| {
        ApiError::forbidden("only project owners can delete projects")
    })?;
    conn.execute(
        "UPDATE projects SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?1",
        libsql::params![id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("Project deleted successfully"))
}
