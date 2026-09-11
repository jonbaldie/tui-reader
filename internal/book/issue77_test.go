package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue77_SoftWrapParagraph(t *testing.T) {
	raw := []string{
		"This is one Markdown paragraph deliberately split across",
		"physical source lines without a blank line.",
	}

	// At width 120, both source lines reflow into a single display line.
	out := FormatParagraphs(raw, 120)
	want := []string{
		"This is one Markdown paragraph deliberately split across physical source lines without a blank line.",
	}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("FormatParagraphs (width 120) =\n%q\nwant:\n%q", out, want)
	}

	// There must not be a blank line between lines of the same paragraph.
	for i, line := range out {
		if line == "" {
			t.Errorf("unexpected blank line at index %d: %v", i, out)
		}
	}
}

func TestIssue77_AdjacentListItemsNoBlankLines(t *testing.T) {
	tests := []struct {
		name string
		raw  []string
		want []string
	}{
		{
			name: "unordered hyphen",
			raw:  []string{"- first list item", "- second list item"},
			want: []string{"- first list item", "- second list item"},
		},
		{
			name: "unordered asterisk",
			raw:  []string{"* first list item", "* second list item"},
			want: []string{"* first list item", "* second list item"},
		},
		{
			name: "ordered numeric",
			raw:  []string{"1. first list item", "2. second list item"},
			want: []string{"1. first list item", "2. second list item"},
		},
		{
			name: "ordered parenthesis",
			raw:  []string{"1) first list item", "2) second list item"},
			want: []string{"1) first list item", "2) second list item"},
		},
		{
			name: "multi-digit ordered",
			raw:  []string{"10. tenth list item", "11. eleventh list item"},
			want: []string{"10. tenth list item", "11. eleventh list item"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatParagraphs(tt.raw, 80)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FormatParagraphs =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestIssue77_SoftWrapFixture(t *testing.T) {
	raw := []string{
		"# Soft Wrap Probe",
		"",
		"This is one Markdown paragraph deliberately split across physical source lines without a blank line. The reader should reflow the prose continuously rather than treating each source line as a new paragraph.",
		"The second source line continues the same paragraph and should not receive a blank display row or a new paragraph indent.",
		"",
		"The next paragraph is separated by a real blank source line.",
		"",
		"- first list item",
		"- second list item",
	}

	out := FormatParagraphs(raw, 62)

	// Check that the two list items appear consecutively without a blank row.
	foundItem1 := -1
	foundItem2 := -1
	for i, line := range out {
		if strings.Contains(line, "- first list item") {
			foundItem1 = i
		}
		if strings.Contains(line, "- second list item") {
			foundItem2 = i
		}
	}

	if foundItem1 < 0 || foundItem2 < 0 {
		t.Fatalf("could not find list items in output: %q", out)
	}

	if foundItem2 != foundItem1+1 {
		t.Errorf("list items are not adjacent (got %d and %d): %q", foundItem1, foundItem2, out)
	}

	// Check that "The second source line" is not preceded by a blank line
	// and does not have an artificial paragraph indent.
	for i, line := range out {
		if strings.Contains(line, "The second source line") {
			if strings.HasPrefix(line, "  The second source line") {
				t.Errorf("unexpected paragraph indentation on soft-wrapped continuation line: %q", line)
			}
			if i > 0 && out[i-1] == "" {
				t.Errorf("unexpected blank line before soft-wrapped continuation line: %q", out)
			}
		}
	}
}

func TestIssue77_ReflowedLinesLinksPreserved(t *testing.T) {
	raw := []string{
		"Paragraph starting here with [link one](#one)",
		"and continuation on line two with [link two](#two).",
	}
	pages := Paginate(raw, 120, 10)
	pages = AttachLinks(pages, raw, 120, 10)
	if len(pages[0].Links) != 2 {
		t.Fatalf("expected 2 links attached to page 0, got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
	if pages[0].Links[0].Target != "one" || pages[0].Links[1].Target != "two" {
		t.Errorf("expected targets one and two, got %+v", pages[0].Links)
	}
}

func TestIssue77_SecondMultiLineParagraphIndented(t *testing.T) {
	raw := []string{
		"First paragraph.",
		"",
		"Second paragraph line one",
		"and line two.",
	}
	out := FormatParagraphs(raw, 80)
	if len(out) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(out), out)
	}
	if !strings.HasPrefix(out[2], "  Second") {
		t.Errorf("expected 2-space indent on second multi-line paragraph, got %q", out[2])
	}
}
