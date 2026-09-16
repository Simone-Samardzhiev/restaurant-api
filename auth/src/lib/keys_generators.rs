use crate::domain::{
    KeyGenerator,
    error::Error,
    keys::{Key, KeyPair},
};
use anyhow::Context;
use rsa::{
    RsaPrivateKey,
    pkcs8::{EncodePrivateKey, EncodePublicKey, LineEnding},
};
use uuid::Uuid;

/// Generates RSA keys.
pub struct RSAKeyGenerator {
    key_size: usize,
}

impl RSAKeyGenerator {
    pub fn new(key_size: usize) -> Self {
        Self { key_size }
    }
}

impl KeyGenerator for RSAKeyGenerator {
    fn generate(&self) -> Result<KeyPair, Error> {
        let mut rng = rand::thread_rng();
        let private_key = RsaPrivateKey::new(&mut rng, self.key_size)
            .context("Failed to generate RSA private key")?;
        let public_key = private_key.to_public_key();

        Ok(KeyPair::new(
            Key::new(
                Uuid::now_v7(),
                private_key
                    .to_pkcs8_pem(LineEnding::LF)
                    .context("Failed to map private ket to pem")?
                    .to_string(),
            ),
            Key::new(
                Uuid::now_v7(),
                public_key
                    .to_public_key_pem(LineEnding::LF)
                    .context("Failed to map public key to pem")?,
            ),
        ))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate() {
        let generator = RSAKeyGenerator::new(2048);
        generator.generate().unwrap();
    }
}
