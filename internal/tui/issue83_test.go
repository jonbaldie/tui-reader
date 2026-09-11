package tui

import (
	"strings"
	"testing"
)

func TestIssue83_OverwideLinkCrossingPageBoundaryNavigation(t *testing.T) {
	doc := `F0

F1

F2

F3

F4

This paragraph begins on page zero and continues across the page boundary onto page one where an overwide link appears [Extremely long link that exceeds twenty-five columns easily](#target)

# Target

Target content.
`
	path := writeTempFile(t, "issue83.md", doc)
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	// A 29x20 terminal gives the reported content dimensions: 25x13.
	m = applyWindowSize(m, 29, 20)
	m = pressKey(m, "home")
	if len(m.book.Pages) < 2 {
		t.Fatalf("expected at least two pages, got %d", len(m.book.Pages))
	}
	if len(m.book.Pages[0].Links) != 0 {
		t.Fatalf("page 0 has %d phantom links, want 0: %+v", len(m.book.Pages[0].Links), m.book.Pages[0].Links)
	}
	if len(m.book.Pages[1].Links) != 1 {
		t.Fatalf("page 1 has %d links, want 1: %+v", len(m.book.Pages[1].Links), m.book.Pages[1].Links)
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != -1 {
		t.Fatalf("Tab on page 0 selected link %d, want no selection", m.SelectedLink())
	}

	m = pressKey(m, "right")
	if m.CurrentPage() != 1 {
		t.Fatalf("expected to advance to page 1, got page %d", m.CurrentPage())
	}
	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("Tab on page 1 selected link %d, want 0", m.SelectedLink())
	}

	m = pressKey(m, "enter")
	targetPage := m.book.PageForAnchor("target")
	if m.CurrentPage() != targetPage {
		t.Fatalf("Enter landed on page %d, want target page %d", m.CurrentPage(), targetPage)
	}
	if !strings.Contains(strings.Join(m.book.Pages[m.CurrentPage()].Lines, "\n"), "# Target") {
		t.Fatalf("target page does not contain heading: %v", m.book.Pages[m.CurrentPage()].Lines)
	}
}
