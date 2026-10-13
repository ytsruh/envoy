use std::{fs, path::PathBuf};

use serde::{Deserialize, Serialize};

const DEFAULT_SERVER_URL: &str = "http://localhost:8080";

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub server_url: String,
    #[serde(default)]
    pub token: String,
}

fn config_path() -> anyhow::Result<PathBuf> {
    let home = dirs::home_dir().ok_or_else(|| anyhow::anyhow!("cannot find home directory"))?;
    Ok(home.join(".envoy").join("config.json"))
}

pub fn load() -> Config {
    let path = match config_path() {
        Ok(p) => p,
        Err(_) => return Config::default(),
    };
    let data = fs::read_to_string(path).unwrap_or_default();
    serde_json::from_str(&data).unwrap_or_default()
}

pub fn save(cfg: &Config) -> anyhow::Result<()> {
    let path = config_path()?;
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(parent, fs::Permissions::from_mode(0o700))?;
        }
    }
    let data = serde_json::to_string_pretty(cfg)?;
    fs::write(&path, data)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o600))?;
    }
    Ok(())
}

pub fn get_token() -> String {
    load().token
}

pub fn set_token(token: &str) -> anyhow::Result<()> {
    let mut cfg = load();
    cfg.token = token.to_string();
    save(&cfg)
}

pub fn clear_token() -> anyhow::Result<()> {
    set_token("")
}

/// Point the CLI at a (self-hosted) server. Clears any stale token since
/// tokens are instance-scoped and must not leak across servers.
pub fn connect(server_url: &str) -> anyhow::Result<()> {
    let url = server_url.trim().trim_end_matches('/').to_string();
    if url.is_empty() {
        anyhow::bail!("server URL is required");
    }
    let mut cfg = load();
    cfg.server_url = url;
    cfg.token = String::new();
    save(&cfg)
}

pub fn server_url() -> String {
    let cfg = load();
    if !cfg.server_url.is_empty() {
        return cfg.server_url;
    }
    if let Ok(url) = std::env::var("ENVOY_SERVER_URL") {
        if !url.is_empty() {
            return url;
        }
    }
    DEFAULT_SERVER_URL.to_string()
}
