// Package book handles loading, parsing, and paginating readable files.
package book

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

const (
	// DefaultPageWidth is the initial number of columns used to lay out a book.
	DefaultPageWidth = 60
	// DefaultPageHeight is the initial number of lines used to lay out a book.
	DefaultPageHeight = 20
	fallbackPageWidth = 80
)

// Anchor represents a named location in the document that can be linked to.
type Anchor struct {
	Name string // normalized anchor name (e.g. "chapter-1")
	Line int    // 0-based line index in the original content
}

// Link represents a clickable reference within the rendered page.
type Link struct {
	Label      string // display text
	Target     string // anchor name this link points to
	LineOnPage int    // 0-based line index within the current page
}

// Page represents a single screen of content.
type Page struct {
	Lines      []string    // lines of text to display
	Links      []Link      // clickable links on this page
	Headings   []bool      // parallel to Lines; true for every display line of a heading block
	LabelSpans []LabelSpan // link label ranges and selected-only openers
}

// LabelSpan identifies a half-open UTF-8 byte range on a page line. LinkIndex
// indexes Links, or is -1 when the link starts on another page. Opener spans
// are rendered only when selected and no label text fits on the owning page.
type LabelSpan struct {
	Line, Start, End, LinkIndex int
	Opener                      bool
}

// Book is a loaded and paginated document. Its source and layout are
// private so that only Reflow can change them, keeping the derived caches
// consistent (#176).
type Book struct {
	Title string

	rawLines      []string
	pages         []Page
	anchors       map[string]int // anchor name -> line index in rawLines
	pageWidth     int
	pageHeight    int
	sourceLinks   sourceLinkSet
	rawLinePages  map[int]int
	pagePositions []Position
	lineSources   []Position
	plainText     bool
}

// Position is a place in the source text: a raw line index and the number of
// non-space runes of that line's display text before it. Unlike a page index,
// it survives a reflow to a new width, even inside a paragraph that spans
// pages (#168).
type Position struct {
	Raw    int
	Offset int
}

func readLines(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read document: %w", err)
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("file is not valid UTF-8")
	}

	// A leading UTF-8 byte-order mark is not content: left in place it would
	// make the first line start with an invisible rune, hiding a heading.
	return splitLines(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))), nil
}

// splitLines splits data into lines, normalizing CRLF (\r\n), CR (\r),
// and LF (\n) in a single pass without intermediate full-string copies.
func splitLines(data []byte) []string {
	var lines []string
	start := 0
	n := len(data)
	for i := 0; i < n; i++ {
		if data[i] == '\r' {
			lines = append(lines, string(data[start:i]))
			if i+1 < n && data[i+1] == '\n' {
				i++
			}
			start = i + 1
		} else if data[i] == '\n' {
			lines = append(lines, string(data[start:i]))
			start = i + 1
		}
	}
	lines = append(lines, string(data[start:n]))
	return lines
}

// deriveTitle extracts a human-readable title from a file path.
func deriveTitle(path string) string {
	// Get basename
	parts := strings.Split(path, "/")
	base := parts[len(parts)-1]
	// Also handle backslash paths
	parts = strings.Split(base, "\\")
	base = parts[len(parts)-1]

	// Remove extension
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}

	// Replace separators with spaces
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")

	return titleCase(base)
}

// titleCase capitalizes the first letter of each word.
func titleCase(s string) string {
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r
		return r
	}, s)
}

// formattedLine is one display line paired with its raw-line provenance: the
// 0-based index of the source line it came from, or -1 for a blank line that
// formatting inserted as a paragraph spacer. Across a document the raw indices
// are non-decreasing. The links field records link tokens whose markup starts
// on the display line; heading marks all content lines of a heading block.
type formattedLine struct {
	text         string
	raw          int
	endRaw       int
	sourceOffset int
	hasOffset    bool
	placed       []placedLink
	heading      bool
}

type formattedDocument struct {
	lines   []formattedLine
	anchors map[string]int
}

type wrappedLine struct {
	text   string
	placed []placedLink
}

type labelPiece struct {
	line, start, end int
}

type placedLink struct {
	link   Link
	pieces []labelPiece
	opener labelPiece
}

type bookLayout struct {
	formatted     []formattedLine
	pages         []Page
	anchors       map[string]int
	rawLinePages  map[int]int
	pagePositions []Position
	lineSources   []Position
	height        int
}

// IsIndentedCodeLine reports whether raw is an indented Markdown code line.
func IsIndentedCodeLine(raw string) bool {
	return codeIndentLength(raw) > 0
}

// codeIndentLength returns the byte length of the leading whitespace that
// indents raw to column 4, or 0 if raw is not indented code. A tab advances
// to the next multiple of 4 columns.
func codeIndentLength(raw string) int {
	column := 0
	for i, c := range []byte(raw) {
		switch c {
		case ' ':
			column++
		case '\t':
			column += 4 - column%4
		default:
			return 0
		}
		if column >= 4 {
			return i + 1
		}
	}
	return 0
}

