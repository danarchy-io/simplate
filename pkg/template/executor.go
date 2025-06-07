package template

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

var funcMap = template.FuncMap{
	"env": os.Getenv,
}

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

	tmpl, err := template.New("generator").Funcs(funcMap).Parse(string(templ))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(output, data)
}
