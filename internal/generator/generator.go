package generator

import (
	"fmt"
	"io"
	"text/template"
)

func Generate(input any, templateContent []byte, output io.Writer) error {

	tmpl, err := template.New("generator").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(output, input)
}
