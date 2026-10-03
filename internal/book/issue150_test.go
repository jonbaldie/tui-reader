package book

import (
	"fmt"
	"strings"
	"testing"
)

func TestIssue150_RepeatedMarkupAttachesToRenderedLine(t *testing.T) {
	content := strings.Join([]string{
		"Opening [a](#title)",
		"words filler filler filler filler filler filler filler filler filler [a](#title)",
	}, "\n")

	for _, size := range [][2]int{{40, 2}, {20, 2}, {30, 3}, {80, 5}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			b, err := Read(strings.NewReader(content), "x", size[0], size[1], false)
			if err != nil {
				t.Fatal(err)
			}
			assertLinksLandWhereRendered(t, b, 2)
		})
	}
}
