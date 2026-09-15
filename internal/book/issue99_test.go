package book

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func issue99Doc() []string {
	return []string{"00000000 00000000000 ", "0 000 000000"}
}

func TestIssue99_PaginateNarrowMultiLineDoesNotPanic(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Paginate panicked: %v", rec)
		}
	}()
	pages := Paginate(issue99Doc(), 5, 5)
	if len(pages) == 0 {
		t.Fatal("Paginate returned no pages")
	}
}

func TestIssue99_NewBookAndReflowDoNotPanic(t *testing.T) {
	path := writeIssue99Doc(t, strings.Join(issue99Doc(), "\n")+"\n")
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("NewBook/Reflow panicked: %v", rec)
		}
	}()
	b, err := NewBook(path, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	b.Reflow(5, 5)
	b.Reflow(60, 20)
	b.Reflow(5, 5)
}

func TestIssue99_WhitespaceAndHardBreakProvenance(t *testing.T) {
	cases := []struct {
		name  string
		raw   []string
		width int
	}{
		{"reported", issue99Doc(), 5},
		{"trailing-spaces", []string{"00000000 00000000000   ", "0 000 000000"}, 5},
		{"tabs", []string{"00000000\t00000000000 ", "0 000 000000"}, 5},
		{"collapsed-spaces", []string{"00000000  00000000000 ", "0  000  000000"}, 5},
		{"hard-broken", []string{"0000000000000000 ", "000000000000"}, 5},
		{"second-paragraph-indent", []string{"Lead.", "", "00000000 00000000000 ", "0 000 000000"}, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("panicked: %v", rec)
				}
			}()
			_ = Paginate(tc.raw, tc.width, 5)
			checkFormatterInvariants(t, tc.raw, tc.width)
		})
	}
}

func writeIssue99Doc(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
