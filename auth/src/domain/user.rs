use time::OffsetDateTime;
use uuid::Uuid;

/// User permission level.
pub enum Role {
    // Has full access to the system
    Admin,
    // Has access to accepting orders and mark them as complete
    Cook,
    // Has access to manage order session and mark orders as delivered
    Waitress,
    // Has access to ordering
    Client,
}

/// Registered user model.
pub struct User {
    pub id: Uuid,
    pub name: String,
    pub email: String,
    pub password: String,
    pub role: Role,
    pub created_at: OffsetDateTime,
    pub updated_at: OffsetDateTime,
}

impl User {
    pub fn new(name: String, email: String, password: String, role: Role) -> Self {
        Self {
            id: Uuid::now_v7(),
            name,
            email,
            password,
            role,
            created_at: OffsetDateTime::now_utc(),
            updated_at: OffsetDateTime::now_utc(),
        }
    }
}
