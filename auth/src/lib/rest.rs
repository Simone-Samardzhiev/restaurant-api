use crate::domain::services::UserService;
use anyhow::Context;
use axum::routing::post;
use handlers::{login, register};
use std::sync::Arc;

mod errors;
mod handlers;

#[derive(Clone)]
struct AppState {
    user_service: Arc<dyn UserService>,
}

impl AppState {
    pub fn new(user_service: Arc<dyn UserService>) -> Self {
        Self { user_service }
    }
}

pub struct Server {
    user_service: Arc<dyn UserService>,
    address: String,
}

impl Server {
    pub fn new(user_service: Arc<dyn UserService>, address: String) -> Self {
        Self {
            user_service,
            address,
        }
    }

    pub fn as_router(&self) -> axum::Router {
        let state = AppState::new(self.user_service.clone());

        axum::Router::new()
            .nest(
                "/api/v1",
                axum::Router::new()
                    .route("/register", post(register))
                    .route("/login", post(login)),
            )
            .with_state(state)
    }

    pub async fn listen<S>(&self, shutdown_signal: S) -> Result<(), anyhow::Error>
    where
        S: Future<Output = ()> + Send + 'static,
    {
        let router = self.as_router();

        let listener = tokio::net::TcpListener::bind(&self.address)
            .await
            .context("Error creating tcp listener")?;

        axum::serve(listener, router)
            .with_graceful_shutdown(shutdown_signal)
            .await
            .context("Error starting server")?;

        Ok(())
    }
}
