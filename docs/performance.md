# Performance Notes

## Link attachment during reflow

`BenchmarkBookReflowLinkDense` uses a generated document with two internal
links per source line, each wrapped at 24 columns. It exercises the same
`Book.Reflow` path used by terminal resize.

Run it with:

    go test -run '^$' -bench '^BenchmarkBookReflowLinkDense$' -benchmem -count=3 ./internal/book

On an Apple M3 before the single-pass attachment change, the median results
were:

| Source lines | Time | Allocations | Bytes |
|---:|---:|---:|---:|
| 32 | 159 µs | 1,630 | 161 KB |
| 128 | 635 µs | 6,448 | 638 KB |
| 512 | 2.61 ms | 25,698 | 2.54 MB |

After the change, the same workload measured:

| Source lines | Time | Allocations | Bytes |
|---:|---:|---:|---:|
| 32 | 70 µs | 758 | 74 KB |
| 128 | 281 µs | 2,932 | 296 KB |
| 512 | 1.13 ms | 11,587 | 1.17 MB |

Each fourfold input increase remains approximately fourfold in time and
memory. The attachment pass parses every raw source line once, visits each
rendered line associated with a source link once, and appends every parsed
link once. Its additional work is therefore `O(B + V + L)` time and
`O(B + L)` space, where `B` is raw input size, `V` rendered-line count, and
`L` parsed link count.
