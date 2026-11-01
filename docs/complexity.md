# Complexity baseline

This inventory names the relevant input sizes so that a benchmark can compare
growth across sizes rather than impose machine-specific time thresholds.

| Symbol | Meaning |
| --- | --- |
| `B` | input bytes |
| `N` | raw input lines |
| `d` | path bytes used to derive a title |
| `x` | bytes in one line |
| `k` | links found in one line |
| `M`, `S` | candidate link ranges and inline-code spans |
| `V` | formatted display lines after wrapping |
| `P` | pages returned |
| `a` | raw lines when an anchor lookup must build its page map |
| `L` | links in a document |
| `C` | command-line arguments |
| `T` | rendered output bytes |

| Operation | Time | Additional or returned space | Notes |
| --- | --- | --- | --- |
| `Load` | `O(B)` | `O(B + N)` returned | reads, validates, normalizes, and splits input |
| title derivation | `O(d)` | `O(d)` | path and title transformations |
| link/code overlap classification | `O(M + S)` | `O(S)` span metadata | two advancing cursors over ordered ranges, shared by extraction and wrapping |
| `ExtractLinks` | regex and code-span discovery, plus `O(M + S)` | `O(M + S + k)` | link parsing for one line |
| pagination | `O(V)` | `O(V + P)` returned | formatting precedes page slicing |
| raw-line page map | expected `O(V)` | `O(N)` | one insertion per source line |
| argument parsing | `O(C)` | `O(1)` | last non-flag argument wins |
| dump rendering | `O(T)` | `O(T)` returned | output builder contents |
| cached `PageForAnchor` | expected `O(a)` lookup | `O(1)` | `a` is one hash-table lookup; expected case |
| uncached `PageForAnchor` | inherits layout construction | map cache | builds the raw-line page map once |
| ordinary navigation | `O(1)` | `O(1)` | page, link-selection, and back actions |
| follow link | expected `O(a)` | amortized `O(1)` | cached anchor lookup plus history append |
| terminal resize | inherits `Book.Reflow` | inherits `Book.Reflow` | recalculates layout |
| `NormalizeAnchor` | `O(t)` | `O(t)` output | one pass over heading bytes/runes |
| wrapping | candidate and code-span discovery plus output | `O(x)` output | overlap checks do not restart the span scan |
| link attachment | `O(B + V + L)` | `O(N + L)` metadata | indexes source link occurrences, then pairs wrapped link starts to them in order |
| `NewBook` | includes token discovery plus `O(B + V + L)` | `O(B + V + P + L)` returned | composes load, anchors, layout, and links |
| `Reflow` | includes token discovery plus `O(B + V + L)` | `O(V + P + L)` replaced layout | reuses raw lines and anchors |
| `View` | `O(T)` | `O(T)` returned | renders the current screen |

`a` is deliberately distinct from the number of anchor entries: the cached
lookup is expected constant time; the uncached path builds a page map from the
layout, so it has the same bound as layout construction.

## Scaling checks

Run the caller-visible suite and compare adjacent sizes for approximately
proportional growth. Allocation counts should likewise remain proportional;
do not assert wall-clock cutoffs in tests.

```sh
go test -run '^$' -bench '^(BenchmarkCallerVisibleOperations|BenchmarkExtractLinks|BenchmarkParseArgs|BenchmarkRenderDump)$' -benchmem ./...
```

The dedicated wrapping, link-dense reflow, rendering, and anchor benchmarks
cover their adversarial shapes alongside this suite.

`BenchmarkIssue193_CodeOverlap` compares mixed inline-code/link paragraphs with
link-only controls at 2,000, 4,000 and 8,000 repetitions, for extraction,
wrapping and reflow. Each cursor crosses each code span at most once. This
bounds overlap classification only. Unmatched backtick discovery is separate
work, so this is not a claim that all layout is linear.
