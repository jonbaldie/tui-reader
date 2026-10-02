package book

import (
	"fmt"
	"strings"
	"testing"
)

// singleParagraph returns n consecutive prose lines with no blank separator,
// with an internal link on every tenth line.
func singleParagraph(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line number %d of the document.", i)
		if i%10 == 0 {
			lines[i] += fmt.Sprintf(" See [line %d](#target-%d).", i, i)
		}
	}
	return strings.Join(lines, "\n")
}

// A quadratic link attachment makes this exceed the test timeout; issue #146.
func TestIssue146_ReadLargeSingleParagraph(t *testing.T) {
	const n = 100000
	b, err := Read(strings.NewReader(singleParagraph(n)), "large", 80, 25, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Pages) < 1500 {
		t.Fatalf("expected 1500+ pages, got %d", len(b.Pages))
	}

	var links []Link
	for _, page := range b.Pages {
		links = append(links, page.Links...)
	}
	if len(links) != n/10 {
		t.Fatalf("expected %d links, got %d", n/10, len(links))
	}
	last := b.Pages[len(b.Pages)-1].Links
	if len(last) == 0 || last[len(last)-1].Target != fmt.Sprintf("target-%d", n-10) {
		t.Fatalf("expected the final link on the last page, got %+v", last)
	}
}

func BenchmarkIssue146_ReadSingleParagraph(b *testing.B) {
	for _, lines := range []int{2000, 4000, 8000} {
		b.Run(fmt.Sprintf("lines=%d", lines), func(b *testing.B) {
			content := singleParagraph(lines)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := Read(strings.NewReader(content), "bench", 80, 25, false); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Reflowed paragraphs join source lines, so a display line can carry links
// from source lines other than its provenance line. Each link must still land
// on the display line that renders it.
func TestIssue146_ReflowedLinksLandWhereRendered(t *testing.T) {
	content := strings.Join([]string{
		"# Title",
		"",
		"Intro [a](#title).",
		"",
		"Opening words",
		"[a](#title) and [b](#title) then filler filler filler filler filler [a](#title)",
		"filler filler filler filler filler filler filler filler [d](#title)",
		"",
		"Single [c](#title) line with filler filler filler filler filler [c](#title).",
		"",
		"- [e](#title) item filler filler filler filler filler filler filler",
		"- [e](#title) item filler filler filler filler filler filler filler",
	}, "\n")
	for _, width := range []int{30, 40, 60} {
		b, err := Read(strings.NewReader(content), "reflow", width, 3, false)
		if err != nil {
			t.Fatal(err)
		}
		assertLinksLandWhereRendered(t, b, 9)
	}
}

// assertLinksLandWhereRendered checks that every page's links match, label by
// label and line by line, the link markup its lines render.
func assertLinksLandWhereRendered(t *testing.T, b *Book, total int) {
	t.Helper()
	got := 0
	for pi, page := range b.Pages {
		want := map[string]int{}
		for li, line := range page.Lines {
			for _, link := range ExtractLinks(line) {
				want[fmt.Sprintf("%d:%s", li, link.Label)]++
			}
		}
		have := map[string]int{}
		for _, link := range page.Links {
			have[fmt.Sprintf("%d:%s", link.LineOnPage, link.Label)]++
		}
		if fmt.Sprint(have) != fmt.Sprint(want) {
			t.Errorf("width %d page %d: links %v, rendered %v\n%s", b.PageWidth, pi, have, want, strings.Join(page.Lines, "\n"))
		}
		got += len(page.Links)
	}
	if got != total {
		t.Errorf("width %d: expected %d links, got %d", b.PageWidth, total, got)
	}
}