// fenceDelimiter returns the delimiter run of a Markdown fenced-code opener,
// or "" if raw is not an opener. An opener has at least three backticks or
// tildes; a backtick opener's info string may not contain backticks.
func fenceDelimiter(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || (trimmed[0] != '`' && trimmed[0] != '~') {
		return ""
	}
	run := trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1]))]
	if len(run) < 3 || (run[0] == '`' && strings.Contains(trimmed[len(run):], "`")) {
		return ""
	}
	return run
}

// isFenceDelimiter reports whether raw is a Markdown fenced-code opener.
func isFenceDelimiter(raw string) bool {
	return fenceDelimiter(raw) != ""
}

// isClosingFence reports whether raw closes a fence opened with fence. A
// closing fence contains only the opener's character and is at least as long
// as its opener.
func isClosingFence(raw string, fence string) bool {
	trimmed := strings.TrimSpace(raw)
	return len(trimmed) >= len(fence) && strings.Trim(trimmed, fence[:1]) == ""
}

func findFenceBlockEnd(rawLines []string, start int) int {
	n := len(rawLines)
	fence := fenceDelimiter(rawLines[start])
	for end := start + 1; end < n; end++ {
		if IsIndentedCodeLine(rawLines[end]) {
			continue
		}
		if isClosingFence(rawLines[end], fence) {
			return end + 1
		}
	}
	return n
}

// formatParagraphsWithProvenance exposes Markdown display lines and their raw
// source-line provenance to in-package formatting tests. Book construction uses
// formatDocument, which also captures anchors and supports plain text.
// Both paths share formatMarkdownDocument rather than re-deriving formatting
// to recover provenance.
func formatParagraphsWithProvenance(rawLines []string, width int) []formattedLine {
	return formatMarkdownDocument(rawLines, width, false).lines
}

func formatMarkdownDocument(rawLines []string, width int, captureAnchors bool) formattedDocument {
	width = normalizePageWidth(width)

	var result []formattedLine
	var anchors map[string]int
	if captureAnchors {
		anchors = make(map[string]int)
	}
	firstParagraph := true
	prevCode := false
	prevList := false
	n := len(rawLines)

	for ri := 0; ri < n; ri++ {
		raw := rawLines[ri]
		if strings.TrimSpace(raw) == "" {
			if needsBlankSeparator(result) {
				result = append(result, formattedLine{text: "", raw: ri})
			}
			prevCode, prevList = false, false
			continue
		}

		block := formatNextBlock(rawLines, ri, firstParagraph, prevCode, prevList, width, result)
		if title := headingTitle(raw, block); title != "" {
			for i := range block.lines {
				block.lines[i].heading = block.lines[i].text != ""
			}
			if anchors != nil {
				anchors[NormalizeAnchor(title)] = ri
			}
		}
		result = append(result, block.lines...)
		firstParagraph = false
		prevCode, prevList = block.isCode, block.isList
		ri = block.lastRi
	}

	return formattedDocument{lines: result, anchors: anchors}
}

// headingTitle classifies a Markdown block for both anchors and display lines.
func headingTitle(raw string, block formattedBlock) string {
	if block.setextTitle != "" {
		return block.setextTitle
	}
	if block.isCode {
		return ""
	}
	return atxHeadingTitle(raw)
}

func atxHeadingTitle(raw string) string {
	if IsIndentedCodeLine(raw) {
		return ""
	}
	if m := headingRegex.FindStringSubmatch(strings.TrimSpace(raw)); m != nil {
		return m[2]
	}
	return ""
}

type formattedBlock struct {
	lines  []formattedLine
	isCode bool
	isList bool
	lastRi int
	// setextTitle is the heading text of a setext heading block, else "".
	setextTitle string
}

func formatNextBlock(rawLines []string, ri int, firstParagraph, prevCode, prevList bool, width int, result []formattedLine) formattedBlock {
	raw := rawLines[ri]
	if end := codeBlockEnd(rawLines, ri); end > ri {
		return formatCodeRange(rawLines, ri, end, firstParagraph, prevCode, width, result)
	}

	if lineBlockMarker(raw) != 0 {
		return formatLineBlock(rawLines, ri, width, firstParagraph, result)
	}

	if isListItem(raw) {
		var lines []formattedLine
		if needsBlockSeparator(firstParagraph, prevList, result) {
			lines = append(lines, formattedLine{text: "", raw: -1})
		}
		lines = append(lines, formatParagraph(raw, ri, true, width)...)
		return formattedBlock{lines: lines, isList: true, lastRi: ri}
	}

	if isSpecialLine(raw) {
		var lines []formattedLine
		if needsBlockSeparator(firstParagraph, false, result) {
			lines = append(lines, formattedLine{text: "", raw: -1})
		}
		lines = append(lines, formatParagraph(raw, ri, true, width)...)
		return formattedBlock{lines: lines, lastRi: ri}
	}

	end := findProseBlockEnd(rawLines, ri)
	var lines []formattedLine
	if needsBlockSeparator(firstParagraph, false, result) {
		lines = append(lines, formattedLine{text: "", raw: -1})
	}
	if level := setextUnderlineAt(rawLines, end); level > 0 {
		title := setextTitle(rawLines[ri:end])
		heading := strings.Repeat("#", level) + " " + title
		lines = append(lines, formatParagraph(heading, ri, true, width)...)
		return formattedBlock{lines: lines, lastRi: end, setextTitle: title}
	}
	lines = append(lines, formatMultiLineParagraph(rawLines[ri:end], ri, firstParagraph, width)...)
	return formattedBlock{lines: lines, lastRi: end - 1}
}

