package book

import "testing"

// TestIssue106_LinkSpanningPhysicalLines reproduces #106: a link whose markup
// spans two physical source lines within a soft-wrapped paragraph is dropped
// from page navigation, even though the reflowed text displays it correctly.
func TestIssue106_LinkSpanningPhysicalLines(t *testing.T) {
	raw := []string{
		"Here is a [link",
		"label](#target) across lines.",
	}

	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)

	if len(pages) == 0 {
		t.Fatalf("expected at least one page")
	}
	if len(pages[0].Lines) == 0 || pages[0].Lines[0] != "Here is a [link label](#target) across lines." {
		t.Fatalf("expected reflowed prose line, got %+v", pages[0].Lines)
	}

	if len(pages[0].Links) != 1 {
		t.Fatalf("expected 1 link attached to page 0, got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
	if pages[0].Links[0].Target != "target" || pages[0].Links[0].Label != "link label" {
		t.Errorf("expected link to target %q with label %q, got %+v", "target", "link label", pages[0].Links[0])
	}
}
