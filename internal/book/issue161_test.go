package book

import (
	"strings"
	"testing"
)

func TestIssue161_InlineCodeMatchingLinkMarkupDoesNotStealAttachment(t *testing.T) {
	content := "`[a](#t)` [a](#t)"
	b, err := Read(strings.NewReader(content), "x", 10, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.pages) < 2 {
		t.Fatalf("expected at least 2 pages, got %d", len(b.pages))
	}
	if len(b.pages[0].Links) != 0 {
		t.Errorf("expected page 0 to have 0 links, got %d", len(b.pages[0].Links))
	}
	if len(b.pages[1].Links) != 1 {
		t.Fatalf("expected page 1 to have 1 link, got %d", len(b.pages[1].Links))
	}
	if link := b.pages[1].Links[0]; link.Label != "a" || link.Target != "t" || link.LineOnPage != 0 {
		t.Errorf("unexpected link on page 1: %+v", link)
	}
	assertLinksLandWhereRendered(t, b, 1)

	// Reflow to another geometry
	b.Reflow(40, 2)
	assertLinksLandWhereRendered(t, b, 1)
}

func TestIssue161_InlineCodeWithLinkInParagraph(t *testing.T) {
	content := "`[a](#t)` filler filler filler filler filler\nfiller filler filler [a](#t)"
	for _, size := range [][2]int{{40, 2}, {20, 2}, {30, 3}, {80, 5}} {
		b, err := Read(strings.NewReader(content), "x", size[0], size[1], false)
		if err != nil {
			t.Fatal(err)
		}
		assertLinksLandWhereRendered(t, b, 1)
	}
}

func TestIssue161_MultipleLinksWithInlineCode(t *testing.T) {
	content := "`[first](#1)` [first](#1) middle `[second](#2)` [second](#2)"
	for _, size := range [][2]int{{15, 2}, {30, 2}, {50, 4}} {
		b, err := Read(strings.NewReader(content), "x", size[0], size[1], false)
		if err != nil {
			t.Fatal(err)
		}
		assertLinksLandWhereRendered(t, b, 2)
	}
}
