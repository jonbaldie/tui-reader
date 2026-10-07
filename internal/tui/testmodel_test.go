package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/jonbaldie/tui-reader/internal/book"
)

// newTestModel builds a model from an in-memory document. name selects the
// title and plain-text routing the same way book.NewBook does for a path,
// without writing a file.
func newTestModel(tb testing.TB, name, content string) Model {
	tb.Helper()
	b, err := book.Read(strings.NewReader(content), titleFromName(name), book.DefaultPageWidth, book.DefaultPageHeight, !markdownName(name))
	if err != nil {
		tb.Fatalf("book.Read: %v", err)
	}
	return NewModelFromBook(b)
}

// newTestModelFromFile loads a fixture path through book.NewBook. Use it when
// the subject is file-type detection or a title derived from a real path.
func newTestModelFromFile(tb testing.TB, path string) Model {
	tb.Helper()
	b, err := book.NewBook(path, book.DefaultPageWidth, book.DefaultPageHeight)
	if err != nil {
		tb.Fatalf("book.NewBook: %v", err)
	}
	return NewModelFromBook(b)
}

// newErrorTestModel builds the error screen for a path that fails to load.
func newErrorTestModel(tb testing.TB, path string) Model {
	tb.Helper()
	_, err := book.NewBook(path, book.DefaultPageWidth, book.DefaultPageHeight)
	if err == nil {
		tb.Fatalf("book.NewBook(%s) succeeded, want a load error", path)
	}
	return NewErrorModel(err)
}

func markdownName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

// titleFromName matches book title derivation for a path's base name.
func titleFromName(name string) string {
	parts := strings.Split(name, "/")
	base := parts[len(parts)-1]
	parts = strings.Split(base, "\\")
	base = parts[len(parts)-1]
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r
		return r
	}, base)
}
