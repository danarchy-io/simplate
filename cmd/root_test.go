package cmd

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd(args ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "simplate",
	}
	cmd.SetArgs(args)
	return cmd
}

func Test_runE_NoArgs(t *testing.T) {
	cmd := newTestCmd()
	err := runE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "no template file provided") {
		t.Errorf("expected error for no args, got: %v", err)
	}
}

func Test_runE_InputContentFlag(t *testing.T) {
	inputContent = "foo: bar"
	defer func() { inputContent = "" }()

	tmp, err := os.CreateTemp("", "tmpl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString("{{.foo}}")
	tmp.Close()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newTestCmd(tmp.Name())
	err = runE(cmd, []string{tmp.Name()})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(out)) != "bar" {
		t.Errorf("expected output 'bar', got '%s'", string(out))
	}
}

func Test_runE_TemplateFileNotFound(t *testing.T) {
	inputContent = "foo: bar"
	defer func() { inputContent = "" }()
	cmd := newTestCmd("notfound.tmpl")
	err := runE(cmd, []string{"notfound.tmpl"})
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected file not found error, got: %v", err)
	}
}

func Test_runE_EmptyInputContent(t *testing.T) {
	inputContent = ""
	defer func() { inputContent = "" }()
	cmd := newTestCmd("file.tmpl")
	err := runE(cmd, []string{"file.tmpl"})
	if err == nil || !strings.Contains(err.Error(), "no data provided") {
		t.Errorf("expected error for empty input content, got: %v", err)
	}
}

func Test_runE_FileArgInput(t *testing.T) {
	// Create YAML data file
	dataFile, err := os.CreateTemp("", "data.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dataFile.Name())
	dataFile.WriteString("foo: filebar")
	dataFile.Close()

	// Create template file
	tmplFile, err := os.CreateTemp("", "tmpl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmplFile.Name())
	tmplFile.WriteString("{{.foo}}")
	tmplFile.Close()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newTestCmd(tmplFile.Name(), dataFile.Name())
	err = runE(cmd, []string{tmplFile.Name(), dataFile.Name()})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(out)) != "filebar" {
		t.Errorf("expected output 'filebar', got '%s'", string(out))
	}
}

func Test_CLI_Execute_WithInputContentFlag(t *testing.T) {
	// Create template file
	tmplFile, err := os.CreateTemp("", "tmpl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmplFile.Name())
	tmplFile.WriteString("{{.foo}}")
	tmplFile.Close()

	// Save and restore os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"simplate", "-c", "foo: cli", tmplFile.Name()}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = Execute()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(out)) != "cli" {
		t.Errorf("expected output 'cli', got '%s'", string(out))
	}
}
