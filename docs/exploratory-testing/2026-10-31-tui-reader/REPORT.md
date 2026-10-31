# Exploratory testing report: 2026-10-31

## Scope and setup

The pass exercised reading, internal links, and dump mode on Linux amd64. The build came from [`0fdbb58`](https://github.com/jonbaldie/tui-reader/commit/0fdbb582fa09d16abe4841b2af5f01b6636f1bfa), the fetched `origin/main` tip. Go 1.26.8 built the binary with `-mod=readonly` after `go mod download` and `go mod verify`. The full `go test -mod=readonly ./...` suite passed. Module and build caches, the binary, the home directory, and fixtures were isolated in a worktree.

The interactive runs used a private tmux server, a PTY at 80 columns by 24 rows, `TERM=screen-256color`, and the documented keys. Each run started a fresh reader process. Captures show terminal text, not the selection colors. The driver strips trailing spaces and blank lines from captures. The fixture files and [driver](evidence/drive.sh) are in `evidence`. The driver expects a built binary at `.run/tui-reader` and creates an isolated `.run/home`. Run it with `bash docs/exploratory-testing/2026-10-31-tui-reader/evidence/drive.sh`. It overwrites the fixtures and captures.

## Journeys

### Read and navigate

The goal was to read a multi-page document, move through pages, and keep a useful position after resize. In [`navigation.txt`](evidence/navigation.txt), Right and Left moved between pages 1 and 2, End showed the final Beta section, and Home returned to page 1. Resizing page 2 to 60 by 20 showed intro lines 06 through 11, including the former page's first line 08. Resizing back showed page 1, which included lines 06 and 07 but not line 08. This is a page-granularity limitation, not a new confirmed bug. [#61](https://github.com/jonbaldie/tui-reader/issues/61) explicitly excluded preserving the line offset within a page.

A separate [plain-text run](evidence/plain.txt) retained the two source lines, stayed on its single page with Right and Left, ignored link commands, and exited with status 0. All interactive runs exited with `q` and status 0.

### Follow links and return

The goal was to follow a contents link and return through link history. In [`links.txt`](evidence/links.txt), Tab and Enter moved from page 1 to Alpha on page 4. Alpha was the last line on that page, after seven lines from the preceding section. The link under Alpha appeared on page 5. After Right, Tab, and Enter, the reader showed Beta on page 8. Two presses of `b` returned to page 5 and then page 1. An initial run pressed Tab and Enter immediately after arriving at Alpha and stayed on page 4. That first capture was overwritten when the driver was revised to page forward. A [fresh replay](evidence/links-initial-replay.txt) produced the same result. The [common questions](../../common-questions.md) document that a link below a heading may be on the next page.

In a fresh process, Tab twice followed the second contents link to Beta. After `b`, Shift+Tab and Enter also followed Beta, and `b` returned to page 1. See [`links-variation.txt`](evidence/links-variation.txt).

A second fixture used a setext heading. Tab and Enter found the title on page 4, and `b` returned to page 1. The same capture showed separate quote lines and a tilde code fence, which were the subject of [#179](https://github.com/jonbaldie/tui-reader/issues/179), [#180](https://github.com/jonbaldie/tui-reader/issues/180), and [#178](https://github.com/jonbaldie/tui-reader/issues/178). See [`structures.txt`](evidence/structures.txt).

### Export a preview

The goal was to render a document without a terminal and limit the output when requested. [`dump.txt`](evidence/dump.txt) shows that `--dump=1` emitted only page 1 of 7, while `--dump` emitted all seven pages. The variation `--dump=0` returned usage and status 1. A missing file returned an error and status 1. The captures combine stdout and stderr.

## Findings and limits

No new bug met the confirmation and scope checks, so no issue was filed. The issue list and the bodies and comments of #61, #168, #178, #179, and #180 were checked before classification. The default `gh issue view --comments` command failed with a GitHub Projects classic GraphQL error, so the REST issues and comments endpoints were used instead.

**Unresolved observation.** The table in `structures.md` uses rows without a leading pipe. Its three lines appear as `A | B --|-- 1 | 2` in the [terminal capture](evidence/structures.txt), so their row boundaries are lost. The accepted brief for [#180](https://github.com/jonbaldie/tui-reader/issues/180) explicitly excludes table rows without a leading pipe. The documentation does not promise this table variant, so this pass does not classify the behavior as a confirmed defect. It merits a product scope decision before filing.

**Usability observation.** A destination heading can appear as the last line of its page, with the preceding section above it. A reader must page forward to see the next paragraph or link. The common questions already describe landing on the page that contains the heading.

This pass did not check link highlight colors, mouse input, other terminal emulators or operating systems, or a cross-file link. The captures are text-only. The terminal resize observation was not treated as a new defect because the recorded behavior falls within the page-level scope of #61. The run did not change product code.
