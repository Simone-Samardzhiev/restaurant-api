use super::{AppState, errors::APIError};
use crate::domain::user::RegisterRequest as DomainRegisterRequest;
use axum::{
    extract::{Json, State},
    http::StatusCode,
};
use serde::Deserialize;
use validator::{Validate, ValidationError};

fn trim(s: &mut String) {
    s.truncate(s.trim_end().len());
    let leading_whitespace = s.len() - s.trim_start().len();
    if leading_whitespace > 0 {
        s.drain(..leading_whitespace);
    }
}

#[derive(Debug, Deserialize, Validate)]
pub struct RegisterRequest {
    #[validate(email(message = "Invalid email address."))]
    email: String,
    #[validate(length(
        min = 4,
        max = 128,
        message = "Username length must be between 4 and 128 characters."
    ))]
    username: String,

    #[validate(
        length(
            min = 8,
            max = 32,
            message = "Password length must be between 8 and 32 characters."
        ),
        custom(
            function = "Self::validate_password_complexity",
            message = "Password is not strong enought."
        )
    )]
    password: String,
}

impl RegisterRequest {
    fn sanitize(&mut self) {
        trim(&mut self.email);
        trim(&mut self.username);
    }

    fn validate_password_complexity(password: &str) -> Result<(), ValidationError> {
        let mut has_upper = false;
        let mut has_lower = false;
        let mut has_digit = false;
        let mut has_special = false;

        for c in password.chars() {
            if c.is_uppercase() {
                has_upper = true;
            }
            if c.is_lowercase() {
                has_lower = true;
            }
            if c.is_digit(10) {
                has_digit = true;
            }
            if c.is_ascii_punctuation() {
                has_special = true;
            }
        }

        if !has_upper || !has_lower || !has_digit || !has_special {
            return Err(ValidationError::new("password_complexity_error"));
        }

        Ok(())
    }
}

pub async fn register(
    State(state): State<AppState>,
    Json(mut req): Json<RegisterRequest>,
) -> Result<StatusCode, APIError> {
    req.sanitize();
    req.validate()?;
    state
        .user_service
        .register(DomainRegisterRequest::new(
            req.username,
            req.email,
            req.password,
        ))
        .await?;

    Ok(StatusCode::OK)
}
