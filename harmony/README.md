# 🎼 Harmony Chat Protocol Parser

The `harmony` package implements a high-performance streaming parser for the Harmony chat template format. This format is designed for next-generation local models that support multi-channel streaming (e.g. separate thinking/reasoning and content streams) and native tool routing.

## The Harmony Protocol Spec

The Harmony protocol utilizes specialized markers to structure model inputs and outputs:

- `<|start|>`: Begins a new block or message turn.
- `<|channel|>`: Specifies the target channel for the current stream of tokens.
- `<|message|>`: Denotes the start of the payload content.
- `<|end|>`: Closes the current message or block.

### Channel Types

1. **`analysis`**: Reserved for reasoning/thinking blocks. Tokens emitted on this channel represent the model's internal step-by-step cognitive process before generating a final answer.
2. **`final`**: The primary output channel containing the actual content returned to the user.
3. **`commentary`**: Optional feedback or conversational side-notes.

### Tool Calling Routing

Tools are requested via the `to=` parameter on the `<|channel|>` tag. For example:
```xml
<|start|>assistant<|channel|>analysis to=functions.search_web<|message|>{"query": "latest weather in New York"}
```
This routes the message content as arguments to the `search_web` tool function.

## Parser Design

The `HarmonyParser` is a streaming parser that decodes tokens incrementally. Because delimiters like `<|start|>` and `<|channel|>` can span multiple tokens, the parser uses a custom suffix overlap resolver (`overlap` function) to ensure that a split token boundary is never missed or corrupted in the stream.
