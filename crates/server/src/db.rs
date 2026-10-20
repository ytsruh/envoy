use libsql::{Builder, Connection, Database};

const SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  password TEXT NOT NULL,
  is_admin INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS invites (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL,
  name TEXT,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  accepted_at TIMESTAMP DEFAULT NULL,
  invited_by TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  git_repo TEXT,
  owner_id TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE(owner_id, git_repo)
);
CREATE TABLE IF NOT EXISTS project_users (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'viewer',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE(project_id, user_id)
);
CREATE TABLE IF NOT EXISTS environments (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(project_id, name)
);
CREATE TABLE IF NOT EXISTS environment_variables (
  id TEXT PRIMARY KEY,
  environment_id TEXT NOT NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  description TEXT,
  file_path TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vars_env_key ON environment_variables(environment_id, key);"#;

#[derive(Clone)]
pub struct Db {
    db: std::sync::Arc<Database>,
}

impl Db {
    pub async fn connect_remote(url: &str, token: &str) -> anyhow::Result<Self> {
        let db = Builder::new_remote(url.to_string(), token.to_string())
            .build()
            .await?;
        Ok(Self {
            db: std::sync::Arc::new(db),
        })
    }

    pub async fn connect_local(path: &str) -> anyhow::Result<Self> {
        let db = Builder::new_local(path).build().await?;
        Ok(Self {
            db: std::sync::Arc::new(db),
        })
    }

    pub async fn connect_auto(url: &str, token: &str) -> anyhow::Result<Self> {
        if url.starts_with("http://") || url.starts_with("https://") || url.starts_with("libsql://")
        {
            Self::connect_remote(url, token).await
        } else {
            let path = url.strip_prefix("file:").unwrap_or(url);
            Self::connect_local(path).await
        }
    }

    pub async fn conn(&self) -> anyhow::Result<Connection> {
        Ok(self.db.connect()?)
    }

    pub async fn migrate(&self) -> anyhow::Result<()> {
        let conn = self.conn().await?;
        conn.execute_batch(SCHEMA).await?;
        // Backfill for databases created before is_admin / invites existed.
        // Ignore "duplicate column" when the column is already present.
        if let Err(e) = conn
            .execute(
                "ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0",
                libsql::params![],
            )
            .await
        {
            let msg = e.to_string();
            if !msg.contains("duplicate column") && !msg.contains("already exists") {
                return Err(e.into());
            }
        }
        // Backfill for databases created before file_path existed: each
        // variable keeps its value, legacy rows read as '' (single-file).
        if let Err(e) = conn
            .execute(
                "ALTER TABLE environment_variables ADD COLUMN file_path TEXT NOT NULL DEFAULT ''",
                libsql::params![],
            )
            .await
        {
            let msg = e.to_string();
            if !msg.contains("duplicate column") && !msg.contains("already exists") {
                return Err(e.into());
            }
        }
        // Same key in two files of one env must fail instead of forking.
        // Fails loudly here if an old DB already has duplicates.
        conn.execute(
            "CREATE UNIQUE INDEX IF NOT EXISTS idx_vars_env_key ON environment_variables(environment_id, key)",
            libsql::params![],
        )
        .await?;
        Ok(())
    }
}
