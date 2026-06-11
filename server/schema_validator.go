package server

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ValidateJSONSchema validates a JSON string or raw byte slice against a JSON Schema.
func ValidateJSONSchema(output string, schemaRaw json.RawMessage) error {
	if len(schemaRaw) == 0 {
		return nil
	}

	// If the format is just the string "json", verify it's valid JSON
	var formatStr string
	if err := json.Unmarshal(schemaRaw, &formatStr); err == nil && formatStr == "json" {
		if !json.Valid([]byte(output)) {
			return fmt.Errorf("output is not valid JSON")
		}
		return nil
	}

	// Otherwise, it must be a JSON Schema object
	var schema map[string]any
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return fmt.Errorf("invalid JSON schema definition: %w", err)
	}

	var data any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return fmt.Errorf("failed to parse output as JSON: %w", err)
	}

	return validateValue(data, schema, "", 0)
}

func validateValue(val any, schema map[string]any, path string, depth int) error {
	if depth > 32 {
		return fmt.Errorf("schema nesting depth limit exceeded")
	}

	if schema == nil {
		return nil
	}

	// Validate type
	if typeVal, ok := schema["type"].(string); ok {
		switch typeVal {
		case "object":
			obj, ok := val.(map[string]any)
			if !ok {
				return fmt.Errorf("path %q: expected object, got %T", path, val)
			}
			// Validate required properties
			if reqs, ok := schema["required"].([]any); ok {
				for _, r := range reqs {
					reqKey, ok := r.(string)
					if ok {
						if _, exists := obj[reqKey]; !exists {
							return fmt.Errorf("path %q: missing required property %q", path, reqKey)
						}
					}
				}
			}
			// Validate properties
			if props, ok := schema["properties"].(map[string]any); ok {
				for propKey, propSchemaVal := range props {
					propSchema, ok := propSchemaVal.(map[string]any)
					if ok {
						if propVal, exists := obj[propKey]; exists {
							newPath := propKey
							if path != "" {
								newPath = path + "." + propKey
							}
							if err := validateValue(propVal, propSchema, newPath, depth+1); err != nil {
								return err
							}
						}
					}
				}
			}
		case "array":
			arr, ok := val.([]any)
			if !ok {
				return fmt.Errorf("path %q: expected array, got %T", path, val)
			}
			if itemsSchemaVal, ok := schema["items"].(map[string]any); ok {
				for i, item := range arr {
					newPath := fmt.Sprintf("%s[%d]", path, i)
					if err := validateValue(item, itemsSchemaVal, newPath, depth+1); err != nil {
						return err
					}
				}
			}
		case "string":
			if _, ok := val.(string); !ok {
				return fmt.Errorf("path %q: expected string, got %T", path, val)
			}
		case "number":
			if _, ok := val.(float64); !ok {
				return fmt.Errorf("path %q: expected number, got %T", path, val)
			}
		case "integer":
			f, ok := val.(float64)
			if !ok {
				return fmt.Errorf("path %q: expected integer, got %T", path, val)
			}
			if f != float64(int64(f)) {
				return fmt.Errorf("path %q: expected integer, got float %f", path, f)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("path %q: expected boolean, got %T", path, val)
			}
		}
	}

	// Validate enum
	if enumVals, ok := schema["enum"].([]any); ok {
		matched := false
		for _, e := range enumVals {
			if reflect.DeepEqual(e, val) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("path %q: value %v not in enum %v", path, val, enumVals)
		}
	}

	return nil
}
