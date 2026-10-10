# Exploratory testing report: 2026-10-10

## Scope and build

I tested remote `main` at commit
[`7afda9d`](https://github.com/jonbaldie/tui-reader/commit/7afda9dbf311c96a88141112fa443ae39ac7cc52)
(after v0.1.13 and the #176 refactor), built with `go build` on Go 1.26.3
darwin/arm64. Modules were fetched with `go mod download` and checked with
`go mod verify`. Environment details and the binary hash are in
[`evidence/environment.txt`](evidence/environment.txt).

I drove the TUI through tmux PTYs with the documented keys. The settings were
`TERM=screen-256color`, `NO_COLOR` unset, and an isolated `HOME`. The driver is
[`evidence/drive.sh`](evidence/drive.sh). Usage is
`drive.sh <session> <cols> <rows> <file> K:<key>... R:<cols>x<rows>...`, and you
set `READER` to point at a binary. Fixtures are in
[`evidence/fixtures/`](evidence/fixtures/). No product source was changed.

Earlier passes (2026-09-09 to 09-26) covered basic paging, link history,
missing anchors, error screens and resize. So this pass looked at Markdown
variants that real books use, a large book, and unusual inputs.

## Journeys

### 1. Follow a table of contents in a real-world Markdown book

**Goal:** a contents link jumps to the page with its heading, and `b` returns.

- ATX headings with a closing sequence (`## Closed heading ##`) and headings
  indented up to 3 spaces both navigate correctly: Page 1 → 4 and Page 1 → 8,
  and `b` returns. See [`journey-heading-variants.txt`](evidence/journey-heading-variants.txt).
- **Confirmed bug:** a `# install` comment inside a `~~~` fence takes over the
  `#install` link. The reader jumps to page 5 instead of staying on page 1, and
  the fence renders as prose. The backtick control works. See
  [#178](https://github.com/jonbaldie/tui-reader/issues/178).
- **Confirmed bug:** setext headings (`Title` / `===`) get no anchor, so links
  to them do nothing. See [#179](https://github.com/jonbaldie/tui-reader/issues/179).
- A CRLF file renders cleanly. Carriage returns are stripped and the headings
  are recognised. See [`dump-constructs.txt`](evidence/dump-constructs.txt).

### 2. Read a large book

**Goal:** a 3.5 MB book with 200 chapters and 8,200 links opens and stays
responsive through linking, paging, resizing and history.

- `--dump` of all 3,709 pages took 0.16 s with a 49 MB peak RSS.
- In the TUI at 100×30 (2,793 pages), I pressed `Tab`×3 and `Enter`, which
  opened Chapter 3 on page 37. `Right` went to 38. Resizing to 60×20 gave page
  78 of 6,116. `b` returned to page 1, and `End` went to 6,116. Resizing back
  to 100×30 gave page 2,792 of 2,793. `Home` went to page 1 and `q` exited 0.
  Every step completed within the 0.5 s capture interval. See
  [`journey-large-book.txt`](evidence/journey-large-book.txt) and
  [`gen-novel.py`](evidence/gen-novel.py) to regenerate the fixture.
- The landing on 2,792 of 2,793 rather than the last page after the
  resize-back fits the page-granular resize scope already accepted in #61 (see
  the 2026-09-26 report). I am classifying it as **rejected** (it is in scope).

### 3. Read Markdown with tables, quotes and unusual inputs

**Goal:** the content stays readable in its source structure.

- **Confirmed bug:** table rows and blockquote lines are merged into one
  reflowed paragraph, which makes tables unreadable. See
  [#180](https://github.com/jonbaldie/tui-reader/issues/180).
- Lists, nested lists, ordered items, `---` rules and `#nothing` lines (no
  space after the `#`) keep their own lines. See [`dump-constructs.txt`](evidence/dump-constructs.txt).
- With an empty file, the TUI shows Page 1 of 1. `Right`, `End`, `Tab` and
  `Enter` do nothing, and `q` exits 0. Passing a directory shows a wrapped
  `is a directory` error and the process exits 1, in both TUI and `--dump`
  modes. See [`input-variations.txt`](evidence/input-variations.txt).

## Findings

| Issue | Summary | Repeats |
|---|---|---|
| [#178](https://github.com/jonbaldie/tui-reader/issues/178) | `~~~` fences not recognised; comments inside become headings that hijack links | 2/2 fresh processes, plus a backtick control |
| [#179](https://github.com/jonbaldie/tui-reader/issues/179) | Setext headings have no anchor; links to them do nothing | 2/2 |
| [#180](https://github.com/jonbaldie/tui-reader/issues/180) | Tables and blockquotes are reflowed into one paragraph | 2/2 |

Each issue lists the starting conditions, replay steps, expected and actual
results, and links to the evidence.

## Unresolved and observations

- **Anchor slugs differ from GitHub's.** `# snake_case_names` gives
  `snakecasenames` and `# Q & A` gives `q-a`. GitHub gives `snake_case_names`
  and `q--a`, so a contents list copied from GitHub-rendered Markdown would
  have dead links. The docs promise only the `chapter-1-introduction` style,
  not GitHub compatibility, so I am leaving this **unresolved** and have not
  filed it. Fixture: [`anchors.md`](evidence/fixtures/anchors.md).
- **Usability observation:** the reader shows Markdown source as written, with
  `#` markers, `[label](#target)` and `>` all visible. A link's target is
  always on screen, which makes dead links easy to spot, but tables and quotes
  depend on line structure, and that structure is lost in the reflow.

## Limits

This pass did not cover the following: mouse input; terminals other than tmux;
other operating systems; external or cross-file links; and how a `---` under a
paragraph should be handled when it could be a setext underline or a rule.
Baseline captures are plain text (`capture-pane -p`), so link highlighting was
not checked visually.

An untracked `docs/exploratory-testing/2026-09-12/` directory in the main
checkout, and the unmerged `.worktrees/exploratory-2026-09-19` and
`exploratory-2026-10-03` branches, come from earlier runs. I left them
untouched.
