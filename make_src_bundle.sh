#!/bin/sh
set -o errexit
set -o nounset

# Builds cmd/srcbundle/blue-src.tar.gz from tracked source files.
# Excludes vendor, test programs, and other test specific files so the
# embedded bundle stays small and contains only what `go build ./cmd/bluerun`
# needs plus the language stdlib.
#
# Usage:
#   ./make_src_bundle.sh [output-path]
# Defaults to cmd/srcbundle/blue-src.tar.gz relative to the repo root.

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
OUT="${1:-$ROOT/cmd/srcbundle/blue-src.tar.gz}"

mkdir -p "$(dirname "$OUT")"

cd "$ROOT"

FILTERED_LIST="$(mktemp)"
trap 'rm -f "$FILTERED_LIST"' EXIT INT TERM

git ls-files \
  | grep -v '^vendor/' \
  | grep -v '^ignored/' \
  | grep -v '^playground/' \
  | grep -v '^b_test_programs/' \
  | grep -v '^manual_tests/' \
  | grep -v '^man/' \
  | grep -v '^\.github/' \
  | grep -v '_test\.go$' \
  | grep -v 'testdata' \
  | grep -v '\.out$' \
  | grep -v '^test\.b$' \
  | grep -v '/test\.b$' \
  | grep -v '^README\.md$' \
  | grep -v '^blue-TODO\.txt$' \
  | grep -v '^b\.txt$' \
  | grep -v '^hf\.md$' \
  | grep -v '^scratchfile\.b$' \
  | grep -v '^gen-man\.sh$' \
  | grep -v '^benchmark-things\.' \
  | grep -v '^make_' \
  | grep -v '^parser/parser_illegal_tok' \
  | grep -v '^\.gitignore$' \
  | grep -v '^cmd/srcbundle/blue-src\.tar\.gz$' \
  > "$FILTERED_LIST"

tar -cf - -T "$FILTERED_LIST" | gzip -9n > "$OUT"

ls -lh "$OUT"
