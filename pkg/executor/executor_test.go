package executor

import (
	"bytes"
	"testing"
)

func TestWithJsonSchemaValidation_Success(t *testing.T) {
	schema := []byte(`{
		"type":"object",
		"properties":{
			"foo":{"type":"string"}
		},
		"required":["foo"]
	}`)
	validate := WithJsonSchemaValidation(schema)
	input := map[string]interface{}{"foo": "bar"}
	if err := validate(input); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWithJsonSchemaValidation_Failure(t *testing.T) {
	schema := []byte(`{
		"type":"object",
		"properties":{
			"foo":{"type":"string"}
		},
		"required":["foo"]
	}`)
	validate := WithJsonSchemaValidation(schema)
	// foo is wrong type
	input := map[string]interface{}{"foo": 123}
	if err := validate(input); err == nil {
		t.Fatal("expected validation error for wrong type, got nil")
	}
}

func TestWithJsonSchemaValidation_InvalidSchema(t *testing.T) {
	badSchema := []byte("not a valid schema")
	validate := WithJsonSchemaValidation(badSchema)
	if err := validate(nil); err == nil {
		t.Fatal("expected error compiling invalid schema, got nil")
	}
}

func TestExecute_InvalidYAML(t *testing.T) {
	input := []byte("key: : bad")
	tmpl := []byte("{{.key}}")
	var out bytes.Buffer
	if err := Execute(input, tmpl, &out); err == nil {
		t.Fatal("expected YAML unmarshal error, got nil")
	}
}

func TestExecute_NoValidation_Success(t *testing.T) {
	input := []byte("greeting: Hello\nname: World")
	tmpl := []byte("{{.greeting}}, {{.name}}!")
	var out bytes.Buffer
	if err := Execute(input, tmpl, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	want := "Hello, World!"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestExecute_WithValidation_Success(t *testing.T) {
	input := []byte("foo: baz")
	tmpl := []byte("Value: {{.foo}}")
	var out bytes.Buffer
	schema := []byte(`{
		"type":"object",
		"properties":{"foo":{"type":"string"}},
		"required":["foo"]
	}`)
	validate := WithJsonSchemaValidation(schema)
	if err := Execute(input, tmpl, &out, validate); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	want := "Value: baz"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestExecute_WithValidation_Failure(t *testing.T) {
	input := []byte("foo: 999")
	tmpl := []byte("{{.foo}}")
	var out bytes.Buffer
	schema := []byte(`{
		"type":"object",
		"properties":{"foo":{"type":"string"}},
		"required":["foo"]
	}`)
	validate := WithJsonSchemaValidation(schema)
	if err := Execute(input, tmpl, &out, validate); err == nil {
		t.Fatal("expected validation failure, got nil")
	}
}

func TestExecute_BadTemplate(t *testing.T) {
	input := []byte("key: value")
	// missing closing brace
	tmpl := []byte("{{.key")
	var out bytes.Buffer
	if err := Execute(input, tmpl, &out); err == nil {
		t.Fatal("expected template parse error, got nil")
	}
}
