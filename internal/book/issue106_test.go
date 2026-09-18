package book

import (
	"testing"
)

func TestIssue106_SoftWrappedParagraphSpanningLinkAttached(t *testing.T) {
	raw := []string{
		"[link",
		"label](#target)",
	}
	pages := AttachLinks(Paginate(raw, 80, 20), raw, 80, 20)
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	if len(pages[0].Links) == 0 {
		t.Fatalf("expected page 0 to have links, got 0; links dropped across soft-wrapped line boundary")
	}
	if pages[0].Links[0].Target != "target" {
		t.Fatalf("expected link target to be 'target', got %q", pages[0].Links[0].Target)
	}
	if pages[0].Links[0].Label != "link label" {
		t.Fatalf("expected link label to be 'link label', got %q", pages[0].Links[0].Label)
	}
}

func TestIssue106_OriginalIssueReproduction(t *testing.T) {
	raw := []string{
		"# Section",
		"Here is a [link",
		"label](#target) across lines.",
		"",
		"# Target",
		"Target content.",
	}
	pages := AttachLinks(Paginate(raw, 80, 20), raw, 80, 20)
	if len(pages) < 1 {
		t.Fatalf("expected at least 1 page, got %d", len(pages))
	}
	if len(pages[0].Links) == 0 {
		t.Fatalf("expected page 0 to have links, got 0; links dropped across soft-wrapped line boundary")
	}
	if pages[0].Links[0].Target != "target" {
		t.Fatalf("expected link target to be 'target', got %q", pages[0].Links[0].Target)
	}
	if pages[0].Links[0].Label != "link label" {
		t.Fatalf("expected link label to be 'link label', got %q", pages[0].Links[0].Label)
	}
}

func TestIssue106_MultipleLinksSpanningLinesInSameParagraph(t *testing.T) {
	raw := []string{
		"Intro text [first",
		"link](#first-target) middle text",
		"and [second",
		"link](#second-target) end text.",
	}
	pages := AttachLinks(Paginate(raw, 80, 20), raw, 80, 20)
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	if len(pages[0].Links) != 2 {
		t.Fatalf("expected 2 links on page 0, got %d: %+v", len(pages[0].Links), pages[0].Links)
	}
	if pages[0].Links[0].Target != "first-target" || pages[0].Links[0].Label != "first link" {
		t.Errorf("link 0 mismatch: %+v", pages[0].Links[0])
	}
	if pages[0].Links[1].Target != "second-target" || pages[0].Links[1].Label != "second link" {
		t.Errorf("link 1 mismatch: %+v", pages[0].Links[1])
	}
}
