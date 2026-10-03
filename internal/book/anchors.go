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
	linkRegex = regexp.MustCompile(`\[([^\[\]]+)\]\(#([^)]+)\)`)
)

// ExtractAnchors returns the heading anchors produced by Markdown layout.
func ExtractAnchors(lines []string) map[string]int {
	return formatDocument(lines, fallbackPageWidth, false).anchors
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

// isAnchorChar reports whether r is a letter or digit, in any script, suitable
// for an anchor fragment.
func isAnchorChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

type linkMatch struct {
	link  Link
	start int
}

func extractLinkMatches(text string) []linkMatch {
	matches := linkRegex.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return nil
	}
	spans := InlineCodeSpans(text)
	var links []linkMatch
	for _, m := range matches {
		if IsInlineCodeRange(spans, m[0], m[1]) {
			continue
		}
		links = append(links, linkMatch{
			link: Link{
				Label:  text[m[2]:m[3]],
				Target: text[m[4]:m[5]],
			},
			start: m[0],
		})
	}
	return links
}

// ExtractLinks finds markdown-style internal links in a line of text.
// Link markup inside an inline code span is literal text, not a link.
func ExtractLinks(line string) []Link {
	matches := extractLinkMatches(line)
	if len(matches) == 0 {
		return nil
	}
	links := make([]Link, len(matches))
	for i, m := range matches {
		links[i] = m.link
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
	n := len(rawLines)
	inFence := false
	for ri := 0; ri < n; ri++ {
		raw := rawLines[ri]
		if IsIndentedCodeLine(raw) {
			continue
		}
		if isFenceDelimiter(raw) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if !isProseLine(raw) {
			if links := ExtractLinks(raw); len(links) > 0 {
				sourceLinks[ri] = links
				sourceOrder = append(sourceOrder, ri)
			}
			continue
		}

		end := findProseBlockEnd(rawLines, ri)
		lines := rawLines[ri:end]
		collectProseBlockLinks(sourceLinks, &sourceOrder, lines, ri)
		ri = end - 1
	}
	return sourceLinkSet{links: sourceLinks, order: sourceOrder}
}

func sourceLinksFor(rawLines []string, plainText bool) sourceLinkSet {
	if plainText {
		return sourceLinkSet{links: map[int][]Link{}}
	}
	return collectSourceLinks(rawLines)
}

func collectProseBlockLinks(sourceLinks map[int][]Link, sourceOrder *[]int, lines []string, startRi int) {
	joined := strings.Join(lines, " ")
	matches := extractLinkMatches(joined)
	if len(matches) == 0 {
		return
	}
	offsets := lineOffsets(lines)
	currentLine := 0
	nLines := len(lines)
	for _, m := range matches {
		for currentLine+1 < nLines && offsets[currentLine+1] <= m.start {
			currentLine++
		}
		rawIndex := startRi + currentLine
		if _, ok := sourceLinks[rawIndex]; !ok {
			*sourceOrder = append(*sourceOrder, rawIndex)
		}
		sourceLinks[rawIndex] = append(sourceLinks[rawIndex], m.link)
	}
}

func attachLinks(pages []Page, rawLines []string, formatted []formattedLine, height int, source sourceLinkSet) []Page {
	height = normalizePageHeight(height)

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
	reflowed := newReflowIndex(rawLines, sourceLinks)
	n := len(rawLines)
	for formattedIndex, line := range formatted {
		if line.raw < 0 || line.raw >= n {
			continue
		}
		recordFormattedLine(locations, formattedIndex, line, sourceLinks, reflowed)
	}
	return locations
}

func recordFormattedLine(locations map[int]*linkLocation, fi int, line formattedLine, sourceLinks map[int][]Link, reflowed reflowIndex) {
	// Direct raw provenance associates link starts with their source line,
	// including links whose markup was broken across display lines.
	if links, ok := sourceLinks[line.raw]; ok && len(links) > 0 {
		entry := locationFor(locations, line.raw, fi)
		entry.record(line.text, fi, links)
		entry.recordStarts(line.text, fi, line.links)
	}
	recordReflowedLinks(locations, fi, line, reflowed)
}

// reflowIndex maps link starts in a prose block to their source lines in
// order, without rescanning the paragraph.
type reflowIndex struct {
	// blockStart holds the first raw index of each line's prose block, or -1
	// for lines that are not prose.
	blockStart []int
	// lines lists, per prose block and link markup, the raw indices whose
	// links include that markup, in ascending order. Display lines consume
	// them from the front as wrapped link starts appear.
	lines map[reflowKey][]int
}

type reflowKey struct {
	block  int
	markup string
}

// newReflowIndex leaves non-prose lines out of the index, so their display
// lines, whose block is -1, match nothing.
func newReflowIndex(rawLines []string, sourceLinks map[int][]Link) reflowIndex {
	blockStart := proseBlockStarts(rawLines)
	lines := make(map[reflowKey][]int)
	for rawIndex := range rawLines {
		links := sourceLinks[rawIndex]
		if blockStart[rawIndex] < 0 {
			continue
		}
		for _, link := range links {
			key := reflowKey{block: blockStart[rawIndex], markup: linkMarkup(link)}
			lines[key] = append(lines[key], rawIndex)
		}
	}
	return reflowIndex{blockStart: blockStart, lines: lines}
}

// proseBlockStarts returns, per raw line, the first raw index of its prose
// block, or -1 for lines that are not prose.
func proseBlockStarts(rawLines []string) []int {
	blockStart := make([]int, len(rawLines))
	for i, raw := range rawLines {
		switch {
		case !isProseLine(raw):
			blockStart[i] = -1
		case i > 0 && blockStart[i-1] >= 0:
			blockStart[i] = blockStart[i-1]
		default:
			blockStart[i] = i
		}
	}
	return blockStart
}

// recordReflowedLinks attributes each wrapped link start to the next source
// occurrence of the same markup in the prose block. Wrapping preserves order,
// so a display line cannot claim another source line's occurrence merely
// because its text contains identical markup.
func recordReflowedLinks(locations map[int]*linkLocation, fi int, line formattedLine, reflowed reflowIndex) {
	block := reflowed.blockStart[line.raw]
	for _, link := range line.links {
		key := reflowKey{block: block, markup: linkMarkup(link)}
		pending := reflowed.lines[key]
		if len(pending) == 0 {
			continue
		}

		rawIndex := pending[0]
		reflowed.lines[key] = pending[1:]
		if rawIndex <= line.raw {
			continue
		}

		entry := locationFor(locations, rawIndex, fi)
		if entry.links == nil {
			entry.links = make(map[Link][]int)
		}
		entry.links[link] = append(entry.links[link], fi)
	}
}

func locationFor(locations map[int]*linkLocation, rawIndex, fi int) *linkLocation {
	entry := locations[rawIndex]
	if entry == nil {
		entry = &linkLocation{first: fi}
		locations[rawIndex] = entry
	}
	return entry
}

func linkMarkup(link Link) string {
	return "[" + link.Label + "](#" + link.Target + ")"
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
