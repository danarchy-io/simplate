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

## Notes

- Templates should conform to the Go `text/template` format.
  - You can access the values of environment variables using the `env` function, like this: `{{ env "HOME" }}`.
- YAML input should be properly structured and optionally validated using a JSON Schema.
- Use `-` to read input from stdin if the second positional argument is not provided.
