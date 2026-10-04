package book

import (
	"strings"
	"testing"
)

// readBook builds a Book through Read, the in-memory construction path that
// production uses. plainText selects plain-text layout instead of Markdown.
func readBook(t testing.TB, doc string, plainText bool, w, h int) *Book {
	t.Helper()
	b, err := Read(strings.NewReader(doc), "", w, h, plainText)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return b
}

// readMarkdown reads raw lines as one Markdown document.
func readMarkdown(t testing.TB, raw []string, w, h int) *Book {
	t.Helper()
	return readBook(t, strings.Join(raw, "\n"), false, w, h)
}

// displayLines returns the book's display lines across all pages, in order.
func displayLines(b *Book) []string {
	var lines []string
	for _, p := range b.Pages {
		lines = append(lines, p.Lines...)
	}
	return lines
}

// formatMarkdown returns the display lines Read lays out for raw Markdown at
// the given width.
func formatMarkdown(t testing.TB, raw []string, width int) []string {
	t.Helper()
	return displayLines(readMarkdown(t, raw, width, DefaultPageHeight))
}
