package book

import (
	"strings"
	"testing"
)

func logFileContent() string {
	return "" +
		"2026-09-19 10:00:01 INFO server started\n" +
		"2026-09-19 10:00:02 WARN disk 91%\n" +
		"2026-09-19 10:00:03 ERROR request failed\n"
}

func pageText(p Page) string {
	return strings.Join(p.Lines, "\n")
}

// TestIssue114_PlainTextLogLinesStaySeparate reproduces issue #114: consecutive
// lines in a .log file must stay independent instead of joining into one
// Markdown paragraph.
func TestIssue114_PlainTextLogLinesStaySeparate(t *testing.T) {
	path := writeTempFile(t, "app.log", logFileContent())

	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Pages) == 0 {
		t.Fatal("expected at least 1 page")
	}

	joined := pageText(b.Pages[0])
	want := []string{
		"2026-09-19 10:00:01 INFO server started",
		"2026-09-19 10:00:02 WARN disk 91%",
		"2026-09-19 10:00:03 ERROR request failed",
	}
	for _, line := range want {
		if !strings.Contains(joined, line) {
			t.Errorf("page lines =\n%s\nmissing independent log line %q", joined, line)
		}
	}
	if strings.Contains(joined, "started 2026-09-19") {
		t.Errorf("consecutive log lines were merged:\n%s", joined)
	}
}

func TestIssue114_PlainTextTabsFileNotCollapsed(t *testing.T) {
	content := "col1\tcol2\tcol3\n\tindented with tab\nnormal\n"
	path := writeTempFile(t, "tabs.txt", content)

	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	joined := pageText(b.Pages[0])
	if strings.Contains(joined, "col3 indented") || strings.Contains(joined, "tab normal") {
		t.Errorf("tabs.txt lines were collapsed:\n%s", joined)
	}
	if !strings.Contains(joined, "col1") || !strings.Contains(joined, "indented with tab") || !strings.Contains(joined, "normal") {
		t.Errorf("tabs.txt missing source lines:\n%s", joined)
	}
}

func TestIssue114_PlainTextLongLineStillWraps(t *testing.T) {
	long := strings.Repeat("word ", 20) + "end"
	path := writeTempFile(t, "wide.log", long+"\nshort\n")

	b, err := NewBook(path, 20, 10)
	if err != nil {
		t.Fatal(err)
	}
	lines := b.Pages[0].Lines
	if len(lines) < 3 {
		t.Fatalf("expected wrapped long line plus short line, got %q", lines)
	}
	if strings.Contains(strings.Join(lines, " "), "end short") && !strings.Contains(strings.Join(lines, "\n"), "short") {
		t.Errorf("short line was merged into wrapped paragraph: %q", lines)
	}
	foundShort := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "short" {
			foundShort = true
		}
		if stringWidth(line) > 20 {
			t.Errorf("display line %q wider than 20", line)
		}
	}
	if !foundShort {
		t.Errorf("expected independent short line, got %q", lines)
	}
}

func TestIssue114_MarkdownStillSoftWraps(t *testing.T) {
	content := "This is one Markdown paragraph deliberately split across\nphysical source lines without a blank line.\n"
	path := writeTempFile(t, "soft.md", content)

	b, err := NewBook(path, 120, 10)
	if err != nil {
		t.Fatal(err)
	}
	joined := pageText(b.Pages[0])
	want := "This is one Markdown paragraph deliberately split across physical source lines without a blank line."
	if !strings.Contains(joined, want) {
		t.Errorf("markdown page =\n%s\nwant soft-wrapped paragraph %q", joined, want)
	}
}

func TestIssue114_PlainTextProvenanceMapsToEachSourceLine(t *testing.T) {
	raw := []string{
		"2026-09-19 10:00:01 INFO server started",
		"2026-09-19 10:00:02 WARN disk 91%",
		"2026-09-19 10:00:03 ERROR request failed",
	}
	formatted := formatPlainTextWithProvenance(raw, 80)
	if len(formatted) != 3 {
		t.Fatalf("formatted lines = %d, want 3: %v", len(formatted), formatted)
	}
	for i, fl := range formatted {
		if fl.raw != i {
			t.Errorf("line %d raw = %d, want %d", i, fl.raw, i)
		}
		if fl.text != raw[i] {
			t.Errorf("line %d text = %q, want %q", i, fl.text, raw[i])
		}
	}
}

func TestIssue114_PlainTextReflowKeepsLinesSeparate(t *testing.T) {
	path := writeTempFile(t, "app.log", logFileContent())
	b, err := NewBook(path, 80, 10)
	if err != nil {
		t.Fatal(err)
	}
	b.Reflow(40, 8)
	joined := pageText(b.Pages[0])
	if strings.Contains(joined, "started 2026-09-19") {
		t.Errorf("reflow merged log lines:\n%s", joined)
	}
	if !strings.Contains(joined, "INFO server started") {
		t.Errorf("reflow lost first log line:\n%s", joined)
	}
}
