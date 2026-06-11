package api

import "encoding/json"

type StructuredRequest struct {
	Model      string          `json:"model"`
	Prompt     string          `json:"prompt"`
	Schema     json.RawMessage `json:"schema"`
	MaxRetries int             `json:"max_retries,omitempty"` // default 3
	Options    map[string]any  `json:"options,omitempty"`
}

type StructuredResponse struct {
	Output   string   `json:"output"`
	Valid    bool     `json:"valid"`
	Attempts int      `json:"attempts"`
	Errors   []string `json:"errors,omitempty"`
}
