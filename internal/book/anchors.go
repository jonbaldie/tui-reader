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
// BuildFormattedToRawMap builds a mapping from formatted line indices to raw
// line indices. Blank lines inserted by formatting map to -1.
func BuildFormattedToRawMap(rawLines []string, width int) []int {
	formatted := FormatParagraphs(rawLines, width)
	fmtToRaw := make([]int, len(formatted))
	for i := range fmtToRaw {
		fmtToRaw[i] = -1
	}

	// Walk rawLines and formatted in lockstep.
	// FormatParagraphs processes rawLines sequentially, so the formatted
	// output for raw line ri appears as a contiguous block in the output.
	fi := 0
	firstParagraph := true
	for ri, raw := range rawLines {
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" {
			// Blank raw line may have produced a spacer
			if fi < len(formatted) && formatted[fi] == "" {
				fmtToRaw[fi] = ri
				fi++
			}
			continue
		}

		// Non-first paragraphs have a blank spacer before content
		if !firstParagraph && fi < len(formatted) && formatted[fi] == "" {
			fmtToRaw[fi] = -1
			fi++
		}

		// Determine how many formatted lines this raw line produced
		isSpecial := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "    ")
		wrapWidth := width
		if !firstParagraph && !isSpecial {
			wrapWidth = width - 2
			if wrapWidth < 10 {
				wrapWidth = 10
			}
		}
		w := WrapLines([]string{raw}, wrapWidth)

		for range w {
			if fi < len(fmtToRaw) {
				fmtToRaw[fi] = ri
				fi++
			}
		}

		firstParagraph = false
	}
	return fmtToRaw
}

func AttachLinks(pages []Page, rawLines []string, width, height int) []Page {
	fmtToRaw := BuildFormattedToRawMap(rawLines, width)

	for pi := range pages {
		var pageLinks []Link
		startLine := pi * height
		for li, line := range pages[pi].Lines {
			globalIdx := startLine + li
			if globalIdx >= len(fmtToRaw) {
				continue
			}
			rawIdx := fmtToRaw[globalIdx]
			if rawIdx < 0 || rawIdx >= len(rawLines) {
				continue // spacer line
			}
			rawLine := rawLines[rawIdx]
			links := ExtractLinks(rawLine)
			for _, lnk := range links {
				lnk.LineOnPage = li
				pageLinks = append(pageLinks, lnk)
			}
			// Also check the formatted line itself for links
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
		pages[pi].Links = pageLinks
	}
	return pages
}
