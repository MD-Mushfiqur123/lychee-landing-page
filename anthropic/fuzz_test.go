package anthropic

import (
	"encoding/json"
	"testing"
)

func FuzzMessagesRequestUnmarshal(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(`{"model": "gemma3", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}]}`))
	f.Add([]byte(`{"model": "gemma3", "max_tokens": 1024, "messages": [{"role": "user", "content": [{"type": "text", "text": "Hello"}]}]}`))
	f.Add([]byte(`{"model": "gemma3", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}], "system": "You are a helpful assistant"}`))
	f.Add([]byte(`{"model": "gemma3", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}], "system": [{"type": "text", "text": "You are a helpful assistant"}]}`))
	f.Add([]byte(`invalid json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var req MessagesRequest
		_ = json.Unmarshal(data, &req)
	})
}
