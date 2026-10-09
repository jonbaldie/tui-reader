package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue167_TabIndentedCodeDoesNotCreateAnchorsOrLinks(t *testing.T) {
	content := "# Intro\n\t# Not a heading\n\t[Not a link](#intro)\n[Real link](#intro)"
	path := writeTempFile(t, "issue-167.md", content)

	b, err := NewBook(path, 80, 24)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := b.anchors, map[string]int{"intro": 0}; !reflect.DeepEqual(got, want) {
		t.Errorf("anchors = %#v, want %#v", got, want)
	}

	var links []Link
	var lines []string
	for _, page := range b.pages {
		links = append(links, page.Links...)
		lines = append(lines, page.Lines...)
	}
	if len(links) != 1 || links[0].Label != "Real link" || links[0].Target != "intro" {
		t.Errorf("links = %+v, want only the real link to intro", links)
	}

	want := []string{"    # Not a heading", "    [Not a link](#intro)"}
	for _, w := range want {
		found := false
		for _, line := range lines {
			if line == w {
				found = true
			}
		}
		if !found {
			t.Errorf("display lines %q missing code line %q", lines, w)
		}
	}
	for _, line := range lines {
		if strings.Contains(line, "\t") {
			t.Errorf("display line %q contains a raw tab", line)
		}
	}
}

func TestIssue167_IndentedCodeLineTabStops(t *testing.T) {
	for raw, want := range map[string]bool{
		"\tcode":    true,
		" \tcode":   true,
		"   \tcode": true,
		"    code":  true,
		"   code":   false,
		"code":      false,
		"":          false,
	} {
		if got := IsIndentedCodeLine(raw); got != want {
			t.Errorf("IsIndentedCodeLine(%q) = %v, want %v", raw, got, want)
		}
	}
}
