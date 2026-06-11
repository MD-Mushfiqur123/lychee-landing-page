package server

import (
	"sync"
	"testing"
	"time"

	"github.com/lychee/lychee/api"
)

func TestMemoryStore(t *testing.T) {
	tempDir := t.TempDir()
	store := NewMemoryStore(tempDir)

	convID := "test-conv-123"
	conv := &Conversation{
		ID:    convID,
		Model: "gemma3",
		Messages: []api.Message{
			{Role: "user", Content: "Hello world!"},
			{Role: "assistant", Content: "Hi there! How can I help you today?"},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	t.Run("save and load roundtrip", func(t *testing.T) {
		err := store.Save(conv)
		if err != nil {
			t.Fatalf("failed to save conversation: %v", err)
		}

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load conversation: %v", err)
		}

		if loaded.ID != conv.ID {
			t.Errorf("expected ID %q, got %q", conv.ID, loaded.ID)
		}
		if loaded.Model != conv.Model {
			t.Errorf("expected Model %q, got %q", conv.Model, loaded.Model)
		}
		if len(loaded.Messages) != len(conv.Messages) {
			t.Fatalf("expected %d messages, got %d", len(conv.Messages), len(loaded.Messages))
		}
		if loaded.Messages[0].Content != conv.Messages[0].Content {
			t.Errorf("expected first message content %q, got %q", conv.Messages[0].Content, loaded.Messages[0].Content)
		}
	})

	t.Run("list conversations", func(t *testing.T) {
		list, err := store.List()
		if err != nil {
			t.Fatalf("failed to list conversations: %v", err)
		}

		if len(list) != 1 {
			t.Fatalf("expected 1 conversation summary, got %d", len(list))
		}

		summary := list[0]
		if summary.ID != convID {
			t.Errorf("expected ID %q, got %q", convID, summary.ID)
		}
		if summary.Title != "Hello world!" {
			t.Errorf("expected Title %q, got %q", "Hello world!", summary.Title)
		}
		if summary.Messages != 2 {
			t.Errorf("expected Messages count 2, got %d", summary.Messages)
		}
	})

	t.Run("append message", func(t *testing.T) {
		newMsg := api.Message{Role: "user", Content: "One more question."}
		err := store.AppendMessage(convID, newMsg)
		if err != nil {
			t.Fatalf("failed to append message: %v", err)
		}

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if len(loaded.Messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(loaded.Messages))
		}
		if loaded.Messages[2].Content != "One more question." {
			t.Errorf("expected last message content 'One more question.', got %q", loaded.Messages[2].Content)
		}
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(num int) {
				defer wg.Done()
				_ = store.AppendMessage(convID, api.Message{Role: "user", Content: "Concurrent message"})
				_, _ = store.List()
			}(i)
		}

		wg.Wait()

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}
		expectedLen := 3 + concurrency
		if len(loaded.Messages) != expectedLen {
			t.Errorf("expected %d messages, got %d", expectedLen, len(loaded.Messages))
		}
	})

	t.Run("delete conversation", func(t *testing.T) {
		err := store.Delete(convID)
		if err != nil {
			t.Fatalf("failed to delete conversation: %v", err)
		}

		_, err = store.Load(convID)
		if err == nil {
			t.Error("expected error when loading deleted conversation, got nil")
		}

		list, err := store.List()
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected 0 summaries, got %d", len(list))
		}
	})

	t.Run("load non-existent ID", func(t *testing.T) {
		_, err := store.Load("does-not-exist")
		if err == nil {
			t.Error("expected error loading non-existent conversation, got nil")
		}
	})
}
