package book

import (
	"fmt"
	"testing"

	"github.com/mattn/go-runewidth"
)

// TestIssue105_IndentedCodeBlockRespectsDisplayWidth is the regression test
// for #105: an indented code block containing wide runes must not gain the
// 4-space indent when doing so would push any wrapped display line over the
// requested page width.
func TestIssue105_IndentedCodeBlockRespectsDisplayWidth(t *testing.T) {
	cases := []struct {
		name string
		raw  []string
	}{
		{"cjk-single-source-line", []string{"    第一段落"}},
		{"cjk-multiple-source-lines", []string{"    第一段落", "    次の段落"}},
		{"ascii", []string{"    a code line"}},
	}

	for _, tc := range cases {
		for width := 1; width <= 6; width++ {
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
