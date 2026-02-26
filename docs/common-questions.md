# Common Questions

## I pressed b and nothing happened

The b key goes back through your link history. If you
have not followed any links yet, there is nowhere to go
back to, so nothing happens. Navigate with arrow keys
instead.

## My link goes to the wrong heading

If your document has two headings with the same text,
like two separate "# Chapter" headings, links will
always jump to the last one. Anchors are derived from
heading text, so duplicates overwrite each other. Give
your headings unique names to avoid this.

## I followed a link but the next link is not on this page

When you follow a link to a heading, you land on the
page that contains that heading. If the next link in the
document is several lines below the heading, it may be
on the next page. Press the right arrow to find it. This
matches how Kindle works: you land at the chapter start,
not at a specific line.

## My PDF or EPUB will not open

The reader only supports plain text and UTF-8 encoded
files. PDFs, EPUBs, and other binary formats will show
an error. Convert them to plain text or markdown first.
Tools like pandoc can do this:

    pandoc book.epub -t markdown -o book.md

## A link is underlined but the highlight looks wrong

If a link appears in a very long line that gets word-
wrapped, the link text may be split across two displayed
lines. The link still works (press Tab to select it and
Enter to follow it) but the visual highlight may only
appear on part of the text. This is a cosmetic issue.

## The text looks too narrow or too wide

The content width is capped at 72 characters, which is
a standard comfortable reading width. It cannot be
changed. If your terminal is narrower than 72 characters
the text will adapt to fit. If your terminal is very
wide, the text stays centered with whitespace on either
side.

## Resizing the terminal lost my page position

When you resize, the document is re-paginated from
scratch. The reader tries to stay on the same page
number, but if the new layout has fewer pages, it clamps
to the last page. You will not lose your place by much.

## Nothing happens when I press Enter

Enter only works when a link is selected. Press Tab
first to highlight a link, then press Enter to follow
it. If the current page has no links, Tab and Enter are
both no-ops.
