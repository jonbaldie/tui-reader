package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIssue167_TabSelectsRealLinkNotTabIndentedCode(t *testing.T) {
	m := newTestModel(t, "tab.md", "# Intro\n\t# Not a heading\n\t[Not a link](#intro)\n[Real link](#intro)\n")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	links := currentLinks(m)
	if len(links) != 1 {
		t.Fatalf("current links = %+v, want only the real link", links)
	}
	if got := links[m.selectedLink]; got.Label != "Real link" || got.Target != "intro" {
		t.Fatalf("first Tab selected %+v, want Real link to intro", got)
	}
}
