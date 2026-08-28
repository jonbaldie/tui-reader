package book

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkCallerVisibleOperations(b *testing.B) {
	for _, lines := range []int{64, 256, 1024} {
		b.Run(fmt.Sprintf("Load/lines=%d", lines), func(b *testing.B) {
			path := complexityBookFile(b, complexityDocument(lines))
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, _, err := Load(path); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Paginate/lines=%d", lines), func(b *testing.B) {
			rawLines := strings.Split(complexityDocument(lines), "\n")
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_ = Paginate(rawLines, 60, 20)
			}
		})

		b.Run(fmt.Sprintf("NewBook/lines=%d", lines), func(b *testing.B) {
			path := complexityBookFile(b, complexityDocument(lines))
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := NewBook(path, 60, 20); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Reflow/lines=%d", lines), func(b *testing.B) {
			path := complexityBookFile(b, complexityDocument(lines))
			book, err := NewBook(path, 60, 20)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				book.Reflow(72, 20)
			}
		})

		b.Run(fmt.Sprintf("PageForAnchor/uncached-lines=%d", lines), func(b *testing.B) {
			path := complexityBookFile(b, complexityDocument(lines))
			book, err := NewBook(path, 60, 20)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				book.rawLinePages = nil
				_ = book.PageForAnchor("section-0")
			}
		})

		b.Run(fmt.Sprintf("PageForAnchor/cached-lines=%d", lines), func(b *testing.B) {
			path := complexityBookFile(b, complexityDocument(lines))
			book, err := NewBook(path, 60, 20)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_ = book.PageForAnchor("section-0")
			}
		})
	}
}

func BenchmarkExtractLinks(b *testing.B) {
	for _, links := range []int{16, 64, 256} {
		b.Run(fmt.Sprintf("links=%d", links), func(b *testing.B) {
			line := strings.Repeat("[label](#target) ", links)
			b.SetBytes(int64(len(line)))
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_ = ExtractLinks(line)
			}
		})
	}
}

func complexityBookFile(b *testing.B, content string) string {
	b.Helper()
	path := filepath.Join(b.TempDir(), "complexity.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}
	return path
}

func complexityDocument(lines int) string {
	var document strings.Builder
	for i := range lines {
		fmt.Fprintf(&document, "# Section %d\n", i)
		document.WriteString("A paragraph with [a link](#section-0) and enough words to wrap at ordinary widths.\n\n")
	}
	return document.String()
}
