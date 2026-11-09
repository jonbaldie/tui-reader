# Reader interaction checks

`Reader.tla` models complete synchronous input updates. Each action maps to the corresponding navigation branch in `internal/tui/model.go`. `Resize` abstracts source-position capture and restoration as a map from a two-page layout into one page. It preserves a valid representative when expanding again. Selection cycling uses a one-link equivalence class.

`Styling.tla` models one iteration per regex match in `styleLinkMarkup`, after extraction and code-span classification. The source renderer's byte cursor becomes an ordered token cursor. Only attached matches consume page-level selection indices.

## Results

The baseline Reader check reached 34 states and 240 steps. TypeOK, SelectionValid, HistoryValid, and NoDeadlock held. Its reader-language view produced six chart states and fourteen arrows. The view hides Idle steps and merges steps within one visible label. It includes no bug marks.

The baseline Styling check reached 108 states and 108 steps. OccurrenceAccounting, NoDuplicateOccurrence, and NoDeadlock held. Its raw graph is a diagnostic graph rather than a primary user-flow chart.

All results are bounded model results, not proofs of the Go code. The skill lists the bounds and exclusions.

## Demonstrated model mutations

Use `/statechart` on scratch copies with these exact changes. Rename the module to match each scratch filename. Keep the original invariants.

- Set Reader.cfg's BadReset constant to TRUE. SelectionValid fails after select a link, then turn the page. Selection remains zero on a page without links.
- Replace Reader's `page < Last` guard with `page < 2`. TypeOK fails after go to the end, then turn the page. The page becomes two in a layout whose last page is one.
- Replace `Append(history,page)` with `Append(history,2)`. HistoryValid fails after select a link, then follow it. Back subsequently violates TypeOK.
- Replace Styling's `occurrence+1` with `occurrence+2`. OccurrenceAccounting fails on the first attached match. NoDuplicateOccurrence fails on the second attached match because the indices are zero and two.

The real-Go mutation demonstration is executable with `python3 verification/reader-interaction/check.py --mutate`. It removes the next-page selection reset through a temporary Go overlay. TestFocusClearsOnMove rejects that code with `page movement retained link focus`. It never edits the production file. The command reruns the unchanged source and requires it to pass.
