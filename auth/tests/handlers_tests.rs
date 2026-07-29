use auth::domain::repositories::UserRepository;
use axum::http::{Request, StatusCode};
use std::sync::Arc;
use tower::ServiceExt;

#[sqlx::test]
#[ignore]
async fn test_register(pool: sqlx::PgPool) {
    let repo = auth::repositories::postgres::PostgresUserRepository::new(pool);
    let hasher = auth::hashers::argon::ArgonPasswordHasher::new(argon2::Argon2::default());
    let service = auth::domain::services::DefaultUserService::new(Arc::new(repo), Arc::new(hasher));
    let server = auth::rest::Server::new(Arc::new(service), "0.0.0.0:8000".into());
    let response = server
        .as_router()
        .oneshot(
            Request::builder()
                .method("POST")
                .uri("/api/v1/register")
                .header("content-type", "application/json")
                .body(axum::body::Body::from(
                    serde_json::json!({
                        "email":"example@email.com",
                        "username":"Username123",
                        "password":"Password_123"}
                    )
                    .to_string(),
                ))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::OK);
}

#[sqlx::test]
#[ignore]
async fn test_register_conflict(pool: sqlx::PgPool) {
    let repo = auth::repositories::postgres::PostgresUserRepository::new(pool);
    repo.save(&auth::domain::user::User::new(
        "User1234".into(),
        "example@email.com".into(),
        "Password_123".into(),
        auth::domain::user::Role::Client,
    ))
    .await
    .unwrap();

    let hasher = auth::hashers::argon::ArgonPasswordHasher::new(argon2::Argon2::default());
    let service = auth::domain::services::DefaultUserService::new(Arc::new(repo), Arc::new(hasher));
    let server = auth::rest::Server::new(Arc::new(service), "0.0.0.0:8000".into());

    let response = server
        .as_router()
        .oneshot(
            Request::builder()
                .method("POST")
                .uri("/api/v1/register")
                .header("content-type", "application/json")
                .body(axum::body::Body::from(
                    serde_json::json!({
                        "email":"example@email.com",
                        "username":"Username123",
                        "password":"Password_123"}
                    )
                    .to_string(),
                ))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::CONFLICT);
}
