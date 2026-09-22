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

/// TokenCoder describes methods for encoding and decoding tokens.
#[automock]
#[async_trait::async_trait]
pub trait TokenCoder: Send + Sync {
    /// Encodes the given token into a string.
    async fn encode(&self, token: &token::AccessToken) -> Result<String, Error>;

    /// Decodes the given token string into a token.
    async fn decode(&self, token: &str) -> Result<token::AccessToken, Error>;

    /// Parses and stores the given key to reduce keys parsing.
    async fn store_keys(&self, pair: keys::KeyPair) -> Result<(), Error>;
}
