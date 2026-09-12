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

// TestRawLineForPage_MultipleHeadingsAnchorsFirstHeading covers the case where
// a page contains multiple headings: Page 0 anchors to the first content line
// (so initial startup does not skip Page 1), and subsequent pages anchor to the
// first heading on that page (issue #91).
func TestRawLineForPage_MultipleHeadingsAnchorsFirstHeading(t *testing.T) {
	lines := []string{
		"# Title",
		"",
		"## TOC",
		"",
		"# Chapter 1",
		"",
		"Content of chapter 1.",
		"",
		"# Chapter 2",
		"",
		"Content of chapter 2.",
	}
	b := &Book{RawLines: lines}
	b.Reflow(60, 20)

	// Page 0 has Title (0), TOC (2), Chapter 1 (4), Chapter 2 (8).
	// Anchors must point to first content (raw 0), not last heading (raw 8).
	if got := b.RawLineForPage(0); got != 0 {
		t.Errorf("expected page 0 to anchor at raw 0, got %d", got)
	}

	// Reflow to smaller height: Page 0 must remain Page 0.
	b.Reflow(72, 8)
	if got := b.PageForRawLine(0); got != 0 {
		t.Errorf("expected raw 0 on page 0 after reflow, got %d", got)
	}

	// Subsequent page with multiple headings: must anchor to its first heading.
	subsequentDoc := []string{
		"# P0",
		"",
		"Content P0",
		"",
		"# P1 Section A",
		"",
		"# P1 Section B",
	}
	b2 := &Book{RawLines: subsequentDoc}
	b2.Reflow(60, 4)
	// Page 0 has raw 0 (P0) and raw 2 (Content P0) (4 formatted lines).
	// Page 1 has both raw 4 (Section A) and raw 6 (Section B).
	// Page 1 must anchor to Section A (raw 4), not Section B (raw 6).
	p1Anchor := b2.RawLineForPage(1)
	if p1Anchor != 4 {
		t.Errorf("expected page 1 to anchor at first heading (raw 4), got %d", p1Anchor)
	}
}
