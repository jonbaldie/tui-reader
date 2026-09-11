package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestIssue82_OverwideLinkSelectionDesync(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := `[This is an extraordinarily long link label that exceeds width](#target-1)

[Short Link](#target-2)

# Target 1

Target 1 body content.

# Target 2

Target 2 body content.
`
	path := writeTempFile(t, "issue82.md", doc)
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	// Terminal dimensions 30x20 gives content width 25
	m = applyWindowSize(m, 30, 20)

	// Verify page 0 has 2 links
	page0 := m.book.Pages[0]
	if len(page0.Links) != 2 {
		t.Fatalf("expected 2 links on page 0, got %d", len(page0.Links))
	}
	if page0.Links[0].Target != "target-1" || page0.Links[1].Target != "target-2" {
		t.Fatalf("unexpected links: %+v", page0.Links)
	}

	// Press Tab to select link 0
	m = pressKey(m, "tab")
	if m.selectedLink != 0 {
		t.Fatalf("expected selectedLink=0, got %d", m.selectedLink)
	}

	view0 := m.View()

	// Short Link should NOT be styled with selectedLinkStyle when link 0 is selected
	selectedShortLink := selectedLinkStyle.Render("Short Link")
	if strings.Contains(view0, selectedShortLink) {
		t.Errorf("desync detected: Short Link is styled as selected when link 0 (over-wide link) is selected")
	}

	// Press Tab to select link 1 (Short Link)
	m = pressKey(m, "tab")
	if m.selectedLink != 1 {
		t.Fatalf("expected selectedLink=1, got %d", m.selectedLink)
	}

	view1 := m.View()

	// Short Link SHOULD be styled with selectedLinkStyle when link 1 is selected
	if !strings.Contains(view1, selectedShortLink) {
		t.Errorf("Short Link should be styled as selected when link 1 is selected")
	}

	// When Short Link (link 1) is selected, pressing Enter navigates to Target 2
	mFollow := pressKey(m, "enter")
	// Target 2 should be in current page view
	followView := mFollow.View()
	if !strings.Contains(followView, "Target 2 body content") {
		t.Errorf("following link 1 should navigate to Target 2, got view:\n%s", followView)
	}
}

func TestIssue82_OverwideLinkFollowTarget(t *testing.T) {
	doc := `[This is an extraordinarily long link label that exceeds width](#target-1)

[Short Link](#target-2)

# Target 1

Target 1 body content.

# Target 2

Target 2 body content.
`
	path := writeTempFile(t, "issue82_follow.md", doc)
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	m = applyWindowSize(m, 30, 20)

	// Press Tab to select link 0 (over-wide link)
	m = pressKey(m, "tab")
	if m.selectedLink != 0 {
		t.Fatalf("expected selectedLink=0, got %d", m.selectedLink)
	}

	// Press Enter to follow link 0
	mFollow := pressKey(m, "enter")
	followView := mFollow.View()
	if !strings.Contains(followView, "Target 1 body content") {
		t.Errorf("following link 0 should navigate to Target 1, got view:\n%s", followView)
	}
}

func TestIssue82_OverwideLinkBetweenNormalLinks(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(profile) })

	doc := `[First Short](#first)

[This is an extraordinarily long link label that exceeds width](#overwide)

[Second Short](#second)

# First

First target content.

# Overwide

Overwide target content.

# Second

Second target content.
`
	path := writeTempFile(t, "issue82_between.md", doc)
	m := NewModel(path)
	if m.Err() != nil {
		t.Fatalf("failed to open fixture: %v", m.Err())
	}

	m = applyWindowSize(m, 30, 20)
	m = pressKey(m, "home")
	selectedFirst := selectedLinkStyle.Render("First Short")
	selectedSecond := selectedLinkStyle.Render("Second Short")

	m = pressKey(m, "tab")
	if m.selectedLink != 0 {
		t.Fatalf("expected selectedLink=0, got %d", m.selectedLink)
	}
	v0 := m.View()
	if !strings.Contains(v0, selectedFirst) {
		t.Errorf("link 0 should visually highlight First Short")
	}
	if strings.Contains(v0, selectedSecond) {
		t.Errorf("link 0 should not visually highlight Second Short")
	}

	// Select link 1: Over-wide link
	m = pressKey(m, "tab")
	if m.selectedLink != 1 {
		t.Fatalf("expected selectedLink=1, got %d", m.selectedLink)
	}
	v1 := m.View()
	if strings.Contains(v1, selectedFirst) {
		t.Errorf("link 1 should not visually highlight First Short")
	}
	if strings.Contains(v1, selectedSecond) {
		t.Errorf("link 1 should not visually highlight Second Short")
	}

	// Select link 2: Second Short
	m = pressKey(m, "tab")
	if m.selectedLink != 2 {
		t.Fatalf("expected selectedLink=2, got %d", m.selectedLink)
	}
	v2 := m.View()
	if strings.Contains(v2, selectedFirst) {
		t.Errorf("link 2 should not visually highlight First Short")
	}
	if !strings.Contains(v2, selectedSecond) {
		t.Errorf("link 2 should visually highlight Second Short")
	}

	// Reverse with shift+tab back to link 1
	m = pressKey(m, "shift+tab")
	if m.selectedLink != 1 {
		t.Fatalf("expected selectedLink=1 after shift+tab, got %d", m.selectedLink)
	}
	v1Rev := m.View()
	if strings.Contains(v1Rev, selectedFirst) || strings.Contains(v1Rev, selectedSecond) {
		t.Errorf("reverse selection to link 1 should not highlight short links")
	}

	// Enter on link 1 navigates to Overwide target
	mFollowOverwide := pressKey(m, "enter")
	if !strings.Contains(mFollowOverwide.View(), "Overwide target content") {
		t.Errorf("expected follow to reach Overwide target")
	}
}
