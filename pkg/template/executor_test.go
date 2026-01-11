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

// TestExecuteWithFiles_NoFileDirectives verifies backward compatibility
// when template has no FILE directives.
func TestExecuteWithFiles_NoFileDirectives(t *testing.T) {
	data := map[string]interface{}{"name": "World"}
	tmpl := []byte("Hello {{.name}}!")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should render to stdout
	if stdout.String() != "Hello World!" {
		t.Errorf("expected stdout 'Hello World!', got %q", stdout.String())
	}

	// Should not create any files
	if len(memWriter.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(memWriter.Files))
	}
}

// TestExecuteWithFiles_SingleFile tests basic FILE directive functionality.
func TestExecuteWithFiles_SingleFile(t *testing.T) {
	data := map[string]interface{}{"name": "World"}
	tmpl := []byte("#FILE:output.txt#\nHello {{.name}}!\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stdout should be empty
	if stdout.String() != "" {
		t.Errorf("expected empty stdout, got %q", stdout.String())
	}

	// Should create one file
	if len(memWriter.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(memWriter.Files))
	}

	content, exists := memWriter.Files["output.txt"]
	if !exists {
		t.Fatal("expected file 'output.txt' to exist")
	}

	expected := "\nHello World!\n"
	if string(content) != expected {
		t.Errorf("expected file content %q, got %q", expected, content)
	}
}

// TestExecuteWithFiles_TemplateFilename tests filename template rendering.
func TestExecuteWithFiles_TemplateFilename(t *testing.T) {
	data := map[string]interface{}{"id": "123", "env": "prod"}
	tmpl := []byte("#FILE:config-{{.env}}-{{.id}}.yml#\nconfig: value\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFilename := "config-prod-123.yml"
	content, exists := memWriter.Files[expectedFilename]
	if !exists {
		t.Fatalf("expected file %q to exist, got files: %v", expectedFilename, memWriter.Files)
	}

	if string(content) != "\nconfig: value\n" {
		t.Errorf("unexpected file content: %q", content)
	}
}

// TestExecuteWithFiles_MixedContent tests stdout and FILE segments together.
func TestExecuteWithFiles_MixedContent(t *testing.T) {
	data := map[string]interface{}{"msg": "test"}
	tmpl := []byte("STDOUT: {{.msg}}\n#FILE:file.txt#\nFILE: {{.msg}}\n#FILE#\nMore stdout")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check stdout content
	expectedStdout := "STDOUT: test\n\nMore stdout"
	if stdout.String() != expectedStdout {
		t.Errorf("expected stdout %q, got %q", expectedStdout, stdout.String())
	}

	// Check file content
	fileContent, exists := memWriter.Files["file.txt"]
	if !exists {
		t.Fatal("expected file 'file.txt' to exist")
	}

	expectedFile := "\nFILE: test\n"
	if string(fileContent) != expectedFile {
		t.Errorf("expected file content %q, got %q", expectedFile, fileContent)
	}
}

// TestExecuteWithFiles_MultipleFiles tests multiple FILE directives.
func TestExecuteWithFiles_MultipleFiles(t *testing.T) {
	data := map[string]interface{}{"app": "myapp"}
	tmpl := []byte(`#FILE:config.yml#
app: {{.app}}
#FILE#
#FILE:readme.txt#
App name: {{.app}}
#FILE#`)
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(memWriter.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(memWriter.Files))
	}

	// Check config.yml
	configContent := string(memWriter.Files["config.yml"])
	if configContent != "\napp: myapp\n" {
		t.Errorf("unexpected config.yml content: %q", configContent)
	}

	// Check readme.txt
	readmeContent := string(memWriter.Files["readme.txt"])
	if readmeContent != "\nApp name: myapp\n" {
		t.Errorf("unexpected readme.txt content: %q", readmeContent)
	}
}

