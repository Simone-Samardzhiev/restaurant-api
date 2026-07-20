use crate::domain::error::Error;
use crate::domain::user::User;

/// UserRepository describes how used data is accessed.
#[mockall::automock]
#[async_trait::async_trait]
pub trait UserRepository: Send + Sync {
    async fn save(&self, user: &User) -> Result<(), Error>;
}
