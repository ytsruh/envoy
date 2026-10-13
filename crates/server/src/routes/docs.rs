use axum::Json;
use serde_json::{Value, json};

pub async fn openapi() -> Json<Value> {
    Json(json!({
        "openapi": "3.1.0",
        "info": {"title": "Envoy API", "version": "0.1.0"},
        "paths": {}
    }))
}

pub async fn docs() -> axum::response::Html<String> {
    axum::response::Html(
        r#"<!doctype html><html><head><title>Envoy API Docs</title><meta charset="utf-8"/></head><body><h1>Envoy API</h1><p>OpenAPI: <a href="/openapi.json">/openapi.json</a></p></body></html>"#.to_string(),
    )
}
