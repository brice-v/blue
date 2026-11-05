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

# The native library branches are pinned to immutable commits. Update these
# only together with the expected checksums below.
LIBS_COMMIT_darwin=17c42a85217861f845ab1418a071b3b7c8f2b471
LIBS_COMMIT_linux=41e8774c38c431156245f1531a50f33a8743cc9f
LIBS_COMMIT_windows=f47d1b682eb818318a2e590970dfe2ae4bcf248b

# Expected sha256 values for every native library that may be copied into
# vendor/. Checked before anything lands in the tree.
expected_sha256() {
	case "$1" in
	darwin/amd64/libwgpu_native.a) echo dd61b0d36a1950352aec0fe377eece39849970f4a1be406aae8c4aa1e11af7f3 ;;
	darwin/arm64/libwgpu_native.a) echo 7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248 ;;
	linux/amd64/libwgpu_native.a) echo 011dc082d31c060202e519548fdc2163974b3d2042e64be95f5d143ce03dfbd4 ;;
	linux/arm64/libwgpu_native.a) echo 15b28059923b0d65e560ac40b35cfec1b8dc94f35383adadb5ac1eb7f890b64c ;;
	windows/386/wgpu_native.lib) echo e1789915a2c747457ecb390632767bcf88780c9f4b46b966617e09d5cbbebc6c ;;
	windows/amd64/libwgpu_native.a) echo 0b5206ac8313a33d5f276417328ed8d077e1c0cc1b25f3a700da6dd0f284b326 ;;
	windows/arm64/wgpu_native.lib) echo ecc7f102c06040c836b0d09ffc66596ff48913d662f60208c5aed9cc01f69cb3 ;;
	*) return 1 ;;
	esac
}

sha256_of() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		echo "error: no sha256sum or shasum available for verification" >&2
		return 1
	fi
}

verify_lib() {
	rel_path="$1"
	file="$2"
	expected="$(expected_sha256 "$rel_path")" || {
		echo "error: no pinned checksum for $rel_path; refusing to vendor an unverified library" >&2
		return 1
	}
	actual="$(sha256_of "$file")" || return 1
	if [ "$actual" != "$expected" ]; then
		echo "error: checksum mismatch for $rel_path" >&2
		echo "  expected: $expected" >&2
		echo "  actual:   $actual" >&2
		return 1
	fi
}

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

	eval "commit=\${LIBS_COMMIT_${system}}"
	if [ -z "${commit:-}" ]; then
		echo "error: no pinned commit for libs-$system" >&2
		return 1
	fi

	tmp="$(mktemp -d)"
	# shellcheck disable=SC2064
	trap "rm -rf '$tmp'" RETURN
	git -c core.autocrlf=false clone --depth 1 --branch "libs-$system" --single-branch "$REPO" "$tmp/repo" >/dev/null 2>&1
	# Check out the pinned commit, not whatever the branch points at today.
	git -C "$tmp/repo" fetch --depth 1 origin "$commit" >/dev/null 2>&1
	git -C "$tmp/repo" checkout --quiet --detach "$commit"

	mkdir -p "$VENDOR/libs-$system"
	for arch in "$tmp/repo/libs-$system"/*/; do
		[ -d "$arch" ] || continue
		name="$(basename "$arch")"
		for lib in "$arch"*; do
			[ -f "$lib" ] || continue
			verify_lib "$system/$name/$(basename "$lib")" "$lib" || return 1
		done
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
