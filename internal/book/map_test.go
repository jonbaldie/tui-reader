package book

import (
	"strings"
	"testing"
)

// checkMapInvariants verifies the structural contract of BuildFormattedToRawMap
// against the public FormatParagraphs output, without duplicating its internal
// logic. These invariants are sensitive to off-by-one, sign, and wrap-width
// mutations in the mapping walk.
func checkMapInvariants(t *testing.T, rawLines []string, width int) []int {
	t.Helper()
	formatted := FormatParagraphs(rawLines, width)
	m := BuildFormattedToRawMap(rawLines, width)

	// INV1: one map entry per formatted line.
	if len(m) != len(formatted) {
		t.Fatalf("map len = %d, want %d (one per formatted line)", len(m), len(formatted))
	}

	prevRaw := -1
	for fi, ri := range m {
		// INV2: raw indices, when present, never move backwards.
		if ri >= 0 {
			if ri < prevRaw {
				t.Errorf("fi=%d maps to raw %d which is before previous %d (must be monotonic)", fi, ri, prevRaw)
			}
			if ri >= len(rawLines) {
				t.Errorf("fi=%d maps to raw %d out of range (len %d)", fi, ri, len(rawLines))
			}
			prevRaw = ri
		}

		if formatted[fi] != "" {
			// INV3: a non-blank formatted line must map to a non-blank raw line.
			if ri < 0 {
				t.Errorf("non-blank formatted line %d (%q) maps to -1; expected a raw line", fi, formatted[fi])
			} else if strings.TrimSpace(rawLines[ri]) == "" {
				t.Errorf("non-blank formatted line %d (%q) maps to blank raw line %d", fi, formatted[fi], ri)
			}
		} else {
			// INV4: a blank formatted line is either an inserted spacer (-1)
			// or a preserved blank source line.
			if ri >= 0 && strings.TrimSpace(rawLines[ri]) != "" {
				t.Errorf("blank formatted line %d maps to non-blank raw line %d (%q)", fi, ri, rawLines[ri])
			}
		}
	}
	return m
}

func TestBuildMap_ExactMapping(t *testing.T) {
	rawLines := []string{"# Heading", "", "First para.", "Second para."}
	got := checkMapInvariants(t, rawLines, 60)
	want := []int{0, 1, 2, -1, 3}
	if len(got) != len(want) {
		t.Fatalf("map = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("map = %v, want %v (index %d differs)", got, want, i)
		}
	}
}

func TestBuildMap_SpacerBeforeSecondParagraph(t *testing.T) {
	// Two adjacent paragraphs: FormatParagraphs inserts a blank spacer that
	// must map to -1 (not to a raw line).
	rawLines := []string{"Alpha line one.", "Beta line two."}
	got := checkMapInvariants(t, rawLines, 60)
	want := []int{0, -1, 1}
	if len(got) != len(want) {
		t.Fatalf("map = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("map = %v, want %v", got, want)
		}
	}
}

func TestBuildMap_FirstParagraphNotIndented(t *testing.T) {
	// A 60-char first "word" fits at width 60 (no indent) but would wrap at the
	// indented width 58. If the walk treated the first paragraph as non-first,
	// the second paragraph's mapping would shift. This pins firstParagraph=true.
	word := strings.Repeat("a", 60)
	rawLines := []string{word, "Second"}
	got := checkMapInvariants(t, rawLines, 60)
	want := []int{0, -1, 1}
	if len(got) != len(want) {
		t.Fatalf("map = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("map = %v, want %v", got, want)
		}
	}
}

func TestBuildMap_BlankSourceLinesPreserved(t *testing.T) {
	rawLines := []string{"Intro line.", "", "", "Body line."}
	checkMapInvariants(t, rawLines, 60)
}

func TestBuildMap_WrappingNonFirstParagraph(t *testing.T) {
	// A non-first paragraph long enough to wrap at the indented width. The map
	// must allocate one entry per wrapped formatted line, all pointing at the
	// same raw index. Catches wrong wrap-width in the walk.
	long := strings.Repeat("word ", 40) // ~200 chars, wraps to several lines
	rawLines := []string{"Short intro.", long, "Tail."}
	checkMapInvariants(t, rawLines, 40)
}

func TestBuildMap_TinyWidthClampedTo10(t *testing.T) {
	// width-2 = 9 < 10, so the indented wrap width is clamped to 10. Exercises
	// the clamp branch in the mapping walk.
	long := strings.Repeat("ab ", 20)
	rawLines := []string{"Intro.", long}
	checkMapInvariants(t, rawLines, 11)
}

