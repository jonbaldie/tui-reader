# Link-attribution verification coverage

Sync revision is `dab9f7557cd033d89e80e732fab78d3c0934502e`.

## Model and bounds

`LinkAttribution.tla` represents `collectSourceLinks`, `attachLinks`, `recordFormattedLine`, and `assignLinksToPages`. It explores all three-event streams of opening a fence, closing a fence, and reading one link. Fence syntax is abstracted to an already-classified event. An unclosed fence stays active to EOF. Source cache construction precedes attachment. Each source occurrence has a distinct raw index. Duplicate markup is abstracted as occurrences in one queue, not collapsed into a set.

Extraction, cache construction, and occurrence consumption are internal algorithm steps. They are not concurrent UI events. The model describes their ordered execution. Actual Go executes the full attachment call synchronously.

The baseline reaches 36 distinct states and 43 transitions. The statechart groups them into four states and five arrows. Changes within the same chart label are hidden by the chart projection.

| Rule | Meaning | Sensitivity check |
|---|---|---|
| NoCodeLinks | No extracted source occurrence comes from inside a fence. | Permit collection while fenced. |
| ConsumedOnce | The attached count equals the consumed source count. | Leave the source cursor unchanged. |
| AttachmentOrder | Attached raw indices preserve source occurrence order. | Attach a fabricated raw index. |
| CacheBeforeAttach | Attachment starts only after source extraction is cached. | Skip setting the cache marker. |
| Complete | Completion retains all collected occurrences. | Finish before the source queue is exhausted. |

NoDeadlock is checked, with explicit stuttering only after completion. It is a checker diagnostic rather than an asserted parser guarantee.

## Correspondence to the scratch inventory

Paths below are relative to the repository root. Line numbers are at the sync revision.

| Scratch entries | Source | Evidence and abstraction |
|---|---|---|
| 8, cache sentinel | `internal/book/anchors.go:220-258`, `internal/book/book.go:1091-1103` | Model separates unknown from collected cache. Go checks assert nil triggers collection, non-nil empty suppresses it, plain text is computed-empty, and reflow retains source data. No concurrent invalidation is modeled. |
| 11, fence extraction | `internal/book/anchors.go:179-218` | Model checks fenced suppression. Go exhausts 512 three-line streams over eight tokens, including both fence characters, lengths three and four, indented delimiters, links, and blanks. Syntax beyond that alphabet is excluded. |
| 13, inline backtick spans | `internal/book/anchors.go:55-156` | `LinkScanners` Inline mode retains an opener length and closes only an equal run over four events with lengths one or two. It checks that every closed pair matches. This streaming abstraction is only a delimiter-pair contract, not the actual outer scanner. `InlinePairs.tla` models the actual outer cursor and exact-match lookahead over every four-run sequence with run lengths one or two. A missing closer advances only past the opener, permitting a later different-length pair. Its three invariants check exact pairs, ordered non-overlapping pairs, and a specific unmatched-opener recovery case. All three have demonstrated failing mutants. Go checks all 16 opener/closer length pairs from one through four, an unmatched opener, and code inside a link label. The occurrence model assumes classified active links; the scanner abstraction does not prove arbitrary byte-level scanning. |
| 20, raw attribution cursor | `internal/book/anchors.go:227-246` | `LinkScanners` SourceCursor mode retains ordered match positions and a source cursor over three source lines at offsets 0, 10, and 20, for up to three matches in positions 0 through 29. It checks that the cursor owns each latest match. Go checks a label crossing source lines and following matches attributed to their opening raw line. It checks source order and first-line registration. Arbitrary whitespace and Unicode are excluded. |
| 21, prose block grouping | `internal/book/anchors.go:336-352` | `LinkScanners` Grouping mode retains active block start and output groups over four prose, boundary, or setext events. It checks that a block never starts after its member. Go checks continuation, blank boundaries, and setext boundaries. Occurrence matching uses one block-key queue. Predicate recognition is supplied by the scanner model event classes rather than Go syntax. |
| 22, source occurrence queues | `internal/book/anchors.go:262-332`, `internal/book/anchors.go:354-360` | Model checks each source consumed once and in order. Go checks repeated identical markup on separate raw lines with distinct formatted positions. A real-source cursor mutation must fail. |
| 23, assignment queues | `internal/book/anchors.go:369-394` | Model checks ordered completion. Go checks total occurrence preservation and page-line bounds for one through four repeated paragraph links plus another paragraph link, widths one through 24, heights one through four, and reflows through width 25 and back. Missing-location and invalid-index fallback paths are not exhaustively modeled. |
| 25, NormalizeAnchor | `internal/book/anchors.go:24-42` | `LinkScanners` Anchor mode retains pending separator and emitted output over four letter, separator, or punctuation classes. It checks normalized separator placement. Go exhausts all strings of length at most four over seven runes, 2,801 strings total. It checks idempotence, absence of leading, trailing, or doubled separators, allowed rune classes, and selected exact results. This is bounded behavioral verification, not an unbounded proof. |

## Reproduce and maintain

Run `python3 verification/link-attribution/check.py` from the repository root. It records TLC output for the occurrence model and four scanner configurations, generated Go results, three real-source mutant failures, and model failures for all five protocol rules plus each scanner rule. No product mutation remains after the command because mutations occur only in temporary source copies.

Tested tools were Python 3.14, Go 1.26.3, OpenJDK 22.0.1, and TLC 1.7.4. The tool download is pinned by SHA-256. The script allows Java 11 or later and uses the installed Go compiler against the current module.

A failure on a changed source line can mean a real regression or an outdated model. Inspect both before changing an invariant. Use `/maintain-formal-verification-skill` to reconcile later changes.

## Scanner configurations

`LinkScanners.tla` uses `Anchor.cfg`, `Inline.cfg`, `Grouping.cfg`, and `SourceCursor.cfg`. `LinkScanners.cfg` is the default Anchor configuration for tools that require matching basenames. The runner checks all four modes separately. Scanner state is deliberately independent of the occurrence protocol, so no claim about their full cross-product is made.

The occurrence protocol collapses the two implementation matching passes into a single ordered consumption step. It proves the consumption contract at that abstraction, not missing-candidate fallback or arbitrary mixtures of block keys. Actual Go checks exercise both implementation passes with repeated identical markup and multiple prose blocks. The scanner rules each have a failing model mutation. That demonstrates sensitivity within the finite abstraction, not an automatic refinement proof from Go.

`InlinePairs.cfg` checks the outer inline scanner independently. The checked recovery example has run lengths 1, 2, 2, 2 and must emit the pair at positions 2 and 3. Go checks include the same unmatched-opener behavior.
