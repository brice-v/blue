#!/bin/sh
# make_wasm.sh builds the blue language runtime as a WebAssembly module for
# the browser playground and copies everything the playground needs into
# ./playground.
#
# Usage:
#   ./make_wasm.sh            build blue.wasm and copy support files
#
# Then serve the playground directory with any static file server, eg:
#   cd playground && python3 -m http.server 8787
set -o errexit
set -o nounset

cd "$(dirname "$0")"

PLAYGROUND_DIR="playground"
WASM_OUT="$PLAYGROUND_DIR/blue.wasm"
GO_WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"

if [ ! -f "$GO_WASM_EXEC" ]; then
	echo "error: could not find $GO_WASM_EXEC" >&2
	echo "(go releases ship it under lib/wasm since go 1.21)" >&2
	exit 1
fi

echo "==> building $WASM_OUT"
# - static: drop the raylib/fyne gui modules which have no js/wasm support
#   (same tag used by make_release_static)
# - wasm:   swap the db/config/psutil modules for stubs that report clear
#   runtime errors instead of failing to compile (see object/*_wasm.go)
CGO_ENABLED=0 GOOS=js GOARCH=wasm \
	go build -tags "static,wasm" -trimpath \
	-ldflags="-s -w" \
	-o "$WASM_OUT" ./wasmmain

echo "==> copying $(basename "$GO_WASM_EXEC") into $PLAYGROUND_DIR/"
cp "$GO_WASM_EXEC" "$PLAYGROUND_DIR/wasm_exec.js"

echo "==> done. serve it up with:"
echo "    cd $PLAYGROUND_DIR && python3 -m http.server 8787"
