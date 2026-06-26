package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/jonbaldie/tui-reader/internal/tui"
)

const usage = "Usage: tui-reader [--dump[=N]] <file>\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the program and returns a process exit code. args are the
// command-line arguments excluding the program name.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprint(stderr, usage)
		return 1
	}

	path, dumpMode, dumpPages := parseArgs(args)

	if path == "" {
		fmt.Fprint(stderr, usage)
		return 1
	}

	if dumpMode {
		b, err := book.NewBook(path, 62, 20)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Fprint(stdout, renderDump(b, dumpPages))
		return 0
	}

	model := tui.NewModel(path)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// parseArgs parses CLI arguments into a file path and dump options. The last
// non-flag argument wins as the path. --dump dumps all pages; --dump=N limits
// the dump to N pages.
func parseArgs(args []string) (path string, dumpMode bool, dumpPages int) {
	for _, arg := range args {
		if arg == "--dump" {
			dumpMode = true
		} else if strings.HasPrefix(arg, "--dump=") {
			dumpMode = true
			fmt.Sscanf(arg, "--dump=%d", &dumpPages)
		} else {
			path = arg
		}
	}
	return path, dumpMode, dumpPages
}

// renderDump renders a textual dump of the book's pages. maxPages limits the
// number of pages rendered; 0 means render all pages.
func renderDump(b *book.Book, maxPages int) string {
	var sb strings.Builder

	total := len(b.Pages)
	if maxPages > 0 && maxPages < total {
		total = maxPages
	}

	for i := 0; i < total; i++ {
		fmt.Fprintf(&sb, "┌─── %s ── Page %d of %d ───┐\n", b.Title, i+1, len(b.Pages))
		fmt.Fprintln(&sb, "│")
		for _, line := range b.Pages[i].Lines {
			fmt.Fprintf(&sb, "│  %s\n", line)
		}
		// Pad to page height
		for j := len(b.Pages[i].Lines); j < 20; j++ {
			fmt.Fprintln(&sb, "│")
		}
		fmt.Fprintln(&sb, "│")
		fmt.Fprintln(&sb, "└────────────────────────────────┘")
		if i < total-1 {
			fmt.Fprintln(&sb)
		}
	}
	return sb.String()
}
