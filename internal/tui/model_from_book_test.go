package tui

import (
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

func TestNewModelFromBook(t *testing.T) {
	b, err := book.Read(strings.NewReader("# In Memory\n\nReadable content.\n"), "In Memory", book.DefaultPageWidth, book.DefaultPageHeight, false)
	if err != nil {
		t.Fatalf("book.Read: %v", err)
	}

	m := NewModelFromBook(b)
	if m.Err() != nil {
		t.Fatalf("unexpected model error: %v", m.Err())
	}
	if m.BookRef() != b {
		t.Fatal("model does not retain the supplied book")
	}
	if m.CurrentPage() != 0 {
		t.Errorf("initial page = %d, want 0", m.CurrentPage())
	}
}
