package book

import (
	"reflect"
	"strings"
	"testing"
)

func minFenceLines() []string {
	return []string{
		"# Guide",
		"",
		"[Usage](#usage)",
		"",
		"## Usage",
		"",
		"Real section.",
		"",
		"## Example",
		"",
		"```sh",
		"# usage",
		"tool --help",
		"```",
	}
}

func TestIssue111_FencedCodeDoesNotHijackAnchors(t *testing.T) {
	anchors := ExtractAnchors(minFenceLines())
	got, ok := anchors["usage"]
	if !ok {
		t.Fatal("missing usage anchor")
	}
	if got != 4 {
		t.Fatalf("anchors[usage] = %d, want 4 (## Usage); anchors=%#v", got, anchors)
	}
}

func TestIssue111_FencedCodeLinesAreNotMerged(t *testing.T) {
	formatted := FormatParagraphs(minFenceLines(), 60)
	for i, line := range formatted {
		if strings.Contains(line, "tool --help") && strings.Contains(line, "```") {
			t.Fatalf("line %d joins code with fence: %q", i, line)
		}
		if strings.Contains(line, "```sh") && strings.Contains(line, "# usage") {
			t.Fatalf("line %d joins opening fence with comment: %q", i, line)
		}
	}
	for _, want := range []string{"```sh", "# usage", "tool --help"} {
		if !formattedContains(formatted, want) {
			t.Fatalf("formatted output missing %q: %q", want, formatted)
		}
	}
	if !formattedContainsFenceClose(formatted) {
		t.Fatalf("formatted output missing closing fence: %q", formatted)
	}
}

func TestIssue111_FencedCodeDoesNotCreateAnchorsOrLinks(t *testing.T) {
	path := writeTempFile(t, "issue-111.md", "# Intro\n```\n# Not a heading\n[Not a link](#intro)\n```\n[Real link](#intro)\n")
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.Anchors, map[string]int{"intro": 0}; !reflect.DeepEqual(got, want) {
		t.Errorf("anchors = %#v, want %#v", got, want)
	}

	var links int
	var found Link
	for _, page := range b.Pages {
		links += len(page.Links)
		if len(page.Links) == 1 {
			found = page.Links[0]
		}
	}
	if links != 1 {
		t.Fatalf("total links = %d, want 1", links)
	}
	if found.Label != "Real link" || found.Target != "intro" {
		t.Fatalf("link = %+v, want the real link to intro", found)
	}
}

func TestIssue111_UsageLinkNavigatesToRealHeading(t *testing.T) {
	path := writeTempFile(t, "min-fence.md", strings.Join(minFenceLines(), "\n")+"\n")
	b, err := NewBook(path, 60, 8)
	if err != nil {
		t.Fatal(err)
	}
	got := b.PageForAnchor("usage")
	want := b.PageForRawLine(4)
	if got != want {
		t.Fatalf("PageForAnchor(usage) = %d, want page %d of ## Usage", got, want)
	}
	page := b.Pages[got]
	if !pageContains(page, "## Usage") {
		t.Fatalf("usage page %d missing ## Usage: %q", got, page.Lines)
	}
}

func TestIssue111_MultipleFencesDoNotAffectProse(t *testing.T) {
	lines := []string{
		"# Title",
		"```",
		"# fake one",
		"```",
		"## Real",
		"```sh",
		"# fake two",
		"echo hi",
		"```",
		"After fences.",
	}
	anchors := ExtractAnchors(lines)
	if _, ok := anchors["fake-one"]; ok {
		t.Fatalf("fake-one should not be an anchor: %#v", anchors)
	}
	if _, ok := anchors["fake-two"]; ok {
		t.Fatalf("fake-two should not be an anchor: %#v", anchors)
	}
	if got, ok := anchors["real"]; !ok || got != 4 {
		t.Fatalf("anchors[real] = %d (ok=%v), want 4; %#v", got, ok, anchors)
	}

	formatted := FormatParagraphs(lines, 60)
	if !formattedContains(formatted, "After fences.") {
		t.Fatalf("prose after fences missing: %q", formatted)
	}
	if !formattedContains(formatted, "echo hi") {
		t.Fatalf("second fence body missing: %q", formatted)
	}
	for i, line := range formatted {
		if strings.Contains(line, "echo hi") && strings.Contains(line, "```") {
			t.Fatalf("line %d joins second fence body with delimiter: %q", i, line)
		}
	}
}

func TestIssue111_FencedHashCommentIsNotAHeadingLine(t *testing.T) {
	formatted := formatParagraphsWithProvenance(minFenceLines(), 60)
	var comment formattedLine
	found := false
	for _, fl := range formatted {
		if strings.Contains(fl.text, "# usage") && !strings.Contains(fl.text, "##") {
			comment = fl
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing # usage display line: %v", projectText(formatted))
	}
	if isHeadingLine(comment.text) {
		t.Fatalf("fenced comment %q treated as a heading", comment.text)
	}
}

func formattedContains(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

func formattedContainsFenceClose(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "```" {
			return true
		}
	}
	return false
}

func pageContains(page Page, want string) bool {
	for _, line := range page.Lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}
