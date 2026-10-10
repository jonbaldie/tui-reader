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

func TestIssue179_SetextHeadingsDisplayAsATX(t *testing.T) {
	formatted := FormatParagraphs(setextLines(), 60)
	for _, want := range []string{"# Chapter Two", "## A Section"} {
		found := false
		for _, line := range formatted {
			found = found || line == want
		}
		if !found {
			t.Fatalf("formatted output lacks %q: %q", want, formatted)
		}
	}
}

func TestIssue179_RuleWithoutPrecedingProseStaysRule(t *testing.T) {
	for _, raw := range [][]string{
		{"Para", "", "---"},
		{"- item", "---"},
		{"# Heading", "---"},
		{"---"},
	} {
		formatted := FormatParagraphs(raw, 60)
		if got := formatted[len(formatted)-1]; got != "---" {
			t.Fatalf("FormatParagraphs(%q) = %q, want trailing rule", raw, formatted)
		}
	}
}

func TestIssue179_EqualsWithoutPrecedingProseStaysProse(t *testing.T) {
	formatted := FormatParagraphs([]string{"# Heading", "==="}, 60)
	if got := formatted[len(formatted)-1]; got != "  ===" {
		t.Fatalf("formatted = %q, want === as a prose paragraph", formatted)
	}
}

func TestIssue179_MultiLineSetextHeadingJoinsText(t *testing.T) {
	anchors := ExtractAnchors([]string{"Part One", "The Beginning", "==="})
	if got, ok := anchors["part-one-the-beginning"]; !ok || got != 0 {
		t.Fatalf("anchors = %#v, want part-one-the-beginning at 0", anchors)
	}
}

func TestIssue179_PlainTextUnaffected(t *testing.T) {
	b, err := Read(strings.NewReader("Chapter Two\n===========\n"), "t", 60, 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.PageForAnchor("chapter-two"); got != -1 {
		t.Fatalf("plain text PageForAnchor = %d, want -1", got)
	}
	if got := b.Page(0).Lines; len(got) < 2 || got[0] != "Chapter Two" || got[1] != "===========" {
		t.Fatalf("plain text lines = %q, want unchanged", got)
	}
}
