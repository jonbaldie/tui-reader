# Novel opening and terminal resize

## Scope

Apply `hill-climbing` to `tui-reader` at baseline
`303b194` (main, 2026-09-24). Candidate inspection covered BookBeam CLI,
Foldback and tui-reader: local document layout offers a repeatable user journey
without remote API or disk-backup variability. No usage telemetry was available;
selection was based on the reader's core workflow.

Target: reduce resize latency by at least 10% on a novel-sized document,
preserve output, and protect the reduction in work. Limit this pass to one
profiled bottleneck and two measurement rounds, then finish review and CI.

## Change

Paragraph wrapping previously collected every word in a growing token slice,
then iterated the slice. It now calls the wrapping consumer as each token is
scanned. Whitespace-only input still returns before allocating wrapping buffers.
Token recognition, word widths, line breaks and link handling are unchanged.

The allocation profile attributed 55.5% of startup allocation bytes to token
slices on this fixture (`allocations-before.txt`). Removing them eliminates
seven allocations per ordinary 40-word paragraph.

## Workload and measurement

`internal/tui/novel_benchmark_test.go` builds a synthetic manuscript with
120,000 prose words, 30 chapters, 3,000 paragraphs, chapter links and accented
Unicode. It produces 658 pages at the measured sizes.

- Open: `NewModel` (including file read and default pagination), initial 80x30
  window event (including reflow), then `View`.
- Resize: alternate 70x30 and 80x30 window events, then `View`.
- Timings end when the rendered string is ready. They exclude process launch,
  terminal painting and human input. File reads are warm-cache reads.
- Apple M3, macOS 26.6.2, Go 1.26.3, default GC settings.
- Five alternating before/after pairs, reversing execution order every pair.
  Final round: 60 operations per journey per run, `GOMAXPROCS=1`.
- Each run reports its nearest-rank p75. The table reports the median of the
  five run-level p75 values, not a pooled percentile.

| Measurement | Before | After | Reduction |
|---|---:|---:|---:|
| Open to first view, median p75 | 55.30 ms | 53.66 ms | 3.0% |
| Resize to view, median p75 | 28.83 ms | 25.44 ms | 11.8% |
| Open, allocated bytes/op | 37,379,153 | 16,774,791 | 55.1% |
| Resize, allocated bytes/op | 17,843,187 | 7,521,101 | 57.8% |
| Open, allocations/op | 109,385 | 67,133 | 38.6% |
| Resize, allocations/op | 53,221 | 32,094 | 39.7% |

The resize p75 improved in all five final pairs (11.8–20.5% per pair).
The initial two-worker round also showed a lower median resize p75
(30.56 to 27.16 ms), but inconclusive startup timing (53.46 to 54.46 ms).
Those raw results remain in `before.txt` and `after.txt`; the final round is in
`single-worker-before.txt` and `single-worker-after.txt`. The initial prototype
handled empty input after allocating buffers; the final version preserves the
early whitespace return. This does not change the manuscript's output.

Absolute timings varied substantially on this shared machine. Do not claim a
reliable startup speedup or extrapolate these timings to other computers. The
allocated-byte reduction is consistent, and the repeated resize improvement
supports using allocations as a regression signal here.

## Verification and ratchet

- Full Go tests pass before and after; race tests and `go vet` pass.
- `messgo . github unusedcode,design,codesize --ignore-tests` passes.
- `TestNovelJourneyOutput` checks baseline SHA-256 hashes of every page, its
  links, source position and rendered view, across 80→70→80 column resizes.
  Existing tests cover Unicode, narrow widths, code blocks and navigation.
- `TestWrapParagraphAllocationBudget` caps a 40-word paragraph at 12
  allocations: baseline fails at 16; optimized code passes at 9. The ordinary
  test suite runs this guard, including CI's mutation-test baseline check.
- Changed-line mutation testing killed all 11 mutants (none escaped or errored).
- Timing remains a benchmark, not a flaky CI gate.

Reproduce the benchmark on each version with the same test file:

```sh
GOMAXPROCS=1 go test ./internal/tui -run '^$' \
  -bench BenchmarkNovelJourney -benchtime=60x -count=5
GOMAXPROCS=2 go test ./internal/book -run TestWrapParagraphAllocationBudget -v
GOMAXPROCS=2 go test -race ./...
GOMAXPROCS=2 go vet ./...
```

For paired measurements, compile the test package on each revision with
`go test -c`, then alternate the two executables with `GOMAXPROCS=1`,
`-test.run=^$ -test.bench=BenchmarkNovelJourney -test.benchtime=60x`.
Copy the benchmark test onto the baseline checkout before compiling it.
Allocation profiling uses the same benchmark with `-memprofile`, then
`go tool pprof -top -alloc_space`.

## Stop and pending field check

The resize target was met with a small change and identical output. Stop after
this bottleneck; further changes would need a fresh profile and separate
justification. These are lab-only results. No release or installation is part
of this pass. Check opening and resizing real manuscripts after the next release,
including actual terminal paint time and other CPU/Go configurations.
