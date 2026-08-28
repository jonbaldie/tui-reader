package book_test

import (
	"strings"
	"testing"
)

// FuzzSupportedHeadingAnchors compares the documented, deliberately narrow
// heading subset against a separately specified fragment rule. The domain is
// ASCII letters/digits separated by whitespace; punctuation and presentation
// choices are intentionally outside this comparison.
func FuzzSupportedHeadingAnchors(f *testing.F) {
	for _, seed := range []string{"Chapter One", "Version 2", "Mixed CASE 42"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed string) {
		heading := supportedHeading(seed)
		if heading == "" {
			t.Skip()
		}
		anchor := referenceAnchor(heading)
		content := "# " + heading + "\n\n[go](#" + anchor + ")\n"
		b := bookFromContent(t, content, 60, 10)
		if got := b.Anchors[anchor]; got != 0 {
			t.Fatalf("anchor %q line = %d, want 0; anchors=%v", anchor, got, b.Anchors)
		}
		if page := b.PageForAnchor(anchor); page != 0 {
			t.Fatalf("PageForAnchor(%q) = %d, want 0", anchor, page)
		}
	})
}

func supportedHeading(seed string) string {
	var out strings.Builder
	space := false
	for _, r := range seed {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
			out.WriteRune(r)
			space = false
		} else if !space && out.Len() > 0 {
			out.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(out.String())
}

func referenceAnchor(heading string) string {
	return strings.ToLower(strings.Join(strings.Fields(heading), "-"))
}
