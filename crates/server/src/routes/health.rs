use axum::Json;
use envoy_shared::responses::HealthResponse;

pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        health: "ok".to_string(),
    })
}

pub async fn home() -> axum::response::Html<&'static str> {
    axum::response::Html(
        r#"<!doctype html><html><head><title>Envoy</title></head><body><h1>Envoy</h1><p>Secure environment variable management. See <a href="/docs">/docs</a>.</p></body></html>"#,
    )
}

pub async fn favicon() -> axum::http::StatusCode {
    axum::http::StatusCode::NO_CONTENT
}
