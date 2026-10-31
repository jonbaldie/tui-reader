#!/bin/bash
# snap.sh SESSION [keys...]: send keys one by one then print pane
S=$1; shift
for k in "$@"; do tmux send-keys -t "$S" "$k"; sleep 0.25; done
sleep 0.3
echo "----- [$S] after: $* -----"
tmux capture-pane -p ${ESC:+-e} -t "$S"
