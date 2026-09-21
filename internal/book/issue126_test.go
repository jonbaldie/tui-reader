package book

import "testing"

func TestDefaultPageGeometry(t *testing.T) {
	if DefaultPageWidth != 60 {
		t.Errorf("DefaultPageWidth = %d, want 60", DefaultPageWidth)
	}
	if DefaultPageHeight != 20 {
		t.Errorf("DefaultPageHeight = %d, want 20", DefaultPageHeight)
	}
}

func TestBookRecordsEffectivePageHeight(t *testing.T) {
	path := writeTempFile(t, "height.md", "one\ntwo\n")

	t.Run("NewBook", func(t *testing.T) {
		b, err := NewBook(path, 60, 0)
		if err != nil {
			t.Fatalf("NewBook: %v", err)
		}
		if b.PageHeight != 20 {
			t.Errorf("PageHeight = %d, want effective default height 20", b.PageHeight)
		}
	})

	t.Run("Reflow", func(t *testing.T) {
		b, err := NewBook(path, 60, 20)
		if err != nil {
			t.Fatalf("NewBook: %v", err)
		}
		b.Reflow(60, 0)
		if b.PageHeight != 20 {
			t.Errorf("PageHeight = %d, want effective default height 20", b.PageHeight)
		}
	})
}

func TestBookRecordsEffectivePageWidth(t *testing.T) {
	path := writeTempFile(t, "width.md", "one\ntwo\n")

	t.Run("NewBook", func(t *testing.T) {
		b, err := NewBook(path, 0, DefaultPageHeight)
		if err != nil {
			t.Fatalf("NewBook: %v", err)
		}
		if b.PageWidth != 80 {
			t.Errorf("PageWidth = %d, want effective fallback width 80", b.PageWidth)
		}
	})

	t.Run("Reflow", func(t *testing.T) {
		b, err := NewBook(path, DefaultPageWidth, DefaultPageHeight)
		if err != nil {
			t.Fatalf("NewBook: %v", err)
		}
		b.Reflow(0, DefaultPageHeight)
		if b.PageWidth != 80 {
			t.Errorf("PageWidth = %d, want effective fallback width 80", b.PageWidth)
		}
	})
}
