# Link-attribution verification coverage

Sync revision is `9b158e9ce76bb27a71f41ec24481500a7e70625a`.

## Current pipeline and drift

The reviewed changes since `dab9f7557cd033d89e80e732fab78d3c0934502e` include all edits to `internal/book/anchors.go` and `internal/book/book.go`. Commit `c6edb5b` removes source-link caches, occurrence queues, and both matching passes. Wrapping now emits `placedLink` occurrences with opener coordinates and label pieces. `paginateFormatted` walks those placements once, adds navigation entries to the opener page, and calls `assignLabelSpans` for owner and continuation pages. Those retired mechanisms are no longer modeled as current behavior.

The first unchanged checker run on the integrated source failed to compile its Go harness because `collectSourceLinks`, `proseBlockStarts`, `buildLocations`, and `Book.sourceLinks` were removed. The old TLA+ models passed before that compile failure, demonstrating that a stale model pass alone is insufficient.

Commits `e1475f9` and `1a4175b` replace repeated regex searches and span rescans with persistent ordered cursors. Commit `4c5be4a` carries occurrence label ranges through hard wraps. Commits `1f6b43f`, `24b9cb1`, `c77c1ea`, and `bc420ee` change line blocks, normalized provenance, and shared heading classification. The document-layout verifier owns provenance and heading semantics. This verifier covers the link metadata that passes through those paths.

## Models and bounds

`LinkAttribution.tla` models two distinct occurrence IDs, including identical markup, with nondecreasing opener lines 0 through 2, page heights 1 or 2, and one nonempty label piece either on the opener line or the following line. Each placement step represents one synchronous iteration of `paginateFormatted` and the complete `assignLabelSpans` call. It checks the finished metadata of that call, without exposing its internal piece loop. It explores 144 distinct states. The chart projects those states to placements ready, first occurrence attached, and both attached. It hides only transitions that stay within a chart label.

| Rule | Observable contract | Failing model mutation |
|---|---|---|
| OccurrenceOrder | Every visited placement has one navigation entry in occurrence order, with a page-local index. | Fabricate the attached occurrence ID. |
| SpanOwnership | A label piece on another page has index -1; an owner-page piece uses its navigation index. | Give continuation pieces an owner index. |
| OwnerFallback | An opener marker exists exactly when no label piece fits on the owner page. | Omit opener markers. |
| CompleteSpans | Every visited placement retains its label piece and any necessary opener. | Drop the span update. |

The model trusts valid placement geometry from wrapping. It does not prove arbitrary numbers of pieces, byte coordinates, malformed negative coordinates, truncated pages, or Markdown recognition. `Expected` describes the atomic placement result and the independent invariant reference, so this is a bounded contract model rather than a refinement proof of the internal loop. The Go harness separately checks generated coordinates against actual display text.

`CodeCursor.tla` models the actual advancing first/last span cursors for two disjoint spans `[1,3)` and `[4,6)`. Candidate ranges have nondecreasing starts and ends, with endpoints 0 through 7. Its 29 states check `OverlapEquivalent`: cursor classification equals the full overlap predicate, including code wholly contained by a label. Replacing the edge-overlap OR with AND fails that rule. This models the cursor, not the lexical construction of candidate ranges.

`InlinePairs.tla` retains the actual outer cursor and first exact-match lookahead over all four-run sequences with run lengths 1 or 2. Its 52 states check exact pair lengths (`ExactPairs`), ordered disjoint pairs (`OrderedPairs`), and recovery after an unmatched opener (`EOFRecovery`). Respectively, pairing arbitrary run lengths, reversing a pair, and suppressing later candidates fail those rules.

`LinkScanners.tla` now contains only anchor normalization: four events chosen from letter, separator, and punctuation classes. Its 30 states check separator placement (`Normalized`). Emitting a separator without the following letter fails that rule. The obsolete streaming inline scanner, source cursor, and prose grouping modes and configurations are removed. Their former source mechanisms are retired or belong to document-layout verification.

Every named invariant has a separate failing mutant, checked alone so another invariant cannot hide the intended failure. NoDeadlock is a TLC diagnostic, with an explicit terminal stutter, rather than a parser guarantee.

## Actual Go checks

The runner copies current production `internal/book/*.go` files into a temporary module and installs the generated in-package harness there. It checks:

- Persistent `wrapLinkScanner.matchAt` against an independent regex-at-position and full-span reference at every byte of 225 two-fragment inputs. The alphabet includes nested and overlapping targets, empty/unclosed targets, repeated adjacent links, whitespace, and closed/unclosed backtick runs. This tests lexical queue construction in actual Go; no TLA+ lexical-scanner proof is claimed.
- `codeSpanCursor.overlaps` against `overlapsCodeSpan` for every pair of disjoint nonempty spans with endpoints 0 through 7 and every two nonempty candidate ranges with nondecreasing starts and ends through endpoint 8.
- Occurrence counts, valid UTF-8 label byte ranges, exact reconstructed labels, navigation line bounds, continuation indexes, and selectable owner labels/openers for one through three identical links; labels `x`, `long label`, `界é`, and inline code; ordinary prose, an indented second paragraph, quote, and table paths; widths 1 through 24 and heights 1 through 3. Each case reflows through width 25 and back.
- Distinct target order for three occurrences in a single paragraph, joined raw lines and a second paragraph, and a table row at those geometries.
- Literal fenced, unclosed-fenced, indented-code, and plain-text examples; all 16 opener/closer length pairs from 1 through 4; unmatched backtick recovery; and code contained by a link label.
- Anchor idempotence and normalized shape over all 2,801 strings of at most four runes from a seven-rune alphabet, plus selected exact outputs.

Five real-source mutations must fail behavioral tests: duplicating the navigation append, assigning continuation pieces an owner index, reversing the owner-fallback condition, changing the overlap OR to AND, and dropping scanner openers. Each runs against a fresh restored source copy; `finally` restores the edited copy, and the final unmutated Go run must pass. The production working tree is never mutated.

These are finite input families, not claims about arbitrary Markdown or Unicode. The placement model and scanner models are checked separately, not as a full cross-product. TUI selection styling is owned by reader-interaction verification.

## Reproduce and maintain

Run `python3 verification/link-attribution/check.py` from the repository root. `VERIFY_OUT` chooses the evidence directory. The four TLA+ files each have a matching basename configuration, so `verification/check_all.py` discovers them without relying on the deleted mode configurations. The runner uses Python 3, Java 11 or later, the repository Go toolchain, and checksum-pinned TLA+ tools release 1.7.4 (TLC reports 2.19). `JAVA` and `TLA2TOOLS` select existing tools.

Use `/statechart` for each baseline and each mutant in scratch directories outside the repository. Retain its report, chart, command output, and the exact mutated model beside the checker evidence. The landing verification trail is in workspace `.audit/tui-reader-landing/links/`, including `before.txt`, `source.diff`, `current/`, and `statecharts/`. No model violation was reproduced as a product defect during this update.

After later source changes, compare the named functions and commit messages with this mapping before advancing the sync revision. A failure can be a product regression or a stale abstraction; replay it against current source before changing a property.
