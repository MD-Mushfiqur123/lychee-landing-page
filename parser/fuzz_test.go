package parser

import (
	"bytes"
	"testing"
)

func FuzzParseFile(f *testing.F) {
	// Add some baseline corpus seed data
	f.Add([]byte("FROM llama3\nSYSTEM Hello\nPARAMETER temperature 0.7"))
	f.Add([]byte("FROM /path/to/model.gguf\nTEMPLATE \"{{ .System }}\n{{ .Prompt }}\""))
	f.Add([]byte("MESSAGE system \"you are a helpful assistant\"\nMESSAGE user \"hi\"\nFROM gemma"))
	
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseFile(bytes.NewReader(data))
	})
}
