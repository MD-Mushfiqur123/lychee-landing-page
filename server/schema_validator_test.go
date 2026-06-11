package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateJSONSchema(t *testing.T) {
	t.Run("nil or empty schema", func(t *testing.T) {
		err := ValidateJSONSchema(`{"any":"thing"}`, nil)
		if err != nil {
			t.Errorf("expected nil error for nil schema, got %v", err)
		}

		err = ValidateJSONSchema(`{"any":"thing"}`, json.RawMessage(""))
		if err != nil {
			t.Errorf("expected nil error for empty schema, got %v", err)
		}
	})

	t.Run("literal json format formatStr", func(t *testing.T) {
		schema := json.RawMessage(`"json"`)

		err := ValidateJSONSchema(`{"valid": true}`, schema)
		if err != nil {
			t.Errorf("expected valid JSON to pass, got %v", err)
		}

		err = ValidateJSONSchema(`{"invalid": `, schema)
		if err == nil {
			t.Errorf("expected invalid JSON to fail")
		} else if !strings.Contains(err.Error(), "not valid JSON") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("basic types validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "integer"},
				"score": {"type": "number"},
				"active": {"type": "boolean"}
			},
			"required": ["name", "age"]
		}`)

		// Correct input
		err := ValidateJSONSchema(`{"name": "Alice", "age": 30, "score": 95.5, "active": true}`, schema)
		if err != nil {
			t.Errorf("expected valid schema to pass, got %v", err)
		}

		// Missing required property
		err = ValidateJSONSchema(`{"name": "Alice", "score": 95.5}`, schema)
		if err == nil {
			t.Errorf("expected failure for missing required property")
		} else if !strings.Contains(err.Error(), "missing required property") {
			t.Errorf("unexpected error: %v", err)
		}

		// Incorrect type (string instead of integer)
		err = ValidateJSONSchema(`{"name": "Alice", "age": "thirty"}`, schema)
		if err == nil {
			t.Errorf("expected failure for wrong type")
		} else if !strings.Contains(err.Error(), "expected integer") {
			t.Errorf("unexpected error: %v", err)
		}

		// Incorrect integer format (float value)
		err = ValidateJSONSchema(`{"name": "Alice", "age": 30.5}`, schema)
		if err == nil {
			t.Errorf("expected failure for non-integer float")
		} else if !strings.Contains(err.Error(), "expected integer") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("array validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "array",
			"items": {
				"type": "string"
			}
		}`)

		err := ValidateJSONSchema(`["one", "two", "three"]`, schema)
		if err != nil {
			t.Errorf("expected valid array to pass, got %v", err)
		}

		err = ValidateJSONSchema(`["one", 2, "three"]`, schema)
		if err == nil {
			t.Errorf("expected array validation failure on element type mismatch")
		} else if !strings.Contains(err.Error(), "expected string") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("enum validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "string",
			"enum": ["red", "green", "blue"]
		}`)

		err := ValidateJSONSchema(`"green"`, schema)
		if err != nil {
			t.Errorf("expected enum match to pass, got %v", err)
		}

		err = ValidateJSONSchema(`"yellow"`, schema)
		if err == nil {
			t.Errorf("expected enum mismatch to fail")
		} else if !strings.Contains(err.Error(), "not in enum") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nesting depth limit validation", func(t *testing.T) {
		// Generate a very deeply nested JSON object and matching schema to trigger depth limit
		var buildDeepSchema func(d int) string
		buildDeepSchema = func(d int) string {
			if d <= 0 {
				return `{"type": "string"}`
			}
			return `{"type": "object", "properties": {"next": ` + buildDeepSchema(d-1) + `}}`
		}

		schema := json.RawMessage(buildDeepSchema(35))

		// Construct corresponding deep JSON input
		var buildDeepJSON func(d int) string
		buildDeepJSON = func(d int) string {
			if d <= 0 {
				return `"end"`
			}
			return `{"next": ` + buildDeepJSON(d-1) + `}`
		}
		jsonData := buildDeepJSON(35)

		err := ValidateJSONSchema(jsonData, schema)
		if err == nil {
			t.Errorf("expected deep nesting depth limit to trigger failure")
		} else if !strings.Contains(err.Error(), "nesting depth limit exceeded") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("reference resolutions ($ref)", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"billing_address": {"$ref": "#/$defs/address"},
				"shipping_address": {"$ref": "#/$defs/address"}
			},
			"$defs": {
				"address": {
					"type": "object",
					"properties": {
						"city": {"type": "string"},
						"zip": {"type": "integer"}
					},
					"required": ["city"]
				}
			}
		}`)

		// Valid matching ref
		err := ValidateJSONSchema(`{"billing_address": {"city": "New York", "zip": 10001}}`, schema)
		if err != nil {
			t.Errorf("expected reference validation to pass, got %v", err)
		}

		// Invalid reference sub-property type
		err = ValidateJSONSchema(`{"billing_address": {"city": "New York", "zip": "not-a-zip"}}`, schema)
		if err == nil {
			t.Errorf("expected zip type mismatch to fail")
		} else if !strings.Contains(err.Error(), "expected integer") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("logical operators (allOf, anyOf, oneOf)", func(t *testing.T) {
		schemaAllOf := json.RawMessage(`{
			"allOf": [
				{"type": "object", "properties": {"first": {"type": "string"}}},
				{"type": "object", "properties": {"second": {"type": "integer"}}}
			]
		}`)

		// Valid allOf
		err := ValidateJSONSchema(`{"first": "hello", "second": 42}`, schemaAllOf)
		if err != nil {
			t.Errorf("expected allOf to pass, got %v", err)
		}

		// Invalid allOf (one of the schemas fails)
		err = ValidateJSONSchema(`{"first": 123, "second": 42}`, schemaAllOf)
		if err == nil {
			t.Errorf("expected allOf mismatch to fail")
		}

		schemaAnyOf := json.RawMessage(`{
			"anyOf": [
				{"type": "string"},
				{"type": "integer"}
			]
		}`)

		// Valid anyOf
		err = ValidateJSONSchema(`100`, schemaAnyOf)
		if err != nil {
			t.Errorf("expected anyOf integer to pass, got %v", err)
		}

		// Invalid anyOf
		err = ValidateJSONSchema(`true`, schemaAnyOf)
		if err == nil {
			t.Errorf("expected anyOf mismatch to fail")
		}

		schemaOneOf := json.RawMessage(`{
			"oneOf": [
				{"type": "integer", "minimum": 10},
				{"type": "integer", "maximum": 20}
			]
		}`)

		// Matches exactly one (value 5: matches maximum 20, but not minimum 10)
		err = ValidateJSONSchema(`5`, schemaOneOf)
		if err != nil {
			t.Errorf("expected oneOf with single match to pass, got %v", err)
		}

		// Matches both (value 15: matches both minimum 10 and maximum 20) -> should fail oneOf
		err = ValidateJSONSchema(`15`, schemaOneOf)
		if err == nil {
			t.Errorf("expected oneOf with double matches to fail")
		}
	})

	t.Run("strict validation constraints", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"code": {"type": "string", "minLength": 3, "maxLength": 5, "pattern": "^[A-Z]+$"},
				"val": {"type": "number", "minimum": 10.5, "maximum": 20.5}
			},
			"additionalProperties": false
		}`)

		// Valid input
		err := ValidateJSONSchema(`{"code": "ABC", "val": 15.2}`, schema)
		if err != nil {
			t.Errorf("expected valid constraints to pass, got %v", err)
		}

		// minLength fail
		err = ValidateJSONSchema(`{"code": "AB", "val": 15.2}`, schema)
		if err == nil {
			t.Errorf("expected minLength fail to error")
		}

		// pattern regex fail
		err = ValidateJSONSchema(`{"code": "abc", "val": 15.2}`, schema)
		if err == nil {
			t.Errorf("expected pattern match fail to error")
		}

		// maximum fail
		err = ValidateJSONSchema(`{"code": "ABC", "val": 25.0}`, schema)
		if err == nil {
			t.Errorf("expected maximum bounds check to error")
		}

		// additionalProperties fail
		err = ValidateJSONSchema(`{"code": "ABC", "val": 15.2, "extra": true}`, schema)
		if err == nil {
			t.Errorf("expected additionalProperties constraint to error")
		}
	})
}
