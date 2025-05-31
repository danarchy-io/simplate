package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunE_Errors(t *testing.T) {
	origContent := inputContent
	origSchema := inputSchemaFile
	cases := []struct {
		name    string
		init    func()
		args    []string
		wantErr string
	}{
		{
			name:    "no args",
			init:    func() { inputContent, inputSchemaFile = "", "" },
			args:    []string{},
			wantErr: "no template file provided",
		},
		{
			name:    "too many args",
			init:    func() { inputContent, inputSchemaFile = "", "" },
			args:    []string{"a", "b", "c"},
			wantErr: "too many arguments provided",
		},
		{
			name: "no data source",
			init: func() { inputContent, inputSchemaFile = "", "" },
			// one arg, no content flag, no stdin, no file
			args:    []string{"tmpl"},
			wantErr: "no data provided",
		},
		{
			name: "template not found with content flag",
			init: func() { inputContent, inputSchemaFile = "foo: bar", "" },
			args: []string{"nonexistent.tmpl"},
			// the code will try to read the template and fail
			wantErr: "failed to read template file",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// restore globals after
			t.Cleanup(func() {
				inputContent = origContent
				inputSchemaFile = origSchema
			})
			tc.init()
			err := runE(nil, tc.args)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tc.wantErr)) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestRunE_ContentFlag_Success(t *testing.T) {
	origContent := inputContent
	origSchema := inputSchemaFile
	t.Cleanup(func() {
		inputContent = origContent
		inputSchemaFile = origSchema
	})

	// prepare template file
	tmplFile := filepath.Join(t.TempDir(), "tmpl.txt")
	if err := os.WriteFile(tmplFile, []byte("Hello {{.Name}}"), 0644); err != nil {
		t.Fatal(err)
	}
	// set content flag to YAML
	inputContent = "Name: Alice"

	// capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runE(nil, []string{tmplFile})
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}
	got := string(bytes.TrimSpace(out))
	want := "Hello Alice"
	if got != want {
		t.Errorf("output = %q; want %q", got, want)
	}
}

func TestRunE_DataFileArg_Success(t *testing.T) {
	origContent := inputContent
	origSchema := inputSchemaFile
	t.Cleanup(func() {
		inputContent = origContent
		inputSchemaFile = origSchema
	})

	// create template
	tmplFile := filepath.Join(t.TempDir(), "tmpl.txt")
	if err := os.WriteFile(tmplFile, []byte("Age: {{.Age}}"), 0644); err != nil {
		t.Fatal(err)
	}
	// create data file
	dataFile := filepath.Join(t.TempDir(), "data.yml")
	if err := os.WriteFile(dataFile, []byte("Age: 30"), 0644); err != nil {
		t.Fatal(err)
	}

	// capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runE(nil, []string{tmplFile, dataFile})
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}
	got := string(bytes.TrimSpace(out))
	want := "Age: 30"
	if got != want {
		t.Errorf("output = %q; want %q", got, want)
	}
}

func TestRunE_ExplicitStdin_Success(t *testing.T) {
	origContent := inputContent
	origSchema := inputSchemaFile
	origStdin := os.Stdin
	t.Cleanup(func() {
		inputContent = origContent
		inputSchemaFile = origSchema
		os.Stdin = origStdin
	})

	// create template
	tmplFile := filepath.Join(t.TempDir(), "tmpl.txt")
	if err := os.WriteFile(tmplFile, []byte("User: {{.User}}"), 0644); err != nil {
		t.Fatal(err)
	}
	// prepare stdin
	rIn, wIn, _ := os.Pipe()
	wIn.Write([]byte("User: Bob"))
	wIn.Close()
	os.Stdin = rIn

	// capture stdout
	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := runE(nil, []string{tmplFile, "-"})
	wOut.Close()
	out, _ := io.ReadAll(rOut)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runE returned error: %v", err)
	}
	got := string(bytes.TrimSpace(out))
	want := "User: Bob"
	if got != want {
		t.Errorf("output = %q; want %q", got, want)
	}
}
