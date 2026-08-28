# Anchor normalization performance

`NormalizeAnchor` now filters, lowercases, converts runs of spaces and hyphens,
and trims boundary separators in one pass. It retains only ASCII letters and
digits, matching the prior output for punctuation and Unicode input.

The adversarial benchmark combines a long hyphen run with unaffected text:

```sh
go test -run '^$' -bench '^BenchmarkNormalizeAnchorAdversarial$' -benchmem ./internal/book
```

On an Apple M3, the prior implementation took 14.4 µs, 219 µs, and 3.62 ms
for 1 KiB, 16 KiB, and 256 KiB inputs, respectively. The single-pass version
took 3.12 µs, 47.2 µs, and 744 µs (about 4.6–4.9× faster) with linear scaling
and substantially fewer allocations.
