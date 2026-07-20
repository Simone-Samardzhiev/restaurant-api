use std::env;
use std::sync::Arc;

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();

    let database_config = auth::config::DatabaseConfig::new().unwrap();
    let pg_pool = auth::repositories::postgres::connect(&database_config)
        .await
        .unwrap();
    auth::repositories::postgres::apply_migrations(pg_pool.clone())
        .await
        .unwrap();

    let app_config = auth::config::AppConfig::new().unwrap();

    let user_repository = auth::repositories::postgres::PostgresUserRepository::new(pg_pool);
    let password_hasher = auth::hashers::argon::ArgonPasswordHasher::new(argon2::Argon2::default());
    let user_service = auth::domain::services::DefaultUserService::new(
        Arc::new(user_repository),
        Arc::new(password_hasher),
    );

    auth::rest::Router::new(Arc::new(user_service), app_config.addr)
        .listen()
        .await
        .unwrap();
}
