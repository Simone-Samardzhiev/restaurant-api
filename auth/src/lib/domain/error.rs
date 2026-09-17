// Enum describes possible domain errors.
#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("Email is already in use.")]
    EmailAlreadyExists,

    #[error("Invalid user role.")]
    InvalidUserRole,

    #[error("An internal error has occurred: {source}.")]
    Internal {
        #[from]
        source: anyhow::Error,
    },
}

impl Error {
    /// Returns unique error code for the specific error case.
    pub fn code(&self) -> String {
        match self {
            Self::EmailAlreadyExists => "EMAIL_CONFLICT".into(),
            Self::InvalidUserRole => "INVALID_USER_ROLE".into(),
            Self::Internal { source: _ } => "INTERNAL_ERROR".into(),
        }
    }
}
