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
        file_path: row
            .get::<Option<String>>(5)
            .unwrap_or(None)
            .unwrap_or_default(),
        created_at: ts_from_row_string(row.get(6)?),
        updated_at: ts_from_row_string(row.get(7)?),
    })
}

const SELECT: &str = "SELECT id, environment_id, key, value, description, file_path, created_at, updated_at FROM environment_variables";

async fn project_id_for_env(conn: &libsql::Connection, env_id: &str) -> Result<String, ApiError> {
    let mut rows = conn
        .query(
            "SELECT project_id FROM environments WHERE id = ?1 AND deleted_at IS NULL",
            libsql::params![env_id.to_string()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("environment not found or access denied"))?;
    row.get::<String>(0).map_err(|e| ApiError::Internal(e.into()))
}

async fn fetch_var_with_project(
    conn: &libsql::Connection,
    var_id: &str,
) -> Result<(VariableResponse, String), ApiError> {
    let mut rows = conn
        .query(
            "SELECT ev.id, ev.environment_id, ev.key, ev.value, ev.description, ev.file_path, ev.created_at, ev.updated_at, e.project_id FROM environment_variables ev JOIN environments e ON e.id = ev.environment_id WHERE ev.id = ?1 AND e.deleted_at IS NULL",
            libsql::params![var_id.to_string()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::not_found("environment variable not found or access denied"))?;
    let project_id: String = row.get(8).map_err(|e| ApiError::Internal(e.into()))?;
    let var = VariableResponse {
        id: get_string(&row, 0).map_err(ApiError::Internal)?,
        environment_id: get_string(&row, 1).map_err(ApiError::Internal)?,
        key: get_string(&row, 2).map_err(ApiError::Internal)?,
        value: get_string(&row, 3).map_err(ApiError::Internal)?,
        description: get_opt_string(&row, 4).map_err(ApiError::Internal)?,
        file_path: row
            .get::<Option<String>>(5)
            .map_err(|e| ApiError::Internal(e.into()))?
            .unwrap_or_default(),
        created_at: ts_from_row_string(row.get(6).map_err(|e| ApiError::Internal(e.into()))?),
        updated_at: ts_from_row_string(row.get(7).map_err(|e| ApiError::Internal(e.into()))?),
    };
    Ok((var, project_id))
}

pub async fn create_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(env_id): Path<String>,
    Json(req): Json<CreateVariableRequest>,
) -> Result<(StatusCode, Json<VariableResponse>), ApiError> {
    if req.key.is_empty() {
        return Err(ApiError::bad("key is required"));
    }
    if req.value.is_empty() {
        return Err(ApiError::bad("value is required"));
    }
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    validation::validate_file_path(&req.file_path).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let project_id = project_id_for_env(&conn, &env_id).await?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let file_path = req.file_path.trim().to_string();
    if let Err(e) = conn.execute(
        "INSERT INTO environment_variables (id, environment_id, key, value, description, file_path, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?7)",
        libsql::params![id.clone(), env_id.clone(), req.key.clone(), req.value.clone(), desc.clone(), file_path.clone(), now.clone()],
    )
    .await
    {
        let msg = e.to_string();
        if msg.contains("UNIQUE") {
            return Err(ApiError::conflict(format!(
                "variable '{}' already exists in this environment",
                req.key
            )));
        }
        return Err(ApiError::Internal(e.into()));
    }
    let ts = envoy_shared::types::parse_sqlite_time(&now).into();
    Ok((
        StatusCode::CREATED,
        Json(VariableResponse {
            id,
            environment_id: env_id,
            key: req.key,
            value: req.value,
            description: desc,
            file_path,
            created_at: ts,
            updated_at: ts,
        }),
    ))
}

pub async fn list_variables(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(env_id): Path<String>,
) -> Result<Json<Vec<VariableResponse>>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let project_id = project_id_for_env(&conn, &env_id).await?;
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
        out.push(row_to_var(&row).map_err(ApiError::Internal)?);
    }
    Ok(Json(out))
}

pub async fn get_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(var_id): Path<String>,
) -> Result<Json<VariableResponse>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let (var, project_id) = fetch_var_with_project(&conn, &var_id).await?;
    crate::access::require_viewer(&conn, &project_id, &user.user_id).await?;
    Ok(Json(var))
}

pub async fn update_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(var_id): Path<String>,
    Json(req): Json<UpdateVariableRequest>,
) -> Result<Json<VariableResponse>, ApiError> {
    if req.key.is_empty() {
        return Err(ApiError::bad("key is required"));
    }
    if req.value.is_empty() {
        return Err(ApiError::bad("value is required"));
    }
    validation::validate_description(&req.description).map_err(ApiError::bad)?;
    validation::validate_file_path(&req.file_path).map_err(ApiError::bad)?;
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let (existing, project_id) = fetch_var_with_project(&conn, &var_id).await?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    let desc: Option<String> = if req.description.is_empty() {
        None
    } else {
        Some(req.description.clone())
    };
    let file_path = req.file_path.trim().to_string();
    let now = envoy_shared::types::sqlite_now();
    let n = conn
        .execute(
            "UPDATE environment_variables SET key = ?1, value = ?2, description = ?3, file_path = ?4, updated_at = ?5 WHERE id = ?6 AND environment_id = ?7",
            libsql::params![req.key.clone(), req.value.clone(), desc.clone(), file_path.clone(), now.clone(), var_id.clone(), existing.environment_id.clone()],
        )
        .await
        .map_err(|e| {
            let msg = e.to_string();
            if msg.contains("UNIQUE") {
                ApiError::conflict(format!(
                    "variable '{}' already exists in this environment",
                    req.key
                ))
            } else {
                ApiError::Internal(e.into())
            }
        })?;
    if n == 0 {
        return Err(ApiError::not_found("environment variable not found or access denied"));
    }
    let (updated, _) = fetch_var_with_project(&conn, &var_id).await?;
    Ok(Json(updated))
}

pub async fn delete_variable(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(var_id): Path<String>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    let conn = state.db.conn().await.map_err(ApiError::Internal)?;
    let (_, project_id) = fetch_var_with_project(&conn, &var_id).await?;
    crate::access::require_editor(&conn, &project_id, &user.user_id).await?;
    conn.execute(
        "DELETE FROM environment_variables WHERE id = ?1",
        libsql::params![var_id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(message("Environment variable deleted successfully"))
}
