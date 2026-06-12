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

    // Memory endpoints
    pub async fn list_conversations(&self, limit: Option<usize>, offset: Option<usize>) -> Result<Vec<ConversationSummary>, LycheeError> {
        let l = limit.unwrap_or(50);
        let o = offset.unwrap_or(0);
        let url = format!("{}/api/conversations?limit={}&offset={}", self.base_url, l, o);
        let res = self.client.get(&url).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        res.json::<Vec<ConversationSummary>>().await.map_err(LycheeError::Http)
    }

    pub async fn get_conversation(&self, id: &str) -> Result<Conversation, LycheeError> {
        let url = format!("{}/api/conversations/{}", self.base_url, id);
        let res = self.client.get(&url).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        res.json::<Conversation>().await.map_err(LycheeError::Http)
    }

    pub async fn delete_conversation(&self, id: &str) -> Result<(), LycheeError> {
        let url = format!("{}/api/conversations/{}", self.base_url, id);
        let res = self.client.delete(&url).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        Ok(())
    }

    // Router endpoints
    pub async fn create_route(&self, route: ModelRoute) -> Result<(), LycheeError> {
        let url = format!("{}/api/routes", self.base_url);
        let res = self.client.post(&url).json(&route).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        Ok(())
    }

    pub async fn list_routes(&self) -> Result<Vec<ModelRoute>, LycheeError> {
        let url = format!("{}/api/routes", self.base_url);
        let res = self.client.get(&url).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        res.json::<Vec<ModelRoute>>().await.map_err(LycheeError::Http)
    }

    pub async fn delete_route(&self, name: &str) -> Result<(), LycheeError> {
        let url = format!("{}/api/routes/{}", self.base_url, name);
        let res = self.client.delete(&url).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        Ok(())
    }

    // Structured Outputs
    pub async fn structured(&self, req: StructuredRequest) -> Result<StructuredResponse, LycheeError> {
        let url = format!("{}/api/structured", self.base_url);
        let res = self.client.post(&url).json(&req).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        res.json::<StructuredResponse>().await.map_err(LycheeError::Http)
    }

    // Compose Chaining
    pub async fn compose(&self, req: ComposeRequest) -> Result<ComposeResponse, LycheeError> {
        let url = format!("{}/api/compose", self.base_url);
        let mut r = req;
        r.stream = Some(false);
        let res = self.client.post(&url).json(&r).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        res.json::<ComposeResponse>().await.map_err(LycheeError::Http)
    }

    pub async fn compose_stream(&self, req: ComposeRequest) -> Result<impl Stream<Item = Result<ComposeEvent, LycheeError>>, LycheeError> {
        let url = format!("{}/api/compose", self.base_url);
        let mut r = req;
        r.stream = Some(true);
        let res = self.client.post(&url).json(&r).send().await?;
        if !res.status().is_success() {
            let body = res.text().await.unwrap_or_default();
            return Err(LycheeError::Api(body));
        }
        Ok(parse_sse_stream(res.bytes_stream()))
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

// Memory API types
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ConversationSummary {
    pub id: String,
    pub model: String,
    pub title: String,
    pub messages: i64,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Conversation {
    pub id: String,
    pub model: String,
    pub title: String,
    pub messages: Vec<Message>,
    pub created_at: String,
    pub updated_at: String,
}

// Router API types
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ModelEndpoint {
    pub host: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub weight: Option<i32>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ModelRoute {
    pub name: String,
    pub endpoints: Vec<ModelEndpoint>,
    pub strategy: String,
}

// Structured Output API types
#[derive(Serialize, Debug, Clone)]
pub struct StructuredRequest {
    pub model: String,
    pub prompt: String,
    pub schema: serde_json::Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_retries: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub options: Option<serde_json::Value>,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct StructuredResponse {
    pub model: String,
    pub output: String,
    pub valid: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

// Compose Chaining types
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ComposeCondition {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub contains: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub not_contains: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub min_length: Option<usize>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_length: Option<usize>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub always: Option<bool>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ComposeStep {
    pub model: String,
    pub prompt: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub options: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timeout_sec: Option<usize>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub fallback_model: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub parallel: Option<Vec<ComposeStep>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub condition: Option<ComposeCondition>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub skip_on_error: Option<bool>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ComposeRequest {
    pub input: String,
    pub steps: Vec<ComposeStep>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stream: Option<bool>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct StepResult {
    pub model: String,
    pub output: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub skipped: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub parallel_results: Option<Vec<StepResult>>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ComposeResponse {
    pub output: String,
    pub results: Vec<StepResult>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ComposeEvent {
    pub event: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub index: Option<usize>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub output: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub result: Option<ComposeResponse>,
}

fn parse_sse_stream<S>(
    stream: S,
) -> impl Stream<Item = Result<ComposeEvent, LycheeError>>
where
    S: Stream<Item = Result<Bytes, reqwest::Error>> + Send + 'static,
{
    let mut stream = Box::pin(stream);
    futures_util::stream::unfold(
        (stream, Vec::new()),
        |(mut stream, mut buffer)| async move {
            loop {
                if let Some(pos) = find_double_newline(&buffer) {
                    let message = buffer.drain(..pos).collect::<Vec<u8>>();
                    while !buffer.is_empty() && (buffer[0] == b'\n' || buffer[0] == b'\r') {
                        buffer.remove(0);
                    }
                    if let Some(data) = extract_sse_data(&message) {
                        match serde_json::from_slice::<ComposeEvent>(&data) {
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
                            if let Some(data) = extract_sse_data(&buffer) {
                                buffer.clear();
                                match serde_json::from_slice::<ComposeEvent>(&data) {
                                    Ok(val) => return Some((Ok(val), (stream, buffer))),
                                    Err(e) => return Some((Err(LycheeError::Json(e)), (stream, buffer))),
                                }
                            }
                            buffer.clear();
                        }
                        return None;
                    }
                }
            }
        },
    )
}

fn find_double_newline(buf: &[u8]) -> Option<usize> {
    for i in 0..buf.len().saturating_sub(1) {
        if buf[i] == b'\n' && buf[i+1] == b'\n' {
            return Some(i + 2);
        }
        if i + 3 <= buf.len() && &buf[i..i+4] == b"\r\n\r\n" {
            return Some(i + 4);
        }
    }
    None
}

fn extract_sse_data(msg: &[u8]) -> Option<Vec<u8>> {
    let text = String::from_utf8_lossy(msg);
    for line in text.lines() {
        if let Some(data_str) = line.strip_prefix("data:") {
            return Some(data_str.trim().as_bytes().to_vec());
        }
    }
    None
}
