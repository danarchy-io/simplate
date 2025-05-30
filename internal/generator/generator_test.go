package generator

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGenerate_ValidTemplate(t *testing.T) {
	templateContent := []byte("Hello, {{.Name}}! Env: {{env \"FOO_ENV\"}}")
	input := map[string]any{"Name": "World"}
	os.Setenv("FOO_ENV", "bar")
	defer os.Unsetenv("FOO_ENV")

	var buf bytes.Buffer
	if err := Generate(input, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Hello, World!") || !strings.Contains(output, "Env: bar") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestGenerate_EnvOrDefault(t *testing.T) {
	templateContent := []byte("Val: {{envOrDefault \"MY_KEY\" \"defVal\"}}")
	os.Unsetenv("MY_KEY")
	var buf bytes.Buffer
	if err := Generate(nil, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "Val: defVal" {
		t.Errorf("expected 'Val: defVal', got %q", got)
	}

	os.Setenv("MY_KEY", "setVal")
	buf.Reset()
	if err := Generate(nil, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "Val: setVal" {
		t.Errorf("expected 'Val: setVal', got %q", got)
	}
}

func TestGenerate_MissingVariable(t *testing.T) {
	templateContent := []byte("Value: {{.Missing}}")
	input := map[string]any{"Name": "World"}
	var buf bytes.Buffer
	if err := Generate(input, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "<no value>") {
		t.Errorf("expected '<no value>' in output, got %q", buf.String())
	}
}

func TestGenerate_NilInput(t *testing.T) {
	templateContent := []byte("Static content only")
	var buf bytes.Buffer
	if err := Generate(nil, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buf.String() != "Static content only" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestGenerate_UseUniqueFunction(t *testing.T) {
	templateContent := []byte("Unique: {{range unique .Items}}{{.}},{{end}}")
	input := map[string]any{"Items": []int{1, 2, 2, 3, 1}}
	var buf bytes.Buffer
	if err := Generate(input, templateContent, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(strings.TrimSpace(buf.String()), "1,2,3,") {
		t.Errorf("expected '1,2,3,' in output, got %q", buf.String())
	}
}

func TestGenerate_InvalidTemplate(t *testing.T) {
	templateContent := []byte("Hello, {{.Name") // missing closing }}
	input := map[string]any{"Name": "World"}
	var buf bytes.Buffer
	err := Generate(input, templateContent, &buf)
	if err == nil || !strings.Contains(err.Error(), "failed to parse template") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
