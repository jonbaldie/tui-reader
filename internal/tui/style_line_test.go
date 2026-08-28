package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestStyleLinkMarkup_PreservesRepeatedLabelsAndMarkup(t *testing.T) {
	line := "[repeat](#first) then [repeat](#second)"
	got := styleLinkMarkup(line, map[linkKey]struct{}{
		{label: "repeat", target: "first"}:  {},
		{label: "repeat", target: "second"}: {},
	}, "second")
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
}

func TestStyleLinkMarkup_LeavesUnattachedMarkupPlain(t *testing.T) {
	line := "[visible](#attached) and [plain](#unattached)"
	got := styleLinkMarkup(line, map[linkKey]struct{}{
		{label: "visible", target: "attached"}: {},
	}, "")
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
}

func TestStyleLinkMarkup_SelectedLinkUsesDistinctStyle(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	links := map[linkKey]struct{}{{label: "link", target: "target"}: {}}
	selected := styleLinkMarkup("[link](#target)", links, "target")
	unselected := styleLinkMarkup("[link](#target)", links, "")
	if selected == unselected {
		t.Fatal("selected and unselected links must have distinct ANSI styling")
	}
}
