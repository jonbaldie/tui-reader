package book

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// Regression for #40: paragraph indent must not cause display lines to exceed
// page width for narrow widths (width <= 11).
func TestParagraphOverflow_NarrowWidthDoesNotOverflowPage(t *testing.T) {
	raw := []string{"# H", "", "body text here that wraps into multiple lines"}

	for width := 1; width <= 15; width++ {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			pages := readBook(t, strings.Join(raw, "\n"), true, width, 20).pages
			for pi, page := range pages {
				for li, line := range page.Lines {
					if rl := utf8.RuneCountInString(line); rl > width {
						t.Errorf("width %d: page %d line %d: %d runes > %d: %q", width, pi, li, rl, width, line)
					}
				}
			}

			formatted := readBook(t, strings.Join(raw, "\n"), true, width, 1<<30).Page(0).Lines
			for li, line := range formatted {
				if rl := utf8.RuneCountInString(line); rl > width {
					t.Errorf("width %d: formatted line %d: %d runes > %d: %q", width, li, rl, width, line)
				}
			}
		})
	}
}

func TestParagraphOverflow_IndentPreservedWhenWidthPermits(t *testing.T) {
	raw := []string{"First paragraph", "", "Second paragraph"}

	// When width >= 3, the second paragraph should have a 2-space indent on its first line.
	formatted := readBook(t, strings.Join(raw, "\n"), true, 20, 1<<30).Page(0).Lines
	var foundSecond bool
	for _, line := range formatted {
		if strings.Contains(line, "Second") {
			foundSecond = true
			if !strings.HasPrefix(line, "  ") {
				t.Errorf("expected 2-space indent on second paragraph line, got %q", line)
			}
		}
	}
	if !foundSecond {
		t.Error("did not find second paragraph in formatted lines")
	}

	// At the boundary width of 3, the indent should still be applied ("  " + "b" = 3 runes).
	formatted3 := readBook(t, strings.Join([]string{"a", "", "b"}, "\n"), true, 3, 1<<30).Page(0).Lines
	if len(formatted3) < 3 || formatted3[2] != "  b" {
		t.Errorf("expected boundary width 3 to indent as %q, got %v", "  b", formatted3)
	}

	// When width < 3, indent should be omitted because 2 spaces would leave no room for text.
	formattedNarrow := readBook(t, strings.Join(raw, "\n"), true, 2, 1<<30).Page(0).Lines
	for _, line := range formattedNarrow {
		if strings.HasPrefix(line, "  ") {
			t.Errorf("width 2 should not have 2-space indent, got %q", line)
		}
	}
}