func lineBlockMarker(raw string) byte {
	if IsIndentedCodeLine(raw) {
		return 0
	}
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") {
		return trimmed[0]
	}
	return 0
}

func formatLineBlock(rawLines []string, start, width int, firstParagraph bool, result []formattedLine) formattedBlock {
	block := formattedBlock{lastRi: start - 1}
	if needsBlockSeparator(firstParagraph, false, result) {
		block.lines = append(block.lines, formattedLine{text: "", raw: -1})
	}
	marker := lineBlockMarker(rawLines[start])
	n := len(rawLines)
	for ri := start; ri < n && lineBlockMarker(rawLines[ri]) == marker; ri++ {
		raw := strings.TrimSpace(rawLines[ri])
		var wrapped []wrappedLine
		if marker == '|' {
			wrapped = wrapTableRow(raw, width)
		} else {
			wrapped = wrapLineWithLinks(raw, width)
		}
		block.lines = append(block.lines, formatWrappedLines(wrapped, ri, "")...)
		block.lastRi = ri
	}
	return block
}

func wrapTableRow(raw string, width int) []wrappedLine {
	var texts []string
	var current []rune
	currentWidth := 0
	appendWordRunes(&current, &currentWidth, raw, width, &texts)
	if len(current) > 0 {
		texts = append(texts, string(current))
	}
	matches := extractLinkMatches(raw)
	wrapped := make([]wrappedLine, len(texts))
	offset, next := 0, 0
	nMatches := len(matches)
	for i, text := range texts {
		wrapped[i].text = text
		offset += len(text)
		for next < nMatches && matches[next].start < offset {
			match := matches[next]
			placed := placedLink{link: match.link, opener: labelPiece{line: i, start: match.start - (offset - len(text)), end: match.start - (offset - len(text)) + 1}}
			labelStart, labelEnd := match.start+1, match.start+1+len(match.link.Label)
			pieceOffset := offset - len(text)
			nTexts := len(texts)
			for j := i; j < nTexts && pieceOffset < labelEnd; j++ {
				row := texts[j]
				start := max(labelStart, pieceOffset)
				end := min(labelEnd, pieceOffset+len(row))
				if start < end {
					placed.pieces = append(placed.pieces, labelPiece{line: j, start: start - pieceOffset, end: end - pieceOffset})
				}
				pieceOffset += len(row)
			}
			wrapped[i].placed = append(wrapped[i].placed, placed)
			next++
		}
	}
	return wrapped
}

// setextLevel returns 1 for a "=" setext heading underline, 2 for a "-" one,
// or 0 if raw is neither. An underline is one run of the character, indented
// at most three spaces, with optional trailing spaces.
func setextLevel(raw string) int {
	if IsIndentedCodeLine(raw) {
		return 0
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.Trim(trimmed, trimmed[:1]) != "" {
		return 0
	}
	switch trimmed[0] {
	case '=':
		return 1
	case '-':
		return 2
	}
	return 0
}

// setextUnderlineAt returns the setext heading level of rawLines[i], which
// must directly follow a prose block, or 0 if there is no underline there.
func setextUnderlineAt(rawLines []string, i int) int {
	if i >= len(rawLines) {
		return 0
	}
	return setextLevel(rawLines[i])
}

// setextTitle joins a setext heading's paragraph lines into its title.
func setextTitle(lines []string) string {
	parts := make([]string, len(lines))
	for i, line := range lines {
		parts[i] = strings.TrimSpace(line)
	}
	return strings.Join(parts, " ")
}

func codeBlockEnd(rawLines []string, ri int) int {
	if IsIndentedCodeLine(rawLines[ri]) {
		return ri + 1
	}
	if isFenceDelimiter(rawLines[ri]) {
		return findFenceBlockEnd(rawLines, ri)
	}
	return ri
}

func formatCodeRange(rawLines []string, start, end int, firstParagraph, prevCode bool, width int, result []formattedLine) formattedBlock {
	var lines []formattedLine
	if needsBlockSeparator(firstParagraph, prevCode, result) {
		lines = append(lines, formattedLine{text: "", raw: -1})
	}
	for i := start; i < end; i++ {
		if strings.TrimSpace(rawLines[i]) == "" {
			lines = append(lines, formattedLine{text: "", raw: i})
			continue
		}
		lines = append(lines, formatCodeBlock(rawLines[i], i, width)...)
	}
	return formattedBlock{lines: lines, isCode: true, lastRi: end - 1}
}

func findProseBlockEnd(rawLines []string, start int) int {
	n := len(rawLines)
	end := start + 1
	for end < n && isProseLine(rawLines[end]) && setextLevel(rawLines[end]) == 0 {
		end++
	}
	return end
}

// needsBlankSeparator reports whether a blank separator line should be inserted
// before the next content line: only when there is preceding content whose last
// line is not already blank.
func needsBlankSeparator(result []formattedLine) bool {
	return len(result) > 0 && result[len(result)-1].text != ""
}

// needsBlockSeparator reports whether a blank separator line should be inserted
// between two blocks: only when this is not the first block, does not continue
// the previous block, and preceding content does not already end with a blank line.
func needsBlockSeparator(firstParagraph, continuesBlock bool, result []formattedLine) bool {
	return !firstParagraph && !continuesBlock && needsBlankSeparator(result)
}

func isOrderedListItem(trimmed string) bool {
	tLen := len(trimmed)
	i := 0
	for i < tLen && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	return i > 0 && i < tLen && (strings.HasPrefix(trimmed[i:], ". ") || strings.HasPrefix(trimmed[i:], ") "))
}

// isListItem reports whether raw is a markdown list item (ordered or unordered).
func isListItem(raw string) bool {
	if IsIndentedCodeLine(raw) {
		return false
	}
	trimmed := strings.TrimLeft(raw, " ")
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return true
	}
	return isOrderedListItem(trimmed)
}

