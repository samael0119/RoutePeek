#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist}"
BUILD_ALL=0
SKIP_FRONTEND=0

usage() {
  cat <<'USAGE'
Usage: scripts/build-release.sh [--all] [--skip-frontend]

Build RoutePeek release binaries.

Environment:
  VERSION=1.2.3        Override version label used in output paths.
  OUT_DIR=./dist       Override output directory.

Options:
  --all                Build Linux/macOS/Windows amd64+arm64 targets.
  --skip-frontend      Do not rebuild web assets before building server.
  -h, --help           Show this help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)
      BUILD_ALL=1
      shift
      ;;
    --skip-frontend)
      SKIP_FRONTEND=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

host_goos() {
  go env GOOS
}

host_goarch() {
  go env GOARCH
}

build_frontend() {
  if [[ "$SKIP_FRONTEND" -eq 1 ]]; then
    echo "==> Skipping frontend build"
    return
  fi

  if [[ ! -f "$ROOT_DIR/web/package.json" ]]; then
    echo "==> No web/package.json found; skipping frontend build"
    return
  fi

  require_cmd npm
  echo "==> Building Vue frontend"
  pushd "$ROOT_DIR/web" >/dev/null
  if [[ -f package-lock.json ]]; then
    npm ci
  else
    npm install
  fi
  npm run build
  popd >/dev/null
}

package_binary() {
  local import_dir="$1"
  local output_name="$2"
  local goos="$3"
  local goarch="$4"
  local package_dir="$OUT_DIR/routepeek-${VERSION}-${goos}-${goarch}"
  local ext=""

  if [[ "$goos" == "windows" ]]; then
    ext=".exe"
  fi

  mkdir -p "$package_dir"

  echo "==> Building $output_name for $goos/$goarch"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags="-s -w" -o "$package_dir/${output_name}${ext}" "./cmd/$import_dir"
}

archive_package() {
  local goos="$1"
  local goarch="$2"
  local package_dir="$OUT_DIR/routepeek-${VERSION}-${goos}-${goarch}"
  local archive_base="$OUT_DIR/routepeek-${VERSION}-${goos}-${goarch}"

  if [[ "$goos" == "windows" ]]; then
    (cd "$OUT_DIR" && zip -qr "$(basename "$archive_base").zip" "$(basename "$package_dir")")
  else
    tar -C "$OUT_DIR" -czf "${archive_base}.tar.gz" "$(basename "$package_dir")"
  fi
}

main() {
  require_cmd go
  if [[ "$BUILD_ALL" -eq 1 ]]; then
    require_cmd tar
    require_cmd zip
  fi

  cd "$ROOT_DIR"
  mkdir -p "$OUT_DIR"
  build_frontend

  local targets=()
  if [[ "$BUILD_ALL" -eq 1 ]]; then
    targets=(
      "linux/amd64"
      "linux/arm64"
      "darwin/amd64"
      "darwin/arm64"
      "windows/amd64"
      "windows/arm64"
    )
  else
    targets=("$(host_goos)/$(host_goarch)")
  fi

  for target in "${targets[@]}"; do
    local goos="${target%/*}"
    local goarch="${target#*/}"
    package_binary cli routepeek "$goos" "$goarch"
    package_binary server routepeek-web "$goos" "$goarch"
    archive_package "$goos" "$goarch"
  done

  echo "==> Release artifacts written to $OUT_DIR"
}

main "$@"
