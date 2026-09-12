package tui

import (
	"strings"
	"testing"
)

// TestStartup_InitialWindowSizeDisplaysFirstPage reproduces issue #91: when
// opening a document with multiple headings near the beginning, the initial
// WindowSizeMsg must not anchor to the last heading and skip Page 1.
func TestStartup_InitialWindowSizeDisplaysFirstPage(t *testing.T) {
	fixture := strings.Join([]string{
		"# The Art of Reading",
		"",
		"## Table of Contents",
		"",
		"- [Chapter 1](#chapter-1)",
		"- [Chapter 2](#chapter-2)",
		"",
		"# Chapter 1",
		"",
		"Content of chapter 1.",
		"",
		"# Chapter 2",
		"",
		"Content of chapter 2.",
	}, "\n")

	path := writeTempFile(t, "startup-headings.md", fixture)
	m := NewModel(path)
	// Terminal 80x15: content width 72, content height 8 (15-7).
	m = applyWindowSize(m, 80, 15)

	if m.CurrentPage() != 0 {
		t.Fatalf("expected initial page 0 on startup, got page %d (showing %q)", m.CurrentPage(), pageText(m))
	}
	if !strings.Contains(pageText(m), "The Art of Reading") {
		t.Errorf("expected page 0 to contain '# The Art of Reading', got %q", pageText(m))
	}
}

func TestStartup_InitialWindowSizeDisplaysFirstPage_Minimal(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("# Heading 1\n\n")
	for i := 0; i < 6; i++ {
		sb.WriteString("filler\n\n")
	}
	sb.WriteString("# Heading 2\n\n")

	path := writeTempFile(t, "startup-minimal.md", sb.String())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 15)

	if m.CurrentPage() != 0 {
		t.Fatalf("expected initial page 0 on startup, got page %d (showing %q)", m.CurrentPage(), pageText(m))
	}
	if !strings.Contains(pageText(m), "Heading 1") {
		t.Errorf("expected page 0 to contain '# Heading 1', got %q", pageText(m))
	}
}

func TestStartup_DocumentStartingWithBodyText(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("Prologue paragraph 1.\n\n")
	for i := 0; i < 6; i++ {
		sb.WriteString("More prologue.\n\n")
	}
	sb.WriteString("# Chapter 1\n\n")
	sb.WriteString("Chapter 1 content.\n")

	path := writeTempFile(t, "startup-prologue.md", sb.String())
	m := NewModel(path)
	m = applyWindowSize(m, 80, 15)

	if m.CurrentPage() != 0 {
		t.Fatalf("expected initial page 0 on startup, got page %d (showing %q)", m.CurrentPage(), pageText(m))
	}
	if !strings.Contains(pageText(m), "Prologue") {
		t.Errorf("expected page 0 to contain 'Prologue', got %q", pageText(m))
	}
}
