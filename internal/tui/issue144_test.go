package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue144_HeadingWithLinkKeepsHeadingAndLinkStyles(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	m := newTestModel(t, "links.md", "## Read [more](#target)\n\n# Target")
	view := m.View()
	if !strings.Contains(view, "\x1b[1;38;5;117m## Read [") || !strings.Contains(view, "38;5;75") || !strings.Contains(stripAnsi(view), "## Read [more](#target)") {
		t.Fatalf("heading must retain heading styling around its link: %q", view)
	}
	m = pressKey(m, "tab")
	if view = m.View(); !strings.Contains(view, "48;5;117") {
		t.Fatalf("heading link must still show selection: %q", view)
	}
}

func TestIssue144_ViewUsesLayoutHeadingClassification(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	cases := []struct {
		name, doc, line string
		heading         bool
	}{
		{"server.log", "# not a heading in a log", "# not a heading in a log", false},
		{"prose.md", "The fix landed in pull #142 after review.", "#142 after review.", false},
		{"heading.md", "## A heading long enough that it wraps", "that it wraps", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := applyWindowSize(newTestModel(t, tc.name, tc.doc), 28, 20)
			view := m.View()
			want, unwanted := textStyle.Render(tc.line), headingStyle.Render(tc.line)
			if tc.heading {
				want, unwanted = unwanted, want
			}
			if want == unwanted {
				t.Fatal("test requires distinct heading and body styles")
			}
			if !strings.Contains(view, want) || strings.Contains(view, unwanted) {
				t.Fatalf("View must render %q with heading=%v; got %q", tc.line, tc.heading, view)
			}
		})
	}
}
