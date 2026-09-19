#!/usr/bin/env bash
#
# build.sh builds the blue binary. It first fetches the WebGPU native libraries
# that the vendored binding needs (they are not Go packages, so `go mod vendor`
# cannot copy them and they are gitignored).
set -euo pipefail

cd "$(dirname "$0")"

./scripts/fetch-wgpu-libs.sh

go build -o blue .
