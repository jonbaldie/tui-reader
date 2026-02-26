package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonbaldie/tui-reader/internal/book"
	"github.com/jonbaldie/tui-reader/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: tui-reader [--dump[=N]] <file>\n")
		os.Exit(1)
	}

	args := os.Args[1:]
	dumpMode := false
	dumpPages := 0 // 0 means all

	// Parse --dump flag
	var path string
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

	if path == "" {
		fmt.Fprintf(os.Stderr, "Usage: tui-reader [--dump[=N]] <file>\n")
		os.Exit(1)
	}

	if dumpMode {
		dump(path, dumpPages)
		return
	}

	model := tui.NewModel(path)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func dump(path string, maxPages int) {
	b, err := book.NewBook(path, 62, 20)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	total := len(b.Pages)
	if maxPages > 0 && maxPages < total {
		total = maxPages
	}

	for i := 0; i < total; i++ {
		fmt.Printf("┌─── %s ── Page %d of %d ───┐\n", b.Title, i+1, len(b.Pages))
		fmt.Println("│")
		for _, line := range b.Pages[i].Lines {
			fmt.Printf("│  %s\n", line)
		}
		// Pad to page height
		for j := len(b.Pages[i].Lines); j < 20; j++ {
			fmt.Println("│")
		}
		fmt.Println("│")
		fmt.Println("└────────────────────────────────┘")
		if i < total-1 {
			fmt.Println()
		}
	}
}
