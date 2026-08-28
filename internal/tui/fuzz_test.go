package tui

import (
	"fmt"
	"strings"
	"testing"
)

// FuzzModelActions sends generated, valid reader actions through Model.Update
// after loading a generated document with internal links.
func FuzzModelActions(f *testing.F) {
	for _, seed := range [][]byte{
		{0, 1, 2, 3, 4},
		{255, 17, 42, 99, 8, 3},
		{7, 11, 13, 17, 19, 23, 29},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, seed []byte) {
		if len(seed) == 0 || len(seed) > 256 {
			t.Skip()
		}

		m := NewModel(writeTempFile(t, "fuzz.md", tuiFuzzDocument(seed)))
		if m.err != nil || m.book == nil {
			t.Fatalf("NewModel: book=%v err=%v", m.book, m.err)
		}

		for step, action := range seed {
			switch action % 8 {
			case 0:
				m = applyWindowSize(m, int(action%100), int(seed[(step+1)%len(seed)]%60))
			case 1:
				m = pressKey(m, "right")
			case 2:
				m = pressKey(m, "left")
			case 3:
				m = pressKey(m, "tab")
			case 4:
				m = pressKey(m, "shift+tab")
			case 5:
				m = pressKey(m, "enter")
			case 6:
				m = pressKey(m, "b")
			case 7:
				m = pressKey(m, "g")
			}
			checkFuzzModel(t, m, step)
		}
	})
}

func tuiFuzzDocument(seed []byte) string {
	sections := int(seed[0]%5) + 1
	var out strings.Builder
	out.WriteString("# Contents\n\n")
	for i := 0; i < sections; i++ {
		fmt.Fprintf(&out, "[Open section %d](#section-%d)\n", i+1, i+1)
	}
	out.WriteString("\n")
	for i := 0; i < sections; i++ {
		fmt.Fprintf(&out, "# Section %d\n\n", i+1)
		for j := 0; j < int(seed[(i+1)%len(seed)]%8)+1; j++ {
			out.WriteString("Unicode café 東京 reader content.\n")
		}
		out.WriteString("\n")
	}
	return out.String()
}

func checkFuzzModel(t *testing.T, m Model, step int) {
	t.Helper()
	if m.book == nil || len(m.book.Pages) == 0 {
		t.Fatalf("step %d: model has no usable book", step)
	}
	if m.currentPage < 0 || m.currentPage >= len(m.book.Pages) {
		t.Fatalf("step %d: page %d outside [0,%d)", step, m.currentPage, len(m.book.Pages))
	}
	page := m.book.Pages[m.currentPage]
	if m.selectedLink < -1 || m.selectedLink >= len(page.Links) {
		t.Fatalf("step %d: selection %d invalid for %d links", step, m.selectedLink, len(page.Links))
	}
	for index, historyPage := range m.history {
		if historyPage < 0 || historyPage >= len(m.book.Pages) {
			t.Fatalf("step %d: history %d has invalid page %d", step, index, historyPage)
		}
	}
}
