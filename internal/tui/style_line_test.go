package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestStyleLinkMarkup_PreservesRepeatedLabelsAndMarkup(t *testing.T) {
	line := "[repeat](#first) then [repeat](#second)"
	got, _ := styleLinkMarkup(line, map[linkKey]struct{}{
		{label: "repeat", target: "first"}:  {},
		{label: "repeat", target: "second"}: {},
	}, 1, 0)
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
}

func TestStyleLinkMarkup_LeavesUnattachedMarkupPlain(t *testing.T) {
	line := "[visible](#attached) and [plain](#unattached)"
	got, _ := styleLinkMarkup(line, map[linkKey]struct{}{
		{label: "visible", target: "attached"}: {},
	}, -1, 0)
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
}

func TestStyleLinkMarkup_SelectedLinkUsesDistinctStyle(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	links := map[linkKey]struct{}{{label: "link", target: "target"}: {}}
	selected, _ := styleLinkMarkup("[link](#target)", links, 0, 0)
	unselected, _ := styleLinkMarkup("[link](#target)", links, -1, 0)
	if selected == unselected {
		t.Fatal("selected and unselected links must have distinct ANSI styling")
	}
}

func TestStyleLinkMarkup_ExactFormatting(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	line := "prefix [one](#target1) middle [two](#target2) suffix"
	links := map[linkKey]struct{}{
		{label: "one", target: "target1"}: {},
		{label: "two", target: "target2"}: {},
	}
	got, _ := styleLinkMarkup(line, links, 1, 0)

	want := "prefix [" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Underline(true).
		Render("one") + "](#target1) middle [" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("117")).
		Bold(true).
		Underline(true).
		Render("two") + "](#target2) suffix"

	if got != want {
		t.Fatalf("styleLinkMarkup mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestStyleLinkMarkup_DuplicateTargetsSelectOneInstance(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	line := "[First Link](#target) and [Second Link](#target)"
	links := map[linkKey]struct{}{
		{label: "First Link", target: "target"}:  {},
		{label: "Second Link", target: "target"}: {},
	}

	got0, _ := styleLinkMarkup(line, links, 0, 0)
	got1, _ := styleLinkMarkup(line, links, 1, 0)
	if got0 == got1 {
		t.Fatal("selecting first vs second same-target link produced identical markup")
	}

	selectedFirst := selectedLinkStyle.Render("First Link")
	selectedSecond := selectedLinkStyle.Render("Second Link")
	unselectedFirst := unselectedLinkStyle.Render("First Link")
	unselectedSecond := unselectedLinkStyle.Render("Second Link")

	if !strings.Contains(got0, selectedFirst) || !strings.Contains(got0, unselectedSecond) || strings.Contains(got0, selectedSecond) {
		t.Fatalf("index 0 should select only First Link, got %q", got0)
	}
	if !strings.Contains(got1, selectedSecond) || !strings.Contains(got1, unselectedFirst) || strings.Contains(got1, selectedFirst) {
		t.Fatalf("index 1 should select only Second Link, got %q", got1)
	}
}
