use std::{env, fs, path::PathBuf};

use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct Config {
    pub jwt_secret: String,
    pub database_url: String,
    pub auth_token: String,
    pub addr: String,
}

fn state_dir() -> PathBuf {
    if let Ok(d) = env::var("ENVOY_STATE_DIR") {
        if !d.is_empty() {
            return PathBuf::from(d);
        }
    }
    if let Ok(url) = env::var("TURSO_DATABASE_URL") {
        let path = url.strip_prefix("file:").unwrap_or(&url);
        if !path.starts_with("http://")
            && !path.starts_with("https://")
            && !path.starts_with("libsql://")
            && !path.is_empty()
        {
            if let Some(parent) = PathBuf::from(path).parent() {
                if !parent.as_os_str().is_empty() {
                    return parent.to_path_buf();
                }
            }
        }
    }
    PathBuf::from("./data")
}

fn resolve_jwt_secret() -> anyhow::Result<String> {
    if let Ok(s) = env::var("JWT_SECRET") {
        if !s.is_empty() {
            return Ok(s);
        }
    }
    let dir = state_dir();
    let path = dir.join(".jwt_secret");
    if let Ok(existing) = fs::read_to_string(&path) {
        let trimmed = existing.trim().to_string();
        if !trimmed.is_empty() {
            return Ok(trimmed);
        }
    }
    // Fresh self-hosted instance: generate and persist so restarts
    // don't invalidate all sessions.
    let secret = format!("{}{}", Uuid::new_v4().simple(), Uuid::new_v4().simple());
    fs::create_dir_all(&dir)?;
    fs::write(&path, &secret)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o600))?;
    }
    Ok(secret)
}

impl Config {
    pub fn load() -> anyhow::Result<Self> {
        let _ = dotenvy::dotenv();
        let jwt_secret = resolve_jwt_secret()?;
        let database_url = env::var("TURSO_DATABASE_URL")
            .unwrap_or_else(|_| "./data/envoy.db".to_string());
        let auth_token = env::var("TURSO_AUTH_TOKEN").unwrap_or_default();
        let addr = env::var("ADDR").unwrap_or_else(|_| "0.0.0.0:8080".to_string());
        let is_remote = database_url.starts_with("http://")
            || database_url.starts_with("https://")
            || database_url.starts_with("libsql://");
        if jwt_secret.is_empty() || database_url.is_empty() {
            anyhow::bail!("JWT_SECRET and database URL must both be set");
        }
        if is_remote && auth_token.is_empty() {
            anyhow::bail!("TURSO_AUTH_TOKEN is required for remote databases");
        }
        Ok(Self {
            jwt_secret,
            database_url,
            auth_token,
            addr,
        })
    }
}
