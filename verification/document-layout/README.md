# Document layout verification

Source correspondence: `9b158e9ce76bb27a71f41ec24481500a7e70625a`, reviewed 2026-11-09. Sources are `internal/book/book.go`; inventory numbers refer to the parent workspace's `.audit/tui-reader-state-machines/scratch.md`.

```sh
python3 verification/document-layout/model_check.py && python3 verification/document-layout/check.py --mutation
```

The root aggregate `python3 verification/check_all.py` includes these checks. Python 3.11 or later and Go 1.25.0 are required. `model_check.py` resolves Temurin Java 25 and TLA+ tools release **1.7.4** through mise, or uses `JAVA` and `TLA2TOOLS`. That jar has SHA-256 `936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88` and reports TLC 2.19. Temporary modules, checker state and mutations live outside the repository.

Read each model's result through `/statechart`, with its matching `.cfg`, in a scratch directory outside the repository. Use `view.mjs` with `BlockContext`. `/statechart` owns installation and invocation details.

## Source mapping and bounds

| Inventory | Source | Finite abstraction and checked property |
|---|---|---|
| 8 | `Read`, `layout`, `Reflow` | `LayoutCache`: three synchronous layouts, geometry IDs 1–2. Pages and derived maps must match the requested geometry. There is no persistent source-link cache in the current implementation. |
| 9 | `formatMarkdownDocument`, `formatNextBlock`, `formatLineBlock`, separator helpers | `BlockContext`: three blocks/blanks, first-block privilege, prior code/list category and blank tail. Blank lines cannot restore first-block privilege. Heading and line blocks share the same context transition; line-block grouping and separator output are Go checks. |
| 10 | fence helpers, `findFenceBlockEnd` | `FenceContext`: three events, no/backtick/tilde fence and heading-registration flag. Fenced headings cannot register. Delimiter matching is an input classification. |
| 12 | `findProseBlockEnd`, setext dispatch | `ProsePromotion`: three lines, collected prose, category, stop and promotion. Promotion requires collected prose followed immediately by an underline. Category recognition is checked in Go. |
| 14 | `eachWrapToken` | `ScannerDecisions` token mode: three classified token/space events, pending separator and cursor. Separators require preceding whitespace; a supplied link classification advances atomically. Persistent `wrapLinkScanner` recognition and inline-code exclusion belong to `verification/link-attribution`. |
| 15 | `appendWordRunes` | `WidthAccumulator`: three runes of width 1–2, page width 2. Flushing conserves display columns. Exact text and widths are Go checks. |
| 16 | `wrapParagraphWithIndent`, `formatCodeBlock`, `fitsWithIndent` | `ScannerDecisions` indent mode: candidate and retry phases with retained fit result. Acceptance requires a fit; rejection selects full width. |
| 17 | `provenanceCursor.nextContent`, `.line`, `mapWrappedProvenance`, `advanceLine` | `ScannerDecisions` provenance mode: four calls over three content runes with raw IDs 0,0,1. Each call consumes one content rune or retains EOF; raw IDs follow source order. Whitespace and UTF-8 bytes are projected away. Exact rune offsets, raw transitions, normalized whitespace and spanning output rows are checked in Go. This replaces the retired substring-search success/failure abstraction. |
| 18 | `lineSources` | `SourcePositions`: three rows, raw IDs −1/0/1, counts 1–2, optional authoritative source offsets 0–2. Explicit offsets win; otherwise only repeated nonnegative raw IDs accumulate and boundaries reset. The historical invariant name `OffsetsResetAtSourceBoundary` now checks this full decision. |
| 19 | `pagePositions`, `firstPageAnchor`, `pageAnchor`, heading classification | `PageAnchors`: three pages of two rows, first-heading/content latches and incoming carry. Page zero selects first content; later pages prefer the first classified heading, then content, then carry. A heading event means the cached `formattedLine.heading` flag, including wrapped continuations. Recognition is a Go check. |
| 25 | `splitLines`, `titleCase`, `codeIndentLength`, `isOrderedListItem` | Separate `ScannerDecisions` modes, each three character events. CRLF counts once, title casing follows whitespace boundaries, indentation reaches column four, list suffixes require preceding digits. Suffix events include `. ` or `) `; EOF rejects. |
| 26 | anchor registration in `formatMarkdownDocument`, `rawLinePages` | `RegistryFolds`: three registrations over raw IDs 0–1, inclusive start/end ranges, heading/raw maps and one lookup. Headings keep the last start occurrence; raw pages keep the first covering occurrence. One abstract registration equals one formatted row/page. Synthetic rows are excluded. |
| 26 | `PageForPosition`, `PageForRawLine` | `PositionLookup`: three sorted entries with a repeated raw ID and gap, four queries, cursor and best candidate. Selection retains the greatest eligible candidate. Exact-match, fallback and clamping are Go checks. Lazy rebuilding was deleted from production and is no longer claimed. |

