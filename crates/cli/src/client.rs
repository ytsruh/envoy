use serde::de::DeserializeOwned;
use serde_json::Value;

use crate::config;

#[derive(Debug, thiserror::Error)]
pub enum ClientError {
    #[error("no token - please login first using 'envoy auth login'")]
    NoToken,
    #[error("session expired - please login again using 'envoy auth login'")]
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

impl Client {
    pub fn new() -> anyhow::Result<Self> {
        let _ = dotenvy::dotenv();
        Ok(Self {
            base: config::server_url(),
            token: config::get_token(),
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
    ) -> Result<T, ClientError>
    {
        let resp = self.raw(method, path, body, auth).await?;
        let status = resp.status();
        if !status.is_success() {
            let err_msg = resp
                .json::<serde_json::Value>()
                .await
                .ok()
                .and_then(|v| v.get("error").and_then(|e| e.as_str()).map(|s| s.to_string()))
                .unwrap_or_else(|| format!("HTTP {status}"));
            if status == reqwest::StatusCode::UNAUTHORIZED && err_msg == "Token has expired" {
                let _ = config::clear_token();
                return Err(ClientError::Expired);
            }
            return Err(ClientError::Server(err_msg));
        }
        // Allow empty body for message endpoints when T is Value
        let parsed = resp.json::<T>().await?;
        Ok(parsed)
    }

    async fn request_unit(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<Value>,
        auth: bool,
    ) -> Result<serde_json::Value, ClientError> {
        self.request::<serde_json::Value>(method, path, body, auth)
            .await
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

    // --- auth ---
    pub async fn register(
        &self,
        name: &str,
        email: &str,
        password: &str,
    ) -> Result<envoy_shared::responses::AuthResponse, ClientError> {
        let body = serde_json::json!({"name": name, "email": email, "password": password});
        let resp: envoy_shared::responses::AuthResponse = self
            .request(reqwest::Method::POST, "/auth/register", Some(body), false)
            .await?;
        let _ = config::set_token(&resp.token);
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
        let _ = config::set_token(&resp.token);
        Ok(resp)
    }

    pub async fn profile(&self) -> Result<envoy_shared::responses::ProfileResponse, ClientError> {
        self.get("/auth/profile", true).await
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
        let v: Vec<envoy_shared::responses::ProjectResponse> =
            self.get("/projects", true).await.unwrap_or_default();
        Ok(v)
    }

    pub async fn get_project(
        &self,
        id: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        self.get(&format!("/projects/{id}"), true).await
    }

    pub async fn update_project(
        &self,
        id: &str,
        name: &str,
        description: &str,
        git_repo: &str,
    ) -> Result<envoy_shared::responses::ProjectResponse, ClientError> {
        let body =
            serde_json::json!({"name": name, "description": description, "git_repo": git_repo});
        self.request(reqwest::Method::PUT, &format!("/projects/{id}"), Some(body), true)
            .await
    }

    pub async fn delete_project(&self, id: &str) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(reqwest::Method::DELETE, &format!("/projects/{id}"), None, true)
            .await?;
        Ok(())
    }

    // --- environments ---
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
        let v: Vec<envoy_shared::responses::EnvironmentResponse> = self
            .get(&format!("/projects/{project_id}/environments"), true)
            .await
            .unwrap_or_default();
        Ok(v)
    }

    pub async fn get_environment(
        &self,
        project_id: &str,
        env_id: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        self.get(
            &format!("/projects/{project_id}/environments/{env_id}"),
            true,
        )
        .await
    }

    pub async fn update_environment(
        &self,
        project_id: &str,
        env_id: &str,
        name: &str,
        description: &str,
    ) -> Result<envoy_shared::responses::EnvironmentResponse, ClientError> {
        let body = serde_json::json!({"name": name, "description": description});
        self.request(
            reqwest::Method::PUT,
            &format!("/projects/{project_id}/environments/{env_id}"),
            Some(body),
            true,
        )
        .await
    }

    pub async fn delete_environment(
        &self,
        project_id: &str,
        env_id: &str,
    ) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/projects/{project_id}/environments/{env_id}"),
                None,
                true,
            )
            .await?;
        Ok(())
    }

    // --- variables ---
    pub async fn create_variable(
        &self,
        project_id: &str,
        env_id: &str,
        key: &str,
        value: &str,
    ) -> Result<envoy_shared::responses::VariableResponse, ClientError> {
        let body = serde_json::json!({"key": key, "value": value, "description": ""});
        self.request(
            reqwest::Method::POST,
            &format!("/projects/{project_id}/environments/{env_id}/variables"),
            Some(body),
            true,
        )
        .await
    }

    pub async fn list_variables(
        &self,
        project_id: &str,
        env_id: &str,
    ) -> Result<Vec<envoy_shared::responses::VariableResponse>, ClientError> {
        let v: Vec<envoy_shared::responses::VariableResponse> = self
            .get(
                &format!("/projects/{project_id}/environments/{env_id}/variables"),
                true,
            )
            .await
            .unwrap_or_default();
        Ok(v)
    }

    pub async fn get_variable(
        &self,
        project_id: &str,
        env_id: &str,
        var_id: &str,
    ) -> Result<envoy_shared::responses::VariableResponse, ClientError> {
        self.get(
            &format!("/projects/{project_id}/environments/{env_id}/variables/{var_id}"),
            true,
        )
        .await
    }

    pub async fn update_variable(
        &self,
        project_id: &str,
        env_id: &str,
        var_id: &str,
        key: &str,
        value: &str,
    ) -> Result<envoy_shared::responses::VariableResponse, ClientError> {
        let body = serde_json::json!({"key": key, "value": value, "description": ""});
        self.request(
            reqwest::Method::PUT,
            &format!("/projects/{project_id}/environments/{env_id}/variables/{var_id}"),
            Some(body),
            true,
        )
        .await
    }

    pub async fn delete_variable(
        &self,
        project_id: &str,
        env_id: &str,
        var_id: &str,
    ) -> Result<(), ClientError> {
        let _: serde_json::Value = self
            .request(
                reqwest::Method::DELETE,
                &format!("/projects/{project_id}/environments/{env_id}/variables/{var_id}"),
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
        let v: Vec<envoy_shared::responses::UserSearchResponse> = self
            .get(&format!("/users/search?email={email}"), true)
            .await
            .unwrap_or_default();
        Ok(v)
    }

    pub async fn list_members(
        &self,
        project_id: &str,
    ) -> Result<Vec<envoy_shared::responses::ProjectMemberResponse>, ClientError> {
        let v: Vec<envoy_shared::responses::ProjectMemberResponse> = self
            .get(&format!("/projects/{project_id}/members"), true)
            .await
            .unwrap_or_default();
        Ok(v)
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

    pub async fn remove_member(
        &self,
        project_id: &str,
        user_id: &str,
    ) -> Result<(), ClientError> {
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

    pub async fn _raw(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<Value>,
        auth: bool,
    ) -> Result<serde_json::Value, ClientError> {
        self.request_unit(method, path, body, auth).await
    }
}