// isSpecialLine reports whether a line is a heading, horizontal rule, or list
// item — lines that should not receive paragraph indentation.
func isSpecialLine(raw string) bool {
	if isListItem(raw) {
		return true
	}
	trimmed := strings.TrimSpace(raw)
	return atxHeadingTitle(raw) != "" || strings.HasPrefix(trimmed, "---")
}

func isProseLine(raw string) bool {
	return strings.TrimSpace(raw) != "" && !IsIndentedCodeLine(raw) && !isFenceDelimiter(raw) && !isSpecialLine(raw) && lineBlockMarker(raw) == 0
}

func formatMultiLineParagraph(lines []string, startRi int, firstParagraph bool, width int) []formattedLine {
	if len(lines) == 1 {
		return formatParagraph(lines[0], startRi, firstParagraph, width)
	}

	joined := strings.Join(lines, " ")
	wrapped, shouldIndent := wrapParagraphWithIndent(joined, width, !firstParagraph)
	offsets := lineOffsets(lines)
	return mapWrappedProvenance(wrapped, joined, offsets, startRi, shouldIndent)
}

const paragraphIndent = "  "

// wrapParagraphWithIndent reserves the paragraph indent while wrapping, but
// drops it and rewraps at the full width if the first content line contains a
// wide rune that cannot fit alongside the indent.
func wrapParagraphWithIndent(text string, width int, indent bool) ([]wrappedLine, bool) {
	if !indent {
		return wrapLineWithLinks(text, width), false
	}

	indentWidth := stringWidth(paragraphIndent)
	if width <= indentWidth {
		return wrapLineWithLinks(text, width), false
	}

	wrapped := wrapLineWithLinks(text, width-indentWidth)
	if len(wrapped) > 0 && stringWidth(wrapped[0].text)+indentWidth <= width {
		return wrapped, true
	}

	return wrapLineWithLinks(text, width), false
}

func lineOffsets(lines []string) []int {
	n := len(lines)
	offsets := make([]int, n)
	pos := 0
	for i := 0; i < n; i++ {
		offsets[i] = pos
		pos += len(lines[i]) + 1
	}
	return offsets
}

type provenanceCursor struct {
	source  string
	offsets []int
	pos     int
	raw     int
	offset  int
}

func (c *provenanceCursor) skipSpace() {
	if c.pos >= len(c.source) {
		return
	}
	r, size := utf8.DecodeRuneInString(c.source[c.pos:])
	if unicode.IsSpace(r) {
		c.pos += size
	}
}

func (c *provenanceCursor) nextContent() Position {
	n := len(c.source)
	for c.pos < n {
		r, size := utf8.DecodeRuneInString(c.source[c.pos:])
		if unicode.IsSpace(r) {
			c.pos += size
			continue
		}
		nextLine := advanceLine(c.offsets, c.raw, c.pos)
		if nextLine != c.raw {
			c.raw, c.offset = nextLine, 0
		}
		position := Position{Raw: c.raw, Offset: c.offset}
		c.pos += size
		c.offset++
		return position
	}
	return Position{Raw: c.raw, Offset: c.offset}
}

func (c *provenanceCursor) line(text string) (Position, int) {
	start := Position{Raw: -1}
	end := c.raw
	for _, r := range text {
		if unicode.IsSpace(r) {
			c.skipSpace()
			continue
		}
		position := c.nextContent()
		if start.Raw < 0 {
			start = position
		}
		end = position.Raw
	}
	if start.Raw < 0 {
		start = Position{Raw: c.raw, Offset: c.offset}
	}
	return start, end
}

