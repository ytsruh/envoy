use axum::{
    extract::{Query, State},
    http::StatusCode,
    Json,
};
use envoy_shared::{
    requests::{AcceptInviteRequest, LoginRequest, SetupRequest},
    responses::{AuthResponse, ProfileResponse, RegisteredUser, UserSearchResponse},
    types::Timestamp,
    validation,
};
use serde::Deserialize;
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::{
    auth,
    error::ApiError,
    extract::AuthUser,
    rows::{get_string, ts_from_row_string},
    state::AppState,
};

fn invite_token_hash(token: &str) -> String {
    format!("{:x}", Sha256::digest(token.as_bytes()))
}

async fn user_count(conn: &libsql::Connection) -> Result<i64, ApiError> {
    let mut rows = conn
        .query(
            "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL",
            libsql::params![],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::Internal(anyhow::anyhow!("count query returned no rows")))?;
    row.get(0).map_err(|e| ApiError::Internal(e.into()))
}

/// One-time bootstrap: only works on a fresh server with no users.
/// The first user becomes instance admin.
pub async fn setup(
    State(state): State<AppState>,
    Json(req): Json<SetupRequest>,
) -> Result<(StatusCode, Json<AuthResponse>), ApiError> {
    validation::validate_register(&req.name, &req.email, &req.password).map_err(ApiError::bad)?;
    let conn = state
        .db
        .conn()
        .await
        .map_err(ApiError::Internal)?;

    if user_count(&conn).await? > 0 {
        return Err(ApiError::forbidden(
            "server is already set up - ask an admin for an invite",
        ));
    }

    let mut existing = conn
        .query(
            "SELECT id FROM users WHERE email = ?1 AND deleted_at IS NULL",
            libsql::params![req.email.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    if existing
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .is_some()
    {
        return Err(ApiError::conflict("user with this email already exists"));
    }

    let hash = auth::hash_password(&req.password).map_err(ApiError::Internal)?;
    let id = Uuid::new_v4().to_string();
    let now = envoy_shared::types::sqlite_now();
    conn.execute(
        "INSERT INTO users (id, name, email, password, is_admin, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, 1, ?5, ?5)",
        libsql::params![id.clone(), req.name.clone(), req.email.clone(), hash, now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;

    let token = auth::generate_jwt(&id, &req.email, &state.jwt_secret)
        .map_err(ApiError::Internal)?;
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

/// Accept an admin-issued invite: invitee chooses their own name/password.
pub async fn accept_invite(
    State(state): State<AppState>,
    Json(req): Json<AcceptInviteRequest>,
) -> Result<(StatusCode, Json<AuthResponse>), ApiError> {
    if req.token.is_empty() {
        return Err(ApiError::bad("invite token is required"));
    }
    let conn = state
        .db
        .conn()
        .await
        .map_err(ApiError::Internal)?;

    let hash = invite_token_hash(&req.token);
    let mut rows = conn
        .query(
            "SELECT id, email, name, expires_at, accepted_at FROM invites WHERE token_hash = ?1",
            libsql::params![hash],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::unauthorized("invalid or expired invite token"))?;
    let invite_id: String = row.get(0).map_err(|e| ApiError::Internal(e.into()))?;
    let email: String = row.get(1).map_err(|e| ApiError::Internal(e.into()))?;
    let invited_name: Option<String> = row.get(2).map_err(|e| ApiError::Internal(e.into()))?;
    let expires_at: String = row.get(3).map_err(|e| ApiError::Internal(e.into()))?;
    let accepted_at: Option<String> = row.get(4).map_err(|e| ApiError::Internal(e.into()))?;

    if accepted_at.is_some() {
        return Err(ApiError::unauthorized("invite has already been used"));
    }
    let now = envoy_shared::types::sqlite_now();
    if expires_at < now {
        return Err(ApiError::unauthorized("invite has expired"));
    }

    let name = if req.name.trim().is_empty() {
        invited_name.unwrap_or_default()
    } else {
        req.name.clone()
    };
    validation::validate_register(&name, &email, &req.password).map_err(ApiError::bad)?;

    let mut dup = conn
        .query(
            "SELECT id FROM users WHERE email = ?1 AND deleted_at IS NULL",
            libsql::params![email.clone()],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    if dup
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .is_some()
    {
        return Err(ApiError::conflict("user with this email already exists"));
    }

    let pw_hash = auth::hash_password(&req.password).map_err(ApiError::Internal)?;
    let id = Uuid::new_v4().to_string();
    conn.execute(
        "INSERT INTO users (id, name, email, password, is_admin, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, 0, ?5, ?5)",
        libsql::params![id.clone(), name.clone(), email.clone(), pw_hash, now.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    conn.execute(
        "UPDATE invites SET accepted_at = ?1 WHERE id = ?2",
        libsql::params![now.clone(), invite_id],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;

    let token = auth::generate_jwt(&id, &email, &state.jwt_secret)
        .map_err(ApiError::Internal)?;
    Ok((
        StatusCode::CREATED,
        Json(AuthResponse {
            token,
            user: RegisteredUser {
                user_id: id,
                name,
                email,
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
        .map_err(ApiError::Internal)?;
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
    let id = get_string(&row, 0).map_err(ApiError::Internal)?;
    let name = get_string(&row, 1).map_err(ApiError::Internal)?;
    let email = get_string(&row, 2).map_err(ApiError::Internal)?;
    let hash = get_string(&row, 3).map_err(ApiError::Internal)?;
    let created: Option<String> = row.get(4).map_err(|e| ApiError::Internal(e.into()))?;
    if !auth::verify_password(&req.password, &hash) {
        return Err(ApiError::unauthorized("invalid email or password"));
    }
    let token = auth::generate_jwt(&id, &email, &state.jwt_secret)
        .map_err(ApiError::Internal)?;
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
        .map_err(ApiError::Internal)?;
    let pattern = format!("%{needle}%");
    let mut rows = conn
        .query(
            "SELECT id, name, email FROM users WHERE email LIKE ?1 AND deleted_at IS NULL LIMIT 10",
            libsql::params![pattern],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
    {
        out.push(UserSearchResponse {
            user_id: get_string(&row, 0).map_err(ApiError::Internal)?,
            name: get_string(&row, 1).map_err(ApiError::Internal)?,
            email: get_string(&row, 2).map_err(ApiError::Internal)?,
        });
    }
    Ok(Json(out))
}
