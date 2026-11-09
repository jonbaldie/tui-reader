# Exploratory testing report: 2026-11-08

## Setup

The pass used Linux amd64, Go 1.26.8, and `origin/main` at [`beb69b9`](https://github.com/jonbaldie/tui-reader/commit/beb69b987dd8e567fc487a4961085f64fdd70444). A worktree held the module cache, build cache, binary, home directory, and generated fixtures. `go mod download`, `go mod verify`, `go build -mod=readonly`, and `go test -mod=readonly ./...` succeeded. See [setup output](evidence/setup.txt).

The [driver](evidence/drive.sh) starts fresh reader processes in a private tmux server and captures their panes. The initial terminal was 80 by 24 with `TERM=screen-256color`. It generates the [Markdown](evidence/fixtures/tour.md) and [plain-text](evidence/fixtures/plain.txt) fixtures. The captures strip trailing spaces and do not preserve colors. To replay, run `mkdir -p .run && go build -mod=readonly -o .run/tui-reader .` followed by `bash docs/exploratory-testing/2026-11-08-tui-reader/evidence/drive.sh`. The driver replaces the fixtures and captures. Its wrapper prints the reader's exit code before the pane exits; the pane's own status is the wrapper's status.

The first driver attempt lost the pane when the reader quit, before it could capture the exit status. Retaining exited panes repaired the driver, and all journeys below ran again from fresh processes. The private tmux server was stopped after the run.

## Read and turn pages

The goal was to read a document and navigate without losing the text. In [navigation](evidence/navigation.txt), Right moved from page 1 to page 2, Left returned to page 1, End reached Beta on page 7, and Home returned to page 1. At page 2, resize to 60 by 20 retained paragraph 08 among the visible text. Resizing back to 80 by 24 showed page 1, with paragraph 07 as its last text. A [fresh replay](evidence/resize-replay.txt) produced the same round-trip drift. Closed [issue #61](https://github.com/jonbaldie/tui-reader/issues/61) excludes preserving the line offset within a page, but it is unclear whether that also covers losing the original page's first paragraph after two resizes. The observation remains unresolved rather than a confirmed new defect.

The [plain-text variation](evidence/plain.txt) retained its line breaks and stayed on its single page after Right and Left. Tab and Enter did not navigate its Markdown-like text. Both sessions exited with status 0.

## Follow links and return

The goal was to reach a heading and return to the source page. In [links](evidence/links.txt), Tab and Enter reached Alpha on page 4, and `b` returned to page 1. Two Tabs and Enter reached Beta on page 7, and `b` returned again. Shift+Tab and Enter left the reader on page 1. Selection of the last link, whose target is missing, is inferred from the documented key behavior, not observed in the text capture. A [fresh session](evidence/missing.txt) used three Tabs and Enter for the missing target, also remaining on page 1. This agrees with [controls](../../controls.md), which says missing anchors are silently ignored. These text captures cannot independently verify highlight colors.

## Preview and handle errors

The goal was to preview pages without a terminal. [Dump output](evidence/dump.txt) shows `--dump=1` emitted page 1 of 6, while `--dump` emitted all six pages. `--dump=0` and a missing input returned status 1. In a separate [interactive missing-file run](evidence/error.txt), the error screen identified the missing path and quitting returned status 1.

## Findings and limits

No new bug was confirmed or filed. The resize observation above remains unresolved against the scope of [#61](https://github.com/jonbaldie/tui-reader/issues/61). Existing open issues, closed issue titles, labels, worktrees, releases, and the comments on #61 were checked before classification. The repository had no `.out-of-scope` directory. This pass did not assess other terminal emulators, operating systems, link highlight colors, or multi-step link history after a resize. It did not change product code.

The resize round trip can return the reader to an earlier page that no longer contains the text they were reading. A reader may need to turn a page again.
