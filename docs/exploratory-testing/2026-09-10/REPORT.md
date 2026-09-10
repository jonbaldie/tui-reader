# Exploratory testing report — 2026-09-10

## Scope and build

I tested the latest `origin/main`, not the stale local checkout. The source
revision was `4639b9365a92d211e91dc73b69d96679ba7493af` and the clean test
build was `/tmp/tui-reader-exploratory-20260910` (SHA-256
`44ad7d958d2528d2821bfa688d2178dcec016d21f93e2d04d1bf7ecdcdbcbff8`).

The build passed `go test ./...` with Go `go1.26.3 darwin/arm64`. Live checks
used real PTYs and the supported keyboard interface; `--dump` was used where
it made text-layout differences unambiguous. Every bug below reproduced from
a clean process at least twice.

## Critical journeys

### 1. Read ordinary Markdown and reflow it

The normal reading path opened a UTF-8 Markdown fixture, rendered it, and
quit cleanly. A meaningful Markdown variation split one logical paragraph
across physical source lines and placed adjacent list items on consecutive
lines. The reader inserted blank display rows and paragraph indentation at
each source-line boundary. See
[`bug-soft-wrap.txt`](evidence/bug-soft-wrap.txt).

### 2. Follow an internal link and return

The ordinary link/history journey passed, including a resize from 80x24 to
40x12 after following the link: the destination remained visible and `b`
returned to the visible return marker. See
[`journey-links-resize.txt`](evidence/journey-links-resize.txt).

A variation placed link-shaped text inside inline code. `Tab`, `Enter`
navigated to the matching heading twice from clean processes, even though
the text was a literal example rather than a link. See
[`bug-inline-code-link.txt`](evidence/bug-inline-code-link.txt).

### 3. Use compact terminal dimensions and a long valid filename

At 18x12, the reader clipped both content and footer text at the terminal
edge. At 40x10, the title line disappeared. A valid long filename caused the
beginning of its four-line title to be clipped at 80x24. These were repeated
from clean processes. See [`bug-terminal-layout.txt`](evidence/bug-terminal-layout.txt).

## Confirmed bugs filed

1. **Markdown source-line boundaries create false paragraph breaks.** A
   standard soft-wrapped paragraph and adjacent list items receive extra
   blank rows and indentation, wasting reading space and changing pagination.
   [Issue #77](https://github.com/jonbaldie/tui-reader/issues/77).

2. **Internal links inside inline code are selectable and navigable.** A
   literal backtick example can unexpectedly move the reader to a matching
   heading when the user presses the normal link controls. [Issue #78](https://github.com/jonbaldie/tui-reader/issues/78).

3. **The TUI clips layout at compact dimensions and with long titles.** The
   minimum content width exceeds an 18-column terminal, the fixed layout drops
   the title at 40x10, and a wrapped filename title loses its beginning at
   80x24. [Issue #79](https://github.com/jonbaldie/tui-reader/issues/79).

## Passed, rejected, and unexplored

- Passed: opening, page movement, normal internal links, link history, resize
  position preservation on the latest revision, and clean terminal restoration
  after `q`.
- Passed: the previously reported indented-code navigation/spacing cases and
  strict malformed `--dump=N` handling on this revision; no duplicate issues
  were filed for the already-closed reports.
- Rejected: no additional bug from normal link navigation or resize was
  confirmed; the prior report's #61 behavior is fixed on this revision.
- Not covered: external URLs and cross-file links, mouse input, performance
  under very large files, non-Darwin terminals, and visual color/underline
  fidelity. Terminal sizes below the dimensions recorded above were not used
  as acceptance criteria.

## Cleanup

All live TUI processes exited with code 0 and restored the PTY. Temporary
fixture files are preserved with this report so the issue replays remain
available; no production source files were changed.
