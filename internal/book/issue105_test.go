package book

import "testing"

// Regression for #105: at width 5, formatCodeBlock reserves 4 columns for the
// indent and wraps the code at width-4=1, but a wide (CJK) rune cannot fit in
// 1 column. WrapLines still emits it whole, so the prefixed display line ends
// up 6 columns wide, exceeding the requested page width of 5.
func TestIndentedCodeBlock_WideRuneDoesNotOverflowAtWidth5(t *testing.T) {
	formatted := FormatParagraphs([]string{"    第一段落"}, 5)
	for i, line := range formatted {
		if sw := stringWidth(line); sw > 5 {
			t.Errorf("line %d exceeds width 5: %d columns: %q", i, sw, line)
		}
	}
}
