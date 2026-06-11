package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/types/model"
)

// StructuredOpts configures structured output generation with retries.
type StructuredOpts struct {
	Model      string
	Prompt     string
	Schema     json.RawMessage
	MaxRetries int // default: 3
	Options    map[string]any
}

// StructuredResult contains the generation result and retry metadata.
type StructuredResult struct {
	Output   string   `json:"output"`
	Valid    bool     `json:"valid"`
	Attempts int      `json:"attempts"`
	Errors   []string `json:"errors,omitempty"`
}

// generateStructured runs generation with schema validation and auto-retry.
// On each failed attempt, it constructs a correction prompt containing
// the previous output and validation error, asking the model to fix it.
func (s *Server) generateStructured(ctx context.Context, opts StructuredOpts) (*StructuredResult, error) {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}

	modelRef, err := parseAndValidateModelRef(opts.Model)
	if err != nil {
		return nil, err
	}
	name := modelRef.Name
	name, err = getExistingName(name)
	if err != nil {
		return nil, err
	}
	m, err := GetModel(name.String())
	if err != nil {
		return nil, err
	}

	caps := []model.Capability{model.CapabilityCompletion}
	r, m, runOpts, err := s.scheduleRunner(ctx, name.String(), caps, opts.Options, nil, nil)
	if err != nil {
		return nil, err
	}

	currentPrompt := opts.Prompt
	var errorsList []string
	var lastOutput string

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		var responseSB strings.Builder
		leadingBOS := leadingBOSForModel(m)

		err = r.Completion(ctx, llm.CompletionRequest{
			Prompt:     currentPrompt,
			Options:    runOpts,
			LeadingBOS: leadingBOS,
		}, func(cr llm.CompletionResponse) {
			responseSB.WriteString(cr.Content)
		})
		if err != nil {
			return nil, err
		}

		lastOutput = responseSB.String()

		// If Schema is empty/nil/empty object/null, bypass validation and return success
		if len(opts.Schema) == 0 || string(opts.Schema) == "null" || string(opts.Schema) == "{}" || string(opts.Schema) == "" {
			return &StructuredResult{
				Output:   lastOutput,
				Valid:    true,
				Attempts: attempt,
			}, nil
		}

		// Validate schema
		valErr := ValidateJSONSchema(lastOutput, opts.Schema)
		if valErr == nil {
			return &StructuredResult{
				Output:   lastOutput,
				Valid:    true,
				Attempts: attempt,
			}, nil
		}

		errorsList = append(errorsList, valErr.Error())

		// Construct correction prompt
		currentPrompt = fmt.Sprintf(
			"%s\n\nYour previous response was invalid JSON or did not conform to the schema.\nError: %s\nPrevious response: %s\nPlease output corrected JSON matching the schema.",
			opts.Prompt, valErr.Error(), lastOutput,
		)
	}

	return &StructuredResult{
		Output:   lastOutput,
		Valid:    false,
		Attempts: opts.MaxRetries,
		Errors:   errorsList,
	}, nil
}
