package book

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIssue194_NormalizedWhitespaceProvenance(t *testing.T) {
	lines := []string{"alpha  beta", "gamma\tdelta", "éclair  漢字", "supercalifragilistic"}
	got := formatMultiLineParagraph(lines, 0, true, 11)
	wantText := []string{"alpha beta", "gamma delta", "éclair 漢字", "supercalifr", "agilistic"}
	wantRaw := []int{0, 1, 2, 3, 3}
	if len(got) != len(wantText) {
		t.Fatalf("got %+v, want %q", got, wantText)
	}
	for i := range got {
		if got[i].text != wantText[i] || got[i].raw != wantRaw[i] {
			t.Errorf("line %d = %q, raw %d; want %q, raw %d", i, got[i].text, got[i].raw, wantText[i], wantRaw[i])
		}
	}
}

func TestIssue194_ReflowPositionAndLinks(t *testing.T) {
	content := "alpha  beta\ngamma\tdelta\n[long link label](#target)  trailing\nlast  marker\n"
	path := filepath.Join(t.TempDir(), "book.md")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	b, err := NewBook(path, 12, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.PositionForPage(1); got.Raw != 1 {
		t.Fatalf("page 1 position = %+v, want raw 1", got)
	}
	pos := b.PositionForPage(1)
	if got := b.pages[2].Links; len(got) != 1 || got[0] != (Link{Label: "long link label", Target: "target", LineOnPage: 0}) {
		t.Fatalf("width 12 link = %+v", got)
	}
	if got := b.pages[3].LabelSpans; len(got) != 1 || got[0] != (LabelSpan{Line: 0, Start: 0, End: 4, LinkIndex: -1}) {
		t.Fatalf("width 12 continuation = %+v", got)
	}
	b.Reflow(20, 1)
	page := b.PageForPosition(pos)
	if got := strings.Join(b.pages[page].Lines, " "); !strings.Contains(got, "gamma") {
		t.Errorf("reflow page %d = %q, want gamma", page, got)
	}
	if got := b.PositionForPage(1); got != (Position{Raw: 1, Offset: 5}) {
		t.Errorf("width 20 page 1 position = %+v, want raw 1 offset 5", got)
	}
	if got := b.PageForPosition(Position{Raw: 1, Offset: 4}); got != 0 {
		t.Errorf("width 20 raw 1 offset 4 page = %d, want 0", got)
	}
	if got := b.pages[2].Links; len(got) != 1 || got[0] != (Link{Label: "long link label", Target: "target", LineOnPage: 0}) {
		t.Fatalf("width 20 link = %+v", got)
	}
	if got := b.pages[2].LabelSpans; len(got) != 1 || got[0] != (LabelSpan{Line: 0, Start: 1, End: 16, LinkIndex: 0}) {
		t.Errorf("width 20 link span = %+v", got)
	}
}

func BenchmarkIssue194_NormalizedParagraph(b *testing.B) {
	for _, n := range []int{2000, 4000, 8000} {
		b.Run(fmt.Sprintf("NewBook/lines=%d", n), func(b *testing.B) {
			path := complexityBookFile(b, strings.Repeat("alpha  beta\tgamma  delta  epsilon\n", n))
			b.ResetTimer()
			for range b.N {
				if _, err := NewBook(path, 72, 20); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("Reflow/lines=%d", n), func(b *testing.B) {
			path := complexityBookFile(b, strings.Repeat("alpha  beta\tgamma  delta  epsilon\n", n))
			book, err := NewBook(path, 72, 20)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for range b.N {
				book.Reflow(72, 20)
			}
		})
	}
}
