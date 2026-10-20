use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
};
use chrono::{TimeZone, Utc};
use envoy_shared::{
    requests::CreateInviteRequest,
    responses::{InviteListItem, InviteResponse},
    validation,
};
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::{error::ApiError, extract::AuthUser, state::AppState};

const INVITE_TTL_SECS: i64 = 7 * 24 * 3600;

fn new_token() -> String {
    format!("{}{}", Uuid::new_v4().simple(), Uuid::new_v4().simple())
}

fn token_hash(token: &str) -> String {
    format!("{:x}", Sha256::digest(token.as_bytes()))
}

/// Admin-only: create a single-use invite. The plaintext token is returned
/// once and never stored.
pub async fn create_invite(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Json(req): Json<CreateInviteRequest>,
) -> Result<(StatusCode, Json<InviteResponse>), ApiError> {
    validation::validate_email(&req.email).map_err(ApiError::bad)?;
    if !req.name.is_empty() && req.name.len() > 50 {
        return Err(ApiError::bad("name must be between 1 and 50 characters"));
    }
    let conn = state
        .db
        .conn()
        .await
        .map_err(ApiError::Internal)?;
    crate::access::require_admin(&conn, &user.user_id).await?;

    let mut dup = conn
        .query(
            "SELECT id FROM users WHERE email = ?1 AND deleted_at IS NULL",
            libsql::params![req.email.clone()],
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

    let token = new_token();
    let id = Uuid::new_v4().to_string();
    let now = Utc::now();
    let expires = now + chrono::Duration::seconds(INVITE_TTL_SECS);
    let fmt = "%Y-%m-%d %H:%M:%S";
    let expires_str = expires.format(fmt).to_string();
    let name: Option<String> = if req.name.is_empty() {
        None
    } else {
        Some(req.name.clone())
    };
    conn.execute(
        "INSERT INTO invites (id, email, name, token_hash, expires_at, invited_by, created_at) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7)",
        libsql::params![
            id.clone(),
            req.email.clone(),
            name,
            token_hash(&token),
            expires_str.clone(),
            user.user_id.clone(),
            now.format(fmt).to_string()
        ],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok((
        StatusCode::CREATED,
        Json(InviteResponse {
            invite_id: id,
            email: req.email,
            token,
            expires_at: expires_str,
        }),
    ))
}

pub async fn list_invites(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
) -> Result<Json<Vec<InviteListItem>>, ApiError> {
    let conn = state
        .db
        .conn()
        .await
        .map_err(ApiError::Internal)?;
    crate::access::require_admin(&conn, &user.user_id).await?;
    let now = Utc::now().format("%Y-%m-%d %H:%M:%S").to_string();
    let mut rows = conn
        .query(
            "SELECT id, email, name, expires_at, created_at FROM invites WHERE accepted_at IS NULL AND expires_at > ?1 ORDER BY created_at DESC",
            libsql::params![now],
        )
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let mut out = Vec::new();
    while let Some(row) = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
    {
        let name: Option<String> = row.get(2).map_err(|e| ApiError::Internal(e.into()))?;
        let expires: String = row.get(3).map_err(|e| ApiError::Internal(e.into()))?;
        let created: Option<String> = row.get(4).map_err(|e| ApiError::Internal(e.into()))?;
        // Normalize to RFC3339 for API consumers.
        let created_at = created
            .as_deref()
            .and_then(|s| chrono::NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S").ok())
            .map(|n| Utc.from_utc_datetime(&n).to_rfc3339())
            .unwrap_or_default();
        out.push(InviteListItem {
            invite_id: row.get(0).map_err(|e| ApiError::Internal(e.into()))?,
            email: row.get(1).map_err(|e| ApiError::Internal(e.into()))?,
            name,
            expires_at: expires,
            created_at,
        });
    }
    Ok(Json(out))
}

pub async fn revoke_invite(
    State(state): State<AppState>,
    AuthUser(user): AuthUser,
    Path(id): Path<String>,
) -> Result<Json<envoy_shared::responses::MessageBody>, ApiError> {
    let conn = state
        .db
        .conn()
        .await
        .map_err(ApiError::Internal)?;
    crate::access::require_admin(&conn, &user.user_id).await?;
    conn.execute(
        "DELETE FROM invites WHERE id = ?1",
        libsql::params![id.clone()],
    )
    .await
    .map_err(|e| ApiError::Internal(e.into()))?;
    Ok(crate::error::message("Invite revoked successfully"))
}
