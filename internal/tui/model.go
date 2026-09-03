// Package tui implements the Bubble Tea TUI for the e-reader.
package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonbaldie/tui-reader/internal/book"
)

// Model is the Bubble Tea model for the reader.
type Model struct {
	book          *book.Book
	currentPage   int
	selectedLink  int // -1 means no link selected
	termWidth     int
	termHeight    int
	contentWidth  int
	contentHeight int
	err           error
	quitting      bool
	history       []int // page history stack for back navigation
}

// NewModel creates a new TUI model for the given file.
func NewModel(path string) Model {
	// We'll start with default dimensions; they'll be updated on WindowSizeMsg
	b, err := book.NewBook(path, 60, 20)
	return Model{
		book:          b,
		currentPage:   0,
		selectedLink:  -1,
		contentWidth:  60,
		contentHeight: 20,
		err:           err,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m = m.recalcLayout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) recalcLayout() Model {
	// Content area: max 72 chars wide, centered, with margins
	newWidth := min(72, max(20, m.termWidth-4))

	// Content height: terminal height minus header (2), footer (3), and top/bottom padding (2)
	newHeight := max(5, m.termHeight-7)

	if newWidth == m.contentWidth && newHeight == m.contentHeight {
		return m
	}

	m.contentWidth = newWidth
	m.contentHeight = newHeight

	if m.book != nil {
		m.book.Reflow(m.contentWidth, m.contentHeight)
		// Try to stay on the same page, clamped
		numPages := len(m.book.Pages)
		m.currentPage = max(0, min(m.currentPage, numPages-1))
		for i := range m.history {
			m.history[i] = max(0, min(m.history[i], numPages-1))
		}
		m.selectedLink = -1
	}
	return m
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, keys.NextPage):
		if canNextPage(m) {
			m.currentPage++
			m.selectedLink = -1
		}
		return m, nil

	case key.Matches(msg, keys.PrevPage):
		if canPrevPage(m) {
			m.currentPage--
			m.selectedLink = -1
		}
		return m, nil

	case key.Matches(msg, keys.FirstPage):
		m.currentPage = 0
		m.selectedLink = -1
		return m, nil

	case key.Matches(msg, keys.LastPage):
		if m.book != nil {
			m.currentPage = len(m.book.Pages) - 1
		}
		m.selectedLink = -1
		return m, nil

	default:
		return m.handleLinkKey(msg)
	}
}

func (m Model) handleLinkKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.NextLink):
		links := currentLinks(m)
		if len(links) > 0 {
			m.selectedLink = (m.selectedLink + 1) % len(links)
		}
		return m, nil

	case key.Matches(msg, keys.PrevLink):
		links := currentLinks(m)
		if len(links) > 0 {
			m.selectedLink--
			if m.selectedLink < 0 {
				m.selectedLink = len(links) - 1
			}
		}
		return m, nil

	case key.Matches(msg, keys.FollowLink):
		return followLink(m)

	case key.Matches(msg, keys.GoBack):
		if len(m.history) > 0 {
			m.currentPage = m.history[len(m.history)-1]
			m.history = m.history[:len(m.history)-1]
			m.selectedLink = -1
		}
		return m, nil
	}
	return m, nil
}

// canNextPage reports whether there is a next page to advance to.
func canNextPage(m Model) bool {
	return m.book != nil && m.currentPage < len(m.book.Pages)-1
}

// canPrevPage reports whether there is a previous page to go back to.
func canPrevPage(m Model) bool {
	return m.book != nil && m.currentPage > 0
}

// currentLinks returns the links on the current page, or nil if there is no
// valid current page.
func currentLinks(m Model) []book.Link {
	if m.book == nil || m.currentPage >= len(m.book.Pages) {
		return nil
	}
	return m.book.Pages[m.currentPage].Links
}

// followLink navigates to the anchor target of the selected link, pushing the
// current page onto the history stack.
func followLink(m Model) (tea.Model, tea.Cmd) {
	links := currentLinks(m)
	if m.selectedLink < 0 {
		return m, nil
	}
	if m.selectedLink >= len(links) {
		return m, nil
	}
	dest := m.book.PageForAnchor(links[m.selectedLink].Target)
	if dest < 0 {
		return m, nil
	}
	m.history = append(m.history, m.currentPage)
	m.currentPage = dest
	m.selectedLink = -1
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.err != nil {
		return m.renderError()
	}
	if m.book == nil {
		return "Loading..."
	}

	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// Stack vertically and center horizontally in terminal
	full := lipgloss.JoinVertical(lipgloss.Center, header, content, footer)

	// Place with horizontal centering; use top position so we control vertical padding
	// Add 1 empty line top padding for symmetry with the bottom padding
	return "\n" + lipgloss.Place(m.termWidth, m.termHeight-1, lipgloss.Center, lipgloss.Top, full)
}

