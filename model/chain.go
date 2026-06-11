package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/lychee/lychee/api"
)

// ExecuteChain runs a composition request sequentially, substituting templates.
func ExecuteChain(ctx context.Context, req *api.ComposeRequest, runStep func(ctx context.Context, modelName string, prompt string, options map[string]any) (string, error)) (*api.ComposeResponse, error) {
	results := make([]api.StepResult, 0, len(req.Steps))
	currentInput := req.Input

	for i, step := range req.Steps {
		// Replace templates in the prompt
		prompt := step.Prompt
		prompt = strings.ReplaceAll(prompt, "{{input}}", currentInput)
		for j, res := range results {
			placeholder := fmt.Sprintf("{{step[%d].output}}", j)
			prompt = strings.ReplaceAll(prompt, placeholder, res.Output)
		}

		output, err := runStep(ctx, step.Model, prompt, step.Options)
		if err != nil {
			return nil, fmt.Errorf("step %d (%s) failed: %w", i, step.Model, err)
		}

		results = append(results, api.StepResult{
			Model:  step.Model,
			Output: output,
		})
		currentInput = output
	}

	finalOutput := ""
	if len(results) > 0 {
		finalOutput = results[len(results)-1].Output
	}

	return &api.ComposeResponse{
		Output:  finalOutput,
		Results: results,
	}, nil
}
