package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue188_OpenerOnlyOwnerPage(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	row := "|" + strings.Repeat(" ", 70) + "[Wide target](#destination) |"
	doc := strings.Repeat("| filler\n", 16) + row + "\n\n" + strings.Repeat("Filler.\n\n", 30) + "# Destination\nTarget content.\n"
	m := applyWindowSize(newTestModel(t, "opener.md", doc), 80, 24)
	if got := len(m.BookRef().Page(0).Links); got != 1 {
		t.Fatalf("owner links = %d, want one", got)
	}
	m = pressKey(m, "tab")
	if !strings.Contains(m.View(), selectedLinkStyle.Render("[")) {
		t.Fatal("Tab cannot visibly mark the opener-only owner page")
	}
	m = pressKey(m, "right")
	if got := len(m.BookRef().Page(1).Links); got != 0 {
		t.Fatalf("continuation page gained %d selectable links", got)
	}
	if m.SelectedLink() != -1 || strings.Contains(m.View(), selectedLinkStyle.Render("Wide target")) {
		t.Fatal("page turn retained the selection")
	}
	m = pressKey(m, "left")
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	if got, want := m.CurrentPage(), m.BookRef().PageForAnchor("destination"); got != want || got == 0 {
		t.Fatalf("Enter landed on page %d, want %d", got, want)
	}
}

func TestIssue188_DuplicateOccurrencesAndUnicodeResize(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := "`[same](#target)` [same](#target) [same](#target) [界界界界界](#target)\n\n# Target\n"
	m := applyWindowSize(newTestModel(t, "duplicates.md", doc), 80, 24)
	plain := stripAnsi(m.View())
	for _, selected := range []int{0, 1} {
		m = pressKey(m, "tab")
		view := m.View()
		if m.SelectedLink() != selected || strings.Count(view, selectedLinkStyle.Render("same")) != 1 || !strings.Contains(view, "`[same](#target)`") {
			t.Fatalf("selection %d styled the wrong occurrence: %q", selected, view)
		}
		if stripAnsi(view) != plain {
			t.Fatal("selection changed the displayed text")
		}
	}
	m = applyWindowSize(m, 13, 24)
	if m.SelectedLink() != -1 {
		t.Fatal("resize did not clear selection")
	}
	m = pressKey(m, "tab")
	m = pressKey(m, "tab")
	m = pressKey(m, "tab")
	if m.SelectedLink() != 2 {
		t.Fatalf("selected link = %d, want Unicode link", m.SelectedLink())
	}
	if !strings.Contains(m.View(), selectedLinkStyle.Render("界界界界")) || !strings.Contains(m.View(), selectedLinkStyle.Render("界")) {
		t.Fatalf("Unicode link fragments are not both highlighted: %q", m.View())
	}
}

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
			wantBackgrounds := 0
			for _, fragment := range tt.fragments {
				wantBackgrounds += utf8.RuneCountInString(fragment)
			}
			if got := strings.Count(view, "48;5;117"); got != wantBackgrounds {
				t.Errorf("selected background runes = %d, want %d", got, wantBackgrounds)
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
