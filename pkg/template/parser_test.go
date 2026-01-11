package template

import (
	"reflect"
	"testing"
)

func TestParseSegments_EmptyTemplate(t *testing.T) {
	segments, err := ParseSegments([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Type != SegmentStdout {
		t.Errorf("expected SegmentStdout, got %v", segments[0].Type)
	}
	if len(segments[0].Content) != 0 {
		t.Errorf("expected empty content, got %q", segments[0].Content)
	}
}

func TestParseSegments_SingleStdout(t *testing.T) {
	template := []byte("Hello {{.name}}")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Type != SegmentStdout {
		t.Errorf("expected SegmentStdout, got %v", segments[0].Type)
	}
	if !reflect.DeepEqual(segments[0].Content, template) {
		t.Errorf("expected content %q, got %q", template, segments[0].Content)
	}
}

func TestParseSegments_SingleFile(t *testing.T) {
	template := []byte("#FILE:output.txt#\nHello {{.name}}\n#FILE#")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Type != SegmentFile {
		t.Errorf("expected SegmentFile, got %v", segments[0].Type)
	}
	if string(segments[0].Filename) != "output.txt" {
		t.Errorf("expected filename 'output.txt', got %q", segments[0].Filename)
	}
	expectedContent := "\nHello {{.name}}\n"
	if string(segments[0].Content) != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, segments[0].Content)
	}
}

func TestParseSegments_FileWithTemplateFilename(t *testing.T) {
	template := []byte("#FILE:output-{{.id}}.txt#\ncontent\n#FILE#")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if string(segments[0].Filename) != "output-{{.id}}.txt" {
		t.Errorf("expected filename 'output-{{.id}}.txt', got %q", segments[0].Filename)
	}
}

func TestParseSegments_MixedContent(t *testing.T) {
	template := []byte("Stdout content\n#FILE:file.txt#\nFile content\n#FILE#\nMore stdout")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First segment: stdout
	if segments[0].Type != SegmentStdout {
		t.Errorf("segment 0: expected SegmentStdout, got %v", segments[0].Type)
	}
	if string(segments[0].Content) != "Stdout content\n" {
		t.Errorf("segment 0: expected 'Stdout content\\n', got %q", segments[0].Content)
	}

	// Second segment: file
	if segments[1].Type != SegmentFile {
		t.Errorf("segment 1: expected SegmentFile, got %v", segments[1].Type)
	}
	if string(segments[1].Filename) != "file.txt" {
		t.Errorf("segment 1: expected filename 'file.txt', got %q", segments[1].Filename)
	}
	if string(segments[1].Content) != "\nFile content\n" {
		t.Errorf("segment 1: expected '\\nFile content\\n', got %q", segments[1].Content)
	}

	// Third segment: stdout
	if segments[2].Type != SegmentStdout {
		t.Errorf("segment 2: expected SegmentStdout, got %v", segments[2].Type)
	}
	if string(segments[2].Content) != "\nMore stdout" {
		t.Errorf("segment 2: expected '\\nMore stdout', got %q", segments[2].Content)
	}
}

func TestParseSegments_MultipleFiles(t *testing.T) {
	template := []byte("#FILE:first.txt#\nFirst\n#FILE#\n#FILE:second.txt#\nSecond\n#FILE#")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// There will be 3 segments: file, stdout (the \n between files), file
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	if segments[0].Type != SegmentFile || string(segments[0].Filename) != "first.txt" {
		t.Errorf("segment 0: expected file 'first.txt', got type=%v filename=%q", segments[0].Type, segments[0].Filename)
	}
	if segments[1].Type != SegmentStdout {
		t.Errorf("segment 1: expected stdout segment between files, got type=%v", segments[1].Type)
	}
	if segments[2].Type != SegmentFile || string(segments[2].Filename) != "second.txt" {
		t.Errorf("segment 2: expected file 'second.txt', got type=%v filename=%q", segments[2].Type, segments[2].Filename)
	}
}

func TestParseSegments_EmptyFileContent(t *testing.T) {
	template := []byte("#FILE:empty.txt#\n#FILE#")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Type != SegmentFile {
		t.Errorf("expected SegmentFile, got %v", segments[0].Type)
	}
	if string(segments[0].Filename) != "empty.txt" {
		t.Errorf("expected filename 'empty.txt', got %q", segments[0].Filename)
	}
	// Content should be just the newline between markers
	if string(segments[0].Content) != "\n" {
		t.Errorf("expected content '\\n', got %q", segments[0].Content)
	}
}

