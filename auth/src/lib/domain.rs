use crate::domain::error::Error;

pub mod error;
pub mod repositories;
pub mod user;


/// PasswordHasher describes methods for hashing and varifying passwords.
pub trait PasswordHasher {
    fn hash(&self, password: &str) -> Result<String, Error>;

    fn verify(&self, password: &str, hash: &str) -> Result<bool, Error>;
}
