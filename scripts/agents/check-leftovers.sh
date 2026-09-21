#!/bin/sh
# Fails (exit 1) if vocabulary that should be gone is still in the tree.
#   check-leftovers.sh base                 stripped integrations, upstream season game, LEDs, Lite-isms
#   check-leftovers.sh game <words-file>    outgoing game vocabulary, one extended regex per line
# Update BASE_PATTERN when upstream's season game changes (it names the 2026 game today).
set -eu
cd "$(git rev-parse --show-toplevel)"
BASE_PATTERN='hub|fuel|tower|shift_change|energized|supercharged|traversal|twitch|nexus|tbaclient|tbapublish|teamsign|team_sign|ledcontroller|/api/scores|FoulPointsAgainst|cheesy-arena-lite'
case "${1:-}" in
  base) pattern="$BASE_PATTERN" ;;
  game) pattern=$(grep -v '^[[:space:]]*\(#\|$\)' "${2:?words file required}" | paste -sd'|' -) ;;
  *) echo "usage: $0 base | game <words-file>" >&2; exit 2 ;;
esac
hits=$(git grep -n -i -E "$pattern" -- '*.go' '*.html' '*.js' '*.css' '*.csv' \
  ':!static/js/lib' ':!static/css/lib' ':!docs' ':!specs' ':!scripts' | grep -v -E 'TbaMatchKey|github\.com' || true)
if [ -n "$hits" ]; then
  echo "$hits"
  echo "leftovers: $(printf '%s\n' "$hits" | wc -l | tr -d ' ') line(s)" >&2
  exit 1
fi
echo "no leftovers"
