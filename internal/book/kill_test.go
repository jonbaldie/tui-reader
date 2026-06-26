package book

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// ==================== deriveTitle ====================

func TestDeriveTitle_LeadingDotFileKeepsName(t *testing.T) {
	// A leading-dot filename has its dot at index 0; the extension strip must
	// only apply when the dot is *after* index 0 (idx > 0).
	got := deriveTitle("notes/.env")
	if got == "" {
		t.Fatal("deriveTitle stripped a leading-dot filename to empty; idx must be > 0")
	}
	if got != ".env" {
		t.Errorf("deriveTitle(.env) = %q, want %q", got, ".env")
	}
}

func TestDeriveTitle_NormalExtensionStripped(t *testing.T) {
	if got := deriveTitle("dir/my-book.md"); got != "My Book" {
		t.Errorf("deriveTitle = %q, want %q", got, "My Book")
	}
}

// ==================== NormalizeAnchor character boundaries ====================

func TestNormalizeAnchor_KeepsBoundaryAlphanumerics(t *testing.T) {
	// The kept-character ranges are inclusive: 'a'..'z' and '0'..'9'. Boundary
	// runes ('a','z','0','9') must survive normalization.
	if got := NormalizeAnchor("a0 z9"); got != "a0-z9" {
		t.Errorf("NormalizeAnchor(%q) = %q, want %q", "a0 z9", got, "a0-z9")
	}
	if got := NormalizeAnchor("Amazing Zebra 0 to 9"); got != "amazing-zebra-0-to-9" {
		t.Errorf("NormalizeAnchor lost a boundary rune: %q", got)
	}
}

// ==================== Load error wrapping ====================

func TestLoad_MissingFileWrapsOSError(t *testing.T) {
	_, _, err := Load("/no/such/file/at/all.md")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("error must wrap the underlying os error (%%w), got %v", err)
	}
}

// ==================== FormatParagraphs width guard ====================

func TestFormatParagraphs_WidthZeroResetsTo80(t *testing.T) {
	word := strings.Repeat("a", 81)
	out := FormatParagraphs([]string{word}, 0)
	if len(out) != 2 {
		t.Fatalf("width 0 must reset to 80: got %d lines, want 2 (%q)", len(out), out)
	}
	if rl := runeLen(out[0]); rl != 80 {
		t.Errorf("first wrapped line = %d runes, want 80 (width must reset to exactly 80)", rl)
	}
}

func TestFormatParagraphs_WidthOneIsHonored(t *testing.T) {
	// width 1 is valid (>= 1) and must NOT be reset to 80.
	word := strings.Repeat("a", 81)
	out := FormatParagraphs([]string{word}, 1)
	if len(out) != 81 {
		t.Fatalf("width 1 must wrap to 81 single-rune lines, got %d", len(out))
	}
}

// ==================== FormatParagraphs blank handling ====================

func TestFormatParagraphs_TrailingBlankPreserved(t *testing.T) {
	out := FormatParagraphs([]string{"text", ""}, 60)
	if len(out) != 2 || out[len(out)-1] != "" {
		t.Errorf("trailing blank source line must be preserved: got %q", out)
	}
}

func TestFormatParagraphs_NoConsecutiveBlanks(t *testing.T) {
	out := FormatParagraphs([]string{"a", "", "", "b"}, 60)
	for i := 1; i < len(out); i++ {
		if out[i] == "" && out[i-1] == "" {
			t.Errorf("consecutive blank lines at %d in %q (blanks must collapse)", i, out)
		}
	}
}

func TestFormatParagraphs_SecondParagraphIndented(t *testing.T) {
	out := FormatParagraphs([]string{"First paragraph.", "Second paragraph."}, 60)
	var second string
	for _, l := range out {
		if strings.Contains(l, "Second") {
			second = l
			break
		}
	}
	if second == "" {
		t.Fatal("second paragraph not found in output")
	}
	if !strings.HasPrefix(second, "  ") {
		t.Errorf("non-first paragraph must be indented by 2 spaces, got %q", second)
	}
}

// ==================== Paginate page structure ====================

func TestPaginate_EmptyPageHasNonNilSlices(t *testing.T) {
	pages := Paginate(nil, 60, 20)
	if len(pages) != 1 {
		t.Fatalf("empty content must yield 1 page, got %d", len(pages))
	}
	if pages[0].Lines == nil {
		t.Error("empty page must have non-nil Lines slice")
	}
	if pages[0].Links == nil {
		t.Error("empty page must have non-nil Links slice")
	}
}

func TestPaginate_EveryPageHasNonNilLinks(t *testing.T) {
	var raw []string
	for i := 0; i < 30; i++ {
		raw = append(raw, "line of content here")
	}
	pages := Paginate(raw, 80, 20)
	if len(pages) < 2 {
		t.Fatalf("expected multiple pages, got %d", len(pages))
	}
	for pi, p := range pages {
		if p.Links == nil {
			t.Errorf("page %d has nil Links slice", pi)
		}
	}
}

func TestPaginate_HeightZeroResetsTo20(t *testing.T) {
	// Each full page must hold exactly 20 lines; a reset to 19 would shrink them.
	var raw []string
	for i := 0; i < 40; i++ {
		raw = append(raw, "content line")
	}
	pages := Paginate(raw, 80, 0)
	if len(pages) < 2 {
		t.Fatalf("expected >=2 pages, got %d", len(pages))
	}
	if got := len(pages[0].Lines); got != 20 {
		t.Errorf("first page has %d lines, want 20 (height must reset to 20)", got)
	}
}

