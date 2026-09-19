package tui

import (
	"strings"
	"testing"
)

func TestIssue112_FollowCJKHeadingLink(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("[Chapter](#第一章) and [Control](#chapter-two)\n\n")
	for i := 0; i < 4; i++ {
		sb.WriteString("A short paragraph of filler text.\n\n")
	}
	sb.WriteString("## 第一章\n\n")
	for i := 0; i < 4; i++ {
		sb.WriteString("A short paragraph of filler text.\n\n")
	}
	sb.WriteString("## Chapter Two\n")
	path := writeTempFile(t, "min-cjk.md", sb.String())
	m := NewModel(path)
	m = applyWindowSize(m, 60, 14)

	want := m.book.PageForAnchor("第一章")
	if want <= 0 {
		t.Fatalf("heading 第一章 page = %d, want a later page", want)
	}
	m = pressKey(m, "tab")
	m = pressKey(m, "enter")
	if m.CurrentPage() != want {
		t.Errorf("after following [Chapter](#第一章), page = %d, want %d", m.CurrentPage(), want)
	}
}
