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
			model := linkDenseModel(linkCount)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = model.View()
			}
		})
	}
}

func linkDenseModel(linkCount int) Model {
	var line strings.Builder
	links := make([]book.Link, 0, linkCount)
	for i := 0; i < linkCount; i++ {
		if i > 0 {
			line.WriteByte(' ')
		}
		label := fmt.Sprintf("link-%03d", i)
		target := fmt.Sprintf("target-%03d", i)
		fmt.Fprintf(&line, "[%s](#%s)", label, target)
		links = append(links, book.Link{Label: label, Target: target, LineOnPage: 0})
	}

	return Model{
		book: &book.Book{
			Title: "Benchmark",
			Pages: []book.Page{{Lines: []string{line.String()}, Links: links}},
		},
		selectedLink:  linkCount / 2,
		termWidth:     80,
		termHeight:    10,
		contentWidth:  72,
		contentHeight: 3,
	}
}
