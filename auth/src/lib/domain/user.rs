use super::error::Error;
use std::str::FromStr;
use time::OffsetDateTime;
use uuid::Uuid;

/// User permission level.
#[derive(Debug, PartialEq)]
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

impl AsRef<str> for Role {
    fn as_ref(&self) -> &str {
        match self {
            Role::Admin => "admin",
            Role::Cook => "cook",
            Role::Waitress => "waitress",
            Role::Client => "client",
        }
    }
}

impl FromStr for Role {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "admin" => Ok(Role::Admin),
            "cook" => Ok(Role::Cook),
            "waitress" => Ok(Role::Waitress),
            "client" => Ok(Role::Client),
            _ => Err(Error::InvalidUserRole),
        }
    }
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
        let now = OffsetDateTime::now_utc();

        Self {
            id: Uuid::now_v7(),
            name,
            email,
            password,
            role,
            created_at: now,
            updated_at: now,
        }
    }
}

/// Request for a user to register in the system.
pub struct RegisterRequest {
    pub name: String,
    pub email: String,
    pub password: String,
}

impl RegisterRequest {
    pub fn new(name: String, email: String, password: String) -> Self {
        Self {
            name,
            email,
            password,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_role_from_str() {
        assert_eq!(Role::from_str("admin").unwrap(), Role::Admin);
        assert_eq!(Role::from_str("cook").unwrap(), Role::Cook);
        assert_eq!(Role::from_str("waitress").unwrap(), Role::Waitress);
        assert_eq!(Role::from_str("client").unwrap(), Role::Client);
        assert!(Role::from_str("").is_err());
    }
}
