---
name: verify-reader-interaction
description: Check reader navigation, focus, link history, resize restoration, lifecycle, and layout-provided link-label styling when internal/tui changes.
---

# Verify reader interaction

The reader must keep page, selection, and return-history indices valid across input sequences. A stale index can select the wrong link, lose the return location, or crash rendering.

## Run

From the repository root, run `python3 verification/reader-interaction/check.py --mutate`. It runs source behavior checks, demonstrates a removed focus reset failing against actual Go through an overlay, and checks continuation styling against a second source mutation. Both mutations require a restored-source passing rerun. It needs Python 3 and the Go version in go.mod. It leaves production source unchanged.

Use `/statechart` to check `verification/reader-interaction/Reader.tla` with `Reader.cfg` and `view.mjs`, and `Styling.tla` with `Styling.cfg`. Copy models to scratch outside the repo before checking. Read every result with `/statechart`, including the shortest paths for broken rules. Report each invariant separately. A command succeeding does not establish that its rules hold.

The combined command `python3 verification/check_all.py` checks both baseline models and calls `model_check.py` to require an intentional violation of each reader and styling invariant. The standalone mutation script requires `JAVA` and `TLA2TOOLS` pointing to the tool versions selected by the combined runner.

## Coverage and synchronization

Confirmed against commit `9b158e9ce76bb27a71f41ec24481500a7e70625a`.

- `internal/tui/model.go` constructors, Update, handleKey, handleLinkKey, followLink, and View correspond to Reader's error, quitting, page, selected, and history state.
- `recalcLayout` corresponds to atomic Resize. Capture, Reflow, current restoration, history restoration, and selection clearing occur in one synchronous Update.
- `internal/tui/keys.go` defines the input aliases. The source check uses representative aliases.
- `styleLine` corresponds to Styling's span cursor, retained byte end, rendered ranges and selected owners. Spans on another line or before the retained end are skipped. An opener is included only for its selected occurrence. Continuation owner -1 is never selected. Layout supplies span ownership; the TUI does not reparse links.
- Scratch inventory entries 3 through 7 and 24 are covered here.

Before advancing the synchronization commit, inspect every changed covered function and update the model and source tests to describe its intended behavior. Do not advance it on a passing stale model alone. Use `/maintain-formal-verification-skill` after changes made without this check.

## Bounds and assumptions

Reader explores all action sequences over two abstract pages, two effective geometries, one selectable link on the start page, and history length at most two. A follow at that history bound is omitted. Shrinking maps all source positions to page zero. Expanding preserves their resulting representatives. This verifies index safety under that map, not semantic source preservation for every Markdown document. The real-code follow/resize/Back test separately checks source restoration on a multi-page fixture.

The single link always records a page-zero origin. Consequently, this model does not distinguish correct history remapping from retaining that zero. The generated fixture also starts at page zero. The existing `internal/tui` tests run by the command, including `TestResize_HistoryRemappedNotClamped`, check nonzero-origin remapping separately. The HistoryValid sensitivity mutant fabricates an out-of-range origin.

Reader permits input after quitting, matching Update's lack of a quitting guard. Error is fixed at construction. Nil book with nil error is excluded because normal main construction cannot produce it. Same-geometry resize, unknown input, unresolved link, no-selection Enter, and blocked movement are represented by stuttering. Reverse link cycling reduces to the same single-link Focus action. It is not a proof for all link counts.

Styling checks every two-span stream over five representative descriptors: first label, next label, continuation, opener and another-line span. Selected indices are -1, 0 and 1. Ranges use one-byte units at offsets zero or one. It checks ordered nonoverlapping rendered ranges, selected-owner validity, opener selection and line isolation. Valid span bounds are assumed from layout. Arbitrary UTF-8 ranges, heading styling and exact ANSI bytes are outside this model. Six exact Go styling fixtures and the existing TUI suite check output separately.

The Go sequence check covers all nine representative actions through depth three on one Markdown fixture, plus a dedicated follow/resize/Back test. Neither model is a formal proof of the Go implementation. No liveness/fairness claim is made.

## Interpret failures

For Reader, TypeOK means pages stay within layout bounds. SelectionValid means focus refers to a link on the current page. HistoryValid means every Back destination belongs to the current layout. For Styling, OrderedRanges means rendered ranges advance without overlap. SelectedOwner prevents the continuation sentinel from becoming selected. OpenerSelection restricts opener styling to its selected owner. LineIsolation excludes spans from other display lines.

Replay a failing model path through Update or the rendering tests. If the source does not reproduce the failure, examine the abstraction before declaring a product defect. Keep actual findings and their paths. Change rules only when the intended behavior changes.

## Maintenance result

Startup and navigation behavior were unchanged between dab9f75 and 9b158e9. Commits bc420ee and 4c5be4a intentionally moved heading classification into page metadata and link styling into layout-provided label spans. Styling was replaced to describe that design. The old occurrence-counter model could pass after its implementation was deleted, so it is no longer run. The source check now also removes styleLine's nonnegative-selection guard in a temporary copy, requires the continuation assertion to fail, restores the copy and requires a passing rerun. No product defect was found by these bounded checks.
