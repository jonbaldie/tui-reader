# Differential Testing Scope

The comparisons use only behavior shared with an independent specification.

* Heading oracle: ATX level-one headings whose text contains ASCII letters,
  digits, and whitespace. The reference fragment is lowercased words joined by
  one hyphen.
* Navigation oracle: next, previous, first, and last page transitions over a
  fixed page count.

The oracle does not call production anchor or navigation code. Allowed
differences are the reader's paragraph indentation, line wrapping, page size,
styling, unsupported Markdown, link-following layout, punctuation handling,
and Unicode anchor normalization. Fuzz failures are automatically minimized by
Go and replayable with the printed `go test -run` command.
