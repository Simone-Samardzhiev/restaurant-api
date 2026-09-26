use crate::config::ValkeyConfig;
use crate::domain::{error::Error, repositories::TokenRepository, token::RefreshToken};
use anyhow::Context;
use fred::prelude::{
    Builder, ClientLike, Config, Expiration, KeysInterface, Pool, ReconnectPolicy,
};
use rmp_serde::to_vec;
use time::OffsetDateTime;

/// Connects to the Valkey server returning a pool.
pub async fn connect(c: &ValkeyConfig) -> Result<Pool, anyhow::Error> {
    let config = Config::from_url(&c.url)?;
    let reconnect_policy = ReconnectPolicy::new_exponential(30, 10, 10_000, 2);

    let client = Builder::from_config(config)
        .set_policy(reconnect_policy)
        .build_pool(c.pool_size)?;

    client.init().await?;
    Ok(client)
}

/// Valkey implementation of the [TokenRepository].
pub struct ValkeyTokenRepository {
    pool: Pool,
}

impl ValkeyTokenRepository {
    pub fn new(pool: Pool) -> Self {
        Self { pool }
    }
}

#[async_trait::async_trait]
impl TokenRepository for ValkeyTokenRepository {
    async fn save(&self, token: &RefreshToken) -> Result<(), Error> {
        let data = to_vec(token).context("Failed to encode refresh token as message pack")?;

        let _: () = self
            .pool
            .set(
                &token.key,
                data,
                Some(Expiration::EX(
                    (token.expires_at - OffsetDateTime::now_utc()).whole_seconds(),
                )),
                None,
                false,
            )
            .await
            .context("Error saving token")?;
        Ok(())
    }
}
