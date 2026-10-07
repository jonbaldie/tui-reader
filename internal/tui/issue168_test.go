package tui

import (
	"strconv"
	"strings"
	"testing"
)

func issue168Chapter(sentences int) string {
	var s []string
	for i := 0; i < sentences; i++ {
		s = append(s, "sentence "+strconv.Itoa(i)+" of the chapter.")
	}
	return "# Chapter\n\n" + strings.Join(s, " ") + "\n"
}

func TestIssue168_ResizeKeepsPositionInsideMultiPageParagraph(t *testing.T) {
	m := newTestModel(t, "chapter.md", issue168Chapter(5))
	m = applyWindowSize(m, 48, 12)
	if n := len(m.BookRef().Pages); n != 2 {
		t.Fatalf("setup: %d pages, want 2", n)
	}
	m = pressKey(m, "right")
	if !strings.Contains(pageText(m), "the chapter.") || strings.Contains(pageText(m), "# Chapter") {
		t.Fatalf("setup: page 2 = %q", pageText(m))
	}

	m = applyWindowSize(m, 47, 12)
	if got := pageText(m); strings.Contains(got, "# Chapter") || !strings.Contains(got, "the chapter.") {
		t.Fatalf("resize jumped to page %d: %q", m.CurrentPage(), got)
	}
}

func TestIssue168_ResizeKeepsPositionInLongParagraph(t *testing.T) {
	m := newTestModel(t, "chapter.md", issue168Chapter(80))
	m = applyWindowSize(m, 48, 12)
	for i := 0; i < 3; i++ {
		m = pressKey(m, "right")
	}
	before := strings.Fields(pageText(m))
	// A word sequence from mid-page that must still be visible.
	mid := strings.Join(before[len(before)/2:len(before)/2+4], " ")

	m = applyWindowSize(m, 47, 12)
	after := strings.Join(strings.Fields(pageText(m)), " ")
	if !strings.Contains(after, mid) {
		t.Fatalf("resize from page 4 lost position: now page %d showing %q, want it to contain %q", m.CurrentPage(), after, mid)
	}
}
