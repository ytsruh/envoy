use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use envoy_shared::{
    requests::{CreateVariableRequest, UpdateVariableRequest},
    responses::VariableResponse,
    validation,
};
use uuid::Uuid;

use crate::{
    error::{ApiError, message},
    extract::AuthUser,
    rows::{get_opt_string, get_string, ts_from_row_string},
    state::AppState,
};

fn row_to_var(row: &libsql::Row) -> anyhow::Result<VariableResponse> {
    Ok(VariableResponse {
        id: get_string(row, 0)?,
        environment_id: get_string(row, 1)?,
        key: get_string(row, 2)?,
        value: get_string(row, 3)?,
        description: get_opt_string(row, 4)?,
        created_at: ts_from_row_string(row.get(5)?),
        updated_at: ts_from_row_string(row.get(6)?),
    })
}

const SELECT: &str = "SELECT id, environment_id, key, value, description, created_at, updated_at FROM environment_variables";

pub async fn create_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, env_id)): Path<(String, String)>,
    Json(req): Json<CreateVariableRequest>,
) -> Result<(StatusCode, Json<VariableResponse>), ApiError> {
    if req.key.is_empty() {
        return Err(ApiError::bad("key is required"));
    }
    if req.value.is_empty() {
        return Err(ApiError::bad("value is required"));
    }
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    conn.execute(
        "INSERT INTO environment_variables (id, environment_id, key, value, description, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?6)",
        libsql::params![id.clone(), env_id.clone(), req.key.clone(), req.value.clone(), desc.clone(), now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    let ts = envoy_shared::types::parse_sqlite_time(&now).into();
    Ok((
        StatusCode::CREATED,
        Json(VariableResponse {
            id,
            environment_id: env_id,
            key: req.key,
            value: req.value,
            description: desc,
            created_at: ts,
            updated_at: ts,
        }),
    ))
}

pub async fn list_variables(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, env_id)): Path<(String, String)>,
) -> Result<Json<Vec<VariableResponse>>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_viewer(&conn, &project_id, &user.user_id).await?;
    let mut rows = conn
        .query(
            &format!("{SELECT} WHERE environment_id = ?1 ORDER BY key"),
            libsql::params![env_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows.next().await.map_err(|e| ApiError::Internal(e.into()))? {
        out.push(row_to_var(&row).map_err(|e| ApiError::Internal(e.into()))?);
    }
    Ok(Json(out))
}

pub async fn get_variable(
    State(state): State<AppState>,
    _user: AuthUser,
    Path((_, _, var_id)): Path<(String, String, String)>,
) -> Result<Json<VariableResponse>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    let mut rows = conn
        .query(
            &format!("{SELECT} WHERE id = ?1"),
            libsql::params![var_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("environment variable not found or access denied"))?;
    Ok(Json(row_to_var(&row).map_err(|e| ApiError::Internal(e.into()))?))
}

pub async fn update_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, env_id, var_id)): Path<(String, String, String)>,
    Json(req): Json<UpdateVariableRequest>,
) -> Result<Json<VariableResponse>, ApiError> {
    if req.key.is_empty() {
        return Err(ApiError::bad("key is required"));
    }
    if req.value.is_empty() {
        return Err(ApiError::bad("value is required"));
    }
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let now = envoy_shared::types::sqlite_now();
    let n = conn
        .execute(
            "UPDATE environment_variables SET key = ?1, value = ?2, description = ?3, updated_at = ?4 WHERE id = ?5 AND environment_id = ?6",
            libsql::params![req.key.clone(), req.value.clone(), desc.clone(), now.clone(), var_id.clone(), env_id.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    if n == 0 {
        return Err(ApiError::not_found("environment variable not found or access denied"));
    }
    get_variable(State(state), AuthUser(user), Path((project_id, env_id, var_id))).await
}

pub async fn delete_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path((project_id, _env_id, var_id)): Path<(String, String, String)>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    let conn = state.db.conn().await.map_err(|e| ApiError::Internal(e.into()))?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    conn.execute(
        "DELETE FROM environment_variables WHERE id = ?1",
        libsql::params![var_id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("Environment variable deleted successfully"))
}
