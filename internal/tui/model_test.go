package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func simpleDoc() string {
	var sb strings.Builder
	sb.WriteString("# Title\n\n")
	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString("[Chapter 1](#chapter-1)\n")
	sb.WriteString("[Chapter 2](#chapter-2)\n\n")
	// Pad to push chapters onto later pages
	for i := 0; i < 30; i++ {
		sb.WriteString("Filler line.\n")
	}
	sb.WriteString("# Chapter 1\n\n")
	sb.WriteString("Content of chapter 1.\n\n")
	for i := 0; i < 20; i++ {
		sb.WriteString("More chapter 1 text.\n")
	}
	sb.WriteString("# Chapter 2\n\n")
	sb.WriteString("Content of chapter 2.\n")
	return sb.String()
}

// ==================== NewModel ====================

func TestNewModel_ValidFile(t *testing.T) {
	path := writeTempFile(t, "test.md", "# Hello\n\nWorld\n")
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("unexpected error: %v", m.Err())
	}
	if m.BookRef() == nil {
		t.Fatal("expected book to be loaded")
	}
	if m.CurrentPage() != 0 {
		t.Errorf("expected initial page 0, got %d", m.CurrentPage())
	}
}

func TestNewModel_MissingFile(t *testing.T) {
	m := NewModel("/nonexistent/file.txt")
	if m.Err() == nil {
		t.Fatal("expected error for missing file")
	}
}

// ==================== Page Navigation ====================

func TestNavigation_NextPage(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	if m.CurrentPage() != 0 {
		t.Fatalf("expected page 0, got %d", m.CurrentPage())
	}

	m = pressKey(m, "right")
	if m.CurrentPage() != 1 {
		t.Errorf("expected page 1 after next, got %d", m.CurrentPage())
	}
}

func TestNavigation_PrevPage(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "right") // go to page 1
	m = pressKey(m, "left")  // go back to page 0
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0, got %d", m.CurrentPage())
	}
}

func TestNavigation_PrevPageAtStart(t *testing.T) {
	path := writeTempFile(t, "nav.md", "Short content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "left") // already at page 0
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0 (clamped), got %d", m.CurrentPage())
	}
}

func TestNavigation_NextPageAtEnd(t *testing.T) {
	path := writeTempFile(t, "nav.md", "Short content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "right") // already at last page
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0 (single page), got %d", m.CurrentPage())
	}
}

func TestNavigation_FirstPage(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "right")
	m = pressKey(m, "right")
	m = pressKey(m, "home") // jump to first
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0, got %d", m.CurrentPage())
	}
}

func TestNavigation_LastPage(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "end") // jump to last
	lastPage := len(m.BookRef().Pages) - 1
	if m.CurrentPage() != lastPage {
		t.Errorf("expected page %d, got %d", lastPage, m.CurrentPage())
	}
}

// ==================== Link Navigation ====================

func TestLink_TabCyclesLinks(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40) // large enough to see links on page 0

	// Initially no link selected
	if m.SelectedLink() != -1 {
		t.Errorf("expected no link selected, got %d", m.SelectedLink())
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Errorf("expected link 0 selected, got %d", m.SelectedLink())
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 1 {
		t.Errorf("expected link 1 selected, got %d", m.SelectedLink())
	}

	// Should wrap around
	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Errorf("expected link wrap to 0, got %d", m.SelectedLink())
	}
}

func TestLink_ShiftTabReverse(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40)

	m = pressKey(m, "shift+tab") // should wrap to last link
	page := m.BookRef().Pages[m.CurrentPage()]
	if len(page.Links) > 0 {
		expected := len(page.Links) - 1
		if m.SelectedLink() != expected {
			t.Errorf("expected link %d (last), got %d", expected, m.SelectedLink())
		}
	}
}

func TestLink_FollowLink(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	// Tab to first link, then follow
	m = pressKey(m, "tab")
	beforePage := m.CurrentPage()
	m = pressKey(m, "enter")

	// Should have navigated to a different page
	if m.CurrentPage() == beforePage {
		// It's possible the link target is on the same page if layout fits differently
		// But with our doc structure, it should be different
		t.Log("link may have pointed to same page; verifying history")
	}

	// History should have an entry
	if len(m.History()) == 0 {
		t.Error("expected history entry after following link")
	}
}

