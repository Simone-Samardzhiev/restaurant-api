use auth::config;
use auth::domain::KeyGenerator;
use auth::domain::repositories::{KeyRepository};
use auth::hashers::ArgonPasswordHasher;
use auth::keys_generators::RSAKeyGenerator;
use auth::repositories::{
    postgres::{
        PostgresKeyRepository, PostgresUserRepository, apply_migrations,
        connect as connect_to_postgres,
    },
    valkey::{ValkeyTokenRepository, connect as connect_to_valkey},
};
use auth::token_coders::JWTCoder;
use std::sync::Arc;
use tokio::signal;
use tokio::signal::ctrl_c;
use tokio::sync::broadcast;

async fn shutdown_signal() {
    let ctrl_c = async {
        ctrl_c().await.expect("Error installing ctrl-c handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("Error installing SIGTERM signal handler")
            .recv()
            .await
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = terminate => {},
        _ = ctrl_c => {},
    }

    println!("Shutting down http server...");
}

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();

    let config = config::Config::new().unwrap();

    let pg_pool = connect_to_postgres(&config.database_config).await.unwrap();
    let valkey_pool = connect_to_valkey(&config.valkey_config).await.unwrap();
    apply_migrations(pg_pool.clone()).await.unwrap();

    let user_repository = Arc::new(PostgresUserRepository::new(pg_pool.clone()));
    let token_repository = Arc::new(ValkeyTokenRepository::new(valkey_pool.clone()));
    let key_repository = Arc::new(PostgresKeyRepository::new(pg_pool.clone()));
    let password_hasher = Arc::new(ArgonPasswordHasher::new(argon2::Argon2::default()));
    let key_generator = Arc::new(RSAKeyGenerator::new(2048));

    let mut keys = key_repository.get().await.unwrap();
    if keys.is_empty() {
        let pair = key_generator.generate().unwrap();
        key_repository.save(&pair).await.unwrap();
        keys.push(pair);
    }

    let token_coder = Arc::new(
        JWTCoder::new(&keys, config.jwt_config.audience, config.jwt_config.issuer).unwrap(),
    );

    let user_service = auth::domain::services::DefaultUserService::new(
        user_repository,
        token_repository,
        token_coder,
        password_hasher,
    );

    let (sender, _) = broadcast::channel::<()>(1);

    let mut http_subscriber = sender.subscribe();
    let http_task = tokio::spawn(async move {
        auth::rest::Server::new(Arc::new(user_service), config.app_config.http_addr)
            .listen(async move {
                let _ = http_subscriber.recv().await;
            })
            .await
            .unwrap()
    });

    let mut grpc_subscriber = sender.subscribe();
    let grpc_task = tokio::spawn(async move {
        auth::grpc::Server::new(key_repository, config.app_config.grpc_addr.parse().unwrap())
            .listen(async move {
                let _ = grpc_subscriber.recv().await;
            })
            .await
            .unwrap()
    });

    shutdown_signal().await;
    let _ = sender.send(());
    let _ = tokio::join!(http_task, grpc_task);
    pg_pool.close().await;
    println!("Database connection pool closed");
}
