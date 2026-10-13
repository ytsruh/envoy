//! End-to-end API test against a local libsql DB.
//! Covers the self-hosted model: setup bootstrap, invite-only join,
//! projects/envs/vars CRUD, and RBAC.

use axum::{
    body::{to_bytes, Body},
    http::{Request, StatusCode},
};
use envoy_server::{db::Db, routes, state::AppState};
use serde_json::{json, Value};

async fn test_app() -> axum::Router {
    let path = std::env::temp_dir().join(format!(
        "envoy-e2e-{}-{}.db",
        std::process::id(),
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos()
    ));
    let db = Db::connect_local(path.to_str().unwrap()).await.unwrap();
    db.migrate().await.unwrap();
    let state = AppState {
        db,
        jwt_secret: "test-secret-for-e2e".to_string(),
    };
    routes::router().with_state(state)
}

async fn call(
    app: &mut axum::Router,
    method: &str,
    uri: &str,
    token: Option<&str>,
    body: Option<Value>,
) -> (StatusCode, Value) {
    let mut b = Request::builder()
        .method(method)
        .uri(uri)
        .header("content-type", "application/json");
    if let Some(t) = token {
        b = b.header("authorization", format!("Bearer {t}"));
    }
    let body = body
        .map(|v| Body::from(v.to_string()))
        .unwrap_or_else(Body::empty);
    let req = b.body(body).unwrap();
    let resp = tower::ServiceExt::oneshot(app.clone(), req).await.unwrap();
    let status = resp.status();
    let bytes = to_bytes(resp.into_body(), 1024 * 1024).await.unwrap();
    let v: Value = if bytes.is_empty() {
        Value::Null
    } else {
        serde_json::from_slice(&bytes).unwrap_or(Value::Null)
    };
    (status, v)
}

