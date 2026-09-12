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
		if IsIndentedCodeLine(line) {
			continue
		}
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
// Link markup inside an inline code span is literal text, not a link.
func ExtractLinks(line string) []Link {
	matches := linkRegex.FindAllStringSubmatchIndex(line, -1)
	if matches == nil {
		return nil
	}
	spans := InlineCodeSpans(line)
	var links []Link
	for _, m := range matches {
		if IsInlineCodeRange(spans, m[0], m[1]) {
			continue
		}
		links = append(links, Link{
			Label:  line[m[2]:m[3]],
			Target: line[m[4]:m[5]],
		})
	}
	return links
}

// InlineCodeSpans returns the [start, end) byte ranges of inline code spans.
func InlineCodeSpans(line string) [][2]int {
	return codeSpans(line)
}

// IsInlineCodeRange reports whether the range [start, end) overlaps an inline
// code span outside the range. It lets callers distinguish literal Markdown
// from active markup using the same rules as ExtractLinks.
func IsInlineCodeRange(spans [][2]int, start, end int) bool {
	return overlapsCodeSpan(start, end, spans)
}

// codeSpans returns the [start, end) byte ranges of inline code spans: a run
// of backticks closed by the next run of the same length.
func codeSpans(line string) [][2]int {
	var spans [][2]int
	n := len(line)
	for i := 0; i < n; {
		if line[i] != '`' {
			i++
			continue
		}
		start := i
		for i < n && line[i] == '`' {
			i++
		}
		run := i - start
		if end := closingBacktickRun(line, i, run); end >= 0 {
			spans = append(spans, [2]int{start, end})
			i = end
		}
	}
	return spans
}

// closingBacktickRun returns the end of the first backtick run of exactly
// length run at or after from, or -1 if there is none.
func closingBacktickRun(line string, from, run int) int {
	n := len(line)
	for j := from; j < n; {
		if line[j] != '`' {
			j++
			continue
		}
		k := j
		for k < n && line[k] == '`' {
			k++
		}
		if k-j == run {
			return k
		}
		j = k
	}
	return -1
}

// overlapsCodeSpan reports whether [start, end) intersects a code span it does
// not wholly contain; a link label may contain code, but code cannot contain
// or split a link.
func overlapsCodeSpan(start, end int, spans [][2]int) bool {
	for _, s := range spans {
		if s[0] < end && start < s[1] && (s[0] < start || s[1] > end) {
			return true
		}
	}
	return false
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
		if IsIndentedCodeLine(rawLine) {
			continue
		}
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
	n := len(rawLines)
	for formattedIndex, line := range formatted {
		if line.raw < 0 || line.raw >= n {
			continue
		}
		recordFormattedLine(locations, formattedIndex, line, sourceLinks, rawLines)
	}
	return locations
}

func recordFormattedLine(locations map[int]*linkLocation, fi int, line formattedLine, sourceLinks map[int][]Link, rawLines []string) {
	// Direct raw provenance associates link starts with their source line,
	// including links whose markup was broken across display lines.
	if links, ok := sourceLinks[line.raw]; ok && len(links) > 0 {
		entry := locations[line.raw]
		if entry == nil {
			entry = &linkLocation{first: fi}
			locations[line.raw] = entry
		}
		entry.record(line.text, fi, links)
		entry.recordStarts(line.text, fi, line.links)
	}
	recordReflowedLinks(locations, fi, line.text, line.raw, sourceLinks, rawLines)
}

func recordReflowedLinks(locations map[int]*linkLocation, fi int, text string, lineRaw int, sourceLinks map[int][]Link, rawLines []string) {
	if lineRaw < 0 || lineRaw >= len(rawLines) || !isProseLine(rawLines[lineRaw]) {
		return
	}
	start := lineRaw
	for start > 0 && isProseLine(rawLines[start-1]) {
		start--
	}
	end := findProseBlockEnd(rawLines, start)
	for rawIndex := start; rawIndex < end; rawIndex++ {
		if rawIndex == lineRaw {
			continue
		}
		if links := sourceLinks[rawIndex]; len(links) > 0 {
			recordMatchingLinks(locations, fi, text, rawIndex, links)
		}
	}
}

func recordMatchingLinks(locations map[int]*linkLocation, fi int, text string, rawIndex int, links []Link) {
	for _, link := range links {
		linkMarkup := "[" + link.Label + "](#" + link.Target + ")"
		if strings.Contains(text, linkMarkup) {
			entry := locations[rawIndex]
			if entry == nil {
				entry = &linkLocation{first: fi}
				locations[rawIndex] = entry
			}
			entry.record(text, fi, links)
			break
		}
	}
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

func (l *linkLocation) recordStarts(line string, formattedIndex int, starts []Link) {
	if len(starts) == 0 {
		return
	}
	counts := make(map[Link]int, len(starts))
	for _, link := range starts {
		counts[link]++
	}

	for link, count := range counts {
		linkMarkup := "[" + link.Label + "](#" + link.Target + ")"
		missing := count - strings.Count(line, linkMarkup)
		if missing <= 0 {
			continue
		}
		if l.links == nil {
			l.links = make(map[Link][]int)
		}
		for k := 0; k < missing; k++ {
			l.links[link] = append(l.links[link], formattedIndex)
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
