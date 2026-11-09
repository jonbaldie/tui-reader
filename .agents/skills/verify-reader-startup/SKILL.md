---
name: verify-reader-startup
description: Check startup when changing argument parsing, dump selection, load-error handling, or exit status in main.go.
---

# Verify reader startup

A failed load must never report successful interactive completion. Otherwise scripts can mistake a failed read for success.

From the repository root, run `python3 verification/startup/verify.py`. Requires Python 3, Go 1.25, mise, TLC 1.7.4, and Java temurin-25. Install the pinned TLC artifact named in `verification/startup/model_check.py` with mise if missing. The script prints a baseline pass and each expected mutation failure. Unexpected checker errors fail the command.

Then use `/statechart` on `verification/startup/Startup.tla`, its adjacent config, and `view.mjs`. Read every result with `/statechart`. Run model checking in a scratch folder outside this repository. Its chart describes opening, reading, error display, dump output, and exiting.

## Coverage and bounds

Confirmed against `dab9f7557cd033d89e80e732fab78d3c0934502e`.

- `main.go` `run` and `runInteractive` correspond to process phases and exit-status rules in `Startup.tla`.
- `parseArgs` corresponds to the option fold. TLA explores zero through two valid option events, two positive counts, path present or absent, invalid count, and nondeterministic load and terminal-program outcomes.
- `args_test.go.txt` checks the real parser for every argument sequence of length zero through three over seven tokens. It distinguishes two paths, two valid counts, bare dump, invalid zero count, and an unknown flag treated as a path.
- No model proves filesystem behavior, terminal-library internals, numeric overflow parsing, every possible string, or unbounded argument lists. Lifecycle event outcomes are abstract inputs. There is no asynchronous loading phase in the code.

The model checks domains, load-error exit status, dump avoiding the reader, usage errors failing, and deadlock freedom. `model_check.py` breaks each named invariant in a scratch model and requires its failure. The real-code mutation changes `parsed.dumpPages = pages` to zero in a disposable copy. It must fail the parser check. The source repository is never mutated.

## Interpret and maintain

An unexpected failure remains a finding until replayed against Go. Compare each action to its function before deciding whether the code or abstraction is wrong. Do not weaken a rule to hide a failure. For intended behavior changes, update the model, its generated-input oracle, and mutation expectations together. Recheck source correspondence before advancing the sync commit. Use `/maintain-formal-verification-skill` when covered code changed without this check.
