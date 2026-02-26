package book

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mutation tests verify that our test suite catches specific types of bugs.
// Each test simulates a mutation by testing boundary conditions and invariants
// that would break if key logic were altered.

// ==================== Pagination boundary mutations ====================

// Mutation: changing `i += height` to `i += height-1` or `i += height+1`
// would produce overlapping or gapped pages. Verify continuity.
func TestMutation_PaginationContinuity(t *testing.T) {
	lines := make([]string, 25)
	for i := range lines {
		lines[i] = "line"
	}
	pages := Paginate(lines, 80, 10)

	// All lines must appear exactly once across all pages
	totalLines := 0
	for _, p := range pages {
		totalLines += len(p.Lines)
	}
	if totalLines != 25 {
		t.Errorf("expected 25 total lines across pages, got %d", totalLines)
	}

	// No page should exceed height
	for i, p := range pages {
		if len(p.Lines) > 10 {
			t.Errorf("page %d has %d lines, exceeds height 10", i, len(p.Lines))
		}
	}
}

// Mutation: changing `end > len(wrapped)` guard to `>=` or removing it
// would cause out-of-bounds or missing last lines.
func TestMutation_LastPageInclusion(t *testing.T) {
	lines := make([]string, 11)
	for i := range lines {
		lines[i] = "x"
	}
	pages := Paginate(lines, 80, 10)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}
	if len(pages[1].Lines) != 1 {
		t.Errorf("expected 1 line on last page, got %d", len(pages[1].Lines))
	}
	if pages[1].Lines[0] != "x" {
		t.Errorf("expected 'x' on last page, got %q", pages[1].Lines[0])
	}
}

// ==================== Line wrapping boundary mutations ====================

// Mutation: changing `<=` to `<` in width comparison would cause off-by-one.
func TestMutation_WrapExactWidth(t *testing.T) {
	// Line exactly at width should NOT be wrapped
	line := strings.Repeat("a", 40)
	result := WrapLines([]string{line}, 40)
	if len(result) != 1 {
		t.Errorf("line at exact width should not wrap, got %d lines", len(result))
	}
}

// Mutation: changing `>` to `>=` in width comparison would wrap too eagerly.
func TestMutation_WrapOneOverWidth(t *testing.T) {
	line := strings.Repeat("a", 41) // one rune over
	result := WrapLines([]string{line}, 40)
	if len(result) != 2 {
		t.Errorf("expected 2 lines for 41 runes at width 40, got %d", len(result))
	}
}

// Mutation: removing the empty-fields check would panic on whitespace-only lines.
func TestMutation_WrapWhitespaceOnly(t *testing.T) {
	result := WrapLines([]string{"   "}, 40)
	// strings.Fields("   ") returns [], so wrapLine returns [""]
	if len(result) != 1 {
		t.Errorf("expected 1 empty line, got %d", len(result))
	}
}

// ==================== Anchor normalization mutations ====================

// Mutation: removing ToLower would make anchors case-sensitive.
func TestMutation_AnchorCaseInsensitive(t *testing.T) {
	a1 := NormalizeAnchor("Chapter One")
	a2 := NormalizeAnchor("CHAPTER ONE")
	a3 := NormalizeAnchor("chapter one")
	if a1 != a2 || a2 != a3 {
		t.Errorf("expected same anchor for different cases: %q, %q, %q", a1, a2, a3)
	}
}

// Mutation: removing hyphen collapsing would leave double hyphens.
func TestMutation_AnchorNoDoubleHyphens(t *testing.T) {
	result := NormalizeAnchor("Hello   World")
	if strings.Contains(result, "--") {
		t.Errorf("anchor should not contain double hyphens: %q", result)
	}
}

// Mutation: not trimming trailing hyphens.
func TestMutation_AnchorNoTrailingHyphen(t *testing.T) {
	result := NormalizeAnchor("Hello!")
	if strings.HasSuffix(result, "-") {
		t.Errorf("anchor should not end with hyphen: %q", result)
	}
}

