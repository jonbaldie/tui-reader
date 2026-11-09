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
	spans := codeSpanCursor{spans: InlineCodeSpans(text)}
	var links []linkMatch
	for _, m := range matches {
		if spans.overlaps(m[0], m[1]) {
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

// codeSpanCursor classifies ranges with nondecreasing starts and ends. Link
// candidates from linkRegex have that order, including rejected nested targets.
type codeSpanCursor struct {
	spans       [][2]int
	first, last int
}

func (c *codeSpanCursor) overlaps(start, end int) bool {
	count := len(c.spans)
	for c.first < count && c.spans[c.first][1] <= start {
		c.first++
	}
	for c.last < count && c.spans[c.last][0] < end {
		c.last++
	}
	if c.first >= c.last {
		return false
	}
	return c.spans[c.first][0] < start || c.spans[c.last-1][1] > end
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

func assignLabelSpans(pages []Page, placed placedLink, formattedIndex, pageIndex, linkIndex, height int) {
	visibleOnOwner := false
	for _, piece := range placed.pieces {
		fi := formattedIndex + piece.line - placed.opener.line
		pi := fi / height
		if pi >= len(pages) {
			continue
		}
		index := -1
		if pi == pageIndex {
			index = linkIndex
			visibleOnOwner = true
		}
		pages[pi].LabelSpans = append(pages[pi].LabelSpans, LabelSpan{Line: fi % height, Start: piece.start, End: piece.end, LinkIndex: index})
	}
	if !visibleOnOwner {
		piece := placed.opener
		pages[pageIndex].LabelSpans = append(pages[pageIndex].LabelSpans, LabelSpan{Line: formattedIndex % height, Start: piece.start, End: piece.end, LinkIndex: linkIndex, Opener: true})
	}
}
