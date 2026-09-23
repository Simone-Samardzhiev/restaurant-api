use super::error::Error;
use crate::domain::{
    PasswordHasher, TokenCoder,
    repositories::{TokenRepository, UserRepository},
    token::{AccessToken, RefreshToken},
    user::{LoginRequest, LoginResponse, RegisterRequest, Role, User},
};
use std::sync::Arc;
use time::{Duration, OffsetDateTime};
use uuid::Uuid;

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

    /// Logs in a user.
    /// # Returns
    /// [OK] if login is successful.
    ///
    /// [Error::WrongCredentials] if the credentials are invalid.
    ///
    /// [Error::Internal] if unknown error occurred.
    async fn login(&self, req: LoginRequest) -> Result<LoginResponse, Error>;
}

/// Default implementation of [UserService].
pub struct DefaultUserService {
    user_repository: Arc<dyn UserRepository>,
    token_repository: Arc<dyn TokenRepository>,

    token_coder: Arc<dyn TokenCoder>,
    hasher: Arc<dyn PasswordHasher>,
}

impl DefaultUserService {
    pub fn new(
        user_repository: Arc<dyn UserRepository>,
        token_repository: Arc<dyn TokenRepository>,
        token_coder: Arc<dyn TokenCoder>,
        hasher: Arc<dyn PasswordHasher>,
    ) -> Self {
        Self {
            user_repository,
            token_repository,
            token_coder,
            hasher,
        }
    }
}

#[async_trait::async_trait]
impl UserService for DefaultUserService {
    async fn register(&self, req: RegisterRequest) -> Result<(), Error> {
        let hash = self.hasher.hash(&req.password)?;
        let user = User::create(req.name, req.email, hash, Role::Client);
        self.user_repository.save(&user).await
    }

    async fn login(&self, req: LoginRequest) -> Result<LoginResponse, Error> {
        let user = self
            .user_repository
            .get_by_email(&req.email)
            .await?
            .ok_or(Error::WrongCredentials)?;

        let is_valid = self.hasher.verify(&req.password, &user.password)?;
        if !is_valid {
            return Err(Error::WrongCredentials);
        }

        let access_token = AccessToken::new(
            Uuid::now_v7(),
            user.id,
            user.role,
            OffsetDateTime::now_utc() + Duration::minutes(15),
        );
        let refresh_token = RefreshToken::new(
            user.id,
            user.role,
            OffsetDateTime::now_utc() + Duration::days(14),
        );

        self.token_repository.save(&refresh_token).await?;

        Ok(LoginResponse::new(
            self.token_coder.encode(&access_token).await?,
            refresh_token.key,
        ))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::{
        MockPasswordHasher, MockTokenCoder,
        repositories::{MockTokenRepository, MockUserRepository},
        user::RegisterRequest,
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

        let service = DefaultUserService::new(
            Arc::new(repo),
            Arc::new(MockTokenRepository::new()),
            Arc::new(MockTokenCoder::new()),
            Arc::new(hasher),
        );
        assert!(service.register(req).await.is_ok());
    }

    #[tokio::test]
    async fn test_login() {
        let mut user_repo = MockUserRepository::new();
        user_repo
            .expect_get_by_email()
            .with(eq("example@email.com"))
            .times(1)
            .returning(move |_| {
                Ok(Some(User::new(
                    Uuid::new_v4(),
                    "username".into(),
                    "example@email.com".into(),
                    "hash".into(),
                    Role::Client,
                    OffsetDateTime::now_utc(),
                    OffsetDateTime::now_utc(),
                )))
            });

        let mut hasher = MockPasswordHasher::new();
        hasher
            .expect_verify()
            .with(eq("password"), eq("hash"))
            .times(1)
            .returning(move |_, _| Ok(true));

        let mut token_repo = MockTokenRepository::new();
        token_repo.expect_save().times(1).returning(move |_| Ok(()));

        let mut token_coder = MockTokenCoder::new();

        token_coder
            .expect_encode()
            .times(1)
            .returning(|_| Ok("token".into()));

        let service = DefaultUserService::new(
            Arc::new(user_repo),
            Arc::new(token_repo),
            Arc::new(token_coder),
            Arc::new(hasher),
        );

        service
            .login(LoginRequest::new(
                "example@email.com".into(),
                "password".into(),
            ))
            .await
            .expect("Error loggin in");
    }
}
