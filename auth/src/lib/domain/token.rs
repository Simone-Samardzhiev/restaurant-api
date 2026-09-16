use super::user::Role;
use uuid::Uuid;
use time::OffsetDateTime;

/// Token for authorization. 
pub struct AccessToken {
    pub id: Uuid,
    pub user_id: Uuid,
    pub user_role: Role,
    pub expires_at: OffsetDateTime,
}

impl AccessToken {
    pub fn new(id: Uuid, user_id: Uuid, user_role: Role, expires_at: OffsetDateTime) -> Self {
        Self { id, user_id, user_role, expires_at }
    }
}
