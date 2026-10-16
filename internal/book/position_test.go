package book

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// longParagraph returns one raw paragraph of n three-rune words "w01".."wNN",
// so a position's offset of k non-space runes falls inside word k/3+1.
func longParagraph(n int) string {
	words := make([]string, n)
	for i := range words {
		words[i] = fmt.Sprintf("w%02d", i+1)
	}
	return strings.Join(words, " ")
}

func pageWords(b *Book, page int) []string {
	return strings.Fields(strings.Join(b.Page(page).Lines, " "))
}

func TestPositionForPage_ClampsOutOfRangePages(t *testing.T) {
	b := readMarkdown(t, []string{longParagraph(60)}, 20, 2)
	last := b.PageCount() - 1
	if last < 2 {
		t.Fatalf("test setup: want at least 3 pages, got %d", b.PageCount())
	}
	if got, want := b.PositionForPage(-1), b.PositionForPage(0); got != want {
		t.Errorf("PositionForPage(-1) = %+v, want page 0's %+v", got, want)
	}
	if got, want := b.PositionForPage(last+1), b.PositionForPage(last); got != want {
		t.Errorf("PositionForPage(%d) = %+v, want last page's %+v", last+1, got, want)
	}
	if b.PositionForPage(last) == b.PositionForPage(0) {
		t.Errorf("test setup: last page and page 0 share position %+v", b.PositionForPage(0))
	}
}

func TestPositionForPage_SinglePageAnchorsAtFirstContentLine(t *testing.T) {
	b := readMarkdown(t, []string{"", "", "Text here."}, 40, 5)
	if b.PageCount() != 1 {
		t.Fatalf("test setup: want 1 page, got %d", b.PageCount())
	}
	want := Position{Raw: 2}
	for _, page := range []int{-1, 0, 1} {
		if got := b.PositionForPage(page); got != want {
			t.Errorf("PositionForPage(%d) = %+v, want %+v", page, got, want)
		}
	}
}

func TestPageForPosition_RoundTripsEveryPageInsideOneParagraph(t *testing.T) {
	b := readMarkdown(t, []string{longParagraph(60), "", "After."}, 20, 2)
	for page := range b.PageCount() {
		if got := b.PageForPosition(b.PositionForPage(page)); got != page {
			t.Errorf("PageForPosition(PositionForPage(%d)) = %d", page, got)
		}
	}
}

func TestPageForPosition_SurvivesReflowInsideParagraph(t *testing.T) {
	raw := []string{longParagraph(60), "", "After."}
	b := readMarkdown(t, raw, 20, 2)
	wide := readMarkdown(t, raw, 20, 2)
	wide.Reflow(45, 3)
	for page := 1; page < b.PageCount(); page++ {
		pos := b.PositionForPage(page)
		if pos.Raw != 0 {
			continue
		}
		word := fmt.Sprintf("w%02d", pos.Offset/3+1)
		got := wide.PageForPosition(pos)
		if !slices.Contains(pageWords(wide, got), word) {
			t.Errorf("position %+v (word %s) mapped to page %d showing %q", pos, word, got, pageWords(wide, got))
		}
	}
}

func TestPageForPosition_OffsetPastParagraphEndStaysInParagraph(t *testing.T) {
	b := readMarkdown(t, []string{longParagraph(60), "", "After."}, 20, 2)
	got := b.PageForPosition(Position{Raw: 0, Offset: 1000})
	if !slices.Contains(pageWords(b, got), "w60") {
		t.Errorf("offset past the paragraph end mapped to page %d showing %q, want the page ending the paragraph",
			got, pageWords(b, got))
	}
}

func TestPageForPosition_FallsBackToPageForRawLine(t *testing.T) {
	b := readMarkdown(t, []string{longParagraph(60), "", "After."}, 20, 2)
	for _, pos := range []Position{
		{Raw: 2, Offset: 0},   // offset 0: no search needed
		{Raw: 999, Offset: 5}, // raw line not displayed
	} {
		if got, want := b.PageForPosition(pos), b.PageForRawLine(pos.Raw); got != want {
			t.Errorf("PageForPosition(%+v) = %d, want PageForRawLine = %d", pos, got, want)
		}
	}
	if got, want := b.PageForRawLine(2), b.PageCount()-1; got != want {
		t.Errorf("test setup: PageForRawLine(2) = %d, want last page %d", got, want)
	}
}
