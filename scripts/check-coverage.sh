#!/usr/bin/env bash
set -euo pipefail

threshold=${1:-90}

go test ./... -coverprofile=coverage.out -covermode=atomic

total=$(go tool cover -func=coverage.out | tail -1 | awk '{print substr($3, 1, length($3)-1)}')

printf "Total coverage: %s%%\n" "$total"

awk -v t="$threshold" -v v="$total" 'BEGIN { if (v+0 < t+0) { exit 1 } }'
