package api

import (
	"encoding/json"
	"testing"
)

func FuzzChatRequest(f *testing.F) {
	f.Add([]byte(`{"model": "llama3.2", "messages": [{"role": "user", "content": "hello"}], "stream": false}`))
	f.Add([]byte(`{"model": "gemma", "think": "high", "options": {"temperature": 0.5}}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var r ChatRequest
		_ = json.Unmarshal(data, &r)
	})
}

func FuzzGenerateRequest(f *testing.F) {
	f.Add([]byte(`{"model": "llama3.2", "prompt": "hello", "stream": false, "keep_alive": 300}`))
	f.Add([]byte(`{"model": "qwen", "think": "medium", "images": ["data:image/png;base64,abc"]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var r GenerateRequest
		_ = json.Unmarshal(data, &r)
	})
}

func FuzzToolCallFunctionArguments(f *testing.F) {
	f.Add([]byte(`{"arg1": "value1", "arg2": 42}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var r ToolCallFunctionArguments
		_ = json.Unmarshal(data, &r)
	})
}

func FuzzEmbedRequest(f *testing.F) {
	f.Add([]byte(`{"model": "nomic-embed-text", "input": ["hello", "world"], "truncate": true}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var r EmbedRequest
		_ = json.Unmarshal(data, &r)
	})
}
