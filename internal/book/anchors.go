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

func AttachLinks(pages []Page, rawLines []string, width, height int) []Page {
	formatted := formatParagraphsWithProvenance(rawLines, width)
	return attachLinks(pages, rawLines, formatted, height)
}

func attachLinks(pages []Page, rawLines []string, formatted []formattedLine, height int) []Page {
	if height < 1 {
		height = 20
	}

	type location struct {
		first int
		links map[Link][]int
	}

	sourceLinks := make(map[int][]Link)
	sourceOrder := make([]int, 0)
	for rawIndex, rawLine := range rawLines {
		if links := ExtractLinks(rawLine); len(links) > 0 {
			sourceLinks[rawIndex] = links
			sourceOrder = append(sourceOrder, rawIndex)
		}
	}

	locations := make(map[int]*location, len(sourceLinks))
	for formattedIndex, line := range formatted {
		if line.raw < 0 || line.raw >= len(rawLines) {
			continue
		}
		if _, ok := sourceLinks[line.raw]; !ok {
			continue
		}
		entry := locations[line.raw]
		if entry == nil {
			entry = &location{first: formattedIndex}
			locations[line.raw] = entry
		}
		for _, link := range ExtractLinks(line.text) {
			if entry.links == nil {
				entry.links = make(map[Link][]int)
			}
			entry.links[link] = append(entry.links[link], formattedIndex)
		}
	}

	for pageIndex := range pages {
		pages[pageIndex].Links = nil
	}
	for _, rawIndex := range sourceOrder {
		links := sourceLinks[rawIndex]
		entry := locations[rawIndex]
		if entry == nil {
			continue
		}
		for _, link := range links {
			formattedIndex := entry.first
			if candidates := entry.links[link]; len(candidates) > 0 {
				formattedIndex = candidates[0]
				entry.links[link] = candidates[1:]
			}
			pageIndex := formattedIndex / height
			if formattedIndex < 0 || pageIndex >= len(pages) {
				continue
			}
			link.LineOnPage = formattedIndex % height
			pages[pageIndex].Links = append(pages[pageIndex].Links, link)
		}
	}
	return pages
}
