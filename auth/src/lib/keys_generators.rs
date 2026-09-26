use crate::domain::{
    KeyGenerator,
    error::Error,
    keys::KeyPair,
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
            .context("Error generating RSA private key")?;
        let public_key = private_key.to_public_key();

        Ok(KeyPair::new(
            Uuid::now_v7(),
            private_key
                .to_pkcs8_pem(LineEnding::LF)
                .context("Error mapping private key to PEM")?
                .to_string(),
            public_key
                .to_public_key_pem(LineEnding::LF)
                .context("Error mapping public key to PEM")?,
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