func mapWrappedProvenance(wrapped []wrappedLine, joined string, offsets []int, startRi int, shouldIndent bool) []formattedLine {
	result := make([]formattedLine, len(wrapped))
	cursor := provenanceCursor{source: joined, offsets: offsets}
	for i, line := range wrapped {
		position, end := cursor.line(line.text)
		text := line.text
		if i == 0 && shouldIndent {
			text = paragraphIndent + text
			shiftPlaced(line.placed, len(paragraphIndent))
		}
		result[i] = formattedLine{text: text, raw: startRi + position.Raw, endRaw: startRi + end, sourceOffset: position.Offset, hasOffset: true, placed: line.placed}
	}
	return result
}

// advanceLine returns the last line at or after line whose offset is at most
// pos.
func advanceLine(offsets []int, line, pos int) int {
	n := len(offsets)
	for line+1 < n && offsets[line+1] <= pos {
		line++
	}
	return line
}

// formatParagraph wraps a single non-blank raw line into display lines with
// optional 2-space indentation, preserving the source-line index as provenance.
func formatParagraph(raw string, ri int, firstParagraph bool, width int) []formattedLine {
	if IsIndentedCodeLine(raw) {
		return formatCodeBlock(raw, ri, width)
	}

	isSpecial := isSpecialLine(raw)
	wrapped, shouldIndent := wrapParagraphWithIndent(raw, width, !firstParagraph && !isSpecial)
	if shouldIndent && len(wrapped) > 0 {
		wrapped[0].text = paragraphIndent + wrapped[0].text
		shiftPlaced(wrapped[0].placed, len(paragraphIndent))
	}

	return formatWrappedLines(wrapped, ri, "")
}

const codeBlockIndent = "    "

func formatCodeBlock(raw string, ri int, width int) []formattedLine {
	code := raw[codeIndentLength(raw):]
	raw = codeBlockIndent + code
	indentWidth := stringWidth(codeBlockIndent)
	if width <= indentWidth {
		return wrapFormattedLines(WrapLines([]string{raw}, width), ri, "")
	}

	wrapped := WrapLines([]string{code}, width-indentWidth)
	if fitsWithIndent(wrapped, indentWidth, width) {
		return wrapFormattedLines(wrapped, ri, codeBlockIndent)
	}

	return wrapFormattedLines(WrapLines([]string{raw}, width), ri, "")
}

// fitsWithIndent reports whether every wrapped line stays within width once
// the indent is prepended. Wide runes that can't be split may leave a
// wrapped line wider than the space reserved for it.
func fitsWithIndent(wrapped []string, indentWidth, width int) bool {
	for _, w := range wrapped {
		if stringWidth(w)+indentWidth > width {
			return false
		}
	}
	return true
}

func wrapFormattedLines(wrapped []string, ri int, prefix string) []formattedLine {
	lines := make([]formattedLine, len(wrapped))
	for i, w := range wrapped {
		lines[i] = formattedLine{text: prefix + w, raw: ri}
	}
	return lines
}

func formatWrappedLines(wrapped []wrappedLine, ri int, prefix string) []formattedLine {
	lines := make([]formattedLine, len(wrapped))
	for i, w := range wrapped {
		lines[i] = formattedLine{text: prefix + w.text, raw: ri, placed: w.placed}
	}
	return lines
}

// formatPlainTextWithProvenance wraps each source line independently. Consecutive
// lines are not joined, and paragraph indentation is not applied.
func formatPlainTextWithProvenance(rawLines []string, width int) []formattedLine {
	width = normalizePageWidth(width)
	var result []formattedLine
	for ri, raw := range rawLines {
		result = append(result, wrapFormattedLines(WrapLines([]string{raw}, width), ri, "")...)
	}
	return result
}

func formatDocument(rawLines []string, width int, plainText bool) formattedDocument {
	if plainText {
		return formattedDocument{
			lines:   formatPlainTextWithProvenance(rawLines, width),
			anchors: make(map[string]int),
		}
	}
	return formatMarkdownDocument(rawLines, width, true)
}

func isMarkdownPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

func layoutFromFormatted(formatted []formattedLine, height int) bookLayout {
	height = normalizePageHeight(height)
	return bookLayout{
		formatted:     formatted,
		pages:         paginateFormatted(formatted, height),
		rawLinePages:  rawLinePages(formatted, height),
		pagePositions: pagePositions(formatted, height),
		lineSources:   lineSources(formatted),
		height:        height,
	}
}

func layoutFromDocument(document formattedDocument, height int) bookLayout {
	layout := layoutFromFormatted(document.lines, height)
	layout.anchors = document.anchors
	return layout
}

func normalizePageHeight(height int) int {
	if height < 1 {
		return DefaultPageHeight
	}
	return height
}

func normalizePageWidth(width int) int {
	if width < 1 {
		return fallbackPageWidth
	}
	return width
}

func paginateFormatted(formatted []formattedLine, height int) []Page {
	if len(formatted) == 0 {
		// Return a single empty page for empty content
		return []Page{{Lines: []string{}, Links: []Link{}}}
	}

	var pages []Page
	n := len(formatted)
	for i := 0; i < n; i += height {
		end := i + height
		if end > len(formatted) {
			end = len(formatted)
		}
		pageLines := make([]string, end-i)
		headings := make([]bool, end-i)
		for j, fl := range formatted[i:end] {
			pageLines[j] = fl.text
			headings[j] = fl.heading
		}
		pages = append(pages, Page{Lines: pageLines, Links: []Link{}, Headings: headings})
	}
	return pages
}

