use auth::domain::{
    error::Error,
    keys::KeyPair,
    repositories::{KeyRepository, UserRepository},
    user::{Role, User},
};
use auth::repositories::postgres::{PostgresKeyRepository, PostgresUserRepository};
use uuid::Uuid;

#[sqlx::test]
#[ignore]
async fn test_user_repository_save(pool: sqlx::PgPool) {
    let repository = PostgresUserRepository::new(pool);

    let user = User::create(
        "Test name".into(),
        "Test email".into(),
        "Test password".into(),
        Role::Admin,
    );

    repository.save(&user).await.unwrap();
}

#[sqlx::test]
#[ignore]
async fn test_user_repository_save_duplicate(pool: sqlx::PgPool) {
    let repository = PostgresUserRepository::new(pool);

    let mut user = User::create(
        "Test name".into(),
        "Test email".into(),
        "Test password".into(),
        Role::Admin,
    );

    repository.save(&user).await.unwrap();
    user.id = Uuid::now_v7();

    match repository.save(&user).await {
        Ok(_) => panic!("Should not be able to save user with duplicate email"),
        Err(Error::EmailAlreadyExists) => (),
        Err(e) => panic!("Unexpected error: {}", e),
    }
}

#[sqlx::test]
#[ignore]
async fn test_user_repository_get_by_email(pool: sqlx::PgPool) {
    let repository = PostgresUserRepository::new(pool);

    let user = User::create(
        "Test name".into(),
        "Test email".into(),
        "Test password".into(),
        Role::Admin,
    );

    repository.save(&user).await.unwrap();

    let found_user = repository
        .get_by_email("Test email".into())
        .await
        .unwrap()
        .unwrap();
    assert_eq!(found_user.id, user.id);

    let not_found_user = repository
        .get_by_email("Random email".into())
        .await
        .unwrap();
    assert!(not_found_user.is_none());
}

#[sqlx::test]
#[ignore]
async fn test_key_repository_save(pool: sqlx::PgPool) {
    let repository = PostgresKeyRepository::new(pool);
    let key_pair = KeyPair::new(Uuid::now_v7(), "private_key".into(), "public_key".into());

    repository
        .save(&key_pair)
        .await
        .unwrap()
}

#[sqlx::test]
#[ignore]
async fn test_key_repository_get(pool: sqlx::PgPool) {
    let repository = PostgresKeyRepository::new(pool);
    let key_pair = KeyPair::new(Uuid::now_v7(), "private_key".into(), "public_key".into());

    repository
        .save(&key_pair)
        .await
        .unwrap();

    let fethed_key = &repository
        .get()
        .await
        .unwrap()[0];

    assert_eq!(fethed_key.id, key_pair.id);
}
