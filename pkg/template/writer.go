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
	SetBaseDir(dir string) error
}

// DefaultFileWriter is the production implementation of FileWriter that writes
// files to the actual filesystem.
type DefaultFileWriter struct {
	baseDir string
}

// SetBaseDir sets the base directory for file writes. All file paths will be
// relative to this directory. If dir is empty, files are written relative to
// the current working directory.
func (w *DefaultFileWriter) SetBaseDir(dir string) error {
	if dir == "" {
		w.baseDir = ""
		return nil
	}

	// Clean the directory path
	cleanDir := filepath.Clean(dir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", cleanDir, err)
	}

	// Verify it's a directory
	info, err := os.Stat(cleanDir)
	if err != nil {
		return fmt.Errorf("failed to stat output directory %s: %w", cleanDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output path %s is not a directory", cleanDir)
	}

	w.baseDir = cleanDir
	return nil
}

// WriteFile writes content to the specified filename, creating parent directories
// as needed. It performs atomic writes using a temporary file and rename strategy
// to prevent partial writes on error.
//
// If a base directory is set via SetBaseDir, the filename is treated as relative
// to that directory.
//
// Security considerations:
//   - Filenames are sanitized using filepath.Clean()
//   - Path traversal attempts (containing "..") are rejected
//   - Parent directories are created with 0755 permissions
//   - Files are created with 0644 permissions
//   - Final path is verified to be within base directory (if set)
func (w *DefaultFileWriter) WriteFile(filename string, content []byte) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts before joining with base dir
	// This catches patterns like "../" or "..\\"
	if strings.Contains(filename, "..") {
		return fmt.Errorf("path traversal not allowed in filename: %s", filename)
	}

	// Join with base directory if set
	fullPath := filename
	if w.baseDir != "" {
		fullPath = filepath.Join(w.baseDir, filename)
	}

	// Sanitize the full path
	cleanFilename := filepath.Clean(fullPath)

	// Verify the resolved path is still within baseDir (defense in depth)
	if w.baseDir != "" {
		relPath, err := filepath.Rel(w.baseDir, cleanFilename)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("resolved path %s is outside output directory", cleanFilename)
		}
	}

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
	Files   map[string][]byte
	baseDir string
}

// SetBaseDir sets the base directory for file writes in memory.
// The directory path is stored but not validated (since this is in-memory only).
func (w *MemoryFileWriter) SetBaseDir(dir string) error {
	w.baseDir = filepath.Clean(dir)
	return nil
}

// WriteFile stores the content in memory under the given filename.
// If a base directory is set, the filename is joined with it.
func (w *MemoryFileWriter) WriteFile(filename string, content []byte) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if w.Files == nil {
		w.Files = make(map[string][]byte)
	}

	// Join with base directory if set
	fullPath := filename
	if w.baseDir != "" {
		fullPath = filepath.Join(w.baseDir, filename)
	}

	w.Files[fullPath] = content
	return nil
}
