package book

import (
	"reflect"
	"testing"
)

var issue111Lines = []string{
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

func TestIssue111_FencedCodeCommentIsNotAnAnchor(t *testing.T) {
	anchors := ExtractAnchors(issue111Lines)
	if got, want := anchors["usage"], 4; got != want {
		t.Errorf("anchors[usage] = %d, want %d (the ## Usage heading)", got, want)
	}
}

func TestIssue111_FencedCodeIsVerbatimCode(t *testing.T) {
	got := FormatParagraphs(issue111Lines, 60)[9:]
	want := []string{
		"",
		"```sh",
		"    # usage",
		"    tool --help",
		"```",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fenced block = %q, want %q", got, want)
	}
}

func TestIssue111_FencedCodeLinksAreNotActive(t *testing.T) {
	lines := []string{"# Intro", "", "~~~", "[Not a link](#intro)", "```", "~~~", "[Real link](#intro)"}
	source := collectSourceLinks(lines)
	if got, want := source.order, []int{6}; !reflect.DeepEqual(got, want) {
		t.Errorf("link source lines = %v, want %v", got, want)
	}
}

func TestIssue111_ProseAfterFenceIsParsedNormally(t *testing.T) {
	lines := []string{
		"Intro text",
		"```",
		"# one",
		"```",
		"## After",
		"````go",
		"```",
		"# two",
		"````",
		"Closing text",
	}
	if got, want := ExtractAnchors(lines), map[string]int{"after": 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("anchors = %v, want %v", got, want)
	}
	got := FormatParagraphs(lines, 60)
	want := []string{
		"Intro text",
		"",
		"```",
		"    # one",
		"```",
		"",
		"## After",
		"",
		"````go",
		"    ```",
		"    # two",
		"````",
		"",
		"  Closing text",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("formatted = %q, want %q", got, want)
	}
}

func TestIssue111_UnclosedFenceRunsToEnd(t *testing.T) {
	lines := []string{"````", "# code", "```python"}
	if got := ExtractAnchors(lines); len(got) != 0 {
		t.Errorf("anchors = %v, want none", got)
	}
	got := FormatParagraphs(lines, 60)
	want := []string{"````", "    # code", "    ```python"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("formatted = %q, want %q", got, want)
	}
}
