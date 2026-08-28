# Wrapping Performance

`BenchmarkWrapLines` covers many short words, one long ASCII word, and one
long multibyte Unicode word at three input sizes. Run it with:

    go test -run '^$' -bench '^BenchmarkWrapLines$' -benchmem -count=3 ./internal/book

On an Apple M3, the median baseline and optimized results were:

| Workload | Input | Before | After | Allocations before → after |
|---|---:|---:|---:|---:|
| short words | 1 KB | 9.8 µs | 5.9 µs | 186 → 30 |
| short words | 4 KB | 38.8 µs | 23.4 µs | 726 → 107 |
| long ASCII word | 1 KB | 73 µs | 5.1 µs | 108 → 30 |
| long ASCII word | 4 KB | 1.27 ms | 19.8 µs | 418 → 107 |
| long Unicode word | 1 KB | 30.1 µs | 6.2 µs | 39 → 13 |
| long Unicode word | 4 KB | 418 µs | 24.3 µs | 145 → 39 |

The optimized 16 KB measurements were 92.8 µs for short words, 80.3 µs for a
long ASCII word, and 96.9 µs for a long Unicode word, confirming linear growth
as input quadruples. The wrapper scans fields and runes once and writes each
output rune once, for `O(x + output)` time and `O(width + output)` space.
