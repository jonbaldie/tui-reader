# Evidence index

The combined TUI logs are ordered screen snapshots from real tmux PTYs. The ANSI log retains style sequences from the diagnostic run with `NO_COLOR` unset.

| Journey | Actions and captures |
| --- | --- |
| Plain-text reading & navigation | [`plain-navigation.txt`](plain-navigation.txt) includes start screen, Space, Right, Left, l, h, PageDown, PageUp, Home, End, g, and G key navigation. |
| Dump CLI mode | [`dump-cli.txt`](dump-cli.txt) covers single and multi-page preview (`--dump`, `--dump=N`, `--dump=99999`), missing files, directory arguments, invalid UTF-8, missing args, and malformed options with exit codes. |
| Heading links and history | [`link-tour.txt`](link-tour.txt) exercises Tab forward selection, Shift+Tab backward selection, multiple links in prose paragraphs, bracketed link syntax, ignored code-block links, silently ignored missing anchors, and history returns with `b` and Backspace. |
| Link selection styling | [`link-selection-style.ansi`](link-selection-style.ansi) shows no-selection baseline, Alpha-selected inverted styling, and Beta-selected inverted styling. |
| Link history after resize | [`link-reflow.txt`](link-reflow.txt) verifies following a link, resizing from 80×24 to 48×14, and returning via `b` to the origin section page after reflow. |
| Reading position after resize | [`reader-reflow.txt`](reader-reflow.txt) verifies reading position preservation across resize (80×24 -> 48×14 -> 80×24) across two clean reader processes. |
| Input errors | [`input-errors.txt`](input-errors.txt) captures missing-file, directory, and invalid-UTF-8 error screens at 80×24 and wrapped error layout at 40×12, verifying clean exits with `q`. |
| Setup & Environment | [`environment.txt`](environment.txt), [`build-info.txt`](build-info.txt), [`build.sha256`](build.sha256), and isolated input fixtures under [`fixtures/`](fixtures/). |

Baseline runs were captured under `NO_COLOR=1`; link styling sequences were verified in a separate diagnostic run with `NO_COLOR` unset.
