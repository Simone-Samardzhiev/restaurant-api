use crate::domain::{TokenCoder, error::Error, keys::KeyPair, token::AccessToken, user::Role};
use std::collections::VecDeque;

use anyhow::Context;
use jsonwebtoken::{
    Algorithm, DecodingKey, EncodingKey, Header, Validation, decode, decode_header, encode,
};
use serde::{Deserialize, Serialize};
use time::OffsetDateTime;
use tokio::sync::RwLock;
use uuid::Uuid;


/// Encoding key for JWT with id.
#[derive(Debug, Clone)]
struct JWTEncodingKey {
    id: Uuid,
    key: EncodingKey,
}

impl JWTEncodingKey {
    fn new(id: Uuid, key: EncodingKey) -> Self {
        Self { id, key }
    }
}

/// Decoding key for JWT with id.
#[derive(Debug, Clone)]
struct JWTDecodingKey {
    id: Uuid,
    key: DecodingKey,
}

impl JWTDecodingKey {
    fn new(id: Uuid, key: DecodingKey) -> Self {
        Self { id, key }
    }
}

/// State of the JWT store, holding encoding and decoding keys.
struct JWTStoreState {
    encoding_key: JWTEncodingKey,
    decoding_keys: VecDeque<JWTDecodingKey>,
}

impl JWTStoreState {
    fn new(encoding_key: JWTEncodingKey, decoding_keys: VecDeque<JWTDecodingKey>) -> Self {
        Self {
            encoding_key,
            decoding_keys,
        }
    }
}

/// Store for parsed JWT encoding and decoding keys.
struct JWTStore {
    state: RwLock<JWTStoreState>,
}

impl JWTStore {
    pub fn new(encoding_key: JWTEncodingKey, decoding_keys: VecDeque<JWTDecodingKey>) -> Self {
        Self {
            state: RwLock::new(JWTStoreState::new(encoding_key, decoding_keys)),
        }
    }

    /// Returns the current encoding key.
    async fn get_encoding_key(&self) -> JWTEncodingKey {
        self.state.read().await.encoding_key.clone()
    }

    /// Returns the decoding key for the given id, if one exists.
    async fn get_decoding_key(&self, id: Uuid) -> Option<JWTDecodingKey> {
        self.state
            .read()
            .await
            .decoding_keys
            .iter()
            .find(|k| k.id == id)
            .cloned()
    }

    /// Stores the given encoding and decoding keys in the store.
    ///
    /// If there are more than 10 decoding keys, the oldest one is removed.
    async fn store_keys(&self, encoding_key: JWTEncodingKey, decoding_key: JWTDecodingKey) {
        let mut state = self.state.write().await;
        state.encoding_key = encoding_key;
        state.decoding_keys.push_back(decoding_key);
        if state.decoding_keys.len() > 10 {
            state.decoding_keys.pop_front();
        }
    }
}

/// Claims for encoding a JWT.
#[derive(Debug, Serialize)]
struct JWTEncodingClaims<'a> {
    kid: Uuid,
    sub: Uuid,
    exp: i64,
    iat: i64,
    role: Role,
    aud: &'a str,
    iss: &'a str,
}

impl<'a> JWTEncodingClaims<'a> {
    fn new(token: &AccessToken, aud: &'a str, iss: &'a str) -> Self {
        Self {
            kid: token.id,
            sub: token.user_id,
            exp: token.expires_at.unix_timestamp(),
            iat: OffsetDateTime::now_utc().unix_timestamp(),
            role: token.user_role,
            aud,
            iss,
        }
    }
}

/// Claims for decoding a JWT.
#[derive(Debug, Deserialize)]
struct JWTDecodingClaims {
    kid: Uuid,
    sub: Uuid,
    exp: i64,
    role: Role,
}

/// Implementation of [TokenCoder] using JWT.
pub struct JWTCoder {
    store: JWTStore,
    audience: String,
    issuer: String,
}

