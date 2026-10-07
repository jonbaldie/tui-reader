package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIssue99_NarrowResizeDoesNotPanic(t *testing.T) {
	content := "00000000 00000000000 \n0 000 000000\n"
	m := newTestModel(t, "doc.md", content)
	if m.Err() != nil {
		t.Fatalf("newTestModel: %v", m.Err())
	}

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("narrow resize panicked: %v", rec)
		}
	}()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 9, Height: 12})
	_ = updated.(Model)
}
