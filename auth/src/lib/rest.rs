use crate::domain::services::UserService;
use anyhow::Context;
use axum::routing::post;
use handlers::register;
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
                axum::Router::new().route("/register", post(register)),
            )
            .with_state(state)
    }

    pub async fn listen(&self) -> Result<(), anyhow::Error> {
        let router = self.as_router();

        let listener = tokio::net::TcpListener::bind(&self.address)
            .await
            .context("Error creating tcp listener")?;

        axum::serve(listener, router)
            .await
            .context("Error running server")?;

        Ok(())
    }
}
