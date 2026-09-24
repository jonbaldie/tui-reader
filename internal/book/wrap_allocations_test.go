package book

import (
	"strings"
	"testing"
)

func TestWrapParagraphAllocationBudget(t *testing.T) {
	paragraph := strings.Repeat("She crossed the quiet courtyard and opened the café door. ", 4)
	allocations := testing.AllocsPerRun(20, func() {
		lines := WrapLines([]string{paragraph}, 60)
		if len(lines) != 4 {
			t.Fatalf("got %d lines, want 4", len(lines))
		}
	})
	t.Logf("allocations per paragraph: %.0f", allocations)
	if allocations > 12 {
		t.Fatalf("wrapping allocated %.0f times; ceiling is 12", allocations)
	}
}
