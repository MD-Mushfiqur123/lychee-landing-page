# 🎭 Anthropic Messages API Compatibility Layer

Lychee features a native translation layer for the Anthropic Messages API. This allows developers to point Claude-specific tools, pipelines, and SDKs (including Anthropic's official Python and TypeScript/Javascript SDKs) directly to local running models.

## Endpoint

- **POST** `http://localhost:11434/v1/messages`

## Supported Features

- **Single & Multi-turn Chat**: Full support for `user`, `assistant`, and `system` message sequences.
- **Streaming Responses**: Server-Sent Events (SSE) streaming with type `message_start`, `content_block_start`, `content_block_delta`, `message_delta`, `message_stop`.
- **System Prompts**: Pass system instructions via the top-level `system` parameter (either as a string or block list).
- **Tool Use (Function Calling)**: Define tools using the Anthropic schema and receive structured tool calls from models that support it.
- **Thinking Blocks**: Support for models that output reasoned/thinking content.
- **Vision (Multimodal Inputs)**: Pass base64-encoded image parameters (`image/jpeg`, `image/png`, `image/webp`, `image/gif`) to models.

## Usage Examples

### 1. Python SDK

To migrate existing codebase to run locally with Lychee, customize the `base_url` and `api_key`:

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:11434/v1",
    api_key="lychee-dummy-key" # Any value works
)

# Streaming request
with client.messages.stream(
    model="gemma3",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Write a short poem about the ocean."}
    ]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)
print()
```

### 2. Node.js SDK

```javascript
import Anthropic from '@anthropic-ai/sdk';

const anthropic = new Anthropic({
  baseURL: 'http://localhost:11434/v1',
  apiKey: 'lychee-dummy-key',
});

const msg = await anthropic.messages.create({
  model: 'gemma3',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'What is photosynthesis?' }],
});

console.log(msg.content[0].text);
```

### 3. Raw curl (JSON)

```bash
curl http://localhost:11434/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: lychee-dummy-key" \
  -d '{
    "model": "gemma3",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello Claude/Lychee!"}
    ]
  }'
```
