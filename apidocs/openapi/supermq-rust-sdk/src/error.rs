use std::fmt;

#[derive(Debug)]
pub enum Error {
    Http(reqwest::Error),
    Serialization(serde_json::Error),
    InvalidInput(String),
    Unauthorized,
    NotFound,
    ServerError(String),
    ResponseError(ResponseContent<serde_json::Value>),
}

#[derive(Debug)]
pub struct ResponseContent<T> {
    pub status: reqwest::StatusCode,
    pub content: String,
    pub entity: Option<T>,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Error::Http(e) => write!(f, "HTTP error: {}", e),
            Error::Serialization(e) => write!(f, "Serialization error: {}", e),
            Error::InvalidInput(msg) => write!(f, "Invalid input: {}", msg),
            Error::Unauthorized => write!(f, "Unauthorized"),
            Error::NotFound => write!(f, "Not found"),
            Error::ServerError(msg) => write!(f, "Server error: {}", msg),
            Error::ResponseError(resp) => write!(f, "Response error: {} - {}", resp.status, resp.content),
        }
    }
}

impl std::error::Error for Error {}

impl From<reqwest::Error> for Error {
    fn from(error: reqwest::Error) -> Self {
        Error::Http(error)
    }
}

impl From<serde_json::Error> for Error {
    fn from(error: serde_json::Error) -> Self {
        Error::Serialization(error)
    }
}

pub type Result<T> = std::result::Result<T, Error>;
