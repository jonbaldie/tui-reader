package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// ==================== parseArgs ====================

func TestParseArgs_PlainPath(t *testing.T) {
	path, dumpMode, dumpPages := parseArgs([]string{"book.md"})
	if path != "book.md" {
		t.Errorf("path = %q, want %q", path, "book.md")
	}
	if dumpMode {
		t.Error("dumpMode = true, want false")
	}
	if dumpPages != 0 {
		t.Errorf("dumpPages = %d, want 0", dumpPages)
	}
}

func TestParseArgs_DumpFlag(t *testing.T) {
	path, dumpMode, dumpPages := parseArgs([]string{"--dump", "book.md"})
	if path != "book.md" {
		t.Errorf("path = %q, want %q", path, "book.md")
	}
	if !dumpMode {
		t.Error("dumpMode = false, want true")
	}
	if dumpPages != 0 {
		t.Errorf("dumpPages = %d, want 0", dumpPages)
	}
}

func TestParseArgs_DumpWithCount(t *testing.T) {
	path, dumpMode, dumpPages := parseArgs([]string{"--dump=3", "book.md"})
	if path != "book.md" {
		t.Errorf("path = %q, want %q", path, "book.md")
	}
	if !dumpMode {
		t.Error("dumpMode = false, want true")
	}
	if dumpPages != 3 {
		t.Errorf("dumpPages = %d, want 3", dumpPages)
	}
}

func TestParseArgs_OnlyFlagNoPath(t *testing.T) {
	path, dumpMode, _ := parseArgs([]string{"--dump"})
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
	if !dumpMode {
		t.Error("dumpMode = false, want true")
	}
}

func TestParseArgs_OrderIndependent(t *testing.T) {
	path, dumpMode, dumpPages := parseArgs([]string{"book.md", "--dump=5"})
	if path != "book.md" {
		t.Errorf("path = %q, want %q", path, "book.md")
	}
	if !dumpMode || dumpPages != 5 {
		t.Errorf("dumpMode=%v dumpPages=%d, want true 5", dumpMode, dumpPages)
	}
}

func TestParseArgs_NoDumpKeepsPagesZero(t *testing.T) {
	// Guards against dumpPages being parsed from a plain path.
	_, _, dumpPages := parseArgs([]string{"chapter5.md"})
	if dumpPages != 0 {
		t.Errorf("dumpPages = %d, want 0", dumpPages)
	}
}

// ==================== renderDump ====================

// dumpBook builds a small multi-page book for dump tests.
func dumpBook(t *testing.T) *book.Book {
	t.Helper()
	var sb strings.Builder
	// 200 distinct non-blank lines -> many pages at height 20.
	for i := 0; i < 200; i++ {
		sb.WriteString("Line number ")
		sb.WriteByte(byte('A' + i%26))
		sb.WriteString("\n")
	}
	path := writeTempFile(t, "my-book.md", sb.String())
	b, err := book.NewBook(path, 62, 20)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	return b
}

func countPageBlocks(out string) int {
	return strings.Count(out, "── Page ")
}

func TestRenderDump_AllPages(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 0)
	if got := countPageBlocks(out); got != len(b.Pages) {
		t.Errorf("rendered %d page blocks, want %d", got, len(b.Pages))
	}
}

func TestRenderDump_MaxPagesTruncates(t *testing.T) {
	b := dumpBook(t)
	if len(b.Pages) < 3 {
		t.Fatalf("need >=3 pages, got %d", len(b.Pages))
	}
	out := renderDump(b, 2)
	if got := countPageBlocks(out); got != 2 {
		t.Errorf("rendered %d page blocks, want 2", got)
	}
}

func TestRenderDump_MaxPagesZeroMeansAll(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 0)
	if got := countPageBlocks(out); got != len(b.Pages) {
		t.Errorf("with 0 rendered %d blocks, want all %d", got, len(b.Pages))
	}
}

func TestRenderDump_MaxPagesLargerThanTotal(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, len(b.Pages)+100)
	if got := countPageBlocks(out); got != len(b.Pages) {
		t.Errorf("rendered %d blocks, want %d", got, len(b.Pages))
	}
}

func TestRenderDump_PageNumbersAreOneBased(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 2)
	if !strings.Contains(out, "Page 1 of") {
		t.Error("expected 'Page 1 of' in output (1-based numbering)")
	}
	if !strings.Contains(out, "Page 2 of") {
		t.Error("expected 'Page 2 of' in output")
	}
	if strings.Contains(out, "Page 0 of") {
		t.Error("page numbering must not be 0-based")
	}
	if strings.Contains(out, "Page 3 of") {
		t.Error("must not render a third page when maxPages=2")
	}
}

