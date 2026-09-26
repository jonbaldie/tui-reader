# Exploratory testing report — 2026-09-26

## Scope and build

I tested remote `main` at commit
[`9746026`](https://github.com/jonbaldie/tui-reader/commit/974602698c0dabf4acd1095a386ff5d70df3ff56)
with Go `go1.26.3 darwin/arm64`. `go build` succeeded; the binary SHA-256
and build metadata are in [`evidence/build.sha256`](evidence/build.sha256)
and [`evidence/build-info.txt`](evidence/build-info.txt).
Go's embedded module version includes `+dirty` because the report files were
untracked during the build; the checkout was at the commit above and no Go
source was modified.

I drove the TUI through tmux PTYs using the documented keyboard controls. The
normal session used `TERM=screen-256color` at 80×24, with an isolated `HOME`
and fixture data under `evidence/fixtures/`. The host had `NO_COLOR=1`, so
baseline captures contain no styling. I retained those observations and made
a separate diagnostic run with `NO_COLOR` unset; its ANSI capture confirms
that links are underlined and the selected link is inverted. The environment
and the capture index are recorded in [`evidence/`](evidence/README.md).

No product source was changed, and I did not run the automated test suite.

## User journeys

### 1. Open, read, and move through a plain-text book

I opened a 120-line text fixture. The first page showed marker 001. `Space`
advanced to page 2 with marker 018; `Right` advanced to page 3 with marker
035; `Home` returned to page 1; and `End` showed marker 120 on page 8. `q`
exited with status 0. The ordered screen captures are in
[`plain-navigation.txt`](evidence/plain-navigation.txt).

The documented preview path, `--dump=1`, exited 0 and emitted one page from
the same fixture. Malformed `--dump=invalid` printed usage and exited 1.
See [`dump-cli.txt`](evidence/dump-cli.txt).

As input-error variations, I opened a missing file and an invalid UTF-8 file.
Both showed the documented error and exited cleanly with `q`. At 40×12, the
missing-file error wrapped across lines and retained `no such file or
directory`. See [`input-errors.txt`](evidence/input-errors.txt).

### 2. Follow Markdown heading links and return

The fixture put Alpha and Beta on separate pages and included a link to a
missing anchor. `Tab`, `Tab`, `Shift+Tab`, `Enter` opened Alpha on page 3;
`b` returned to the link index on page 1. Two `Tab` presses and `Enter`
opened Beta on page 6; `b` returned to page 1. Three `Tab` presses selected
the missing anchor; `Enter` left the reader on page 1, as documented. The
process exited 0. The ordered screen captures are in
[`link-tour.txt`](evidence/link-tour.txt); selected-link styling is in
[`link-selection-style.ansi`](evidence/link-selection-style.ansi).

I also followed Alpha, resized from 80×24 to 40×12, then pressed `b`. The
reader returned to the link index after reflow. This replay is captured in
[`link-reflow.txt`](evidence/link-reflow.txt).

### 3. Resize while reading

At 80×24, `Right` twice showed marker 035 on page 3 of 8. Resizing to 48×14
kept marker 035 visible on page 5 of 18. Resizing back to 80×24 showed marker
018 on page 2 of 8. The same sequence reproduced from two clean reader
processes; `Home`, `End`, and `q` also worked. See
[`reader-reflow.txt`](evidence/reader-reflow.txt).

I rejected the page-marker movement as a new bug: closed issue
[#61](https://github.com/jonbaldie/tui-reader/issues/61) defines resize
preservation at page granularity and explicitly leaves the offset within a
page out of scope. After the narrower reflow, the next resize maps the
intermediate page's starting source line back to its containing wide page.
No issue was filed for this already-scoped behavior.

## Findings and issue tracker

No new confirmed bugs were found, so no GitHub issue was created. The error
wrapping case covered by closed issue
[#62](https://github.com/jonbaldie/tui-reader/issues/62) passed at 40×12.

## Limits

This pass covered plain text, Markdown heading links and history, missing
anchors, dump preview, file-open and UTF-8 errors, keyboard navigation, and
terminal reflow on macOS in tmux. It did not cover external or cross-file
links, mouse input, other operating systems or terminal emulators, very large
books, or a full unbounded `--dump` of every page. The normal host setting
`NO_COLOR=1` suppressed colors in baseline captures; selected-link styling
was checked only in the separate diagnostic run with that variable unset.
