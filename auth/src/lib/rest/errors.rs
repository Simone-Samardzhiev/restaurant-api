use crate::domain::error::Error;
use axum::{
    Json,
    http::StatusCode,
    response::{IntoResponse, Response},
};
use serde::Serialize;
use std::collections::HashMap;

/// Error returned from handler.
#[derive(Debug, thiserror::Error)]
pub enum APIError {
    #[error("Validation error: {0}")]
    Validation(#[from] validator::ValidationErrors),
    #[error("App error: {0}")]
    AppError(#[from] Error),
}

impl IntoResponse for APIError {
    fn into_response(self) -> Response {
        match self {
            APIError::Validation(e) => ValidationErrorResponse::from(e).into_response(),
            APIError::AppError(e) => APIError::AppError(e).into_response(),
        }
    }
}

/// Response from [Error].
#[derive(Debug, Serialize)]
struct AppErrorResponse {
    message: String,
    code: String,
    #[serde(serialize_with = "serialize_status_code")]
    status: StatusCode,
}

fn serialize_status_code<S>(status: &StatusCode, serializer: S) -> Result<S::Ok, S::Error>
where
    S: serde::Serializer,
{
    serializer.serialize_u16(status.as_u16())
}

impl AppErrorResponse {
    fn new(message: String, code: String, status: StatusCode) -> Self {
        Self {
            message,
            code,
            status,
        }
    }
}

impl From<Error> for AppErrorResponse {
    fn from(err: Error) -> Self {
        match err {
            Error::EmailAlreadyExists => Self::new(
                "EMAIL_CONFLICT".into(),
                "Email is already in use.".into(),
                StatusCode::CONFLICT.into(),
            ),
            Error::Internal { .. } => Self::new(
                "INTERNAL_ERROR".into(),
                "Internal error".into(),
                StatusCode::INTERNAL_SERVER_ERROR.into(),
            ),
        }
    }
}

impl IntoResponse for AppErrorResponse {
    fn into_response(self) -> Response {
        (self.status, Json(self)).into_response()
    }
}

/// Response from [validator::ValidationErrors].
#[derive(Debug, Serialize)]
struct ValidationErrorResponse {
    message: String,
    code: String,
    fields: HashMap<String, Vec<String>>,
}

impl ValidationErrorResponse {
    fn new(message: String, code: String, fields: HashMap<String, Vec<String>>) -> Self {
        Self {
            message,
            code,
            fields,
        }
    }
}

impl From<validator::ValidationErrors> for ValidationErrorResponse {
    fn from(errs: validator::ValidationErrors) -> Self {
        let mut fields: HashMap<String, Vec<String>> = HashMap::new();
        for (name, kind) in errs.errors() {
            match kind {
                validator::ValidationErrorsKind::Field(errors) => {
                    let messages: Vec<String> = errors
                        .iter()
                        .map(|e| {
                            e.message
                                .as_ref()
                                .map(|m| m.to_string())
                                .unwrap_or_else(|| e.code.to_string())
                        })
                        .collect();
                    fields.insert(name.to_string(), messages);
                }
                validator::ValidationErrorsKind::Struct(nested_errs) => {
                    let response = ValidationErrorResponse::from(*nested_errs.clone());
                    for (sub_name, sub_msgs) in response.fields {
                        fields.insert(format!("{}.{}", name, sub_name), sub_msgs);
                    }
                }
                validator::ValidationErrorsKind::List(nested_list) => {
                    for (index, nested_errs) in nested_list.iter() {
                        let sub_response = ValidationErrorResponse::from(*nested_errs.clone());
                        for (sub_field, sub_msgs) in sub_response.fields {
                            fields.insert(format!("{}[{}].{}", name, index, sub_field), sub_msgs);
                        }
                    }
                }
            }
        }

        Self::new(
            "Request data is invalid.".into(),
            "INVALID_ENTITY".into(),
            fields,
        )
    }
}

impl IntoResponse for ValidationErrorResponse {
    fn into_response(self) -> Response {
        (StatusCode::UNPROCESSABLE_ENTITY, Json(self)).into_response()
    }
}
