package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue125_HeadingAndProseBehavior(t *testing.T) {
	cases := []struct {
		name, doc, anchor string
		lines             []string
		headings          []bool
	}{
		{"number after prose", "Intro\n#142", "142", []string{"Intro #142"}, []bool{false}},
		{"hashtag after prose", "Intro\n#tag", "tag", []string{"Intro #tag"}, []bool{false}},
		{"bare hash after prose", "Intro\n#", "", []string{"Intro #"}, []bool{false}},
		{"seven hashes after prose", "Intro\n####### x", "x", []string{"Intro ####### x"}, []bool{false}},
		{"valid heading", "Intro\n## Heading", "heading", []string{"Intro", "", "## Heading"}, []bool{false, false, true}},
		{"indented code", "Intro\n    # Heading", "heading", []string{"Intro", "", "    # Heading"}, []bool{false, false, false}},
		{"fenced code", "```\n# Heading\n#tag\n```", "heading", []string{"    ```", "    # Heading", "    #tag", "    ```"}, []bool{false, false, false, false}},
		{"setext hash title", "Intro\n#142\n===", "intro-142", []string{"# Intro #142"}, []bool{true}},
		{"setext hashtag", "#tag\n---", "tag", []string{"## #tag"}, []bool{true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := readBook(t, tc.doc, true, 80, 20)
			page := b.Page(0)
			if !reflect.DeepEqual(page.Lines, tc.lines) || !reflect.DeepEqual(page.Headings, tc.headings) {
				t.Errorf("lines/headings = %q / %v, want %q / %v", page.Lines, page.Headings, tc.lines, tc.headings)
			}
			if tc.anchor != "" {
				want := -1
				if tc.headings[len(tc.headings)-1] {
					want = 0
				}
				if got := b.PageForAnchor(tc.anchor); got != want {
					t.Errorf("PageForAnchor(%q) = %d, want %d", tc.anchor, got, want)
				}
			}
		})
	}
}

func TestIssue125_LinkAcrossJoinedHashLine(t *testing.T) {
	b := readBook(t, "# Target\n\nSee [label\n#tag](#target).\n\n```\n#comment [hidden](#target)\n```", true, 80, 20)
	page := b.Page(0)
	if len(page.Links) != 1 {
		t.Fatalf("links = %+v, want one visible link", page.Links)
	}
	link := page.Links[0]
	if link.Label != "label #tag" || link.Target != "target" || !strings.Contains(page.Lines[link.LineOnPage], "[label #tag](#target)") {
		t.Errorf("link = %+v on lines %q", link, page.Lines)
	}
	if got := b.PageForAnchor(link.Target); got != 0 {
		t.Errorf("PageForAnchor(%q) = %d, want 0", link.Target, got)
	}
}
