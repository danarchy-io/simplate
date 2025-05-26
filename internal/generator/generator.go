package generator

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

var funcMap = template.FuncMap{
	"env": os.Getenv,
}

func Generate(input any, templateContent []byte, output io.Writer) error {

	tmpl, err := template.New("generator").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(output, input)
}