// TestExecuteWithFiles_WithValidation tests FILE directives with schema validation.
func TestExecuteWithFiles_WithValidation(t *testing.T) {
	data := map[string]interface{}{"name": "test"}
	tmpl := []byte("#FILE:output.txt#\n{{.name}}\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	schema := []byte(`{
		"type":"object",
		"properties":{"name":{"type":"string"}},
		"required":["name"]
	}`)
	validate := WithJsonSchemaValidation(schema)

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter, validate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(memWriter.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(memWriter.Files))
	}
}

// TestExecuteWithFiles_ValidationFailure tests that validation errors are caught.
func TestExecuteWithFiles_ValidationFailure(t *testing.T) {
	data := map[string]interface{}{"name": 123} // Wrong type
	tmpl := []byte("#FILE:output.txt#\n{{.name}}\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	schema := []byte(`{
		"type":"object",
		"properties":{"name":{"type":"string"}},
		"required":["name"]
	}`)
	validate := WithJsonSchemaValidation(schema)

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter, validate)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestExecuteWithFiles_InvalidFilename tests error handling for bad filename templates.
func TestExecuteWithFiles_InvalidFilename(t *testing.T) {
	data := map[string]interface{}{}
	// Use invalid template syntax to trigger parse error
	tmpl := []byte("#FILE:{{.name}#\ncontent\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err == nil {
		t.Fatal("expected error for invalid filename template, got nil")
	}
}

// TestExecuteWithFiles_WriteError tests error handling when file write fails.
func TestExecuteWithFiles_WriteError(t *testing.T) {
	data := map[string]interface{}{"name": "test"}
	tmpl := []byte("#FILE:output.txt#\ncontent\n#FILE#")
	var stdout bytes.Buffer

	// Use DefaultFileWriter with path traversal to trigger write error
	fileWriter := &DefaultFileWriter{}
	// Create a template with path traversal in rendered filename
	tmpl = []byte("#FILE:../forbidden.txt#\ncontent\n#FILE#")

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, fileWriter)
	if err == nil {
		t.Fatal("expected error for write failure, got nil")
	}
}

// TestExecuteWithFiles_DuplicateFilenames tests that duplicate filenames overwrite.
func TestExecuteWithFiles_DuplicateFilenames(t *testing.T) {
	data := map[string]interface{}{}
	tmpl := []byte(`#FILE:same.txt#
first
#FILE#
#FILE:same.txt#
second
#FILE#`)
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Last write should win
	content := string(memWriter.Files["same.txt"])
	if content != "\nsecond\n" {
		t.Errorf("expected last write to win with 'second', got %q", content)
	}
}

// TestExecuteWithFiles_EmptyFileContent tests that empty files are created.
func TestExecuteWithFiles_EmptyFileContent(t *testing.T) {
	data := map[string]interface{}{}
	tmpl := []byte("#FILE:empty.txt#\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, exists := memWriter.Files["empty.txt"]
	if !exists {
		t.Fatal("expected empty.txt to exist")
	}

	// Content should be just the newline
	if string(content) != "\n" {
		t.Errorf("expected content '\\n', got %q", content)
	}
}

// TestExecuteWithFiles_NestedPath tests file creation with directory paths.
func TestExecuteWithFiles_NestedPath(t *testing.T) {
	data := map[string]interface{}{"name": "app"}
	tmpl := []byte("#FILE:logs/{{.name}}/output.log#\nlog entry\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := "logs/app/output.log"
	content, exists := memWriter.Files[expectedPath]
	if !exists {
		t.Fatalf("expected file %q to exist", expectedPath)
	}

	if string(content) != "\nlog entry\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

// TestExecuteWithFiles_CustomFunctions tests that custom template functions work in FILE blocks.
func TestExecuteWithFiles_CustomFunctions(t *testing.T) {
	data := map[string]interface{}{"items": []interface{}{"a", "b", "a", "c"}}
	tmpl := []byte("#FILE:unique.txt#\n{{unique .items}}\n#FILE#")
	var stdout bytes.Buffer
	memWriter := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := ExecuteWithFiles(AnyProvider(data), tmpl, &stdout, memWriter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(memWriter.Files["unique.txt"])
	// unique function should deduplicate the array
	if content != "\n[a b c]\n" {
		t.Errorf("expected unique items, got %q", content)
	}
}
