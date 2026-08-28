# Getting Started

## Install

Build from source. You need Go 1.21 or later.

    git clone https://github.com/jonbaldie/tui-reader.git
    cd tui-reader
    go build -o tui-reader .

This produces a single binary called tui-reader.

## Usage

Pass any readable file as the only argument:

    ./tui-reader mybook.md
    ./tui-reader notes.txt
    ./tui-reader ~/documents/essay.md

The reader opens in fullscreen (alt-screen mode) and
paginates the file to fit your terminal.

## Supported Formats

Any plain text or UTF-8 encoded file works. Markdown
files get the best experience because headings become
navigable anchors and internal links become clickable.

Other formats like .txt, .rst, .org, or .log files
will display fine as plain text with word wrapping.

## What Happens on Errors

If the file does not exist, you see a centered error
message. If the file is not valid UTF-8 (e.g. a binary
file), the reader tells you so. Press q to quit.

## Fuzz Testing

The bounded fuzz checks exercise generated documents through the production
book and TUI entry points. Run them one at a time to keep the test run
reproducible on developer machines and CI:

    GOMAXPROCS=2 go test -parallel=1 -run='^$' -fuzz=FuzzBookLayout -fuzztime=15s ./internal/book
    GOMAXPROCS=2 go test -parallel=1 -run='^$' -fuzz=FuzzModelActions -fuzztime=15s ./internal/tui
