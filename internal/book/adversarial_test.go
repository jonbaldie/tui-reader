package book

import (
	"fmt"
	"strings"
	"testing"
)

// ==================== Stress: extremely long lines ====================

func TestAdversarial_VeryLongLine(t *testing.T) {
	line := strings.Repeat("word ", 2000) // 10,000 chars
	result := WrapLines([]string{line}, 60)
	for i, l := range result {
		if runeLen(l) > 60 {
			t.Errorf("wrapped line %d exceeds width 60: %d runes", i, runeLen(l))
		}
	}
	// Should produce many lines
	if len(result) < 100 {
		t.Errorf("expected 100+ wrapped lines from 10k char line, got %d", len(result))
	}
}

func TestAdversarial_SingleCharWidth(t *testing.T) {
	line := "hello world"
	result := WrapLines([]string{line}, 1)
	// Each character should be on its own line (word "hello" gets hard-broken)
	for i, l := range result {
		if runeLen(l) > 1 {
			t.Errorf("line %d exceeds width 1: %q", i, l)
		}
	}
}

// ==================== Stress: large file ====================

func TestAdversarial_LargeFile(t *testing.T) {
	lines := make([]string, 100000)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line number %d of the document.", i)
	}
	pages := Paginate(lines, 80, 25)
	if len(pages) < 4000 {
		t.Errorf("expected 4000+ pages for 100k lines at height 25, got %d", len(pages))
	}
	// First and last page should have content
	if len(pages[0].Lines) == 0 {
		t.Error("first page has no lines")
	}
	if len(pages[len(pages)-1].Lines) == 0 {
		t.Error("last page has no lines")
	}
}

// ==================== Edge: only headings, no body ====================

func TestAdversarial_OnlyHeadings(t *testing.T) {
	lines := []string{
		"# Heading 1",
		"## Heading 2",
		"### Heading 3",
	}
	anchors := ExtractAnchors(lines)
	if len(anchors) != 3 {
		t.Errorf("expected 3 anchors, got %d", len(anchors))
	}
	pages := Paginate(lines, 80, 10)
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
}

// ==================== Edge: duplicate heading names ====================

func TestAdversarial_DuplicateHeadings(t *testing.T) {
	lines := []string{
		"# Chapter",
		"First chapter text",
		"# Chapter",
		"Second chapter text",
	}
	anchors := ExtractAnchors(lines)
	// Duplicate heading: second overwrites first
	// This is a KNOWN BEHAVIOR - the last occurrence wins
	idx, ok := anchors["chapter"]
	if !ok {
		t.Fatal("expected 'chapter' anchor")
	}
	// The second "# Chapter" is at line 2
	if idx != 2 {
		t.Logf("NOTE: duplicate headings - anchor points to line %d (last wins)", idx)
	}
	// This is a real issue: the first heading's anchor is unreachable.
	// Links to #chapter will always go to the second one.
}

// ==================== Edge: link to nonexistent anchor ====================

func TestAdversarial_LinkToNonexistentAnchor(t *testing.T) {
	content := "[Click here](#does-not-exist)\n# Real Heading\n"
	path := writeTempFile(t, "badlink.md", content)
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	// The link should be detected
	if len(b.Pages[0].Links) == 0 {
		t.Fatal("expected link to be detected even if target doesn't exist")
	}
	// PageForAnchor should return -1
	page := b.PageForAnchor("does-not-exist")
	if page != -1 {
		t.Errorf("expected -1 for nonexistent anchor, got %d", page)
	}
}

// ==================== Edge: whitespace-only file ====================

func TestAdversarial_WhitespaceOnlyFile(t *testing.T) {
	path := writeTempFile(t, "spaces.txt", "   \n  \n    \n")
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Pages) == 0 {
		t.Error("expected at least 1 page")
	}
}

// ==================== Edge: single newline file ====================

func TestAdversarial_SingleNewline(t *testing.T) {
	path := writeTempFile(t, "newline.txt", "\n")
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Pages) == 0 {
		t.Error("expected at least 1 page")
	}
}

// ==================== BUG: Heading regex requires space after # ====================

func TestAdversarial_HeadingNoSpace(t *testing.T) {
	// "#NoSpace" should NOT be treated as a heading (markdown spec requires space)
	lines := []string{"#NoSpace", "# With Space"}
	anchors := ExtractAnchors(lines)
	if _, ok := anchors["nospace"]; ok {
		t.Error("#NoSpace without space should not be an anchor")
	}
	if _, ok := anchors["with-space"]; !ok {
		t.Error("# With Space should be an anchor")
	}
}

