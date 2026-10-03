package book

import (
	"strings"
	"testing"
)

// Identical link markup on two source lines of one paragraph must attach each
// link to the display line rendering its own occurrence; issue #148.
func TestIssue148_RepeatedMarkupInParagraphLandsWhereRendered(t *testing.T) {
	content := strings.Join([]string{
		"Opening words",
		"[a](#title) and [b](#title) then filler filler filler filler filler",
		"filler filler filler filler filler filler filler filler [a](#title)",
	}, "\n")
	for _, size := range [][2]int{{40, 3}, {20, 2}, {30, 3}, {80, 5}} {
		b, err := Read(strings.NewReader(content), "x", size[0], size[1], false)
		if err != nil {
			t.Fatal(err)
		}
		assertLinksLandWhereRendered(t, b, 3)
	}
}

// A display line that ends exactly where the next source line begins renders
// none of that line, so it must not take that line's repeated link.
func TestIssue148_RepeatedMarkupAtSourceLineBoundary(t *testing.T) {
	content := strings.Join([]string{
		"Opening [a](#title) words here filler",
		"[a](#title) more filler filler filler",
	}, "\n")
	b, err := Read(strings.NewReader(content), "x", 40, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	assertLinksLandWhereRendered(t, b, 2)
}
