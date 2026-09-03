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

func TestAttachLinks_WrappedLinkKeepsLabelOnOneLine(t *testing.T) {
	raw := []string{"# Target", "", "[Open Section 1](#section-1)"}
	pages := AttachLinks(Paginate(raw, 20, 20), raw, 20, 20)

	var links []Link
	for _, page := range pages {
		links = append(links, page.Links...)
	}
	if len(links) != 1 {
		t.Fatalf("attached links = %d, want exactly one", len(links))
	}

	link := links[0]
	page := pages[0]
	if !strings.Contains(page.Lines[link.LineOnPage], link.Label) {
		t.Fatalf("selected line %q does not contain label %q", page.Lines[link.LineOnPage], link.Label)
	}
}

func TestWrapLines_PreservesCompleteInternalLinkMarkup(t *testing.T) {
	link := "[Open Section 1](#section-1)"
	// Width 28 fits the link markup exactly, so it is kept whole and isolated
	// on its own line. The over-wide case (link wider than the page) is covered
	// by TestLinkOverflow_WideMarkupWrapsToWidth.
	got := WrapLines([]string{"before " + link + " after"}, 28)
	want := []string{"before", link, "after"}
	if len(got) != len(want) {
		t.Fatalf("wrapped lines = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWrapLines_PreservesPunctuationAfterInternalLink(t *testing.T) {
	link := "[Open Section 1](#section-1)"
	got := WrapLines([]string{link + ", then continue"}, 28)
	want := []string{link, ", then continue"}
	if len(got) != len(want) {
		t.Fatalf("wrapped lines = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAttachLinks_WrappedSourceLinkAttachedOnce(t *testing.T) {
	raw := []string{strings.Repeat("padding ", 10) + "[Chapter 1](#chapter-1) " + strings.Repeat("more ", 10)}
	pages := Paginate(raw, 20, 20)
	pages = AttachLinks(pages, raw, 20, 20)

	var links []Link
	for _, page := range pages {
		links = append(links, page.Links...)
	}
	if len(links) != 1 {
		t.Fatalf("got %d attached links, want one: %+v", len(links), links)
	}
	if links[0].Target != "chapter-1" {
		t.Errorf("target = %q, want chapter-1", links[0].Target)
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

func TestAttachLinks_RepeatedIdenticalLinksAcrossLines(t *testing.T) {
	raw := []string{"[Target](#target) and " + strings.Repeat("filler words ", 10) + " and [Target](#target)"}
	pages := Paginate(raw, 40, 20)
	pages = AttachLinks(pages, raw, 40, 20)

	if len(pages[0].Links) != 2 {
		t.Fatalf("expected 2 attached links, got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
	if pages[0].Links[0].LineOnPage == pages[0].Links[1].LineOnPage {
		t.Errorf("expected links on different lines, got both on line %d", pages[0].Links[0].LineOnPage)
	}
}
