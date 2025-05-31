package executor

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

type ValidateInputFunc func(input any) error

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
func Execute(input []byte, templ []byte, output io.Writer, validateInputFuncs ...ValidateInputFunc) error {

	var data any
	err := yaml.Unmarshal(input, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML input: %w", err)
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
