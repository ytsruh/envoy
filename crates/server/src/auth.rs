use chrono::Utc;
use jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation, decode, encode};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub user_id: String,
    pub email: String,
    pub exp: i64,
    pub iat: i64,
}

const EXPIRY_SECS: i64 = 7 * 24 * 3600;

pub fn generate_jwt(user_id: &str, email: &str, secret: &str) -> anyhow::Result<String> {
    let now = Utc::now().timestamp();
    let claims = Claims {
        user_id: user_id.to_string(),
        email: email.to_string(),
        iat: now,
        exp: now + EXPIRY_SECS,
    };
    let token = encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(secret.as_bytes()),
    )?;
    Ok(token)
}

#[derive(Debug, thiserror::Error)]
pub enum JwtError {
    #[error("expired token")]
    Expired,
    #[error("invalid token")]
    Invalid,
}

pub fn validate_jwt(token: &str, secret: &str) -> Result<Claims, JwtError> {
    let mut validation = Validation::default();
    validation.validate_exp = false; // manual check to control message
    let data = decode::<Claims>(
        token,
        &DecodingKey::from_secret(secret.as_bytes()),
        &validation,
    )
    .map_err(|e| {
        use jsonwebtoken::errors::ErrorKind;
        match e.kind() {
            ErrorKind::ExpiredSignature => JwtError::Expired,
            _ => JwtError::Invalid,
        }
    })?;
    if Utc::now().timestamp() > data.claims.exp {
        return Err(JwtError::Expired);
    }
    Ok(data.claims)
}

pub fn hash_password(password: &str) -> anyhow::Result<String> {
    Ok(bcrypt::hash(password, 10)?)
}

pub fn verify_password(password: &str, hash: &str) -> bool {
    bcrypt::verify(password, hash).unwrap_or(false)
}
