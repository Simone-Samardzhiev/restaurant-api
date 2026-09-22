use super::user::Role;
use time::OffsetDateTime;
use uuid::Uuid;
use rand::Rng;
use base64::Engine;
use serde::{Serialize, Deserialize};


/// Token for authorization.
#[derive(Debug)]
pub struct AccessToken {
    pub id: Uuid,
    pub user_id: Uuid,
    pub user_role: Role,
    pub expires_at: OffsetDateTime,
}

impl AccessToken {
    pub fn new(id: Uuid, user_id: Uuid, user_role: Role, expires_at: OffsetDateTime) -> Self {
        Self {
            id,
            user_id,
            user_role,
            expires_at,
        }
    }
}


/// Token for refreshing session.
#[derive(Debug, Serialize, Deserialize)]
pub struct RefreshToken {
    pub key: String,
    pub user_id: Uuid,
    pub user_role: Role,
    pub expires_at: OffsetDateTime,
}

impl RefreshToken {
    pub fn new(user_id: Uuid, user_role: Role, expires_at: OffsetDateTime) -> Self{
        let mut bytes = [0u8; 32];
        let mut rng = rand::thread_rng();
        rng.fill(&mut bytes);
        let key = base64::engine::general_purpose::STANDARD.encode(bytes);
        
        Self {
            key,
            user_id,
            user_role,
            expires_at,
        }
    }
    
}
