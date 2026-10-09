package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/jonbaldie/tui-reader/internal/tui"
)

const usage = "Usage: tui-reader [--dump[=N]] <file>\n"

var osExit = os.Exit

// interactiveOptions are extra Bubble Tea options. Production leaves this nil.
// Tests set it so interactive mode can run without a terminal.
var interactiveOptions []tea.ProgramOption

type parsedArgs struct {
	path      string
	dumpMode  bool
	dumpPages int
}

func main() {
	osExit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the program and returns a process exit code. args are the
// command-line arguments excluding the program name.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprint(stderr, usage)
		return 1
	}

	parsed, err := parseArgs(args)

	if err != nil {
		fmt.Fprint(stderr, usage)
		return 1
	}

	if parsed.path == "" {
		fmt.Fprint(stderr, usage)
		return 1
	}

	b, err := book.NewBook(parsed.path, book.DefaultPageWidth, book.DefaultPageHeight)
	if parsed.dumpMode {
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Fprint(stdout, renderDump(b, parsed.dumpPages))
		return 0
	}

	return runInteractive(b, err, stderr)
}

// runInteractive starts the reader. A load error still shows the in-TUI error
// screen; after it exits, the process status is 1.
func runInteractive(b *book.Book, loadErr error, stderr io.Writer) int {
	var model tui.Model
	if loadErr != nil {
		model = tui.NewErrorModel(loadErr)
	} else {
		model = tui.NewModelFromBook(b)
	}
	opts := append([]tea.ProgramOption{tea.WithAltScreen()}, interactiveOptions...)
	p := tea.NewProgram(model, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if loadErr != nil {
		return 1
	}
	return 0
}

// parseArgs parses CLI arguments into a file path and dump options. The last
// non-flag argument wins as the path. --dump dumps all pages; --dump=N limits
// the dump to N pages.
func parseArgs(args []string) (parsedArgs, error) {
	var parsed parsedArgs

	for _, arg := range args {
		if arg == "--dump" {
			parsed.dumpMode = true
		} else if strings.HasPrefix(arg, "--dump=") {
			value := strings.TrimPrefix(arg, "--dump=")
			pages, parseErr := strconv.Atoi(value)
			if parseErr != nil || pages <= 0 {
				return parsed, fmt.Errorf("invalid dump page count %q", value)
			}
			parsed.dumpMode = true
			parsed.dumpPages = pages
		} else {
			parsed.path = arg
		}
	}
	return parsed, nil
}

// renderDump renders a textual dump of the book's pages. maxPages limits the
// number of pages rendered; 0 means render all pages.
func renderDump(b *book.Book, maxPages int) string {
	var sb strings.Builder

	total := b.PageCount()
	if maxPages > 0 && maxPages < total {
		total = maxPages
	}

	for i := 0; i < total; i++ {
		fmt.Fprintf(&sb, "┌─── %s ── Page %d of %d ───┐\n", b.Title, i+1, b.PageCount())
		fmt.Fprintln(&sb, "│")
		page := b.Page(i)
		for _, line := range page.Lines {
			fmt.Fprintf(&sb, "│  %s\n", line)
		}
		// Pad to page height
		for j := len(page.Lines); j < b.PageHeight(); j++ {
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
