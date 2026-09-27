mod v1 {
    tonic::include_proto!("v1");
}

use crate::domain::keys::KeyPair;
use crate::domain::repositories::KeyRepository;
use anyhow::Context;
use std::net::SocketAddr;
use std::sync::Arc;
use tonic::transport::Server as TonicServer;
use tonic::{Request, Response, Status};
use v1::{
    GetKeysResponse, Key,
    auth_service_server::{AuthService as TonicAuthService, AuthServiceServer},
};

pub struct Server {
    repository: Arc<dyn KeyRepository>,
    addr: SocketAddr,
}

impl Server {
    pub fn new(repository: Arc<dyn KeyRepository>, addr: SocketAddr) -> Self {
        Self { repository, addr }
    }

    pub async fn listen<S>(&self, shutdown_signal: S) -> Result<(), anyhow::Error>
    where
        S: Future<Output = ()> + Send + 'static,
    {
        TonicServer::builder()
            .add_service(AuthServiceServer::new(AuthService::new(
                self.repository.clone(),
            )))
            .serve_with_shutdown(self.addr.clone(), shutdown_signal)
            .await
            .context("Error starting gRPC server")
    }
}

/// Implementation of gRPC auth service.
struct AuthService {
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
impl TonicAuthService for AuthService {
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
    use super::*;
    use crate::repositories::postgres::PostgresKeyRepository;
    use uuid::Uuid;

    #[sqlx::test]
    #[ignore]
    async fn test_get_keys(pool: sqlx::PgPool) {
        let repository = PostgresKeyRepository::new(pool.clone());
        let pair = KeyPair::new(Uuid::now_v7(), "Private key".into(), "Public ket".into());

        repository.save(&pair).await.unwrap();

        let service = AuthService::new(Arc::new(repository));
        let response = service
            .get_keys(Request::new(()))
            .await
            .unwrap()
            .into_inner();

        assert_eq!(response.keys.len(), 1);
        assert_eq!(response.keys[0].id, pair.id.to_string());
    }
}
