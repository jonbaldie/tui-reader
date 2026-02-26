package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ==================== BUG: 'b' key conflict ====================
// 'b' is bound to BOTH GoBack AND could conflict with other bindings.
// The key is in the GoBack binding. But what happens if user wants
// to use 'b' and there's no history? It should be a no-op.

func TestAdversarial_BKeyNoHistory(t *testing.T) {
	path := writeTempFile(t, "adv.md", "# Title\n\nText\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	pageBefore := m.CurrentPage()
	m = pressKey(m, "b")
	if m.CurrentPage() != pageBefore {
		t.Errorf("'b' with no history changed page: %d -> %d", pageBefore, m.CurrentPage())
	}
}

// ==================== BUG: 'b' key shadows backspace ====================
// Both 'b' and 'backspace' are GoBack. Verify backspace works too.

func TestAdversarial_BackspaceGoBack(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	origPage := m.CurrentPage()
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	// Use backspace instead of 'b'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(Model)
	if m.CurrentPage() != origPage {
		t.Errorf("backspace should go back to %d, got %d", origPage, m.CurrentPage())
	}
}

// ==================== Tiny terminal ====================

func TestAdversarial_TinyTerminal(t *testing.T) {
	path := writeTempFile(t, "adv.md", "# Hello\n\nWorld\n")
	m := NewModel(path)
	// 1x1 terminal - should not panic
	m = applyWindowSize(m, 1, 1)
	view := m.View()
	_ = view // just verifying no panic
}

func TestAdversarial_ZeroTerminal(t *testing.T) {
	path := writeTempFile(t, "adv.md", "# Hello\n\nWorld\n")
	m := NewModel(path)
	// 0x0 terminal
	m = applyWindowSize(m, 0, 0)
	view := m.View()
	_ = view
}

// ==================== Rapid resize ====================

func TestAdversarial_RapidResize(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)

	sizes := []struct{ w, h int }{
		{80, 24}, {40, 12}, {200, 50}, {20, 5}, {1, 1}, {120, 40}, {60, 20},
	}
	for _, s := range sizes {
		m = applyWindowSize(m, s.w, s.h)
		view := m.View()
		_ = view
		if m.CurrentPage() < 0 {
			t.Errorf("negative page after resize to %dx%d", s.w, s.h)
		}
		if m.BookRef() != nil && m.CurrentPage() >= len(m.BookRef().Pages) {
			t.Errorf("page %d exceeds max %d after resize to %dx%d",
				m.CurrentPage(), len(m.BookRef().Pages)-1, s.w, s.h)
		}
	}
}

// ==================== Navigate then resize ====================

func TestAdversarial_NavigateThenResize(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 40, 5) // many pages

	// Go to last page
	m = pressKey(m, "end")
	lastPage := m.CurrentPage()

	// Resize much bigger - should clamp
	m = applyWindowSize(m, 200, 100)
	if m.CurrentPage() >= len(m.BookRef().Pages) {
		t.Errorf("page %d out of bounds (max %d) after resize",
			m.CurrentPage(), len(m.BookRef().Pages)-1)
	}
	// Page should be <= lastPage since there are fewer pages now
	if m.CurrentPage() > lastPage {
		t.Errorf("page increased after enlarging terminal: %d > %d", m.CurrentPage(), lastPage)
	}
}

// ==================== Multiple link follows (deep history) ====================

