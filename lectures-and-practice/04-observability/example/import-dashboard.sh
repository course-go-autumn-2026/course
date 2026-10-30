#!/usr/bin/env sh
# Usage: GRAFANA_URL=http://localhost:22300 OVERWRITE=false sh import-dashboard.sh
# Uses the anonymous Admin access of the local tripgoctl teaching environment.
set -eu
url="${GRAFANA_URL:-http://localhost:22300}"
overwrite="${OVERWRITE:-false}"
case "$overwrite" in true|false) ;; *) echo 'OVERWRITE must be true or false' >&2; exit 2;; esac

dashboard=$(cat "$(dirname -- "$0")/dashboard.json")
printf '{"dashboard":%s,"overwrite":%s}\n' "$dashboard" "$overwrite" |
  curl --fail-with-body --silent --show-error --proto '=http,https' \
    --connect-timeout 3 --max-time 15 \
    --header 'Content-Type: application/json' --data-binary @- \
    --url "${url%/}/api/dashboards/db"
printf '\n'