`ScannerDecisions` chooses one independent scanner at initialization, avoiding a product of unrelated state. Bounds are encoded in the definitions. Each model has a terminal self-loop. These are bounded control-flow abstractions, not a proof of arbitrary Markdown, Unicode, delimiter recognition, allocation behavior or concurrency.

## Go correspondence

`check.py` copies current production files into a temporary Go module and installs `code_test.go.txt` in the same package. It uses the current `formatDocument` API, replacing deleted exported formatter helpers.

The wrapper corpus enumerates lengths 0–4 over `a`, `b`, `界`, space and tab at widths 1–5, checking nonspace conservation and permitted width. Reflow is compared with fresh `Read` construction over five documents, Markdown/plain modes, four widths and two heights. A second generated corpus uses all sixteen pairs of prose lines containing repeated spaces, tabs, Unicode and leading/trailing whitespace at widths 1–8; each output row must identify the exact first source rune and last raw source line it contains.

Focused checks cover fences, setext, ordered lists, indentation fallback, line endings, normalized provenance, authoritative offsets, inclusive raw ranges, heading continuation flags, first/blank page anchor policy, first/last registration, lookup bounds, preserved table and blockquote rows, and clone isolation for every Page slice (`Lines`, `Links`, `Headings`, `LabelSpans`). Link recognition and occurrence-specific placements have a separate verification subsystem.

## Sensitivity and evidence

Ten baseline models and nineteen targeted model mutations must pass the sensitivity checks. `model_check.py` requires the named invariant violation, so parser errors and unrelated checker failures cannot pass as successful mutations.

| Mutation | Required violation |
|---|---|
| Reset first-block privilege on blank | `LaterProseKeepsIndent` |
| Retain stale layout maps | `LayoutConsistent` |
| Drop emitted columns | `NoLostColumns` |
| Register fenced heading | `CodeHeadingNotRegistered` |
| Carry offset across source boundary; ignore explicit source offset | `OffsetsResetAtSourceBoundary`, separately |
| Discard later-page priority | `AnchorPriority` |
| Promote without adjacent prose | `AdjacentOnly` |
| Keep first heading; omit range endpoint | `FirstLast`, separately |
| Return missing raw mapping | `LookupRetainsFirst` |
| Refuse best-candidate update | `GreatestCandidate` |
| Emit unrequested space; accept rejected indentation; advance beyond provenance EOF; accept short indent; accept invalid list suffix; invert title boundary; double-count CRLF | `ObservedRule`, seven separate mutations |

`check.py --mutation` makes five separate changes to isolated production source: reset first-block privilege, ignore authoritative offsets, retain offsets across raw boundaries, omit inclusive range endpoints, and remove heading continuation flags. Each must fail its specific behavioral assertion, then restoration must pass before the next mutation. No production mutation is retained.

Landing evidence is in the parent workspace's `.audit/tui-reader-landing/layout/`: `before.txt` records the unchanged harness's compilation drift, `book.diff` the reviewed source changes, `after.txt` the current Go/mutation results, `models.txt` all baseline and targeted mutant checks, and `statechart/*/report/` each model's report, chart and TLC output. `mutant-statechart/*/report/` preserves each targeted counterexample. The previous reviewed evidence remains under `.audit/tui-reader-formal/document-layout/reviewed-final/`.

The maintenance review inspected every `book.go` change and commit message since `dab9f7557cd033d89e80e732fab78d3c0934502e`. It replaced stale cache/lazy-map claims, updated heading and line-block correspondence, and added normalized source offsets and inclusive registration ranges. Existing event bounds remain, except provenance now takes four calls to include an EOF call after three content runes. Source-offset choices and inclusive ranges broaden their existing three-row bounds. No production bug was reproduced by the refreshed checks.


Changes to the promised behavior were intentional: `c93e41c` deleted the source construction helpers and lazy anchor rebuilding; `bc420ee` and `c77c1ea` made cached heading classification authoritative; `1f6b43f` introduced preserved table/quote line blocks; `24b9cb1` replaced substring matching with normalized content-rune provenance and authoritative offsets; `c6edb5b` removed the source-link cache in favor of layout placements. Their source, tests and commit messages support the updated contracts. `4c5be4a` adds page metadata cloned by `Page`. `e1475f9` and `1a4175b` change recognition inside the link subsystem, whose scanner checks cover that boundary. The old “failed substring search retains cursors” and “source cache is initialized” statements described deleted mechanisms and have been replaced by “content calls consume one source rune or retain EOF” and “all derived geometry belongs to the requested layout.”
