package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/muesli/termenv"
)

func TestStyleLine_PreservesRepeatedLabelsAndMarkup(t *testing.T) {
	line := "[repeat](#first) then [repeat](#second)"
	got := styledBookLine(t, line, 1)
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
}

func TestStyleLine_LeavesUnattachedMarkupPlain(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })
	line := "[visible](#attached) and `[plain](#unattached)`"
	got := styledBookLine(t, line, -1)
	if plain := stripAnsi(got); plain != line {
		t.Errorf("styled line = %q, want original markup %q after stripping ANSI", plain, line)
	}
	if strings.Contains(got, unselectedLinkStyle.Render("plain")) {
		t.Errorf("inline code link was styled: %q", got)
	}
}

func TestStyleLine_SelectedLinkUsesDistinctStyle(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	selected := styledBookLine(t, "[link](#target)", 0)
	unselected := styledBookLine(t, "[link](#target)", -1)
	if selected == unselected {
		t.Fatal("selected and unselected links must have distinct ANSI styling")
	}
}

func TestStyleLine_ExactFormatting(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	line := "prefix [one](#target1) middle [two](#target2) suffix"
	got := styledBookLine(t, line, 1)
	want := textStyle.Render("prefix [" + unselectedLinkStyle.Render("one") + "](#target1) middle [" + selectedLinkStyle.Render("two") + "](#target2) suffix")
	if got != want {
		t.Fatalf("styleLine mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestStyleLine_DuplicateTargetsSelectOneInstance(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	line := "[First Link](#target) and [Second Link](#target)"
	got0 := styledBookLine(t, line, 0)
	got1 := styledBookLine(t, line, 1)
	if got0 == got1 {
		t.Fatal("selecting first vs second same-target link produced identical markup")
	}
	if !strings.Contains(got0, selectedLinkStyle.Render("First Link")) || !strings.Contains(got0, unselectedLinkStyle.Render("Second Link")) || strings.Contains(got0, selectedLinkStyle.Render("Second Link")) {
		t.Fatalf("index 0 should select only First Link, got %q", got0)
	}
	if !strings.Contains(got1, selectedLinkStyle.Render("Second Link")) || !strings.Contains(got1, unselectedLinkStyle.Render("First Link")) || strings.Contains(got1, selectedLinkStyle.Render("First Link")) {
		t.Fatalf("index 1 should select only Second Link, got %q", got1)
	}
}

func styledBookLine(t *testing.T, line string, selected int) string {
	t.Helper()
	b, err := book.Read(strings.NewReader(line), "test", 120, 20, false)
	if err != nil {
		t.Fatal(err)
	}
	page := b.Page(0)
	return styleLine(page.Lines[0], page.Headings[0], 0, page.LabelSpans, selected)
}
