// Package book handles loading, parsing, and paginating readable files.
package book

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
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
	Lines []string // lines of text to display
	Links []Link   // clickable links on this page
}

// Book is a loaded and paginated document.
type Book struct {
	Title      string
	RawLines   []string
	Pages      []Page
	Anchors    map[string]int // anchor name -> line index in RawLines
	PageWidth  int
	PageHeight int

	rawLinePages map[int]int
}

// Load reads a file from disk and returns its raw content lines.
func Load(path string) (title string, lines []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("cannot open file: %w", err)
	}

	content := string(data)
	if !utf8.ValidString(content) {
		return "", nil, fmt.Errorf("file is not valid UTF-8")
	}

	// Derive title from filename
	title = deriveTitle(path)

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines = strings.Split(content, "\n")
	return title, lines, nil
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
// are non-decreasing.
type formattedLine struct {
	text string
	raw  int
}

type bookLayout struct {
	formatted    []formattedLine
	pages        []Page
	rawLinePages map[int]int
	height       int
}

// formatParagraphsWithProvenance is the single owner of the paragraph
// formatting rules. In one pass it produces each display line together with the
// raw source line it came from, so callers never re-derive the formatting
// algorithm to recover provenance. FormatParagraphs is a thin projection over
// it.
func formatParagraphsWithProvenance(rawLines []string, width int) []formattedLine {
	if width < 1 {
		width = 80
	}

	var result []formattedLine
	firstParagraph := true

	for ri, raw := range rawLines {
		trimmed := strings.TrimSpace(raw)

		// Blank lines in source: preserve as spacing, mapped to the source line.
		if trimmed == "" {
			// Only add a blank line if we haven't just added one
			if len(result) > 0 && result[len(result)-1].text != "" {
				result = append(result, formattedLine{text: "", raw: ri})
			}
			continue
		}

		// Insert blank line between paragraphs (not before the first). This
		// spacer has no source line, so its provenance is -1.
		if !firstParagraph {
			if len(result) > 0 && result[len(result)-1].text != "" {
				result = append(result, formattedLine{text: "", raw: -1})
			}
		}

		// Detect if this is a heading or special line (don't indent those)
		isSpecial := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "    ")

		// For non-first, non-special paragraphs: wrap at width-2 to leave room
		// for the 2-space indent, then prepend it to the first line.
		shouldIndent := !firstParagraph && !isSpecial
		wrapWidth := width
		if shouldIndent {
			wrapWidth = width - 2
			if wrapWidth < 10 {
				wrapWidth = 10
			}
		}

		wrapped := WrapLines([]string{raw}, wrapWidth)

		if shouldIndent && len(wrapped) > 0 {
			wrapped[0] = "  " + wrapped[0]
		}

		// Every wrapped line of this paragraph shares the same source line.
		for _, w := range wrapped {
			result = append(result, formattedLine{text: w, raw: ri})
		}
		firstParagraph = false
	}

	return result
}

// FormatParagraphs takes raw lines and produces display-ready lines with
// paragraph indentation and spacing. Each non-empty raw line is treated as
// a paragraph. Non-first paragraphs get a 2-space indent on their first
// wrapped line, and a blank line is inserted between paragraphs.
func FormatParagraphs(rawLines []string, width int) []string {
	formatted := formatParagraphsWithProvenance(rawLines, width)
	result := make([]string, len(formatted))
	for i, fl := range formatted {
		result[i] = fl.text
	}
	return result
}

// Paginate splits raw lines into pages of the given dimensions, wrapping long lines.
func Paginate(rawLines []string, width, height int) []Page {
	return buildBookLayout(rawLines, width, height).pages
}

func buildBookLayout(rawLines []string, width, height int) bookLayout {
	formatted := formatParagraphsWithProvenance(rawLines, width)
	height = normalizePageHeight(height)
	return bookLayout{
		formatted:    formatted,
		pages:        paginateFormatted(formatted, height),
		rawLinePages: rawLinePages(formatted, height),
		height:       height,
	}
}

func normalizePageHeight(height int) int {
	if height < 1 {
		return 20
	}
	return height
}

func paginateFormatted(formatted []formattedLine, height int) []Page {
	if len(formatted) == 0 {
		// Return a single empty page for empty content
		return []Page{{Lines: []string{}, Links: []Link{}}}
	}

	var pages []Page
	for i := 0; i < len(formatted); i += height {
		end := i + height
		if end > len(formatted) {
			end = len(formatted)
		}
		pageLines := make([]string, end-i)
		for j, fl := range formatted[i:end] {
			pageLines[j] = fl.text
		}
		pages = append(pages, Page{Lines: pageLines, Links: []Link{}})
	}
	return pages
}

func rawLinePages(formatted []formattedLine, height int) map[int]int {
	pages := make(map[int]int)
	for fi, fl := range formatted {
		if fl.raw < 0 {
			continue
		}
		if _, ok := pages[fl.raw]; !ok {
			pages[fl.raw] = fi / height
		}
	}
	return pages
}

// WrapLines wraps each line to fit within the given width.
func WrapLines(lines []string, width int) []string {
	var result []string
	for _, line := range lines {
		if runeLen(line) <= width {
			result = append(result, line)
			continue
		}
		wrapped := wrapLine(line, width)
		result = append(result, wrapped...)
	}
	return result
}

// wrapLine breaks a single line into multiple lines of at most `width` runes.
func wrapLine(line string, width int) []string {
	if width < 1 {
		return []string{line}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""

	for _, word := range words {
		if current == "" {
			current = word
		} else if runeLen(current)+1+runeLen(word) <= width {
			current += " " + word
		} else {
			// Flush current, hard-breaking if needed
			for runeLen(current) > width {
				lines = append(lines, string([]rune(current)[:width]))
				current = string([]rune(current)[width:])
			}
			if current != "" {
				lines = append(lines, current)
			}
			current = word
		}
		// Hard-break current if it's a single word longer than width
		for runeLen(current) > width {
			lines = append(lines, string([]rune(current)[:width]))
			current = string([]rune(current)[width:])
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

// NewBook creates a fully paginated book from a file path.
func NewBook(path string, width, height int) (*Book, error) {
	title, lines, err := Load(path)
	if err != nil {
		return nil, err
	}

	anchors := ExtractAnchors(lines)
	layout := buildBookLayout(lines, width, height)
	pages := attachLinks(layout.pages, lines, layout.formatted, layout.height)

	return &Book{
		Title:        title,
		RawLines:     lines,
		Pages:        pages,
		Anchors:      anchors,
		PageWidth:    width,
		PageHeight:   height,
		rawLinePages: layout.rawLinePages,
	}, nil
}

// Reflow re-paginates the book for new dimensions.
func (b *Book) Reflow(width, height int) {
	b.PageWidth = width
	b.PageHeight = height
	layout := buildBookLayout(b.RawLines, width, height)
	b.Pages = attachLinks(layout.pages, b.RawLines, layout.formatted, layout.height)
	b.rawLinePages = layout.rawLinePages
}

// PageForAnchor returns the page index containing the given anchor.
// Returns -1 if the anchor is not found.
func (b *Book) PageForAnchor(anchor string) int {
	lineIdx, ok := b.Anchors[anchor]
	if !ok {
		return -1
	}

	if b.PageHeight <= 0 {
		return 0
	}

	if b.rawLinePages == nil {
		layout := buildBookLayout(b.RawLines, b.PageWidth, b.PageHeight)
		b.rawLinePages = layout.rawLinePages
	}

	if page, ok := b.rawLinePages[lineIdx]; ok {
		return page
	}

	return -1
}
