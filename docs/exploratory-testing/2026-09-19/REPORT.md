# Exploratory testing report — 2026-09-19

## Scope and build

- Revision: `origin/main` at `41668ab` (fix(book): preserve links spanning physical lines in soft-wrapped paragraphs, #110).
- Binary: `go build -o /tmp/tr-et0919 .`, SHA-256 `94cccc169c349847f0f18eb9cda2b473f9a7eafb3fd996cac4d58d0c4bcac506`, Go `go1.26.3 darwin/arm64`.
- `go test ./...` passed before exploration.
- Driver: tmux 80x24, 60x14, 60x16 and 100x40 panes, `TERM=xterm-256color`, keys sent with
  `tmux send-keys`, screens captured with `tmux capture-pane` (`-e` when checking link selection
  highlight, background `48;5;117`). Scripts: [`evidence/driver/`](evidence/driver/).
- Fixtures were isolated under `/tmp/et0919/fx` and copied to [`evidence/fixtures/`](evidence/fixtures/).

Previous passes (2026-09-09/10/12) covered basic navigation, resize, compact layout, and
dump-mode argument errors, and their bugs are closed. This pass focused on realistic Markdown
authoring: README-style files, multi-chapter link history, non-English text, and files from other editors.

## Journeys

### 1. Read a README-style Markdown file with code

Goal: open a typical README (TOC, fenced code sample, prose) and read it with headings and links working.

- Ordinary path: [`fixtures/readme.md`](evidence/fixtures/readme.md) at 80x30. Prose, lists and
  headings render. The fenced ```` ```bash ```` block does not: `# Install dependencies first`
  and `# Configuration` inside the fence are styled as headings, and the code lines are reflowed into
  indented paragraphs with the fence markers attached (`widget init ````).
  See [`journey-plaintext-cli.txt`](evidence/journey-plaintext-cli.txt).
- Variation: a TOC link whose target name also appears as a comment in a later fenced block.
  The link jumps to the code comment, not the real section → **bug #111**.

### 2. Follow links through a multi-chapter book and return through history

Goal: jump between contents and chapters, then return with `b` to each earlier page.

- Ordinary path: [`fixtures/handbook.md`](evidence/fixtures/handbook.md) (5 chapters, contents and
  "Return to contents / on to next chapter" links). Contents → Ch3 → Contents → Ch5, then `b` ×3
  returned to 1 → 4 → 1 exactly. `b`/`Backspace` with empty history are no-ops. `l`, `Space`,
  `PgDn`, `h`, `PgUp`, `G`, `g` all moved as documented. See
  [`journey-links-history.txt`](evidence/journey-links-history.txt).
- Variation: resize 80x24 → 60x16 after following a link. The landed heading stayed visible
  (page 6/9 → 13/21), and `b` returned to the contents page. Resize 100x40 → 80x24 kept the heading on screen.
- Variation: headings in non-English text, and link fragments in GitHub style
  ([`fixtures/anchors.md`](evidence/fixtures/anchors.md)). Links to `## 第一章` and `## Été` are
  selectable but `Enter` does nothing → **bug #112**. `#install_deps` → `## install_deps` and
  `#Usage` → `## Usage` also do nothing. `docs/differential-testing.md` lists punctuation handling
  as an allowed difference, so these are recorded in #112 as related observations rather than as separate bugs.

### 3. Open files produced by other tools, and use the CLI

Goal: read files with Windows or exporter conventions, and get sensible CLI behaviour.

- CRLF line endings: headings, links and paragraph joining behave identically to LF. Passed.
- Plain-text `.log`/`.txt`: consecutive lines are merged into one reflowed paragraph, so
  log entries run together. The docs promise these formats "display fine as plain text" → **bug #114**.
- UTF-8 BOM: the first heading is not recognised (body style, missing anchor), so a "back to top"
  link is dead. The no-BOM control works → **bug #113**.
- CLI: `--dump=0` prints usage and exits 1. `--dump a.md b.md` silently uses the last path,
  as documented in `parseArgs`.

## Confirmed bugs

Each was replayed twice from fresh processes with minimal fixtures (60x14 for #111–#113, 80x10 for #114).

1. **#111: fenced code blocks are parsed as prose.**
   Impact: in a README, comment lines in a code block become headings and can take over TOC links. Code lines are joined into a paragraph.
   Replay: [`min-fence.md`](evidence/fixtures/min-fence.md), `Tab`, `Enter`.
   Expected: stay on page 1 (`## Usage`). Actual: page 2, showing the code comment `# usage` as a heading.
   Evidence: [`bug-fenced-code-heading.txt`](evidence/bug-fenced-code-heading.txt).
   [Issue #111](https://github.com/jonbaldie/tui-reader/issues/111).
2. **#112: headings with non-ASCII letters get truncated or empty anchors.**
   Impact: in CJK or accented-language books, links to chapters do nothing. An all-CJK heading cannot be reached by any link.
   Replay: [`min-cjk.md`](evidence/fixtures/min-cjk.md), `Tab`, `Enter`; control `Tab Tab Enter`.
   Expected: page 2 (`## 第一章`). Actual: stays on page 1. The control reaches page 4.
   Evidence: [`bug-non-ascii-anchor.txt`](evidence/bug-non-ascii-anchor.txt).
   [Issue #112](https://github.com/jonbaldie/tui-reader/issues/112).
3. **#113: a UTF-8 BOM stops the first heading being recognised.**
   Impact: files saved by Windows editors lose their title heading style, and links to the title do nothing.
   Replay: [`min-bom.md`](evidence/fixtures/min-bom.md), `End`, `Tab`, `Enter`.
   Expected: page 1. Actual: stays on page 3. `# Start` renders in body colour.
   Evidence: [`bug-bom-first-heading.txt`](evidence/bug-bom-first-heading.txt).
   [Issue #113](https://github.com/jonbaldie/tui-reader/issues/113).
4. **#114: plain-text files have consecutive lines merged.**
   Impact: `.log` entries and other line-oriented text run together. The Markdown soft-wrap joining from #77 is applied to every file type.
   Replay: `tui-reader app.log` ([`app.log`](evidence/fixtures/app.log)) or `--dump`.
   Expected: one entry per line. Actual: all three entries are reflowed into one paragraph.
   Evidence: [`bug-plaintext-lines-merged.txt`](evidence/bug-plaintext-lines-merged.txt).
   [Issue #114](https://github.com/jonbaldie/tui-reader/issues/114).

## Unresolved and rejected

- Rejected: link history after resize loses position. Observed correct in journey 2.
- Rejected: CRLF files break paragraph joining or anchors. They behave identically to LF.
- None unresolved.

## Usability observations

These are observations, not bugs. Suggestions are marked as such.

- A chapter heading can be the last line on a page, with its body on the next page (handbook,
  page 4 of 9 at 80x24). Following a link lands on that page, so the reader sees the end of the
  previous chapter first. The FAQ describes landing at the chapter start. *Suggestion:* keep a heading
  with at least one line of its following content.
- `tui-reader --help` opens the full-screen error `cannot open file: open --help`. It does not print usage.
- In TUI mode, a missing file shows the error screen, and quitting exits with status 0. `--dump` exits with 1 for the same file.
- Links show their raw Markdown (`[label](#anchor)`), consistent with earlier passes.

## Not covered

Mouse input, non-macOS terminals, terminal colour fidelity below 256 colours, very large (MB-scale) files,
and setext (`===`) headings.

## Cleanup

All tmux sessions were killed. The reader exited with status 0 on `q` where checked. Scratch files in
`/tmp/et0919` and the binary `/tmp/tr-et0919` were removed after evidence was copied here.