// ==================== WrapLines ====================

func TestWrapLines_PreservesLeadingSpacesWhenShort(t *testing.T) {
	// A short line must be passed through verbatim, preserving leading spaces
	// (it must not be re-tokenized through wrapLine).
	out := WrapLines([]string{"  indented code"}, 80)
	if len(out) != 1 || out[0] != "  indented code" {
		t.Errorf("short line not preserved verbatim: %q", out)
	}
}

func TestWrapLines_PreservesLeadingSpacesAtExactWidth(t *testing.T) {
	// runeLen == width must still be treated as "fits" (<=, not <).
	line := "  hi" // 4 runes
	out := WrapLines([]string{line}, 4)
	if len(out) != 1 || out[0] != line {
		t.Errorf("line at exact width not preserved: %q", out)
	}
}

func TestWrapLines_DoesNotOverpackWords(t *testing.T) {
	// "aaaa bbbbb" cannot fit in width 9 (would need 10 incl. the space), so it
	// must break between the words, not be joined and hard-split.
	out := WrapLines([]string{"aaaa bbbbb"}, 9)
	want := []string{"aaaa", "bbbbb"}
	if len(out) != len(want) {
		t.Fatalf("got %q, want %q", out, want)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("got %q, want %q", out, want)
		}
	}
}

func TestWrapLines_PacksWordsThatExactlyFit(t *testing.T) {
	// "ab cd" is exactly 5 runes incl. the space; at width 5 it must stay on one
	// line (<= width, not < width).
	out := WrapLines([]string{"ab cd"}, 5)
	if len(out) != 1 || out[0] != "ab cd" {
		t.Errorf("words that exactly fit must pack onto one line, got %q", out)
	}
}

func TestWrapLines_LongWhitespaceOnlyLine(t *testing.T) {
	// A whitespace-only line longer than width reaches wrapLine, whose empty
	// words case must yield a single empty string.
	out := WrapLines([]string{strings.Repeat(" ", 20)}, 5)
	if len(out) != 1 || out[0] != "" {
		t.Errorf("long whitespace line must wrap to a single empty line, got %q", out)
	}
}

func TestWrapLines_HardBreakKeepsAllRunes(t *testing.T) {
	// A single 7-rune word at width 5 must split into "12345" + "67" with no
	// runes lost or duplicated.
	out := WrapLines([]string{"1234567"}, 5)
	want := []string{"12345", "67"}
	if len(out) != len(want) || out[0] != want[0] || out[1] != want[1] {
		t.Errorf("hard break = %q, want %q", out, want)
	}
}

// ==================== NewBook field population ====================

func TestNewBook_PopulatesAllFields(t *testing.T) {
	path := writeTempFile(t, "my-doc.md", "# Title\n\nSome content.\n")
	b, err := NewBook(path, 55, 12)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	if b.PageWidth != 55 {
		t.Errorf("PageWidth = %d, want 55", b.PageWidth)
	}
	if b.PageHeight != 12 {
		t.Errorf("PageHeight = %d, want 12", b.PageHeight)
	}
	if b.Anchors == nil {
		t.Error("Anchors must be populated")
	}
	if len(b.RawLines) == 0 {
		t.Error("RawLines must be populated")
	}
	if len(b.Pages) == 0 {
		t.Error("Pages must be populated")
	}
	if b.Title != "My Doc" {
		t.Errorf("Title = %q, want %q", b.Title, "My Doc")
	}
}

// ==================== PageForAnchor edge cases ====================

func TestPageForAnchor_HeightOneGivesNonZeroPage(t *testing.T) {
	// With PageHeight 1 (which is > 0), an anchor below the first line must map
	// to a page index > 0. A `<= 1` guard would wrongly short-circuit to 0.
	path := writeTempFile(t, "doc.md", "Intro line one.\n\nMore intro text.\n\n# Target Heading\n")
	b, err := NewBook(path, 60, 1)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	got := b.PageForAnchor("target-heading")
	if got <= 0 {
		t.Errorf("PageForAnchor with height 1 = %d, want > 0", got)
	}
}

func TestPageForAnchor_ZeroHeightReturnsZero(t *testing.T) {
	b := &Book{
		Anchors:    map[string]int{"intro": 0},
		RawLines:   []string{"# Intro"},
		PageWidth:  60,
		PageHeight: 0,
	}
	if got := b.PageForAnchor("intro"); got != 0 {
		t.Errorf("PageForAnchor with height 0 = %d, want 0", got)
	}
}

func TestPageForAnchor_AnchorLineMissingFromMapReturnsNeg1(t *testing.T) {
	// Anchor exists in the map but points at a raw line that never appears in
	// the formatted output: the loop falls through to the final return -1.
	b := &Book{
		Anchors:    map[string]int{"ghost": 999},
		RawLines:   []string{"# Real Heading"},
		PageWidth:  60,
		PageHeight: 20,
	}
	if got := b.PageForAnchor("ghost"); got != -1 {
		t.Errorf("PageForAnchor for unreachable anchor line = %d, want -1", got)
	}
}
