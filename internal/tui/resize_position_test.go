package tui

import (
	"os"
	"strings"
	"testing"
)

// resizePositionFixturesDir is the checked-in exploratory-testing fixture set
// used to reproduce issue #61.
const resizePositionFixturesDir = "../../docs/exploratory-testing/2026-09-09/evidence/fixtures"

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(resizePositionFixturesDir + "/" + name)
	if err != nil {
		t.Fatalf("fixture missing: %v", err)
	}
	return string(data)
}

// pageText returns the display text of the model's current page.
func pageText(m Model) string {
	b := m.BookRef()
	if b == nil || m.CurrentPage() >= len(b.Pages) {
		return ""
	}
	return strings.Join(b.Pages[m.CurrentPage()].Lines, " ")
}

// TestResize_KeepsReadingPosition reproduces issue #61: resize must preserve
// the reader's position by source location, not page index.
func TestResize_KeepsReadingPosition(t *testing.T) {
	path := writeTempFile(t, "reading-journey.md", loadFixture(t, "reading-journey.md"))
	m := NewModel(path)
	m = applyWindowSize(m, 90, 24) // wide: 72x17 content

	// Jump to the end of the document.
	m = pressKey(m, "G")
	before := pageText(m)
	if !strings.Contains(before, "The journey ends") {
		t.Fatalf("setup: expected last page to show the ending paragraph, got page %d: %q", m.CurrentPage(), before)
	}

	// Resize to a narrow terminal (40x12 → 36x5 content). Page-level
	// granularity: the reader resumes at the paragraph the wide page
	// started with, one page short of the very end.
	m = applyWindowSize(m, 40, 12)
	after := pageText(m)
	if !strings.Contains(after, "Page-marker paragraph 22") {
		t.Errorf("resize lost reading position: was showing the ending paragraphs, now on page %d showing %q",
			m.CurrentPage(), firstNonEmpty(after))
	}
}

// TestResize_HistoryKeepsLinkSource reproduces the link-history path of
// issue #61: follow a link, resize, go back — `b` must return to the page
// containing RETURN MARKER.
func TestResize_HistoryKeepsLinkSource(t *testing.T) {
	path := writeTempFile(t, "link-history-journey.md", loadFixture(t, "link-history-journey.md"))
	m := NewModel(path)
	m = applyWindowSize(m, 90, 24) // wide: 72x17 content

	// Advance to the page containing RETURN MARKER and its link.
	for i := 0; i < 10 && !strings.Contains(pageText(m), "RETURN MARKER"); i++ {
		m = pressKey(m, " ")
	}
	if !strings.Contains(pageText(m), "RETURN MARKER") {
		t.Fatal("setup: RETURN MARKER page not reached")
	}

	// Select the link and follow it.
	m = pressKey(m, "tab")
	if m.SelectedLink() < 0 {
		t.Fatal("setup: no link selected on RETURN MARKER page")
	}
	m = pressKey(m, "enter")
	if !strings.Contains(pageText(m), "The Ending") {
		t.Fatalf("setup: expected to land on The Ending, got page %d: %q", m.CurrentPage(), pageText(m))
	}

	// Resize to a narrow terminal, then go back.
	m = applyWindowSize(m, 40, 12)
	m = pressKey(m, "b")

	if !strings.Contains(pageText(m), "RETURN MARKER") {
		t.Errorf("back-history lost source position after resize: on page %d showing %q",
			m.CurrentPage(), firstNonEmpty(pageText(m)))
	}
}

func firstNonEmpty(s string) string {
	for line := range strings.SplitSeq(s, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return "<empty>"
}

// TestResize_HistoryRemappedNotClamped covers the acceptance criterion that
// every back-history entry is remapped through its source position — not
// clamped — when a resize increases the page count.
func TestResize_HistoryRemappedNotClamped(t *testing.T) {
	path := writeTempFile(t, "stateful.md", statefulDocument())
	m := NewModel(path)
	m = applyWindowSize(m, 90, 24) // wide: fewer pages

	// Chain through three chapters; every landed page carries a
	// "[Next chapter]" link, so tab+enter follows it again.
	for follows := 0; follows < 3; follows++ {
		m = pressKey(m, "tab")
		m = pressKey(m, "enter")
	}
	if len(m.History()) != 3 {
		t.Fatalf("setup: expected 3 history entries, got %v", m.History())
	}

	// Snapshot the heading each history entry's wide page ends with.
	heading := func(page int) string {
		for _, line := range m.BookRef().Pages[page].Lines {
			if strings.HasPrefix(line, "# ") {
				return line
			}
		}
		return ""
	}
	want := make([]string, len(m.History()))
	for i, page := range m.History() {
		want[i] = heading(page)
		if want[i] == "" {
			t.Fatalf("setup: history page %d has no heading", page)
		}
	}

	// Resize to a narrow terminal: the page count grows.
	widePages := len(m.BookRef().Pages)
	m = applyWindowSize(m, 40, 12)
	if len(m.BookRef().Pages) <= widePages {
		t.Fatalf("setup: expected page count to grow, %d -> %d", widePages, len(m.BookRef().Pages))
	}

	for i, page := range m.History() {
		if page < 0 || page >= len(m.BookRef().Pages) {
			t.Fatalf("history entry %d (%d) outside the new page range", i, page)
		}
		if got := strings.Join(m.BookRef().Pages[page].Lines, " "); !strings.Contains(got, want[i]) {
			t.Errorf("history entry %d was clamped, not remapped: expected page %d to contain %q, got %q",
				i, page, want[i], firstNonEmpty(got))
		}
	}
}

// TestResize_UnchangedDimensionsNoOp covers the acceptance criterion that a
// resize with unchanged content dimensions leaves the page and history alone.
func TestResize_UnchangedDimensionsNoOp(t *testing.T) {
	path := writeTempFile(t, "reading-journey.md", loadFixture(t, "reading-journey.md"))
	m := NewModel(path)
	m = applyWindowSize(m, 90, 24)
	m = pressKey(m, " ")
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")

	page, history := m.CurrentPage(), append([]int(nil), m.History()...)
	m = applyWindowSize(m, 90, 24) // duplicate WindowSizeMsg, same dims

	if m.CurrentPage() != page {
		t.Errorf("current page changed on no-op resize: %d -> %d", page, m.CurrentPage())
	}
	if !equalInts(m.History(), history) {
		t.Errorf("history changed on no-op resize: %v -> %v", history, m.History())
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
