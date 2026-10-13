use axum::{
    Json,
    extract::{Query, State},
    http::StatusCode,
};
use envoy_shared::{
    requests::{LoginRequest, RegisterRequest},
    responses::{AuthResponse, ProfileResponse, RegisteredUser, UserSearchResponse},
    types::Timestamp,
    validation,
};
use serde::Deserialize;
use uuid::Uuid;

use crate::{
    auth,
    error::ApiError,
    extract::AuthUser,
    rows::{get_string, ts_from_row_string},
    state::AppState,
};

pub async fn register(
    State(state): State<AppState>,
    Json(req): Json<RegisterRequest>,
) -> Result<(StatusCode, Json<AuthResponse>), ApiError> {
    validation::validate_register(&req.name, &req.email, &req.password)
        .map_err(ApiError::bad)?;
    let conn = state
        .db
        .conn()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;

    let mut rows = conn
        .query(
            "SELECT id FROM users WHERE email = ?1 AND deleted_at IS NULL",
            libsql::params![req.email.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    if rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .is_some()
    {
        return Err(ApiError::conflict("user with this email already exists"));
    }

    let hash = auth::hash_password(&req.password).map_err(|e| ApiError::Internal(e.into()))?;
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    conn.execute(
        "INSERT INTO users (id, name, email, password, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5, ?5)",
        libsql::params![id.clone(), req.name.clone(), req.email.clone(), hash, now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;

    let token = auth::generate_jwt(&id, &req.email, &state.jwt_secret).map_err(|e| ApiError::Internal(e.into()))?;
    Ok((
        StatusCode::CREATED,
        Json(AuthResponse {
            token,
            user: RegisteredUser {
                user_id: id,
                name: req.name,
                email: req.email,
                created_at: Timestamp(envoy_shared::types::parse_sqlite_time(&now)),
            },
        }),
    ))
}

pub async fn login(
    State(state): State<AppState>,
    Json(req): Json<LoginRequest>,
) -> Result<Json<AuthResponse>, ApiError> {
    validation::validate_login(&req.email, &req.password).map_err(ApiError::bad)?;
    let conn = state
        .db
        .conn()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut rows = conn
        .query(
            "SELECT id, name, email, password, created_at FROM users WHERE email = ?1 AND deleted_at IS NULL",
            libsql::params![req.email.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::unauthorized("invalid email or password"))?;
    let id = get_string(&row, 0).map_err(|e| ApiError::Internal(e.into()))?;
    let name = get_string(&row, 1).map_err(|e| ApiError::Internal(e.into()))?;
    let email = get_string(&row, 2).map_err(|e| ApiError::Internal(e.into()))?;
    let hash = get_string(&row, 3).map_err(|e| ApiError::Internal(e.into()))?;
    let created: Option<String> = row.get(4).map_err(|e| ApiError::Internal(e.into()))?;
    if !auth::verify_password(&req.password, &hash) {
        return Err(ApiError::unauthorized("invalid email or password"));
    }
    let token = auth::generate_jwt(&id, &email, &state.jwt_secret).map_err(|e| ApiError::Internal(e.into()))?;
    Ok(Json(AuthResponse {
        token,
        user: RegisteredUser {
            user_id: id,
            name,
            email,
            created_at: ts_from_row_string(created),
        },
    }))
}

pub async fn profile(AuthUser(claims): AuthUser) -> Json<ProfileResponse> {
    Json(ProfileResponse {
        user_id: claims.user_id,
        email: claims.email,
        issued_at: claims.iat,
        expires_at: claims.exp,
    })
}

#[derive(Debug, Deserialize)]
pub struct SearchQuery {
    pub email: Option<String>,
}

pub async fn search_users(
    State(state): State<AppState>,
    _user: AuthUser,
    Query(q): Query<SearchQuery>,
) -> Result<Json<Vec<UserSearchResponse>>, ApiError> {
    let needle = q
        .email
        .filter(|s| !s.is_empty())
        .ok_or_else(|| ApiError::bad("email query parameter is required"))?;
    let conn = state
        .db
        .conn()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let pattern = format!("%{needle}%");
    let mut rows = conn
        .query(
            "SELECT id, name, email FROM users WHERE email LIKE ?1 AND deleted_at IS NULL LIMIT 10",
            libsql::params![pattern],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows.next().await.map_err(|e| ApiError::Internal(e.into()))? {
        out.push(UserSearchResponse {
            user_id: get_string(&row, 0).map_err(|e| ApiError::Internal(e.into()))?,
            name: get_string(&row, 1).map_err(|e| ApiError::Internal(e.into()))?,
            email: get_string(&row, 2).map_err(|e| ApiError::Internal(e.into()))?,
        });
    }
    Ok(Json(out))
}
