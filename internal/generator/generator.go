package generator

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

// Generate parses and executes a text template using the provided input data.
// It registers three helper functions for use within templates:
//   - "env": wraps os.Getenv, returning the value of the named environment variable.
//   - "envOrDefault": returns a default value if the environment variable is unset or empty.
//   - "unique": removes duplicate elements from a slice, preserving first‐occurrence order.
//
// Parameters:
//   - input:    the data model (any Go value) supplied to the template.
//   - templateContent: raw template text as a byte slice.
//   - output:   destination io.Writer for the rendered template output.
//
// Returns:
//   - error: non‐nil if template parsing or execution fails.
func Generate(input any, templateContent []byte, output io.Writer) error {

	tmpl, err := template.New("generator").Funcs(template.FuncMap{
		"env":          os.Getenv,
		"envOrDefault": getEnvOrDefault,
		"unique":       unique,
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(output, input)
}
