package book

import (
	"strings"
	"testing"
)

func TestIssue139_PlainTextLinksStayInert(t *testing.T) {
	content := "# Target\n\nSee [jump](#target) here.\n"
	txt := writeTempFile(t, "a.txt", content)
	md := writeTempFile(t, "a.md", content)

	txtBook, err := NewBook(txt, 60, 20)
	if err != nil {
		t.Fatal(err)
	}
	mdBook, err := NewBook(md, 60, 20)
	if err != nil {
		t.Fatal(err)
	}

	assertPlainTextLinksInert(t, txtBook)
	assertMarkdownLinkResolves(t, mdBook)

	txtBook.Reflow(40, 20)
	mdBook.Reflow(40, 20)

	assertPlainTextLinksInert(t, txtBook)
	assertMarkdownLinkResolves(t, mdBook)
}

func assertPlainTextLinksInert(t *testing.T, b *Book) {
	t.Helper()
	var links int
	var joined strings.Builder
	for _, page := range b.Pages {
		links += len(page.Links)
		joined.WriteString(strings.Join(page.Lines, "\n"))
		joined.WriteByte('\n')
	}
	if links != 0 {
		t.Errorf("plain-text Page.Links = %d, want 0", links)
	}
	if !strings.Contains(joined.String(), "[jump](#target)") {
		t.Errorf("plain-text pages missing literal link text:\n%s", joined.String())
	}
}

func assertMarkdownLinkResolves(t *testing.T, b *Book) {
	t.Helper()
	var links []Link
	for _, page := range b.Pages {
		links = append(links, page.Links...)
	}
	if len(links) == 0 {
		t.Fatal("markdown book has no Page.Links")
	}
	if page := b.PageForAnchor(links[0].Target); page < 0 {
		t.Errorf("PageForAnchor(%q) = %d, want a resolved page", links[0].Target, page)
	}
}
