package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/muesli/termenv"
)

func TestIssue111_FencedHashCommentIsNotStyledAsHeading(t *testing.T) {
	raw := []string{
		"# Guide",
		"",
		"## Usage",
		"",
		"```sh",
		"# usage",
		"tool --help",
		"```",
	}
	b, err := book.Read(strings.NewReader(strings.Join(raw, "\n")), "Guide", 60, 20, false)
	if err != nil {
		t.Fatal(err)
	}
	formatted := b.Page(0).Lines
	var comment string
	for _, line := range formatted {
		if strings.Contains(line, "# usage") && !strings.Contains(line, "##") {
			comment = line
			break
		}
	}
	if comment == "" {
		t.Fatalf("missing # usage display line: %q", formatted)
	}

	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })
	got := NewModelFromBook(b).View()
	if want := textStyle.Render(comment); !strings.Contains(got, want) {
		t.Fatalf("fenced comment %q style = %q, want body style %q", comment, got, want)
	}
}
