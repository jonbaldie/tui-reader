package book

import (
	"strings"
	"testing"
)

func TestAttachLinks_LinkOnSecondPage(t *testing.T) {
	var raw []string
	for i := 0; i < 8; i++ {
		raw = append(raw, "Filler paragraph line.")
		raw = append(raw, "")
	}
	raw = append(raw, "Visit [Chapter 2](#chapter-2) now.")

	width, height := 80, 5
	pages := Paginate(raw, width, height)
	pages = AttachLinks(pages, raw, width, height)

	if len(pages) < 2 {
		t.Fatalf("expected >=2 pages, got %d", len(pages))
	}

	pagesWithLink := 0
	linkPage := -1
	for pi, p := range pages {
		if len(p.Links) > 0 {
			pagesWithLink++
			linkPage = pi
		}
	}
	if pagesWithLink != 1 {
		t.Fatalf("expected exactly 1 page with links, got %d", pagesWithLink)
	}
	if linkPage != len(pages)-1 {
		t.Errorf("link on page %d, want last page %d", linkPage, len(pages)-1)
	}

	link := pages[linkPage].Links[0]
	if link.Target != "chapter-2" {
		t.Errorf("target = %q, want chapter-2", link.Target)
	}
	if link.LineOnPage < 0 || link.LineOnPage >= len(pages[linkPage].Lines) {
		t.Fatalf("LineOnPage = %d, out of range for page with %d lines", link.LineOnPage, len(pages[linkPage].Lines))
	}
	if !strings.Contains(pages[linkPage].Lines[link.LineOnPage], "Chapter 2") {
		t.Errorf("LineOnPage %d does not point at link line: %q", link.LineOnPage, pages[linkPage].Lines[link.LineOnPage])
	}
}

func TestAttachLinks_DeduplicatesSameLink(t *testing.T) {
	raw := []string{"See [Chapter 1](#chapter-1) here."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)

	if len(pages[0].Links) != 1 {
		t.Fatalf("expected exactly 1 link, got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
}

func TestAttachLinks_TwoDistinctLinksSameLine(t *testing.T) {
	raw := []string{"[One](#one) and [Two](#two)."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)

	if len(pages[0].Links) != 2 {
		t.Fatalf("expected 2 distinct links, got %d", len(pages[0].Links))
	}

	targets := map[string]bool{}
	for _, l := range pages[0].Links {
		targets[l.Target] = true
	}
	if !targets["one"] || !targets["two"] {
		t.Errorf("expected targets one+two, got %v", targets)
	}
}

func TestAttachLinks_NoLinksLeavesEmpty(t *testing.T) {
	raw := []string{"Just some text.", "More text."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)

	for pi, p := range pages {
		if len(p.Links) != 0 {
			t.Errorf("page %d has %d links, want 0", pi, len(p.Links))
		}
	}
}
