package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryFileWriter_WriteFile(t *testing.T) {
	writer := &MemoryFileWriter{Files: make(map[string][]byte)}

	content := []byte("test content")
	err := writer.WriteFile("test.txt", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(writer.Files))
	}

	storedContent, exists := writer.Files["test.txt"]
	if !exists {
		t.Fatal("expected file 'test.txt' to exist")
	}

	if string(storedContent) != string(content) {
		t.Errorf("expected content %q, got %q", content, storedContent)
	}
}

func TestMemoryFileWriter_EmptyFilename(t *testing.T) {
	writer := &MemoryFileWriter{Files: make(map[string][]byte)}

	err := writer.WriteFile("", []byte("content"))
	if err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}

func TestMemoryFileWriter_MultipleFiles(t *testing.T) {
	writer := &MemoryFileWriter{Files: make(map[string][]byte)}

	files := map[string][]byte{
		"file1.txt": []byte("content 1"),
		"file2.txt": []byte("content 2"),
		"file3.txt": []byte("content 3"),
	}

	for filename, content := range files {
		if err := writer.WriteFile(filename, content); err != nil {
			t.Fatalf("unexpected error writing %s: %v", filename, err)
		}
	}

	if len(writer.Files) != len(files) {
		t.Fatalf("expected %d files, got %d", len(files), len(writer.Files))
	}

	for filename, expectedContent := range files {
		storedContent, exists := writer.Files[filename]
		if !exists {
			t.Errorf("expected file %s to exist", filename)
			continue
		}
		if string(storedContent) != string(expectedContent) {
			t.Errorf("file %s: expected content %q, got %q", filename, expectedContent, storedContent)
		}
	}
}

func TestMemoryFileWriter_Overwrite(t *testing.T) {
	writer := &MemoryFileWriter{Files: make(map[string][]byte)}

	writer.WriteFile("test.txt", []byte("first"))
	writer.WriteFile("test.txt", []byte("second"))

	if string(writer.Files["test.txt"]) != "second" {
		t.Errorf("expected file to be overwritten with 'second', got %q", writer.Files["test.txt"])
	}
}

func TestDefaultFileWriter_WriteFile(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "simplate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writer := &DefaultFileWriter{}
	filename := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err = writer.WriteFile(filename, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	storedContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(storedContent) != string(content) {
		t.Errorf("expected content %q, got %q", content, storedContent)
	}

	// Verify file permissions
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	expectedPerms := os.FileMode(0644)
	if info.Mode().Perm() != expectedPerms {
		t.Errorf("expected permissions %v, got %v", expectedPerms, info.Mode().Perm())
	}
}

func TestDefaultFileWriter_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simplate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writer := &DefaultFileWriter{}
	filename := filepath.Join(tmpDir, "subdir", "nested", "test.txt")
	content := []byte("nested content")

	err = writer.WriteFile(filename, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	storedContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(storedContent) != string(content) {
		t.Errorf("expected content %q, got %q", content, storedContent)
	}

	// Verify directories were created
	dirInfo, err := os.Stat(filepath.Join(tmpDir, "subdir", "nested"))
	if err != nil {
		t.Fatalf("failed to stat created directory: %v", err)
	}

	if !dirInfo.IsDir() {
		t.Error("expected nested path to be a directory")
	}
}

func TestDefaultFileWriter_EmptyFilename(t *testing.T) {
	writer := &DefaultFileWriter{}

	err := writer.WriteFile("", []byte("content"))
	if err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}

func TestDefaultFileWriter_PathTraversal(t *testing.T) {
	writer := &DefaultFileWriter{}

	// Try various path traversal attempts (relative paths with ..)
	pathTraversalAttempts := []string{
		"../../../etc/passwd",
		"subdir/../../outside.txt",
		"./subdir/../../../etc/passwd",
		"test/../../../dangerous.txt",
	}

	for _, filename := range pathTraversalAttempts {
		err := writer.WriteFile(filename, []byte("malicious"))
		if err == nil {
			t.Errorf("expected error for path traversal attempt %q, got nil", filename)
		}
		if err != nil && !contains(err.Error(), "path traversal") {
			t.Errorf("expected 'path traversal' error for %q, got: %v", filename, err)
		}
	}
}

func TestDefaultFileWriter_AtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simplate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writer := &DefaultFileWriter{}
	filename := filepath.Join(tmpDir, "atomic.txt")

	// Write file
	err = writer.WriteFile(filename, []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify temp file was cleaned up
	tmpFile := filename + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("expected temp file to be cleaned up")
	}
}

func TestDefaultFileWriter_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simplate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writer := &DefaultFileWriter{}
	filename := filepath.Join(tmpDir, "overwrite.txt")

	// Write first version
	err = writer.WriteFile(filename, []byte("first"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Overwrite with second version
	err = writer.WriteFile(filename, []byte("second"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify final content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != "second" {
		t.Errorf("expected content 'second', got %q", content)
	}
}

func TestDefaultFileWriter_WithBaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "output")

	writer := &DefaultFileWriter{}
	err := writer.SetBaseDir(baseDir)
	if err != nil {
		t.Fatalf("unexpected error setting base dir: %v", err)
	}

	// Write a file relative to base dir
	filename := "test.txt"
	err = writer.WriteFile(filename, []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created in base directory
	expectedPath := filepath.Join(baseDir, filename)
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file at expected path %s: %v", expectedPath, err)
	}

	if string(content) != "content" {
		t.Errorf("expected content 'content', got %q", content)
	}
}

func TestDefaultFileWriter_WithBaseDir_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "output")

	writer := &DefaultFileWriter{}
	err := writer.SetBaseDir(baseDir)
	if err != nil {
		t.Fatalf("unexpected error setting base dir: %v", err)
	}

	// Write a file with nested path
	filename := "subdir/nested/test.txt"
	err = writer.WriteFile(filename, []byte("nested content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created in correct location
	expectedPath := filepath.Join(baseDir, filename)
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file at expected path %s: %v", expectedPath, err)
	}

	if string(content) != "nested content" {
		t.Errorf("expected content 'nested content', got %q", content)
	}
}

func TestDefaultFileWriter_BaseDir_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "safe")

	writer := &DefaultFileWriter{}
	err := writer.SetBaseDir(baseDir)
	if err != nil {
		t.Fatalf("unexpected error setting base dir: %v", err)
	}

	// Try path traversal with base dir set
	err = writer.WriteFile("../escape.txt", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !contains(err.Error(), "path traversal") {
		t.Errorf("expected 'path traversal' error, got: %v", err)
	}
}

func TestDefaultFileWriter_BaseDir_InvalidPath(t *testing.T) {
	writer := &DefaultFileWriter{}

	// Try to set a file as base dir
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = writer.SetBaseDir(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error when setting file as base dir, got nil")
	}
	if !contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", err)
	}
}

func TestMemoryFileWriter_WithBaseDir(t *testing.T) {
	writer := &MemoryFileWriter{Files: make(map[string][]byte)}
	err := writer.SetBaseDir("output")
	if err != nil {
		t.Fatalf("unexpected error setting base dir: %v", err)
	}

	// Write a file
	err = writer.WriteFile("test.txt", []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file is stored with base dir prefix
	expectedKey := filepath.Join("output", "test.txt")
	content, exists := writer.Files[expectedKey]
	if !exists {
		t.Fatalf("expected file at key %q, but not found. Keys: %v", expectedKey, mapKeys(writer.Files))
	}

	if string(content) != "content" {
		t.Errorf("expected content 'content', got %q", content)
	}
}

func mapKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
