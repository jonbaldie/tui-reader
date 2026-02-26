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

// Paginate splits raw lines into pages of the given dimensions, wrapping long lines.
func Paginate(rawLines []string, width, height int) []Page {
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 20
	}

	wrapped := WrapLines(rawLines, width)
	if len(wrapped) == 0 {
		// Return a single empty page for empty content
		return []Page{{Lines: []string{}, Links: []Link{}}}
	}

	var pages []Page
	for i := 0; i < len(wrapped); i += height {
		end := i + height
		if end > len(wrapped) {
			end = len(wrapped)
		}
		pageLines := make([]string, end-i)
		copy(pageLines, wrapped[i:end])
		pages = append(pages, Page{Lines: pageLines, Links: []Link{}})
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
	pages := Paginate(lines, width, height)
	pages = AttachLinks(pages, lines, width, height)

	return &Book{
		Title:      title,
		RawLines:   lines,
		Pages:      pages,
		Anchors:    anchors,
		PageWidth:  width,
		PageHeight: height,
	}, nil
}

// Reflow re-paginates the book for new dimensions.
func (b *Book) Reflow(width, height int) {
	b.PageWidth = width
	b.PageHeight = height
	b.Pages = Paginate(b.RawLines, width, height)
	b.Pages = AttachLinks(b.Pages, b.RawLines, width, height)
}

// PageForAnchor returns the page index containing the given anchor.
// Returns -1 if the anchor is not found.
func (b *Book) PageForAnchor(anchor string) int {
	lineIdx, ok := b.Anchors[anchor]
	if !ok {
		return -1
	}

	// We need to figure out which page this raw line ended up on after wrapping.
	// Rewrap lines up to lineIdx to count how many wrapped lines precede it.
	wrappedBefore := 0
	for i := 0; i < lineIdx && i < len(b.RawLines); i++ {
		w := WrapLines([]string{b.RawLines[i]}, b.PageWidth)
		wrappedBefore += len(w)
	}

	if b.PageHeight <= 0 {
		return 0
	}
	return wrappedBefore / b.PageHeight
}
