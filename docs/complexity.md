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
| `ExtractLinks` | `O(x)` | `O(k)` returned | link parsing for one line |
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
| wrapping | `O(x)` per line | `O(x)` output | see wrapping benchmark |
| link attachment | `O(B + V + L)` | `O(L)` metadata | scans source and formatted lines once |
| `NewBook` | `O(B + V + L)` | `O(B + V + P + L)` returned | composes load, anchors, layout, and links |
| `Reflow` | `O(B + V + L)` | `O(V + P + L)` replaced layout | reuses raw lines and anchors |
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
