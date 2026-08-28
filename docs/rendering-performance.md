# Rendering Performance

`BenchmarkModelViewLinkDense` renders a real `Model.View` with distinct
internal links on one line, including both selected and unselected links. It
forces ANSI-256 color so the benchmark includes the production terminal output
cost:

    go test -run '^$' -bench '^BenchmarkModelViewLinkDense$' -benchmem -count=3 ./internal/tui

On an Apple M3, median measurements were:

| Links | Before | After | Bytes before → after | Allocations before → after |
|---:|---:|---:|---:|---:|
| 16 | 181 µs | 191 µs | 123 KB → 114 KB | 1,198 → 1,242 |
| 64 | 613 µs | 577 µs | 813 KB → 470 KB | 4,086 → 4,231 |
| 256 | 3.29 ms | 2.12 ms | 8.60 MB → 2.06 MB | 15,588 → 16,111 |

The renderer now parses Markdown link occurrences once, left to right, and
styles attached labels as it encounters them. It avoids rescanning strings
that already contain ANSI output, giving `O(x + r)` lookup work plus final
output size for a line of `x` bytes and `r` link occurrences.
