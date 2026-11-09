# Document layout verification

Source correspondence: `dab9f7557cd033d89e80e732fab78d3c0934502e`, reviewed 2026-11-09. All source references below are in `internal/book/book.go`. Inventory numbers refer to `.audit/tui-reader-state-machines/scratch.md` in the parent workspace.

From the repository root, one command reruns all document models, targeted model mutations, generated Go checks, and the isolated Go mutation/restoration check:

```sh
python3 verification/document-layout/model_check.py && python3 verification/document-layout/check.py --mutation
```

`model_check.py` uses an already installed TLA+ tools release **1.7.4**, SHA-256 `936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88`, whose checker reports **TLC 2.19 (08 August 2024, rev 5a47802)**. It resolves that pinned artifact and Temurin Java 25 through `mise`; alternatively set `JAVA` to the Java executable and `TLA2TOOLS` to that exact jar. The recorded Go run used Go 1.25.0. Python requires 3.11 or later. Temporary checker state and mutated source live outside the repository.

Read every TLA+ result through `/statechart`, using matching same-name `.cfg` files and scratch copies outside the repository. `BlockContext` also has `view.mjs`. The root aggregate `python3 verification/check_all.py` includes this subsystem. `/statechart` owns its own setup and invocation details.

## Source mapping and finite abstractions

| Inventory | Source | Model state and bound | Checked property |
|---|---|---|---|
| 8 | `Read`, `Reflow`, `layout` (1070–1104) | `LayoutCache`: geometry, pages/map geometry, source cache and layout calls; three synchronous layouts, geometry IDs 1–2 | Pages and maps belong to the requested geometry; source cache is initialized. |
| 9 | `formatMarkdownDocument`, `formatNextBlock`, separator helpers (265–365,439–448) | `BlockContext`: first-block privilege, prior code/list category, blank tail; three blocks/blanks | A blank cannot restore the first-block privilege after an emitted block. Separator predicates are represented by context, but exact separator output is checked in Go. |
| 10 | fence helpers and `findFenceBlockEnd` (217–259) | `FenceContext`: no fence/backtick/tilde, scanned event count, heading-registration flag; three events | A heading inside a fence cannot register. Delimiter-length and indentation parsing are abstracted into matching/nonmatching events. |
| 12 | `findProseBlockEnd`, setext promotion (349–365,367–400,427–437) | `ProsePromotion`: input cursor, collected contiguous prose count, current line category, promotion and stop state; three consumed lines | Promotion requires collected prose and an adjacent underline; a boundary terminates collection. Line categories stand for dispatch results, not literal Markdown recognition. |
| 14 | `eachWrapToken`, `matchLinkToken` (999–1048) | `ScannerDecisions` token mode: cursor, pending whitespace, previous pending separator, atomic word/link emission; three events | Emitted separators require preceding whitespace; a recognized link advances as one token. Link recognition and inline-code exclusion are input classifications, not proven regex semantics. |
| 15 | `appendWordRunes` (961–982) | `WidthAccumulator`: current/emitted/consumed display columns; three runes, rune widths 1–2, page width 2 | Flushing never loses columns. Exact token-prefix and link-start metadata are exercised in Go and the link subsystem, not proved here. |
| 16 | `wrapParagraphWithIndent`, `formatCodeBlock`, `fitsWithIndent` (504–520,589–615) | `ScannerDecisions` indent mode: candidate phase, retained fit decision, accept or full-width retry phase | A candidate is accepted exactly when it fits; rejection selects full width. The fit predicate and resulting text are tested against Go. |
| 17 | `mapWrappedProvenance`, `advanceLine` (533–571) | `ScannerDecisions` provenance mode: retained search/raw cursors, prior cursors, success/failure event; three matches | Both cursors move forward; a failed search retains them. Matching and offset-to-raw conversion are reduced to atomic success events. |
| 18 | `lineSources` (780–789) | `SourcePositions`: current/prior raw index, prior nonspace count, offset; three lines, raw IDs −1/0/1 and counts 1–2 | A synthetic spacer or changed raw index resets continuation; only repeated nonnegative raw IDs accumulate. |
| 19 | `pagePositions`, `firstPageAnchor`, `pageAnchor` (759–778,801–832) | `PageAnchors`: page and line cursors, first-content/heading latches, last nonblank position, incoming carry, selected anchor; three pages of two lines | Page zero selects first content; later pages prefer first heading, then content, then incoming carry. Persistent scanning and carry are explicit. Source positions are scalar IDs. |
| 25 | `splitLines`, `titleCase`, `codeIndentLength`, `isOrderedListItem` (103–122,145–160,196–215,450–458) | `ScannerDecisions`: separate modes retain terminator/cursor/output-count state, previous whitespace, indentation column and stopped/rejected state, or consumed-digit/suffix state; three character events | CRLF emits one terminator, word boundaries title-case, indentation stops at four, ordered-list acceptance requires prior digits and a valid suffix. A suffix event includes `. ` or `) `; EOF rejects. Exact strings and EOF empty tail are Go checks. |
| 26 | `recordHeadingAnchor`, `rawLinePages` (303–323,739–757) | `RegistryFolds`: history, heading/raw-page maps, query and answer; three registrations over two keys | Heading registration keeps the last occurrence; raw pages keep the first; lookup returns that retained first value. Synthetic rows and normalization are reduced away. |
| 26 | `PageForPosition`, `PageForRawLine` (1179–1227) | `PositionLookup`: scan cursor, best candidate, eligible candidates, stop flag, query, query mode; fixed three-entry sorted example with a repeated raw ID and a gap, four queries | The retained best is the greatest eligible page/index; a greater raw ID terminates the position scan. Exact-match, fallback, clamping and lazy-map behavior are checked in Go. This finite example is not a proof over arbitrary maps. |