func (m Model) renderError() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true).
		Padding(1, 2)

	msg := fmt.Sprintf("Error: %v", m.err)
	box := style.Render(msg)
	return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderHeader() string {
	title := m.book.Title

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.contentWidth)

	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(m.contentWidth)

	divider := dividerStyle.Render(strings.Repeat("─", m.contentWidth))
	return lipgloss.JoinVertical(lipgloss.Center, titleStyle.Render(title), divider)
}

func (m Model) renderContent() string {
	if m.currentPage >= len(m.book.Pages) {
		return ""
	}

	page := m.book.Pages[m.currentPage]

	// Build a set of link keys per line for highlighting
	var linksByLine map[int]map[linkKey]struct{}
	if len(page.Links) > 0 {
		linksByLine = make(map[int]map[linkKey]struct{})
		for _, lnk := range page.Links {
			set := linksByLine[lnk.LineOnPage]
			if set == nil {
				set = make(map[linkKey]struct{})
				linksByLine[lnk.LineOnPage] = set
			}
			set[linkKey{label: lnk.Label, target: lnk.Target}] = struct{}{}
		}
	}

	// Find the selected link if any
	var selectedTarget string
	if m.selectedLink >= 0 && m.selectedLink < len(page.Links) {
		selectedTarget = page.Links[m.selectedLink].Target
	}

	contentStyle := lipgloss.NewStyle().
		Width(m.contentWidth).
		Height(m.contentHeight)

	// Render each line
	rendered := make([]string, 0, m.contentHeight)
	for i, line := range page.Lines {
		styledLine := styleLine(line, i, linksByLine, selectedTarget)
		rendered = append(rendered, styledLine)
	}

	// Pad to full height
	padCount := m.contentHeight - len(rendered)
	for i := 0; i < padCount; i++ {
		rendered = append(rendered, "")
	}

	body := strings.Join(rendered, "\n")
	return contentStyle.Render(body)
}

var (
	headingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true)

	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedLinkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("117")).
				Bold(true).
				Underline(true)

	unselectedLinkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("75")).
				Underline(true)
)

func styleLine(line string, lineIdx int, linksByLine map[int]map[linkKey]struct{}, selectedTarget string) string {
	// Check if this line has links
	links, hasLinks := linksByLine[lineIdx]

	if !hasLinks {
		// Check for heading styling
		if isHeading(line) {
			return headingStyle.Render(line)
		}
		return textStyle.Render(line)
	}

	result := styleLinkMarkup(line, links, selectedTarget)
	return textStyle.Render(result)
}

type linkKey struct {
	label  string
	target string
}

var internalLinkMarkup = regexp.MustCompile(`\[([^\]]+)\]\(#([^)]+)\)`)

// styleLinkMarkup scans Markdown link markup once from left to right. It keeps
// link syntax intact and styles only labels that are attached to this page
// line, so repeated labels do not cause ANSI-decorated text to be revisited.
func styleLinkMarkup(line string, links map[linkKey]struct{}, selectedTarget string) string {
	matches := internalLinkMarkup.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return line
	}

	var sb strings.Builder
	sb.Grow(len(line) + len(matches)*32)
	lastIdx := 0

	for _, loc := range matches {
		matchStart, matchEnd := loc[0], loc[1]
		labelStart, labelEnd := loc[2], loc[3]
		targetStart, targetEnd := loc[4], loc[5]

		sb.WriteString(line[lastIdx:matchStart])
		lastIdx = matchEnd

		label := line[labelStart:labelEnd]
		target := line[targetStart:targetEnd]

		if _, ok := links[linkKey{label: label, target: target}]; ok {
			sb.WriteByte('[')
			if target == selectedTarget {
				sb.WriteString(selectedLinkStyle.Render(label))
			} else {
				sb.WriteString(unselectedLinkStyle.Render(label))
			}
			sb.WriteString("](#")
			sb.WriteString(target)
			sb.WriteByte(')')
		} else {
			sb.WriteString(line[matchStart:matchEnd])
		}
	}
	sb.WriteString(line[lastIdx:])
	return sb.String()
}

func isHeading(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) > 0 && trimmed[0] == '#'
}

func (m Model) renderFooter() string {
	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(m.contentWidth)

	pageInfo := fmt.Sprintf("Page %d of %d", m.currentPage+1, len(m.book.Pages))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(m.contentWidth)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(m.contentWidth)

	divider := dividerStyle.Render(strings.Repeat("─", m.contentWidth))
	info := infoStyle.Render(pageInfo)
	help := helpStyle.Render("←/→ page • tab link • enter follow • b back • q quit")

	return lipgloss.JoinVertical(lipgloss.Center, divider, info, help)
}

// Exported accessors for testing

// CurrentPage returns the current page index.
func (m Model) CurrentPage() int { return m.currentPage }

// SelectedLink returns the currently selected link index.
func (m Model) SelectedLink() int { return m.selectedLink }

// Err returns any error in the model.
func (m Model) Err() error { return m.err }

// BookRef returns the loaded book (may be nil).
func (m Model) BookRef() *book.Book { return m.book }

// History returns the navigation history stack.
func (m Model) History() []int { return m.history }
