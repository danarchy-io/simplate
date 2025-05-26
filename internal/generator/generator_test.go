package generator

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGenerate_ValidTemplate(t *testing.T) {
	templateContent := []byte("Hello, {{.Name}}! Env: {{env \"FOO_ENV\"}}")
	input := map[string]interface{}{"Name": "World"}
	os.Setenv("FOO_ENV", "bar")
	defer os.Unsetenv("FOO_ENV")

	var buf bytes.Buffer
	err := Generate(input, templateContent, &buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Hello, World!") || !strings.Contains(output, "Env: bar") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestGenerate_InvalidTemplate(t *testing.T) {
	templateContent := []byte("Hello, {{.Name") // missing closing }}
	input := map[string]interface{}{"Name": "World"}
	var buf bytes.Buffer
	err := Generate(input, templateContent, &buf)
	if err == nil {
		t.Fatal("expected error for invalid template, got nil")
	}
}

func TestGenerate_MissingVariable(t *testing.T) {
	templateContent := []byte("Hello, {{.Missing}}!")
	input := map[string]interface{}{"Name": "World"}
	var buf bytes.Buffer
	err := Generate(input, templateContent, &buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "<no value>") {
		t.Errorf("expected '<no value>' in output, got %q", output)
	}
}

func TestGenerate_NilInput(t *testing.T) {
	templateContent := []byte("Static content only")
	var buf bytes.Buffer
	err := Generate(nil, templateContent, &buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buf.String() != "Static content only" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}
