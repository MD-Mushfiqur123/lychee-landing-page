package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
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

	return validateValue(data, schema, "", 0, schema)
}

func resolveRef(ref string, root map[string]any) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("external or relative refs not supported: %s", ref)
	}

	parts := strings.Split(ref[2:], "/")
	var current any = root

	for _, part := range parts {
		// Escape JSON pointer characters
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		switch c := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = c[part]
			if !ok {
				return nil, fmt.Errorf("reference %q: property %q not found", ref, part)
			}
		default:
			return nil, fmt.Errorf("reference %q: cannot traverse non-object type %T at %q", ref, current, part)
		}
	}

	resolved, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("reference %q did not resolve to a valid schema object", ref)
	}

	return resolved, nil
}

func validateValue(val any, schema map[string]any, path string, depth int, root map[string]any) error {
	if depth > 32 {
		return fmt.Errorf("schema nesting depth limit exceeded")
	}

	if schema == nil {
		return nil
	}

	// Handle $ref
	if refVal, ok := schema["$ref"].(string); ok {
		resolvedSchema, err := resolveRef(refVal, root)
		if err != nil {
			return err
		}
		return validateValue(val, resolvedSchema, path, depth+1, root)
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
							if err := validateValue(propVal, propSchema, newPath, depth+1, root); err != nil {
								return err
							}
						}
					}
				}
			}
			// Validate additionalProperties
			if addProps, exists := schema["additionalProperties"]; exists {
				if addPropsBool, ok := addProps.(bool); ok && !addPropsBool {
					allowedProps := make(map[string]bool)
					if props, ok := schema["properties"].(map[string]any); ok {
						for k := range props {
							allowedProps[k] = true
						}
					}
					for k := range obj {
						if !allowedProps[k] {
							return fmt.Errorf("path %q: additional property %q is not allowed", path, k)
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
					if err := validateValue(item, itemsSchemaVal, newPath, depth+1, root); err != nil {
						return err
					}
				}
			}
		case "string":
			sVal, ok := val.(string)
			if !ok {
				return fmt.Errorf("path %q: expected string, got %T", path, val)
			}
			if minLength, ok := schema["minLength"].(float64); ok && float64(len(sVal)) < minLength {
				return fmt.Errorf("path %q: string length %d is less than minLength %g", path, len(sVal), minLength)
			}
			if maxLength, ok := schema["maxLength"].(float64); ok && float64(len(sVal)) > maxLength {
				return fmt.Errorf("path %q: string length %d is greater than maxLength %g", path, len(sVal), maxLength)
			}
			if patternStr, ok := schema["pattern"].(string); ok {
				matched, err := regexp.MatchString(patternStr, sVal)
				if err != nil {
					return fmt.Errorf("path %q: invalid pattern regex %q: %w", path, patternStr, err)
				}
				if !matched {
					return fmt.Errorf("path %q: string %q does not match pattern %q", path, sVal, patternStr)
				}
			}
		case "number", "integer":
			f, ok := val.(float64)
			if !ok {
				return fmt.Errorf("path %q: expected number, got %T", path, val)
			}
			if typeVal == "integer" && f != float64(int64(f)) {
				return fmt.Errorf("path %q: expected integer, got float %f", path, f)
			}
			if minimum, ok := schema["minimum"].(float64); ok && f < minimum {
				return fmt.Errorf("path %q: value %g is less than minimum %g", path, f, minimum)
			}
			if maximum, ok := schema["maximum"].(float64); ok && f > maximum {
				return fmt.Errorf("path %q: value %g is greater than maximum %g", path, f, maximum)
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

	// Validate allOf
	if allOf, ok := schema["allOf"].([]any); ok {
		for i, subSchemaVal := range allOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid allOf schema at index %d", i)
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err != nil {
				return fmt.Errorf("allOf schema validation failed at index %d: %w", i, err)
			}
		}
	}

	// Validate anyOf
	if anyOf, ok := schema["anyOf"].([]any); ok {
		matched := false
		var lastErr error
		for _, subSchemaVal := range anyOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				continue
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err == nil {
				matched = true
				break
			} else {
				lastErr = err
			}
		}
		if !matched {
			return fmt.Errorf("anyOf schema validation failed: value does not match any allowed schema (last error: %v)", lastErr)
		}
	}

	// Validate oneOf
	if oneOf, ok := schema["oneOf"].([]any); ok {
		matches := 0
		var lastErr error
		for _, subSchemaVal := range oneOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				continue
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err == nil {
				matches++
			} else {
				lastErr = err
			}
		}
		if matches != 1 {
			return fmt.Errorf("oneOf schema validation failed: expected exactly 1 match, found %d matches (last error: %v)", matches, lastErr)
		}
	}

	return nil
}
