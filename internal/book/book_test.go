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

// ==================== Load ====================

func TestLoad_ValidTextFile(t *testing.T) {
	path := writeTempFile(t, "test.txt", "Hello\nWorld\n")
	title, lines, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Test" {
		t.Errorf("expected title 'Test', got %q", title)
	}
	if len(lines) != 3 { // "Hello", "World", ""
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "Hello" {
		t.Errorf("expected first line 'Hello', got %q", lines[0])
	}
}

func TestLoad_ValidMarkdownFile(t *testing.T) {
	content := "# My Book\n\nSome text.\n\n## Chapter 1\n\nMore text.\n"
	path := writeTempFile(t, "my-book.md", content)
	title, lines, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "My Book" {
		t.Errorf("expected title 'My Book', got %q", title)
	}
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := writeTempFile(t, "empty.txt", "")
	_, lines, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 { // Split of "" gives [""]
		t.Errorf("expected 1 line (empty), got %d", len(lines))
	}
}

func TestLoad_WindowsLineEndings(t *testing.T) {
	path := writeTempFile(t, "crlf.txt", "Line1\r\nLine2\r\nLine3\r\n")
	_, lines, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lines[0] != "Line1" {
		t.Errorf("expected 'Line1', got %q", lines[0])
	}
	if lines[1] != "Line2" {
		t.Errorf("expected 'Line2', got %q", lines[1])
	}
}

func TestLoad_ClassicMacLineEndings(t *testing.T) {
	path := writeTempFile(t, "cr.txt", "Alpha\rBeta\rGamma\r")
	_, lines, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lines[0] != "Alpha" {
		t.Errorf("expected 'Alpha', got %q", lines[0])
	}
}

// Unhappy paths

func TestLoad_MissingFile(t *testing.T) {
	_, _, err := Load("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.bin")
	if err := os.WriteFile(path, []byte{0xff, 0xfe, 0x80, 0x81}, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	_, _, err := Load(path)
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

// ==================== Paginate ====================

func TestPaginate_Basic(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	pages := Paginate(lines, 80, 10)
	if len(pages) != 5 {
		t.Errorf("expected 5 pages, got %d", len(pages))
	}
}

func TestPaginate_PartialLastPage(t *testing.T) {
	lines := make([]string, 15)
	for i := range lines {
		lines[i] = "line"
	}
	pages := Paginate(lines, 80, 10)
	if len(pages) != 2 {
		t.Errorf("expected 2 pages, got %d", len(pages))
	}
	if len(pages[1].Lines) != 5 {
		t.Errorf("expected last page to have 5 lines, got %d", len(pages[1].Lines))
	}
}

func TestPaginate_EmptyContent(t *testing.T) {
	pages := Paginate([]string{}, 80, 10)
	if len(pages) != 1 {
		t.Errorf("expected 1 empty page, got %d", len(pages))
	}
}

func TestPaginate_SingleLine(t *testing.T) {
	pages := Paginate([]string{"hello"}, 80, 10)
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
	if pages[0].Lines[0] != "hello" {
		t.Errorf("expected 'hello', got %q", pages[0].Lines[0])
	}
}

func TestPaginate_InvalidDimensions(t *testing.T) {
	lines := []string{"test"}
	// Should fall back to defaults
	pages := Paginate(lines, 0, 0)
	if len(pages) == 0 {
		t.Error("expected at least 1 page with zero dimensions")
	}
}

func TestPaginate_ExactFit(t *testing.T) {
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	pages := Paginate(lines, 80, 10)
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
	// Create a document with a heading on a known line
	var lines []string
	for i := 0; i < 30; i++ {
		lines = append(lines, "filler line")
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
	if page != 3 { // line 30, with height 10 -> page 3
		t.Errorf("expected page 3, got %d", page)
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
