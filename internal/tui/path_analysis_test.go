package tui

import "testing"

func TestPathAnalysis_InitHasNoStartupCommand(t *testing.T) {
	m := Model{}
	if cmd := m.Init(); cmd != nil {
		t.Fatalf("Init command = %v, want nil", cmd)
	}
}

func TestPathAnalysis_UnknownUpdateMessagePreservesModel(t *testing.T) {
	m := newTestModel(t, "path-analysis.md", "# Title\n\nBody\n")
	updated, cmd := m.Update(struct{}{})
	if cmd != nil {
		t.Fatalf("unknown message command = %v, want nil", cmd)
	}
	got, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}
	if got.book != m.book || got.currentPage != m.currentPage || got.selectedLink != m.selectedLink || got.err != m.err {
		t.Fatalf("unknown message changed model: before=%+v after=%+v", m, got)
	}
}
