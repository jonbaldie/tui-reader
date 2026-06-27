package book_test

// Characterization tests that lock in the reader's current navigation behavior
// through the public book.Book interface only (NewBook, Pages, Links,
// PageForAnchor, Reflow). They deliberately live in the external book_test
// package so the compiler forbids reaching into internal state: these tests
// describe what a reader observes, not how the package computes it, and must
// survive the internal refactor that follows (collapsing the duplicate
// formatted-to-raw mapping behind a single one-pass formatter).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

// bookFromContent writes content to a temp file and builds a paginated Book,
// the only way to construct a Book through the public interface.
func bookFromContent(t *testing.T, content string, width, height int) *book.Book {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	b, err := book.NewBook(path, width, height)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	return b
}

// findLink returns the first link with the given target across all pages, plus
// the index of the page it sits on.
func findLink(b *book.Book, target string) (book.Link, int, bool) {
	for pi, p := range b.Pages {
		for _, l := range p.Links {
			if l.Target == target {
				return l, pi, true
			}
		}
	}
	return book.Link{}, -1, false
}

// pageContains reports whether any line on the page contains substr.
func pageContains(p book.Page, substr string) bool {
	for _, line := range p.Lines {
		if strings.Contains(line, substr) {
			return true
		}
	}
	return false
}

// A link's LineOnPage must index a line that actually contains its display
// text, so the reader's highlight sits on the text they see.
func TestNav_LinkLineIndexPointsAtItsDisplayText(t *testing.T) {
	content := "# Intro\n\nSee [Chapter 1](#chapter-1) for details.\n\n# Chapter 1\n\nChapter body.\n"
	b := bookFromContent(t, content, 60, 10)

	link, pi, ok := findLink(b, "chapter-1")
	if !ok {
		t.Fatal("link with target chapter-1 not found on any page")
	}
	page := b.Pages[pi]
	if link.LineOnPage < 0 || link.LineOnPage >= len(page.Lines) {
		t.Fatalf("LineOnPage = %d, out of range for page of %d lines", link.LineOnPage, len(page.Lines))
	}
	if got := page.Lines[link.LineOnPage]; !strings.Contains(got, link.Label) {
		t.Errorf("LineOnPage %d = %q, does not contain link label %q", link.LineOnPage, got, link.Label)
	}
}

// Following a link (resolving its target anchor) must land on a page whose
// lines contain the target heading, so "follow link" reaches the heading.
func TestNav_FollowingLinkLandsOnHeadingPage(t *testing.T) {
	// The link label ("the next chapter") is distinct from the heading text
	// ("Chapter One"), so finding "Chapter One" proves we landed on the heading,
	// not merely back on the link line. Filler pushes the heading onto a later
	// page, exercising the pagination math in PageForAnchor.
	var sb strings.Builder
	sb.WriteString("# Intro\n\nJump to [the next chapter](#chapter-one).\n\n")
	for i := 0; i < 12; i++ {
		sb.WriteString("Filler paragraph that occupies space.\n\n")
	}
	sb.WriteString("# Chapter One\n\nBody of chapter one.\n")
	b := bookFromContent(t, sb.String(), 60, 6)

	link, _, ok := findLink(b, "chapter-one")
	if !ok {
		t.Fatal("link with target chapter-one not found")
	}

	page := b.PageForAnchor(link.Target)
	if page < 0 {
		t.Fatalf("PageForAnchor(%q) = -1, anchor not resolved", link.Target)
	}
	if page >= len(b.Pages) {
		t.Fatalf("PageForAnchor(%q) = %d, out of range for %d pages", link.Target, page, len(b.Pages))
	}
	if !pageContains(b.Pages[page], "Chapter One") {
		t.Errorf("page %d does not contain target heading %q; lines=%v", page, "Chapter One", b.Pages[page].Lines)
	}
}

// countLinks returns the total number of links attached across all pages.
func countLinks(b *book.Book) int {
	n := 0
	for _, p := range b.Pages {
		n += len(p.Links)
	}
	return n
}

// A link that appears once in the source must be selectable exactly once: the
// reader tabbing through links must not hit a phantom duplicate.
func TestNav_SingleLinkAttachedExactlyOnce(t *testing.T) {
	b := bookFromContent(t, "See [Chapter 1](#chapter-1) here.\n", 80, 20)
	if got := countLinks(b); got != 1 {
		t.Fatalf("expected exactly 1 attached link, got %d", got)
	}
}

