
#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("Email is already in use.")]
    EmailAlreadyExists,

    #[error("An internal error has occurred.")]
    Internal(#[from(anyhow::Error)] anyhow::Error),
}

impl Error {
    pub fn code(&self) -> String {
        match self {
            Self::EmailAlreadyExists => "EMAIL_CONFLICT".into(),
            Self::Internal(_) => "INTERNAL_ERROR".into(),
        }
    }
}
