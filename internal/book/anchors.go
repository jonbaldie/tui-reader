package book

import (
	"regexp"
	"strings"
)

var (
	// Markdown headings: # Heading, ## Heading, etc.
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	// Markdown links: [text](#anchor)
	linkRegex = regexp.MustCompile(`\[([^\]]+)\]\(#([^)]+)\)`)
)

// ExtractAnchors scans raw lines for headings and returns a map of
// normalized anchor names to their line indices.
func ExtractAnchors(lines []string) map[string]int {
	anchors := make(map[string]int)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if m := headingRegex.FindStringSubmatch(trimmed); m != nil {
			anchor := NormalizeAnchor(m[2])
			anchors[anchor] = i
		}
	}
	return anchors
}

// NormalizeAnchor converts heading text to a URL-fragment style anchor.
// "Chapter 1: Introduction" -> "chapter-1-introduction"
func NormalizeAnchor(text string) string {
	text = strings.ToLower(text)
	// Remove non-alphanumeric chars except spaces and hyphens
	var b strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	// Replace spaces with hyphens
	result = strings.ReplaceAll(result, " ", "-")
	// Collapse multiple hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	return result
}

// ExtractLinks finds markdown-style internal links in a line of text.
func ExtractLinks(line string) []Link {
	matches := linkRegex.FindAllStringSubmatch(line, -1)
	if matches == nil {
		return nil
	}
	var links []Link
	for _, m := range matches {
		links = append(links, Link{
			Label:  m[1],
			Target: m[2],
		})
	}
	return links
}

// AttachLinks scans pages for internal links and attaches them.
func AttachLinks(pages []Page, rawLines []string, width, height int) []Page {
	// First, build a mapping from wrapped line index to original line index
	wrapped := WrapLines(rawLines, width)

	wrappedToRaw := make([]int, len(wrapped))
	wi := 0
	for ri, line := range rawLines {
		w := WrapLines([]string{line}, width)
		for range w {
			if wi < len(wrappedToRaw) {
				wrappedToRaw[wi] = ri
				wi++
			}
		}
	}

	for pi := range pages {
		var pageLinks []Link
		startLine := pi * height
		for li, line := range pages[pi].Lines {
			globalWrappedIdx := startLine + li
			if globalWrappedIdx >= len(wrappedToRaw) {
				continue
			}
			rawIdx := wrappedToRaw[globalWrappedIdx]
			rawLine := rawLines[rawIdx]
			links := ExtractLinks(rawLine)
			for _, lnk := range links {
				lnk.LineOnPage = li
				pageLinks = append(pageLinks, lnk)
			}
			// Also check if the wrapped line itself has links
			if rawIdx < len(rawLines) {
				lineLinks := ExtractLinks(line)
				for _, lnk := range lineLinks {
					lnk.LineOnPage = li
					// Deduplicate
					found := false
					for _, existing := range pageLinks {
						if existing.Target == lnk.Target && existing.LineOnPage == lnk.LineOnPage {
							found = true
							break
						}
					}
					if !found {
						pageLinks = append(pageLinks, lnk)
					}
				}
			}
		}
		pages[pi].Links = pageLinks
	}
	return pages
}
