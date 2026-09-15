package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue100_TabHighlightsLinkAtSourceLineBoundary(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := "000[\n[0](#00)\n\n" + strings.Repeat("filler\n\n", 4) + "# 00\nTarget body.\n"
	m := NewModel(writeTempFile(t, "issue100.md", doc))
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}
	m = applyWindowSize(m, 14, 12) // content width 10, height 5

	if len(m.BookRef().Pages[0].Links) != 1 {
		t.Fatalf("page 0 links = %+v, want one link", m.BookRef().Pages[0].Links)
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("selected link = %d, want 0", m.SelectedLink())
	}

	selectedMarkup := "[" + selectedLinkStyle.Render("0") + "](#00)"
	if !strings.Contains(m.View(), selectedMarkup) {
		t.Fatalf("selected link is not visibly highlighted:\n%s", m.View())
	}

	m = pressKey(m, "enter")
	targetPage := m.BookRef().PageForAnchor("00")
	if m.CurrentPage() != targetPage {
		t.Fatalf("Enter landed on page %d, want target page %d", m.CurrentPage(), targetPage)
	}
}

func TestIssue100_TabHighlightsUnwrappedLinkAtSourceLineBoundary(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	m := NewModel(writeTempFile(t, "issue100-wide.md", "000[\n[0](#00)\n"))
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	if len(m.BookRef().Pages[0].Links) != 1 {
		t.Fatalf("page 0 links = %+v, want one link", m.BookRef().Pages[0].Links)
	}
	m = pressKey(m, "tab")
	selectedMarkup := "[" + selectedLinkStyle.Render("0") + "](#00)"
	if !strings.Contains(m.View(), selectedMarkup) {
		t.Fatalf("selected link is not visibly highlighted in unwrapped prose:\n%s", m.View())
	}
}
