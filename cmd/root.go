package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/danarchy-io/simplate/pkg/executor"
	"github.com/spf13/cobra"
)

var (
	inputContent    string
	inputSchemaFile string

	rootCmd = &cobra.Command{
		Use:   "simplate [flags] [--] <template-file> [input-file | -]",
		Short: "A simple YAML-Powered Template Engine",
		Long: `Simplate CLI is a straightforward template engine. It takes a template
file and a YAML data file as input, then uses the data to fill in your
template and produce the final output.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runE,
	}
)

func init() {

	rootCmd.Flags().StringVarP(&inputContent, "input-content", "c", "", "Input content")
	rootCmd.Flags().StringVarP(&inputSchemaFile, "input-schema-file", "s", "", "Input jsonschema file")
}

func Execute() error {
	return rootCmd.Execute()
}

func runE(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("no template file provided")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments provided")
	}

	templateFile := args[0] // Template file is the first required arg

	// --- Determine Input Source ---
	var dataBytes []byte
	var err error
	var inputSourceType string // For better logging messages

	// 1. Highest priority: --content flag
	if inputContent != "" {
		dataBytes = []byte(inputContent)
		inputSourceType = "content flag"
	} else if len(args) == 2 && args[1] == "-" {
		// 2. Next priority: Explicit '-' argument for stdin
		dataBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read data from stdin (via '-'): %w", err)
		}
		inputSourceType = "explicit stdin ('-')"
	} else {
		// 3. Next priority: Implicit stdin (pipe/redirect)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 { // If stdin is NOT a character device
			dataBytes, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read YAML data from stdin: %w", err)
			}
			inputSourceType = "implicit stdin (pipe/redirect)"
		} else if len(args) == 2 {
			// 4. Lowest priority: Positional argument (yaml-data-file)
			dataFilePath := args[1]
			dataBytes, err = os.ReadFile(dataFilePath)
			if err != nil {
				return fmt.Errorf("failed to read YAML data from file '%s': %w", dataFilePath, err)
			}
			inputSourceType = "file argument"
		} else {
			// No input source found (no --content, no stdin, no file arg)
			return fmt.Errorf("no data provided. Use a data file argument, the '-' argument for stdin, --content flag, or pipe via stdin")
		}
	}

	if len(dataBytes) == 0 {
		return fmt.Errorf("no input provided from %s", inputSourceType)
	}

	templateBytes, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file '%s': %w", templateFile, err)
	}

	if inputSchemaFile != "" {
		inputSchemaBytes, err := os.ReadFile(inputSchemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file '%v': %w", inputSchemaFile, err)
		}
		return executor.Execute(dataBytes, templateBytes, os.Stdout,
			executor.WithJsonSchemaValidation(inputSchemaBytes))
	}

	return executor.Execute(dataBytes, templateBytes, os.Stdout)
}
