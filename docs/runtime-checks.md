# Runtime Check Record

Environment: macOS on Apple silicon, Go `go1.26.3 darwin/arm64`, clean build
from the issue branch.

Commands run:

    go test -race -count=1 ./...
    go build -o /tmp/tui-reader-p09 .
    /tmp/tui-reader-p09 --dump sample.md
    go test -race -count=1 ./internal/tui -run '^FuzzModelActions/491d46a7afd24cc2$' -v
    script -q tty.log /tmp/tui-reader-p09 sample.md   # send q on the PTY

The full race suite passed. Dump mode rendered the sample document through the
real binary. The saved minimized stateful-fuzz corpus replayed under `-race`.
The PTY workflow exited with status 0 and recorded both alternate-screen entry
(`?1049h`) and cleanup (`?1049l`) after the `q` key.

No race, panic, hang, leak report, or invalid concurrent behavior was found,
so there are no reports to group, minimize, or file. Go's race detector is the
supported runtime data-race check used here; Go does not provide a portable
general heap-leak detector for this application, and address/memory sanitizers
are not supported by this macOS Go setup.
