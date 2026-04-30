#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SMOKE_OUT_DIR="${SMOKE_OUT_DIR:-/tmp/routepeek-release-check}"

echo "==> Go tests"
go test ./...

echo "==> Frontend tests"
pushd "$ROOT_DIR/web" >/dev/null
npm test

echo "==> Frontend build"
npm run build
popd >/dev/null

echo "==> Go CLI build"
CGO_ENABLED=0 go build -trimpath -o /tmp/routepeek-check-cli "$ROOT_DIR/cmd/cli"

echo "==> CLI check/report smoke"
/tmp/routepeek-check-cli check --json >/dev/null
/tmp/routepeek-check-cli report --format json >/dev/null

echo "==> Go web server build"
CGO_ENABLED=0 go build -trimpath -o /tmp/routepeek-check-web "$ROOT_DIR/cmd/server"

echo "==> Release script smoke build (Linux/macOS/Windows)"
rm -rf "$SMOKE_OUT_DIR"
OUT_DIR="$SMOKE_OUT_DIR" "$ROOT_DIR/scripts/build-release.sh" --all --skip-frontend

echo "==> Release check passed"
