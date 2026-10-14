use axum::{extract::FromRequestParts, http::request::Parts};

use crate::{auth::Claims, error::ApiError, state::AppState};

#[derive(Debug, Clone)]
pub struct AuthUser(pub Claims);

impl FromRequestParts<AppState> for AuthUser {
    type Rejection = ApiError;

    async fn from_request_parts(
        parts: &mut Parts,
        state: &AppState,
    ) -> Result<Self, Self::Rejection> {
        let header = parts
            .headers
            .get(axum::http::header::AUTHORIZATION)
            .and_then(|v| v.to_str().ok())
            .ok_or_else(|| ApiError::unauthorized("Missing authorization header"))?;
        let token = header
            .strip_prefix("Bearer ")
            .ok_or_else(|| {
                ApiError::unauthorized(
                    "Invalid authorization header format. Expected: Bearer <token>",
                )
            })?;
        match crate::auth::validate_jwt(token, &state.jwt_secret) {
            Ok(claims) => Ok(AuthUser(claims)),
            Err(crate::auth::JwtError::Expired) => Err(ApiError::unauthorized("Token has expired")),
            Err(crate::auth::JwtError::Invalid) => Err(ApiError::unauthorized("Invalid token")),
        }
    }
}
