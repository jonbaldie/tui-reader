package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/muesli/termenv"
)

func BenchmarkModelViewLinkDense(b *testing.B) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	b.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	for _, linkCount := range []int{16, 64, 256} {
		b.Run(fmt.Sprintf("links=%d", linkCount), func(b *testing.B) {
			model := linkDenseModel(b, linkCount)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = model.View()
			}
		})
	}
}

func linkDenseModel(b *testing.B, linkCount int) Model {
	var line strings.Builder
	for i := 0; i < linkCount; i++ {
		if i > 0 {
			line.WriteByte(' ')
		}
		fmt.Fprintf(&line, "[link-%03d](#target-%03d)", i, i)
	}

	// A page as wide as the line keeps every link on one dense line.
	reader, err := book.Read(strings.NewReader(line.String()), "Benchmark", line.Len(), 3, false)
	if err != nil {
		b.Fatal(err)
	}
	if got := len(reader.Page(0).Links); got != linkCount {
		b.Fatalf("page 0 has %d links, want %d", got, linkCount)
	}

	return Model{
		book:          reader,
		selectedLink:  linkCount / 2,
		termWidth:     80,
		termHeight:    10,
		contentWidth:  72,
		contentHeight: 3,
	}
}
