pub mod auth;
pub mod docs;
pub mod environments;
pub mod health;
pub mod invites;
pub mod projects;
pub mod sharing;
pub mod variables;

use axum::{
    routing::{delete, get, post, put},
    Router,
};

use crate::state::AppState;

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/health", get(health::health))
        .route("/", get(health::home))
        .route("/favicon.ico", get(health::favicon))
        .route("/openapi.json", get(docs::openapi))
        .route("/docs", get(docs::docs))
        .route("/auth/setup", post(auth::setup))
        .route("/auth/accept-invite", post(auth::accept_invite))
        .route("/auth/login", post(auth::login))
        .route("/auth/profile", get(auth::profile))
        .route("/users/search", get(auth::search_users))
        .route(
            "/users/invites",
            post(invites::create_invite).get(invites::list_invites),
        )
        .route("/users/invites/{id}", delete(invites::revoke_invite))
        .route(
            "/projects",
            post(projects::create_project).get(projects::list_projects),
        )
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
            "/environments/{id}",
            get(environments::get_environment)
                .put(environments::update_environment)
                .delete(environments::delete_environment),
        )
        .route(
            "/environments/{environment_id}/variables",
            post(variables::create_variable).get(variables::list_variables),
        )
        .route(
            "/variables/{id}",
            get(variables::get_variable)
                .put(variables::update_variable)
                .delete(variables::delete_variable),
        )
}
