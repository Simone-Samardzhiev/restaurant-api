use super::error::Error;
use serde::{Deserialize, Serialize};
use std::str::FromStr;
use time::OffsetDateTime;
use uuid::Uuid;

/// User permission level.
#[derive(Debug, Clone, Copy, PartialEq)]
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

impl From<Role> for String {
    fn from(role: Role) -> Self {
        role.as_ref().to_string()
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

impl Serialize for Role {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(self.as_ref())
    }
}

impl<'de> Deserialize<'de> for Role {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        Self::from_str(&s).map_err(|_| serde::de::Error::custom("invalid role"))
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
    pub fn new(
        id: Uuid,
        name: String,
        email: String,
        password: String,
        role: Role,
        created_at: OffsetDateTime,
        updated_at: OffsetDateTime,
    ) -> Self {
        Self {
            id,
            name,
            email,
            password,
            role,
            created_at,
            updated_at,
        }
    }

    /// Creates a new user with the given name, email, password, and role.
    /// The id is generated using the V7 UUID variant.
    /// The created_at and updated_at fields are set to the current UTC time.
    pub fn create(name: String, email: String, password: String, role: Role) -> Self {
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

/// Request for a user to login in the system.
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

impl LoginRequest {
    pub fn new(email: String, password: String) -> Self {
        Self { email, password }
    }
}

/// Response for a user login request.
pub struct LoginResponse {
    pub access_token: String,
    pub refresh_token: String,
}

impl LoginResponse {
    pub fn new(access_token: String, refresh_token: String) -> Self {
        Self { access_token, refresh_token }
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
