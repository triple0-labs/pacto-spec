#!/bin/sh
set -eu
if [ -n "$(git status --porcelain 2>/dev/null || true)" ]; then
  echo "Git worktree has uncommitted changes" >&2
  exit 2
fi
exit 0
