package book

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRead_BuildsBookFromReader(t *testing.T) {
	content := "\xef\xbb\xbf# Chapter One\r\n\r\nSee [the index](#index).\r\n\r\n# Index\r\n"
	b, err := Read(strings.NewReader(content), "Stream Title", 60, 10, false)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if b.Title != "Stream Title" {
		t.Errorf("Title = %q, want %q", b.Title, "Stream Title")
	}
	if b.RawLines[0] != "# Chapter One" {
		t.Errorf("first raw line = %q, want BOM-free heading", b.RawLines[0])
	}
	if got, want := b.RawLines[1], ""; got != want {
		t.Errorf("second raw line = %q, want %q", got, want)
	}
	if got, want := b.RawLines[2], "See [the index](#index)."; got != want {
		t.Errorf("third raw line = %q, want %q", got, want)
	}
	if _, ok := b.Anchors["chapter-one"]; !ok {
		t.Error("expected chapter-one anchor")
	}
	if b.PageForAnchor("index") < 0 {
		t.Error("expected index anchor to have a page")
	}
	if len(b.Pages) == 0 {
		t.Fatal("expected a laid out page")
	}
	var foundIndexLink bool
	for _, page := range b.Pages {
		for _, link := range page.Links {
			if link.Target == "index" {
				foundIndexLink = true
			}
		}
	}
	if !foundIndexLink {
		t.Error("expected the index link to be attached to a page")
	}
}

func TestRead_RejectsInvalidUTF8(t *testing.T) {
	_, err := Read(bytes.NewReader([]byte{0xff, 0xfe}), "Invalid", 60, 10, false)
	if err == nil {
		t.Fatal("expected invalid UTF-8 error")
	}
}

func TestNewBookMatchesReadForFileFormat(t *testing.T) {
	for _, tc := range []struct {
		name      string
		plainText bool
	}{
		{name: "my-book.md"},
		{name: "my-book.txt", plainText: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := "\xef\xbb\xbf# Heading\r\n\r\nA paragraph.\r\n"
			path := filepath.Join(t.TempDir(), tc.name)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			fromFile, err := NewBook(path, 60, 10)
			if err != nil {
				t.Fatalf("NewBook: %v", err)
			}
			fromReader, err := Read(strings.NewReader(content), "My Book", 60, 10, tc.plainText)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if !reflect.DeepEqual(fromFile, fromReader) {
				t.Fatal("NewBook result differs from Read for the same document")
			}
		})
	}
}
