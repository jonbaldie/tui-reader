# Evidence index

The combined TUI logs are ordered screen snapshots from real tmux PTYs. Empty
rows after the last visible line are omitted. The ANSI log retains style
sequences from the diagnostic run with `NO_COLOR` unset.

| Journey | Actions and captures |
| --- | --- |
| Plain-text reading | [`plain-navigation.txt`](plain-navigation.txt) includes the start screen, Space, Right, Home, and End states. |
| Dump mode | [`dump-cli.txt`](dump-cli.txt) includes the first-page output and malformed-option response with exit codes. |
| Heading links and history | [`link-tour.txt`](link-tour.txt) includes Tab, Shift+Tab, Alpha and Beta destinations, returns, and the ignored missing anchor. |
| Link selection styling | [`link-selection-style.ansi`](link-selection-style.ansi) shows the no-selection, Alpha-selected, and Beta-selected states. |
| Link history after resize | [`link-reflow.txt`](link-reflow.txt) records the destination, 40×12 reflow, and return to the link index. |
| Reading position after resize | [`reader-reflow.txt`](reader-reflow.txt) contains the round trip and its fresh-process replay. |
| Input errors | [`input-errors.txt`](input-errors.txt) contains missing-file and invalid-UTF-8 screens at 80×24 and a missing-file screen at 40×12. |
| Setup | [`environment.txt`](environment.txt), [`build-info.txt`](build-info.txt), [`build.sha256`](build.sha256), and isolated inputs in [`fixtures/`](fixtures/). |

The host had `NO_COLOR=1`, so baseline runs correctly suppressed link styling.
