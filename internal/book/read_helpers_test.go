package book

import (
	"strings"
	"testing"
)

// readMarkdown lays out raw Markdown lines through Read, the entry point
// production uses, so tests exercise the same pipeline as the reader (#145).
func readMarkdown(t testing.TB, rawLines []string, width, height int) *Book {
	t.Helper()
	b, err := Read(strings.NewReader(strings.Join(rawLines, "\n")), "", width, height, false)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return b
}

// readDisplayLines returns every display line of the Markdown layout of
// rawLines at width, in page order.
func readDisplayLines(t testing.TB, rawLines []string, width int) []string {
	t.Helper()
	lines := []string{}
	for _, p := range readMarkdown(t, rawLines, width, DefaultPageHeight).pages {
		lines = append(lines, p.Lines...)
	}
	return lines
}
