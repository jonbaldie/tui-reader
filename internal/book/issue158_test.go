package book

import (
	"reflect"
	"strings"
	"testing"
)

// A book read at one geometry and reflowed to another must be laid out exactly
// like a book read at the second geometry: every laid-out field, exported or
// not, has to match.
func TestIssue158_ReflowMatchesFreshRead(t *testing.T) {
	content := "# Intro\n\nSee [usage](#usage) and [site](https://example.com) for a long wrapped line of prose.\n\n" +
		"```\ncode [x](#intro)\n```\n\n## Usage\n\n- item [back](#intro)\n- second item\n\nTail paragraph.\n"
	geometries := [][2]int{{60, 20}, {20, 4}, {0, 0}, {-5, -5}, {200, 1}}

	for _, plainText := range []bool{false, true} {
		for _, from := range geometries {
			for _, to := range geometries {
				want, err := Read(strings.NewReader(content), "t", to[0], to[1], plainText)
				if err != nil {
					t.Fatal(err)
				}
				got, err := Read(strings.NewReader(content), "t", from[0], from[1], plainText)
				if err != nil {
					t.Fatal(err)
				}
				got.Reflow(to[0], to[1])
				if !reflect.DeepEqual(got, want) {
					t.Errorf("plainText=%v %v→%v: reflowed book differs from fresh read\ngot:  %#v\nwant: %#v", plainText, from, to, got, want)
				}
			}
		}
	}
}
