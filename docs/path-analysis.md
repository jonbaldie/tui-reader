# User-Facing Path Analysis

Baseline command (run from the repository root):

    go test -coverpkg=./... -coverprofile=/tmp/tui-reader-p06.out ./...
    go tool cover -func=/tmp/tui-reader-p06.out

The baseline exercised 95.4% of production statements. The remaining
user-facing decisions were selected by impact and traced as follows:

| Decision | Prior conditions | Reversing input/state | Result |
|---|---|---|---|
| CLI usage path | no positional argument | `run(nil, stdout, stderr)` | documented existing test: exits 1 and prints usage |
| CLI dump load error | `--dump` and a path | nonexistent path | documented existing test: exits 1 without stdout |
| TUI initialization | a model has been created | call `Init` | new test: no unexpected startup command |
| TUI message guard | a valid loaded model | pass an unknown `tea.Msg` | new test: model and command remain unchanged |
| Interactive program error | a non-dump invocation needs a terminal | Bubble Tea startup failure | not reproducible in the non-interactive test harness; `run` already reports its returned error |

No stable failure was found. The invalid-dimension, link-deduplication, and
resize/navigation guards are exercised by the generated Book and TUI test
suites introduced in the preceding issues. No confirmed bug was filed.
