use crate::domain::{error::Error, repositories::TokenRepository, token::RefreshToken};
use anyhow::Context;
use fred::{
    clients::Client,
    prelude::{Expiration, KeysInterface},
};
use rmp_serde::to_vec;
use time::OffsetDateTime;

pub struct ValkeyTokenRepository {
    client: Client,
}

impl ValkeyTokenRepository {
    pub fn new(client: Client) -> Self {
        Self { client }
    }
}

#[async_trait::async_trait]
impl TokenRepository for ValkeyTokenRepository {
    async fn save(&self, token: &RefreshToken) -> Result<(), Error> {
        let data = to_vec(token).context("Failed to encode refresh token as message pack")?;

        let _: () = self
            .client
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
            .context("Failed to save token")?;
        Ok(())
    }
}
