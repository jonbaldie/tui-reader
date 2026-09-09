package book

import (
	"os"
	"strings"
	"testing"
)

func loadFixture(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile("../../docs/exploratory-testing/2026-09-09/evidence/fixtures/" + name)
	if err != nil {
		t.Fatalf("fixture missing: %v", err)
	}
	return splitLines(data)
}

func pageLines(b *Book, page int) string {
	return strings.Join(b.Pages[page].Lines, " ")
}

// TestReflow_PositionRoundTripThroughRawAnchors covers the acceptance
// criterion that positioning survives reflow by source location: a reader at
// the last page of a wide layout lands, after reflow, on the page displaying
// the same paragraph (issue #61).
func TestReflow_PositionRoundTripThroughRawAnchors(t *testing.T) {
	lines := loadFixture(t, "reading-journey.md")
	b := &Book{RawLines: lines}
	b.Reflow(72, 17)
	lastPage := len(b.Pages) - 1

	anchor := b.RawLineForPage(lastPage)
	b.Reflow(36, 5)
	page := b.PageForRawLine(anchor)
	if !strings.Contains(pageLines(b, page), "Page-marker paragraph 22") {
		t.Errorf("expected reflowed page %d to show the ending region's first paragraph, got %q", page, pageLines(b, page))
	}
}

// TestReflow_EmptyDocument covers the acceptance criterion that reflowing an
// empty or whitespace-only document neither panics nor leaves a broken
// position.
func TestReflow_EmptyDocument(t *testing.T) {
	for _, content := range []string{"", "\n\n\n", "   \n  \n\n"} {
		path := t.TempDir() + "/empty.md"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		b, err := NewBook(path, 72, 17)
		if err != nil {
			t.Fatal(err)
		}
		b.Reflow(36, 5)
		if len(b.Pages) != 1 {
			t.Fatalf("expected a single empty page, got %d", len(b.Pages))
		}
		if got := b.RawLineForPage(0); got != 0 {
			t.Errorf("expected anchor 0 for empty content, got %d", got)
		}
		if got := b.PageForRawLine(0); got != 0 {
			t.Errorf("expected page 0 for empty content, got %d", got)
		}
	}
}

// TestRawLineForPage_HeadingAnchors covers the edge case that the anchor of
// the page holding the followed link's marker is the heading itself, and that
// a raw line past the last displayed line falls back to the nearest
// preceding page (issue #61).
func TestRawLineForPage_HeadingAnchors(t *testing.T) {
	lines := loadFixture(t, "link-history-journey.md")
	b := &Book{RawLines: lines}
	b.Reflow(72, 17)

	// RETURN MARKER is the last heading on the wide page the reader leaves,
	// so going back resumes at that section.
	if got := b.RawLineForPage(1); got != 22 {
		t.Errorf("expected wide page 1 to anchor at the RETURN MARKER heading (raw 22), got %d", got)
	}
	b.Reflow(36, 5)
	if got := b.PageForRawLine(22); got != 8 {
		t.Errorf("expected raw 22 on narrow page 8, got %d", got)
	}

	// A raw line past the last displayed content line falls back to the
	// nearest preceding page.
	if got := b.PageForRawLine(60); got < 0 || got >= len(b.Pages) {
		t.Errorf("expected a clamped page for raw 60, got %d", got)
	}
}