func TestBuildMap_HeadingsAndCode(t *testing.T) {
	rawLines := []string{
		"# Title",
		"",
		"Normal paragraph here.",
		"",
		"## Section",
		"    indented code block",
		"Another paragraph.",
	}
	checkMapInvariants(t, rawLines, 50)
}

func TestBuildMap_LongNonFirstHeading(t *testing.T) {
	// A long heading that is NOT the first paragraph: headings are "special",
	// so they are wrapped at full width (no indent) in both FormatParagraphs and
	// the mapping walk. If the walk misclassified it as a normal paragraph it
	// would wrap at width-2 and the formatted/raw alignment would break.
	heading := "# " + strings.Repeat("word ", 15)
	rawLines := []string{"Intro paragraph.", heading, "Closing paragraph."}
	checkMapInvariants(t, rawLines, 40)
}

func TestBuildMap_LongNonFirstCodeBlock(t *testing.T) {
	code := "    " + strings.Repeat("token ", 15)
	rawLines := []string{"Intro paragraph.", code, "Closing paragraph."}
	checkMapInvariants(t, rawLines, 40)
}

func TestBuildMap_EmptyInput(t *testing.T) {
	m := BuildFormattedToRawMap([]string{}, 60)
	if len(m) != 0 {
		t.Errorf("empty input map len = %d, want 0", len(m))
	}
}

// ==================== AttachLinks ====================

func TestAttachLinks_LinkOnSecondPage(t *testing.T) {
	// Build content so the link lands on page index 1. This pins
	// startLine = pi * height (a divide or off-by-one breaks placement).
	var raw []string
	for i := 0; i < 8; i++ {
		raw = append(raw, "Filler paragraph line.")
		raw = append(raw, "")
	}
	raw = append(raw, "Visit [Chapter 2](#chapter-2) now.")

	width, height := 80, 5
	pages := Paginate(raw, width, height)
	pages = AttachLinks(pages, raw, width, height)

	if len(pages) < 2 {
		t.Fatalf("expected >=2 pages, got %d", len(pages))
	}

	// Exactly one page should carry the link, and it must be the last page.
	pagesWithLink := 0
	linkPage := -1
	for pi, p := range pages {
		if len(p.Links) > 0 {
			pagesWithLink++
			linkPage = pi
		}
	}
	if pagesWithLink != 1 {
		t.Fatalf("expected exactly 1 page with links, got %d", pagesWithLink)
	}
	if linkPage != len(pages)-1 {
		t.Errorf("link on page %d, want last page %d", linkPage, len(pages)-1)
	}
	if pages[linkPage].Links[0].Target != "chapter-2" {
		t.Errorf("target = %q, want chapter-2", pages[linkPage].Links[0].Target)
	}
	// LineOnPage must be the local index within that page, not a global index.
	lop := pages[linkPage].Links[0].LineOnPage
	if lop < 0 || lop >= len(pages[linkPage].Lines) {
		t.Errorf("LineOnPage = %d, out of range for page of %d lines", lop, len(pages[linkPage].Lines))
	}
	if pages[linkPage].Lines[lop] == "" || !strings.Contains(pages[linkPage].Lines[lop], "Chapter 2") {
		t.Errorf("LineOnPage %d does not point at the link line: %q", lop, pages[linkPage].Lines[lop])
	}
}

func TestAttachLinks_DeduplicatesSameLink(t *testing.T) {
	// A single link on a single line must be attached exactly once, even though
	// AttachLinks scans both the raw line and the formatted line.
	raw := []string{"See [Chapter 1](#chapter-1) here."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)
	if len(pages[0].Links) != 1 {
		t.Fatalf("expected exactly 1 link (deduplicated), got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
}

func TestAttachLinks_TwoDistinctLinksSameLine(t *testing.T) {
	raw := []string{"[One](#one) and [Two](#two)."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)
	if len(pages[0].Links) != 2 {
		t.Fatalf("expected 2 distinct links, got %d", len(pages[0].Links))
	}
	targets := map[string]bool{}
	for _, l := range pages[0].Links {
		targets[l.Target] = true
	}
	if !targets["one"] || !targets["two"] {
		t.Errorf("expected targets one+two, got %v", targets)
	}
}

func TestAttachLinks_NoLinksLeavesEmpty(t *testing.T) {
	raw := []string{"Just some text.", "More text."}
	pages := Paginate(raw, 80, 20)
	pages = AttachLinks(pages, raw, 80, 20)
	for pi, p := range pages {
		if len(p.Links) != 0 {
			t.Errorf("page %d has %d links, want 0", pi, len(p.Links))
		}
	}
}
