use auth::domain::{
    error::Error,
    repositories::UserRepository,
    user::{Role, User},
};
use auth::repositories::postgres::PostgresUserRepository;
use uuid::Uuid;

#[sqlx::test]
#[ignore]
async fn test_user_repository_save(pool: sqlx::PgPool) {
    let repository = PostgresUserRepository::new(pool);

    let user = User::new(
        "Test name".into(),
        "Test email".into(),
        "Test password".into(),
        Role::Admin,
    );

    repository.save(&user).await.expect("Failed to save user");
}

#[sqlx::test]
#[ignore]
async fn test_user_repository_save_duplicate(pool: sqlx::PgPool) {
    let repository = PostgresUserRepository::new(pool);

    let mut user = User::new(
        "Test name".into(),
        "Test email".into(),
        "Test password".into(),
        Role::Admin,
    );

    repository.save(&user).await.expect("Failed to save user");
    user.id = Uuid::now_v7();

    match repository.save(&user).await {
        Ok(_) => panic!("Should not be able to save user with duplicate email"),
        Err(Error::EmailAlreadyExists) => (),
        Err(e) => panic!("Unexpected error: {}", e),
    }
}
