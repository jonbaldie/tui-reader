package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue106_TabSelectsSoftWrappedSpanningLink(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := "# Section\nHere is a [link\nlabel](#target) across lines.\n\n# Target\nTarget content.\n"
	m := NewModel(writeTempFile(t, "issue106.md", doc))
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}
	m = applyWindowSize(m, 80, 24)

	if len(m.BookRef().Pages[0].Links) != 1 {
		t.Fatalf("page 0 links = %d, want 1: %+v", len(m.BookRef().Pages[0].Links), m.BookRef().Pages[0].Links)
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("selected link = %d, want 0", m.SelectedLink())
	}

	selectedMarkup := "[" + selectedLinkStyle.Render("link label") + "](#target)"
	if !strings.Contains(m.View(), selectedMarkup) {
		t.Fatalf("selected link is not visibly highlighted:\n%s", m.View())
	}

	m = pressKey(m, "enter")
	targetPage := m.BookRef().PageForAnchor("target")
	if m.CurrentPage() != targetPage {
		t.Fatalf("Enter landed on page %d, want target page %d", m.CurrentPage(), targetPage)
	}
}
