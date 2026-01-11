package template

import (
	"bytes"
	"fmt"
	"strings"
)

// SegmentType represents the type of template segment.
type SegmentType int

const (
	// SegmentStdout indicates content that should be written to stdout.
	SegmentStdout SegmentType = iota
	// SegmentFile indicates content that should be written to a file.
	SegmentFile
)

// Segment represents a portion of a template that is either directed to stdout
// or to a specific file.
type Segment struct {
	Type     SegmentType
	Content  []byte // Raw template content to be rendered
	Filename []byte // Template expression for filename (FILE segments only)
}

const (
	fileOpenPrefix = "#FILE:"
	fileOpenSuffix = "#"
	fileClose      = "#FILE#"
)

// ParseSegments parses a template into segments based on FILE directive markers.
// It identifies #FILE:filename# ... #FILE# blocks and separates them from
// content that should go to stdout.
//
// Returns a slice of Segment objects representing the parsed template, or an
// error if the template contains malformed FILE directives.
//
// Error conditions:
//   - Unclosed FILE directive (missing closing #FILE#)
//   - Nested FILE directives (FILE directive inside another FILE directive)
//   - Empty filename in FILE directive
func ParseSegments(templateBytes []byte) ([]Segment, error) {
	if len(templateBytes) == 0 {
		return []Segment{{Type: SegmentStdout, Content: []byte{}}}, nil
	}

	var segments []Segment
	template := string(templateBytes)
	pos := 0
	inFileBlock := false
	fileBlockStart := 0

	for pos < len(template) {
		// Look for FILE directive markers
		openIdx := strings.Index(template[pos:], fileOpenPrefix)
		closeIdx := strings.Index(template[pos:], fileClose)

		// No more directives found
		if openIdx == -1 && closeIdx == -1 {
			if inFileBlock {
				return nil, fmt.Errorf("unclosed FILE directive starting at position %d", fileBlockStart)
			}
			// Add remaining content as stdout segment
			if pos < len(template) {
				segments = append(segments, Segment{
					Type:    SegmentStdout,
					Content: []byte(template[pos:]),
				})
			}
			break
		}

		if !inFileBlock {
			// We're looking for an opening directive
			if closeIdx != -1 && (openIdx == -1 || closeIdx < openIdx) {
				return nil, fmt.Errorf("unexpected FILE closing marker at position %d", pos+closeIdx)
			}

			if openIdx != -1 {
				// Found opening directive
				// Add any content before this as stdout segment
				if openIdx > 0 {
					segments = append(segments, Segment{
						Type:    SegmentStdout,
						Content: []byte(template[pos : pos+openIdx]),
					})
				}

				// Find the end of the opening marker (the second #)
				openStart := pos + openIdx
				filenameStart := openStart + len(fileOpenPrefix)
				filenameEnd := strings.Index(template[filenameStart:], fileOpenSuffix)

				if filenameEnd == -1 {
					return nil, fmt.Errorf("malformed FILE directive at position %d: missing closing # in filename", openStart)
				}

				filename := template[filenameStart : filenameStart+filenameEnd]
				if strings.TrimSpace(filename) == "" {
					return nil, fmt.Errorf("empty filename in FILE directive at position %d", openStart)
				}

				// Check for nested FILE directive in filename
				if strings.Contains(filename, fileOpenPrefix) {
					return nil, fmt.Errorf("nested FILE directive not allowed at position %d", openStart)
				}

				fileBlockStart = openStart
				inFileBlock = true
				pos = filenameStart + filenameEnd + len(fileOpenSuffix)

				// Store filename for later when we find the closing marker
				segments = append(segments, Segment{
					Type:     SegmentFile,
					Filename: []byte(filename),
					Content:  nil, // Will be filled when we find the closing marker
				})
			}
		} else {
			// We're inside a FILE block, looking for closing directive
			if openIdx != -1 && (closeIdx == -1 || openIdx < closeIdx) {
				return nil, fmt.Errorf("nested FILE directive not allowed at position %d", pos+openIdx)
			}

			if closeIdx != -1 {
				// Found closing directive
				// Extract content between opening and closing
				content := template[pos : pos+closeIdx]
				segments[len(segments)-1].Content = []byte(content)

				inFileBlock = false
				pos = pos + closeIdx + len(fileClose)
			} else {
				return nil, fmt.Errorf("unclosed FILE directive starting at position %d", fileBlockStart)
			}
		}
	}

	if inFileBlock {
		return nil, fmt.Errorf("unclosed FILE directive starting at position %d", fileBlockStart)
	}

	// If no segments were created, return a single stdout segment with all content
	if len(segments) == 0 {
		return []Segment{{Type: SegmentStdout, Content: templateBytes}}, nil
	}

	// Filter out empty stdout segments at the beginning and end
	return filterEmptyEdgeSegments(segments), nil
}

// filterEmptyEdgeSegments removes empty stdout segments from the beginning
// and end of the segments slice, but preserves empty segments in the middle
// and all FILE segments (even if empty).
func filterEmptyEdgeSegments(segments []Segment) []Segment {
	if len(segments) == 0 {
		return segments
	}

	start := 0
	end := len(segments)

	// Find first non-empty segment or first FILE segment
	for start < len(segments) {
		seg := segments[start]
		if seg.Type == SegmentFile || len(bytes.TrimSpace(seg.Content)) > 0 {
			break
		}
		start++
	}

	// Find last non-empty segment or last FILE segment
	for end > start {
		seg := segments[end-1]
		if seg.Type == SegmentFile || len(bytes.TrimSpace(seg.Content)) > 0 {
			break
		}
		end--
	}

	if start >= end {
		// All segments were empty stdout segments
		return []Segment{{Type: SegmentStdout, Content: []byte{}}}
	}

	return segments[start:end]
}
