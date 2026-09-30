package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func TestIssue92_FooterPageInfoTruncated(t *testing.T) {
	// Narrow content width where "Page 1 of 1" (11 cols) would wrap without truncate.
	ftr := renderFooter(nil, 0, 8)
	lines := strings.Split(ftr, "\n")
	if len(lines) != 3 {
		t.Fatalf("footer lines = %d, want 3 (divider, pageInfo, help); got %q", len(lines), ftr)
	}
	for i, line := range lines {
		w := runewidth.StringWidth(stripAnsi(line))
		if w > 8 {
			t.Errorf("footer line %d width = %d > 8: %q", i, w, stripAnsi(line))
		}
	}
}

// TestIssue92_FooterPageInfoFitsCompactTerminal reproduces the footer wrapping
// that can push the title off the top of a short terminal.
func TestIssue92_FooterPageInfoFitsCompactTerminal(t *testing.T) {
	m := NewModel(writeTempFile(t, "compact-footer-probe.md", "Short.\n"))
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	const termWidth, termHeight = 12, 10
	m = applyWindowSize(m, termWidth, termHeight)

	footer := renderFooter(m.book, m.currentPage, m.contentWidth)
	plainFooter := stripAnsi(footer)
	wantPageInfo := truncate("Page 1 of 1", m.contentWidth)
	if !strings.Contains(plainFooter, wantPageInfo) {
		t.Errorf("footer is missing truncated page info %q: %q", wantPageInfo, plainFooter)
	}
	if got := lipgloss.Height(footer); got != 3 {
		t.Errorf("footer height = %d, want 3 lines at content width %d: %q", got, m.contentWidth, plainFooter)
	}

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > termHeight {
		t.Errorf("view height = %d, exceeds terminal height %d: %q", len(lines), termHeight, stripAnsi(view))
	}

	wantTitle := truncate(m.book.Title, m.contentWidth)
	if !strings.Contains(stripAnsi(view), wantTitle) {
		t.Errorf("view is missing truncated title %q: %q", wantTitle, stripAnsi(view))
	}
}

func TestIssue92_CompactTerminalKeepsTitle(t *testing.T) {
	fixture := "../../docs/exploratory-testing/2026-09-12/evidence/fixtures/compact-footer-wrap.md"
	m := NewModel(fixture)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	termW, termH := 12, 10
	m = applyWindowSize(m, termW, termH)

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > termH {
		t.Errorf("view height = %d lines, exceeds terminal height %d", len(lines), termH)
	}

	// Title may be truncated to fit contentWidth; require a recognizable prefix on line 0.
	first := stripAnsi(lines[0])
	if !strings.Contains(first, "Compa") {
		t.Errorf("title prefix missing from 12x10 view line 0: %q", first)
	}

	for i, line := range lines {
		w := runewidth.StringWidth(stripAnsi(line))
		if w > termW {
			t.Errorf("line %d width = %d > %d: %q", i, w, termW, stripAnsi(line))
		}
	}
}
