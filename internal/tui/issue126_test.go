package tui

import (
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

func TestNewModelUsesBookDefaultPageGeometry(t *testing.T) {
	m := NewModel(writeTempFile(t, "default-geometry.md", "# Title\n"))

	if m.contentWidth != book.DefaultPageWidth {
		t.Errorf("contentWidth = %d, want %d", m.contentWidth, book.DefaultPageWidth)
	}
	if m.contentHeight != book.DefaultPageHeight {
		t.Errorf("contentHeight = %d, want %d", m.contentHeight, book.DefaultPageHeight)
	}
	if m.BookRef().PageWidth != book.DefaultPageWidth {
		t.Errorf("book PageWidth = %d, want %d", m.BookRef().PageWidth, book.DefaultPageWidth)
	}
	if m.BookRef().PageHeight != book.DefaultPageHeight {
		t.Errorf("book PageHeight = %d, want %d", m.BookRef().PageHeight, book.DefaultPageHeight)
	}
}
