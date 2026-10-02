package book

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Test fixtures ---

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// ==================== NewBook loading ====================

func TestNewBook_ValidTextFile(t *testing.T) {
	path := writeTempFile(t, "test.txt", "Hello\nWorld\n")
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", b.Title)
	}
	if len(b.RawLines) != 3 { // "Hello", "World", ""
		t.Errorf("expected 3 lines, got %d", len(b.RawLines))
	}
	if b.RawLines[0] != "Hello" {
		t.Errorf("expected first line 'Hello', got %q", b.RawLines[0])
	}
}

func TestNewBook_ValidMarkdownFile(t *testing.T) {
	content := "# My Book\n\nSome text.\n\n## Chapter 1\n\nMore text.\n"
	path := writeTempFile(t, "my-book.md", content)
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Title != "My Book" {
		t.Errorf("expected title 'My Book', got %q", b.Title)
	}
	if len(b.RawLines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(b.RawLines))
	}
}

func TestRead_EmptyDocument(t *testing.T) {
	b := readBook(t, "", true, 80, 20)
	if len(b.RawLines) != 1 { // Split of "" gives [""]
		t.Errorf("expected 1 line (empty), got %d", len(b.RawLines))
	}
}

func TestRead_WindowsLineEndings(t *testing.T) {
	lines := readBook(t, "Line1\r\nLine2\r\nLine3\r\n", true, 80, 20).RawLines
	if lines[0] != "Line1" {
		t.Errorf("expected 'Line1', got %q", lines[0])
	}
	if lines[1] != "Line2" {
		t.Errorf("expected 'Line2', got %q", lines[1])
	}
}

func TestRead_ClassicMacLineEndings(t *testing.T) {
	lines := readBook(t, "Alpha\rBeta\rGamma\r", true, 80, 20).RawLines
	if lines[0] != "Alpha" {
		t.Errorf("expected 'Alpha', got %q", lines[0])
	}
}

func TestRead_LineEndingEquivalence(t *testing.T) {
	cases := []string{
		"",
		"single line no newline",
		"single line with unix\n",
		"single line with windows\r\n",
		"single line with mac\r",
		"one\ntwo\rthree\r\nfour",
		"trailing\r\n\r\n",
		"trailing\n\n",
		"trailing\r\r",
		"\r\nleading\r\n\r\nmiddle\r\n\r\ntrailing\r\n",
		"mixed\r\nand\nand\rand\r\nend",
	}

	for _, tc := range cases {
		// Canonical reference
		canon := tc
		canon = strings.ReplaceAll(canon, "\r\n", "\n")
		canon = strings.ReplaceAll(canon, "\r", "\n")
		wantLines := strings.Split(canon, "\n")

		gotLines := readBook(t, tc, true, 80, 20).RawLines

		if len(gotLines) != len(wantLines) {
			t.Fatalf("case %q: line count mismatch: got %d, want %d", tc, len(gotLines), len(wantLines))
		}
		for i := range wantLines {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("case %q line %d: got %q, want %q", tc, i, gotLines[i], wantLines[i])
			}
		}
	}
}

// Unhappy paths

func TestNewBook_InvalidUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.bin")
	if err := os.WriteFile(path, []byte{0xff, 0xfe, 0x80, 0x81}, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	_, err := NewBook(path, 80, 20)
	if err == nil {
		t.Fatal("expected error for invalid UTF-8, got nil")
	}
}

// ==================== deriveTitle ====================

