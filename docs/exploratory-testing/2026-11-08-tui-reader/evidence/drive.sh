#!/bin/bash
set -eu
cd "$(dirname "$0")/../../../.."
root=$PWD
out=docs/exploratory-testing/2026-11-08-tui-reader/evidence
sock=$root/.run/explore-tmux.sock
cleanup() { tmux -S "$sock" kill-server 2>/dev/null || true; }
trap cleanup EXIT HUP INT TERM
mkdir -p .run/home "$out/fixtures"
python3 - "$out/fixtures" <<'PY'
import pathlib, sys
p = pathlib.Path(sys.argv[1])
p.joinpath('tour.md').write_text('# Contents\n\n[Alpha](#alpha) [Beta](#beta) [Missing](#missing)\n\n' + ''.join(f'Intro paragraph {i:02d}.\n\n' for i in range(1, 25)) + '# Alpha\n\n[Go to Beta](#beta)\n\n' + ''.join(f'Alpha paragraph {i:02d}.\n\n' for i in range(1, 25)) + '# Beta\n\nBeta destination.\n')
p.joinpath('plain.txt').write_text('First line.\nSecond line.\n\n[Not a link](#missing)\n')
PY
run() {
  local name=$1 file=$2; shift 2
  tmux -S "$sock" new-session -d -s "$name" -x 80 -y 24 "cd '$root'; env HOME='$root/.run/home' TERM=screen-256color .run/tui-reader '$file'; printf '\\nEXIT=%s\\n' \$?"
  tmux -S "$sock" set-option -t "$name" remain-on-exit on
  sleep 1
  { echo '--- start 80x24'; tmux -S "$sock" capture-pane -p -t "$name"; } > "$out/$name.txt"
  for action in "$@"; do
    case "$action" in
      resize:*) size=${action#resize:}; tmux -S "$sock" resize-window -t "$name" -x "${size%x*}" -y "${size#*x}" ;;
      *) tmux -S "$sock" send-keys -t "$name" "$action" ;;
    esac
    sleep 0.4
    { printf '\n--- %s\n' "$action"; tmux -S "$sock" capture-pane -p -t "$name"; } >> "$out/$name.txt"
  done
  tmux -S "$sock" kill-session -t "$name"
}
run navigation "$out/fixtures/tour.md" Right Left End Home Right resize:60x20 resize:80x24 q
run resize-replay "$out/fixtures/tour.md" Right resize:60x20 resize:80x24 q
run links "$out/fixtures/tour.md" Tab Enter b Tab Tab Enter b BTab Enter b q
run missing "$out/fixtures/tour.md" Tab Tab Tab Enter b q
run plain "$out/fixtures/plain.txt" Right Tab Enter Left q
run error "$out/fixtures/not-found.md" q
{
  echo '$ tui-reader --dump=1 tour.md'; .run/tui-reader --dump=1 "$out/fixtures/tour.md"
  echo '$ tui-reader --dump tour.md'; .run/tui-reader --dump "$out/fixtures/tour.md"
  echo '$ tui-reader --dump=0 tour.md'; if .run/tui-reader --dump=0 "$out/fixtures/tour.md"; then echo EXIT=0; else echo EXIT=$?; fi
  echo '$ tui-reader --dump not-found.md'; if .run/tui-reader --dump "$out/fixtures/not-found.md"; then echo EXIT=0; else echo EXIT=$?; fi
} > "$out/dump.txt" 2>&1
python3 - "$out" <<'PY'
import pathlib, sys
for path in pathlib.Path(sys.argv[1]).glob('*.txt'):
    if path.name != 'setup.txt':
        path.write_text('\n'.join(line.rstrip() for line in path.read_text().splitlines()).rstrip() + '\n')
PY
