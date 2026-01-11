package template

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type InputProvider func() (any, error)
type ValidateInputFunc func(input any) error

// AnyProvider returns an InputProvider that simply wraps the given Go value.
// When the returned provider is invoked, it returns the original input value.
// If the input is nil, the provider returns an error instead.
// This is useful when you already have a Go value as input and don’t need to
// unmarshal from YAML or JSON.
//
// Example:
//
//	provider := AnyProvider(map[string]interface{}{"foo": "bar"})
//	data, err := provider()
//	// data == map[string]interface{}{"foo":"bar"}, err == nil
func AnyProvider(input any) InputProvider {
	return func() (any, error) {
		if input == nil {
			return nil, fmt.Errorf("input is nil")
		}
		return input, nil
	}
}

// YamlProvider returns an InputProvider that unmarshals the provided YAML bytes
// into a Go data structure (map[string]interface{} for objects or []interface{} for arrays).
// When the returned provider is invoked, it parses the YAML input and returns
// the resulting data or an error if unmarshalling fails.
//
// Example:
//
//	provider := YamlProvider([]byte("foo: bar\nbaz:\n  - qux\n"))
//	data, err := provider()
//	// data == map[string]interface{}{"foo":"bar","baz":[]interface{}{"qux"}}, err == nil
func YamlProvider(input []byte) InputProvider {
	return func() (any, error) {
		var data any
		if err := yaml.Unmarshal(input, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML input: %w", err)
		}
		return data, nil
	}
}

// WithJsonSchemaValidation returns a ValidateInputFunc that validates
// a parsed YAML input (the result of yaml.Unmarshal) against the
// provided JSON Schema.
// The schema parameter must be the JSON Schema definition as raw bytes.
// The returned function compiles this schema and applies it to the input,
// returning an error if schema compilation or validation fails.
func WithJsonSchemaValidation(schema []byte) ValidateInputFunc {
	return func(input any) error {
		schema, err := jsonschema.CompileString("schema.json", string(schema))
		if err != nil {
			return fmt.Errorf("failed to compile JSONSchema: %w", err)
		}

		return schema.Validate(input)
	}
}

// Execute parses the given YAML input, optionally validates it,
// then applies a Go html/template and writes the result to output.
//
// Parameters:
//   - input: raw YAML bytes to unmarshal (resulting in map[string]interface{}
//     or []interface{}).
//   - template: Go text/template source as bytes.
//   - output: destination io.Writer for the rendered template.
//   - validateInputFuncs: zero or more validation functions (ValidateInputFunc)
//     which are invoked on the unmarshaled data before rendering.
//
// It returns an error if any of the following steps fail:
//  1. YAML unmarshalling of input
//  2. any validation function
//  3. parsing the template
//  4. executing the template
func Execute(inputProvider InputProvider, templ []byte, output io.Writer, validateInputFuncs ...ValidateInputFunc) error {

	data, err := inputProvider()
	if err != nil {
		return fmt.Errorf("failed to get input data: %w", err)
	}

	for _, validateFunc := range validateInputFuncs {
		if err := validateFunc(data); err != nil {
			return fmt.Errorf("input validation failed: %w", err)
		}
	}

	tmpl, err := template.New("generator").Funcs(template.FuncMap{
		"env":          os.Getenv,
		"envOrDefault": envOrDefault,
		"unique":       unique,
	}).Parse(string(templ))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(output, data)
}

// ExecuteWithFiles parses the given template for FILE directives, validates input,
// and renders segments to either stdout or files based on the directives.
//
// This function extends Execute() with multi-file generation support using FILE
// directive markers in the template:
//
//	#FILE:filename.txt#
//	content for this file
//	#FILE#
//
// Parameters:
//   - inputProvider: provider function that returns the input data
//   - templ: Go text/template source as bytes, may contain FILE directives
//   - output: destination io.Writer for stdout segments (content outside FILE blocks)
//   - fileWriter: FileWriter implementation for writing file segments
//   - validateInputFuncs: zero or more validation functions (ValidateInputFunc)
//     which are invoked on the unmarshaled data before rendering.
//
// Features:
//   - Filenames can contain template expressions (e.g., #FILE:output-{{.id}}.txt#)
//   - Each FILE block is rendered as an independent template with full data access
//   - Content outside FILE blocks is rendered to the output writer
//   - Parent directories are automatically created for file paths
//   - Templates without FILE directives work identically to Execute()
//
// It returns an error if any of the following steps fail:
//  1. Getting input data from the provider
//  2. Any validation function
//  3. Parsing FILE directives (malformed syntax, unclosed blocks, etc.)
//  4. Parsing or executing templates (for filenames or content)
//  5. Writing files
func ExecuteWithFiles(
	inputProvider InputProvider,
	templ []byte,
	output io.Writer,
	fileWriter FileWriter,
	validateInputFuncs ...ValidateInputFunc,
) error {
	// Get input data
	data, err := inputProvider()
	if err != nil {
		return fmt.Errorf("failed to get input data: %w", err)
	}

	// Run validation functions
	for _, validateFunc := range validateInputFuncs {
		if err := validateFunc(data); err != nil {
			return fmt.Errorf("input validation failed: %w", err)
		}
	}

	// Parse template into segments
	segments, err := ParseSegments(templ)
	if err != nil {
		return fmt.Errorf("failed to parse template segments: %w", err)
	}

	// Process each segment
	for i, segment := range segments {
		switch segment.Type {
		case SegmentStdout:
			// Render stdout segment
			if err := renderSegment(segment.Content, data, output); err != nil {
				return fmt.Errorf("failed to render stdout segment %d: %w", i, err)
			}

		case SegmentFile:
			// Render filename template
			var filenameBuf bytes.Buffer
			if err := renderSegment(segment.Filename, data, &filenameBuf); err != nil {
				return fmt.Errorf("failed to render filename template for segment %d: %w", i, err)
			}
			filename := filenameBuf.String()

			// Render file content template
			var contentBuf bytes.Buffer
			if err := renderSegment(segment.Content, data, &contentBuf); err != nil {
				return fmt.Errorf("failed to render file content for %s: %w", filename, err)
			}

			// Write file
			if err := fileWriter.WriteFile(filename, contentBuf.Bytes()); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filename, err)
			}
		}
	}

	return nil
}

// renderSegment parses and executes a template segment with the given data,
// writing the result to the provided writer.
func renderSegment(templateContent []byte, data any, output io.Writer) error {
	tmpl, err := template.New("segment").Funcs(template.FuncMap{
		"env":          os.Getenv,
		"envOrDefault": envOrDefault,
		"unique":       unique,
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(output, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

