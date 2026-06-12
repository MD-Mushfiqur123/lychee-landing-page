package api

import "encoding/json"

type StructuredRequest struct {
	Model      string          `json:"model"`
	Prompt     string          `json:"prompt"`
	Schema     json.RawMessage `json:"schema"`
	MaxRetries int             `json:"max_retries,omitempty"` // default 3
	Options    map[string]any  `json:"options,omitempty"`
	Stream     bool            `json:"stream,omitempty"`
	TimeoutSec int             `json:"timeout_sec,omitempty"` // default 60
}

type StructuredResponse struct {
	Output   string   `json:"output"`
	Valid    bool     `json:"valid"`
	Attempts int      `json:"attempts"`
	Errors   []string `json:"errors,omitempty"`
}

type StructuredEvent struct {
	Event    string              `json:"event"`
	Attempt  int                 `json:"attempt,omitempty"`
	Output   string              `json:"output,omitempty"`
	Error    string              `json:"error,omitempty"`
	Response *StructuredResponse `json:"response,omitempty"`
}
