package harmony

import (
	"testing"
)

func FuzzHarmonyParser(f *testing.F) {
	// Seed corpus
	f.Add("<|start|>user<|message|>Hello<|end|>")
	f.Add("<|start|>assistant<|channel|>analysis<|message|>Thinking...<|end|>")
	f.Add("<|start|>assistant<|channel|>commentary to=functions.calc<|message|>{\"x\": 5}<|end|>")
	f.Add("<")
	f.Add("|")
	f.Add("start|>user<|message|>test")
	f.Add("assistant to=functions.get_weather<|channel|>analysis")

	f.Fuzz(func(t *testing.T, data string) {
		// Test parsing complete content
		parser := HarmonyParser{
			MessageStartTag: "<|start|>",
			MessageEndTag:   "<|end|>",
			HeaderEndTag:    "<|message|>",
		}
		_ = parser.AddContent(data)

		// Test streaming parsing by splitting content
		parser2 := HarmonyParser{
			MessageStartTag: "<|start|>",
			MessageEndTag:   "<|end|>",
			HeaderEndTag:    "<|message|>",
		}
		if len(data) > 1 {
			mid := len(data) / 2
			_ = parser2.AddContent(data[:mid])
			_ = parser2.AddContent(data[mid:])
		}
	})
}
