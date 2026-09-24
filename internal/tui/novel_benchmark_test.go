package tui

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// novelFile represents a 120,000-word manuscript with chapter links and Unicode.
func novelFile(t testing.TB) string {
	t.Helper()
	var text strings.Builder
	for chapter := range 30 {
		fmt.Fprintf(&text, "# Chapter %d\n\n[Contents](#chapter-0)\n\n", chapter)
		for range 100 {
			text.WriteString(strings.Repeat("She crossed the quiet courtyard and opened the café door. ", 4))
			text.WriteString("\n\n")
		}
	}
	path := filepath.Join(t.TempDir(), "novel.md")
	if err := os.WriteFile(path, []byte(text.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func BenchmarkNovelJourney(b *testing.B) {
	path := novelFile(b)
	b.Run("OpenToFirstView", func(b *testing.B) {
		samples := make([]time.Duration, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			start := time.Now()
			model, _ := NewModel(path).Update(tea.WindowSizeMsg{Width: 80, Height: 30})
			if model.(Model).err != nil {
				b.Fatal(model.(Model).err)
			}
			_ = model.View()
			samples[i] = time.Since(start)
		}
		b.StopTimer()
		slices.Sort(samples)
		b.ReportMetric(float64(samples[(len(samples)*3+3)/4-1].Nanoseconds()), "p75-ns/op")
	})
	b.Run("ResizeToView", func(b *testing.B) {
		model, _ := NewModel(path).Update(tea.WindowSizeMsg{Width: 80, Height: 30})
		samples := make([]time.Duration, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			start := time.Now()
			model, _ = model.Update(tea.WindowSizeMsg{Width: 70 + i%2*10, Height: 30})
			_ = model.View()
			samples[i] = time.Since(start)
		}
		b.StopTimer()
		slices.Sort(samples)
		b.ReportMetric(float64(samples[(len(samples)*3+3)/4-1].Nanoseconds()), "p75-ns/op")
	})
}

// Hash every page and link plus rendered views across resize widths. The hashes
// are captured before optimizing; they protect wrapping, provenance and links.
func TestNovelJourneyOutput(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })
	model := NewModel(novelFile(t))
	if model.err != nil {
		t.Fatal(model.err)
	}
	want := map[int]string{
		80: "aac9b5e3b7e43633f8b5a9d8351784451e722db3a471c14ee9df3bed07a8981d",
		70: "aa9d03bc4562f7de132074ca30d4fb20ca3dfe3c8adca0de97ba5ecdefd4e39e",
	}
	for _, width := range []int{80, 70, 80} {
		updated, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: 30})
		model = updated.(Model)
		digest := sha256.New()
		for page := range model.book.Pages {
			model.currentPage = page
			fmt.Fprintf(digest, "%v\n%d\n%s\n", model.book.Pages[page], model.book.RawLineForPage(page), model.View())
		}
		if got := fmt.Sprintf("%x", digest.Sum(nil)); got != want[width] {
			t.Fatalf("width=%d: output hash %s, want %s", width, got, want[width])
		}
	}
}
