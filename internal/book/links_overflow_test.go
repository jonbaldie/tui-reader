package book

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Regression for #39: a markdown link whose [label](#target) markup is wider
// than the page width must not produce a display line longer than the width.
// The link markup is hard-broken at the character level, like any over-long
// word, so the width invariant holds. The link stays attached (selectable via
// tab, followable via enter) through its raw-source-line provenance.
func TestLinkOverflow_WideMarkupWrapsToWidth(t *testing.T) {
	raw := []string{"# target", "", "[This is a long link label](#target)"}
	pages := Paginate(raw, 20, 20)

	for pi, page := range pages {
		for li, line := range page.Lines {
			if rl := utf8.RuneCountInString(line); rl > 20 {
				t.Errorf("page %d line %d: %d runes > 20: %q", pi, li, rl, line)
			}
		}
	}

	// The link stays attached and resolvable through NewBook (Paginate alone
	// does not attach links; NewBook is the entry point that does).
	content := strings.Join(raw, "\n") + "\n"
	path := writeTempFile(t, "wide.md", content)
	b, err := NewBook(path, 20, 20)
	if err != nil {
		t.Fatalf("NewBook: %v", err)
	}
	var found bool
	for _, page := range b.Pages {
		for _, lnk := range page.Links {
			if lnk.Target == "target" {
				found = true
			}
		}
	}
	if !found {
		t.Error("over-wide link was not attached to any page")
	}
	if dest := b.PageForAnchor("target"); dest < 0 {
		t.Errorf("PageForAnchor(target) = %d, want >= 0", dest)
	}
}

// Short links that fit within the width are kept whole on one display line, so
// their label is fully visible and stylable (the #23 behaviour).
func TestLinkOverflow_ShortLinkStaysWhole(t *testing.T) {
	raw := []string{"# target", "", "[go](#target)"}
	pages := Paginate(raw, 20, 20)

	var whole bool
	for _, page := range pages {
		for _, line := range page.Lines {
			if strings.Contains(line, "[go](#target)") {
				whole = true
			}
		}
	}
	if !whole {
		t.Error("short link markup was not kept whole on a display line")
	}
}