package tui

import (
	"strings"
	"testing"
)

// Regression for #41: CJK characters must not cause pages to overflow the terminal.
func TestCJKDisplayWidth_TUIViewDoesNotOverflowTerminal(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 23; i++ {
		sb.WriteString(strings.Repeat("東", 20))
		sb.WriteString("\n")
	}
	path := writeTempFile(t, "cjk.md", sb.String())
	m := NewModel(path)
	m = applyWindowSize(m, 40, 30) // contentWidth=36, contentHeight=23

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 30 {
		t.Errorf("View() returned %d lines for 30-line terminal: overflows terminal", len(lines))
	}
}
