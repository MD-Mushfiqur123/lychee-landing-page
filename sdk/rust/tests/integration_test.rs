use lychee_rs::{Lychee, GenerateRequest, ChatRequest, Message, PullRequest};
use mockito::Server;
use futures_util::StreamExt;

#[tokio::test]
async fn test_generate() {
    let mut server = Server::new_async().await;
    let mock = server.mock("POST", "/api/generate")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"model":"test-model","response":"hello world","done":true}"#)
        .create_async()
        .await;

    let client = Lychee::new_with_host(server.url());
    let req = GenerateRequest {
        model: "test-model".to_string(),
        prompt: "hello".to_string(),
        stream: Some(false),
    };

    let res = client.generate(req).await.unwrap();
    assert_eq!(res.model, "test-model");
    assert_eq!(res.response, "hello world");
    assert!(res.done);

    mock.assert_async().await;
}

#[tokio::test]
async fn test_generate_stream() {
    let mut server = Server::new_async().await;
    let body = "{\"model\":\"test-model\",\"response\":\"hello\",\"done\":false}\n{\"model\":\"test-model\",\"response\":\" world\",\"done\":true}\n";
    let mock = server.mock("POST", "/api/generate")
        .with_status(200)
        .with_header("content-type", "application/x-ndjson")
        .with_body(body)
        .create_async()
        .await;

    let client = Lychee::new_with_host(server.url());
    let req = GenerateRequest {
        model: "test-model".to_string(),
        prompt: "hello".to_string(),
        stream: Some(true),
    };

    let mut stream = client.generate_stream(req).await.unwrap();
    let mut responses = Vec::new();
    while let Some(res) = stream.next().await {
        responses.push(res.unwrap());
    }

    assert_eq!(responses.len(), 2);
    assert_eq!(responses[0].response, "hello");
    assert!(!responses[0].done);
    assert_eq!(responses[1].response, " world");
    assert!(responses[1].done);

    mock.assert_async().await;
}

#[tokio::test]
async fn test_chat() {
    let mut server = Server::new_async().await;
    let mock = server.mock("POST", "/api/chat")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"model":"test-model","message":{"role":"assistant","content":"response content"},"done":true}"#)
        .create_async()
        .await;

    let client = Lychee::new_with_host(server.url());
    let req = ChatRequest {
        model: "test-model".to_string(),
        messages: vec![Message {
            role: "user".to_string(),
            content: "hello".to_string(),
        }],
        stream: Some(false),
    };

    let res = client.chat(req).await.unwrap();
    assert_eq!(res.model, "test-model");
    assert_eq!(res.message.role, "assistant");
    assert_eq!(res.message.content, "response content");
    assert!(res.done);

    mock.assert_async().await;
}

#[tokio::test]
async fn test_list() {
    let mut server = Server::new_async().await;
    let mock = server.mock("GET", "/api/tags")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"models":[{"name":"test:latest","model":"test:latest","size":123456,"digest":"sha256:123"}]}"#)
        .create_async()
        .await;

    let client = Lychee::new_with_host(server.url());
    let res = client.list().await.unwrap();
    assert_eq!(res.models.len(), 1);
    assert_eq!(res.models[0].name, "test:latest");
    assert_eq!(res.models[0].size, 123456);

    mock.assert_async().await;
}
