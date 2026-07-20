use crate::domain::error::Error;
use mockall::automock;

pub mod error;
pub mod repositories;
pub mod services;
pub mod user;

/// PasswordHasher describes methods for hashing and varifying passwords.
#[automock]
pub trait PasswordHasher: Send + Sync {
    fn hash(&self, password: &str) -> Result<String, Error>;

    fn verify(&self, password: &str, hash: &str) -> Result<bool, Error>;
}
