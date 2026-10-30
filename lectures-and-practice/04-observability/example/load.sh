#!/usr/bin/env sh
# Usage: sh load.sh [url] [rps]; runs until Ctrl+C or SIGTERM.
set -eu
[ "$#" -le 2 ] || { echo 'Usage: sh load.sh [url] [rps]' >&2; exit 2; }
cd "$(dirname "$0")"
go build -o bin/load ./load
exec ./bin/load -url="${1:-http://127.0.0.1:8080/orders}" -rps="${2:-5}"
