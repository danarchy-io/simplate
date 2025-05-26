package validator

import (
	"testing"
)

func TestValidateYamlWithSchema_Valid(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name", "age"]
	}`)
	data := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	err := ValidateYamlWithSchema(data, schema)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateYamlWithSchema_InvalidData(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name", "age"]
	}`)
	data := map[string]interface{}{
		"name": "Bob",
		"age":  "not-an-integer",
	}

	err := ValidateYamlWithSchema(data, schema)
	if err == nil {
		t.Fatal("expected error for invalid data, got nil")
	}
}

func TestValidateYamlWithSchema_InvalidSchema(t *testing.T) {
	invalidSchema := []byte(`{ "type": "object", "properties": { "foo": { "type": "unknown" } } }`)
	data := map[string]interface{}{
		"foo": "bar",
	}

	err := ValidateYamlWithSchema(data, invalidSchema)
	if err == nil {
		t.Fatal("expected error for invalid schema, got nil")
	}
}