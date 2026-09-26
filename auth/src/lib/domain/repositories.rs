use crate::domain::{error::Error, keys::KeyPair, token::RefreshToken, user::User};

/// UserRepository describes how used data is accessed.
#[mockall::automock]
#[async_trait::async_trait]
pub trait UserRepository: Send + Sync {
    /// Saves a user.
    /// # Returns
    /// [Ok] if the user is saved successfully.
    ///
    /// [Error::EmailAlreadyExists] if the email is already in use.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn save(&self, user: &User) -> Result<(), Error>;

    /// Retrieves a user by email.
    ///
    /// # Returns
    /// [Ok] if the user is found.
    ///
    /// [Ok(None)] if no user is found.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn get_by_email(&self, email: &str) -> Result<Option<User>, Error>;
}

/// KeyRepository describes how key pairs data is accessed.
#[mockall::automock]
#[async_trait::async_trait]
pub trait KeyRepository: Send + Sync {
    /// Saves a key pair.
    ///
    /// # Returns
    /// [Ok] if the key pair is saved successfully.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn save(&self, pair: &KeyPair) -> Result<(), Error>;

    /// Retrieves all key pairs.
    ///
    /// # Returns
    /// [Ok] if the key pairs are retrieved successfully.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn get(&self) -> Result<Vec<KeyPair>, Error>;
}

/// TokenRepository describes how refresh tokens are accessed.
#[mockall::automock]
#[async_trait::async_trait]
pub trait TokenRepository: Send + Sync {
    /// Saves a refresh token.
    ///
    /// # Returns
    /// [Ok] if the token is saved successfully.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn save(&self, token: &RefreshToken) -> Result<(), Error>;
}
