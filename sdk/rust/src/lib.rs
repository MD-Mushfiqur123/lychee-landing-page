pub mod error;
pub use error::LycheeError;

use reqwest::Client;
use serde::{Deserialize, Serialize};
use futures_util::stream::{Stream, StreamExt};
use bytes::Bytes;

#[derive(Clone)]
pub struct Lychee {
    client: Client,
    base_url: String,
}

impl Default for Lychee {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct GenerateRequest {
    pub model: String,
    pub prompt: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stream: Option<bool>,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct GenerateResponse {
    pub model: String,
    pub response: String,
    pub done: bool,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Message {
    pub role: String,
    pub content: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ChatRequest {
    pub model: String,
    pub messages: Vec<Message>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stream: Option<bool>,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct ChatResponse {
    pub model: String,
    pub message: Message,
    pub done: bool,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct ModelInfo {
    pub name: String,
    pub model: String,
    pub size: i64,
    pub digest: String,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct ListResponse {
    pub models: Vec<ModelInfo>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct PullRequest {
    pub model: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stream: Option<bool>,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct ProgressResponse {
    pub status: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub digest: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub total: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub completed: Option<i64>,
}

impl Lychee {
    pub fn new() -> Self {
        Self::new_with_host("http://localhost:11434".to_string())
    }

    pub fn new_with_host(host: String) -> Self {
        Self {
            client: Client::new(),
            base_url: host,
        }
    }

    pub async fn generate(&self, req: GenerateRequest) -> Result<GenerateResponse, LycheeError> {
        let url = format!("{}/api/generate", self.base_url);
        let mut request = req;
        request.stream = Some(false);
        let res = self.client.post(&url).json(&request).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        res.json::<GenerateResponse>().await.map_err(LycheeError::Http)
    }

    pub async fn generate_stream(&self, req: GenerateRequest) -> Result<impl Stream<Item = Result<GenerateResponse, LycheeError>>, LycheeError> {
        let url = format!("{}/api/generate", self.base_url);
        let mut request = req;
        request.stream = Some(true);
        let res = self.client.post(&url).json(&request).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        let stream = res.bytes_stream();
        Ok(parse_ndjson_stream(stream))
    }

    pub async fn chat(&self, req: ChatRequest) -> Result<ChatResponse, LycheeError> {
        let url = format!("{}/api/chat", self.base_url);
        let mut request = req;
        request.stream = Some(false);
        let res = self.client.post(&url).json(&request).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        res.json::<ChatResponse>().await.map_err(LycheeError::Http)
    }

    pub async fn chat_stream(&self, req: ChatRequest) -> Result<impl Stream<Item = Result<ChatResponse, LycheeError>>, LycheeError> {
        let url = format!("{}/api/chat", self.base_url);
        let mut request = req;
        request.stream = Some(true);
        let res = self.client.post(&url).json(&request).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        let stream = res.bytes_stream();
        Ok(parse_ndjson_stream(stream))
    }

    pub async fn list(&self) -> Result<ListResponse, LycheeError> {
        let url = format!("{}/api/tags", self.base_url);
        let res = self.client.get(&url).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        res.json::<ListResponse>().await.map_err(LycheeError::Http)
    }

    pub async fn pull(&self, req: PullRequest) -> Result<impl Stream<Item = Result<ProgressResponse, LycheeError>>, LycheeError> {
        let url = format!("{}/api/pull", self.base_url);
        let mut request = req;
        request.stream = Some(true);
        let res = self.client.post(&url).json(&request).send().await?;
        if !res.status().is_success() {
            let err_text = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(err_text));
        }
        let stream = res.bytes_stream();
        Ok(parse_ndjson_stream(stream))
    }
}

fn parse_ndjson_stream<T, S>(
    stream: S,
) -> impl Stream<Item = Result<T, LycheeError>>
where
    T: serde::de::DeserializeOwned + Send + 'static,
    S: Stream<Item = Result<Bytes, reqwest::Error>> + Send + 'static,
{
    let mut stream = Box::pin(stream);
    futures_util::stream::unfold(
        (stream, Vec::new()),
        |(mut stream, mut buffer)| async move {
            loop {
                if let Some(pos) = buffer.iter().position(|&b| b == b'\n') {
                    let line = buffer.drain(..=pos).collect::<Vec<u8>>();
                    if line.len() > 1 {
                        match serde_json::from_slice::<T>(&line[..line.len() - 1]) {
                            Ok(val) => return Some((Ok(val), (stream, buffer))),
                            Err(e) => return Some((Err(LycheeError::Json(e)), (stream, buffer))),
                        }
                    }
                }

                match stream.next().await {
                    Some(Ok(bytes)) => {
                        buffer.extend_from_slice(&bytes);
                    }
                    Some(Err(err)) => {
                        return Some((Err(LycheeError::Http(err)), (stream, buffer)));
                    }
                    None => {
                        if !buffer.is_empty() {
                            let val = serde_json::from_slice::<T>(&buffer);
                            buffer.clear();
                            match val {
                                Ok(val) => return Some((Ok(val), (stream, buffer))),
                                Err(e) => return Some((Err(LycheeError::Json(e)), (stream, buffer))),
                            }
                        }
                        return None;
                    }
                }
            }
        },
    )
}
