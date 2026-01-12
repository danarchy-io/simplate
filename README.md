# Simplate CLI

[![CI](https://github.com/danarchy-io/simplate/workflows/CI/badge.svg)](https://github.com/danarchy-io/simplate/actions?query=workflow%3ACI)
[![Release](https://github.com/danarchy-io/simplate/workflows/Release/badge.svg)](https://github.com/danarchy-io/simplate/actions?query=workflow%3ARelease)

**Simplate** is a simple, YAML-powered template engine written in Go. It uses Go's `text/template` system to process templates with YAML input, making it easy to generate content dynamically.

## Installation

### Download Pre-built Binary

Download the latest release for your architecture from the [releases page](https://github.com/danarchy-io/simplate/releases):

```bash
# For Linux amd64
VERSION=v1.0.0  # Replace with the latest version
ARCH=amd64      # Or arm64

# Download binary and checksums
wget https://github.com/danarchy-io/simplate/releases/download/${VERSION}/simplate-${VERSION}-linux-${ARCH}.tar.gz
wget https://github.com/danarchy-io/simplate/releases/download/${VERSION}/SHA256SUMS

# Verify checksum
sha256sum -c SHA256SUMS --ignore-missing

# Extract and use
tar -xzf simplate-${VERSION}-linux-${ARCH}.tar.gz
cd simplate-${VERSION}-linux-${ARCH}
./simplate --version

# Optionally, move to PATH
sudo mv simplate /usr/local/bin/
```

### Go Install (from source)

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

- `--input-content` or `-c`: Pass YAML input directly as a string instead of a file.
- `--input-schema-file` or `-s`: Specify a [JSON Schema](https://json-schema.org/) file to validate the input YAML.
- `--output-dir` or `-o`: Specify output directory for FILE directives (default: current directory).

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

You can embed Simplate’s core functionality in your own Go programs by calling the `Execute` function from the `template` package. This lets you render templates with YAML input (and optional JSON-Schema validation) without invoking the CLI.

```go
import (
    "bytes"
    "fmt"

    "github.com/danarchy-io/simplate/pkg/template"
)

func main() {
    // raw YAML input
    inputYAML := []byte(`
name: Alice
age: 30
`)

    // Go text/template source
    tmplSrc := []byte("Name: {{.name}}, Age: {{.age}}")

    // optional JSON Schema for validation
    schema := []byte(`{
        "type":"object",
        "properties":{
            "name":{"type":"string"},
            "age":{"type":"integer"}
        },
        "required":["name","age"]
    }`)

    // buffer to capture output
    var buf bytes.Buffer

    // render: wrap the raw bytes in a provider, then call Execute
    err := template.Execute(
        template.YamlProvider(inputYAML),
        tmplSrc,
        &buf,
        template.WithJsonSchemaValidation(schema),
    )
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
    inputProvider InputProvider,
    templ         []byte,
    output        io.Writer,
    validateFuncs ...ValidateInputFunc,
) error
```

- inputProvider:
    - YamlProvider(rawYAML []byte) to unmarshal YAML
    - AnyProvider(value interface{}) for already–parsed Go values
- templ: Go text/template source as bytes
- output: any io.Writer
- validateFuncs: zero or more ValidateInputFunc (e.g. WithJsonSchemaValidation(schemaBytes))

## Multi-File Generation with FILE Directives

Simplate supports generating multiple files from a single template using FILE directives. This allows you to create complex output structures with one command.

### FILE Directive Syntax

Use the `#FILE:filename#` and `#FILE#` markers to define file sections:

```
#FILE:filename.txt#
content for this file
#FILE#
```

### Features

- **Template-rendered filenames**: Filenames can contain template expressions
  ```
  #FILE:config-{{.environment}}.yml#
  ```
- **Multiple files**: Define as many FILE blocks as needed
- **Nested directories**: Parent directories are created automatically
  ```
  #FILE:logs/{{.name}}/output.log#
  ```
- **Mixed output**: Content outside FILE blocks goes to stdout
- **Full template support**: Each FILE block has access to all template data and functions

### Example

Template (`config.tmpl`):
```
Generated at {{.timestamp}}

#FILE:config-{{.environment}}.yml#
server:
  name: {{.name}}
  port: {{.port}}
#FILE#

#FILE:logs/{{.name}}.log#
Application: {{.name}}
Started: {{.timestamp}}
#FILE#

Summary: Configuration created for {{.name}}
```

Data (`data.yaml`):
```yaml
name: api-service
environment: production
port: 8080
timestamp: 2025-01-11T10:00:00Z
```

Running:
```bash
simplate config.tmpl data.yaml
```

**Output to stdout:**
```
Generated at 2025-01-11T10:00:00Z



Summary: Configuration created for api-service
```

**Files created:**
- `config-production.yml` - Server configuration
- `logs/api-service.log` - Log file (directory created automatically)

### Using Output Directory

Specify where generated files should be written:

```bash
# Write all files to ./build directory
simplate --output-dir build config.tmpl data.yaml

# Short form
simplate -o ./dist config.tmpl data.yaml

# Nested paths in FILE directives are relative to output dir
# #FILE:config/app.yml# will write to build/config/app.yml
```

The output directory will be created automatically if it doesn't exist. All FILE directive paths are treated as relative to this directory.

## Library Usage with Multi-File Generation

Use `ExecuteWithFiles` for FILE directive support:

```go
import (
    "bytes"
    "github.com/danarchy-io/simplate/pkg/template"
)

func main() {
    inputYAML := []byte(`name: myapp`)

    tmplSrc := []byte(`#FILE:{{.name}}.conf#
app_name={{.name}}
#FILE#`)

    var stdout bytes.Buffer
    fileWriter := &template.DefaultFileWriter{}

    err := template.ExecuteWithFiles(
        template.YamlProvider(inputYAML),
        tmplSrc,
        &stdout,
        fileWriter,
    )
    // Creates file: myapp.conf
}
```

For testing, use `MemoryFileWriter`:

```go
memWriter := &template.MemoryFileWriter{Files: make(map[string][]byte)}

err := template.ExecuteWithFiles(
    template.YamlProvider(inputYAML),
    tmplSrc,
    &stdout,
    memWriter,
)

// Access files in memory
content := memWriter.Files["myapp.conf"]
```

Function signature:

```go
func ExecuteWithFiles(
    inputProvider InputProvider,
    templ         []byte,
    output        io.Writer,
    fileWriter    FileWriter,
    validateFuncs ...ValidateInputFunc,
) error
```

## Development

### Running Tests

All pull requests must pass unit tests before merging. The CI workflow automatically runs tests on:
- Go 1.24.3 (project version)
- Go 1.24.x (latest patch)

To run tests locally:

```bash
# Run all tests
go test ./... -v

# Run tests with race detection and coverage
go test ./... -race -coverprofile=coverage.out -covermode=atomic

# View coverage report
go tool cover -html=coverage.out

# Display coverage summary
go tool cover -func=coverage.out
```

### Building

```bash
# Build the binary
go build -v .

# Build with version information
go build -ldflags="-X main.version=v1.0.0" -o simplate .

# Run the binary
./simplate version
```

### Continuous Integration

The project uses GitHub Actions for automated testing and releases:

- **CI Workflow**: Runs on all pull requests to main
  - Executes unit tests with race detection
  - Builds binary to verify compilation
  - Generates coverage reports
  - Tests on multiple Go versions

- **Release Workflow**: Triggers on GitHub releases
  - Builds static binaries for Linux (amd64, arm64)
  - Generates SHA256 checksums
  - Uploads release artifacts

## Notes

- Templates should conform to the Go `text/template` format.
  - You can access the values of environment variables using the `env` function, like this: `{{ env "HOME" }}`.
  - You can access an environment variable with a fallback using `envOrDefault`, e.g. `{{ envOrDefault "LOG_LEVEL" "info" }}`.
  - You can remove duplicate elements from a slice (preserving order) using `unique`, e.g. `{{ unique .items }}`.
- YAML input should be properly structured and optionally validated using a JSON Schema.
- Use `-` to read input from stdin if the second positional argument is not provided.
- FILE directives cannot be nested.
- Filenames are sanitized to prevent path traversal attacks (e.g., `../` is rejected).
