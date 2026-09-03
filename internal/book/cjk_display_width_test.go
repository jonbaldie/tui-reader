package book

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

// Regression for #41: CJK and wide characters must be wrapped by display width,
// not rune count, so display lines do not exceed the page width in visual columns.
func TestCJKDisplayWidth_WrappingRespectsDisplayWidth(t *testing.T) {
	// "東" has display width 2.
	// 20 "東" characters = 20 runes = 40 display columns.
	line := strings.Repeat("東", 20)
	wrapped := WrapLines([]string{line}, 30)

	for i, l := range wrapped {
		sw := runewidth.StringWidth(l)
		if sw > 30 {
			t.Errorf("line %d has display width %d > 30: %q", i, sw, l)
		}
	}
}

func TestCJKDisplayWidth_MixedAsciiAndCJK(t *testing.T) {
	// "hello 東京 world" -> "hello " (6) + "東京" (4) + " world" (6) = 16 display cols, 14 runes.
	line := "hello 東京 world"
	wrapped := WrapLines([]string{line}, 10)

	for i, l := range wrapped {
		sw := runewidth.StringWidth(l)
		if sw > 10 {
			t.Errorf("line %d has display width %d > 10: %q", i, sw, l)
		}
	}
}

func TestCJKDisplayWidth_SingleCharWiderThanPage(t *testing.T) {
	// A wide character (width 2) wrapped at width 1 should not produce empty lines.
	out := WrapLines([]string{"東"}, 1)
	if len(out) != 1 || out[0] != "東" {
		t.Errorf("expected [\"東\"], got %v", out)
	}
}

func TestCJKDisplayWidth_WordFillsWidthExact(t *testing.T) {
	// Word "ab" fills width 2 exactly. A following link must start on a new line.
	out := WrapLines([]string{"ab [link](#t)"}, 2)
	for i, l := range out {
		if sw := runewidth.StringWidth(l); sw > 2 {
			t.Errorf("line %d has display width %d > 2: %q", i, sw, l)
		}
	}
}

func TestCJKDisplayWidth_LinkExactWidthKeptWhole(t *testing.T) {
	// "[a](#target)" is exactly 12 characters / display columns.
	// At width 12, it must be kept whole on one display line.
	out := WrapLines([]string{"[a](#target)"}, 12)
	if len(out) != 1 || out[0] != "[a](#target)" {
		t.Errorf("expected link to be kept whole at exact width 12, got %v", out)
	}
}

func TestCJKDisplayWidth_ResetWidthOnFlushPreservesFollowers(t *testing.T) {
	// Width 8. "12345" (5) + "東" (2) = 7.
	// Next "東" (2): 7+2=9 > 8 -> flushes line 1 ("12345東", 7 cols).
	// Line 2 receives "東" (2) + "ab" (2) = 4 cols <= 8, so "東ab" must be on the same line.
	out := WrapLines([]string{"12345東東ab"}, 8)
	if len(out) < 2 || out[1] != "東ab" {
		t.Errorf("expected line 2 to be %q, got %v", "東ab", out)
	}
}
