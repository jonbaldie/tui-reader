package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIssue99_NarrowResizeDoesNotPanic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	content := "00000000 00000000000 \n0 000 000000\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("NewModel: %v", m.Err())
	}

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("narrow resize panicked: %v", rec)
		}
	}()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 9, Height: 12})
	_ = updated.(Model)
}
