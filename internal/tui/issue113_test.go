package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestIssue113_BOMFirstHeadingStyledAndLinkable reproduces issue #113 at the
// TUI seam: with a UTF-8 BOM at the start of the file, the first heading
// renders in body style and a link back to it does nothing.
func TestIssue113_BOMFirstHeadingStyledAndLinkable(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	var sb strings.Builder
	sb.WriteString("\ufeff# Start\n\n")
	for i := 0; i < 8; i++ {
		sb.WriteString("A short paragraph of filler text.\n\n")
	}
	sb.WriteString("[Top](#start)\n")
	path := writeTempFile(t, "min-bom.md", sb.String())

	m := NewModel(path)
	m = applyWindowSize(m, 60, 14)

	first := m.View()
	if !strings.Contains(first, headingStyle.Render("# Start")) {
		t.Errorf("page 1 does not render %q in heading style; got %q", "# Start", first)
	}

	m = pressKey(m, "end")
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	if m.CurrentPage() != 0 {
		t.Errorf("after following [Top](#start), page = %d, want 0", m.CurrentPage())
	}
}
