package book

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// FuzzBookLayout drives the public Book API with compact, structurally valid
// documents. A byte seed controls headings, internal links, whitespace, and
// Unicode while keeping every generated link target resolvable.
func FuzzBookLayout(f *testing.F) {
	for _, seed := range [][]byte{
		{0, 1, 2, 3},
		{5, 255, 0, 17, 42},
		{2, 7, 11, 13, 17, 19, 23, 29},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, seed []byte) {
		if len(seed) == 0 || len(seed) > 256 {
			t.Skip()
		}

		content, headings := fuzzDocument(seed)
		width, height := fuzzDimensions(seed)
		path := writeFuzzDocument(t, "doc.md", content)
		b, err := NewBook(path, width, height)
		if err != nil {
			t.Fatalf("NewBook: %v", err)
		}
		checkFuzzBook(t, b, headings)

		// A reflow can change page numbers, but not whether the generated links
		// display their label or resolve to their generated heading.
		secondWidth, secondHeight := fuzzDimensions(append(seed, 97))
		b.Reflow(secondWidth, secondHeight)
		checkFuzzBook(t, b, headings)
		first := snapshotPages(b.Pages)
		b.Reflow(secondWidth, secondHeight)
		if !reflect.DeepEqual(first, b.Pages) {
			t.Fatalf("same reflow changed pages\nbefore: %#v\nafter:  %#v", first, b.Pages)
		}

		// The loader normalizes all supported line endings before formatting.
		crlf := writeFuzzDocument(t, "crlf.md", strings.ReplaceAll(content, "\n", "\r\n"))
		crlfBook, err := NewBook(crlf, width, height)
		if err != nil {
			t.Fatalf("NewBook with CRLF: %v", err)
		}
		if !reflect.DeepEqual(b.RawLines, crlfBook.RawLines) {
			t.Fatalf("line-ending normalization changed raw lines\nLF:   %q\nCRLF: %q", b.RawLines, crlfBook.RawLines)
		}
	})
}

func fuzzDocument(seed []byte) (string, map[string]string) {
	sectionCount := int(seed[0]%6) + 1
	var out strings.Builder
	headingText := make(map[string]string, sectionCount)

	out.WriteString("# Contents\n\n")
	for i := 0; i < sectionCount; i++ {
		heading := fmt.Sprintf("Section %d", i+1)
		target := NormalizeAnchor(heading)
		headingText[target] = heading
		fmt.Fprintf(&out, "[Open %s](#%s)\n", heading, target)
	}
	out.WriteString("\n")

	words := []string{"lorem", "rune", "café", "東京", "🙂", "spaces", "reader"}
	for i := 0; i < sectionCount; i++ {
		heading := fmt.Sprintf("Section %d", i+1)
		fmt.Fprintf(&out, "# %s\n\n", heading)
		wordCount := int(seed[(i+1)%len(seed)]%12) + 1
		for j := 0; j < wordCount; j++ {
			if j > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(words[int(seed[(i+j+2)%len(seed)])%len(words)])
		}
		out.WriteString("\n\n")
	}
	return out.String(), headingText
}

func fuzzDimensions(seed []byte) (int, int) {
	// Generated links are deliberately kept whole on a display line. Inputs
	// where a label itself is split are a separate product bug (#23), not a
	// meaningful counterexample to this link-line property.
	width := int(seed[0]%41) + 40
	height := int(seed[len(seed)-1]%16) + 1
	return width, height
}

func writeFuzzDocument(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func checkFuzzBook(t *testing.T, b *Book, headings map[string]string) {
	t.Helper()
	if len(b.Pages) == 0 {
		t.Fatal("book has no pages")
	}

	formatted := formatParagraphsWithProvenance(b.RawLines, b.PageWidth)
	previousRaw := -1
	for index, line := range formatted {
		if line.raw < 0 {
			continue
		}
		if line.raw < previousRaw || line.raw >= len(b.RawLines) {
			t.Fatalf("formatted line %d has invalid provenance %d", index, line.raw)
		}
		if line.text != "" && strings.TrimSpace(b.RawLines[line.raw]) == "" {
			t.Fatalf("non-blank formatted line %d maps to blank source line %d", index, line.raw)
		}
		previousRaw = line.raw
	}

	seenTargets := make(map[string]int, len(headings))
	for pageIndex, page := range b.Pages {
		if page.Lines == nil {
			t.Fatalf("page %d has nil lines", pageIndex)
		}
		for _, link := range page.Links {
			seenTargets[link.Target]++
			if link.LineOnPage < 0 || link.LineOnPage >= len(page.Lines) {
				t.Fatalf("page %d link %#v has invalid line index", pageIndex, link)
			}
			if !strings.Contains(page.Lines[link.LineOnPage], link.Label) {
				t.Fatalf("page %d link %#v points to %q", pageIndex, link, page.Lines[link.LineOnPage])
			}
			heading, ok := headings[link.Target]
			if !ok {
				t.Fatalf("generated link target %q has no heading", link.Target)
			}
			destination := b.PageForAnchor(link.Target)
			if destination < 0 || destination >= len(b.Pages) {
				t.Fatalf("PageForAnchor(%q) = %d for %d pages", link.Target, destination, len(b.Pages))
			}
			if !pageHasText(b.Pages[destination], heading) {
				t.Fatalf("link target %q resolves to page %d without heading %q", link.Target, destination, heading)
			}
		}
	}
	for target := range headings {
		if seenTargets[target] != 1 {
			t.Fatalf("target %q attached %d times, want once", target, seenTargets[target])
		}
	}
}

func pageHasText(page Page, text string) bool {
	for _, line := range page.Lines {
		if strings.Contains(line, text) {
			return true
		}
	}
	return false
}

func snapshotPages(pages []Page) []Page {
	return append([]Page(nil), pages...)
}