func TestParseSegments_UnclosedBlock(t *testing.T) {
	template := []byte("#FILE:output.txt#\nContent without closing")
	_, err := ParseSegments(template)
	if err == nil {
		t.Fatal("expected error for unclosed FILE block, got nil")
	}
	if !contains(err.Error(), "unclosed FILE directive") {
		t.Errorf("expected 'unclosed FILE directive' error, got: %v", err)
	}
}

func TestParseSegments_MissingFilename(t *testing.T) {
	template := []byte("#FILE:#\ncontent\n#FILE#")
	_, err := ParseSegments(template)
	if err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
	if !contains(err.Error(), "empty filename") {
		t.Errorf("expected 'empty filename' error, got: %v", err)
	}
}

func TestParseSegments_MalformedOpeningMarker(t *testing.T) {
	template := []byte("#FILE:filename\ncontent\n#FILE#")
	_, err := ParseSegments(template)
	if err == nil {
		t.Fatal("expected error for malformed FILE directive, got nil")
	}
	// This will be detected as unclosed since the opening marker is never properly closed with #
	if !contains(err.Error(), "unclosed FILE directive") && !contains(err.Error(), "malformed FILE directive") {
		t.Errorf("expected 'unclosed FILE directive' or 'malformed FILE directive' error, got: %v", err)
	}
}

func TestParseSegments_NestedBlocks(t *testing.T) {
	template := []byte("#FILE:outer.txt#\n#FILE:inner.txt#\nNested\n#FILE#\n#FILE#")
	_, err := ParseSegments(template)
	if err == nil {
		t.Fatal("expected error for nested FILE blocks, got nil")
	}
	if !contains(err.Error(), "nested FILE directive") {
		t.Errorf("expected 'nested FILE directive' error, got: %v", err)
	}
}

func TestParseSegments_UnexpectedClosingMarker(t *testing.T) {
	template := []byte("Some content\n#FILE#\n")
	_, err := ParseSegments(template)
	if err == nil {
		t.Fatal("expected error for unexpected closing marker, got nil")
	}
	if !contains(err.Error(), "unexpected FILE closing marker") {
		t.Errorf("expected 'unexpected FILE closing marker' error, got: %v", err)
	}
}

func TestParseSegments_WhitespaceOnly(t *testing.T) {
	template := []byte("   \n\t\n   ")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be filtered to empty stdout segment
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Type != SegmentStdout {
		t.Errorf("expected SegmentStdout, got %v", segments[0].Type)
	}
}

func TestParseSegments_FileWithPathSeparators(t *testing.T) {
	template := []byte("#FILE:output/nested/file.txt#\ncontent\n#FILE#")
	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if string(segments[0].Filename) != "output/nested/file.txt" {
		t.Errorf("expected filename 'output/nested/file.txt', got %q", segments[0].Filename)
	}
}

func TestParseSegments_ComplexMixedContent(t *testing.T) {
	template := []byte(`Start of template
#FILE:config-{{.env}}.yml#
config: {{.config}}
#FILE#
Middle content
#FILE:logs/{{.name}}.log#
log entry
#FILE#
End of template`)

	segments, err := ParseSegments(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 5 {
		t.Fatalf("expected 5 segments, got %d", len(segments))
	}

	// Validate segment types and filenames
	expected := []struct {
		typ      SegmentType
		filename string
	}{
		{SegmentStdout, ""},
		{SegmentFile, "config-{{.env}}.yml"},
		{SegmentStdout, ""},
		{SegmentFile, "logs/{{.name}}.log"},
		{SegmentStdout, ""},
	}

	for i, exp := range expected {
		if segments[i].Type != exp.typ {
			t.Errorf("segment %d: expected type %v, got %v", i, exp.typ, segments[i].Type)
		}
		if exp.typ == SegmentFile && string(segments[i].Filename) != exp.filename {
			t.Errorf("segment %d: expected filename %q, got %q", i, exp.filename, segments[i].Filename)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAny(s, substr))
}

func containsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