#[tokio::test]
async fn full_e2e_flow() {
    let mut app = test_app().await;

    // health + home + docs (no auth)
    let (s, v) = call(&mut app, "GET", "/health", None, None).await;
    assert_eq!(s, StatusCode::OK, "{v}");
    assert_eq!(v["health"], "ok");

    // bootstrap: first user becomes admin
    let (s, v) = call(
        &mut app,
        "POST",
        "/auth/setup",
        None,
        Some(json!({"name": "Alice", "email": "alice@example.com", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let token_a = v["token"].as_str().unwrap().to_string();
    assert!(!token_a.is_empty());

    // second setup -> 403 (already set up)
    let (s, _) = call(
        &mut app,
        "POST",
        "/auth/setup",
        None,
        Some(json!({"name": "Eve", "email": "eve@example.com", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::FORBIDDEN);

    // login
    let (s, v) = call(
        &mut app,
        "POST",
        "/auth/login",
        None,
        Some(json!({"email": "alice@example.com", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::OK, "{v}");

    // bad login -> 401
    let (s, _) = call(
        &mut app,
        "POST",
        "/auth/login",
        None,
        Some(json!({"email": "alice@example.com", "password": "wrongpass1"})),
    )
    .await;
    assert_eq!(s, StatusCode::UNAUTHORIZED);

    // profile (auth)
    let (s, v) = call(&mut app, "GET", "/auth/profile", Some(&token_a), None).await;
    assert_eq!(s, StatusCode::OK, "{v}");
    assert_eq!(v["email"], "alice@example.com");

    // profile without token -> 401
    let (s, _) = call(&mut app, "GET", "/auth/profile", None, None).await;
    assert_eq!(s, StatusCode::UNAUTHORIZED);

    // create project
    let (s, v) = call(
        &mut app,
        "POST",
        "/projects",
        Some(&token_a),
        Some(json!({"name": "demo", "description": "d", "git_repo": "acme/demo"})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let pid = v["id"].as_str().unwrap().to_string();

    // duplicate git_repo -> 409
    let (s, _) = call(
        &mut app,
        "POST",
        "/projects",
        Some(&token_a),
        Some(json!({"name": "demo2", "description": "", "git_repo": "acme/demo"})),
    )
    .await;
    assert_eq!(s, StatusCode::CONFLICT);

    // list + get + update
    let (s, v) = call(&mut app, "GET", "/projects", Some(&token_a), None).await;
    assert_eq!(s, StatusCode::OK);
    assert_eq!(v.as_array().unwrap().len(), 1);
    let (s, _) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, v) = call(
        &mut app,
        "PUT",
        &format!("/projects/{pid}"),
        Some(&token_a),
        Some(json!({"name": "demo-renamed", "description": "d2", "git_repo": "acme/demo"})),
    )
    .await;
    assert_eq!(s, StatusCode::OK, "{v}");
    assert_eq!(v["name"], "demo-renamed");

    // admin creates invite for Bob (non-admin cannot)
    let (s, _) = call(
        &mut app,
        "POST",
        "/users/invites",
        None,
        Some(json!({"email": "bob@example.com", "name": "Bob"})),
    )
    .await;
    assert_eq!(s, StatusCode::UNAUTHORIZED);
    let (s, v) = call(
        &mut app,
        "POST",
        "/users/invites",
        Some(&token_a),
        Some(json!({"email": "bob@example.com", "name": "Bob"})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let invite_token = v["token"].as_str().unwrap().to_string();
    assert!(!invite_token.is_empty());

    // pending invites list (admin)
    let (s, v) = call(&mut app, "GET", "/users/invites", Some(&token_a), None).await;
    assert_eq!(s, StatusCode::OK, "{v}");
    assert_eq!(v.as_array().unwrap().len(), 1);

    // Bob accepts invite with his own password
    let (s, v) = call(
        &mut app,
        "POST",
        "/auth/accept-invite",
        None,
        Some(json!({"token": invite_token, "name": "Bob", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let user_b = v["user"]["user_id"].as_str().unwrap().to_string();

    // reused invite -> 401
    let (s, _) = call(
        &mut app,
        "POST",
        "/auth/accept-invite",
        None,
        Some(json!({"token": invite_token, "name": "Bob", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::UNAUTHORIZED);

    // bad invite token -> 401
    let (s, _) = call(
        &mut app,
        "POST",
        "/auth/accept-invite",
        None,
        Some(json!({"token": "bogus", "name": "X", "password": "password123"})),
    )
    .await;
    assert_eq!(s, StatusCode::UNAUTHORIZED);

    // search finds Bob
    let (s, v) = call(
        &mut app,
        "GET",
        "/users/search?email=bob",
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    assert_eq!(v.as_array().unwrap().len(), 1);

    // invite for existing email -> 409
    let (s, _) = call(
        &mut app,
        "POST",
        "/users/invites",
        Some(&token_a),
        Some(json!({"email": "bob@example.com", "name": "Bob"})),
    )
    .await;
    assert_eq!(s, StatusCode::CONFLICT);

    // Bob (non-admin) cannot create invites -> 403
    let token_b = {
        let (s, v) = call(
            &mut app,
            "POST",
            "/auth/login",
            None,
            Some(json!({"email": "bob@example.com", "password": "password123"})),
        )
        .await;
        assert_eq!(s, StatusCode::OK);
        v["token"].as_str().unwrap().to_string()
    };
    let (s, _) = call(
        &mut app,
        "POST",
        "/users/invites",
        Some(&token_b),
        Some(json!({"email": "carol@example.com", "name": "Carol"})),
    )
    .await;
    assert_eq!(s, StatusCode::FORBIDDEN);

    // invite + revoke flow for Carol
    let (s, v) = call(
        &mut app,
        "POST",
        "/users/invites",
        Some(&token_a),
        Some(json!({"email": "carol@example.com", "name": ""})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let carol_id = v["invite_id"].as_str().unwrap().to_string();
    let (s, _) = call(
        &mut app,
        "DELETE",
        &format!("/users/invites/{carol_id}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);

    // sharing: add/list/update/remove member
    let (s, _) = call(
        &mut app,
        "POST",
        &format!("/projects/{pid}/members"),
        Some(&token_a),
        Some(json!({"user_id": user_b, "role": "viewer"})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED);
    let (s, v) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/members"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    assert_eq!(v.as_array().unwrap().len(), 1);
    let (s, _) = call(
        &mut app,
        "PUT",
        &format!("/projects/{pid}/members/{user_b}"),
        Some(&token_a),
        Some(json!({"role": "editor"})),
    )
    .await;
    assert_eq!(s, StatusCode::OK);

    // environments CRUD
    let (s, v) = call(
        &mut app,
        "POST",
        &format!("/projects/{pid}/environments"),
        Some(&token_a),
        Some(json!({"name": "dev", "description": ""})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let eid = v["id"].as_str().unwrap().to_string();
    let (s, v) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/environments"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    assert_eq!(v.as_array().unwrap().len(), 1);
    let (s, _) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/environments/{eid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);

    // variables CRUD
    let (s, v) = call(
        &mut app,
        "POST",
        &format!("/projects/{pid}/environments/{eid}/variables"),
        Some(&token_a),
        Some(json!({"key": "API_KEY", "value": "secret", "description": ""})),
    )
    .await;
    assert_eq!(s, StatusCode::CREATED, "{v}");
    let vid = v["id"].as_str().unwrap().to_string();
    let (s, v) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/environments/{eid}/variables"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    assert_eq!(v.as_array().unwrap().len(), 1);
    let (s, _) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/environments/{eid}/variables/{vid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, v) = call(
        &mut app,
        "PUT",
        &format!("/projects/{pid}/environments/{eid}/variables/{vid}"),
        Some(&token_a),
        Some(json!({"key": "API_KEY", "value": "rotated", "description": ""})),
    )
    .await;
    assert_eq!(s, StatusCode::OK, "{v}");
    assert_eq!(v["value"], "rotated");

    // RBAC: downgrade Bob to viewer, check write denied
    let (s, _) = call(
        &mut app,
        "PUT",
        &format!("/projects/{pid}/members/{user_b}"),
        Some(&token_a),
        Some(json!({"role": "viewer"})),
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, _) = call(
        &mut app,
        "POST",
        &format!("/projects/{pid}/environments"),
        Some(&token_b),
        Some(json!({"name": "nope", "description": ""})),
    )
    .await;
    assert_eq!(s, StatusCode::FORBIDDEN);
    // viewer can read
    let (s, _) = call(
        &mut app,
        "GET",
        &format!("/projects/{pid}/environments"),
        Some(&token_b),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);

    // cleanup: remove member, delete var/env/project
    let (s, _) = call(
        &mut app,
        "DELETE",
        &format!("/projects/{pid}/members/{user_b}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, _) = call(
        &mut app,
        "DELETE",
        &format!("/projects/{pid}/environments/{eid}/variables/{vid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, _) = call(
        &mut app,
        "DELETE",
        &format!("/projects/{pid}/environments/{eid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
    let (s, _) = call(
        &mut app,
        "DELETE",
        &format!("/projects/{pid}"),
        Some(&token_a),
        None,
    )
    .await;
    assert_eq!(s, StatusCode::OK);
}
