package book

import (
	"strconv"
	"strings"
	"testing"
)

func TestNormalizeAnchorParity(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"long hyphen run", "start" + strings.Repeat("-", 256) + "end", "start-end"},
		{"spaces", "  several   spaces  ", "several-spaces"},
		{"punctuation", "hello, world! (again)", "hello-world-again"},
		{"boundaries", "---heading---", "heading"},
		{"digits", "Part 42 - Section 7", "part-42-section-7"},
		{"unicode removed", "Café — 東京", "caf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeAnchor(tt.input); got != tt.want {
				t.Fatalf("NormalizeAnchor(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkNormalizeAnchorAdversarial(b *testing.B) {
	for _, size := range []int{1 << 10, 1 << 14, 1 << 18} {
		b.Run("bytes="+strconv.Itoa(size), func(b *testing.B) {
			input := "start" + strings.Repeat("-", size/2) + strings.Repeat("unaffected text ", size/32) + "end"
			b.SetBytes(int64(len(input)))
			b.ResetTimer()
			for range b.N {
				NormalizeAnchor(input)
			}
		})
	}
}
