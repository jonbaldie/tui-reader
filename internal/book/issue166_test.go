package book

import (
	"strings"
	"testing"
)

func TestIssue166_LongFencePreservesHeadingsAndLinks(t *testing.T) {
	content := strings.Join([]string{
		"# Title",
		"",
		"See [usage](#usage).",
		"",
		"````",
		"```",
		"sample",
		"# not a heading",
		"````",
		"",
		"## Usage",
		"",
		"Real section.",
	}, "\n")
	rawLines := strings.Split(content, "\n")

	if end := findFenceBlockEnd(rawLines, 4); end != 9 {
		t.Errorf("findFenceBlockEnd = %d, want 9 (after the four-backtick closer)", end)
	}

	path := writeTempFile(t, "issue-166.md", content)
	b, err := NewBook(path, 80, 40)
	if err != nil {
		t.Fatal(err)
	}

	usageLine, ok := b.Anchors["usage"]
	if !ok {
		t.Errorf("anchors missing usage: %#v", b.Anchors)
	} else if usageLine != 10 {
		t.Errorf("anchors[usage] = %d, want 10 (## Usage); anchors=%#v", usageLine, b.Anchors)
	}
	if _, ok := b.Anchors["not-a-heading"]; ok {
		t.Errorf("anchors contain not-a-heading, want it inside the fence: %#v", b.Anchors)
	}
	usagePage := b.PageForAnchor("usage")
	if usagePage < 0 {
		t.Errorf("PageForAnchor(usage) = %d, want a displayed heading page", usagePage)
	} else if !pageContains(b.Pages[usagePage], "## Usage") {
		t.Errorf("usage page %d does not contain ## Usage: %q", usagePage, b.Pages[usagePage].Lines)
	}

	var links []Link
	for _, page := range b.Pages {
		links = append(links, page.Links...)
	}
	if len(links) != 1 || links[0].Label != "usage" || links[0].Target != "usage" {
		t.Errorf("links = %+v, want one link to usage", links)
	} else if page := b.PageForAnchor(links[0].Target); page != usagePage {
		t.Errorf("PageForAnchor(%q) = %d, want usage page %d", links[0].Target, page, usagePage)
	}

	formatted := FormatParagraphs(rawLines, 80)
	for _, want := range []string{"    ````", "    ```", "    # not a heading", "## Usage"} {
		if !containsFormattedLine(formatted, want) {
			t.Errorf("formatted output missing %q: %q", want, formatted)
		}
	}
	if containsFormattedLine(formatted, "    ## Usage") {
		t.Errorf("## Usage rendered as code: %q", formatted)
	}
}

func TestIssue166_ClosingFenceMayBeLongerThanOpener(t *testing.T) {
	rawLines := []string{"```", "````", "## Usage"}
	if end := findFenceBlockEnd(rawLines, 0); end != 2 {
		t.Errorf("findFenceBlockEnd = %d, want 2 (after the longer closing fence)", end)
	}
}
