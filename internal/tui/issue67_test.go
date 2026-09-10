package tui

import "testing"

func TestIndentedCodeIsNotStyledAsHeading(t *testing.T) {
	got, _ := styleLine("    # Not a heading", 0, nil, -1, 0)
	if want := textStyle.Render("    # Not a heading"); got != want {
		t.Fatalf("indented code style = %q, want regular text style %q", got, want)
	}

	got, _ = styleLine("# Real heading", 0, nil, -1, 0)
	if want := headingStyle.Render("# Real heading"); got != want {
		t.Fatalf("heading style = %q, want heading style %q", got, want)
	}
}
