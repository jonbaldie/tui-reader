package tui

import (
	"strings"
	"testing"
)

func TestIssue180_FollowLinksInLineBlocks(t *testing.T) {
	for _, line := range []string{"| [Go](#destination) |", "> [Go](#destination)"} {
		t.Run(line, func(t *testing.T) {
			doc := line + "\n\n" + strings.Repeat("Filler.\n\n", 30) + "# Destination\nTarget content.\n"
			m := applyWindowSize(newTestModel(t, "issue180.md", doc), 80, 24)
			if m.Err() != nil {
				t.Fatal(m.Err())
			}
			m = pressKey(m, "tab")
			if m.SelectedLink() != 0 {
				t.Fatalf("selected link = %d, want 0", m.SelectedLink())
			}
			m = pressKey(m, "enter")
			if got, want := m.CurrentPage(), m.BookRef().PageForAnchor("destination"); got != want || got == 0 {
				t.Fatalf("Enter landed on page %d, want destination page %d beyond page 0", got, want)
			}
			if !strings.Contains(m.View(), "Target content.") {
				t.Fatalf("destination content is not visible:\n%s", m.View())
			}
		})
	}
}
