use thiserror::Error;

#[derive(Error, Debug)]
pub enum LycheeError {
    #[error("http request failed: {0}")]
    Http(#[from] reqwest::Error),

    #[error("json parsing failed: {0}")]
    Json(#[from] serde_json::Error),

    #[error("stream error: {0}")]
    Stream(String),

    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    #[error("API error response: {0}")]
    Api(String),
}
