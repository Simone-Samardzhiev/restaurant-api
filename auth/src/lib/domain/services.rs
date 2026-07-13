use super::error::Error;
use crate::domain::{
    PasswordHasher,
    repositories::UserRepository,
    user::{RegisterRequest, Role, User},
};

// UserService describes how user business logic is accessed.
pub trait UserService: Send + Sync + 'static {
    //
    fn register(&self, req: RegisterRequest) -> impl Future<Output = Result<(), Error>> + Send;
}

pub struct DefaultUserService<U: UserRepository, P: PasswordHasher> {
    repository: U,
    hasher: P,
}
impl<U: UserRepository, P: PasswordHasher> DefaultUserService<U, P> {
    pub fn new(repository: U, hasher: P) -> Self {
        Self { repository, hasher }
    }
}

impl<U: UserRepository, P: PasswordHasher> UserService for DefaultUserService<U, P> {
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
            .returning(move |_| Box::pin(async { Ok(()) }));

        let req = RegisterRequest::new(
            "example@email.com".into(),
            "username".into(),
            "password".into(),
        );

        let service = DefaultUserService::new(repo, hasher);
        assert!(service.register(req).await.is_ok());
    }
}
