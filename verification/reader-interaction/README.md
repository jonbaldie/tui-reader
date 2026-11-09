# Reader interaction checks

`Reader.tla` models complete synchronous input updates. Each action maps to the corresponding navigation branch in `internal/tui/model.go`. `Resize` abstracts source-position capture and restoration as a map from a two-page layout into one page. It preserves a valid representative when expanding again. Selection cycling uses a one-link equivalence class.

`Styling.tla` models layout-provided label spans consumed by `styleLine`. It tracks the span cursor, retained byte end, rendered ranges and highlighted owners. It covers continuation owners, selected openers, overlapping ranges and spans on other lines.

## Results

At synchronization commit `9b158e9`, Reader reached 34 states and 240 steps. TypeOK, SelectionValid, HistoryValid and NoDeadlock held. Its reader-language view produced six chart states and fourteen arrows. The view hides Idle steps and merges steps within one visible label.

Styling reached 225 states and 225 steps. OrderedRanges, SelectedOwner, OpenerSelection, LineIsolation and NoDeadlock held. Its raw graph is a diagnostic graph rather than a primary user-flow chart.

All results are bounded model results. The skill lists the bounds and exclusions, including Reader's page-zero history reduction and Styling's assumed valid span bounds.

## Mutation demonstrations

`model_check.py` requires a named invariant violation for each of seven isolated mutations. Reader mutations fabricate page two, retain stale focus and append an invalid history origin. Styling mutations allow overlapping ranges, select continuation owner -1, render unselected openers and accept spans from another line. Use `/statechart` on scratch copies to read their shortest counterexamples.

`python3 verification/reader-interaction/check.py --mutate` also checks actual Go behavior. It removes the next-page selection reset through a temporary Go overlay and requires TestFocusClearsOnMove to reject the change. In a temporary module copy, it removes styleLine's nonnegative-selection guard and requires the continuation styling fixture to fail. Both mutations are followed by a restored-source passing run. Production files remain unchanged.

Six exact styling fixtures check continuation, opener selection, overlap, line isolation and adjacent owners. The command also runs the existing TUI tests and generated navigation sequences.
