package readerinteraction_test

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/jonbaldie/tui-reader/internal/tui"
	"strings"
	"testing"
)

func key(m tui.Model, k string) tui.Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	switch k {
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	}
	next, _ := m.Update(msg)
	return next.(tui.Model)
}
func makeModel(t *testing.T) tui.Model {
	t.Helper()
	b, err := book.Read(strings.NewReader("# Start\n\n[End](#end)\n\n"+strings.Repeat("filler words here\n\n", 12)+"# End\n"), "formal", 30, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	m := tui.NewModelFromBook(b)
	n, _ := m.Update(tea.WindowSizeMsg{Width: 34, Height: 12})
	return n.(tui.Model)
}
func TestFocusClearsOnMove(t *testing.T) {
	m := key(makeModel(t), "tab")
	if m.SelectedLink() < 0 {
		t.Fatal("fixture has no selectable link")
	}
	m = key(m, "right")
	if m.SelectedLink() != -1 {
		t.Fatal("page movement retained link focus")
	}
}
func TestBoundedEventSequences(t *testing.T) {
	actions := []string{"tab", "enter", "right", "left", "g", "G", "b", "resize", "q"}
	var visit func([]string, int)
	visit = func(trace []string, left int) {
		m := makeModel(t)
		for _, action := range trace {
			if action == "resize" {
				n, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 32})
				m = n.(tui.Model)
			} else {
				m = key(m, action)
			}
			b := m.BookRef()
			if m.CurrentPage() < 0 || m.CurrentPage() >= b.PageCount() {
				t.Fatal(trace, "page out of bounds")
			}
			if m.SelectedLink() < -1 || m.SelectedLink() >= len(b.Page(m.CurrentPage()).Links) {
				t.Fatal(trace, "selection out of bounds")
			}
			for _, p := range m.History() {
				if p < 0 || p >= b.PageCount() {
					t.Fatal(trace, "history out of bounds")
				}
			}
		}
		if left > 0 {
			for _, a := range actions {
				visit(append(append([]string{}, trace...), a), left-1)
			}
		}
	}
	t.Run(fmt.Sprint("depth-3"), func(t *testing.T) { visit(nil, 3) })
}
func TestFollowResizeBack(t *testing.T) {
	m := key(key(makeModel(t), "tab"), "enter")
	if len(m.History()) != 1 {
		t.Fatal("follow did not retain origin")
	}
	origin := m.BookRef().PositionForPage(m.History()[0])
	n, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 32})
	m = n.(tui.Model)
	expected := m.BookRef().PageForPosition(origin)
	m = key(m, "b")
	if m.CurrentPage() != expected || len(m.History()) != 0 {
		t.Fatal("resize/back lost origin")
	}
}
