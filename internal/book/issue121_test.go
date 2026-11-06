package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue121_SetextLinkWithIndentedContinuation(t *testing.T) {
	b, err := Read(strings.NewReader("[a\n   b](#target)\n===\n"), "heading", 40, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	page := b.Page(0)
	want := []Link{{Label: "a b", Target: "target", LineOnPage: 0}}
	if !reflect.DeepEqual(page.Links, want) {
		t.Fatalf("links = %+v, want %+v", page.Links, want)
	}
	if len(page.LabelSpans) != 1 || page.Lines[0][page.LabelSpans[0].Start:page.LabelSpans[0].End] != "a b" {
		t.Fatalf("heading = %q, spans = %+v", page.Lines, page.LabelSpans)
	}
}

func TestIssue121_LinkLocationsOnReadAndReflow(t *testing.T) {
	const source = "[one](#a) [one](#b)\n[split\nlabel](#c)\n\n`[inert](#x)`\n\n```\n[hidden](#x)\n```\n\n    [indented](#x)\n"
	b, err := Read(strings.NewReader(source), "links", 12, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	check := func(want [][]Link) {
		t.Helper()
		if b.PageCount() != len(want) {
			t.Fatalf("page count = %d, want %d", b.PageCount(), len(want))
		}
		for i, expected := range want {
			got := b.Page(i).Links
			if len(got) == 0 {
				got = nil
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("page %d links = %+v, want %+v", i, got, expected)
			}
		}
	}
	check([][]Link{
		{{Label: "one", Target: "a", LineOnPage: 0}, {Label: "one", Target: "b", LineOnPage: 1}},
		{{Label: "split label", Target: "c", LineOnPage: 0}},
		nil, nil, nil, nil, nil, nil,
	})
	b.Reflow(40, 2)
	check([][]Link{
		{{Label: "one", Target: "a", LineOnPage: 0}, {Label: "one", Target: "b", LineOnPage: 0}, {Label: "split label", Target: "c", LineOnPage: 0}},
		nil, nil, nil, nil,
	})
	plain, err := Read(strings.NewReader(source), "plain", 40, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < plain.PageCount(); i++ {
		if p := plain.Page(i); len(p.Links) != 0 || len(p.LabelSpans) != 0 {
			t.Errorf("plain text page %d has link metadata: %+v", i, p)
		}
	}
}