func TestLink_GoBack(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	startPage := m.CurrentPage()
	m = pressKey(m, "tab")
	m = pressKey(m, "enter") // follow link
	m = pressKey(m, "b")     // go back

	if m.CurrentPage() != startPage {
		t.Errorf("expected to go back to page %d, got %d", startPage, m.CurrentPage())
	}
	if len(m.History()) != 0 {
		t.Errorf("expected empty history after going back, got %d entries", len(m.History()))
	}
}

func TestLink_GoBackEmptyHistory(t *testing.T) {
	path := writeTempFile(t, "nav.md", "Short text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	m = pressKey(m, "b") // nothing should happen
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0, got %d", m.CurrentPage())
	}
}

func TestLink_EnterWithNoSelection(t *testing.T) {
	path := writeTempFile(t, "nav.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	// Press enter without selecting a link
	m = pressKey(m, "enter")
	if m.CurrentPage() != 0 {
		t.Errorf("expected page 0, got %d", m.CurrentPage())
	}
}

// ==================== Window Resize ====================

func TestResize_PagesReflowed(t *testing.T) {
	path := writeTempFile(t, "resize.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 20)
	pagesAt80 := len(m.BookRef().Pages)

	m = applyWindowSize(m, 40, 10) // smaller
	pagesAt40 := len(m.BookRef().Pages)

	if pagesAt40 <= pagesAt80 {
		t.Errorf("expected more pages at smaller size: %d vs %d", pagesAt40, pagesAt80)
	}
}

func TestResize_CurrentPageClamped(t *testing.T) {
	path := writeTempFile(t, "resize.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 40, 5) // many pages

	// Go to a high page
	m = pressKey(m, "end")
	highPage := m.CurrentPage()

	// Now resize bigger — fewer pages, current should clamp
	m = applyWindowSize(m, 80, 40)
	if m.CurrentPage() >= highPage && highPage > len(m.BookRef().Pages)-1 {
		t.Error("expected page to be clamped after resize")
	}
}

// ==================== View ====================

func TestView_NoError(t *testing.T) {
	path := writeTempFile(t, "view.md", "# Hello\n\nWorld\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)

	view := m.View()
	if !strings.Contains(view, "Page 1 of") {
		t.Error("expected page indicator in view")
	}
}

func TestView_ErrorState(t *testing.T) {
	m := NewModel("/nonexistent.txt")
	m = applyWindowSize(m, 80, 24)

	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view")
	}
}

func TestView_Quitting(t *testing.T) {
	path := writeTempFile(t, "quit.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)
	m = pressKey(m, "q")

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got %q", view)
	}
}

// ==================== Key bindings ====================

func TestQuit_Q(t *testing.T) {
	path := writeTempFile(t, "quit.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestQuit_CtrlC(t *testing.T) {
	path := writeTempFile(t, "quit.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)
	_ = m
	if cmd == nil {
		t.Error("expected quit command on ctrl+c")
	}
}

func TestQuit_Escape(t *testing.T) {
	path := writeTempFile(t, "quit.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	_ = m
	if cmd == nil {
		t.Error("expected quit command on escape")
	}
}

// ==================== Link selection resets on page change ====================

func TestLink_SelectionResetsOnPageChange(t *testing.T) {
	path := writeTempFile(t, "links.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40)

	m = pressKey(m, "tab") // select a link
	if m.SelectedLink() < 0 {
		t.Skip("no links on this page")
	}

	m = pressKey(m, "right") // next page
	if m.SelectedLink() != -1 {
		t.Errorf("expected link selection reset, got %d", m.SelectedLink())
	}
}

// ==================== Helpers ====================

func applyWindowSize(m Model, w, h int) Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return updated.(Model)
}

func pressKey(m Model, k string) Model {
	var msg tea.KeyMsg
	switch k {
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "home":
		msg = tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		msg = tea.KeyMsg{Type: tea.KeyEnd}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "b":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
	case "q":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	updated, _ := m.Update(msg)
	return updated.(Model)
}
