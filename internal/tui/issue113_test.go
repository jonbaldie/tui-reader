package tui

import (
	"strings"
	"testing"
)

func TestIssue113_BOMFirstHeadingNavigationAndStyle(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("\xef\xbb\xbf# Start\n\n")
	for i := 0; i < 8; i++ {
		sb.WriteString("A short paragraph of filler text.\n\n")
	}
	sb.WriteString("[Top](#start)\n")

	path := writeTempFile(t, "min-bom.md", sb.String())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 14)

	// In 60x14 terminal, [Top](#start) is on page 2 (0-indexed page 2, i.e. page 3 of 3)
	m = pressKey(m, "end")
	if m.CurrentPage() == 0 {
		t.Fatalf("expected to navigate away from page 0 on End, got page %d", m.CurrentPage())
	}

	// Press Tab to select [Top](#start) and Enter to follow it
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")

	if m.CurrentPage() != 0 {
		t.Errorf("after pressing Enter on [Top](#start), current page = %d, want 0", m.CurrentPage())
	}
}
