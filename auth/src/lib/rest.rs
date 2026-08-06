use crate::domain::services::UserService;
use anyhow::Context;
use axum::routing::post;
use handlers::register;
use std::sync::Arc;
use tokio::signal;
use tokio::signal::ctrl_c;

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

async fn shutdown_signal() {
    let ctrl_c = async {
        ctrl_c().await.expect("Failed to install ctrl-c handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("Failed to install SIGTERM signal handler")
            .recv()
            .await
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = terminate => {},
        _ = ctrl_c => {},
    }

    println!("Shutting down http server...");
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
            .with_graceful_shutdown(shutdown_signal())
            .await
            .context("Error starting server")?;

        Ok(())
    }
}
