package book

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkBookReflowLinkDense(b *testing.B) {
	for _, lines := range []int{32, 128, 512} {
		b.Run(fmt.Sprintf("lines=%d", lines), func(b *testing.B) {
			path := benchmarkBookFile(b, linkDenseDocument(lines))
			book, err := NewBook(path, 24, 20)
			if err != nil {
				b.Fatalf("NewBook: %v", err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				book.Reflow(24, 20)
			}
		})
	}
}

func benchmarkBookFile(b *testing.B, content string) string {
	b.Helper()
	path := filepath.Join(b.TempDir(), "link-dense.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatalf("write benchmark document: %v", err)
	}
	return path
}

func linkDenseDocument(lines int) string {
	var out strings.Builder
	out.WriteString("# target\n\n")
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&out, "padding [first link %d](#target) more padding [second link %d](#target) trailing words\n", i, i)
	}
	return out.String()
}
