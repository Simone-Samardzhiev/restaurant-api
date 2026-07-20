use anyhow::Context;
use serde::{Deserialize, Deserializer};
use validator::Validate;

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
        match envy::prefixed("DB_").from_env::<Self>() {
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

#[derive(Debug, Deserialize, Validate)]
pub struct AppConfig {
    pub addr: String,
}

impl AppConfig {
    pub fn new() -> Result<Self, anyhow::Error> {
        match envy::from_env::<Self>() {
            Ok(config) => Ok(config),
            Err(e) => Err(anyhow::anyhow!("Error parsing app config: {}", e)),
        }
    }
}
