package book

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestIssue192Wrapping(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  []string
	}{
		{"external links", "[see](https://example.org) text [x](https://example.org)", 16, []string{"[see](https://ex", "ample.org) text", "[x](https://exam", "ple.org)"}},
		{"unmatched brackets", "[[[ words [[ here", 9, []string{"[[[ words", "[[ here"}},
		{"internal after external", "[external](url) ([wide label](#target)) end", 23, []string{"[external](url)", "([wide label](#target))", "end"}},
		{"unicode and code", "界 `[a](#b)` [x y](#z)", 12, []string{"界 `[a](#b)`", "[x y](#z)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapLines([]string{tt.input}, tt.width)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("WrapLines() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkIssue192Wrap(b *testing.B) {
	for _, size := range []int{500, 1000, 2000, 4000} {
		for _, tc := range []struct{ name, unit string }{
			{"external", "[see](https://example.org) text "},
			{"unmatched", "[not-a-link text "},
		} {
			b.Run(fmt.Sprintf("%s/%d", tc.name, size), func(b *testing.B) {
				benchmarkWrapLines(b, strings.Repeat(tc.unit, size), 80)
			})
		}
	}
}
