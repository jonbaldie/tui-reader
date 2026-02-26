package tui

import (
	"strings"
	"testing"
)

func longProseDoc() string {
	// Simulate the manuscript: long lines that wrap heavily
	return `"—I'm telling you, the flux ratio is within tolerance. It's point-three above baseline, which on a system this size is basically a rounding error."
"It's not a rounding error, Joss."
"Sorry—Commander Rist. Forgot we were being formal today."
"We're on an open channel. You're reporting drive status to the bridge. So yes, we're being formal." I kept my voice level because seventeen bridge officers were very carefully pretending not to listen. "Give me a cause for the deviation."
She took a beat. My sister always took a beat before answering a direct question—not because she didn't know, but because she was deciding what to leave out.
"Coolant microvariance in the tertiary loop. We swapped a gasket six hours ago, new seal's still bedding in. It'll settle before we hit ignition threshold." Another beat. "I'd bet my commission on it."
"You don't have a commission. You're civilian-contract."
"Right. So I'd be betting nothing. That's how confident I am."
A couple of the junior officers at the sensor banks smiled. I didn't. I flagged the reading, stamped it with my authentication code, and closed the channel.
Look—I could have walked down there. Checked the gasket myself. Part of me wanted to, the same part that always wanted to double-check Joss's work, not because she wasn't brilliant but because she was my little sister and I'd been double-checking her work since she was six years old and building railgun models out of kitchen parts.
`
}

func TestRendering_FirstLineVisible(t *testing.T) {
	path := writeTempFile(t, "prose.txt", longProseDoc())
	m := NewModel(path)
	// Simulate a wide, tall terminal
	m = applyWindowSize(m, 200, 50)

	view := m.View()
	lines := strings.Split(view, "\n")

	// The first line of content should contain the opening quote
	foundOpening := false
	for _, line := range lines {
		if strings.Contains(line, "flux ratio is within tolerance") {
			foundOpening = true
			break
		}
	}
	if !foundOpening {
		// Print first 20 lines for debugging
		t.Log("First 20 lines of rendered view:")
		for i := 0; i < 20 && i < len(lines); i++ {
			t.Logf("  [%d] %q", i, lines[i])
		}
		t.Error("BUG: first line of content not visible in rendered view")
	}
}

func TestRendering_ViewFitsTerminal(t *testing.T) {
	path := writeTempFile(t, "prose.txt", longProseDoc())
	m := NewModel(path)

	sizes := []struct{ w, h int }{
		{80, 24},
		{200, 50},
		{120, 30},
		{250, 60},
	}

	for _, s := range sizes {
		m = applyWindowSize(m, s.w, s.h)
		view := m.View()
		lines := strings.Split(view, "\n")

		// View should not exceed terminal height
		if len(lines) > s.h {
			t.Errorf("at %dx%d: view has %d lines, exceeds terminal height %d",
				s.w, s.h, len(lines), s.h)
		}

		// No line should exceed terminal width in rune count
		// (ANSI escape codes inflate byte count, so measure runes after stripping)
		for i, line := range lines {
			stripped := stripAnsi(line)
			runeCount := len([]rune(stripped))
			if runeCount > s.w+2 { // small tolerance for rounding
				t.Errorf("at %dx%d: line %d is %d runes wide, exceeds width",
					s.w, s.h, i, runeCount)
			}
		}
	}
}

func TestRendering_ContentNotClipped(t *testing.T) {
	path := writeTempFile(t, "prose.txt", longProseDoc())
	m := NewModel(path)
	m = applyWindowSize(m, 200, 50)

	view := m.View()

	// "to leave out" should NOT be the first content line
	lines := strings.Split(view, "\n")
	firstContentLine := ""
	for _, line := range lines {
		stripped := strings.TrimSpace(stripAnsi(line))
		// Skip empty lines, the title, and the divider
		if stripped == "" || strings.Contains(stripped, "────") {
			continue
		}
		if strings.Contains(stripped, "Prose") {
			continue // title
		}
		firstContentLine = stripped
		break
	}

	if strings.HasPrefix(firstContentLine, "to leave out") {
		t.Errorf("BUG: content starts at 'to leave out' — first lines are clipped. First content: %q", firstContentLine)
	}
	if !strings.Contains(firstContentLine, "flux ratio") {
		t.Errorf("expected first content to contain 'flux ratio', got: %q", firstContentLine)
	}
}

// stripAnsi removes ANSI escape sequences from a string
func stripAnsi(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until we find the terminator
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++ // skip the terminator
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