func rawLinePages(formatted []formattedLine, height int) map[int]int {
	pages := make(map[int]int)
	for fi, fl := range formatted {
		if fl.raw < 0 {
			continue
		}
		for raw := fl.raw; raw <= max(fl.raw, fl.endRaw); raw++ {
			if _, ok := pages[raw]; !ok {
				pages[raw] = fi / height
			}
		}
	}
	return pages
}

// pagePositions maps each page index to a representative source position: the
// first heading displayed on the page — the section the reader was in — else
// its first content line, falling back to the nearest preceding content
// line, then to 0. Page 0 always anchors to its first content line so the
// reader never skips the start of the document (#91). It is the reverse query
// of rawLinePages and stays in sync with it: both divide formatted indices by
// the same height.
func pagePositions(formatted []formattedLine, height int) []Position {
	if len(formatted) == 0 {
		// Empty content paginates to a single empty page.
		return []Position{{}}
	}
	n := len(formatted)
	sources := lineSources(formatted)
	anchors := make([]Position, (n+height-1)/height)
	var last Position
	end0 := min(height, n)
	anchors[0], last = firstPageAnchor(formatted[:end0], sources[:end0])
	for start := height; start < n; start += height {
		end := min(start+height, n)
		anchors[start/height], last = pageAnchor(formatted[start:end], sources[start:end], last)
	}
	return anchors
}

func lineSources(formatted []formattedLine) []Position {
	sources := make([]Position, len(formatted))
	for i, fl := range formatted {
		sources[i] = Position{Raw: fl.raw}
		if fl.hasOffset {
			sources[i].Offset = fl.sourceOffset
		} else if i > 0 && fl.raw >= 0 && fl.raw == formatted[i-1].raw {
			sources[i].Offset = sources[i-1].Offset + nonSpaceRunes(formatted[i-1].text)
		}
	}
	return sources
}

func nonSpaceRunes(s string) int {
	n := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}

func firstPageAnchor(lines []formattedLine, sources []Position) (Position, Position) {
	for i, fl := range lines {
		if fl.text != "" {
			return sources[i], sources[i]
		}
	}
	return Position{}, Position{}
}

func pageAnchor(pageLines []formattedLine, sources []Position, last Position) (Position, Position) {
	heading, content := -1, -1
	for i, fl := range pageLines {
		if fl.text == "" {
			continue
		}
		last = sources[i]
		if content < 0 {
			content = i
		}
		if heading < 0 && fl.heading {
			heading = i
		}
	}
	if heading >= 0 {
		return sources[heading], last
	}
	if content >= 0 {
		return sources[content], last
	}
	return last, last
}

// WrapLines wraps each line to fit within the given width.
func WrapLines(lines []string, width int) []string {
	var result []string
	for _, line := range lines {
		if stringWidth(line) <= width {
			result = append(result, line)
			continue
		}
		wrapped := wrapLine(line, width)
		result = append(result, wrapped...)
	}
	return result
}

// wrapLine breaks a single line into multiple lines of at most `width` columns.
func wrapLine(line string, width int) []string {
	wrapped := wrapLineWithLinks(line, width)
	result := make([]string, len(wrapped))
	for i, w := range wrapped {
		result[i] = w.text
	}
	return result
}

func wrapLineWithLinks(line string, width int) []wrappedLine {
	if width < 1 {
		return []wrappedLine{{text: line}}
	}

	if strings.TrimSpace(line) == "" {
		return []wrappedLine{{text: ""}}
	}

	lines := make([]string, 0, len(line)/width+1)
	var placements []placedLink
	current := make([]rune, 0, width)
	currentWidth := 0

	flush := func() {
		if len(current) > 0 {
			lines = append(lines, string(current))
			current = current[:0]
			currentWidth = 0
		}
	}

	eachWrapToken(line, func(token wrapToken) {
		wordLength := stringWidth(token.text)
		if shouldFlush(currentWidth, len(current) > 0, token.spaceBefore, wordLength, width) {
			flush()
		}
		appendPlacedWrapToken(token, &current, &currentWidth, width, &lines, &placements)
	})
	flush()

	result := make([]wrappedLine, len(lines))
	for i, text := range lines {
		result[i] = wrappedLine{text: text}
	}
	for _, placed := range placements {
		result[placed.opener.line].placed = append(result[placed.opener.line].placed, placed)
	}
	return result
}

func appendPlacedWrapToken(token wrapToken, current *[]rune, currentWidth *int, width int, lines *[]string, placements *[]placedLink) {
	if len(*current) > 0 && token.spaceBefore {
		*current = append(*current, ' ')
		(*currentWidth)++
	}
	if !token.link {
		appendWordRunes(current, currentWidth, token.text, width, lines)
		return
	}
	startLine, startByte := len(*lines), len(string(*current))
	appendLinkRunes(current, currentWidth, token.text, width, lines)
	placeTokenLinks(token, startLine, startByte, *lines, *current, placements)
}

