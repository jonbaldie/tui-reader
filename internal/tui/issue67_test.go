package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIndentedCodeIsNotStyledAsHeading(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	m := newTestModel(t, "code.md", "    # Not a heading\n\n# Real heading")
	got := m.View()
	if want := textStyle.Render("    # Not a heading"); !strings.Contains(got, want) {
		t.Fatalf("indented code style = %q, want regular text style %q", got, want)
	}

	if want := headingStyle.Render("# Real heading"); !strings.Contains(got, want) {
		t.Fatalf("heading style = %q, want heading style %q", got, want)
	}
}
