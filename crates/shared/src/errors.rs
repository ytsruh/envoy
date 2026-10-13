use thiserror::Error;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum ApiError {
    #[error("bad request: {0}")]
    BadRequest(String),
    #[error("unauthorized: {0}")]
    Unauthorized(String),
    #[error("access denied")]
    AccessDenied,
    #[error("not a project member")]
    NotMember,
    #[error("expired token")]
    ExpiredToken,
    #[error("invalid token")]
    InvalidToken,
}
