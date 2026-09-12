package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue90_InlineCodeMatchingLinkDoesNotStealSelection(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := "`[demo](#demo)`[demo](#demo)"
	m := NewModel(writeTempFile(t, "issue90.md", doc))
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	if got := len(m.book.Pages[0].Links); got != 1 {
		t.Fatalf("page 0 has %d links, want 1: %+v", got, m.book.Pages[0].Links)
	}

	m = pressKey(m, "tab")
	if m.SelectedLink() != 0 {
		t.Fatalf("Tab selected link %d, want 0", m.SelectedLink())
	}

	view := m.View()
	selectedMarkup := "[" + selectedLinkStyle.Render("demo") + "](#demo)"
	if !strings.Contains(view, "](#demo)`"+selectedMarkup) {
		t.Fatalf("active link should be selected in view:\n%s", view)
	}
	if strings.Contains(view, "`"+selectedMarkup+"`") {
		t.Fatalf("inline-code occurrence should not receive selected-link styling:\n%s", view)
	}
	if !strings.Contains(view, "`[demo](#demo)`") {
		t.Fatalf("inline-code occurrence should remain literal:\n%s", view)
	}
}
