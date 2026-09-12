package tui

import (
	"strings"
	"testing"
)

func TestIssue89_RepeatedInternalLinksStayNavigableOnLaterPages(t *testing.T) {
	doc := `# Chapter 1

First chapter content.

[Back to Contents](#contents)

# Chapter 2

Second chapter content.

[Back to Contents](#contents)

# Contents

Table of contents.
`
	path := writeTempFile(t, "issue89.md", doc)
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	// A 60x15 terminal gives the reported content dimensions: 56x8.
	m = applyWindowSize(m, 60, 15)
	m = pressKey(m, "home")
	if len(m.book.Pages) < 2 {
		t.Fatalf("expected at least two pages, got %d", len(m.book.Pages))
	}
	if len(m.book.Pages[0].Links) != 1 {
		t.Fatalf("page 0 has %d links, want 1: %+v", len(m.book.Pages[0].Links), m.book.Pages[0].Links)
	}
	if len(m.book.Pages[1].Links) != 1 {
		t.Fatalf("page 1 has %d links, want 1: %+v", len(m.book.Pages[1].Links), m.book.Pages[1].Links)
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("Tab on page 0 selected link %d, want 0", m.SelectedLink())
	}
	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("second Tab on page 0 selected link %d, want 0 (single link)", m.SelectedLink())
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
	targetPage := m.book.PageForAnchor("contents")
	if m.CurrentPage() != targetPage {
		t.Fatalf("Enter landed on page %d, want contents page %d", m.CurrentPage(), targetPage)
	}
	if !strings.Contains(strings.Join(m.book.Pages[m.CurrentPage()].Lines, "\n"), "# Contents") {
		t.Fatalf("target page does not contain heading: %v", m.book.Pages[m.CurrentPage()].Lines)
	}
}
