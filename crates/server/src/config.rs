use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub jwt_secret: String,
    pub database_url: String,
    pub auth_token: String,
    pub addr: String,
}

impl Config {
    pub fn load() -> anyhow::Result<Self> {
        let _ = dotenvy::dotenv();
        let jwt_secret =
            env::var("JWT_SECRET").map_err(|_| anyhow::anyhow!("missing JWT_SECRET"))?;
        let database_url = env::var("TURSO_DATABASE_URL")
            .or_else(|_| env::var("DB_URL"))
            .map_err(|_| anyhow::anyhow!("missing TURSO_DATABASE_URL (or DB_URL)"))?;
        let auth_token = env::var("TURSO_AUTH_TOKEN")
            .or_else(|_| env::var("DB_TOKEN"))
            .unwrap_or_default();
        let addr = env::var("ADDR").unwrap_or_else(|_| "0.0.0.0:8080".to_string());
        let is_remote = database_url.starts_with("http://")
            || database_url.starts_with("https://")
            || database_url.starts_with("libsql://");
        if jwt_secret.is_empty() || database_url.is_empty() {
            anyhow::bail!("JWT_SECRET and database URL must both be set");
        }
        if is_remote && auth_token.is_empty() {
            anyhow::bail!("TURSO_AUTH_TOKEN (or DB_TOKEN) is required for remote databases");
        }
        Ok(Self {
            jwt_secret,
            database_url,
            auth_token,
            addr,
        })
    }
}
