pub mod auth;
pub mod docs;
pub mod environments;
pub mod health;
pub mod projects;
pub mod sharing;
pub mod variables;

use axum::{
    Router,
    routing::{delete, get, post, put},
};

use crate::state::AppState;

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/health", get(health::health))
        .route("/", get(health::home))
        .route("/favicon.ico", get(health::favicon))
        .route("/openapi.json", get(docs::openapi))
        .route("/docs", get(docs::docs))
        .route("/auth/register", post(auth::register))
        .route("/auth/login", post(auth::login))
        .route("/auth/profile", get(auth::profile))
        .route("/users/search", get(auth::search_users))
        .route("/projects", post(projects::create_project).get(projects::list_projects))
        .route(
            "/projects/{id}",
            get(projects::get_project)
                .put(projects::update_project)
                .delete(projects::delete_project),
        )
        .route(
            "/projects/{id}/members",
            post(sharing::add_member).get(sharing::list_members),
        )
        .route(
            "/projects/{id}/members/{user_id}",
            put(sharing::update_member_role).delete(sharing::remove_member),
        )
        .route("/user/projects", get(projects::list_user_projects))
        .route(
            "/projects/{project_id}/environments",
            post(environments::create_environment).get(environments::list_environments),
        )
        .route(
            "/projects/{project_id}/environments/{id}",
            get(environments::get_environment)
                .put(environments::update_environment)
                .delete(environments::delete_environment),
        )
        .route(
            "/projects/{project_id}/environments/{environment_id}/variables",
            post(variables::create_variable).get(variables::list_variables),
        )
        .route(
            "/projects/{project_id}/environments/{environment_id}/variables/{id}",
            get(variables::get_variable)
                .put(variables::update_variable)
                .delete(variables::delete_variable),
        )
        // Back-compat alias for Go CLI bug (DELETE /api/projects/:id)
        .route(
            "/api/projects/{id}",
            delete(projects::delete_project),
        )
}
