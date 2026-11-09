package book

import "testing"

func TestRead_WrappedSourcePositionsSurviveReflow(t *testing.T) {
	b := readBook(t, "one two three four five six seven eight nine ten", true, 10, 2)
	// Display lines are: "one two", "three four", "five six", "seven",
	// "eight nine", "ten". Offsets count non-space runes, not display columns.
	for page, want := range []Position{{Raw: 0, Offset: 0}, {Raw: 0, Offset: 15}, {Raw: 0, Offset: 27}} {
		if got := b.PositionForPage(page); got != want {
			t.Errorf("PositionForPage(%d) = %+v, want %+v", page, got, want)
		}
	}
	for _, tc := range []struct{ offset, page int }{
		{-1, 0}, {0, 0}, {1, 0}, {6, 0}, {14, 0}, {15, 1},
		{22, 1}, {26, 1}, {27, 2}, {36, 2}, {99, 2},
	} {
		if got := b.PageForPosition(Position{Raw: 0, Offset: tc.offset}); got != tc.page {
			t.Errorf("PageForPosition(offset %d) = %d, want %d", tc.offset, got, tc.page)
		}
	}

	pos := b.PositionForPage(1)
	b.Reflow(15, 1)
	// "four five six" starts after the first eleven non-space runes.
	if got := b.PageForPosition(pos); got != 1 {
		t.Errorf("reflowed position %+v is on page %d, want 1", pos, got)
	}
	if got := b.PositionForPage(1); got != (Position{Raw: 0, Offset: 11}) {
		t.Errorf("reflowed page 1 position = %+v, want raw 0 offset 11", got)
	}
}

func TestRead_PositionQueriesClampToDocument(t *testing.T) {
	b := readBook(t, "alpha\n\nbeta\n\n# End\n\nlast", true, 80, 2)
	for _, tc := range []struct{ page, raw int }{
		{-1, 0}, {0, 0}, {1, 2}, {2, 4}, {3, 6}, {4, 6}, {99, 6},
	} {
		want := Position{Raw: tc.raw}
		if got := b.PositionForPage(tc.page); got != want {
			t.Errorf("PositionForPage(%d) = %+v, want %+v", tc.page, got, want)
		}
	}
	for _, tc := range []struct{ raw, page int }{
		{-1, 0}, {0, 0}, {2, 1}, {4, 2}, {6, 3}, {7, 3}, {99, 3},
	} {
		if got := b.PageForRawLine(tc.raw); got != tc.page {
			t.Errorf("PageForRawLine(%d) = %d, want %d", tc.raw, got, tc.page)
		}
		if got := b.PageForPosition(Position{Raw: tc.raw, Offset: 99}); got != tc.page {
			t.Errorf("PageForPosition(raw %d, offset 99) = %d, want %d", tc.raw, got, tc.page)
		}
	}
}
