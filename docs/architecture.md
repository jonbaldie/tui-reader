# Architecture

## Project Structure

    main.go               Entry point
    internal/book/book.go Loading, wrapping, pagination
    internal/book/anchors.go  Headings and links
    internal/tui/model.go Bubble Tea model and view
    internal/tui/keys.go  Key binding definitions
    docs/                 This documentation
    sample.md             Sample book for testing

## How It Works

### Loading

The reader loads a file into memory as a slice of lines.
Line endings are normalized (CRLF and CR both become LF).
The file must be valid UTF-8 or an error is shown.

### Pagination

Lines are word-wrapped to fit the content width, then
split into pages of equal height. Long words that exceed
the width are hard-broken at the character level.

The content width is capped at 72 characters and centered
in the terminal. This keeps lines at a comfortable reading
length regardless of terminal size.

### Anchors

Markdown headings (lines starting with # through ######)
are extracted as named anchors. The heading text is
normalized to a URL-fragment style name:

    # Chapter 1: The Beginning
    becomes: chapter-1-the-beginning

### Links

Internal markdown links like [text](#anchor) are detected
on each page. The reader maps them back to the raw source
lines so that links still work even when the display text
has been word-wrapped across multiple lines.

### Reflow

When the terminal is resized, the entire document is
re-paginated on the fly. The reader tries to stay on the
same page, clamping if the new layout has fewer pages.

## Dependencies

    charmbracelet/bubbletea    TUI framework
    charmbracelet/lipgloss     Styling and layout
    charmbracelet/bubbles      Key binding helpers

## Design Decisions

No configuration. The reader picks sensible defaults for
content width, colors, and key bindings. There are no
config files, no flags beyond the file path, and no
settings menus. This is deliberate.
