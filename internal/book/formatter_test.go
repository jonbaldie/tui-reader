package book

import (
	"strings"
	"testing"
)

// projectText extracts just the display text from the formatter output, which
// is exactly what FormatParagraphs must return.
func projectText(lines []formattedLine) []string {
	out := make([]string, len(lines))
	for i, fl := range lines {
		out[i] = fl.text
	}
	return out
}

// formatterInputs is a representative set of documents exercising every
// formatting rule the one-pass formatter owns: headings, code blocks, blank
// collapsing, preserved blank source lines, indentation, and wrapping.
var formatterInputs = []struct {
	name  string
	raw   []string
	width int
}{
	{"heading-then-two-paras", []string{"# Heading", "", "First para.", "Second para."}, 60},
	{"adjacent-paras", []string{"Alpha line one.", "Beta line two."}, 60},
	{"preserved-blanks", []string{"Intro line.", "", "", "Body line."}, 60},
	{"wrapping-non-first", []string{"Short intro.", strings.Repeat("word ", 40), "Tail."}, 40},
	{"tiny-width-clamp", []string{"Intro.", strings.Repeat("ab ", 20)}, 11},
	{"headings-and-code", []string{"# Title", "", "Normal paragraph here.", "", "## Section", "    indented code block", "Another paragraph."}, 50},
	{"empty", []string{}, 60},
	{"whitespace-only", []string{"   ", "\t"}, 60},
}

// TestFormatter_TextMatchesFormatParagraphs is the tracer: the formatter's text
// projection must be byte-for-byte identical to FormatParagraphs today.
func TestFormatter_TextMatchesFormatParagraphs(t *testing.T) {
	for _, tc := range formatterInputs {
		t.Run(tc.name, func(t *testing.T) {
			want := FormatParagraphs(tc.raw, tc.width)
			got := projectText(formatParagraphsWithProvenance(tc.raw, tc.width))
			if len(got) != len(want) {
				t.Fatalf("len = %d, want %d\ngot  %q\nwant %q", len(got), len(want), got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
				}
			}
		})
	}
}

// checkFormatterInvariants verifies the four provenance invariants directly on
// the one-pass formatter output. These are sensitive to off-by-one, sign, and
// wrap-width mutations in the provenance assignment.
func checkFormatterInvariants(t *testing.T, rawLines []string, width int) []formattedLine {
	t.Helper()
	lines := formatParagraphsWithProvenance(rawLines, width)

	// INV1: exactly one provenance entry per formatted line (the projection has
	// the same length as the formatter output).
	if len(lines) != len(FormatParagraphs(rawLines, width)) {
		t.Fatalf("formatter produced %d lines, want one per formatted line", len(lines))
	}

	prevRaw := -1
	for fi, fl := range lines {
		if fl.raw >= 0 {
			// INV2: present raw indices are non-decreasing and in range.
			if fl.raw < prevRaw {
				t.Errorf("fi=%d raw %d precedes previous %d (must be non-decreasing)", fi, fl.raw, prevRaw)
			}
			if fl.raw >= len(rawLines) {
				t.Errorf("fi=%d raw %d out of range (len %d)", fi, fl.raw, len(rawLines))
			}
			prevRaw = fl.raw
		}

		if fl.text != "" {
			// INV3: a non-blank formatted line maps to a non-blank raw line.
			if fl.raw < 0 {
				t.Errorf("non-blank line %d (%q) maps to -1; want a raw line", fi, fl.text)
			} else if strings.TrimSpace(rawLines[fl.raw]) == "" {
				t.Errorf("non-blank line %d (%q) maps to blank raw line %d", fi, fl.text, fl.raw)
			}
		} else {
			// INV4: a blank formatted line is an inserted spacer (-1) or a
			// preserved blank source line.
			if fl.raw >= 0 && strings.TrimSpace(rawLines[fl.raw]) != "" {
				t.Errorf("blank line %d maps to non-blank raw line %d (%q)", fi, fl.raw, rawLines[fl.raw])
			}
		}
	}
	return lines
}

func TestFormatter_Invariants(t *testing.T) {
	for _, tc := range formatterInputs {
		t.Run(tc.name, func(t *testing.T) {
			checkFormatterInvariants(t, tc.raw, tc.width)
		})
	}
}

// TestFormatter_ExactProvenance pins the precise raw index of each formatted
// line: a heading (0), a preserved blank source line (1), an unindented first
// paragraph (2), an inserted inter-paragraph spacer (-1), and a second
// paragraph (3). Mutating any single provenance assignment breaks this.
func TestFormatter_ExactProvenance(t *testing.T) {
	raw := []string{"# Heading", "", "First para.", "Second para."}
	lines := checkFormatterInvariants(t, raw, 60)
	want := []int{0, 1, 2, -1, 3}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %+v", len(lines), len(want), lines)
	}
	for i, w := range want {
		if lines[i].raw != w {
			t.Errorf("line %d (%q) raw = %d, want %d", i, lines[i].text, lines[i].raw, w)
		}
	}
}

// TestFormatter_SpacerIsMinusOne pins that the spacer inserted between two
// adjacent paragraphs carries -1, not the index of either neighbour.
func TestFormatter_SpacerIsMinusOne(t *testing.T) {
	lines := checkFormatterInvariants(t, []string{"Alpha.", "Beta."}, 60)
	want := []int{0, -1, 1}
	for i, w := range want {
		if lines[i].raw != w {
			t.Errorf("line %d raw = %d, want %d", i, lines[i].raw, w)
		}
	}
}
