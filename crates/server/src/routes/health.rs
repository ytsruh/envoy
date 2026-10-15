use axum::Json;
use envoy_shared::responses::{HealthResponse, MessageBody};

pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        health: "ok".to_string(),
    })
}

pub async fn home() -> Json<MessageBody> {
    Json(MessageBody {
        message: "welcome to envoy".to_string(),
    })
}
