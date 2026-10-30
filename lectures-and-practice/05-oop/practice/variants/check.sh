#!/bin/sh
set -eu

variants_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

for variant in "$variants_dir"/[0-9][0-9]-*; do
    [ -d "$variant" ] || continue
    printf '\n%s\n' "Checking $(basename "$variant")"
    (
        cd "$variant"
        go test ./...
        go vet ./...
        go run ./cmd/demo
    )
done