func shiftPlaced(placed []placedLink, indent int) {
	for i := range placed {
		if placed[i].opener.line == 0 {
			placed[i].opener.start += indent
			placed[i].opener.end += indent
		}
		for j := range placed[i].pieces {
			if placed[i].pieces[j].line == 0 {
				placed[i].pieces[j].start += indent
				placed[i].pieces[j].end += indent
			}
		}
	}
}

func placeTokenLinks(token wrapToken, startLine, startByte int, lines []string, current []rune, placements *[]placedLink) {
	match := extractLinkMatches(token.text[token.linkStart:])[0]
	match.start += token.linkStart
	placed := placedLink{link: match.link}
	labelStart, labelEnd := match.start+1, match.start+1+len(match.link.Label)
	line, pos := startLine, startByte
	currentBytes := len(string(current))
	for offset, r := range token.text {
		if pos >= tokenLineLength(lines, currentBytes, line) {
			line++
			pos = 0
		}
		size := utf8.RuneLen(r)
		if offset == match.start {
			placed.opener = labelPiece{line: line, start: pos, end: pos + size}
		}
		if offset >= labelStart && offset < labelEnd {
			n := len(placed.pieces)
			if n > 0 && placed.pieces[n-1].line == line && placed.pieces[n-1].end == pos {
				placed.pieces[n-1].end += size
			} else {
				placed.pieces = append(placed.pieces, labelPiece{line: line, start: pos, end: pos + size})
			}
		}
		pos += size
	}
	*placements = append(*placements, placed)
}

func tokenLineLength(lines []string, currentBytes, line int) int {
	if line < len(lines) {
		return len(lines[line])
	}
	return currentBytes
}

// shouldFlush reports whether the current line should be flushed before adding
// a word of wordLen, optionally preceded by a space separator.
func shouldFlush(currentLen int, hasContent, hasSpaceBefore bool, wordLen, width int) bool {
	if !hasContent {
		return false
	}
	if hasSpaceBefore {
		return currentLen+1+wordLen > width
	}
	return currentLen+wordLen > width
}

// appendWordRunes appends a word's runes to current, flushing when the line
// reaches the configured width.
func appendWordRunes(current *[]rune, currentWidth *int, word string, width int, lines *[]string) {
	for _, r := range word {
		rw := runewidth.RuneWidth(r)
		if len(*current) > 0 && *currentWidth+rw > width {
			*lines = append(*lines, string(*current))
			*current = (*current)[:0]
			*currentWidth = 0
		}
		*current = append(*current, r)
		*currentWidth += rw
		if *currentWidth >= width {
			*lines = append(*lines, string(*current))
			*current = (*current)[:0]
			*currentWidth = 0
		}
	}
}

// appendLinkRunes appends a link token to current. A link that fits within the
// width is kept whole so its markup stays on one display line. A link wider
// than the page is hard-broken at the character level via appendWordRunes, so
// no display line exceeds width.
func appendLinkRunes(current *[]rune, currentWidth *int, text string, width int, lines *[]string) {
	if stringWidth(text) > width {
		appendWordRunes(current, currentWidth, text, width, lines)
		return
	}
	*current = append(*current, []rune(text)...)
	*currentWidth += stringWidth(text)
}

type wrapToken struct {
	text        string
	spaceBefore bool
	link        bool
	linkStart   int
}

func matchLinkToken(line string, i int, spans [][2]int) int {
	if line[i] != '[' {
		return 0
	}
	match := linkRegex.FindStringIndex(line[i:])
	if match == nil || match[0] != 0 || IsInlineCodeRange(spans, i, i+match[1]) {
		return 0
	}
	return match[1]
}

// eachWrapToken consumes words as they are scanned, without retaining a token slice.
func eachWrapToken(line string, consume func(wrapToken)) {
	spans := InlineCodeSpans(line)
	spaceBefore := false
	n := len(line)
	for i := 0; i < n; {
		r, size := utf8.DecodeRuneInString(line[i:])
		if unicode.IsSpace(r) {
			spaceBefore = true
			i += size
			continue
		}

		// Scan a word. A link may start mid-word when prefix punctuation (e.g.
		// "(", '"') abuts the opening bracket; the prefix stays glued to the
		// link token so the markup is kept whole on one display line.
		start := i
		link := false
		for i < n {
			r, size = utf8.DecodeRuneInString(line[i:])
			if unicode.IsSpace(r) {
				break
			}
			if matchLen := matchLinkToken(line, i, spans); matchLen > 0 {
				consume(wrapToken{text: line[start : i+matchLen], spaceBefore: spaceBefore, link: true, linkStart: i - start})
				i += matchLen
				link = true
				break
			}
			i += size
		}
		if link {
			spaceBefore = false
			continue
		}
		consume(wrapToken{text: line[start:i], spaceBefore: spaceBefore})
		spaceBefore = false
	}
}

func stringWidth(s string) int {
	return runewidth.StringWidth(s)
}

func runeLen(s string) int {
	return stringWidth(s)
}

