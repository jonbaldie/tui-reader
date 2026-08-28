package tui

import (
	"strings"
	"testing"
)

// FuzzPageNavigationAgainstReference compares page-only navigation with a
// small independent state machine. Link following, wrapping, and styling are
// outside the shared domain because those are intentionally reader-specific.
func FuzzPageNavigationAgainstReference(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 2, 0})
	f.Add([]byte{3, 3, 1, 0, 2})
	f.Fuzz(func(t *testing.T, actions []byte) {
		if len(actions) == 0 || len(actions) > 256 {
			t.Skip()
		}
		m := NewModel(writeTempFile(t, "navigation.md", strings.Repeat("content line\n", 100)))
		m = applyWindowSize(m, 40, 12)
		last := len(m.book.Pages) - 1
		expected := 0
		for step, action := range actions {
			switch action % 4 {
			case 0:
				m = pressKey(m, "right")
				if expected < last {
					expected++
				}
			case 1:
				m = pressKey(m, "left")
				if expected > 0 {
					expected--
				}
			case 2:
				m = pressKey(m, "g")
				expected = 0
			case 3:
				m = pressKey(m, "G")
				expected = last
			}
			if m.currentPage != expected {
				t.Fatalf("step %d: page = %d, want %d", step, m.currentPage, expected)
			}
		}
	})
}
