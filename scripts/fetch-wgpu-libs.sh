#!/usr/bin/env bash
#
# Fetch the prebuilt WebGPU native artifacts that the vendored
# github.com/oliverbestmann/webgpu module needs, but that `go mod vendor`
# cannot copy because they are not Go packages:
#
#   * headers: the module ships them as wgpu/lib/<system>/<arch>/*.h
#   * libraries: they live on per-system git branches (libs-linux, libs-darwin,
#     libs-windows) as libs-<system>/<arch>/libwgpu_native.a (or wgpu_native.lib)
#
# Both are copied under vendor/ so the cgo ${SRCDIR} flags in the bindings
# (-I.../wgpu/lib/<system>/<arch> and -L.../libs-<system>/<arch>) resolve.
# Re-running `go mod vendor` deletes them, so run this again after.
#
# Usage:
#   scripts/fetch-wgpu-libs.sh           fetch for the host GOOS
#   scripts/fetch-wgpu-libs.sh --all     fetch linux, darwin and windows
#   scripts/fetch-wgpu-libs.sh --force   re-download even when already present
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VENDOR="$ROOT/vendor/github.com/oliverbestmann/webgpu"
MODULE="github.com/oliverbestmann/webgpu"
REPO="https://github.com/oliverbestmann/webgpu"

FORCE=0
ALL=0
for arg in "$@"; do
	case "$arg" in
	--all) ALL=1 ;;
	--force) FORCE=1 ;;
	*)
		echo "unknown argument: $arg" >&2
		exit 2
		;;
	esac
done

if [ ! -d "$VENDOR" ]; then
	echo "error: $VENDOR not found; run 'go mod vendor' first" >&2
	exit 1
fi

# The headers live in the module source, so make sure it is downloaded and then
# ask Go where it landed. -mod=mod is required because the main module vendors
# its dependencies.
go mod download "$MODULE"
# tr keeps the path usable under Git Bash on Windows, where go prints backslashes
MODDIR="$(go list -m -mod=mod -f '{{.Dir}}' "$MODULE" | tr '\\' '/')"
if [ -z "$MODDIR" ] || [ ! -d "$MODDIR" ]; then
	echo "error: could not locate $MODULE in the module cache" >&2
	exit 1
fi

if [ "$ALL" -eq 1 ]; then
	systems="linux darwin windows"
else
	case "$(go env GOOS)" in
	linux | darwin | windows) systems="$(go env GOOS)" ;;
	*)
		echo "no WebGPU native libraries needed for GOOS=$(go env GOOS)"
		exit 0
		;;
	esac
fi

copy_headers() {
	system="$1"
	[ -d "$MODDIR/wgpu/lib/$system" ] || return 0
	mkdir -p "$VENDOR/wgpu/lib/$system"
	# module cache files are read-only, so make any previous copy writable first
	chmod -R u+w "$VENDOR/wgpu/lib/$system" 2>/dev/null || true
	cp -Rf "$MODDIR/wgpu/lib/$system"/. "$VENDOR/wgpu/lib/$system"/
	chmod -R u+w "$VENDOR/wgpu/lib/$system" 2>/dev/null || true
	echo "headers: $system"
}

has_libs() {
	find "$VENDOR/libs-$1" -maxdepth 2 \( -name '*.a' -o -name '*.lib' \) 2>/dev/null | grep -q .
}

copy_libs() {
	system="$1"
	if [ "$FORCE" -eq 0 ] && has_libs "$system"; then
		echo "libs: $system (already present)"
		return 0
	fi

	tmp="$(mktemp -d)"
	# shellcheck disable=SC2064
	trap "rm -rf '$tmp'" RETURN
	git -c core.autocrlf=false clone --depth 1 --branch "libs-$system" --single-branch "$REPO" "$tmp/repo" >/dev/null 2>&1

	mkdir -p "$VENDOR/libs-$system"
	for arch in "$tmp/repo/libs-$system"/*/; do
		[ -d "$arch" ] || continue
		name="$(basename "$arch")"
		mkdir -p "$VENDOR/libs-$system/$name"
		chmod -R u+w "$VENDOR/libs-$system/$name" 2>/dev/null || true
		cp -Rf "$arch". "$VENDOR/libs-$system/$name"/
	done
	echo "libs: $system"
}

for system in $systems; do
	copy_headers "$system"
	copy_libs "$system"
done

echo "WebGPU native libraries are ready"
