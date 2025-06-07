package template

import (
	"bytes"
	"reflect"
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
	if err := Execute(YamlProvider(input), tmpl, &out); err == nil {
		t.Fatal("expected YAML unmarshal error, got nil")
	}
}

func TestExecute_NoValidation_Success(t *testing.T) {
	input := []byte("greeting: Hello\nname: World")
	tmpl := []byte("{{.greeting}}, {{.name}}!")
	var out bytes.Buffer
	if err := Execute(YamlProvider(input), tmpl, &out); err != nil {
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
	if err := Execute(YamlProvider(input), tmpl, &out, validate); err != nil {
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
	if err := Execute(YamlProvider(input), tmpl, &out, validate); err == nil {
		t.Fatal("expected validation failure, got nil")
	}
}

func TestExecute_BadTemplate(t *testing.T) {
	input := []byte("key: value")
	// missing closing brace
	tmpl := []byte("{{.key")
	var out bytes.Buffer
	if err := Execute(YamlProvider(input), tmpl, &out); err == nil {
		t.Fatal("expected template parse error, got nil")
	}
}

// TestAnyProvider_Success verifies that AnyProvider returns the original value.
func TestAnyProvider_Success(t *testing.T) {
	input := map[string]interface{}{"foo": "bar"}
	provider := AnyProvider(input)
	got, err := provider()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Errorf("expected %v, got %v", input, got)
	}
}

// TestAnyProvider_Nil verifies that AnyProvider(nil) returns an error.
func TestAnyProvider_Nil(t *testing.T) {
	provider := AnyProvider(nil)
	if _, err := provider(); err == nil {
		t.Fatal("expected error for nil input, got nil")
	}
}

// TestExecute_WithAnyProvider ensures Execute works with AnyProvider.
func TestExecute_WithAnyProvider(t *testing.T) {
	data := map[string]interface{}{"greeting": "Hi", "name": "Tester"}
	tmpl := []byte("{{.greeting}} {{.name}}")
	var out bytes.Buffer
	if err := Execute(AnyProvider(data), tmpl, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Hi Tester"
	if got := out.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
