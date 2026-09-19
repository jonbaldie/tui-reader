package tui

import "testing"

func TestIssue111_FencedCodeCommentDoesNotHijackLink(t *testing.T) {
	content := "# Guide\n\n[Usage](#usage)\n\n## Usage\n\nReal section.\n\n## Example\n\n```sh\n# usage\ntool --help\n```\n"
	path := writeTempFile(t, "min-fence.md", content)
	m := NewModel(path)
	m = applyWindowSize(m, 60, 14)

	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	if m.CurrentPage() != 0 {
		t.Errorf("after following [Usage](#usage), page = %d, want 0", m.CurrentPage())
	}
	for _, page := range m.book.Pages {
		for i, line := range page.Lines {
			if line == "# usage" || line == "    # usage" {
				if got, want := styleLineText(line, i), textStyle.Render(line); got != want {
					t.Errorf("fenced comment styled as %q, want text style %q", got, want)
				}
			}
		}
	}
}

func styleLineText(line string, i int) string {
	got, _ := styleLine(line, i, nil, -1, 0)
	return got
}
