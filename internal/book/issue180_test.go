package book

import (
	"slices"
	"strings"
	"testing"
)

func TestIssue180_LineBlocks(t *testing.T) {
	for _, tt := range []struct {
		name  string
		raw   []string
		width int
		want  []string
	}{
		{
			name: "table padding and quotes",
			raw: []string{
				"| Version | Date       | Notes          |",
				"|---------|------------|----------------|",
				"| 0.1.12  | 2026-09-30 | Fence fixes    |",
				"| 0.1.13  | 2026-10-05 | Resize fixes   |",
				"",
				"> First quoted line.", "> Second quoted line.", "> Third quoted line.",
			},
			width: 80,
			want: []string{
				"| Version | Date       | Notes          |",
				"|---------|------------|----------------|",
				"| 0.1.12  | 2026-09-30 | Fence fixes    |",
				"| 0.1.13  | 2026-10-05 | Resize fixes   |",
				"",
				"> First quoted line.", "> Second quoted line.", "> Third quoted line.",
			},
		},
		{
			name:  "prose boundaries and outer whitespace",
			raw:   []string{"Intro text", "  | a | b |  ", " | c | d | ", " > Quote one", "> Quote two", "Outro text", "still prose"},
			width: 80,
			want:  []string{"Intro text", "", "| a | b |", "| c | d |", "", "> Quote one", "> Quote two", "", "  Outro text still prose"},
		},
		{
			name: "hard wrap keeps spaces",
			raw:  []string{"| ab  cd  ef |"}, width: 6,
			want: []string{"| ab  ", "cd  ef", " |"},
		},
		{
			name: "hard wrap uses display columns",
			raw:  []string{"| 猫 犬 |"}, width: 5,
			want: []string{"| 猫 ", "犬 |"},
		},
		{
			name: "quote wraps words",
			raw:  []string{"> one two three", "> next"}, width: 9,
			want: []string{"> one two", "three", "> next"},
		},
		{
			name: "code stays code",
			raw:  []string{"    | code  row |", "    > code  quote"}, width: 80,
			want: []string{"    | code  row |", "    > code  quote"},
		},
		{
			name: "fence stays code",
			raw:  []string{"```", "| code  row |", "> code  quote", "```"}, width: 80,
			want: []string{"    ```", "    | code  row |", "    > code  quote", "    ```"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := readDisplayLines(t, tt.raw, tt.width); !slices.Equal(got, tt.want) {
				t.Fatalf("display lines = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssue180_TableLinksAcrossHardWrapAndReflow(t *testing.T) {
	b := readBook(t, "|   [first](#one)  [second](#two) |", true, 12, 1)
	for _, width := range []int{12, 8, 80} {
		b.Reflow(width, 1)
		var got []Link
		var pages []int
		for i := 0; i < b.PageCount(); i++ {
			for _, link := range b.Page(i).Links {
				got = append(got, link)
				pages = append(pages, i)
			}
		}
		want := []Link{{Label: "first", Target: "one", LineOnPage: 0}, {Label: "second", Target: "two", LineOnPage: 0}}
		wantPages := map[int][]int{12: {0, 1}, 8: {0, 2}, 80: {0, 0}}
		if !slices.Equal(got, want) || !slices.Equal(pages, wantPages[width]) {
			t.Fatalf("width %d links = %+v on pages %v, want %+v on pages %v", width, got, pages, want, wantPages[width])
		}
	}
}

func TestIssue180_SourcePositions(t *testing.T) {
	b := readBook(t, "| ab  cd  ef |\n| next |\n\n> quote\n> next", true, 6, 1)
	if got := b.PageForRawLine(1); got != 3 {
		t.Fatalf("second table row page = %d, want 3", got)
	}
	if got := b.PageForRawLine(4); got != 8 {
		t.Fatalf("second quote line page = %d, want 8", got)
	}
}

func TestIssue180_PlainTextUnchanged(t *testing.T) {
	b, err := Read(strings.NewReader("| a  | b |\n| c  | d |\n> one\n> two"), "Text", 80, 20, true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"| a  | b |", "| c  | d |", "> one", "> two"}
	if got := b.Page(0).Lines; !slices.Equal(got, want) {
		t.Fatalf("plain text lines = %q, want %q", got, want)
	}
}
