#!/usr/bin/env sh
set -eu

# This script records environment metadata and runs hey if installed.
# Run only against a service you own or are authorized to test.
URL=${1:-http://127.0.0.1:8080/hello}
DURATION=${DURATION:-20s}
CONCURRENCY=${CONCURRENCY:-20}
OUT=${OUT:-results-$(date +%Y%m%d-%H%M%S)}
mkdir -p "$OUT"
{
  echo "date=$(date -Is)"
  echo "url=$URL"
  echo "duration=$DURATION"
  echo "concurrency=$CONCURRENCY"
  go version 2>&1 || true
  uname -a 2>&1 || true
  if command -v lscpu >/dev/null 2>&1; then lscpu 2>&1; fi
} > "$OUT/environment.txt"
if ! command -v hey >/dev/null 2>&1; then
  echo "hey not found. Install it from a trusted source, then rerun." >&2
  exit 2
fi
hey -z "$DURATION" -c "$CONCURRENCY" "$URL" > "$OUT/hey.txt" 2>&1
printf 'Results saved under %s\n' "$OUT"
