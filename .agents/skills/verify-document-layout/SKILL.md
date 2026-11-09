---
name: verify-document-layout
description: Verify document parsing, wrapping, layout cache replacement, and reading-position mapping after changes to internal/book/book.go.
---

# Verify document layout

A layout change must preserve text and source positions while replacing every geometry-dependent cache. A broken cache can move the reader to the wrong section after resize.

## Run

1. Read `verification/document-layout/README.md` for the code mapping, model bounds, and known exclusions.
2. From the repository root, run `python3 verification/document-layout/model_check.py && python3 verification/document-layout/check.py --mutation`. This checks ten bounded models and seventeen invariant-specific model mutations, then executes generated-input checks and a mutation/restoration check against an isolated copy of the current Go source. A failure is a code finding until its expected behavior is disproved from the source.
3. Use `/statechart` to check each TLA+ module and matching configuration in `verification/document-layout`. Work outside the repository. Use `view.mjs` with `BlockContext.tla`. Read every result through `/statechart`. A green Go command alone does not establish that the models pass.
4. To confirm that the behavioral checks detect a real source defect, run `python3 verification/document-layout/check.py --mutation`. It resets `firstParagraph` on a blank line in an isolated copy, requires the indentation assertion to fail, restores the copy, and requires the baseline to pass.
5. If a model fails, replay the report's path against Go with a focused test before reporting a production bug. If it cannot be replayed, correct the abstraction. Keep the original failing result as evidence.

The Go command and model checks together are the rerun procedure. `/statechart` owns its installation and invocation details. Python 3.11 or later and Go 1.25.0 are required. The recorded model run used TLA+ tools release 1.7.4 (whose checker reports TLC 2.19), installed by `/statechart`; the README pins its checksum.

## Coverage and synchronization

The last source correspondence review was at `dab9f7557cd033d89e80e732fab78d3c0934502e`.

Covered code lives in `internal/book/book.go`. `BlockContext` abstracts `formatMarkdownDocument`, `formatNextBlock`, and paragraph indentation context. `FenceContext` abstracts `findFenceBlockEnd` and heading suppression. `WidthAccumulator` abstracts `appendWordRunes`. `LayoutCache` abstracts `Read`, `layout`, and `Reflow`. `SourcePositions` abstracts `lineSources`. `ScannerDecisions` represents bounded persistent token, indentation-retry, provenance, column, digit-suffix, title-boundary and CRLF scans. `PageAnchors` retains per-page first-heading/content latches and cross-page carry with the page-zero exception. `ProsePromotion` collects prose until a boundary or adjacent underline. `RegistryFolds` checks first/last registration policy, and `PositionLookup` retains the greatest eligible candidate through a bounded lookup scan. Read the README for reductions and exclusions.

Generated checks additionally exercise line endings, title boundaries, indentation recognition, ordered-list recognition, setext promotion, indent fallback, source provenance, first-page and blank-page anchor policy, first-occurrence raw-page mapping, last-occurrence heading registration, lazy anchor mapping, page-copy isolation, position lookup bounds, and reflow equivalence to fresh construction. The README maps these to the scratch inventory.

The finite models check their abstractions, not Go execution. Generated Go inputs check the implementation within the stated corpus. Neither proves arbitrary Markdown correctness, unbounded input safety, or preservation of every exact reading offset across every possible reflow.

When a covered function changes, compare its control flow with its mapped model and checks. Update the model if intended behavior changes. Move the sync commit forward only after checking this correspondence and rerunning both Go checks and `/statechart`. Use `/maintain-formal-verification-skill` when later changes have not kept this coverage current.
