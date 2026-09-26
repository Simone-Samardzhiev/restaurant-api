use auth::domain::{repositories::TokenRepository, token::RefreshToken, user::Role};
use auth::repositories::valkey::ValkeyTokenRepository;
use fred::prelude::*;
use time::OffsetDateTime;
use uuid::Uuid;

async fn connect() -> ValkeyTokenRepository {
    let url = std::env::var("VALKEY_URL").unwrap();
    let config = Config::from_url(&url).unwrap();
    let client = Builder::from_config(config).build_pool(1).unwrap();
    
    client.init().await.unwrap();
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
