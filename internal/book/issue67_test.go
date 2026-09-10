package book

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIndentedCodeDoesNotCreateAnchorsOrLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "issue-67.md")
	content := "# Intro\n    # Not a heading\n    [Not a link](#intro)\n[Real link](#intro)"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	b, err := NewBook(path, 80, 2)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := b.Anchors, map[string]int{"intro": 0}; !reflect.DeepEqual(got, want) {
		t.Errorf("anchors = %#v, want %#v", got, want)
	}

	var links int
	var found Link
	for _, page := range b.Pages {
		links += len(page.Links)
		if len(page.Links) == 1 {
			found = page.Links[0]
		}
	}
	if links != 1 {
		t.Fatalf("total links = %d, want 1", links)
	}
	if found.Label != "Real link" || found.Target != "intro" {
		t.Fatalf("link = %+v, want the real link to intro", found)
	}
}

func TestIndentedCodeDoesNotAnchorPagePositionAsHeading(t *testing.T) {
	layout := buildBookLayout([]string{
		"Body text",
		"    # Not a heading",
	}, 80, 3)

	if got, want := layout.pageRawLines[0], 0; got != want {
		t.Fatalf("page raw anchor = %d, want %d for the body line", got, want)
	}
}
