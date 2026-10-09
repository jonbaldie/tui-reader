package book

import (
	"strings"
	"testing"
)

// Issue #165: a ```go line inside an open fence is content, not a closer.
// The false close turns the sample heading into an anchor and swallows later headings.
func TestIssue165_InfoStringLineDoesNotCloseFence(t *testing.T) {
	content := strings.Join([]string{
		"```",
		"```go",
		"# Fake Heading",
		"```",
		"## Usage",
		"See [usage](#usage).",
	}, "\n")
	rawLines := strings.Split(content, "\n")

	if end := findFenceBlockEnd(rawLines, 0); end != 4 {
		t.Fatalf("findFenceBlockEnd = %d, want 4 (the bare closing fence)", end)
	}

	path := writeTempFile(t, "issue-165.md", content)
	b, err := NewBook(path, 80, 40)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := b.anchors["usage"]; !ok {
		t.Errorf("anchors missing usage: %#v", b.anchors)
	}
	if _, ok := b.anchors["fake-heading"]; ok {
		t.Errorf("anchors contain fake-heading, want it inside the fence: %#v", b.anchors)
	}
	if page := b.PageForAnchor("usage"); page < 0 {
		t.Errorf("PageForAnchor(usage) = %d, want a displayed heading page", page)
	}

	var links []Link
	for _, page := range b.pages {
		links = append(links, page.Links...)
	}
	if len(links) != 1 || links[0].Label != "usage" || links[0].Target != "usage" {
		t.Errorf("links = %+v, want one link to usage", links)
	} else if got := b.PageForAnchor(links[0].Target); got < 0 {
		t.Errorf("PageForAnchor(%q) = %d, want the usage heading", links[0].Target, got)
	}

	formatted := FormatParagraphs(rawLines, 80)
	if !containsFormattedLine(formatted, "    # Fake Heading") || containsFormattedLine(formatted, "# Fake Heading") {
		t.Errorf("# Fake Heading rendered as a heading, want code: %q", formatted)
	}
	if containsFormattedLine(formatted, "    ## Usage") || !containsFormattedLine(formatted, "## Usage") {
		t.Errorf("## Usage not rendered as a heading: %q", formatted)
	}
}
