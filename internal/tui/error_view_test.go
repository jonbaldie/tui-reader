package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func maxLineDisplayWidth(s string) int {
	maxW := 0
	for _, line := range strings.Split(s, "\n") {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}
	return maxW
}

func visibleWords(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func visibleCompact(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func TestView_ErrorWrapsAtWidth40(t *testing.T) {
	m := NewModel("missing.md")
	m = applyWindowSize(m, 40, 12)

	view := m.View()
	if w := maxLineDisplayWidth(view); w > 40 {
		t.Errorf("error view line is %d cells wide at termWidth 40", w)
	}
	got := visibleWords(view)
	if !strings.Contains(got, "missing.md") {
		t.Error("expected filename missing.md in error view")
	}
	if !strings.Contains(got, "no such file or directory") {
		t.Error("expected OS reason in error view")
	}
}

func TestView_ErrorUnclippedAtWidth80(t *testing.T) {
	m := NewModel("missing.md")
	m = applyWindowSize(m, 80, 24)

	view := m.View()
	if !strings.Contains(view, "missing.md") {
		t.Error("expected filename missing.md in error view")
	}
	if !strings.Contains(view, "no such file or directory") {
		t.Error("expected OS reason in error view")
	}
	if w := maxLineDisplayWidth(view); w > 80 {
		t.Errorf("error view line is %d cells wide at termWidth 80", w)
	}
}

func TestView_ErrorTinyTerminals(t *testing.T) {
	m := NewModel("missing.md")

	view0 := m.View()
	if !strings.Contains(visibleCompact(view0), "Error") {
		t.Error("width 0 should still emit the error message")
	}

	m = applyWindowSize(m, 5, 5)
	view5 := m.View()
	if !strings.Contains(visibleCompact(view5), "Error") {
		t.Error("width 5 should still emit the error message")
	}
}

func TestQuit_ErrorScreen(t *testing.T) {
	m := NewModel("missing.md")
	m = applyWindowSize(m, 40, 12)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)
	if cmd == nil {
		t.Error("expected quit command from error screen")
	}
	if m.View() != "" {
		t.Errorf("expected empty view after quit from error screen, got %q", m.View())
	}
}
