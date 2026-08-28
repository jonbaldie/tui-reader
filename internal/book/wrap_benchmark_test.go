package book

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkWrapLines(b *testing.B) {
	for _, size := range []int{1 << 10, 1 << 12, 1 << 14} {
		b.Run(fmt.Sprintf("short-words/bytes=%d", size), func(b *testing.B) {
			benchmarkWrapLines(b, strings.Repeat("word ", size/5), 40)
		})
		b.Run(fmt.Sprintf("long-ascii/bytes=%d", size), func(b *testing.B) {
			benchmarkWrapLines(b, strings.Repeat("a", size), 40)
		})
		b.Run(fmt.Sprintf("long-unicode/bytes=%d", size), func(b *testing.B) {
			benchmarkWrapLines(b, strings.Repeat("界", size/3), 40)
		})
	}
}

func benchmarkWrapLines(b *testing.B, line string, width int) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapLines([]string{line}, width)
	}
}
