use uuid::Uuid;

pub struct Key {
    pub id: Uuid,
    pub pem: String,
}

impl Key {
    pub fn new(id: Uuid, pem: String) -> Self {
        Self { id, pem }
    }
}

/// Represents pair of keys.
pub struct KeyPair {
    pub private_key: Key,
    pub public_key: Key,
}

impl KeyPair {
    pub fn new(private_key: Key, public_key: Key) -> Self {
        Self {
            private_key,
            public_key,
        }
    }
}
