use crate::domain::error::Error;
use crate::domain::user::User;

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
}
