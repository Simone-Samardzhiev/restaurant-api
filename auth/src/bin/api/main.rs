use auth::config;
use auth::domain::KeyGenerator;
use auth::domain::repositories::KeyRepository;
use auth::repositories::{
    postgres::{
        PostgresKeyRepository, PostgresUserRepository, apply_migrations,
        connect as connect_to_postgres,
    },
    valkey::{ValkeyTokenRepository, connect as connect_to_valkey},
};
use auth::token_coders::JWTCoder;

use std::sync::Arc;

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();

    let config = config::Config::new().unwrap();

    let pg_pool = connect_to_postgres(&config.database_config).await.unwrap();
    let valkey_pool = connect_to_valkey(&config.valkey_config).await.unwrap();
    apply_migrations(pg_pool.clone()).await.unwrap();

    let user_repository = PostgresUserRepository::new(pg_pool.clone());
    let token_repository = ValkeyTokenRepository::new(valkey_pool.clone());
    let key_repository = PostgresKeyRepository::new(pg_pool.clone());

    let password_hasher = auth::hashers::ArgonPasswordHasher::new(argon2::Argon2::default());

    let key_generator = auth::keys_generators::RSAKeyGenerator::new(2048);
    let mut keys = key_repository.get().await.unwrap();
    if keys.is_empty() {
        keys.push(key_generator.generate().unwrap());
    }

    let token_coder =
        JWTCoder::new(&keys, config.jwt_config.audience, config.jwt_config.issuer).unwrap();

    let user_service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repository),
        Arc::new(token_repository),
        Arc::new(token_coder),
        Arc::new(password_hasher),
    );

    auth::rest::Server::new(Arc::new(user_service), config.app_config.addr)
        .listen()
        .await
        .unwrap();

    println!("HTTP server closed down successfully");
    pg_pool.close().await;
    println!("Database connection pool closed");
    std::process::exit(0);
}
