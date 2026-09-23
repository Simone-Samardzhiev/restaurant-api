use auth::domain::KeyGenerator;
use auth::domain::repositories::KeyRepository;
use std::sync::Arc;

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();

    let database_config = auth::config::DatabaseConfig::new().unwrap();
    let valkey_config = auth::config::ValkeyConfig::new().unwrap();
    let jwt_config = auth::config::JWTConfig::new().unwrap();

    let pg_pool = auth::repositories::postgres::connect(&database_config)
        .await
        .unwrap();
    let valkey_pool = auth::repositories::valkey::connect(&valkey_config)
        .await
        .unwrap();

    auth::repositories::postgres::apply_migrations(pg_pool.clone())
        .await
        .unwrap();

    let user_repository =
        auth::repositories::postgres::PostgresUserRepository::new(pg_pool.clone());
    let token_repository =
        auth::repositories::valkey::ValkeyTokenRepository::new(valkey_pool.clone());
    let key_repository = auth::repositories::postgres::PostgresKeyRepository::new(pg_pool.clone());

    let password_hasher = auth::hashers::ArgonPasswordHasher::new(argon2::Argon2::default());

    let key_generator = auth::keys_generators::RSAKeyGenerator::new(2048);
    let mut keys = key_repository.get().await.unwrap();
    if keys.is_empty() {
        keys.push(key_generator.generate().unwrap());
    }

    let token_coder =
        auth::token_coders::JWTCoder::new(&keys, jwt_config.audience, jwt_config.issuer).unwrap();

    let user_service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repository),
        Arc::new(token_repository),
        Arc::new(token_coder),
        Arc::new(password_hasher),
    );

    let app_config = auth::config::AppConfig::new().unwrap();
    auth::rest::Server::new(Arc::new(user_service), app_config.addr)
        .listen()
        .await
        .unwrap();

    println!("HTTP server closed down successfully");
    pg_pool.close().await;
    println!("Database connection pool closed");
    std::process::exit(0);
}