// NewBook creates a fully paginated book from a file path.
func NewBook(path string, width, height int) (*Book, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	return Read(f, deriveTitle(path), width, height, !isMarkdownPath(path))
}

// Read creates a fully paginated book from a reader.
func Read(r io.Reader, title string, width, height int, plainText bool) (*Book, error) {
	lines, err := readLines(r)
	if err != nil {
		return nil, err
	}
	b := &Book{
		Title:     title,
		rawLines:  lines,
		plainText: plainText,
	}
	b.layout(width, height)
	return b, nil
}

// Reflow re-paginates the book for new dimensions.
func (b *Book) Reflow(width, height int) {
	b.layout(width, height)
}

// layout paginates rawLines at the given geometry and sets every laid-out
// field. Source links are computed once and cached across reflows.
func (b *Book) layout(width, height int) {
	if b.sourceLinks.links == nil {
		b.sourceLinks = sourceLinksFor(b.rawLines, b.plainText)
	}
	width = normalizePageWidth(width)
	layout := layoutFromDocument(formatDocument(b.rawLines, width, b.plainText), height)
	b.pageWidth = width
	b.pageHeight = layout.height
	b.pages = attachLinks(layout.pages, b.rawLines, layout.formatted, layout.height, b.sourceLinks)
	b.anchors = layout.anchors
	b.rawLinePages = layout.rawLinePages
	b.pagePositions = layout.pagePositions
	b.lineSources = layout.lineSources
}

// PageCount returns the number of laid-out pages.
func (b *Book) PageCount() int {
	return len(b.pages)
}

// Page returns a copy of the page at index i, so callers cannot change the
// book's own page data.
func (b *Book) Page(i int) Page {
	p := b.pages[i]
	return Page{
		Lines:      slices.Clone(p.Lines),
		Links:      slices.Clone(p.Links),
		Headings:   slices.Clone(p.Headings),
		LabelSpans: slices.Clone(p.LabelSpans),
	}
}

// PageHeight returns the number of lines per laid-out page.
func (b *Book) PageHeight() int {
	return b.pageHeight
}

// PageForAnchor returns the page index containing the given anchor.
// The anchor is normalized like heading text, so "#Usage" finds "## Usage".
// Returns -1 if the anchor is not found.
func (b *Book) PageForAnchor(anchor string) int {
	lineIdx, ok := b.anchors[NormalizeAnchor(anchor)]
	if !ok {
		return -1
	}

	if page, ok := b.rawLinePages[lineIdx]; ok {
		return page
	}

	return -1
}

// RawLineForPage returns the raw source line the page is anchored to; see
// PositionForPage.
func (b *Book) RawLineForPage(page int) int {
	return b.PositionForPage(page).Raw
}

// PositionForPage returns the source position the page is anchored to: the
// first content line for page 0, otherwise the first heading displayed on
// the page, else its first content line, falling back to the nearest
// preceding content line, then to 0. Call it before Reflow; PageForPosition
// is the matching query after, so a page index survives a re-pagination by
// round-tripping through its source position.
func (b *Book) PositionForPage(page int) Position {
	if len(b.pagePositions) == 0 {
		return Position{}
	}
	if page < 0 {
		return b.pagePositions[0]
	}
	if page >= len(b.pagePositions) {
		return b.pagePositions[len(b.pagePositions)-1]
	}
	return b.pagePositions[page]
}

// PageForPosition returns the page displaying the given source position: the
// page of the last display line of pos.Raw that starts at or before
// pos.Offset, falling back to PageForRawLine.
func (b *Book) PageForPosition(pos Position) int {
	first, ok := b.rawLinePages[pos.Raw]
	if !ok || pos.Offset <= 0 {
		return b.PageForRawLine(pos.Raw)
	}
	start := min(first*b.pageHeight, len(b.lineSources))
	best := -1
	for i, src := range b.lineSources[start:] {
		if src.Raw > pos.Raw {
			break
		}
		if src.Raw == pos.Raw && src.Offset <= pos.Offset {
			best = start + i
		}
	}
	if best < 0 {
		return b.PageForRawLine(pos.Raw)
	}
	return b.clampPage(best / b.pageHeight)
}

// PageForRawLine returns the page on which the given raw source line is
// displayed: its first display page, falling back to the nearest preceding
// raw line, then to a clamp into the page range.
func (b *Book) PageForRawLine(raw int) int {
	if page, ok := b.rawLinePages[raw]; ok {
		return b.clampPage(page)
	}

	// The raw line is not displayed (e.g. trailing blank lines): fall back
	// to the nearest preceding raw line. Formatted raw indices are
	// non-decreasing, so the largest mapped line below raw owns the
	// largest such page.
	best := -1
	for line, page := range b.rawLinePages {
		if line < raw && page > best {
			best = page
		}
	}
	if best >= 0 {
		return b.clampPage(best)
	}

	return 0
}

// clampPage pins a page index into the valid page range, treating an empty
// book as a single page 0.
func (b *Book) clampPage(page int) int {
	last := len(b.pages) - 1
	if last < 0 {
		return 0
	}
	return max(0, min(page, last))
}
