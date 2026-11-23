#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/code"
gofmt -w $(find . -type f -name '*.go')
go test -race ./...
go vet ./...
go build ./...
for module in "$ROOT"/stages/* "$ROOT"/experiments/epoll-demo; do
  [[ -f "$module/go.mod" ]] || continue
  echo "==> $module"
  (cd "$module" && go test ./... && go vet ./... && go build ./...)
done
