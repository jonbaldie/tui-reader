package book

import (
	"strings"
	"testing"
)

func setextLines() []string {
	return []string{
		"[Chapter Two](#chapter-two)",
		"",
		"Filler.",
		"",
		"Chapter Two",
		"===========",
		"",
		"A Section",
		"---------",
		"",
		"Body text.",
	}
}

func TestIssue179_SetextHeadingsAreAnchors(t *testing.T) {
	anchors := ExtractAnchors(setextLines())
	if got, ok := anchors["chapter-two"]; !ok || got != 4 {
		t.Fatalf("anchors[chapter-two] = %d (ok=%v), want 4; %#v", got, ok, anchors)
	}
	if got, ok := anchors["a-section"]; !ok || got != 7 {
		t.Fatalf("anchors[a-section] = %d (ok=%v), want 7; %#v", got, ok, anchors)
	}
}

func TestIssue179_SetextUnderlineNotMergedIntoProse(t *testing.T) {
	for i, line := range FormatParagraphs(setextLines(), 60) {
		if strings.Contains(line, "===") || strings.Contains(line, "---") {
			t.Fatalf("formatted line %d still contains setext underline: %q", i, line)
		}
	}
}

func TestIssue179_LinkNavigatesToSetextHeading(t *testing.T) {
	path := writeTempFile(t, "setext.md", strings.Join(setextLines(), "\n")+"\n")
	b, err := NewBook(path, 60, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.PageForAnchor("chapter-two"), b.PageForRawLine(4); got != want || got == 0 {
		t.Fatalf("PageForAnchor(chapter-two) = %d, want page %d", got, want)
	}
}