func TestDeriveTitle_Simple(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"my-book.md", "My Book"},
		{"chapter_one.txt", "Chapter One"},
		{"/some/path/great-gatsby.epub", "Great Gatsby"},
		{"simple", "Simple"},
		{"a-b-c.md", "A B C"},
	}
	for _, tt := range tests {
		got := deriveTitle(tt.path)
		if got != tt.want {
			t.Errorf("deriveTitle(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// ==================== WrapLines ====================

func TestWrapLines_ShortLines(t *testing.T) {
	lines := []string{"hello", "world"}
	result := WrapLines(lines, 80)
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d", len(result))
	}
}

func TestWrapLines_LongLine(t *testing.T) {
	line := strings.Repeat("word ", 20) // 100 chars
	result := WrapLines([]string{line}, 40)
	if len(result) < 2 {
		t.Errorf("expected line to be wrapped, got %d lines", len(result))
	}
	for _, l := range result {
		if runeLen(l) > 40 {
			t.Errorf("wrapped line exceeds width: %q (%d runes)", l, runeLen(l))
		}
	}
}

func TestWrapLines_VeryLongWord(t *testing.T) {
	word := strings.Repeat("x", 100)
	result := WrapLines([]string{word}, 30)
	if len(result) < 2 {
		t.Errorf("expected long word to be hard-broken, got %d lines", len(result))
	}
	for _, l := range result {
		if runeLen(l) > 30 {
			t.Errorf("hard-broken line exceeds width: %q", l)
		}
	}
}

func TestWrapLines_EmptyLine(t *testing.T) {
	result := WrapLines([]string{""}, 80)
	if len(result) != 1 || result[0] != "" {
		t.Errorf("expected single empty line, got %v", result)
	}
}

func TestWrapLines_UnicodeContent(t *testing.T) {
	// Japanese text - each char is 1 rune
	line := "こんにちは世界テスト文字列"
	result := WrapLines([]string{line}, 5)
	// The line is 12 runes, wrapping at 5 should produce multiple lines
	if len(result) < 2 {
		t.Errorf("expected unicode wrapping, got %d lines", len(result))
	}
}

// ==================== Pagination ====================

func TestRead_PaginateBasic(t *testing.T) {
	// 50 paragraphs separated by blank lines = 50 content + 49 blanks = 99 lines
	// At height 10 = 10 pages
	lines := make([]string, 99)
	for i := range lines {
		if i%2 == 0 {
			lines[i] = "line"
		}
	}
	pages := readMarkdown(t, lines, 80, 10).Pages
	if len(pages) != 10 {
		t.Errorf("expected 10 pages, got %d", len(pages))
	}
}

func TestRead_PaginatePartialLastPage(t *testing.T) {
	// 15 paragraphs separated by blank lines = 15 content + 14 blanks = 29 formatted lines
	// at height 10 = 3 pages (10, 10, 9)
	lines := make([]string, 29)
	for i := range lines {
		if i%2 == 0 {
			lines[i] = "line"
		}
	}
	pages := readMarkdown(t, lines, 80, 10).Pages
	if len(pages) != 3 {
		t.Errorf("expected 3 pages, got %d", len(pages))
	}
	lastPage := pages[len(pages)-1]
	if len(lastPage.Lines) != 9 {
		t.Errorf("expected last page to have 9 lines, got %d", len(lastPage.Lines))
	}
}

func TestRead_PaginateEmptyContent(t *testing.T) {
	pages := readMarkdown(t, []string{}, 80, 10).Pages
	if len(pages) != 1 {
		t.Errorf("expected 1 empty page, got %d", len(pages))
	}
}

func TestRead_PaginateSingleLine(t *testing.T) {
	pages := readMarkdown(t, []string{"hello"}, 80, 10).Pages
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
	if pages[0].Lines[0] != "hello" {
		t.Errorf("expected 'hello', got %q", pages[0].Lines[0])
	}
}

func TestRead_PaginateInvalidDimensions(t *testing.T) {
	lines := []string{"test"}
	// Should fall back to defaults
	pages := readMarkdown(t, lines, 0, 0).Pages
	if len(pages) == 0 {
		t.Error("expected at least 1 page with zero dimensions")
	}
}

func TestRead_PaginateExactFit(t *testing.T) {
	// A single raw line should produce exactly 1 page if it fits
	pages := readMarkdown(t, []string{"single paragraph"}, 80, 10).Pages
	if len(pages) != 1 {
		t.Errorf("expected exactly 1 page, got %d", len(pages))
	}
}

// ==================== NewBook ====================

func TestNewBook_ValidFile(t *testing.T) {
	content := "# Title\n\nParagraph one.\n\nParagraph two.\n"
	path := writeTempFile(t, "book.md", content)
	b, err := NewBook(path, 60, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Title != "Book" {
		t.Errorf("expected title 'Book', got %q", b.Title)
	}
	if len(b.Pages) == 0 {
		t.Error("expected at least 1 page")
	}
	if len(b.Anchors) == 0 {
		t.Error("expected at least 1 anchor from heading")
	}
}

func TestNewBook_MissingFile(t *testing.T) {
	_, err := NewBook("/nonexistent.txt", 60, 10)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ==================== Reflow ====================

func TestReflow_ChangeDimensions(t *testing.T) {
	content := strings.Repeat("word ", 100) // long text
	path := writeTempFile(t, "reflow.txt", content)
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	origPages := len(b.Pages)

	b.Reflow(40, 10) // smaller view
	if len(b.Pages) <= origPages {
		t.Error("expected more pages after reducing dimensions")
	}
}

// ==================== PageForAnchor ====================

func TestPageForAnchor_Found(t *testing.T) {
	// Create a document with a heading after some filler
	var lines []string
	for i := 0; i < 30; i++ {
		lines = append(lines, "filler line", "")
	}
	lines = append(lines, "# Target Heading")
	for i := 0; i < 10; i++ {
		lines = append(lines, "more text")
	}

	content := strings.Join(lines, "\n")
	path := writeTempFile(t, "anchor.md", content)
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	page := b.PageForAnchor("target-heading")
	if page < 0 {
		t.Fatal("expected to find anchor 'target-heading'")
	}
	// With paragraph spacing, 30 filler lines become 30 + 29 spacers = 59 formatted lines
	// Heading is at formatted line 59 (after a spacer). At height 10, that's page 5 or 6
	if page < 5 {
		t.Errorf("expected page >= 5, got %d", page)
	}
}

func TestPageForAnchor_NotFound(t *testing.T) {
	content := "# Heading\n\nText\n"
	path := writeTempFile(t, "noanchor.md", content)
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	page := b.PageForAnchor("nonexistent")
	if page != -1 {
		t.Errorf("expected -1, got %d", page)
	}
}

func TestPageForAnchor_PageShowsHeading(t *testing.T) {
	rawLines := []string{
		"# Introduction",
		"",
		"See [Chapter 1](#chapter-1) or [Chapter 2](#chapter-2).",
		"",
		"# Chapter 1",
		"",
		"Some long paragraph that will take up several display lines to ensure pagination happens properly across multiple pages.",
		"",
		"# Chapter 2",
		"",
		"Content of chapter 2.",
	}
	b := readMarkdown(t, rawLines, 40, 4)

	for anchor, line := range b.Anchors {
		page := b.PageForAnchor(anchor)
		if page < 0 {
			t.Fatalf("expected valid page for %q, got %d", anchor, page)
		}
		found := false
		for _, l := range b.Pages[page].Lines {
			if l == rawLines[line] {
				found = true
			}
		}
		if !found {
			t.Fatalf("anchor %q page %d = %q, want heading %q", anchor, page, b.Pages[page].Lines, rawLines[line])
		}
	}

	// Non-existent anchor
	if b.PageForAnchor("non-existent") != -1 {
		t.Fatalf("expected -1 for non-existent anchor")
	}
}
