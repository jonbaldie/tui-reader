package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

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
