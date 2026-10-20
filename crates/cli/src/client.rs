use serde::de::DeserializeOwned;
use serde_json::Value;

use crate::config;

#[derive(Debug, thiserror::Error)]
pub enum ClientError {
    #[error("no token - please login first using 'envoy manage login'")]
    NoToken,
    #[error("session expired - please login again using 'envoy manage login'")]
    Expired,
    #[error("server error: {0}")]
    Server(String),
    #[error(transparent)]
    Transport(#[from] reqwest::Error),
    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

pub struct Client {
    base: String,
    token: String,
    http: reqwest::Client,
}

fn edit_distance(a: &str, b: &str) -> usize {
    let a: Vec<char> = a.chars().collect();
    let b: Vec<char> = b.chars().collect();
    let mut prev: Vec<usize> = (0..=b.len()).collect();
    for (i, ca) in a.iter().enumerate() {
        let mut cur = vec![i + 1];
        for (j, cb) in b.iter().enumerate() {
            let cost = if ca.eq_ignore_ascii_case(cb) {
                0
            } else {
                1
            };
            cur.push((prev[j] + cost).min(cur[j] + 1).min(prev[j + 1] + 1));
        }
        prev = cur;
    }
    prev[b.len()]
}

fn closest_name(input: &str, names: &[String]) -> Option<String> {
    let mut best: Option<(usize, &String)> = None;
    for n in names {
        let d = edit_distance(input, n);
        if d <= 2 && best.is_none_or(|(bd, _)| d < bd) {
            best = Some((d, n));
        }
    }
    best.map(|(_, n)| n.clone())
}

impl Client {
    pub fn new() -> anyhow::Result<Self> {
        Ok(Self {
            base: config::server_url()?,
            token: config::get_token()?,
            http: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(30))
                .build()?,
        })
    }

    pub fn require_token() -> Result<Self, ClientError> {
        let c = Self::new().map_err(ClientError::Other)?;
        if c.token.is_empty() {
            return Err(ClientError::NoToken);
        }
        Ok(c)
    }

    async fn request<T: DeserializeOwned>(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<Value>,
        auth: bool,
    ) -> Result<T, ClientError> {
        let resp = self.raw(method, path, body, auth).await?;
        let status = resp.status();
        if !status.is_success() {
            let err_msg = resp
                .json::<serde_json::Value>()
                .await
                .ok()
                .and_then(|v| {
                    v.get("error")
                        .and_then(|e| e.as_str())
                        .map(|s| s.to_string())
                })
                .unwrap_or_else(|| format!("HTTP {status}"));
            if status == reqwest::StatusCode::UNAUTHORIZED {
                // Token is unusable (expired, invalid, revoked, or server rotated
                // its secret). Clear it so the next run reports NoToken instead of
                // retrying a dead credential, then tell the user to log in again.
                if let Err(e) = config::clear_token() {
                    eprintln!("Warning: failed to clear expired token: {e:?}");
                }
                if err_msg.to_ascii_lowercase().contains("expir") {
                    return Err(ClientError::Expired);
                }
                return Err(ClientError::Server(format!(
                    "{err_msg} - please login again using 'envoy manage login'"
                )));
            }
            return Err(ClientError::Server(err_msg));
        }
        // Allow empty body for message endpoints when T is Value
        let parsed = resp.json::<T>().await?;
        Ok(parsed)
    }

    async fn raw(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<Value>,
        auth: bool,
    ) -> Result<reqwest::Response, ClientError> {
        let mut req = self.http.request(method, format!("{}{path}", self.base));
        if auth && !self.token.is_empty() {
            req = req.bearer_auth(&self.token);
        }
        req = req.header("Content-Type", "application/json");
        if let Some(b) = body {
            req = req.json(&b);
        }
        Ok(req.send().await?)
    }

    async fn get<T: DeserializeOwned>(&self, path: &str, auth: bool) -> Result<T, ClientError> {
        self.request(reqwest::Method::GET, path, None, auth).await
    }

    // --- setup / session ---
    pub async fn setup_init(
        &self,
        name: &str,
        email: &str,
        password: &str,
    ) -> Result<envoy_shared::responses::AuthResponse, ClientError> {
        let body = serde_json::json!({"name": name, "email": email, "password": password});
        let resp: envoy_shared::responses::AuthResponse = self
            .request(reqwest::Method::POST, "/auth/setup", Some(body), false)
            .await?;
        config::set_token(&resp.token).map_err(ClientError::Other)?;
        Ok(resp)
    }

    pub async fn accept_invite(
        &self,
        token: &str,
        name: &str,
        password: &str,
    ) -> Result<envoy_shared::responses::AuthResponse, ClientError> {
        let body = serde_json::json!({"token": token, "name": name, "password": password});
        let resp: envoy_shared::responses::AuthResponse = self
            .request(
                reqwest::Method::POST,
                "/auth/accept-invite",
                Some(body),
                false,
            )
            .await?;
        config::set_token(&resp.token).map_err(ClientError::Other)?;
        Ok(resp)
    }

    pub async fn login(
        &self,
        email: &str,
        password: &str,
    ) -> Result<envoy_shared::responses::AuthResponse, ClientError> {
        let body = serde_json::json!({"email": email, "password": password});
        let resp: envoy_shared::responses::AuthResponse = self
            .request(reqwest::Method::POST, "/auth/login", Some(body), false)
            .await?;
        config::set_token(&resp.token).map_err(ClientError::Other)?;
        Ok(resp)
    }

    pub async fn profile(&self) -> Result<envoy_shared::responses::ProfileResponse, ClientError> {
        self.get("/auth/profile", true).await
    }

    pub async fn health(&self) -> Result<serde_json::Value, ClientError> {
        self.request(reqwest::Method::GET, "/health", None, false)
            .await
    }

    // --- projects ---
    pub async fn create_project(
        &self,
        name: &str,
        description: &str,
        git_repo: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        let body =
            serde_json::json!({"name": name, "description": description, "git_repo": git_repo});
        self.request(reqwest::Method::POST, "/projects", Some(body), true)
            .await
    }

    pub async fn list_projects(
        &self,
    ) -> Result<Vec<envoy_shared::responses::ProjectResponse>, ClientError> {
        self.get("/projects", true).await
    }

    pub async fn get_project(
        &self,
        id: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        self.get(&format!("/projects/{id}"), true).await
    }

    #[allow(dead_code)]
    pub async fn update_project(
        &self,
        id: &str,
        name: &str,
        description: &str,
        git_repo: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        let body =
            serde_json::json!({"name": name, "description": description, "git_repo": git_repo});
        self.request(
            reqwest::Method::PUT,
            &format!("/projects/{id}"),
            Some(body),
            true,
        )
        .await
    }

    #[allow(dead_code)]
    pub async fn delete_project(&self, id: &str) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/projects/{id}"),
                None,
                true,
            )
            .await?;
        Ok(())
    }

    // --- environments (flat: only create/list are project-scoped) ---
    pub async fn create_environment(
        &self,
        project_id: &str,
        name: &str,
        description: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        let body = serde_json::json!({"name": name, "description": description});
        self.request(
            reqwest::Method::POST,
            &format!("/projects/{project_id}/environments"),
            Some(body),
            true,
        )
        .await
    }

    pub async fn list_environments(
        &self,
        project_id: &str,
    ) -> Result<Vec<envoy_shared::responses::EnvironmentResponse>, ClientError> {
        self.get(&format!("/projects/{project_id}/environments"), true)
            .await
    }

    pub async fn get_environment(
        &self,
        env_id: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        self.get(&format!("/environments/{env_id}"), true).await
    }

    #[allow(dead_code)]
    pub async fn update_environment(
        &self,
        env_id: &str,
        name: &str,
        description: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        let body = serde_json::json!({"name": name, "description": description});
        self.request(
            reqwest::Method::PUT,
            &format!("/environments/{env_id}"),
            Some(body),
            true,
        )
        .await
    }

    #[allow(dead_code)]
    pub async fn delete_environment(&self, env_id: &str) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/environments/{env_id}"),
                None,
                true,
            )
            .await?;
        Ok(())
    }

    // --- variables (flat: only create/list are env-scoped) ---
    pub async fn create_variable(
        &self,
        env_id: &str,
        key: &str,
        value: &str,
        file_path: &str,
    ) -> Result<envoy_shared::responses::VariableResponse, ClientError> {
        let body =
            serde_json::json!({"key": key, "value": value, "description": "", "file_path": file_path});
        self.request(
            reqwest::Method::POST,
            &format!("/environments/{env_id}/variables"),
            Some(body),
            true,
        )
        .await
    }

    pub async fn list_variables(
        &self,
        env_id: &str,
    ) -> Result<Vec<envoy_shared::responses::VariableResponse>, ClientError> {
        self.get(&format!("/environments/{env_id}/variables"), true)
            .await
    }

    pub async fn update_variable(
        &self,
        var_id: &str,
        key: &str,
        value: &str,
        file_path: &str,
    ) -> Result<envoy_shared::responses::VariableResponse, ClientError> {
        let body =
            serde_json::json!({"key": key, "value": value, "description": "", "file_path": file_path});
        self.request(
            reqwest::Method::PUT,
            &format!("/variables/{var_id}"),
            Some(body),
            true,
        )
        .await
    }

    // --- name-or-id resolution (client-side, via list + match) ---
    pub async fn resolve_project(
        &self,
        input: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        let projects = self.list_projects().await?;
        projects
            .into_iter()
            .find(|p| p.id == input || p.name == input || p.name.eq_ignore_ascii_case(input))
            .ok_or_else(|| ClientError::Server(format!("project '{input}' not found")))
    }

    pub async fn resolve_environment(
        &self,
        project_id: &str,
        input: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        let envs = self.list_environments(project_id).await?;
        if let Some(e) = envs
            .iter()
            .find(|e| e.id == input || e.name == input || e.name.eq_ignore_ascii_case(input))
        {
            return Ok(e.clone());
        }
        // Strict: never create on typo. Suggest closest name instead.
        let available: Vec<String> = envs.iter().map(|e| e.name.clone()).collect();
        let suggestion = closest_name(input, &available);
        let mut msg = format!("environment '{input}' not found");
        if let Some(s) = suggestion {
            msg.push_str(&format!(". Did you mean '{s}'?"));
        }
        if available.is_empty() {
            msg.push_str(". No environments exist yet — run 'envoy manage envs create'");
        } else {
            msg.push_str(" — run 'envoy manage envs list' to see available environments");
        }
        Err(ClientError::Server(msg))
    }

    // --- admin / invites ---
    pub async fn create_invite(
        &self,
        email: &str,
        name: &str,
    ) -> Result<envoy_shared::responses::InviteResponse, ClientError> {
        let body = serde_json::json!({"email": email, "name": name});
        self.request(reqwest::Method::POST, "/users/invites", Some(body), true)
            .await
    }

    pub async fn list_invites(
        &self,
    ) -> Result<Vec<envoy_shared::responses::InviteListItem>, ClientError> {
        self.get("/users/invites", true).await
    }

    pub async fn revoke_invite(&self, invite_id: &str) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/users/invites/{invite_id}"),
                None,
                true,
            )
            .await?;
        Ok(())
    }

    // --- users/sharing ---
    pub async fn search_users(
        &self,
        email: &str,
    ) -> Result<Vec<envoy_shared::responses::UserSearchResponse>, ClientError> {
        self.get(&format!("/users/search?email={email}"), true)
            .await
    }

    pub async fn list_members(
        &self,
        project_id: &str,
    ) -> Result<Vec<envoy_shared::responses::ProjectMemberResponse>, ClientError> {
        self.get(&format!("/projects/{project_id}/members"), true)
            .await
    }

    pub async fn add_member(
        &self,
        project_id: &str,
        user_id: &str,
        role: &str,
    ) -> Result<(), ClientError> {
        let body = serde_json::json!({"user_id": user_id, "role": role});
        let _: serde_json::Value = self
            .request(
                reqwest::Method::POST,
                &format!("/projects/{project_id}/members"),
                Some(body),
                true,
            )
            .await?;
        Ok(())
    }

    pub async fn remove_member(&self, project_id: &str, user_id: &str) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/projects/{project_id}/members/{user_id}"),
                None,
                true,
            )
            .await?;
        Ok(())
    }
}
