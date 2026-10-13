use std::time::Duration;

use axum::Router;
use envoy_server::{config, db, routes, state};
use tower_http::{limit::RequestBodyLimitLayer, timeout::TimeoutLayer, trace::TraceLayer};
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();
    let cfg = config::Config::load()?;
    let db = db::Db::connect_auto(&cfg.database_url, &cfg.auth_token).await?;
    db.migrate().await?;
    let state = state::AppState {
        db,
        jwt_secret: cfg.jwt_secret,
    };
    let app: Router = routes::router()
        .with_state(state)
        .layer(TraceLayer::new_for_http())
        .layer(RequestBodyLimitLayer::new(2 * 1024 * 1024))
        .layer(TimeoutLayer::with_status_code(
            axum::http::StatusCode::REQUEST_TIMEOUT,
            Duration::from_secs(30),
        ));

    let listener = tokio::net::TcpListener::bind(&cfg.addr).await?;
    println!("envoy-server listening on http://{}", cfg.addr);
    axum::serve(listener, app).await?;
    Ok(())
}
