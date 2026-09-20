package book

import (
	"strings"
	"testing"
)

// TestIssue113_BOMAtFileStartDoesNotHideFirstHeading reproduces issue #113: a
// UTF-8 byte-order mark left on the first line stops "# Start" being read as a
// heading, so it has no anchor and links to it are dead.
func TestIssue113_BOMAtFileStartDoesNotHideFirstHeading(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("\ufeff# Start\n\n")
	for i := 0; i < 8; i++ {
		sb.WriteString("A short paragraph of filler text.\n\n")
	}
	sb.WriteString("[Top](#start)\n")
	path := writeTempFile(t, "min-bom.md", sb.String())

	b, err := NewBook(path, 60, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.RawLines[0]; got != "# Start" {
		t.Errorf("first raw line = %q, want %q (BOM stripped)", got, "# Start")
	}
	if _, ok := b.Anchors["start"]; !ok {
		t.Errorf("anchors = %v, want an entry for %q", b.Anchors, "start")
	}
	if page := b.PageForAnchor("start"); page != 0 {
		t.Errorf("PageForAnchor(%q) = %d, want 0", "start", page)
	}
}
