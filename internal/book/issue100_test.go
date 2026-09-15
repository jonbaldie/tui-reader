package book

import (
	"strings"
	"testing"
)

func TestIssue100_LinkAttachmentStartsWhereMarkupStarts(t *testing.T) {
	raw := []string{"000[", "[0](#00)"}
	pages := AttachLinks(Paginate(raw, 10, 3), raw, 10, 3)

	if len(pages) != 1 {
		t.Fatalf("pages = %d, want 1: %+v", len(pages), pages)
	}
	if len(pages[0].Links) != 1 {
		t.Fatalf("links = %d, want 1: %+v", len(pages[0].Links), pages[0].Links)
	}

	link := pages[0].Links[0]
	line := pages[0].Lines[link.LineOnPage]
	if !strings.Contains(line, "[0](#00)") {
		t.Fatalf("link attached to line %d = %q; want the display line containing its complete markup, lines=%q", link.LineOnPage, line, pages[0].Lines)
	}
}
