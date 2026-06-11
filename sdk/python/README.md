# Lychee Python SDK

The official Python client for Lychee — the universal local inference engine.

## Installation

```bash
pip install lychee-ai
```

## Quick Start

```python
import lychee

# Simple chat completion
response = lychee.chat(
    model="gemma3",
    messages=[
        {"role": "user", "content": "Explain quantum physics in one sentence."}
    ]
)
print(response['message']['content'])
```

## Client Object

For custom configurations, use the `Client` object:

```python
from lychee import Client

client = Client(host="http://localhost:11434")

# Generate text
response = client.generate(
    model="gemma3",
    prompt="Write a poem about trees."
)
print(response['response'])

# List local models
models = client.list()
for model in models['models']:
    print(model['name'])
```