`ScannerDecisions` is a disjoint union: each initialized execution selects one scanner. Its registers are reused across modes; this avoids a large product of unrelated scanners. Bounds are part of the definitions, not hidden TLC constraints. Each module has a terminal self-loop and a matching configuration.

## Go correspondence checks

`check.py` copies current production Go files into a temporary module and adds `code_test.go.txt`. It never changes the repository's production files. The generated wrapper corpus enumerates strings of length 0–4 over `a`, `b`, `界`, space and tab, at widths 1–5. Checks compare nonspace text preservation and allowed display width. Layout checks compare reflow with fresh construction over five documents, Markdown/plain modes, four widths and two heights. Focused checks exercise line endings including trailing empty segments, title boundaries, tab/space indentation, ordered-list suffixes, fence heading suppression, multi-line setext promotion, paragraph/code indentation fallback, provenance, offset resets, first and blank page anchors, first/last registry policies, lazy maps, copy isolation and lookup bounds.

These are bounded generated checks against real code. They complement finite control-flow abstractions; they do not turn the abstractions into proofs of the Go parser or of all Markdown, Unicode, offsets, cache states or input sizes. File I/O, invalid UTF-8 errors, concurrency, allocation failure, full inline-code/link matching, delimiter lengths, and exact regex recognition are outside the formal models. Links have a separate verification subsystem.

## Sensitivity checks

`model_check.py` checks ten baseline models, prepares seventeen isolated mutants with `prepare_mutants.py`, and requires the **specific invariant name** to be violated. A parser error, crash or unrelated TLC error cannot satisfy that check.

| Mutant | Required violation |
|---|---|
| Reset first-block state on a blank | `LaterProseKeepsIndent` |
| Retain stale layout map geometry | `LayoutConsistent` |
| Drop an emitted full line's columns | `NoLostColumns` |
| Register a heading inside a fence | `CodeHeadingNotRegistered` |
| Carry a source offset across raw-line changes | `OffsetsResetAtSourceBoundary` |
| Discard a later page's heading/content priority | `AnchorPriority` |
| Promote an underline without collected prose | `AdjacentOnly` |
| Keep the first heading instead of last | `FirstLast` |
| Return a missing value from a populated raw map | `LookupRetainsFirst` |
| Refuse to update the best position candidate | `GreatestCandidate` |
| Emit unrequested token space; accept rejected indentation; advance failed provenance; accept too-short code indent; accept invalid digit suffix; invert title boundary; double-count CRLF | `ObservedRule`, independently for each of seven mutants |

`check.py --mutation` additionally inserts `firstParagraph = true` into the actual production blank-line branch **in the isolated copy**, requires the later-paragraph indentation assertion to fail, restores that source, and requires all checks to pass again. This is a real implementation mutation, not merely a model mutation. No defect is claimed in the unchanged production source.

Recorded baseline and mutation reports, shortest counterexample paths and charts are under the parent workspace's `.audit/tui-reader-formal/document-layout/reviewed-final/`. `report-BlockContext/chart.mmd` is the compact overview; its five states hide steps that keep the same displayed context. Detailed reports retain all model state. Other charts are larger mechanical scanner graphs. Refresh this record after any model or source change.
