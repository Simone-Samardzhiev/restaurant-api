use crate::domain::error::Error;
use mockall::automock;

pub mod error;
pub mod keys;
pub mod repositories;
pub mod services;
pub mod token;
pub mod user;

/// PasswordHasher describes methods for hashing and varifying passwords.
#[automock]
pub trait PasswordHasher: Send + Sync {
    fn hash(&self, password: &str) -> Result<String, Error>;

    fn verify(&self, password: &str, hash: &str) -> Result<bool, Error>;
}

/// KeyGenerator describes methods for generating token keys.
#[automock]
pub trait KeyGenerator: Send + Sync {
    fn generate(&self) -> Result<keys::KeyPair, Error>;
}

#[async_trait::async_trait]
#[automock]
pub trait TokenCoder: Send + Sync {
    async fn encode(&self, token: &token::AccessToken) -> Result<String, Error>;
    async fn decode(&self, token: &str) -> Result<token::AccessToken, Error>;
    async fn store_keys(&self, pair: keys::KeyPair) -> Result<(), Error>;
}
