package book

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestIssue188_TableSpansAndContinuationOwnership(t *testing.T) {
	row := "|" + strings.Repeat(" ", 70) + "[Wide target](#destination) |"
	b, err := Read(strings.NewReader(strings.Repeat("| filler\n", 16)+row), "table", 72, 17, false)
	if err != nil {
		t.Fatal(err)
	}
	owner, next := b.Page(0), b.Page(1)
	if len(owner.Links) != 1 || owner.Links[0].LineOnPage != 16 || len(next.Links) != 0 {
		t.Fatalf("owner links = %+v, continuation links = %+v", owner.Links, next.Links)
	}
	var table strings.Builder
	for _, line := range owner.Lines[16:] {
		table.WriteString(line)
	}
	for _, line := range next.Lines {
		table.WriteString(line)
	}
	if table.String() != row {
		t.Fatalf("table bytes = %q, want %q", table.String(), row)
	}
	want := LabelSpan{Line: 16, Start: 71, End: 72, LinkIndex: 0, Opener: true}
	if !reflect.DeepEqual(owner.LabelSpans, []LabelSpan{want}) {
		t.Fatalf("owner spans = %+v, want opener %+v", owner.LabelSpans, want)
	}
	if len(next.LabelSpans) == 0 || next.LabelSpans[0].LinkIndex != -1 {
		t.Fatalf("continuation spans = %+v", next.LabelSpans)
	}
	owner.LabelSpans[0].Start = 0
	if b.Page(0).LabelSpans[0] != want {
		t.Fatal("Page did not clone label spans")
	}
}

func TestIssue188_IndentedJoinedParagraphSpans(t *testing.T) {
	b, err := Read(strings.NewReader("# Heading\n\nFirst paragraph.\n\n[link\nlabel](#target)\n"), "indent", 72, 20, false)
	if err != nil {
		t.Fatal(err)
	}
	p := b.Page(0)
	if len(p.Links) != 1 {
		t.Fatalf("links = %+v", p.Links)
	}
	var label string
	for _, span := range p.LabelSpans {
		label += p.Lines[span.Line][span.Start:span.End]
		if span.Start < 2 || !strings.HasPrefix(p.Lines[span.Line], "  [") {
			t.Fatalf("indent not accounted for in span %+v on %q", span, p.Lines[span.Line])
		}
	}
	if label != "link label" {
		t.Fatalf("label = %q, want link label", label)
	}
}

func TestIssue188_ReflowUnicodeAndDuplicateOccurrences(t *testing.T) {
	doc := "`[same](#target)` [same](#target) [same](#target) [界界界界](#target)\n"
	b, err := Read(strings.NewReader(doc), "links", 20, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{20, 8, 72} {
		b.Reflow(width, 30)
		p := b.Page(0)
		if len(p.Links) != 3 {
			t.Fatalf("width %d links = %+v", width, p.Links)
		}
		var labels [3]string
		for _, span := range p.LabelSpans {
			if span.Opener {
				continue
			}
			line := p.Lines[span.Line]
			if span.Start < 0 || span.End > len(line) || !utf8.ValidString(line[:span.Start]) || !utf8.ValidString(line[:span.End]) {
				t.Fatalf("width %d invalid span %+v in %q", width, span, line)
			}
			labels[span.LinkIndex] += line[span.Start:span.End]
		}
		if labels != [3]string{"same", "same", "界界界界"} {
			t.Fatalf("width %d labels = %q", width, labels)
		}
	}
	plain, err := Read(strings.NewReader(doc), "plain", 20, 30, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(plain.Page(0).Links) != 0 || len(plain.Page(0).LabelSpans) != 0 {
		t.Fatal("plain text acquired link metadata")
	}
}
