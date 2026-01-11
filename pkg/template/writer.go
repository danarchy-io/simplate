package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileWriter provides an abstraction for writing files to enable testing
// without actual filesystem I/O.
type FileWriter interface {
	WriteFile(filename string, content []byte) error
}

// DefaultFileWriter is the production implementation of FileWriter that writes
// files to the actual filesystem.
type DefaultFileWriter struct{}

// WriteFile writes content to the specified filename, creating parent directories
// as needed. It performs atomic writes using a temporary file and rename strategy
// to prevent partial writes on error.
//
// Security considerations:
//   - Filenames are sanitized using filepath.Clean()
//   - Path traversal attempts (containing "..") are rejected
//   - Parent directories are created with 0755 permissions
//   - Files are created with 0644 permissions
func (w *DefaultFileWriter) WriteFile(filename string, content []byte) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts before cleaning
	// This catches patterns like "../" or "..\\"
	if strings.Contains(filename, "..") {
		return fmt.Errorf("path traversal not allowed in filename: %s", filename)
	}

	// Sanitize filename
	cleanFilename := filepath.Clean(filename)

	// Get directory path
	dir := filepath.Dir(cleanFilename)

	// Create parent directories if needed
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write to temporary file first for atomic write
	tmpFile := cleanFilename + ".tmp"
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", cleanFilename, err)
	}

	// Rename temporary file to final filename (atomic on most filesystems)
	if err := os.Rename(tmpFile, cleanFilename); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file to %s: %w", cleanFilename, err)
	}

	return nil
}

// MemoryFileWriter is a test implementation of FileWriter that stores files
// in memory rather than writing to the filesystem. This enables fast, isolated
// testing without filesystem side effects.
type MemoryFileWriter struct {
	Files map[string][]byte
}

// WriteFile stores the content in memory under the given filename.
func (w *MemoryFileWriter) WriteFile(filename string, content []byte) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if w.Files == nil {
		w.Files = make(map[string][]byte)
	}

	w.Files[filename] = content
	return nil
}