impl JWTCoder {
    pub fn new(keys: KeyPair, audience: String, issuer: String) -> Result<Self, Error> {
        let encoding_key = EncodingKey::from_rsa_pem(keys.private_key.as_bytes())
            .context("Failed to parse private key")?;
        let decoding_key = DecodingKey::from_rsa_pem(keys.public_key.as_bytes())
            .context("Failed to parse public key")?;

        let store = JWTStore::new(
            JWTEncodingKey::new(keys.id, encoding_key),
            [JWTDecodingKey::new(keys.id, decoding_key)].into(),
        );

        Ok(Self {
            store,
            audience,
            issuer,
        })
    }
}

#[async_trait::async_trait]
impl TokenCoder for JWTCoder {
    async fn encode(&self, token: &AccessToken) -> Result<String, Error> {
        let encoding_key = self.store.get_encoding_key().await;

        let mut header = Header::new(Algorithm::RS256);
        header.kid = Some(encoding_key.id.to_string());

        let claims = JWTEncodingClaims::new(token, &self.audience, &self.issuer);
        let token =
            encode(&header, &claims, &encoding_key.key).context("Failed to encode token")?;

        Ok(token)
    }

    async fn decode(&self, token: &str) -> Result<AccessToken, Error> {
        let header = decode_header(token).map_err(|_| Error::InvalidToken)?;
        let kid: Uuid = header
            .kid
            .ok_or(Error::InvalidToken)?
            .parse()
            .map_err(|_| Error::InvalidToken)?;

        let decoding_key = self
            .store
            .get_decoding_key(kid)
            .await
            .ok_or(Error::InvalidToken)?;

        let mut validation = Validation::new(Algorithm::RS256);
        validation.set_audience(&[&self.audience]);
        validation.set_issuer(&[&self.issuer]);

        let claims = decode::<JWTDecodingClaims>(token, &decoding_key.key, &validation)
            .map_err(|_| Error::InvalidToken)?
            .claims;

        let token = AccessToken::new(
            claims.kid,
            claims.sub,
            claims.role,
            OffsetDateTime::from_unix_timestamp(claims.exp).map_err(|_| Error::InvalidToken)?,
        );
        Ok(token)
    }

    async fn store_keys(&self, pair: KeyPair) -> Result<(), Error> {
        let encoding_key = EncodingKey::from_rsa_pem(pair.private_key.as_bytes())
            .context("Failed to parse private key")?;
        let decoding_key = DecodingKey::from_rsa_pem(pair.public_key.as_bytes())
            .context("Failed to parse public key")?;

        self.store
            .store_keys(
                JWTEncodingKey::new(pair.id, encoding_key),
                JWTDecodingKey::new(pair.id, decoding_key),
            )
            .await;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        domain::{KeyGenerator, TokenCoder},
        keys_generators::RSAKeyGenerator,
    };
    use time::{Duration, OffsetDateTime};

    #[tokio::test]
    async fn test_jwt_coder() {
        let key_generator = RSAKeyGenerator::new(2048);
        let key_pair = key_generator.generate().expect("Error generating key");

        let coder = JWTCoder::new(key_pair, "test-aud".into(), "test-iss".into())
            .expect("Error creating JWT coder");
        let token = AccessToken::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            Role::Admin,
            OffsetDateTime::now_utc() + Duration::days(2),
        );

        let encoded = coder.encode(&token).await.expect("Error encoding token");

        let decoded = coder.decode(&encoded).await.expect("Error decoding token");
        assert_eq!(token.id, decoded.id);
        assert_eq!(token.user_id, decoded.user_id);
        assert_eq!(token.user_role, decoded.user_role);
        assert_eq!(
            token.expires_at.unix_timestamp(),
            decoded.expires_at.unix_timestamp()
        );

        // Test that tokens using old encoding key are valid
        let key_pair = key_generator.generate().expect("Error generating key");
        coder
            .store_keys(key_pair)
            .await
            .expect("Error storing keys");

        let decoded = coder.decode(&encoded).await.expect("Error decoding token");
        assert_eq!(token.id, decoded.id);
        assert_eq!(token.user_id, decoded.user_id);
        assert_eq!(token.user_role, decoded.user_role);
        assert_eq!(
            token.expires_at.unix_timestamp(),
            decoded.expires_at.unix_timestamp()
        );
    }
}
