package tui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestIssue79_WidthClipping_18x12(t *testing.T) {
	fixture := "../../docs/exploratory-testing/2026-09-10/evidence/fixtures/small-terminal.md"
	m := NewModel(fixture)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	termW, termH := 18, 12
	m = applyWindowSize(m, termW, termH)

	view := m.View()
	lines := strings.Split(view, "\n")

	// View must not exceed terminal height
	if len(lines) > termH {
		t.Errorf("view height = %d lines, exceeds terminal height %d", len(lines), termH)
	}

	// No line should exceed terminal width
	for i, line := range lines {
		stripped := stripAnsi(line)
		w := runewidth.StringWidth(stripped)
		if w > termW {
			t.Errorf("line %d width = %d, exceeds terminal width %d: %q", i, w, termW, stripped)
		}
	}
}

func TestIssue79_HeightClipping_40x10(t *testing.T) {
	fixture := "../../docs/exploratory-testing/2026-09-10/evidence/fixtures/small-terminal.md"
	m := NewModel(fixture)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	termW, termH := 40, 10
	m = applyWindowSize(m, termW, termH)

	view := m.View()
	lines := strings.Split(view, "\n")

	// View must not exceed terminal height (otherwise top lines scroll off)
	if len(lines) > termH {
		t.Errorf("view height = %d lines, exceeds terminal height %d (title will scroll off)", len(lines), termH)
	}

	// Title must be present in view
	foundTitle := false
	for _, line := range lines {
		if strings.Contains(stripAnsi(line), "Small Terminal") {
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Errorf("title 'Small Terminal' is absent from view")
	}
}

func TestIssue79_LongFilenameTitleClipping_80x24(t *testing.T) {
	fixture := "../../docs/exploratory-testing/2026-09-10/evidence/fixtures/this-is-an-extremely-long-reader-file-name-created-to-test-whether-the-title-header-can-grow-beyond-two-lines-and-hide-the-footer-and-controls-without-clipping-the-reading-content-or-page-indicator-or-key-help-during-real-terminal-use.md"
	m := NewModel(fixture)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	termW, termH := 80, 24
	m = applyWindowSize(m, termW, termH)

	view := m.View()
	lines := strings.Split(view, "\n")

	// View must not exceed terminal height
	if len(lines) > termH {
		t.Errorf("view height = %d lines, exceeds terminal height %d", len(lines), termH)
	}

	// Title prefix must not be dropped off the top
	foundPrefix := false
	for _, line := range lines {
		if strings.Contains(stripAnsi(line), "This Is An Extremely Long") {
			foundPrefix = true
			break
		}
	}
	if !foundPrefix {
		t.Errorf("title prefix 'This Is An Extremely Long' is missing from view")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"hello", -1, ""},
		{"hello", 0, ""},
		{"hello", 1, "."},
		{"hello", 2, ".."},
		{"hello", 3, "..."},
		{"hello", 4, "h..."},
		{"hello", 5, "hello"},
		{"hello", 6, "hello"},
		{"abcdefgh", 6, "abc..."},
		{"abcdefgh", 7, "abcd..."},
	}

	for _, tt := range tests {
		got := truncate(tt.s, tt.width)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
		}
	}
}

func TestRecalcLayout_NarrowTerminalWidth(t *testing.T) {
	path := writeTempFile(t, "narrow.md", "text\n")
	m := NewModel(path)

	testCases := []struct {
		termWidth int
		wantWidth int
	}{
		{20, 16},
		{18, 14},
		{5, 1},
		{4, 1},
		{3, 1},
		{1, 1},
		{0, 1},
	}

	for _, tc := range testCases {
		res := applyWindowSize(m, tc.termWidth, 20)
		if res.contentWidth != tc.wantWidth {
			t.Errorf("applyWindowSize(%d, 20) contentWidth = %d, want %d", tc.termWidth, res.contentWidth, tc.wantWidth)
		}
	}
}

func TestView_VerticalPaddingSwitch(t *testing.T) {
	path := writeTempFile(t, "vp.md", "# Title\nline 1\nline 2\nline 3\nline 4\nline 5\n")
	m := NewModel(path)

	// In 80x24, full height is 22 < 24, so View starts with top padding newline "\n"
	m24 := applyWindowSize(m, 80, 24)
	v24 := m24.View()
	if !strings.HasPrefix(v24, "\n") {
		t.Errorf("expected View() in 80x24 to start with newline padding")
	}

	// In 40x10, full height is 10 == termHeight, so View does not start with top padding newline
	m10 := applyWindowSize(m, 40, 10)
	v10 := m10.View()
	if strings.HasPrefix(v10, "\n") {
		t.Errorf("expected View() in 40x10 to not start with newline padding")
	}
}

func TestRenderHeaderFooter_NilBook(t *testing.T) {
	hdr := renderHeader(nil, 20)
	if !strings.Contains(hdr, "───") {
		t.Errorf("expected header with nil book to contain divider: %q", hdr)
	}

	ftr := renderFooter(nil, 0, 20)
	if !strings.Contains(ftr, "Page 1 of 0") {
		t.Errorf("expected footer with nil book to contain Page 1 of 0: %q", ftr)
	}
}

