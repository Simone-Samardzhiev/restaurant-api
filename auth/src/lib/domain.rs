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
pub trait KeyGenerator {
    fn generate(&self) -> Result<keys::KeyPair, Error>;
}
