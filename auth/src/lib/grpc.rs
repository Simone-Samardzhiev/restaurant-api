mod v1 {
    tonic::include_proto!("v1");
}

use crate::domain::keys::KeyPair;
use crate::domain::repositories::KeyRepository;
use std::sync::Arc;
use tonic::{Request, Response, Status};
use v1::{GetKeysResponse, Key};

/// Implementation of gRPC auth service.
pub struct AuthService {
    repository: Arc<dyn KeyRepository>,
}

impl AuthService {
    pub fn new(repository: Arc<dyn KeyRepository>) -> Self {
        Self { repository }
    }
}

impl From<KeyPair> for Key {
    fn from(key: KeyPair) -> Self {
        Self {
            id: key.id.to_string(),
            pem: key.public_key,
        }
    }
}

#[tonic::async_trait]
impl v1::auth_service_server::AuthService for AuthService {
    async fn get_keys(&self, _request: Request<()>) -> Result<Response<GetKeysResponse>, Status> {
        let keys: Vec<Key> = self
            .repository
            .get()
            .await
            .map_err(|e| Status::internal(e.to_string()))?
            .into_iter()
            .map(Key::from)
            .collect();

        Ok(Response::new(GetKeysResponse { keys }))
    }
}

#[cfg(test)]
mod tests {
    use super::v1::auth_service_server::AuthService as _;
    use super::*;
    use crate::repositories::postgres::PostgresKeyRepository;
    use uuid::Uuid;

    #[sqlx::test]
    #[ignore]
    async fn test_get_keys(pool: sqlx::PgPool) {
        let repository = PostgresKeyRepository::new(pool.clone());
        let pair = KeyPair::new(Uuid::now_v7(), "Private key".into(), "Public ket".into());

        repository.save(&pair).await.expect("Failed to save key");

        let service = AuthService::new(Arc::new(repository));
        let response = service
            .get_keys(Request::new(()))
            .await
            .expect("Failed to call get_keys")
            .into_inner();

        assert_eq!(response.keys.len(), 1);
        assert_eq!(response.keys[0].id, pair.id.to_string());
    }
}
