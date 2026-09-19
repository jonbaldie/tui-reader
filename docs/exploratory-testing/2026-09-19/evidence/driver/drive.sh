#!/bin/bash
# usage: drive.sh SESSION COLSxROWS FILE  -> starts reader in tmux
# then: tmux send-keys -t S ... ; tmux capture-pane -p -t S
S=$1; W=${2%x*}; H=${2#*x}; F=$3
tmux kill-session -t "$S" 2>/dev/null
tmux new-session -d -s "$S" -x "$W" -y "$H" "TERM=xterm-256color /tmp/tr-et0919 '$F'; echo EXIT=\$?; sleep 30"
sleep 0.6