// ==================== BUG HUNT: link text split across wrapped lines ====================

func TestAdversarial_LinkTextWrapped(t *testing.T) {
	// If a line with a link is long enough to wrap, the link markdown syntax
	// might get split across lines, making it undetectable in the wrapped version.
	// The raw line should still have the link though.
	longLine := strings.Repeat("padding ", 10) + "[Very Important Link](#target) " + strings.Repeat("more ", 10)
	rawLines := []string{"# Target", "", longLine}

	pages := Paginate(rawLines, 40, 10) // narrow enough to force wrapping
	pages = AttachLinks(pages, rawLines, 40, 10)

	// The link should still be found (from the raw line, not the wrapped version)
	found := false
	for _, p := range pages {
		for _, lnk := range p.Links {
			if lnk.Target == "target" {
				found = true
			}
		}
	}
	if !found {
		t.Error("BUG: link in wrapped line was not detected")
	}
}

// ==================== BUG HUNT: wrapLine with width=1 ====================

func TestAdversarial_WrapWidth1(t *testing.T) {
	result := wrapLine("ab cd", 1)
	for i, l := range result {
		if runeLen(l) > 1 {
			t.Errorf("line %d exceeds width 1: %q", i, l)
		}
	}
	// Should produce 4 lines: a, b, c, d
	if len(result) != 4 {
		t.Errorf("expected 4 lines at width 1 for 'ab cd', got %d: %v", len(result), result)
	}
}

// ==================== BUG HUNT: wrapLine with width=0 ====================

func TestAdversarial_WrapWidth0(t *testing.T) {
	// Should not panic
	result := wrapLine("hello", 0)
	if len(result) == 0 {
		t.Error("expected at least 1 line")
	}
}

// ==================== BUG HUNT: Paginate with height=1 ====================

func TestAdversarial_PaginateHeight1(t *testing.T) {
	// 3 raw lines -> "a", "", "  b", "", "  c" = 5 formatted lines
	// At height 1: 5 pages, each with 1 line
	lines := []string{"a", "b", "c"}
	pages := Paginate(lines, 80, 1)
	if len(pages) != 5 {
		t.Errorf("expected 5 pages at height 1 (3 content + 2 spacers), got %d", len(pages))
	}
	for i, p := range pages {
		if len(p.Lines) != 1 {
			t.Errorf("page %d should have exactly 1 line, got %d", i, len(p.Lines))
		}
	}
}

// ==================== BUG HUNT: empty anchor name ====================

func TestAdversarial_EmptyAnchorName(t *testing.T) {
	result := NormalizeAnchor("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// ==================== BUG HUNT: link with empty target ====================

func TestAdversarial_LinkEmptyTarget(t *testing.T) {
	links := ExtractLinks("[text](#)")
	// The regex requires at least one char after #
	// This depends on regex: `#([^)]+)` requires 1+ chars
	if len(links) != 0 {
		t.Log("NOTE: empty anchor target is accepted - may want to reject")
	}
}

// ==================== BUG HUNT: consecutive links on same line ====================

func TestAdversarial_ConsecutiveLinks(t *testing.T) {
	line := "[A](#a)[B](#b)[C](#c)"
	links := ExtractLinks(line)
	if len(links) != 3 {
		t.Errorf("expected 3 consecutive links, got %d", len(links))
	}
}

// ==================== BUG HUNT: markdown in link labels ====================

func TestAdversarial_MarkdownInLinkLabel(t *testing.T) {
	// Bold text in link label
	line := "[**Bold Label**](#target)"
	links := ExtractLinks(line)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Label != "**Bold Label**" {
		t.Logf("NOTE: link label includes markdown formatting: %q", links[0].Label)
	}
}

// ==================== BUG HUNT: PageForAnchor with zero PageHeight ====================

func TestAdversarial_PageForAnchorZeroHeight(t *testing.T) {
	content := "# Heading\ntext\n"
	path := writeTempFile(t, "zeroh.md", content)
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Manually set PageHeight to 0 (shouldn't happen normally but let's be safe)
	b.PageHeight = 0
	// This should not panic (division by zero)
	page := b.PageForAnchor("heading")
	_ = page // just checking it doesn't panic
}

// ==================== BUG HUNT: Reflow to very small dimensions ====================

func TestAdversarial_ReflowTiny(t *testing.T) {
	content := "# Hello\n\nSome text here.\n"
	path := writeTempFile(t, "tiny.md", content)
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic
	b.Reflow(1, 1)
	if len(b.Pages) == 0 {
		t.Error("expected at least 1 page after tiny reflow")
	}
}
