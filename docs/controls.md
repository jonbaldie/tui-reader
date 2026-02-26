# Controls

## Page Navigation

    Right arrow, l, Space, PgDn    Next page
    Left arrow, h, PgUp            Previous page
    Home, g                        First page
    End, G                         Last page

## Link Navigation

    Tab                Select next link on current page
    Shift+Tab          Select previous link
    Enter              Follow the selected link
    b, Backspace       Go back to where you came from

Links are highlighted with an underline. The currently
selected link is shown with inverted colors so you can
see exactly which one you are about to follow.

## Quit

    q, Esc, Ctrl+C     Quit the reader

## How Links Work

In markdown files, internal links like [Chapter 1](#chapter-1)
are detected automatically. When you follow one, the reader
jumps to the page containing that heading.

Every time you follow a link, your previous position is
saved to a history stack. Press b to pop back through
your reading history, one step at a time.

Links to anchors that do not exist in the document are
silently ignored when you press Enter.
