package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lychee/lychee/api"
	_ "github.com/mattn/go-sqlite3"
)

// ConversationSummary provides high-level info about a saved conversation.
type ConversationSummary struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Title     string    `json:"title"`
	Messages  int       `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Conversation holds the full metadata and list of messages for a chat.
type Conversation struct {
	ID        string        `json:"id"`
	Model     string        `json:"model"`
	Messages  []api.Message `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// MemoryStore manages conversation persistence using JSON and SQLite.
type MemoryStore struct {
	dir    string
	dbPath string
	db     *sql.DB
	mu     sync.RWMutex
}

// NewMemoryStore initializes a new MemoryStore at the specified directory.
func NewMemoryStore(dir string) *MemoryStore {
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create conversation directory: %v\n", err)
	}

	dbPath := filepath.Join(dir, "conversations.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err == nil {
		schema := `
		CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			model TEXT,
			title TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT,
			role TEXT,
			content TEXT,
			images TEXT,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);`
		if _, err := db.Exec(schema); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize sqlite schema: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "failed to open sqlite db: %v\n", err)
	}

	return &MemoryStore{
		dir:    dir,
		dbPath: dbPath,
		db:     db,
	}
}

// Save writes a conversation to both SQLite and JSON file.
func (ms *MemoryStore) Save(conv *Conversation) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 1. Write to JSON
	jsonPath := filepath.Join(ms.dir, conv.ID+".json")
	jsonData, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	// 2. Write to SQLite
	if ms.db != nil {
		title := "New Conversation"
		for _, m := range conv.Messages {
			if m.Role == "user" && m.Content != "" {
				title = m.Content
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				break
			}
		}

		_, err = ms.db.Exec(`
			INSERT OR REPLACE INTO conversations (id, model, title, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`,
			conv.ID, conv.Model, title, conv.CreatedAt, conv.UpdatedAt,
		)
		if err != nil {
			return err
		}

		_, _ = ms.db.Exec("DELETE FROM messages WHERE conversation_id = ?", conv.ID)

		for _, msg := range conv.Messages {
			var imagesJSON []byte
			if len(msg.Images) > 0 {
				imagesJSON, _ = json.Marshal(msg.Images)
			}
			_, err = ms.db.Exec(`
				INSERT INTO messages (conversation_id, role, content, images)
				VALUES (?, ?, ?, ?)`,
				conv.ID, msg.Role, msg.Content, string(imagesJSON),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Load retrieves a conversation by ID, checking SQLite and falling back to JSON.
func (ms *MemoryStore) Load(id string) (*Conversation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.db != nil {
		var conv Conversation
		err := ms.db.QueryRow(`
			SELECT id, model, created_at, updated_at FROM conversations WHERE id = ?`,
			id,
		).Scan(&conv.ID, &conv.Model, &conv.CreatedAt, &conv.UpdatedAt)
		if err == nil {
			rows, err := ms.db.Query(`
				SELECT role, content, images FROM messages WHERE conversation_id = ? ORDER BY id ASC`,
				id,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var msg api.Message
					var imagesStr sql.NullString
					if err := rows.Scan(&msg.Role, &msg.Content, &imagesStr); err == nil {
						if imagesStr.Valid && imagesStr.String != "" {
							_ = json.Unmarshal([]byte(imagesStr.String), &msg.Images)
						}
						conv.Messages = append(conv.Messages, msg)
					}
				}
				return &conv, nil
			}
		}
	}

	jsonPath := filepath.Join(ms.dir, id+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// List returns a list of summaries for all saved conversations.
func (ms *MemoryStore) List() ([]ConversationSummary, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.db != nil {
		rows, err := ms.db.Query(`
			SELECT c.id, c.model, c.title, COUNT(m.id), c.created_at, c.updated_at
			FROM conversations c
			LEFT JOIN messages m ON c.id = m.conversation_id
			GROUP BY c.id
			ORDER BY c.updated_at DESC`,
		)
		if err == nil {
			defer rows.Close()
			var summaries []ConversationSummary
			for rows.Next() {
				var s ConversationSummary
				if err := rows.Scan(&s.ID, &s.Model, &s.Title, &s.Messages, &s.CreatedAt, &s.UpdatedAt); err == nil {
					summaries = append(summaries, s)
				}
			}
			return summaries, nil
		}
	}

	files, err := os.ReadDir(ms.dir)
	if err != nil {
		return nil, err
	}

	var summaries []ConversationSummary
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ms.dir, f.Name()))
		if err != nil {
			continue
		}
		var conv Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			continue
		}
		title := "New Conversation"
		for _, m := range conv.Messages {
			if m.Role == "user" && m.Content != "" {
				title = m.Content
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				break
			}
		}
		summaries = append(summaries, ConversationSummary{
			ID:        conv.ID,
			Model:     conv.Model,
			Title:     title,
			Messages:  len(conv.Messages),
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
		})
	}
	return summaries, nil
}

// Delete removes a conversation.
func (ms *MemoryStore) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.db != nil {
		_, _ = ms.db.Exec("DELETE FROM conversations WHERE id = ?", id)
		_, _ = ms.db.Exec("DELETE FROM messages WHERE conversation_id = ?", id)
	}

	jsonPath := filepath.Join(ms.dir, id+".json")
	return os.Remove(jsonPath)
}

// AppendMessage appends a single message to a conversation.
func (ms *MemoryStore) AppendMessage(id string, msg api.Message) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var conv *Conversation
	jsonPath := filepath.Join(ms.dir, id+".json")
	data, err := os.ReadFile(jsonPath)
	if err == nil {
		var c Conversation
		if err := json.Unmarshal(data, &c); err == nil {
			conv = &c
		}
	}

	if conv == nil && ms.db != nil {
		var c Conversation
		err := ms.db.QueryRow(`
			SELECT id, model, created_at, updated_at FROM conversations WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Model, &c.CreatedAt, &c.UpdatedAt)
		if err == nil {
			rows, err := ms.db.Query(`
				SELECT role, content, images FROM messages WHERE conversation_id = ? ORDER BY id ASC`,
				id,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var m api.Message
					var imagesStr sql.NullString
					if err := rows.Scan(&m.Role, &m.Content, &imagesStr); err == nil {
						if imagesStr.Valid && imagesStr.String != "" {
							_ = json.Unmarshal([]byte(imagesStr.String), &m.Images)
						}
						c.Messages = append(c.Messages, m)
					}
				}
				conv = &c
			}
		}
	}

	if conv == nil {
		return fmt.Errorf("conversation not found: %s", id)
	}

	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()

	// Save JSON
	jsonData, err := json.MarshalIndent(conv, "", "  ")
	if err == nil {
		_ = os.WriteFile(jsonPath, jsonData, 0644)
	}

	// Update SQLite
	if ms.db != nil {
		_, _ = ms.db.Exec("UPDATE conversations SET updated_at = ? WHERE id = ?", conv.UpdatedAt, conv.ID)
		var imagesJSON []byte
		if len(msg.Images) > 0 {
			imagesJSON, _ = json.Marshal(msg.Images)
		}
		_, _ = ms.db.Exec(`
			INSERT INTO messages (conversation_id, role, content, images)
			VALUES (?, ?, ?, ?)`,
			conv.ID, msg.Role, msg.Content, string(imagesJSON),
		)
	}

	return nil
}
