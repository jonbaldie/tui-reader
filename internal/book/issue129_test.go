package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue129_IndentedFenceDelimiterKeepsAnchorsAndLinksInSync(t *testing.T) {
	lines := []string{
		"    ```",
		"# Swallowed Heading",
		"```go",
		"# Fake Heading In Real Fence",
		"```",
		"# Final Heading",
		"Jump to [final](#final-heading).",
	}

	wantAnchors := map[string]int{
		"swallowed-heading": 1,
		"final-heading":     5,
	}
	if got := ExtractAnchors(lines); !reflect.DeepEqual(got, wantAnchors) {
		t.Errorf("ExtractAnchors = %#v, want %#v", got, wantAnchors)
	}
	formatted := FormatParagraphs(lines, 80)
	for _, want := range []string{"# Swallowed Heading", "# Final Heading"} {
		if !containsFormattedLine(formatted, want) {
			t.Errorf("formatted output is missing displayed heading %q: %q", want, formatted)
		}
	}

	path := writeTempFile(t, "issue-129.md", strings.Join(lines, "\n"))
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}

	var links []Link
	for _, page := range b.Pages {
		links = append(links, page.Links...)
	}
	if len(links) != 1 {
		t.Errorf("links = %+v, want one link to final-heading", links)
	} else if links[0].Label != "final" || links[0].Target != "final-heading" {
		t.Errorf("link = %+v, want one link to final-heading", links[0])
	}
	if got := b.PageForAnchor("final-heading"); got < 0 {
		t.Errorf("PageForAnchor(final-heading) = %d, want a displayed heading page", got)
	}
}

func containsFormattedLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}
