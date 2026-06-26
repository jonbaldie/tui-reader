package tui

import (
	"strings"
	"testing"
)

// ==================== NewModel defaults ====================

func TestNewModel_DefaultFieldValues(t *testing.T) {
	path := writeTempFile(t, "d.md", "# Hi\n\nthere\n")
	m := NewModel(path)
	if m.SelectedLink() != -1 {
		t.Errorf("default selectedLink = %d, want -1", m.SelectedLink())
	}
	if m.contentWidth != 60 {
		t.Errorf("default contentWidth = %d, want 60", m.contentWidth)
	}
	if m.contentHeight != 20 {
		t.Errorf("default contentHeight = %d, want 20", m.contentHeight)
	}
	if m.CurrentPage() != 0 {
		t.Errorf("default currentPage = %d, want 0", m.CurrentPage())
	}
}

// ==================== recalcLayout dimensions ====================

func TestRecalcLayout_ContentWidthFromTermWidth(t *testing.T) {
	path := writeTempFile(t, "w.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 50, 40)
	// 50 - 4 padding = 46 (below the 72 cap, above the 20 floor).
	if m.contentWidth != 46 {
		t.Errorf("contentWidth = %d, want 46 (termWidth-4)", m.contentWidth)
	}
}

func TestRecalcLayout_ContentWidthCappedAt72(t *testing.T) {
	path := writeTempFile(t, "w.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 200, 40)
	if m.contentWidth != 72 {
		t.Errorf("contentWidth = %d, want 72 (cap)", m.contentWidth)
	}
}

func TestRecalcLayout_ContentHeightFromTermHeight(t *testing.T) {
	path := writeTempFile(t, "w.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 30)
	// 30 - 7 = 23 (above the 5 floor).
	if m.contentHeight != 23 {
		t.Errorf("contentHeight = %d, want 23 (termHeight-7)", m.contentHeight)
	}
}

func TestRecalcLayout_ContentHeightFloorIsFive(t *testing.T) {
	path := writeTempFile(t, "w.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 8) // 8-7 = 1, must clamp up to exactly 5
	if m.contentHeight != 5 {
		t.Errorf("contentHeight = %d, want exactly 5 (floor)", m.contentHeight)
	}
}

// ==================== recalcLayout page/link handling on resize ====================

func TestRecalcLayout_KeepsCurrentPageOnResize(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 20)
	if len(m.BookRef().Pages) < 3 {
		t.Fatalf("need >=3 pages, got %d", len(m.BookRef().Pages))
	}
	m = pressKey(m, "right")
	m = pressKey(m, "right")
	if m.CurrentPage() != 2 {
		t.Fatalf("setup: expected page 2, got %d", m.CurrentPage())
	}
	// Resize to the same dimensions: page 2 still exists, so the current page
	// must be preserved (not reset to 0).
	m = applyWindowSize(m, 80, 20)
	if m.CurrentPage() != 2 {
		t.Errorf("currentPage after same-size resize = %d, want 2 (must not reset)", m.CurrentPage())
	}
}

func TestRecalcLayout_ResetsSelectedLinkOnResize(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 20)
	if len(m.BookRef().Pages[0].Links) == 0 {
		t.Skip("page 0 has no links in this layout")
	}
	m = pressKey(m, "tab") // selects link 0
	if m.SelectedLink() != 0 {
		t.Fatalf("setup: expected selectedLink 0, got %d", m.SelectedLink())
	}
	m = applyWindowSize(m, 80, 20)
	if m.SelectedLink() != -1 {
		t.Errorf("selectedLink after resize = %d, want -1 (must reset)", m.SelectedLink())
	}
}

func TestRecalcLayout_ClampsPageWhenShrinking(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 40) // few pages
	// Go to the last page.
	m = pressKey(m, "end")
	last := m.CurrentPage()
	if last == 0 {
		t.Skip("only one page at this size")
	}
	// Shrink height a lot: more pages now, but the clamp must keep us in range.
	m = applyWindowSize(m, 80, 12)
	if m.CurrentPage() < 0 || m.CurrentPage() >= len(m.BookRef().Pages) {
		t.Errorf("currentPage %d out of range after shrink (pages=%d)", m.CurrentPage(), len(m.BookRef().Pages))
	}
}

// ==================== nil-book navigation (book != nil guards) ====================

func TestHandleKey_NilBookNeverPanics(t *testing.T) {
	m := NewModel("/no/such/file.md")
	if m.BookRef() != nil {
		t.Fatal("expected nil book for missing file")
	}
	for _, k := range []string{"right", "left", "home", "end", "tab", "shift+tab", "enter", "b"} {
		m = pressKey(m, k) // must not panic dereferencing a nil book
		if m.CurrentPage() != 0 {
			t.Errorf("after %q on nil book, currentPage = %d, want 0", k, m.CurrentPage())
		}
	}
}

// ==================== link selection wrap-around ====================

func TestPrevLink_StopsAtZeroWithoutWrapping(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 20)
	if len(m.BookRef().Pages[0].Links) < 2 {
		t.Skipf("need >=2 links on page 0, got %d", len(m.BookRef().Pages[0].Links))
	}
	m = pressKey(m, "tab") // -> 0
	m = pressKey(m, "tab") // -> 1
	if m.SelectedLink() != 1 {
		t.Fatalf("setup: selectedLink = %d, want 1", m.SelectedLink())
	}
	m = pressKey(m, "shift+tab") // 1 -> 0; must NOT wrap to the last index
	if m.SelectedLink() != 0 {
		t.Errorf("selectedLink after shift+tab from 1 = %d, want 0", m.SelectedLink())
	}
}

// ==================== follow link to a page-0 anchor ====================

func selfLinkDoc() string {
	var sb strings.Builder
	sb.WriteString("# Top Section\n\n")
	sb.WriteString("[Back to top](#top-section)\n\n")
	for i := 0; i < 40; i++ {
		sb.WriteString("Filler paragraph text.\n\n")
	}
	return sb.String()
}

func TestFollowLink_TargetOnFirstPageFollowsAndResets(t *testing.T) {
	path := writeTempFile(t, "self.md", selfLinkDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 20)

	if dest := m.BookRef().PageForAnchor("top-section"); dest != 0 {
		t.Fatalf("setup: anchor resolves to page %d, want 0", dest)
	}
	if len(m.BookRef().Pages[0].Links) == 0 {
		t.Fatalf("setup: no links on page 0")
	}

	m = pressKey(m, "tab") // select the link
	if m.SelectedLink() < 0 {
		t.Fatalf("setup: no link selected")
	}
	m = pressKey(m, "enter") // follow; dest == 0 must still be followed (>= 0)

	if len(m.History()) != 1 {
		t.Errorf("history length = %d, want 1 (a dest of 0 must still push history)", len(m.History()))
	}
	if m.SelectedLink() != -1 {
		t.Errorf("selectedLink after follow = %d, want -1", m.SelectedLink())
	}
	if m.CurrentPage() != 0 {
		t.Errorf("currentPage after follow = %d, want 0", m.CurrentPage())
	}
}

// ==================== isHeading ====================

func TestIsHeading(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"# Heading", true},
		{"### Deep", true},
		{"   # Indented heading", true},
		{"plain text", false},
		{"", false},
		{"   ", false},
		{"a # not heading", false},
	}
	for _, c := range cases {
		if got := isHeading(c.line); got != c.want {
			t.Errorf("isHeading(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}
