// Package tui implements the Bubble Tea TUI for the e-reader.
package tui

import (
	"fmt"
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
	maxWidth := 72
	m.contentWidth = m.termWidth - 4 // 2 chars padding each side
	if m.contentWidth > maxWidth {
		m.contentWidth = maxWidth
	}
	if m.contentWidth < 20 {
		m.contentWidth = 20
	}

	// Content height: terminal height minus header (3) and footer (3)
	m.contentHeight = m.termHeight - 6
	if m.contentHeight < 5 {
		m.contentHeight = 5
	}

	if m.book != nil {
		oldPage := m.currentPage
		m.book.Reflow(m.contentWidth, m.contentHeight)
		// Try to stay on the same page, clamped
		if oldPage >= len(m.book.Pages) {
			m.currentPage = len(m.book.Pages) - 1
		}
		if m.currentPage < 0 {
			m.currentPage = 0
		}
		m.selectedLink = -1
	}
	return m
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Quit
	case key.Matches(msg, keys.Quit):
		m.quitting = true
		return m, tea.Quit

	// Next page
	case key.Matches(msg, keys.NextPage):
		if m.book != nil && m.currentPage < len(m.book.Pages)-1 {
			m.currentPage++
			m.selectedLink = -1
		}
		return m, nil

	// Previous page
	case key.Matches(msg, keys.PrevPage):
		if m.book != nil && m.currentPage > 0 {
			m.currentPage--
			m.selectedLink = -1
		}
		return m, nil

	// First page
	case key.Matches(msg, keys.FirstPage):
		m.currentPage = 0
		m.selectedLink = -1
		return m, nil

	// Last page
	case key.Matches(msg, keys.LastPage):
		if m.book != nil {
			m.currentPage = len(m.book.Pages) - 1
		}
		m.selectedLink = -1
		return m, nil

	// Tab through links
	case key.Matches(msg, keys.NextLink):
		if m.book != nil && m.currentPage < len(m.book.Pages) {
			page := m.book.Pages[m.currentPage]
			if len(page.Links) > 0 {
				m.selectedLink = (m.selectedLink + 1) % len(page.Links)
			}
		}
		return m, nil

	// Shift+Tab: previous link
	case key.Matches(msg, keys.PrevLink):
		if m.book != nil && m.currentPage < len(m.book.Pages) {
			page := m.book.Pages[m.currentPage]
			if len(page.Links) > 0 {
				m.selectedLink--
				if m.selectedLink < 0 {
					m.selectedLink = len(page.Links) - 1
				}
			}
		}
		return m, nil

	// Enter: follow selected link
	case key.Matches(msg, keys.FollowLink):
		if m.book != nil && m.selectedLink >= 0 && m.currentPage < len(m.book.Pages) {
			page := m.book.Pages[m.currentPage]
			if m.selectedLink < len(page.Links) {
				target := page.Links[m.selectedLink].Target
				dest := m.book.PageForAnchor(target)
				if dest >= 0 {
					m.history = append(m.history, m.currentPage)
					m.currentPage = dest
					m.selectedLink = -1
				}
			}
		}
		return m, nil

	// Back navigation
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

	// Stack vertically and center horizontally
	full := lipgloss.JoinVertical(lipgloss.Center, header, content, footer)

	// Center in terminal
	return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, full)
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

	// Build a set of link lines for highlighting
	linkLineSet := make(map[int][]book.Link)
	for _, lnk := range page.Links {
		linkLineSet[lnk.LineOnPage] = append(linkLineSet[lnk.LineOnPage], lnk)
	}

	// Find the selected link if any
	var selectedTarget string
	if m.selectedLink >= 0 && m.selectedLink < len(page.Links) {
		selectedTarget = page.Links[m.selectedLink].Target
	}

	contentStyle := lipgloss.NewStyle().
		Width(m.contentWidth).
		Height(m.contentHeight).
		Padding(0, 1)

	// Render each line
	var rendered []string
	for i, line := range page.Lines {
		styledLine := m.styleLine(line, i, linkLineSet, selectedTarget)
		rendered = append(rendered, styledLine)
	}

	// Pad to full height
	for len(rendered) < m.contentHeight {
		rendered = append(rendered, "")
	}

	body := strings.Join(rendered, "\n")
	return contentStyle.Render(body)
}

func (m Model) styleLine(line string, lineIdx int, linkLineSet map[int][]book.Link, selectedTarget string) string {
	// Check if this line has links
	links, hasLinks := linkLineSet[lineIdx]

	if !hasLinks {
		// Check for heading styling
		if isHeading(line) {
			headingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("117")).
				Bold(true)
			return headingStyle.Render(line)
		}
		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
		return textStyle.Render(line)
	}

	// Highlight links in the line
	result := line
	for _, lnk := range links {
		linkText := lnk.Label
		if lnk.Target == selectedTarget {
			// Selected link: inverse colors
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("117")).
				Bold(true).
				Underline(true)
			result = strings.Replace(result, linkText, selectedStyle.Render(linkText), 1)
		} else {
			linkStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("75")).
				Underline(true)
			result = strings.Replace(result, linkText, linkStyle.Render(linkText), 1)
		}
	}

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	return textStyle.Render(result)
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
