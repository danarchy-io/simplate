# Simplate CLI

**Simplate** is a simple, YAML-powered template engine written in Go. It uses Go's `text/template` system to process templates with YAML input, making it easy to generate content dynamically.

## Installation

```bash
go install github.com/danarchy-io/simplate@latest
```

## Usage

```bash
simplate [flags] [--] <template-file> [input-file | -]
```

- **template-file**: A template file that follows Go's [`text/template`](https://pkg.go.dev/text/template) syntax.
- **input-file**: A YAML file providing the data used to render the template.
  - If not provided as a positional argument, the input data can be passed via:
    - The `--input-content` flag (as a YAML string)
    - Standard input (`-` as input-file)

## Optional Flags

- `--input-content`: Pass YAML input directly as a string instead of a file.
- `--input-schema-file`: Specify a [JSON Schema](https://json-schema.org/) file to validate the input YAML.

## Description

Simplate CLI is a straightforward template engine. It takes a template file and a YAML data file as input, then uses the data to fill in your template and produce the final output.

## Examples

### Basic usage with template and input files

```bash
simplate template.tmpl data.yaml
```

### Reading input from standard input

```bash
cat data.yaml | simplate template.tmpl
```

```bash
cat data.yaml | simplate template.tmpl -
```

### Using inline input content

```bash
simplate --input-content "$(cat data.yaml)" template.tmpl
```

### Validating input with a JSON Schema

```bash
simplate --input-schema-file schema.json template.tmpl data.yaml
```

### Combining stdin with schema validation

```bash
cat data.yaml | simplate --input-schema-file schema.json template.tmpl
```

```bash
cat data.yaml | simplate --input-schema-file schema.json template.tmpl -
```

## Using Simplate as a Library

You can embed Simplate’s core functionality in your own Go programs by calling the `Execute` function from the `executor` package. This lets you render templates with YAML input (and optional JSON-Schema validation) without invoking the CLI.

```go
import (
    "bytes"
    "fmt"

    "github.com/danarchy-io/simplate/pkg/executor"
)

func main() {
    // raw YAML input
    inputYAML := []byte(`
name: Alice
age: 30
`)

    // Go text/template source
    tmplSrc := []byte("Name: {{.name}}, Age: {{.age}}")

	schema := []byte(`{
		"type":"object",
		"properties":{"name":{"type":"string"},"age":{"type":"integer"}},
		"required":["name", "age"]
	}`)

    // buffer to capture output
    var buf bytes.Buffer

    // render
    err := executor.Execute(inputYAML, tmplSrc, &buf, executor.WithJsonSchemaValidation(schema))
    if err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
    // Output: Name: Alice, Age: 30
}
```

Function signature:

```go
func Execute(
    input     []byte,
    templ     []byte,
    output    io.Writer,
    validateFuncs ...ValidateInputFunc,
) error
```

- input: raw YAML bytes to unmarshal.
- templ: Go text/template source.
- output: destination for the rendered template (any io.Writer).
- validateFuncs: zero or more validators, e.g. executor.WithJsonSchemaValidation(schemaBytes).

## Notes

- Templates should conform to the Go `text/template` format.
  - You can access the values of environment variables using the `env` function, like this: `{{ env "HOME" }}`.
- YAML input should be properly structured and optionally validated using a JSON Schema.
- Use `-` to read input from stdin if the second positional argument is not provided.
