package book

import (
	"fmt"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestIssue101_WideRuneDoesNotGainParagraphIndentOverflow(t *testing.T) {
	raw := []string{"# H", "", "第一段落　次の段落", "続き"}
	const width = 3

	pages := Paginate(raw, width, 20)
	for pageIndex, page := range pages {
		for lineIndex, line := range page.Lines {
			if got := runewidth.StringWidth(line); got > width {
				t.Errorf("page %d line %d has display width %d > %d: %q", pageIndex, lineIndex, got, width, line)
			}
		}
	}
}

func TestIssue101_IndentedParagraphsRespectDisplayWidth(t *testing.T) {
	cases := []struct {
		name string
		raw  []string
	}{
		{"ascii", []string{"# H", "", "a paragraph with words"}},
		{"cjk-single-source-line", []string{"# H", "", "第一段落　次の段落"}},
		{"cjk-multiple-source-lines", []string{"# H", "", "第一段落", "次の段落"}},
		{"mixed", []string{"# H", "", "第一 a paragraph"}},
	}

	for _, tc := range cases {
		for width := 1; width <= 5; width++ {
			t.Run(fmt.Sprintf("%s/width=%d", tc.name, width), func(t *testing.T) {
				for pageIndex, page := range Paginate(tc.raw, width, 20) {
					for lineIndex, line := range page.Lines {
						got := runewidth.StringWidth(line)
						if got > width && !isUnavoidableWideRune(line, width) {
							t.Errorf("page %d line %d has display width %d > %d: %q", pageIndex, lineIndex, got, width, line)
						}
					}
				}
			})
		}
	}
}

func isUnavoidableWideRune(line string, width int) bool {
	runes := []rune(line)
	return width < 2 && len(runes) == 1 && runewidth.RuneWidth(runes[0]) > width
}
