package book

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssue132_PlainTextFilesDoNotCreateAnchors(t *testing.T) {
	for _, name := range []string{"events.txt", "events.log"} {
		t.Run(name, func(t *testing.T) {
			path := writeTempFile(t, name, "# Request Failed\nThe service returned an error.\n")
			b, err := NewBook(path, 40, 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.anchors) != 0 {
				t.Errorf("anchors = %#v, want none for plain text", b.anchors)
			}
			if page := b.PageForAnchor("request-failed"); page != -1 {
				t.Errorf("PageForAnchor(request-failed) = %d, want -1", page)
			}

			b.Reflow(20, 1)
			if len(b.anchors) != 0 {
				t.Errorf("anchors after reflow = %#v, want none for plain text", b.anchors)
			}
		})
	}
}

func TestIssue132_MarkdownAnchorsFollowDocumentLayout(t *testing.T) {
	fence := strings.Repeat("`", 3)
	lines := []string{
		"# Start",
		"Opening prose.",
		fence + "go",
		"# Hidden Fenced Heading",
		fence,
		"    # Hidden Indented Heading",
		"",
		"A longer paragraph with enough words to move later headings across pages when the book is reflowed.",
		"# Middle",
		"Middle section content.",
		"#### End & Beyond",
	}
	path := writeTempFile(t, "issue-132.md", strings.Join(lines, "\n"))
	b, err := NewBook(path, 80, 2)
	if err != nil {
		t.Fatal(err)
	}

	wantAnchors := map[string]int{
		"start":      0,
		"middle":     8,
		"end-beyond": 10,
	}
	if !reflect.DeepEqual(b.anchors, wantAnchors) {
		t.Fatalf("anchors = %#v, want %#v", b.anchors, wantAnchors)
	}

	checkHeadingPages := func() {
		t.Helper()
		for anchor, heading := range map[string]string{
			"start":      "Start",
			"middle":     "Middle",
			"end-beyond": "End",
		} {
			page := b.PageForAnchor(anchor)
			if page < 0 || page >= len(b.pages) {
				t.Errorf("PageForAnchor(%q) = %d, want a page in [0, %d)", anchor, page, len(b.pages))
				continue
			}
			found := false
			for _, line := range b.pages[page].Lines {
				if strings.Contains(line, heading) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("PageForAnchor(%q) = %d, whose lines %q do not contain heading %q", anchor, page, b.pages[page].Lines, heading)
			}
		}
	}
	checkHeadingPages()

	b.Reflow(14, 2)
	checkHeadingPages()
}
