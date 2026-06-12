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

use lychee_rs::{
    ConversationSummary, Conversation, ModelRoute, ModelEndpoint,
    StructuredRequest, StructuredResponse, ComposeRequest, ComposeStep, ComposeEvent
};

#[tokio::test]
async fn test_conversations_api() {
    let mut server = Server::new_async().await;
    let list_mock = server.mock("GET", "/api/conversations?limit=2&offset=0")
        .with_status(200)
        .with_body(r#"[{"id":"conv-123","model":"gemma3","title":"quantum","messages":2,"created_at":"","updated_at":""}]"#)
        .create_async().await;

    let get_mock = server.mock("GET", "/api/conversations/conv-123")
        .with_status(200)
        .with_body(r#"{"id":"conv-123","model":"gemma3","title":"quantum","messages":[],"created_at":"","updated_at":""}"#)
        .create_async().await;

    let del_mock = server.mock("DELETE", "/api/conversations/conv-123")
        .with_status(200)
        .create_async().await;

    let client = Lychee::new_with_host(server.url());
    let list = client.list_conversations(Some(2), Some(0)).await.unwrap();
    assert_eq!(list.len(), 1);
    assert_eq!(list[0].id, "conv-123");

    let conv = client.get_conversation("conv-123").await.unwrap();
    assert_eq!(conv.id, "conv-123");

    client.delete_conversation("conv-123").await.unwrap();

    list_mock.assert_async().await;
    get_mock.assert_async().await;
    del_mock.assert_async().await;
}

#[tokio::test]
async fn test_routes_api() {
    let mut server = Server::new_async().await;
    let create_mock = server.mock("POST", "/api/routes")
        .with_status(200)
        .create_async().await;

    let list_mock = server.mock("GET", "/api/routes")
        .with_status(200)
        .with_body(r#"[{"name":"fast","endpoints":[{"host":"http://localhost:11434"}],"strategy":"random"}]"#)
        .create_async().await;

    let del_mock = server.mock("DELETE", "/api/routes/fast")
        .with_status(200)
        .create_async().await;

    let client = Lychee::new_with_host(server.url());
    
    let route = ModelRoute {
        name: "fast".to_string(),
        endpoints: vec![ModelEndpoint { host: "http://localhost:11434".to_string(), model: None, weight: None }],
        strategy: "random".to_string(),
    };
    client.create_route(route).await.unwrap();

    let routes = client.list_routes().await.unwrap();
    assert_eq!(routes.len(), 1);
    assert_eq!(routes[0].name, "fast");

    client.delete_route("fast").await.unwrap();

    create_mock.assert_async().await;
    list_mock.assert_async().await;
    del_mock.assert_async().await;
}

#[tokio::test]
async fn test_structured_api() {
    let mut server = Server::new_async().await;
    let mock = server.mock("POST", "/api/structured")
        .with_status(200)
        .with_body(r#"{"model":"gemma3","output":"{}","valid":true}"#)
        .create_async().await;

    let client = Lychee::new_with_host(server.url());
    let req = StructuredRequest {
        model: "gemma3".to_string(),
        prompt: "JSON profile".to_string(),
        schema: serde_json::json!({"type": "object"}),
        max_retries: Some(3),
        options: None,
    };
    let res = client.structured(req).await.unwrap();
    assert_eq!(res.output, "{}");
    assert!(res.valid);

    mock.assert_async().await;
}

#[tokio::test]
async fn test_compose_api() {
    let mut server = Server::new_async().await;
    let mock = server.mock("POST", "/api/compose")
        .with_status(200)
        .with_body(r#"{"output":"final text","results":[{"model":"gemma3","output":"final text"}]}"#)
        .create_async().await;

    let client = Lychee::new_with_host(server.url());
    let req = ComposeRequest {
        input: "input".to_string(),
        steps: vec![ComposeStep {
            model: "gemma3".to_string(),
            prompt: "Refine: {{input}}".to_string(),
            options: None,
            timeout_sec: None,
            fallback_model: None,
            parallel: None,
            condition: None,
            skip_on_error: None,
        }],
        stream: Some(false),
    };
    let res = client.compose(req).await.unwrap();
    assert_eq!(res.output, "final text");

    mock.assert_async().await;
}
