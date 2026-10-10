package book

import (
	"strings"
	"testing"
)

func tildeFenceLines() []string {
	return []string{
		"[Install](#install)",
		"",
		"# Install",
		"",
		"Real chapter.",
		"",
		"~~~sh",
		"# install",
		"make install",
		"~~~",
	}
}

func TestIssue178_TildeFenceDoesNotHijackAnchors(t *testing.T) {
	anchors := ExtractAnchors(tildeFenceLines())
	if got, ok := anchors["install"]; !ok || got != 2 {
		t.Fatalf("anchors[install] = %d (ok=%v), want 2 (# Install); anchors=%#v", got, ok, anchors)
	}
}

func TestIssue178_TildeFenceLinesAreNotMerged(t *testing.T) {
	formatted := FormatParagraphs(tildeFenceLines(), 60)
	for i, line := range formatted {
		if strings.Contains(line, "~~~") && (strings.Contains(line, "# install") || strings.Contains(line, "make install")) {
			t.Fatalf("line %d joins code with tilde fence: %q", i, line)
		}
	}
}

func TestIssue178_InstallLinkNavigatesToRealHeading(t *testing.T) {
	path := writeTempFile(t, "tilde-fence.md", strings.Join(tildeFenceLines(), "\n")+"\n")
	b, err := NewBook(path, 60, 6)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.PageForAnchor("install"), b.PageForRawLine(2); got != want {
		t.Fatalf("PageForAnchor(install) = %d, want page %d of # Install", got, want)
	}
}

func TestIssue178_FenceClosesOnlyWithMatchingCharacter(t *testing.T) {
	lines := []string{
		"~~~",
		"```",
		"# inside tilde",
		"~~~",
		"```",
		"# inside backtick",
		"~~~",
		"```",
		"# Real",
	}
	anchors := ExtractAnchors(lines)
	for _, fake := range []string{"inside-tilde", "inside-backtick"} {
		if _, ok := anchors[fake]; ok {
			t.Fatalf("%s should not be an anchor: %#v", fake, anchors)
		}
	}
	if got, ok := anchors["real"]; !ok || got != 8 {
		t.Fatalf("anchors[real] = %d (ok=%v), want 8; %#v", got, ok, anchors)
	}
}

func TestIssue178_TildeInfoStringMayContainBackticks(t *testing.T) {
	anchors := ExtractAnchors([]string{"~~~ `weird` info", "# fake", "~~~", "# Real"})
	if _, ok := anchors["fake"]; ok {
		t.Fatalf("fake should not be an anchor: %#v", anchors)
	}
}
