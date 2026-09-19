package book

import (
	"strings"
	"testing"
)

func TestIssue113_UTF8BOMDoesNotHideFirstHeading(t *testing.T) {
	path := writeTempFile(t, "min-bom.md", "\ufeff# Start\n\nSee [Top](#start).\n")

	_, lines, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.HasPrefix(lines[0], "\ufeff") {
		t.Fatalf("first line still starts with BOM: %q", lines[0])
	}
	if lines[0] != "# Start" {
		t.Fatalf("first line = %q, want %q", lines[0], "# Start")
	}

	anchors := ExtractAnchors(lines)
	if _, ok := anchors["start"]; !ok {
		t.Fatalf("anchors missing %q: %v", "start", anchors)
	}

	b, err := NewBook(path, 60, 14)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	if page := b.PageForAnchor("start"); page != 0 {
		t.Fatalf("PageForAnchor(%q) = %d, want 0", "start", page)
	}
}

func TestIssue113_FileWithoutBOMUnaffected(t *testing.T) {
	path := writeTempFile(t, "no-bom.md", "# Start\n\nSee [Top](#start).\n")
	_, lines, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if lines[0] != "# Start" {
		t.Fatalf("first line = %q, want %q", lines[0], "# Start")
	}
	anchors := ExtractAnchors(lines)
	if _, ok := anchors["start"]; !ok {
		t.Fatalf("anchors missing %q: %v", "start", anchors)
	}
}
