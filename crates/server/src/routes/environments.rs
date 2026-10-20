use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use envoy_shared::{
    requests::{CreateEnvironmentRequest, UpdateEnvironmentRequest},
    responses::EnvironmentResponse,
    validation,
};
use uuid::Uuid;

use crate::{
    error::{ApiError, message},
    extract::AuthUser,
    rows::{get_opt_string, get_string, ts_from_row_string},
    state::AppState,
};

fn row_to_env(row: &libsql::Row, project_id: &str) -> anyhow::Result<EnvironmentResponse> {
    Ok(EnvironmentResponse {
        id: get_string(row, 0)?,
        project_id: project_id.to_string(),
        name: get_string(row, 1)?,
        description: get_opt_string(row, 2)?,
        created_at: ts_from_row_string(row.get(3)?),
        updated_at: ts_from_row_string(row.get(4)?),
    })
}

pub async fn create_environment(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(project_id): Path<String>,
    Json(req): Json<CreateEnvironmentRequest>,
) -> Result<(StatusCode, Json<EnvironmentResponse>), ApiError> {
    validation::validate_name(&req.name).map_err(ApiError::bad)?;
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    conn.execute(
        "INSERT INTO environments (id, project_id, name, description, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?5)",
        libsql::params![id.clone(), project_id.clone(), req.name.clone(), desc.clone(), now.clone()],
    )
    .await
    .map_err(|e| {
        let msg = e.to_string();
        if msg.contains("UNIQUE") {
            ApiError::conflict("an environment with this name already exists")
        } else {
            ApiError::Internal(e.into())
        }
    })?;
    let ts = envoy_shared::types::parse_sqlite_time(&now).into();
    Ok((
        StatusCode::CREATED,
        Json(EnvironmentResponse {
            id,
            project_id,
            name: req.name,
            description: desc,
            created_at: ts,
            updated_at: ts,
        }),
    ))
}

pub async fn list_environments(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(project_id): Path<String>,
) -> Result<Json<Vec<EnvironmentResponse>>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    crate::access::require_viewer(&conn, &project_id, &user.user_id).await?;
    let mut rows = conn
        .query(
            "SELECT id, name, description, created_at, updated_at FROM environments WHERE project_id = ?1 AND deleted_at IS NULL ORDER BY created_at",
            libsql::params![project_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows.next().await.map_err(|e| ApiError::Internal(e.into()))? {
        out.push(row_to_env(&row, &project_id).map_err(ApiError::Internal)?);
    }
    Ok(Json(out))
}

async fn fetch_env(
    conn: &libsql::Connection,
    env_id: &str,
) -> Result<EnvironmentResponse, ApiError> {
    let mut rows = conn
        .query(
            "SELECT id, project_id, name, description, created_at, updated_at FROM environments WHERE id = ?1 AND deleted_at IS NULL",
            libsql::params![env_id.to_string()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("environment not found or access denied"))?;
    Ok(EnvironmentResponse {
        id: get_string(&row, 0).map_err(ApiError::Internal)?,
        project_id: get_string(&row, 1).map_err(ApiError::Internal)?,
        name: get_string(&row, 2).map_err(ApiError::Internal)?,
        description: {
            let v: Option<String> = row.get(3).map_err(|e| ApiError::Internal(e.into()))?;
            v.filter(|s| !s.is_empty())
        },
        created_at: ts_from_row_string(row.get(4).map_err(|e| ApiError::Internal(e.into()))?),
        updated_at: ts_from_row_string(row.get(5).map_err(|e| ApiError::Internal(e.into()))?),
    })
}

pub async fn get_environment(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(env_id): Path<String>,
) -> Result<Json<EnvironmentResponse>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let env = fetch_env(&conn, &env_id).await?;
    crate::access::require_viewer(&conn, &env.project_id, &user.user_id).await?;
    Ok(Json(env))
}

pub async fn update_environment(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(env_id): Path<String>,
    Json(req): Json<UpdateEnvironmentRequest>,
) -> Result<Json<EnvironmentResponse>, ApiError> {
    validation::validate_name(&req.name).map_err(ApiError::bad)?;
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let existing = fetch_env(&conn, &env_id).await?;
    crate::access::require_editor(&conn, &existing.project_id, &user.user_id).await?;
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let now = envoy_shared::types::sqlite_now();
    let n = conn
        .execute(
            "UPDATE environments SET name = ?1, description = ?2, updated_at = ?3 WHERE id = ?4 AND deleted_at IS NULL",
            libsql::params![req.name.clone(), desc.clone(), now.clone(), env_id.clone()],
        )
        .await
        .map_err(|e| {
            let msg = e.to_string();
            if msg.contains("UNIQUE") {
                ApiError::conflict("an environment with this name already exists")
            } else {
                ApiError::Internal(e.into())
            }
        })?;
    if n == 0 {
        return Err(ApiError::not_found("environment not found or access denied"));
    }
    Ok(Json(fetch_env(&conn, &env_id).await?))
}

pub async fn delete_environment(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(env_id): Path<String>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let existing = fetch_env(&conn, &env_id).await?;
    crate::access::require_editor(&conn, &existing.project_id, &user.user_id).await?;
    conn.execute(
        "UPDATE environments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?1",
        libsql::params![env_id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("Environment deleted successfully"))
}
