use auth::domain::{KeyGenerator, repositories::UserRepository};
use auth::repositories::valkey::ValkeyTokenRepository;
use axum::http::{Request, StatusCode};
use fred::prelude::{Builder, ClientLike, Config};
use std::sync::Arc;
use tower::ServiceExt;

async fn connect_to_valkey() -> ValkeyTokenRepository {
    let url = std::env::var("VALKEY_URL").expect("Missing valkey url");
    let config = Config::from_url(&url).expect("Failed to get valkey config");

    let client = Builder::from_config(config)
        .build_pool(1)
        .expect("Failed to build valkey client");
    client.init().await.expect("Error connecting to valkey");

    ValkeyTokenRepository::new(client)
}

#[sqlx::test]
#[ignore]
async fn test_register(pool: sqlx::PgPool) {
    let user_repo = auth::repositories::postgres::PostgresUserRepository::new(pool);
    let token_repo = connect_to_valkey().await;
    let hasher = auth::hashers::ArgonPasswordHasher::new(argon2::Argon2::default());

    let pair = auth::keys_generators::RSAKeyGenerator::new(2048)
        .generate()
        .expect("Error generationg keys");

    let coder = auth::token_coders::JWTCoder::new(&[pair], "test-aud".into(), "test-iss".into())
        .expect("Error creating jwt token coder");

    let service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repo),
        Arc::new(token_repo),
        Arc::new(coder),
        Arc::new(hasher),
    );
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

    assert_eq!(response.status(), StatusCode::NO_CONTENT);
}

#[sqlx::test]
#[ignore]
async fn test_register_conflict(pool: sqlx::PgPool) {
    let user_repo = auth::repositories::postgres::PostgresUserRepository::new(pool);
    let token_repo = connect_to_valkey().await;
    let hasher = auth::hashers::ArgonPasswordHasher::new(argon2::Argon2::default());

    let pair = auth::keys_generators::RSAKeyGenerator::new(2048)
        .generate()
        .expect("Error generationg keys");

    let coder = auth::token_coders::JWTCoder::new(&[pair], "test-aud".into(), "test-iss".into())
        .expect("Error creating jwt token coder");

    user_repo
        .save(&auth::domain::user::User::create(
            "User1234".into(),
            "example@email.com".into(),
            "Password_123".into(),
            auth::domain::user::Role::Client,
        ))
        .await
        .unwrap();

    let service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repo),
        Arc::new(token_repo),
        Arc::new(coder),
        Arc::new(hasher),
    );
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

#[sqlx::test]
#[ignore]
async fn test_login(pool: sqlx::PgPool) {
    let user_repo = auth::repositories::postgres::PostgresUserRepository::new(pool);
    let token_repo = connect_to_valkey().await;
    let hasher = auth::hashers::ArgonPasswordHasher::new(argon2::Argon2::default());

    let pair = auth::keys_generators::RSAKeyGenerator::new(2048)
        .generate()
        .expect("Error generationg keys");

    let coder = auth::token_coders::JWTCoder::new(&[pair], "test-aud".into(), "test-iss".into())
        .expect("Error creating jwt token coder");

    let service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repo),
        Arc::new(token_repo),
        Arc::new(coder),
        Arc::new(hasher),
    );
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
                        "password":"Password_123"
                    }
                    )
                    .to_string(),
                ))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::NO_CONTENT);

    let response = server
        .as_router()
        .oneshot(
            Request::builder()
                .method("POST")
                .uri("/api/v1/login")
                .header("content-type", "application/json")
                .body(axum::body::Body::from(
                    serde_json::json!({
                        "email":"example@email.com",
                        "password":"Password_123"
                    }
                    )
                    .to_string(),
                ))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::OK)
}
