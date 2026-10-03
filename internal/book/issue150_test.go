package book

import (
	"strings"
	"testing"
)

// A display line that renders source line r's link and the start of source
// line r+1 must not take r+1's repeated link, which wraps onto a later line;
// issue #150.
func TestIssue150_RepeatedMarkupSpanningSourceLinesLandsWhereRendered(t *testing.T) {
	content := strings.Join([]string{
		"Opening [a](#title)",
		"words filler filler filler filler filler filler filler filler filler [a](#title)",
	}, "\n")
	for _, size := range [][2]int{{40, 2}, {20, 2}, {30, 3}, {80, 5}} {
		b, err := Read(strings.NewReader(content), "repeat", size[0], size[1], false)
		if err != nil {
			t.Fatal(err)
		}
		assertLinksLandWhereRendered(t, b, 2)
	}
}