func TestRenderDump_TotalShowsFullCount(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 1)
	// Even when truncated, "of N" must reflect the real total page count.
	want := "of " + itoa(len(b.Pages))
	if !strings.Contains(out, want) {
		t.Errorf("expected %q in header, got:\n%s", want, firstLine(out))
	}
}

func TestRenderDump_TitleInHeader(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 1)
	if !strings.Contains(out, b.Title) {
		t.Errorf("header missing title %q", b.Title)
	}
	if b.Title != "My Book" {
		t.Errorf("title = %q, want %q", b.Title, "My Book")
	}
}

func TestRenderDump_ContainsPageContent(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 1)
	// First content line of page 0 should appear, prefixed by the gutter.
	first := b.Pages[0].Lines[0]
	if !strings.Contains(out, "│  "+first) {
		t.Errorf("expected gutter-prefixed content line %q in output", first)
	}
}

func TestRenderDump_PadsToTwentyLines(t *testing.T) {
	// A book whose last page has fewer than 20 lines must be padded so each
	// page block spans a fixed height.
	path := writeTempFile(t, "short.md", "only one line\n")
	b, err := book.NewBook(path, 62, 20)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	out := renderDump(b, 0)
	// Count bare gutter lines "│" (no following content).
	bareGutters := 0
	for _, ln := range strings.Split(out, "\n") {
		if ln == "│" {
			bareGutters++
		}
	}
	// 1 leading "│" + (20 - contentLines) padding + 1 trailing "│".
	contentLines := len(b.Pages[0].Lines)
	want := 1 + (20 - contentLines) + 1
	if bareGutters != want {
		t.Errorf("bare gutter lines = %d, want %d (pad to 20, content=%d)", bareGutters, want, contentLines)
	}
}

func TestRenderDump_BlankLineBetweenPagesOnly(t *testing.T) {
	b := dumpBook(t)
	out := renderDump(b, 3)
	// There must be exactly (pages-1) inter-page blank separators: a blank
	// line immediately following a page's closing border.
	border := "└────────────────────────────────┘"
	segments := strings.Split(out, border)
	// segments: [before p1][between p1,p2][between p2,p3][after p3]
	// The first three borders each terminate a page; a separator blank line
	// follows all but the last.
	separators := 0
	for i := 1; i < len(segments)-1; i++ {
		if strings.HasPrefix(segments[i], "\n\n") {
			separators++
		}
	}
	if separators != 2 {
		t.Errorf("inter-page separators = %d, want 2", separators)
	}
	// The final page must not be followed by a trailing separator blank line.
	last := segments[len(segments)-1]
	if last == "\n\n" {
		t.Error("unexpected trailing blank line after final page")
	}
}

func TestRenderDump_EmptyWhenMaxPagesNegativelyUnreachable(t *testing.T) {
	// maxPages must only shrink total, never grow it; with maxPages <= 0 we
	// render all pages (covered above) — here confirm a single-page render.
	path := writeTempFile(t, "one.md", "hello\n")
	b, err := book.NewBook(path, 62, 20)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	out := renderDump(b, 0)
	if got := countPageBlocks(out); got != 1 {
		t.Errorf("page blocks = %d, want 1", got)
	}
}

// ==================== run ====================

func TestRun_NoArgs(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(nil, &out, &errBuf)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "Usage:") {
		t.Errorf("expected usage on stderr, got %q", errBuf.String())
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout, got %q", out.String())
	}
}

func TestRun_OnlyFlagNoPath(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"--dump"}, &out, &errBuf)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "Usage:") {
		t.Errorf("expected usage on stderr, got %q", errBuf.String())
	}
}

func TestRun_DumpValidFile(t *testing.T) {
	path := writeTempFile(t, "doc.md", "# Title\n\nSome content here.\n")
	var out, errBuf bytes.Buffer
	code := run([]string{"--dump", path}, &out, &errBuf)
	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr=%q", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Page 1 of") {
		t.Errorf("expected dumped page on stdout, got %q", out.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected no stderr, got %q", errBuf.String())
	}
}

func TestRun_DumpMissingFile(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"--dump", "/no/such/file/here.md"}, &out, &errBuf)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "Error:") {
		t.Errorf("expected error on stderr, got %q", errBuf.String())
	}
	if out.Len() != 0 {
		t.Errorf("expected no stdout on error, got %q", out.String())
	}
}

func TestRun_DumpWithCountLimitsOutput(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString("Distinct line ")
		sb.WriteByte(byte('A' + i%26))
		sb.WriteString("\n")
	}
	path := writeTempFile(t, "long.md", sb.String())
	var out, errBuf bytes.Buffer
	code := run([]string{"--dump=1", path}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if strings.Contains(out.String(), "Page 2 of") {
		t.Error("--dump=1 should render only one page")
	}
	if !strings.Contains(out.String(), "Page 1 of") {
		t.Error("expected page 1 in output")
	}
}

// small helpers to avoid importing strconv in assertions above
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
