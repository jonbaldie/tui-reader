package book

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestIssue193_CodeAndLinksSurviveReflow(t *testing.T) {
	cases := []struct {
		name, text string
		labels     []string
	}{
		{"code contains link", "`[hidden](#a)` [shown](#a)", []string{"shown"}},
		{"label contains code", "[with `code`](#a) [shown](#a)", []string{"with `code`", "shown"}},
		{"code splits link", "`start [hidden`](#a) [shown](#a)", []string{"shown"}},
		{"unclosed backtick", "`open [shown](#a)", []string{"shown"}},
		{"repeated links", "`code` [same](#a) `code` [same](#a)", []string{"same", "same"}},
		{"nested target", "`[hidden](#outer[also](#a))` [shown](#a)", []string{"shown"}},
		{"source line split", "`code` [two\nwords](#a) [shown](#a)", []string{"two words", "shown"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Read(strings.NewReader("# A\n\n"+tc.text), "mixed", 80, 3, false)
			if err != nil {
				t.Fatal(err)
			}
			for _, width := range []int{80, 30} {
				b.Reflow(width, 3)
				var labels []string
				for pi, p := range b.pages {
					for _, l := range p.Links {
						labels = append(labels, l.Label)
						if l.Target != "a" || l.LineOnPage < 0 || l.LineOnPage >= len(p.Lines) {
							t.Fatalf("page %d: invalid link %+v", pi, l)
						}
						if !strings.Contains(p.Lines[l.LineOnPage], "[") {
							t.Fatalf("link not on rendered line: %+v", l)
						}
					}
				}
				if !reflect.DeepEqual(labels, tc.labels) {
					t.Fatalf("width %d: labels %q, want %q", width, labels, tc.labels)
				}
			}
		})
	}
}

func BenchmarkIssue193_CodeOverlap(b *testing.B) {
	for _, kind := range []string{"mixed", "links"} {
		for _, n := range []int{2000, 4000, 8000} {
			unit := "textxx [link](#a) "
			if kind == "mixed" {
				unit = "`code` [link](#a) "
			}
			text := strings.Repeat(unit, n)
			b.Run(fmt.Sprintf("Extract/%s/%d", kind, n), func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					ExtractLinks(text)
				}
			})
			b.Run(fmt.Sprintf("Wrap/%s/%d", kind, n), func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					eachWrapToken(text, func(wrapToken) {})
				}
			})
			b.Run(fmt.Sprintf("Reflow/%s/%d", kind, n), func(b *testing.B) {
				book, err := Read(strings.NewReader(text), "bench", 80, 25, false)
				if err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					book.Reflow(72, 25)
				}
			})
		}
	}
}
