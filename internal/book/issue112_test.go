package book

import "testing"

func TestIssue112_PageForAnchorNonASCIIAndUnnormalizedFragments(t *testing.T) {
	content := "[Chapter](#第一章)\n\n" +
		"para one\n\npara two\n\npara three\n\npara four\n\n" +
		"## 第一章\n\n" +
		"para five\n\npara six\n\n" +
		"## Été\n\n" +
		"para seven\n\npara eight\n\n" +
		"## install_deps\n\n" +
		"para nine\n\npara ten\n\n" +
		"## Usage\n"
	path := writeTempFile(t, "non-ascii.md", content)
	b, err := NewBook(path, 60, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"第一章", "été", "Été", "install_deps", "Usage"} {
		if page := b.PageForAnchor(target); page <= 0 {
			t.Errorf("PageForAnchor(%q) = %d, want the heading's page (> 0)", target, page)
		}
	}
}
