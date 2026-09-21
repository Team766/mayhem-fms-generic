#!/bin/sh
# Lists upstream commits after the UPSTREAM.md checkpoint, oldest first, with the paths each touches.
# Usage: scripts/agents/upstream-range.sh [upstream-ref]   (default: upstream/main). Does not fetch.
set -eu
cd "$(git rev-parse --show-toplevel)"
ref="${1:-upstream/main}"
checkpoint=$(sed -n 's/^Last reviewed upstream commit: `\([0-9a-f]\{7,40\}\)`.*/\1/p' UPSTREAM.md)
if [ -z "$checkpoint" ]; then
  echo "UPSTREAM.md has no checkpoint yet; use regenerate mode." >&2
  exit 1
fi
echo "# $checkpoint..$ref ($(git rev-list --count "$checkpoint..$ref") commits)"
git log --reverse --date=short --format='%n%h %ad %s' --name-only "$checkpoint..$ref"
