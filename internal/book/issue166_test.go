package book

import (
	"strings"
	"testing"
)

func TestIssue166LongerBacktickFencesPreserveFollowingHeadingsAndLinks(t *testing.T) {
	tests := []struct {
		name    string
		content []string
	}{
		{
			name: "long opener with nested sample fence",
			content: []string{
				"See [usage](#usage).",
				"````",
				"```",
				"# not a heading",
				"````",
				"## Usage",
			},
		},
		{
			name: "long closer ends normal fence",
			content: []string{
				"```",
				"# not a heading",
				"````",
				"## Usage",
				"[usage](#usage)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, "issue-166.md", strings.Join(tt.content, "\n"))
			b, err := NewBook(path, 80, 40)
			if err != nil {
				t.Fatal(err)
			}

			formatted := FormatParagraphs(tt.content, 80)
			if !containsFormattedLine(formatted, "## Usage") {
				t.Errorf("## Usage not rendered as a heading: %q", formatted)
			}

			if _, ok := b.Anchors["not-a-heading"]; ok {
				t.Errorf("anchors contain a heading from inside the fence: %#v", b.Anchors)
			}
			if _, ok := b.Anchors["usage"]; !ok {
				t.Errorf("anchors missing usage: %#v", b.Anchors)
			}
			if page := b.PageForAnchor("usage"); page < 0 {
				t.Errorf("PageForAnchor(usage) = %d, want a displayed heading page", page)
			}

			var links []Link
			for _, page := range b.Pages {
				links = append(links, page.Links...)
			}
			if len(links) != 1 || links[0].Target != "usage" {
				t.Errorf("links = %+v, want one link to usage", links)
			}
		})
	}
}
