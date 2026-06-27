use crate::domain::error::Error;
use crate::domain::user::User;

/// UserRepository describes how used data is accessed.
pub trait UserRepository: Send + Sync {
    fn save(&self, user: &User) -> impl Future<Output = Result<(), Error>> + Send;
}
