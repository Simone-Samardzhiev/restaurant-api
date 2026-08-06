use crate::{
    config::DatabaseConfig,
    domain::{error::Error, repositories::UserRepository, user::User},
};
use anyhow::{Context, anyhow};
use sqlx::PgPool;

// Connects to a postgres database returning a pool of connections.
pub async fn connect(c: &DatabaseConfig) -> Result<PgPool, anyhow::Error> {
    sqlx::postgres::PgPoolOptions::new()
        .max_connections(c.max_connections)
        .max_lifetime(Some(c.max_lifetime))
        .connect(&c.url)
        .await
        .context("Failed to connect to postgres")
}

// Applies pending migrations to the database schema.
pub async fn apply_migrations(pool: PgPool) -> Result<(), anyhow::Error> {
    sqlx::migrate!()
        .run(&pool)
        .await
        .context("Failed to migrate the database")
}

/// Postgres implementation of [UserRepository].
pub struct PostgresUserRepository {
    pool: PgPool,
}

impl PostgresUserRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait::async_trait]
impl UserRepository for PostgresUserRepository {
    #[tracing::instrument(name = "postgres_user_repository.save", skip(self, user), fields(user_id=%user.id
    ))]
    async fn save(&self, user: &User) -> Result<(), Error> {
        let result = sqlx::query(
            "INSERT INTO users (id, name, email, password, role, created_at, updated_at)
                    VALUES ($1, $2, $3, $4, $5::user_roles, $6, $7)",
        )
        .bind(&user.id)
        .bind(&user.name)
        .bind(&user.email)
        .bind(&user.password)
        .bind(user.role.as_ref())
        .bind(&user.created_at)
        .bind(&user.updated_at)
        .execute(&self.pool)
        .await;

        match result {
            Ok(_) => Ok(()),
            Err(sqlx::Error::Database(db_err)) => {
                if db_err.is_unique_violation() {
                    Err(Error::EmailAlreadyExists)
                } else {
                    Err(Error::from(anyhow!(
                        "Unexpected database error: {}",
                        db_err
                    )))
                }
            }

            Err(e) => Err(Error::from(anyhow!("Error saving user: {}", e))),
        }
    }
}
