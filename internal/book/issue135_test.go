package book

import (
	"strings"
	"testing"
)

func TestIssue135_FindFenceBlockEndIgnoresIndentedDelimiters(t *testing.T) {
	lines := []string{
		"```go",
		"    ```",
		"    ```python",
		"```",
		"after",
	}
	end := findFenceBlockEnd(lines, 0)
	if end != 4 {
		t.Fatalf("findFenceBlockEnd = %d, want 4", end)
	}
}

func TestIssue135_IndentedFenceInsideFencedCodeBlock(t *testing.T) {
	content := strings.Join([]string{
		"```",
		"    ```",
		"```",
		"",
		"# Real Heading",
		"",
		"Jump to [real](#real-heading).",
		"",
		"```",
		"# Fake Heading In Code",
		"```",
	}, "\n")

	rawLines := strings.Split(content, "\n")
	formatted := FormatParagraphs(rawLines, 80)
	anchors := ExtractAnchors(rawLines)

	path := writeTempFile(t, "repro.md", content)
	b, err := NewBook(path, 80, 20)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Anchors correctly records # Real Heading (line 4) and excludes fake heading in code (line 10)
	if _, ok := anchors["real-heading"]; !ok {
		t.Errorf("anchors missing 'real-heading'")
	}
	if _, ok := anchors["fake-heading-in-code"]; ok {
		t.Errorf("anchors should not contain 'fake-heading-in-code'")
	}

	// 2. But findFenceBlockEnd prematurely closed at line 1, inverting the formatter's blocks:
	// - '# Real Heading' was indented and rendered as code: "    # Real Heading"
	// - '# Fake Heading In Code' was rendered outside code as a heading: "# Fake Heading In Code"
	for _, l := range formatted {
		if l == "    # Real Heading" {
			t.Errorf("BUG: '# Real Heading' was rendered inside code as an indented code line!")
		}
		if l == "# Fake Heading In Code" {
			t.Errorf("BUG: '# Fake Heading In Code' was rendered outside code as a heading!")
		}
	}

	// 3. Link navigation to #real-heading resolves to the code block page
	page := b.PageForAnchor("real-heading")
	if page < 0 {
		t.Errorf("PageForAnchor('real-heading') = %d, want >= 0", page)
	}
}
