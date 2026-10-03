# Exploratory testing report — 2026-10-03

## Scope and build

I tested remote `main` at commit
[`75f7f97`](https://github.com/jonbaldie/tui-reader/commit/75f7f97915f5b94621ab1f6bb46a0b365dd3fb7a)
with Go `go1.26.3 darwin/arm64`. `go build` succeeded; the binary SHA-256
and build metadata are recorded in [`evidence/build.sha256`](evidence/build.sha256)
and [`evidence/build-info.txt`](evidence/build-info.txt).
Go's embedded module version includes `+dirty` because report and evidence files were
untracked during the build; the checkout was at the commit above and no Go source was modified.

I drove the TUI through real tmux PTYs using the documented keyboard controls. The
standard sessions used `TERM=screen-256color` at 80×24, with an isolated `HOME`
and fixture data under `evidence/fixtures/`. The host environment had `NO_COLOR=1`,
so baseline captures contain no ANSI styling sequences. A separate diagnostic run
with `NO_COLOR` unset confirmed that active links are styled with inverted background
and bold text while inactive links are underlined. The environment, driver setup,
and capture index are recorded in [`evidence/README.md`](evidence/README.md).

No product source code was changed.

## User journeys

### 1. Plain-text and Markdown reading, navigation, and window reflow

I opened a 120-line plain-text fixture (`plain-reading.txt`). The first page
displayed line marker 001. All primary navigation keys were exercised:
`Space` advanced to Page 2 (marker 018); `Right` advanced to Page 3 (marker 035);
`l` advanced to Page 4 (marker 052); `PageDown` advanced to Page 5 (marker 069);
`Left` returned to Page 4 (marker 052); `h` returned to Page 3 (marker 035);
`PageUp` returned to Page 2 (marker 018); `Home` and `g` jumped directly to
Page 1 (marker 001); `End` and `G` jumped directly to the final Page 8 (marker 120).
Pressing `q` cleanly restored the terminal and exited with status 0.
The full ordered screen captures are preserved in
[`plain-navigation.txt`](evidence/plain-navigation.txt).

Plain-text files verified the fix for issue #114: individual source lines were not
collapsed or soft-wrapped into single paragraphs, and paragraph indents were not
injected.

I also tested reading position stability during window resize. From an initial
80×24 viewport, navigating twice forward showed marker 035 on Page 3 of 8.
Resizing the PTY to 48×14 kept marker 035 visible on Page 5 of 18. Resizing back
to 80×24 remapped the reading position back to marker 018 on Page 2 of 8,
conforming to the page-granularity mapping specified in closed issue #61.
The identical sequence was replayed across two clean reader processes,
confirming deterministic behavior in [`reader-reflow.txt`](evidence/reader-reflow.txt).

### 2. Internal link discovery, syntax variations, and history navigation

I tested internal links using `link-tour.md`, which contains links to multiple
headings, links in lists, links in prose paragraphs, bracketed link syntax,
links inside code blocks, and an invalid link targeting a non-existent anchor.

Cycling through links on Page 1:
- `Tab` cycled forward through Chapter Alpha, Chapter Beta, Missing Chapter,
  Bracketed Chapter, and two inline prose links (Alpha Again and Beta Again).
- `Shift+Tab` (`BTab`) cycled backward through the identical set of links in reverse order.
- Links inside inline code spans (`` `[Code Ignored](#chapter-alpha)` ``) and indented
  code blocks were correctly ignored as literal text and not offered for selection.
- Selecting Missing Chapter and pressing `Enter` left the reader on Page 1 without
  altering the page or pushing onto the history stack, as documented.
- Following Chapter Beta jumped to Page 3 of 4; pressing `b` returned to Page 1 of 4.
- Following Chapter Alpha jumped to Page 2 of 4; pressing `Backspace` returned to Page 1 of 4.
- Following Bracketed Chapter (`[[Bracketed Chapter](#chapter-bracketed)]`) jumped to
  Page 4 of 4; pressing `b` returned to Page 1 of 4.
- Quitting with `q` exited cleanly with status 0.
The full interactive captures are in [`link-tour.txt`](evidence/link-tour.txt).
ANSI sequences verifying inverted color highlighting for selected links are in
[`link-selection-style.ansi`](evidence/link-selection-style.ansi).

I tested link history preservation across window resize using `resize-reading.md`.
Starting at 80×24, navigating to Section Two on Page 2, and following `[Jump to Three]`
jumped to Section Three on Page 3 of 4. Resizing the terminal to 48×14 reflowed the
document to 11 pages (Section Three on Page 8). Pressing `b` returned to Section Two
(Page 5 of 11) where the link was originated, confirming history stack reflow.
Captured in [`link-reflow.txt`](evidence/link-reflow.txt).

Internationalization and character sets were verified with `unicode-tour.md`.
Headings with CJK characters (`# 目次`, `# 第1章: はじめに`) and accented Latin
(`# Chapitre: Éléphant`) generated valid anchors, were selectable via `Tab`,
navigated accurately on `Enter`, and returned cleanly with `b`.

### 3. CLI dump mode, edge cases, and input errors

The CLI dump mode was exercised across valid and invalid inputs:
- Single-page preview: `--dump=1` dumped Page 1 of `plain-reading.txt` and exited 0.
- Multi-page preview: `--dump=2` dumped the first two pages of `link-tour.md` and exited 0.
- Large page count: `--dump=99999` capped cleanly at the document's total page count and exited 0.
- Empty files and files with only whitespace/newlines paginated to a single clean page without crashing.
- Input validation: running without arguments, `--dump` without a file, `--dump=0`,
  `--dump=-1`, and `--dump=invalid` each printed the usage message to stderr and exited with code 1.
- Document loading errors: opening missing files, directories, and invalid UTF-8
  binary files emitted the appropriate error to stderr with exit code 1.
See [`dump-cli.txt`](evidence/dump-cli.txt).

In the interactive TUI, opening a missing file, a directory, or an invalid UTF-8
binary file displayed a centered error screen. Resizing the error screen to a compact
40×12 terminal wrapped the message cleanly without overflowing. Pressing `q`
exited the error view cleanly with status 0.
See [`input-errors.txt`](evidence/input-errors.txt).

## Findings and issue tracker

No new confirmed bugs were discovered during this exploratory pass.
All observed behaviors conformed to the documented interface and design requirements.
Previously closed issues (#61, #62, #91, #92, #114, #129, #135, #139) remained resolved.
Existing open issues (#148 and #150) regarding link attachment for identical markup
within prose paragraphs were verified as already tracked in the repository and were
not re-filed.

## Limits

This pass covered plain text files, Markdown headings, lists, code blocks, bracketed
links, Unicode / CJK characters, CLI dump mode, terminal reflow, keyboard controls,
and error handling on macOS (darwin/arm64) using tmux PTYs.
It did not cover unsupported external URLs or cross-file links (out of scope by design),
mouse input, or non-Darwin platforms.
