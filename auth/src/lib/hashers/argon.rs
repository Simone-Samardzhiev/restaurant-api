use crate::domain::{PasswordHasher, error::Error};
use anyhow::Context;
use argon2::{
    Argon2, PasswordHash, PasswordHasher as Argon2PasswordHasher, PasswordVerifier,
    password_hash::{SaltString, rand_core::OsRng},
};

/// Implementation of [PasswordHasher] using argon.
pub struct ArgonPasswordHasher {
    argon2: Argon2<'static>,
}
impl ArgonPasswordHasher {
    pub fn new(argon2: Argon2<'static>) -> Self {
        Self { argon2 }
    }
}

impl<'a> PasswordHasher for ArgonPasswordHasher {
    fn hash(&self, password: &str) -> Result<String, Error> {
        let salt = SaltString::generate(&mut OsRng);

        Ok(self
            .argon2
            .hash_password(password.as_bytes(), &salt)
            .context("Error hashing password")?
            .to_string())
    }
    fn verify(&self, password: &str, hash: &str) -> Result<bool, Error> {
        let parsed = PasswordHash::new(&hash);
        match parsed {
            Ok(hash) => Ok(self
                .argon2
                .verify_password(password.as_ref(), &hash)
                .is_ok()),
            Err(_) => Ok(false),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_password_hash() {
        let hasher = ArgonPasswordHasher::new(Argon2::default());
        let password = "hunter2";
        hasher.hash(password).unwrap();
    }

    #[test]
    fn test_password_verify() {
        let hasher = ArgonPasswordHasher::new(Argon2::default());
        let password = "hunter2";
        let hash = hasher.hash(password).unwrap();
        assert!(hasher.verify(password, &hash).unwrap());
    }
}
