use uuid::Uuid;

pub struct KeyPair {
    pub id: Uuid,
    pub private_key: String,
    pub public_key: String,
}

impl KeyPair {
    pub fn new(id: Uuid, private_key: String, public_key: String) -> Self {
        Self {
            id,
            private_key,
            public_key,
        }
    }
}
