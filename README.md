# tui-reader

A beautiful, opinionated TUI e-reader for the terminal. Opens any text or markdown file and paginates it like a Kindle — centered, word-wrapped, with internal link navigation. No configuration. No settings. Just reading.

## Installation

```bash
go install github.com/jonbaldie/tui-reader@latest
```

*(Or build from source locally with `go build -o tui-reader .`)*

## Quickstart

Open any markdown or text file:

```bash
tui-reader path/to/book.md
```

- **Turn pages**: `Space` or `→` (forward), `←` (back)
- **Follow links**: `Tab` to select, `Enter` to jump, `b` to return
- **Quit**: `q` or `Esc`

## Controls

| Key | Action |
|---|---|
| `→` `l` `Space` | Next page |
| `←` `h` | Previous page |
| `Home` `g` | First page |
| `End` `G` | Last page |
| `Tab` | Select next link |
| `Shift+Tab` | Select previous link |
| `Enter` | Follow selected link |
| `b` `Backspace` | Go back |
| `q` `Esc` | Quit |

## Dump mode

Render pages to stdout without a terminal (useful for scripts and previews):

```
./tui-reader --dump book.md
./tui-reader --dump=3 book.md   # first 3 pages only
```

## Docs

See `docs/` for architecture, controls reference, and common questions.
