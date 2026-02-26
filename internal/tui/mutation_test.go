package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Mutation tests for the TUI model. Each verifies that specific logic
// changes would be caught by tests.

// ==================== Navigation boundary mutations ====================

// Mutation: removing the `currentPage > 0` guard would allow negative pages.
func TestMutation_PrevPageNeverNegative(t *testing.T) {
	path := writeTempFile(t, "mut.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	// Press prev many times
	for i := 0; i < 10; i++ {
		m = pressKey(m, "left")
	}
	if m.CurrentPage() < 0 {
		t.Error("page went negative")
	}
}

// Mutation: removing the `currentPage < len(Pages)-1` guard would allow overflow.
func TestMutation_NextPageNeverExceedsMax(t *testing.T) {
	path := writeTempFile(t, "mut.md", "content\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 15)

	// Press next many times
	for i := 0; i < 10; i++ {
		m = pressKey(m, "right")
	}
	max := len(m.BookRef().Pages) - 1
	if m.CurrentPage() > max {
		t.Errorf("page exceeded max: %d > %d", m.CurrentPage(), max)
	}
}

// ==================== Link selection mutations ====================

// Mutation: not resetting selectedLink on page change.
func TestMutation_LinkResetOnNext(t *testing.T) {
	path := writeTempFile(t, "mut.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40)

	m = pressKey(m, "tab")
	if m.SelectedLink() < 0 {
		t.Skip("no links on page")
	}
	m = pressKey(m, "right")
	if m.SelectedLink() != -1 {
		t.Errorf("expected selectedLink -1 after page change, got %d", m.SelectedLink())
	}
}

func TestMutation_LinkResetOnPrev(t *testing.T) {
	path := writeTempFile(t, "mut.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40)

	m = pressKey(m, "right") // go to page 1
	m = pressKey(m, "tab")   // select a link (if any)
	m = pressKey(m, "left")  // back to page 0
	if m.SelectedLink() != -1 {
		t.Errorf("expected selectedLink -1 after prev page, got %d", m.SelectedLink())
	}
}

// Mutation: tab on page with no links should not crash or change selection.
func TestMutation_TabOnNoLinksPage(t *testing.T) {
	path := writeTempFile(t, "mut.md", "No links here\n")
	m := NewModel(path)
	m = applyWindowSize(m, 60, 40)

	m = pressKey(m, "tab")
	if m.SelectedLink() != -1 {
		t.Errorf("expected -1 on page with no links, got %d", m.SelectedLink())
	}
}

// ==================== History mutations ====================

// Mutation: not pushing to history before navigation.
func TestMutation_HistoryPushedOnFollow(t *testing.T) {
	path := writeTempFile(t, "mut.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	m = pressKey(m, "tab")
	startPage := m.CurrentPage()
	m = pressKey(m, "enter")

	if len(m.History()) == 0 {
		t.Error("expected history entry after following link")
	}
	if m.History()[len(m.History())-1] != startPage {
		t.Errorf("expected history to contain %d, got %v", startPage, m.History())
	}
}

// Mutation: not popping from history on back.
func TestMutation_HistoryPoppedOnBack(t *testing.T) {
	path := writeTempFile(t, "mut.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	histLen := len(m.History())

	m = pressKey(m, "b")
	if len(m.History()) != histLen-1 {
		t.Errorf("expected history length %d after back, got %d", histLen-1, len(m.History()))
	}
}

// Mutation: going back restores wrong page.
func TestMutation_BackRestoresCorrectPage(t *testing.T) {
	path := writeTempFile(t, "mut.md", simpleDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 10)

	origPage := m.CurrentPage()
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	m = pressKey(m, "b")
	if m.CurrentPage() != origPage {
		t.Errorf("expected to return to page %d, got %d", origPage, m.CurrentPage())
	}
}

// ==================== View rendering mutations ====================

// Mutation: header not showing title.
func TestMutation_ViewContainsTitle(t *testing.T) {
	path := writeTempFile(t, "my-doc.md", "# Heading\ntext\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)
	view := m.View()
	if !strings.Contains(view, "My Doc") {
		t.Errorf("view should contain title 'My Doc'")
	}
}

// Mutation: footer not showing page count.
func TestMutation_ViewContainsPageCount(t *testing.T) {
	path := writeTempFile(t, "mut.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)
	view := m.View()
	if !strings.Contains(view, "Page 1") {
		t.Error("view should contain page number")
	}
}

// Mutation: view still renders after quitting.
func TestMutation_QuitClearsView(t *testing.T) {
	path := writeTempFile(t, "mut.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)
	m = pressKey(m, "q")
	if m.View() != "" {
		t.Error("view should be empty after quit")
	}
}

// ==================== Resize mutations ====================

// Mutation: contentWidth not capped at maxWidth.
func TestMutation_ContentWidthCapped(t *testing.T) {
	path := writeTempFile(t, "mut.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 200, 50) // very wide terminal
	if m.contentWidth > 72 {
		t.Errorf("contentWidth should be capped at 72, got %d", m.contentWidth)
	}
}

// Mutation: contentHeight calculation wrong.
func TestMutation_ContentHeightMinimum(t *testing.T) {
	path := writeTempFile(t, "mut.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 8) // very short terminal
	if m.contentHeight < 5 {
		t.Errorf("contentHeight should be at least 5, got %d", m.contentHeight)
	}
}

// ==================== Quit command mutations ====================

// Mutation: quit key produces no command (wouldn't actually quit).
func TestMutation_QuitProducesCommand(t *testing.T) {
	path := writeTempFile(t, "mut.md", "text\n")
	m := NewModel(path)
	m = applyWindowSize(m, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("quit should produce a non-nil command")
	}
}

// ==================== Error state mutations ====================

// Mutation: error state not shown properly.
func TestMutation_ErrorViewShowsMessage(t *testing.T) {
	m := NewModel("/absolutely/nonexistent/path.txt")
	m = applyWindowSize(m, 80, 24)
	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Error("error view should contain 'Error'")
	}
	if !strings.Contains(view, "cannot open file") {
		t.Error("error view should contain the actual error message")
	}
}

// Mutation: BookRef returns non-nil for failed loads.
func TestMutation_FailedLoadNilBook(t *testing.T) {
	m := NewModel("/absolutely/nonexistent/path.txt")
	if m.BookRef() != nil {
		t.Error("book should be nil after failed load")
	}
}
