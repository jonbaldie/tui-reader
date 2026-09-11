package book

import (
	"reflect"
	"testing"
)

func TestIssue78_InlineCodeLinkIsNotSelectable(t *testing.T) {
	b, err := NewBook("../../docs/exploratory-testing/2026-09-10/evidence/fixtures/inline-code.md", 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	for pi, p := range b.Pages {
		if len(p.Links) > 0 {
			t.Errorf("page %d links = %+v, want none: the only link is inside inline code", pi+1, p.Links)
		}
	}
}

func TestIssue78_ExtractLinksSkipsInlineCode(t *testing.T) {
	tests := []struct {
		line string
		want []Link
	}{
		{"`[a](#x)`", nil},
		{"See `[a](#x)` and [b](#y).", []Link{{Label: "b", Target: "y"}}},
		{"``[a](#x) ` still code``", nil},
		{"`code` then [b](#y)", []Link{{Label: "b", Target: "y"}}},
		{"unclosed ` then [b](#y)", []Link{{Label: "b", Target: "y"}}},
		{"[use `x`](#y)", []Link{{Label: "use `x`", Target: "y"}}},
		{"[a `](#x)`", nil},
	}
	for _, tt := range tests {
		if got := ExtractLinks(tt.line); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ExtractLinks(%q) = %+v, want %+v", tt.line, got, tt.want)
		}
	}
}
