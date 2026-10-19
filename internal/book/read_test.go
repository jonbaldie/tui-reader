package book

import (
	"strings"
	"testing"
)

// readBook builds test documents through the same boundary as production.
func readBook(t testing.TB, doc string, md bool, w, h int) *Book {
	t.Helper()
	b, err := Read(strings.NewReader(doc), "Test", w, h, !md)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return b
}

// readDisplayLines projects all pages of a Markdown book for formatting
// assertions, without assuming that the document fits on a single page.
func readDisplayLines(t testing.TB, raw []string, width int) []string {
	t.Helper()
	b := readBook(t, strings.Join(raw, "\n"), true, width, DefaultPageHeight)
	lines := make([]string, 0)
	for i := 0; i < b.PageCount(); i++ {
		lines = append(lines, b.Page(i).Lines...)
	}
	return lines
}
