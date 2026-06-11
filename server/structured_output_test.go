package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/llm"
)

func TestStructuredOutput(t *testing.T) {
	t.Setenv("LYCHEE_MODELS", t.TempDir())
	gin.SetMode(gin.TestMode)

	t.Run("first attempt valid schema", func(t *testing.T) {
		mock := mockRunner{
			CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
				fn(llm.CompletionResponse{
					Content: `{"name":"Alice","age":30}`,
					Done:    true,
				})
				return nil
			},
		}

		s := newServerWithMockRunner(t, &mock)
		createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

		r := gin.New()
		r.POST("/api/structured", s.StructuredHandler)

		reqBody := api.StructuredRequest{
			Model:      "test-model",
			Prompt:     "give me a profile for Alice",
			Schema:     json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"]}`),
			MaxRetries: 3,
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/structured", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.StructuredResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if !resp.Valid {
			t.Error("expected output to be valid")
		}
		if resp.Attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", resp.Attempts)
		}
		if len(resp.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(resp.Errors))
		}
	})

	t.Run("retry correction succeeds", func(t *testing.T) {
		attempts := 0
		mock := mockRunner{
			CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
				attempts++
				content := `{"name":"Bob"}` // missing "age"
				if attempts > 1 {
					if !strings.Contains(req.Prompt, "previous response was invalid") {
						t.Error("expected retry prompt to contain error context")
					}
					content = `{"name":"Bob","age":40}`
				}
				fn(llm.CompletionResponse{
					Content: content,
					Done:    true,
				})
				return nil
			},
		}

		s := newServerWithMockRunner(t, &mock)
		createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

		r := gin.New()
		r.POST("/api/structured", s.StructuredHandler)

		reqBody := api.StructuredRequest{
			Model:      "test-model",
			Prompt:     "give me a profile for Bob",
			Schema:     json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"]}`),
			MaxRetries: 3,
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/structured", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.StructuredResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if !resp.Valid {
			t.Error("expected output to be valid after retry")
		}
		if resp.Attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", resp.Attempts)
		}
		if len(resp.Errors) != 1 {
			t.Errorf("expected 1 error list item, got %d", len(resp.Errors))
		}
	})

	t.Run("exhaust retries", func(t *testing.T) {
		mock := mockRunner{
			CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
				fn(llm.CompletionResponse{
					Content: `{"invalid": true}`,
					Done:    true,
				})
				return nil
			},
		}

		s := newServerWithMockRunner(t, &mock)
		createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

		r := gin.New()
		r.POST("/api/structured", s.StructuredHandler)

		reqBody := api.StructuredRequest{
			Model:      "test-model",
			Prompt:     "give me a profile for Bob",
			Schema:     json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"]}`),
			MaxRetries: 2,
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/structured", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.StructuredResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Valid {
			t.Error("expected output to be invalid")
		}
		if resp.Attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", resp.Attempts)
		}
		if len(resp.Errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(resp.Errors))
		}
	})

	t.Run("bypass schema validation", func(t *testing.T) {
		mock := mockRunner{
			CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
				fn(llm.CompletionResponse{
					Content: `plain text output`,
					Done:    true,
				})
				return nil
			},
		}

		s := newServerWithMockRunner(t, &mock)
		createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

		r := gin.New()
		r.POST("/api/structured", s.StructuredHandler)

		reqBody := api.StructuredRequest{
			Model:      "test-model",
			Prompt:     "say something",
			Schema:     nil,
			MaxRetries: 3,
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/structured", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.StructuredResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if !resp.Valid {
			t.Error("expected output to be valid (schema is nil)")
		}
		if resp.Attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", resp.Attempts)
		}
	})
}
