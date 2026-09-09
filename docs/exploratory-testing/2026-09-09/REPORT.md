# Exploratory test report — 2026-09-09

## Scope and setup

- Budget: 30 minutes, AFK.
- Product revision: `fbb0d9fb23318a804f06903dcdbe2424a114c29d` (`v0.1.4`).
- Build: Go 1.26.3 on Darwin arm64; `go build -o /tmp/tui-reader-exploratory-20260909 .`.
- Binary SHA-256: `7dd5eefabf34aea6aceecb0c44515f5b21930d2bdd745ab70ed6494cf019b6f1`.
- Readiness check: `go test ./...` passed before exploration.
- Driver: the production TUI in a real interactive PTY, initially 80x24 or 90x24; terminal resize was delivered as a real `SIGWINCH` after changing the PTY dimensions.
- Data: isolated fixtures under [`evidence/fixtures`](evidence/fixtures).

## Journeys

### 1. Read and navigate a multi-page document

Goal: open a UTF-8 Markdown document, move forward, jump to the end, resize, and quit with the terminal restored.

Space advanced from page 1 to page 2, `G` reached page 5 of 5, and `q` restored the terminal and exited successfully. Resizing from 90x24 to 40x12 exposed a confirmed position-loss bug: content jumped from paragraphs 22–24 to paragraph 05. See [`journey-1-reading-navigation.txt`](evidence/journey-1-reading-navigation.txt).

### 2. Follow an internal link and return

Goal: select an internal Markdown link, reach its heading, and use `b` to return to the source position.

At a fixed 90x24 size, Space, Tab, Enter, and `b` completed the journey correctly. After resizing between follow and back, `b` returned to opening paragraph 02 rather than the visible RETURN MARKER. The clean replay failed identically. See [`bug-resize-position.txt`](evidence/bug-resize-position.txt).

### 3. Handle unreadable input and retry

Goal: understand missing-file and invalid-UTF-8 failures, quit safely, and successfully reopen readable input.

At 80x24 the complete missing-file reason was visible; the compiled binary used as input produced the documented `file is not valid UTF-8` error. Both states quit cleanly, and readable fixtures reopened normally. At 40x12, the missing-file error was clipped at `missing.` on two clean observations. See [`journey-3-error-handling.txt`](evidence/journey-3-error-handling.txt).

## Confirmed bugs

1. [#61 — Terminal resize loses reading and back-history positions](https://github.com/jonbaldie/tui-reader/issues/61). Reflow reuses page indices instead of preserving document positions, so resize can move the current reading location far backward and make link-history `b` return to the wrong content. Reproduced twice from clean processes, with a fixed-size control comparison.
2. [#62 — Error messages are clipped in narrow terminals](https://github.com/jonbaldie/tui-reader/issues/62). Error text is horizontally clipped at narrow but usable terminal widths, hiding the filename suffix and operating-system reason. Reproduced by resize and by starting a clean 40x12 PTY.

No matching open or closed report was found among GitHub issues 1–56 before filing.

## Rejected and unresolved candidates

- Rejected: link navigation/history itself. It completed correctly at a fixed terminal size; the failure requires reflow.
- Rejected: invalid UTF-8 detection. The documented error was displayed and the TUI remained quittable.
- Unresolved: selected-link styling was not evaluated visually because the PTY transcript records changed cells and ANSI controls rather than a color screenshot. Selection was functionally proven by Enter reaching the target.

## Coverage limits

External URLs and cross-file links are not supported product goals and were not exercised. Performance, very large files, mouse input, non-Darwin terminals, and terminals shorter than 12 rows remain unexplored. No product instrumentation or source changes were used.

## Cleanup

All TUI processes and PTYs created by the pass were closed. The temporary binary and disposable worktree are removed after this report is merged; only this report and its evidence are preserved.
