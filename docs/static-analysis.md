# Static Analysis Record

The production-code investigation was run with:

    go version go1.26.3 darwin/arm64
    staticcheck 2026.2 (0.8.0)
    govulncheck@v1.1.4

No repository-specific analyzer configuration is present. Commands and scope:

    go test -run '^$' ./...            # compile every package
    go vet ./...                       # all repository packages
    staticcheck -tests=false ./...     # production Go files only
    govulncheck ./...                  # reachable vulnerability call paths

Results were ranked by reachable entry point, severity, and confidence:

| Lead | Reachable entry point | Classification | Evidence |
|---|---|---|---|
| `SA4006` in `internal/tui/model_test.go:324` | none (test only) | rejected | the value is assigned only before the test checks that `cmd` is non-nil; it cannot affect the CLI, Book, or TUI runtime |
| `go vet` production diagnostics | none | rejected | command completed without output |
| Staticcheck production diagnostics | none | rejected | `-tests=false` completed without output |
| called vulnerabilities | none | rejected | `govulncheck ./...` reported zero code vulnerabilities; indirect module advisories were not called by this program |

There were no retained, reproducible user-facing leads, so no bug issue was
filed. Staticcheck was included because it adds high-confidence correctness
checks beyond the compiler and `go vet`; its installed version is recorded
above to make the result reproducible.