// ==================== Link extraction mutations ====================

// Mutation: regex captures wrong groups.
func TestMutation_LinkLabelAndTargetCorrect(t *testing.T) {
	links := ExtractLinks("[Go Here](#destination)")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Label != "Go Here" {
		t.Errorf("label should be 'Go Here', got %q", links[0].Label)
	}
	if links[0].Target != "destination" {
		t.Errorf("target should be 'destination', got %q", links[0].Target)
	}
}

// Mutation: regex accepts non-# links (false positives).
func TestMutation_LinkRejectsHTTP(t *testing.T) {
	links := ExtractLinks("[Click](http://example.com)")
	if len(links) != 0 {
		t.Errorf("expected 0 links for HTTP URL, got %d", len(links))
	}
}

// Mutation: regex with wrong anchor group.
func TestMutation_LinkRejectsRelativePaths(t *testing.T) {
	links := ExtractLinks("[File](./other.md)")
	if len(links) != 0 {
		t.Errorf("expected 0 links for relative path, got %d", len(links))
	}
}

// ==================== PageForAnchor mutations ====================

// Mutation: off-by-one in page calculation.
func TestMutation_PageForAnchor_Precision(t *testing.T) {
	// Put a heading exactly at line 20 with page height 10
	lines := make([]string, 25)
	for i := range lines {
		lines[i] = "text"
	}
	lines[20] = "# Exact"

	path := writeTempFileMut(t, "precision.md", strings.Join(lines, "\n"))
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	page := b.PageForAnchor("exact")
	if page != 2 { // line 20 / height 10 = page 2
		t.Errorf("expected page 2, got %d", page)
	}
}

// Mutation: returning 0 instead of -1 for missing anchor.
func TestMutation_PageForAnchor_MissingReturnsNeg1(t *testing.T) {
	path := writeTempFileMut(t, "missing.md", "# Hello\nText\n")
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	result := b.PageForAnchor("nonexistent-anchor")
	if result != -1 {
		t.Errorf("expected -1 for missing anchor, got %d", result)
	}
}

// ==================== Reflow mutations ====================

// Mutation: not updating Pages during reflow.
func TestMutation_ReflowUpdatesPages(t *testing.T) {
	content := strings.Repeat("word ", 100)
	path := writeTempFileMut(t, "reflow.txt", content)
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	pagesBefore := len(b.Pages)
	b.Reflow(20, 5) // much smaller
	pagesAfter := len(b.Pages)
	if pagesAfter <= pagesBefore {
		t.Errorf("expected more pages after shrinking: before=%d, after=%d", pagesBefore, pagesAfter)
	}
}

// Mutation: not updating PageWidth/PageHeight during reflow.
func TestMutation_ReflowUpdatesDimensions(t *testing.T) {
	path := writeTempFileMut(t, "dim.txt", "text\n")
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	b.Reflow(40, 10)
	if b.PageWidth != 40 {
		t.Errorf("expected width 40 after reflow, got %d", b.PageWidth)
	}
	if b.PageHeight != 10 {
		t.Errorf("expected height 10 after reflow, got %d", b.PageHeight)
	}
}

// ==================== Load mutations ====================

// Mutation: not normalizing \r\n to \n.
func TestMutation_LoadNormalizesCRLF(t *testing.T) {
	path := writeTempFileMut(t, "crlf.txt", "A\r\nB\r\nC")
	_, lines, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range lines {
		if strings.ContainsRune(line, '\r') {
			t.Errorf("line %d still contains \\r: %q", i, line)
		}
	}
}

// Mutation: UTF-8 validation removed would accept garbage.
func TestMutation_LoadRejectsInvalidUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.bin")
	os.WriteFile(path, []byte{0xff, 0xfe, 0x80}, 0644)
	_, _, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid UTF-8")
	}
}

func writeTempFileMut(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
