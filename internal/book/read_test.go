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