// Two distinct links on the same line must both be selectable, so the reader
// can reach either target.
func TestNav_TwoDistinctLinksSameLineBothAttached(t *testing.T) {
	b := bookFromContent(t, "[One](#one) and [Two](#two).\n", 80, 20)
	if got := countLinks(b); got != 2 {
		t.Fatalf("expected 2 attached links, got %d", got)
	}
	if _, _, ok := findLink(b, "one"); !ok {
		t.Error("link target one not attached")
	}
	if _, _, ok := findLink(b, "two"); !ok {
		t.Error("link target two not attached")
	}
}

// linkLineHasLabel reports whether the link's highlighted line contains label.
func linkLineHasLabel(b *book.Book, target, label string) bool {
	l, pi, ok := findLink(b, target)
	if !ok {
		return false
	}
	p := b.Pages[pi]
	if l.LineOnPage < 0 || l.LineOnPage >= len(p.Lines) {
		return false
	}
	return strings.Contains(p.Lines[l.LineOnPage], label)
}

// followLandsOnHeading reports whether resolving target lands on a page whose
// lines contain headingText.
func followLandsOnHeading(b *book.Book, target, headingText string) bool {
	page := b.PageForAnchor(target)
	if page < 0 || page >= len(b.Pages) {
		return false
	}
	return pageContains(b.Pages[page], headingText)
}

// Resizing the terminal (Reflow) must not change navigation outcomes: a link
// still highlights its own text and still leads to its heading, even though
// page indices and wrapping change.
func TestNav_OutcomesStableAcrossReflow(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("# Intro\n\nRead [onward](#chapter-two) to continue.\n\n")
	for i := 0; i < 12; i++ {
		sb.WriteString("Filler paragraph that occupies space.\n\n")
	}
	sb.WriteString("# Chapter Two\n\nThe body of chapter two.\n")
	b := bookFromContent(t, sb.String(), 60, 6)

	beforeLabel := linkLineHasLabel(b, "chapter-two", "onward")
	beforeHeading := followLandsOnHeading(b, "chapter-two", "Chapter Two")
	if !beforeLabel || !beforeHeading {
		t.Fatalf("precondition failed before reflow: lineHasLabel=%v landsOnHeading=%v", beforeLabel, beforeHeading)
	}

	b.Reflow(44, 9)

	afterLabel := linkLineHasLabel(b, "chapter-two", "onward")
	afterHeading := followLandsOnHeading(b, "chapter-two", "Chapter Two")
	if afterLabel != beforeLabel {
		t.Errorf("link-line-has-label changed across reflow: %v -> %v", beforeLabel, afterLabel)
	}
	if afterHeading != beforeHeading {
		t.Errorf("follow-lands-on-heading changed across reflow: %v -> %v", beforeHeading, afterHeading)
	}
}

// Degenerate inputs must still paginate into at least one well-formed page
// without panicking, so edge cases never crash the reader.
//
// Note: this characterizes today's behavior, where a link-free page carries a
// nil Links slice (a nil slice still has len 0 and ranges safely). We therefore
// assert len(Links) == 0 rather than Links != nil; normalizing Links to a
// non-nil empty slice belongs to the slice that rebuilds AttachLinks.
func TestNav_DegenerateInputsAreWellFormed(t *testing.T) {
	cases := map[string]string{
		"empty":          "",
		"whitespaceOnly": "   \n\t\n   ",
		"noHeading":      "Just some text.\n\nMore plain text here.",
		"noLink":         "# Heading\n\nBody paragraph with no links.",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			b := bookFromContent(t, content, 60, 10)
			if len(b.Pages) < 1 {
				t.Fatalf("expected >=1 page, got %d", len(b.Pages))
			}
			for i, p := range b.Pages {
				if p.Lines == nil {
					t.Errorf("page %d has nil Lines", i)
				}
				if len(p.Links) != 0 {
					t.Errorf("page %d has %d links, want 0", i, len(p.Links))
				}
			}
			// No anchors exist for these inputs, so resolution must report -1
			// rather than panic.
			if got := b.PageForAnchor("anything"); got != -1 {
				t.Errorf("PageForAnchor on anchorless doc = %d, want -1", got)
			}
		})
	}
}
