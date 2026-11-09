---
name: verify-reader-interaction
description: Check reader navigation, focus, link history, resize restoration, lifecycle, and link-label occurrence styling when internal/tui changes.
---

# Verify reader interaction

The reader must keep page, selection, and return-history indices valid across input sequences. A stale index can select the wrong link, lose the return location, or crash rendering.

## Run

From the repository root, run `python3 verification/reader-interaction/check.py --mutate`. It runs source behavior checks, demonstrates a removed focus reset failing against actual Go through an overlay, and reruns unchanged source. It needs Python 3 and the Go version in go.mod. It leaves production source unchanged.

Use `/statechart` to check `verification/reader-interaction/Reader.tla` with `Reader.cfg` and `view.mjs`, and `Styling.tla` with `Styling.cfg`. Copy models to scratch outside the repo before checking. Read every result with `/statechart`, including the shortest paths for broken rules. Report each invariant separately. A command succeeding does not establish that its rules hold.

The combined command `python3 verification/check_all.py` checks both baseline models and calls `model_check.py` to require an intentional violation of each reader and styling invariant. The standalone mutation script requires `JAVA` and `TLA2TOOLS` pointing to the tool versions selected by the combined runner.

## Coverage and synchronization

Confirmed against commit `dab9f7557cd033d89e80e732fab78d3c0934502e`.

- `internal/tui/model.go` constructors, Update, handleKey, handleLinkKey, followLink, and View correspond to Reader's error, quitting, page, selected, and history state.
- `recalcLayout` corresponds to atomic Resize. Capture, Reflow, current restoration, history restoration, and selection clearing occur in one synchronous Update.
- `internal/tui/keys.go` defines the input aliases. The source check uses representative aliases.
- `styleLinkMarkup` corresponds to Styling's ordered match cursor and attached-occurrence counter. Literal and inline-code matches copy without consuming an attached occurrence. Complete regex matches are abstracted as tokens.
- Scratch inventory entries 3 through 7 and 24 are covered here.

Before advancing the synchronization commit, inspect every changed covered function and update the model and source tests to describe its intended behavior. Do not advance it on a passing stale model alone. Use `/maintain-formal-verification-skill` after changes made without this check.

## Bounds and assumptions

Reader explores all action sequences over two abstract pages, two effective geometries, one selectable link on the start page, and history length at most two. A follow at that history bound is omitted. Shrinking maps all source positions to page zero. Expanding preserves their resulting representatives. This verifies index safety under that map, not semantic source preservation for every Markdown document. The real-code follow/resize/Back test separately checks source restoration on a multi-page fixture.

The single link always records a page-zero origin. Consequently, this model does not distinguish correct history remapping from retaining that zero. The generated fixture also starts at page zero. The existing `internal/tui` tests run by the command, including `TestResize_HistoryRemappedNotClamped`, check nonzero-origin remapping separately. The HistoryValid sensitivity mutant fabricates an out-of-range origin.

Reader permits input after quitting, matching Update's lack of a quitting guard. Error is fixed at construction. Nil book with nil error is excluded because normal main construction cannot produce it. Same-geometry resize, unknown input, unresolved link, no-selection Enter, and blocked movement are represented by stuttering. Reverse link cycling reduces to the same single-link Focus action. It is not a proof for all link counts.

Styling checks all three-match streams over attached, inline-code, and unattached classifications. It verifies occurrence counting and order, not regex correctness, ANSI bytes, split markup, or inline-code parsing. Existing internal/tui style tests run against actual code as part of the command.

The Go sequence check covers all nine representative actions through depth three on one Markdown fixture, plus a dedicated follow/resize/Back test. Neither model is a formal proof of the Go implementation. No liveness/fairness claim is made.

## Interpret failures

For Reader, TypeOK means pages stay within layout bounds. SelectionValid means focus refers to a link on the current page. HistoryValid means every Back destination belongs to the current layout. For Styling, OccurrenceAccounting means only attached links consume a selection index. NoDuplicateOccurrence means successive styled matches receive distinct indices in order.

Replay a failing model path through Update or the rendering tests. If the source does not reproduce the failure, examine the abstraction before declaring a product defect. Keep actual findings and their paths. Change rules only when the intended behavior changes.
