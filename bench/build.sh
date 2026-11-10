#!/bin/sh
set -eu

# Go has no separate install step in its normal build. `go mod download` is the closest equivalent,
# so the install markers wrap it and the build markers wrap the compile and link.

# The commit SHA is embedded at build time, so /version reports what was built, not a runtime setting.
# BENCH_COMMIT is the harness override (Fly passes it as a build arg); the others are provider variables.
commit_sha() {
  for sha in "${BENCH_COMMIT:-}" "${RENDER_GIT_COMMIT:-}" "${RAILWAY_GIT_COMMIT_SHA:-}" "${VERCEL_GIT_COMMIT_SHA:-}"; do
    if [ -n "$sha" ]; then
      printf '%s\n' "$sha"
      return
    fi
  done
  git rev-parse HEAD 2>/dev/null || printf '%s\n' unknown
}

install() {
  printf '%s\n' '::bench::install_start'
  go version
  go mod download
  printf '%s\n' '::bench::install_done'
}

build() {
  printf '%s\n' '::bench::build_start'
  # Static binary without cgo, so the same build runs on native runtimes and in a distroless image.
  CGO_ENABLED=0 go build -ldflags "-X main.commit=$(commit_sha)" -o app .
  printf '%s\n' '::bench::build_done'
}

case "${1:-all}" in
  install) install ;;
  build) build ;;
  all) install; build ;;
  *) printf 'Usage: %s {install|build|all}\n' "$0" >&2; exit 2 ;;
esac
