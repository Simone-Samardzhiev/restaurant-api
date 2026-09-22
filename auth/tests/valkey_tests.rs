use auth::domain::{repositories::TokenRepository, token::RefreshToken, user::Role};
use auth::repositories::valkey::ValkeyTokenRepository;
use fred::prelude::*;
use time::OffsetDateTime;
use uuid::Uuid;

async fn connect() -> ValkeyTokenRepository {
    let url = std::env::var("VALKEY_URL").expect("Missing valkey url");
    let config = Config::from_url(&url).expect("Failed to get valkey config");

    let client = Builder::from_config(config)
        .build()
        .expect("Failed to build valkey client");
    client.init().await.expect("Error connecting to valkey");

    ValkeyTokenRepository::new(client)
}

#[tokio::test]
#[ignore]
async fn test_save() {
    let repo = connect().await;
    let token = RefreshToken::new(
        Uuid::new_v4(),
        Role::Admin,
        OffsetDateTime::now_utc() + time::Duration::hours(1),
    );
    repo.save(&token).await.unwrap();
}
