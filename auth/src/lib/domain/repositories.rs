use crate::domain::error::Error;
use crate::domain::user::User;

/// UserRepository describes how used data is accessed.
#[mockall::automock]
pub trait UserRepository: Send + Sync + 'static {
    fn save(&self, user: &User) -> impl Future<Output = Result<(), Error>> + Send;
}
