package book

import (
	"reflect"
	"testing"
)

func TestIssue144_HeadingMetadataIsCopiedAndReflowed(t *testing.T) {
	b := readBook(t, "Intro\n\n## A heading long enough that it wraps", true, 24, 3)
	page := b.Page(0)
	page.Headings[2] = false
	if !b.Page(0).Headings[2] {
		t.Fatal("changing returned heading flags changed the book")
	}
	b.Reflow(60, 20)
	if got, want := b.Page(0).Headings, []bool{false, false, true}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reflowed headings = %v, want %v", got, want)
	}
}

func TestIssue144_PagePositionUsesLayoutHeadingClassification(t *testing.T) {
	cases := []struct {
		name, doc string
		markdown  bool
		height    int
		want      Position
	}{
		{"plain text hash", "intro\nintro2\nfirst\n# fake", false, 2, Position{Raw: 2}},
		{"wrapped prose hash", "intro\n\nThe fix landed in pull #142 after review.", true, 2, Position{Raw: 2}},
		{"heading continuation before another heading", "start\n\n## A heading long enough that it wraps\n\n# Later", true, 3, Position{Raw: 2, Offset: 20}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := readBook(t, tc.doc, tc.markdown, 24, tc.height)
			if got := b.PositionForPage(1); got != tc.want {
				t.Fatalf("PositionForPage(1) = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestIssue144_LayoutCarriesHeadingClassification(t *testing.T) {
	cases := []struct {
		name     string
		doc      string
		markdown bool
		lines    []string
		headings []bool
	}{
		{"plain text hash", "# not a heading in a log", false, []string{"# not a heading in a log"}, []bool{false}},
		{"wrapped prose hash", "The fix landed in pull #142 after review.", true, []string{"The fix landed in pull", "#142 after review."}, []bool{false, false}},
		{"wrapped ATX heading", "## A heading long enough that it wraps", true, []string{"## A heading long enough", "that it wraps"}, []bool{true, true}},
		{"existing setext heading", "A heading long enough that it wraps\n===", true, []string{"# A heading long enough", "that it wraps"}, []bool{true, true}},
		{"not an anchor heading", "#142 after review.", true, []string{"#142 after review."}, []bool{false}},
		{"fenced code", "```\n# not a heading\n```", true, []string{"    ```", "    # not a heading", "    ```"}, []bool{false, false, false}},
		{"separator", "# Heading\nBody", true, []string{"# Heading", "", "  Body"}, []bool{true, false, false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := readBook(t, tc.doc, tc.markdown, 24, 1)
			var lines []string
			var headings []bool
			for p := 0; p < b.PageCount(); p++ {
				page := b.Page(p)
				if len(page.Headings) != len(page.Lines) {
					t.Fatalf("page %d: %d heading flags for %d lines", p, len(page.Headings), len(page.Lines))
				}
				lines = append(lines, page.Lines...)
				headings = append(headings, page.Headings...)
			}
			if !reflect.DeepEqual(lines, tc.lines) || !reflect.DeepEqual(headings, tc.headings) {
				t.Fatalf("lines/headings = %q / %v, want %q / %v", lines, headings, tc.lines, tc.headings)
			}
		})
	}
}
