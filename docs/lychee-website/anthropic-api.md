# Anthropic API Compatibility

Lychee includes a native `anthropic` translation layer, allowing you to use Local LLMs with SDKs and tools expecting Claude.

## Endpoint

**POST** `/anthropic/v1/messages`

## Usage with Python SDK

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:11434/anthropic",
    api_key="sk-not-needed",
)

message = client.messages.create(
    model="llama3", # Specify your local model
    max_tokens=1000,
    temperature=0,
    messages=[
        {
            "role": "user",
            "content": "Hello, Claude! Actually, you're a local Llama."
        }
    ]
)
print(message.content)
```
