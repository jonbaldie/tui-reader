package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue188_HardWrappedSelection(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	for _, tt := range []struct {
		name      string
		row       string
		fragments []string
	}{
		{"split markup", "|" + strings.Repeat(" ", 59) + "[Wide target](#destination) |", []string{"Wide target"}},
		{"split label", "|" + strings.Repeat(" ", 64) + "[Wide target](#destination) |", []string{"Wide t", "arget"}},
		{"over-wide prose", "[" + strings.Repeat("W", 90) + "](#destination)", []string{strings.Repeat("W", 71), strings.Repeat("W", 19)}},
		{"last line on page", strings.Repeat("| filler\n", 16) + "|" + strings.Repeat(" ", 64) + "[Wide target](#destination) |", []string{"Wide t"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.row + "\n\n" + strings.Repeat("Filler.\n\n", 30) + "# Destination\nTarget content.\n"
			m := applyWindowSize(newTestModel(t, "issue188.md", doc), 80, 24)
			plain := stripAnsi(m.View())
			m = pressKey(m, "tab")
			view := m.View()
			for _, fragment := range tt.fragments {
				if !strings.Contains(view, selectedLinkStyle.Render(fragment)) {
					t.Errorf("selected label fragment %q has no highlight", fragment)
				}
			}
			if got := strings.Count(view, "48;5;117"); got != len(tt.fragments) {
				t.Errorf("selected backgrounds = %d, want %d", got, len(tt.fragments))
			}
			if got := stripAnsi(view); got != plain {
				t.Error("Tab changed the displayed text")
			}
			m = pressKey(m, "enter")
			if got, want := m.CurrentPage(), m.BookRef().PageForAnchor("destination"); got != want || got == 0 {
				t.Fatalf("Enter landed on page %d, want destination page %d beyond page 0", got, want)
			}
			if !strings.Contains(m.View(), "Target content.") {
				t.Fatal("destination content is not visible")
			}
			m = pressKey(m, "b")
			if m.CurrentPage() != 0 || m.SelectedLink() != -1 {
				t.Fatal("back did not restore the first page with selection cleared")
			}
		})
	}
}
