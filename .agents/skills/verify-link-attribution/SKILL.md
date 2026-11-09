---
name: verify-link-attribution
description: Check link extraction and occurrence attribution after changes to fences, inline code spans, anchor normalization, source grouping, link queues, or source-link caching in internal/book/anchors.go.
---

# Verify link attribution

Links inside code must stay literal, and every active occurrence must attach once in source order. A violation can make Enter follow a hidden link or select the wrong occurrence.

1. Read `verification/link-attribution/COVERAGE.md`. Compare its sync revision with changes to the named functions. Confirm that the model still describes the code before interpreting a pass.
2. From the repository root, run `python3 verification/link-attribution/check.py`. Success requires baseline TLC and Go checks, every expected mutant failure, and a restored passing Go check. The command prints the evidence folder. Set `VERIFY_OUT` to retain results at a chosen location.
3. Read TLA+ results with `/statechart`. Use the supplied `LinkAttribution.tla`, configuration, and `view.mjs`. Run it in a scratch folder outside the repository. Preserve its report and chart beside the checker evidence.
4. For a violation, replay the report against actual Go code. A mismatch is a modeling defect until source or runtime evidence establishes otherwise. Report existing defects without changing product behavior.
5. For intended changes, update the model, generated-input checks, source mapping, bounds, and sensitivity mutations together. Advance the sync revision only after confirming correspondence and rerunning the checks.

The runner uses TLC 1.7.4 with a checked SHA-256 download, Java 11 or later, Python 3, and the repository's Go toolchain. `JAVA` and `TLA2TOOLS` select existing tools. The generated Go harness runs against copies of current production files in a temporary module. Its mutants modify those real source copies, never the working tree.

A model pass covers its stated finite protocol or scanner scope. The generated Go checks cover their stated finite input families. Neither result proves arbitrary Markdown, Unicode input, all wrapping geometries, or the complete reader.
