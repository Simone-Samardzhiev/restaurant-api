use anyhow::Context;
use serde::{Deserialize, Deserializer};
use validator::Validate;

/// Configuration for the database.
#[derive(Debug, Deserialize, Validate)]
pub struct DatabaseConfig {
    pub url: String,
    #[validate(range(min = 4, max = 100))]
    pub max_connections: u32,
    #[serde(deserialize_with = "deserialize_duration")]
    pub max_lifetime: std::time::Duration,
}

fn deserialize_duration<'de, D>(deserializer: D) -> Result<std::time::Duration, D::Error>
where
    D: Deserializer<'de>,
{
    let secs = u64::deserialize(deserializer)?;
    Ok(std::time::Duration::from_secs(secs))
}

impl DatabaseConfig {
    pub fn new() -> Result<Self, anyhow::Error> {
        match envy::prefixed("DATABASE_").from_env::<Self>() {
            Ok(config) => {
                config
                    .validate()
                    .context("Invalid database configuration")?;
                Ok(config)
            }
            Err(e) => Err(anyhow::anyhow!("Error parsing database config: {}", e)),
        }
    }
}

/// Configuration for the valkey connection.
#[derive(Debug, Deserialize)]
pub struct ValkeyConfig {
    pub url: String,
    pub pool_size: usize,
}

impl ValkeyConfig {
    pub fn new() -> Result<Self, anyhow::Error> {
        match envy::prefixed("VALKEY_").from_env::<Self>() {
            Ok(config) => Ok(config),
            Err(e) => Err(anyhow::anyhow!("Error parsing valkey config: {}", e)),
        }
    }
}

/// Configuration for the JWT.
#[derive(Debug, Deserialize)]
pub struct JWTConfig {
    pub audience: String,
    pub issuer: String,
}

impl JWTConfig {
    pub fn new() -> Result<Self, anyhow::Error> {
        match envy::prefixed("JWT_").from_env::<Self>() {
            Ok(config) => Ok(config),
            Err(e) => Err(anyhow::anyhow!("Error parsing JWT config: {}", e)),
        }
    }
}

/// Configuration for the rest api.
#[derive(Debug, Deserialize, Validate)]
pub struct AppConfig {
    pub http_addr: String,
    pub grpc_addr: String,
}

impl AppConfig {
    pub fn new() -> Result<Self, anyhow::Error> {
        match envy::from_env::<Self>() {
            Ok(config) => Ok(config),
            Err(e) => Err(anyhow::anyhow!("Error parsing app config: {}", e)),
        }
    }
}

/// Wraps all configurations into one struct.
pub struct Config {
    pub database_config: DatabaseConfig,
    pub valkey_config: ValkeyConfig,
    pub jwt_config: JWTConfig,
    pub app_config: AppConfig,
}

impl Config {
    pub fn new() -> Result<Self, anyhow::Error> {
        Ok(Self {
            database_config: DatabaseConfig::new()?,
            valkey_config: ValkeyConfig::new()?,
            jwt_config: JWTConfig::new()?,
            app_config: AppConfig::new()?,
        })
    }
}
