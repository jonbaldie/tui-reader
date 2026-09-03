package book

import (
	"reflect"
	"strings"
	"testing"
)

// Regression for #42: wrapped indented code-block lines must preserve leading
// 4-space indentation on the first display line and all continuation lines.
func TestIndentedCodeBlock_WrappingPreservesIndentation(t *testing.T) {
	raw := []string{"    this is a long indented code line that wraps past width"}

	// At width 20, lines should wrap and every line must start with "    ".
	formatted := FormatParagraphs(raw, 20)
	if len(formatted) < 2 {
		t.Fatalf("expected multiple wrapped lines, got %v", formatted)
	}

	for i, line := range formatted {
		if !strings.HasPrefix(line, "    ") {
			t.Errorf("line %d does not have leading 4-space indentation: %q", i, line)
		}
		if sw := stringWidth(line); sw > 20 {
			t.Errorf("line %d exceeds width 20: %d columns: %q", i, sw, line)
		}
	}
}

func TestIndentedCodeBlock_ShortLinePreservesIndentation(t *testing.T) {
	raw := []string{"    short code"}
	formatted := FormatParagraphs(raw, 20)
	want := []string{"    short code"}
	if !reflect.DeepEqual(formatted, want) {
		t.Errorf("got %v, want %v", formatted, want)
	}
}

func TestIndentedCodeBlock_NotGivenParagraphIndent(t *testing.T) {
	// A code block following a paragraph must NOT receive a 2-space paragraph indent.
	raw := []string{"First paragraph", "", "    code line"}
	formatted := FormatParagraphs(raw, 40)
	if len(formatted) != 3 {
		t.Fatalf("expected 3 lines, got %v", formatted)
	}
	if formatted[2] != "    code line" {
		t.Errorf("expected %q, got %q", "    code line", formatted[2])
	}
}

func TestIndentedCodeBlock_SpecialLinesNoParagraphIndent(t *testing.T) {
	lines := []string{"para 1", "", "# Heading", "", "---", "", "para 2"}
	formatted := FormatParagraphs(lines, 40)
	if len(formatted) != 7 {
		t.Fatalf("expected 7 lines, got %v", formatted)
	}
	if formatted[2] != "# Heading" {
		t.Errorf("heading should not be indented, got %q", formatted[2])
	}
	if formatted[4] != "---" {
		t.Errorf("horizontal rule should not be indented, got %q", formatted[4])
	}
	if formatted[6] != "  para 2" {
		t.Errorf("normal paragraph 2 should be indented, got %q", formatted[6])
	}
}

func TestIndentedCodeBlock_BoundaryWidth5(t *testing.T) {
	// At exact boundary width 5, width - 4 = 1. "ab" wraps to "a", "b", each prefixed by 4 spaces.
	formatted := FormatParagraphs([]string{"    ab"}, 5)
	want := []string{"    a", "    b"}
	if !reflect.DeepEqual(formatted, want) {
		t.Errorf("at width 5: got %v, want %v", formatted, want)
	}

	// At width 4, code block wrapping does not apply 4-space indent (width < 5).
	// Content must not be dropped (non-empty output), and provenance must be preserved.
	layout := buildBookLayout([]string{"    ab"}, 4, 10)
	if len(layout.formatted) == 0 {
		t.Fatal("at width 4, formatted lines must not be empty")
	}
	for _, fl := range layout.formatted {
		if fl.raw != 0 {
			t.Errorf("expected provenance raw=0, got %d", fl.raw)
		}
		if stringWidth(fl.text) > 4 {
			t.Errorf("at width 4, line %q exceeds width 4", fl.text)
		}
	}
	if len(layout.pages[0].Lines) == 0 {
		t.Fatal("at width 4, pages must not be empty")
	}
}

func TestIndentedCodeBlock_ExactWidthMinus4(t *testing.T) {
	// At width 6, available width for code is 6 - 4 = 2.
	// "abc" must wrap into "ab" and "c", resulting in "    ab" and "    c".
	formatted := FormatParagraphs([]string{"    abc"}, 6)
	want := []string{"    ab", "    c"}
	if !reflect.DeepEqual(formatted, want) {
		t.Errorf("at width 6: got %v, want %v", formatted, want)
	}
}
