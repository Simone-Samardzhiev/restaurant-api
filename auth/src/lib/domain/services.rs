use super::error::Error;
use crate::domain::{
    PasswordHasher,
    repositories::UserRepository,
    user::{RegisterRequest, Role, User},
};
use std::sync::Arc;

// UserService describes how user business logic is accessed.
#[async_trait::async_trait]
pub trait UserService: Send + Sync {
    /// Registers a new user.
    /// # Returns
    /// [OK] if registration is successful.
    ///
    /// [Error::EmailAlreadyExists] if the email is already in use.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn register(&self, req: RegisterRequest) -> Result<(), Error>;
}

/// Default implementation of [UserService].
pub struct DefaultUserService {
    repository: Arc<dyn UserRepository>,
    hasher: Arc<dyn PasswordHasher>,
}
impl DefaultUserService {
    pub fn new(repository: Arc<dyn UserRepository>, hasher: Arc<dyn PasswordHasher>) -> Self {
        Self { repository, hasher }
    }
}

#[async_trait::async_trait]
impl UserService for DefaultUserService {
    async fn register(&self, req: RegisterRequest) -> Result<(), Error> {
        let hash = self.hasher.hash(&req.password)?;
        let user = User::new(req.name, req.email, hash, Role::Client);
        self.repository.save(&user).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::{
        MockPasswordHasher, repositories::MockUserRepository, user::RegisterRequest,
    };
    use mockall::predicate::*;

    #[tokio::test]
    async fn test_register() {
        let mut hasher = MockPasswordHasher::new();
        hasher
            .expect_hash()
            .with(eq("password"))
            .times(1)
            .returning(|_| Ok("hash".into()));

        let mut repo = MockUserRepository::new();
        repo.expect_save()
            .with(function(|u: &User| u.password == "hash"))
            .times(1)
            .returning(move |_| Ok(()));

        let req = RegisterRequest::new(
            "example@email.com".into(),
            "username".into(),
            "password".into(),
        );

        let service = DefaultUserService::new(Arc::new(repo), Arc::new(hasher));
        assert!(service.register(req).await.is_ok());
    }
}
