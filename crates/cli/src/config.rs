use std::{fs, path::PathBuf};

use serde::{Deserialize, Serialize};

const DEFAULT_SERVER_URL: &str = "http://localhost:8080";

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct LinkEntry {
    #[serde(default)]
    pub project_id: String,
    #[serde(default)]
    pub project_name: String,
    #[serde(default)]
    pub env_name: String,
    #[serde(default)]
    pub git_repo: String,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub server_url: String,
    #[serde(default)]
    pub token: String,
    #[serde(default)]
    pub links: std::collections::HashMap<String, LinkEntry>,
}

fn config_path() -> anyhow::Result<PathBuf> {
    let home = dirs::home_dir().ok_or_else(|| anyhow::anyhow!("cannot find home directory"))?;
    Ok(home.join(".envoy").join("config.json"))
}

/// Load config, distinguishing "no config yet" (Ok(default)) from
/// "config exists but is unreadable/corrupt" (Err with path + cause).
pub fn load() -> anyhow::Result<Config> {
    let path = config_path()?;
    let data = match fs::read_to_string(&path) {
        Ok(d) => d,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Config::default()),
        Err(e) => {
            anyhow::bail!("cannot read config file {}: {e}", path.display());
        }
    };
    if data.trim().is_empty() {
        return Ok(Config::default());
    }
    serde_json::from_str(&data).map_err(|e| {
        anyhow::anyhow!(
            "corrupt config file {}: {e}. Fix it or remove it to start fresh",
            path.display()
        )
    })
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

pub fn get_token() -> anyhow::Result<String> {
    Ok(load()?.token)
}

pub fn set_token(token: &str) -> anyhow::Result<()> {
    let mut cfg = load()?;
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
    let mut cfg = load()?;
    cfg.server_url = url;
    cfg.token = String::new();
    save(&cfg)
}

pub fn server_url() -> anyhow::Result<String> {
    let cfg = load()?;
    if !cfg.server_url.is_empty() {
        return Ok(cfg.server_url);
    }
    if let Ok(url) = std::env::var("ENVOY_SERVER_URL") {
        if !url.is_empty() {
            return Ok(url);
        }
    }
    Ok(DEFAULT_SERVER_URL.to_string())
}

fn cwd_key() -> String {
    let cwd = std::env::current_dir().unwrap_or_else(|_| PathBuf::from("."));
    let canon = std::fs::canonicalize(&cwd).unwrap_or(cwd);
    canon.to_string_lossy().to_string()
}

/// Find the link for cwd, walking up parents so subdirs inherit.
pub fn link_for_cwd() -> anyhow::Result<Option<LinkEntry>> {
    let cfg = load()?;
    if cfg.links.is_empty() {
        return Ok(None);
    }
    let cwd = std::env::current_dir().unwrap_or_else(|_| PathBuf::from("."));
    let canon = std::fs::canonicalize(&cwd).unwrap_or(cwd);
    let mut cur: Option<&std::path::Path> = Some(&canon);
    while let Some(p) = cur {
        if let Some(e) = cfg.links.get(&p.to_string_lossy().to_string()) {
            return Ok(Some(e.clone()));
        }
        cur = p.parent();
    }
    Ok(None)
}

pub fn set_link(entry: &LinkEntry) -> anyhow::Result<String> {
    let mut cfg = load()?;
    let key = cwd_key();
    cfg.links.insert(key.clone(), entry.clone());
    save(&cfg)?;
    Ok(key)
}

pub fn remove_link() -> anyhow::Result<bool> {
    let mut cfg = load()?;
    let key = cwd_key();
    let removed = cfg.links.remove(&key).is_some();
    if removed {
        save(&cfg)?;
    }
    Ok(removed)
}

pub fn list_links() -> anyhow::Result<Vec<(String, LinkEntry)>> {
    let mut v: Vec<(String, LinkEntry)> = load()?.links.into_iter().collect();
    v.sort_by(|a, b| a.0.cmp(&b.0));
    Ok(v)
}
