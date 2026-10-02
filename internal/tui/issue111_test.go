package tui

import (
	"strings"
	"testing"

	"github.com/jonbaldie/tui-reader/internal/book"
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
	b, err := book.Read(strings.NewReader(strings.Join(raw, "\n")), "", 60, 20, false)
	if err != nil {
		t.Fatal(err)
	}
	var formatted []string
	for _, page := range b.Pages {
		formatted = append(formatted, page.Lines...)
	}
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

	got, _ := styleLine(comment, 0, nil, -1, 0)
	if want := textStyle.Render(comment); got != want {
		t.Fatalf("fenced comment %q style = %q, want body style %q", comment, got, want)
	}
}
