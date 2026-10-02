package book

import (
	"strings"
	"testing"
)

// Every attached link must sit on a display line that renders its markup, and
// each page must carry exactly as many links as its lines render; issue #148.
func TestIssue148_RepeatedMarkupInParagraphAttachesToOwnLine(t *testing.T) {
	content := strings.Join([]string{
		"Opening words",
		"[a](#title) and [b](#title) then filler filler filler filler filler",
		"filler filler filler filler filler filler filler filler [a](#title)",
	}, "\n")
	b, err := Read(strings.NewReader(content), "x", 40, 3, false)
	if err != nil {
		t.Fatal(err)
	}

	for pi, page := range b.Pages {
		rendered := 0
		for _, line := range page.Lines {
			rendered += len(ExtractLinks(line))
		}
		if len(page.Links) != rendered {
			t.Errorf("page %d renders %d links but has %d attached: lines %q, links %+v",
				pi, rendered, len(page.Links), page.Lines, page.Links)
		}
		for _, link := range page.Links {
			if link.LineOnPage >= len(page.Lines) || !strings.Contains(page.Lines[link.LineOnPage], linkMarkup(link)) {
				t.Errorf("page %d: link %+v is not on a line rendering it: lines %q", pi, link, page.Lines)
			}
		}
	}
}
