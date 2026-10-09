package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
)

func BenchmarkParseArgs(b *testing.B) {
	for _, count := range []int{4, 16, 64} {
		b.Run(fmt.Sprintf("arguments=%d", count), func(b *testing.B) {
			args := make([]string, count)
			args[0] = "--dump=3"
			for i := 1; i < count-1; i++ {
				args[i] = "ignored-flag"
			}
			args[count-1] = "book.md"
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, _ = parseArgs(args)
			}
		})
	}
}

func BenchmarkRenderDump(b *testing.B) {
	for _, pages := range []int{4, 16, 64} {
		b.Run(fmt.Sprintf("pages=%d", pages), func(b *testing.B) {
			content := strings.TrimSuffix(strings.Repeat("content\nmore content\n", pages), "\n")
			reader, err := book.Read(strings.NewReader(content), "Benchmark", 80, 2, true)
			if err != nil {
				b.Fatal(err)
			}
			if got := reader.PageCount(); got != pages {
				b.Fatalf("PageCount() = %d, want %d", got, pages)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_ = renderDump(reader, 0)
			}
		})
	}
}
