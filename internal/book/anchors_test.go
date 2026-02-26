package book

import (
	"testing"
)

// ==================== NormalizeAnchor ====================

func TestNormalizeAnchor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Chapter 1", "chapter-1"},
		{"Chapter 1: Introduction", "chapter-1-introduction"},
		{"Hello World", "hello-world"},
		{"Already-Hyphenated", "already-hyphenated"},
		{"Special!@#$%Characters", "specialcharacters"},
		{"  Spaces  Everywhere  ", "spaces-everywhere"},
		{"UPPER CASE", "upper-case"},
		{"", ""},
		{"123 Numbers", "123-numbers"},
		{"a", "a"},
	}
	for _, tt := range tests {
		got := NormalizeAnchor(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeAnchor(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ==================== ExtractAnchors ====================

func TestExtractAnchors_MarkdownHeadings(t *testing.T) {
	lines := []string{
		"# Chapter 1",
		"Some text",
		"## Section 1.1",
		"More text",
		"### Subsection",
		"Even more",
	}
	anchors := ExtractAnchors(lines)

	expected := map[string]int{
		"chapter-1":  0,
		"section-11": 2,
		"subsection": 4,
	}
	for name, wantLine := range expected {
		gotLine, ok := anchors[name]
		if !ok {
			t.Errorf("missing anchor %q", name)
			continue
		}
		if gotLine != wantLine {
			t.Errorf("anchor %q: expected line %d, got %d", name, wantLine, gotLine)
		}
	}
}

func TestExtractAnchors_NoHeadings(t *testing.T) {
	lines := []string{"plain text", "more text", "no headings here"}
	anchors := ExtractAnchors(lines)
	if len(anchors) != 0 {
		t.Errorf("expected 0 anchors, got %d", len(anchors))
	}
}

func TestExtractAnchors_HashInMiddle(t *testing.T) {
	// "#" must be at the start of the line
	lines := []string{"text with # in middle", "also not ## a heading"}
	anchors := ExtractAnchors(lines)
	if len(anchors) != 0 {
		t.Errorf("expected 0 anchors for mid-line hashes, got %d", len(anchors))
	}
}

func TestExtractAnchors_HeadingLevels(t *testing.T) {
	lines := []string{
		"# H1",
		"## H2",
		"### H3",
		"#### H4",
		"##### H5",
		"###### H6",
		"####### Not a heading",
	}
	anchors := ExtractAnchors(lines)
	if len(anchors) != 6 {
		t.Errorf("expected 6 anchors (h1-h6), got %d", len(anchors))
	}
}

// ==================== ExtractLinks ====================

func TestExtractLinks_SingleLink(t *testing.T) {
	line := "See [Chapter 1](#chapter-1) for details."
	links := ExtractLinks(line)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Label != "Chapter 1" {
		t.Errorf("expected label 'Chapter 1', got %q", links[0].Label)
	}
	if links[0].Target != "chapter-1" {
		t.Errorf("expected target 'chapter-1', got %q", links[0].Target)
	}
}

func TestExtractLinks_MultipleLinks(t *testing.T) {
	line := "[Intro](#intro) and [Conclusion](#conclusion)"
	links := ExtractLinks(line)
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
}

func TestExtractLinks_NoLinks(t *testing.T) {
	line := "No links here, just plain text."
	links := ExtractLinks(line)
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestExtractLinks_ExternalLinkIgnored(t *testing.T) {
	// Only internal (#anchor) links should be extracted
	line := "[Google](https://google.com)"
	links := ExtractLinks(line)
	if len(links) != 0 {
		t.Errorf("expected 0 links for external URL, got %d", len(links))
	}
}

func TestExtractLinks_NestedBrackets(t *testing.T) {
	line := "Text [label](#target) more text"
	links := ExtractLinks(line)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
}

// ==================== AttachLinks ====================

func TestAttachLinks_LinksOnCorrectPage(t *testing.T) {
	rawLines := []string{
		"# Introduction",
		"",
		"See [Chapter 1](#chapter-1) for details.",
		"",
		"# Chapter 1",
		"Content here.",
	}
	pages := Paginate(rawLines, 80, 3) // 3 lines per page -> 2 pages
	pages = AttachLinks(pages, rawLines, 80, 3)

	// The link should be on page index 0 (line 2 is "See [Chapter 1]...")
	if len(pages) < 1 {
		t.Fatal("expected at least 1 page")
	}
	if len(pages[0].Links) == 0 {
		t.Fatal("expected links on page 0")
	}
	if pages[0].Links[0].Target != "chapter-1" {
		t.Errorf("expected target 'chapter-1', got %q", pages[0].Links[0].Target)
	}
}

func TestAttachLinks_NoLinks(t *testing.T) {
	rawLines := []string{"plain text", "no links"}
	pages := Paginate(rawLines, 80, 10)
	pages = AttachLinks(pages, rawLines, 80, 10)
	if len(pages[0].Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(pages[0].Links))
	}
}
