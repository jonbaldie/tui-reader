package book_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

var renderedLink = regexp.MustCompile(`\[([^\]]+)\]\(#([^)]+)\)`)

func TestIssue176_ReflowKeepsRenderedLinksAndPageLinksConsistent(t *testing.T) {
	content := "# Start\n\nSee [the index](#index) and [the start](#start) for more.\n\n" +
		"Some padding text that wraps differently at each width so the links move.\n\n" +
		"Then [the index](#index) again.\n\n# Index\n\nEnd.\n"
	b, err := book.Read(strings.NewReader(content), "t", 80, 6, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, width := range []int{80, 30, 50} {
		b.Reflow(width, 6)
		rendered := 0
		for p := 0; p < b.PageCount(); p++ {
			page := b.Page(p)
			for i, line := range page.Lines {
				for _, m := range renderedLink.FindAllStringSubmatch(line, -1) {
					rendered++
					want := book.Link{Label: m[1], Target: m[2], LineOnPage: i}
					if !slices.Contains(page.Links, want) {
						t.Errorf("width %d page %d: rendered link %+v missing from Links %+v", width, p, want, page.Links)
					}
				}
			}
		}
		if rendered < 3 {
			t.Errorf("width %d: found %d rendered links, want at least 3", width, rendered)
		}
	}
}

func TestIssue176_MutatingReturnedPageDoesNotChangeBook(t *testing.T) {
	b, err := book.Read(strings.NewReader("See [the index](#index).\n\n# Index\n"), "t", 80, 20, false)
	if err != nil {
		t.Fatal(err)
	}

	page := b.Page(0)
	if len(page.Lines) == 0 || len(page.Links) == 0 {
		t.Fatalf("page 0 = %+v, want lines and links", page)
	}
	page.Lines[0] = "overwritten"
	page.Links[0] = book.Link{Label: "x", Target: "y"}

	again := b.Page(0)
	if got, want := again.Lines[0], "See [the index](#index)."; got != want {
		t.Errorf("Lines[0] = %q, want %q", got, want)
	}
	if got, want := again.Links[0], (book.Link{Label: "the index", Target: "index"}); got != want {
		t.Errorf("Links[0] = %+v, want %+v", got, want)
	}
}

func TestIssue176_PageHeightReportsLaidOutHeight(t *testing.T) {
	b, err := book.Read(strings.NewReader("a\n"), "t", 80, 7, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.PageHeight(); got != 7 {
		t.Errorf("PageHeight() = %d, want 7", got)
	}
	b.Reflow(80, 0)
	if got := b.PageHeight(); got != book.DefaultPageHeight {
		t.Errorf("PageHeight() after Reflow(80, 0) = %d, want %d", got, book.DefaultPageHeight)
	}
}

