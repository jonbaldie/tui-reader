package book

import (
	"regexp"
	"strings"
	"unicode"
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
	var b strings.Builder
	pendingHyphen := false
	for _, r := range text {
		r = unicode.ToLower(r)
		if isAnchorChar(r) {
			if b.Len() > 0 && pendingHyphen {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			pendingHyphen = false
			continue
		}
		if r == ' ' || r == '-' {
			pendingHyphen = true
		}
	}
	return b.String()
}

// isAnchorChar reports whether r is a lowercase letter or digit suitable for
// an anchor fragment.
func isAnchorChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
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
	return attachLinks(pages, rawLines, formatted, height, sourceLinkSet{})
}

// linkLocation tracks where a source line's links appear in the formatted
// output: the first formatted line index, and per-link candidate indices.
type linkLocation struct {
	first int
	links map[Link][]int
}

// sourceLinkSet holds pre-extracted links from raw document lines and the order
// in which lines with links were discovered.
type sourceLinkSet struct {
	links map[int][]Link
	order []int
}

// collectSourceLinks scans raw lines for links, returning a sourceLinkSet.
func collectSourceLinks(rawLines []string) sourceLinkSet {
	sourceLinks := make(map[int][]Link)
	sourceOrder := make([]int, 0)
	for rawIndex, rawLine := range rawLines {
		if links := ExtractLinks(rawLine); len(links) > 0 {
			sourceLinks[rawIndex] = links
			sourceOrder = append(sourceOrder, rawIndex)
		}
	}
	return sourceLinkSet{links: sourceLinks, order: sourceOrder}
}

func attachLinks(pages []Page, rawLines []string, formatted []formattedLine, height int, source sourceLinkSet) []Page {
	if height < 1 {
		height = 20
	}

	if source.links == nil {
		source = collectSourceLinks(rawLines)
	}

	locations := buildLocations(formatted, rawLines, source.links)
	assignLinksToPages(pages, source.order, source.links, locations, height)
	return pages
}

// buildLocations maps each source-line index that has links to a linkLocation,
// recording the first formatted-line index and all formatted indices per link.
func buildLocations(formatted []formattedLine, rawLines []string, sourceLinks map[int][]Link) map[int]*linkLocation {
	locations := make(map[int]*linkLocation, len(sourceLinks))
	for formattedIndex, line := range formatted {
		if line.raw < 0 || line.raw >= len(rawLines) {
			continue
		}
		links, ok := sourceLinks[line.raw]
		if !ok || len(links) == 0 {
			continue
		}
		entry := locations[line.raw]
		if entry == nil {
			entry = &linkLocation{first: formattedIndex}
			locations[line.raw] = entry
		}
		entry.record(line.text, formattedIndex, links)
	}
	return locations
}

func (l *linkLocation) record(line string, formattedIndex int, links []Link) {
	seen := make(map[Link]struct{}, len(links))
	for _, link := range links {
		if _, ok := seen[link]; ok {
			continue
		}
		seen[link] = struct{}{}

		linkMarkup := "[" + link.Label + "](#" + link.Target + ")"
		count := strings.Count(line, linkMarkup)
		if count > 0 {
			if l.links == nil {
				l.links = make(map[Link][]int)
			}
			for k := 0; k < count; k++ {
				l.links[link] = append(l.links[link], formattedIndex)
			}
		}
	}
}

// assignLinksToPages clears existing page links and places each source link on
// the page containing its first (or next candidate) formatted line.
func assignLinksToPages(pages []Page, sourceOrder []int, sourceLinks map[int][]Link, locations map[int]*linkLocation, height int) {
	for pageIndex := range pages {
		pages[pageIndex].Links = nil
	}
	numPages := len(pages)
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
			if formattedIndex < 0 || pageIndex >= numPages {
				continue
			}
			link.LineOnPage = formattedIndex % height
			pages[pageIndex].Links = append(pages[pageIndex].Links, link)
		}
	}
}
