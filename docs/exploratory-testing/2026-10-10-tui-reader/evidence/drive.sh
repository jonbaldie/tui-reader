#!/bin/bash
# usage: drive.sh <session> <cols> <rows> <file> <keys...>  ; each key group "K:<tmux keys>" or "R:<cols>x<rows>"
S=$1; C=$2; R=$3; FILE=$4; shift 4
tmux -L explore kill-session -t $S 2>/dev/null
tmux -L explore new-session -d -s $S -x $C -y $R "env -u NO_COLOR TERM=screen-256color HOME=$TMPDIR/home ${READER:-$TMPDIR/tui-reader} '$FILE'; echo EXIT=\$?; sleep 30"
sleep 0.8
echo "--- start (${C}x${R})"; tmux -L explore capture-pane -p -t $S
for a in "$@"; do
  case $a in
    R:*) d=${a#R:}; tmux -L explore resize-window -t $S -x ${d%x*} -y ${d#*x};;
    K:*) tmux -L explore send-keys -t $S ${a#K:};;
  esac
  sleep 0.5
  echo "--- after $a"; tmux -L explore capture-pane -p -t $S
done
tmux -L explore kill-session -t $S 2>/dev/null