func TestAdversarial_DeepHistory(t *testing.T) {
	// Build a doc with links that chain: toc -> ch1 -> ch2 -> ch3
	// Use a large enough page height so heading + link fit on the same page
	var sb strings.Builder
	sb.WriteString("# TOC\n\n[Ch1](#chapter-1)\n\n")
	for i := 0; i < 30; i++ {
		sb.WriteString("filler\n")
	}
	sb.WriteString("# Chapter 1\n[Ch2](#chapter-2)\n\n")
	for i := 0; i < 30; i++ {
		sb.WriteString("filler\n")
	}
	sb.WriteString("# Chapter 2\n[Ch3](#chapter-3)\n\n")
	for i := 0; i < 30; i++ {
		sb.WriteString("filler\n")
	}
	sb.WriteString("# Chapter 3\n\nThe end.\n")

	path := writeTempFile(t, "deep.md", sb.String())
	m := NewModel(path)
	// Use height large enough that heading + link are on the same page
	m = applyWindowSize(m, 60, 32)

	b := m.BookRef()

	// Verify precondition: each chapter page has a link
	for _, anchor := range []string{"toc", "chapter-1", "chapter-2"} {
		pg := b.PageForAnchor(anchor)
		if pg < 0 || pg >= len(b.Pages) {
			t.Fatalf("anchor %q not found or out of range (page %d)", anchor, pg)
		}
		if len(b.Pages[pg].Links) == 0 {
			t.Fatalf("anchor %q on page %d has no links — heading and link are on different pages", anchor, pg)
		}
	}

	// Follow chain: toc -> ch1 -> ch2 -> ch3
	pages := []int{m.CurrentPage()}
	for i := 0; i < 3; i++ {
		m = pressKey(m, "tab")
		m = pressKey(m, "enter")
		pages = append(pages, m.CurrentPage())
	}

	if len(m.History()) != 3 {
		t.Fatalf("expected 3 history entries, got %d", len(m.History()))
	}

	// Now go back 3 times - should retrace in reverse
	for i := 2; i >= 0; i-- {
		m = pressKey(m, "b")
		if m.CurrentPage() != pages[i] {
			t.Errorf("back step %d: expected page %d, got %d", 3-i, pages[i], m.CurrentPage())
		}
	}

	if len(m.History()) != 0 {
		t.Errorf("expected empty history after full unwind, got %d", len(m.History()))
	}
}

// ==================== Space key for next page ====================

func TestAdversarial_SpaceNextPage(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updated.(Model)
	if m.CurrentPage() != 1 {
		t.Errorf("space should advance to page 1, got %d", m.CurrentPage())
	}
}

// ==================== L key for next page ====================

func TestAdversarial_LKeyNextPage(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = updated.(Model)
	if m.CurrentPage() != 1 {
		t.Errorf("'l' should advance to page 1, got %d", m.CurrentPage())
	}
}

// ==================== H key for prev page ====================

func TestAdversarial_HKeyPrevPage(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	m = pressKey(m, "right") // page 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = updated.(Model)
	if m.CurrentPage() != 0 {
		t.Errorf("'h' should go back to page 0, got %d", m.CurrentPage())
	}
}

// ==================== G key for first page ====================

func TestAdversarial_GKeyFirstPage(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	m = pressKey(m, "right")
	m = pressKey(m, "right")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updated.(Model)
	if m.CurrentPage() != 0 {
		t.Errorf("'g' should go to page 0, got %d", m.CurrentPage())
	}
}

// ==================== G key (capital) for last page ====================

func TestAdversarial_GCapKeyLastPage(t *testing.T) {
	path := writeTempFile(t, "adv.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = updated.(Model)
	lastPage := len(m.BookRef().Pages) - 1
	if m.CurrentPage() != lastPage {
		t.Errorf("'G' should go to page %d, got %d", lastPage, m.CurrentPage())
	}
}

// ==================== Unknown key does nothing ====================

func TestAdversarial_UnknownKey(t *testing.T) {
	path := writeTempFile(t, "adv.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	pageBefore := m.CurrentPage()
	linkBefore := m.SelectedLink()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	m = updated.(Model)
	if m.CurrentPage() != pageBefore || m.SelectedLink() != linkBefore {
		t.Error("unknown key should be a no-op")
	}
}

// ==================== View with nil book ====================

func TestAdversarial_ViewNilBook(t *testing.T) {
	// Construct a model with nil book but no error (shouldn't happen, but defensive)
	m := Model{
		book:       nil,
		err:        nil,
		termWidth:  80,
		termHeight: 24,
	}
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("nil book with no error should show 'Loading', got: %q", view)
	}
}
